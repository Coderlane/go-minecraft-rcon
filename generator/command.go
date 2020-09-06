package generator

// Argument holds a single argument for a command
type Argument struct {
	Values   []string
	Optional bool
}

// Command is a single RCON command with its arguments and subcommands.
type Command struct {
	Name        string
	Subcommands []Command
	Arguments   []Argument
}

// Alias maps an alias name to the real command name
type Alias struct {
	AliasName   string
	CommandName string
}
