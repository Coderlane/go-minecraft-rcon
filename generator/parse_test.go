package generator

import (
	"strings"
	"testing"
)

const (
	AliasSample                  string = "tp -> teleport"
	CommandSample                string = "/seed"
	CommandWithArgSample         string = "/me <action>"
	CommandWithOptionalArgSample string = "/clear [<targets>]"
	CommandWithSubcommandSample  string = "/recipe (give|take)"
)

func TestParseAlias(t *testing.T) {
	out, err := ParseCommandWithAlias(AliasSample)
	if err != nil {
		t.Fatalf("Failed to parse alias: %s", err)
	}
	alias, ok := out.(*Alias)
	if !ok {
		t.Fatalf("Output was not an alias: %T", alias)
	}
	if alias.AliasName != "tp" {
		t.Errorf("Expected alias name 'tp' got: '%v'", alias.AliasName)
	}
	if alias.CommandName != "teleport" {
		t.Errorf("Expected command name 'tp' got: '%v'", alias.CommandName)
	}
	cmd, _ := ParseCommand(AliasSample)
	if cmd != nil {
		t.Errorf("Expected nil command")
	}
}

func TestParseAliasFailures(t *testing.T) {
	type testCase struct {
		input string
		err   string
	}
	testCases := []testCase{
		{"/{} -> teleport", "Invalid command name: {}"},
		{"/tp -> []", "Invalid command name: []"},
	}
	for _, tcase := range testCases {
		t.Run(tcase.input, func(t *testing.T) {
			_, err := ParseCommandWithAlias(tcase.input)
			if err == nil {
				t.Fatalf("Expected error parsing: %s", tcase.input)
			}
			if !strings.Contains(err.Error(), tcase.err) {
				t.Errorf("Expected error: '%s' to contain: '%s'",
					err.Error(), tcase.err)
			}
		})
	}
}

func TestParseCommand(t *testing.T) {
	cmd, err := ParseCommand(CommandSample)
	if err != nil {
		t.Fatalf("Failed to parse alias: %s", err)
	}
	if cmd.Name != "seed" {
		t.Errorf("Expected alias name 'tp' got: '%v'", cmd.Name)
	}
}

func TestParseCommandFailures(t *testing.T) {
	type testCase struct {
		input string
		err   string
	}
	testCases := []testCase{
		{"/{}", "Invalid command name: {}"},
	}
	for _, tcase := range testCases {
		t.Run(tcase.input, func(t *testing.T) {
			_, err := ParseCommand(tcase.input)
			if err == nil {
				t.Fatalf("Expected error parsing: %s", tcase.input)
			}
			if !strings.Contains(err.Error(), tcase.err) {
				t.Errorf("Expected error: '%s' to contain: '%s'",
					err.Error(), tcase.err)
			}
		})
	}
}

func TestParseAndPrint(t *testing.T) {
	testCases := []string{
		AliasSample,
		CommandSample,
		CommandWithArgSample,
		CommandWithOptionalArgSample,
		CommandWithSubcommandSample,
	}
	for _, tcase := range testCases {
		t.Run(tcase, func(t *testing.T) {
			cmd, err := ParseCommandWithAlias(tcase)
			if err != nil {
				t.Fatalf("Error parsing: %v", err)
			}
			if tcase[0] == '/' {
				tcase = tcase[1:]
			}
			str := cmd.String()
			if str != tcase {
				t.Errorf("\nExpected: %s\nGot:      %s", tcase, str)
			}
		})
	}
}
