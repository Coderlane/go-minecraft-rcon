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
func ParseCommandWithAlias(line string) (interface{}, error) {
	if len(line) == 0 {
		return nil, nil
	}
	if line[0] == '/' {
		line = line[1:]
	}
	parts := strings.Split(line, " ")
	alias, err := maybeParseAlias(parts)
	if err != nil || alias != nil {
		return alias, err
	}
	return nil, nil
}
