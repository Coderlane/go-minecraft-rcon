package generator

import (
	"fmt"
	"regexp"
	"strings"
)

var cmdRegex = regexp.MustCompile(`^[a-zA-Z_\-]+$`)

func validateCommand(cmd string) error {
	if cmdRegex.MatchString(cmd) {
		return nil
	}
	return fmt.Errorf("Invalid command name: %s", cmd)
}

func maybeParseAlias(parts []string) (*Alias, error) {
	if len(parts) != 3 || parts[1] != "->" {
		return nil, nil
	}
	alias := &Alias{
		AliasName:   parts[0],
		CommandName: parts[2],
	}
	err := validateCommand(alias.AliasName)
	if err != nil {
		return nil, err
	}
	err = validateCommand(alias.CommandName)
	if err != nil {
		return nil, err
	}
	return alias, nil
}

func parseCommandFromParts(parts []string) (*Command, error) {
	name := parts[0]
	err := validateCommand(name)
	if err != nil {
		return nil, err
	}
	cmd := &Command{
		Name: name,
	}
	for _, part := range parts[1:] {
		optional := false
		name := ""
		// Check for optional parameters
		if part[0] == '[' && part[len(part)-1] == ']' {
			optional = true
			part = part[1 : len(part)-1]
		}
		// Strip array parens
		if part[0] == '(' && part[len(part)-1] == ')' {
			part = part[1 : len(part)-1]
		}
		// Check for sub-values
		values := strings.Split(part, "|")
		if len(values) <= 1 {
			values = []string{}
		}
		// Check for an argument name
		if part[0] == '<' && part[len(part)-1] == '>' {
			name = part[1 : len(part)-1]
		}
		cmd.Arguments = append(cmd.Arguments, Argument{
			Optional: optional,
			Name:     name,
			Values:   values,
		})
	}
	return cmd, nil
}

// ParseCommand parses only commands
func ParseCommand(line string) (*Command, error) {
	out, err := ParseCommandWithAlias(line)
	if err != nil {
		return nil, err
	}
	cmd, ok := out.(*Command)
	if ok {
		return cmd, nil
	}
	return nil, nil
}

// ParseCommandWithAlias parses commands as well as aliases
func ParseCommandWithAlias(line string) (CommandInterface, error) {
	if line[0] == '/' {
		line = line[1:]
	}
	parts := strings.Split(line, " ")
	alias, err := maybeParseAlias(parts)
	if err != nil || alias != nil {
		return alias, err
	}
	return parseCommandFromParts(parts)
}
