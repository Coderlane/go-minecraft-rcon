package generator

import "strings"

// CommandInterface represents commands and aliases.
type CommandInterface interface {
	String() string
}

// Argument holds a single argument for a command.
type Argument struct {
	Name     string
	Values   []string
	Optional bool
}

// Command is a single RCON command with its arguments and subcommands.
type Command struct {
	Name      string
	Arguments []Argument
}

// String prints the command in the same format it was parsed in.
func (cmd *Command) String() string {
	if len(cmd.Arguments) == 0 {
		return cmd.Name
	}
	out := cmd.Name
	for _, arg := range cmd.Arguments {
		out += " "
		tail := ""
		if arg.Optional {
			out += "["
			tail = "]"
		} else if len(arg.Values) != 0 {
			out += "("
			tail = ")"
		}
		if len(arg.Name) != 0 {
			out += "<" + arg.Name + ">"
		} else if len(arg.Values) != 0 {
			out += strings.Join(arg.Values, "|")
		}
		out += tail
	}
	return out
}

// Alias maps an alias name to the real command name.
type Alias struct {
	AliasName   string
	CommandName string
}

// String prints the alias in the same format it was printed in.
func (alias *Alias) String() string {
	return alias.AliasName + " -> " + alias.CommandName
}
