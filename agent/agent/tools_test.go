package agent

import (
	_ "embed"
	"reflect"
)

// Shell runs a shell command.
//
// Parameters:
//   - command: The shell command to run.
func Shell(command string) (string, error) {
	return "hello world", nil
}

// PatchFile patches the file at the specified path by replacing before with
// after.
//
// Parameters:
//   - path: The path to the file, in the format "path/to/file.txt"
//   - before: The content that will be replaced
//   - after: The content it will be replaced with
func PatchFile(path, before, after string) (string, error) {
	// ^^ note, all params share a type, which is interesting to parse
	return "Successfully replaced before with after.", nil
}

//go:embed tools_test.go
var toolSource string

var testConfig = &Config{
	SystemPrompt: "You are a friendly assistant that uses tools to help users.",
	ToolSource:   toolSource,
	Tools: map[string]reflect.Value{
		"shell":      reflect.ValueOf(Shell),
		"patch_file": reflect.ValueOf(PatchFile),
	},
}
