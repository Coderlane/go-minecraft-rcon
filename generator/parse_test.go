package generator

import (
	"strings"
	"testing"
)

func TestParseAlias(t *testing.T) {
	aliasStr := "/tp -> teleport"
	out, err := ParseCommandWithAlias(aliasStr)
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
}

func TestParseAliasFailures(t *testing.T) {
	type testCase struct {
		input string
		err   string
	}
	testCases := []testCase{
		{"tp -> teleport", "expected leading"},
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
