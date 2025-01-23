package agent

import (
	_ "embed"
	"reflect"
	"testing"

	"github.com/parakeet-nest/parakeet/llm"
	"github.com/stretchr/testify/require"
)

func TestToLowerSnakeCase(t *testing.T) {
	require.Equal(t, "hello_world", toLowerSnakeCase("HelloWorld"))
	require.Equal(t, "hello_world", toLowerSnakeCase("helloWorld"))
	require.Equal(t, "hello_world", toLowerSnakeCase("Hello_World"))
	require.Equal(t, "hello_world_123", toLowerSnakeCase("HelloWorld123"))
	require.Equal(t, "hello_world_123", toLowerSnakeCase("helloWorld123"))
}

func TestReplaceWholeWord(t *testing.T) {
	text := "Hello World! Hello Universe!"
	require.Equal(t, "Hello Gopher! Hello Universe!", replaceWholeWord(text, "World", "Gopher"))
	require.Equal(t, "Hello World! Hello Gopher!", replaceWholeWord(text, "Universe", "Gopher"))
	require.Equal(t, "Hello World! Hello Universe!", replaceWholeWord(text, "world", "Gopher"))
}

func TestParseFunctions(t *testing.T) {
	a := &Agent{
		q: &llm.Query{
			Messages: []llm.Message{{Role: "system", Content: testConfig.SystemPrompt}},
		},
		goFuncs: map[string]struct {
			fn         reflect.Value
			paramNames []string
		}{},
	}
	err := a.parseFunctions(testConfig)
	require.NoError(t, err)

	require.Equal(t, []llm.Tool{
		{
			Type: "function",
			Function: llm.Function{
				Name: "shell",
				Description: `shell runs a shell command.

Parameters:
  - command: The shell command to run.
`,
				Parameters: llm.Parameters{
					Type: "object",
					Properties: map[string]llm.Property{
						"command": {Type: "string", Description: "command"},
					},
					Required: []string{"command"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.Function{
				Name: "read_file",
				Description: `read_file reads a file.

Parameters:
  - path: The path to the file.
`,
				Parameters: llm.Parameters{
					Type: "object",
					Properties: map[string]llm.Property{
						"path": {Type: "string", Description: "path"},
					},
					Required: []string{"path"},
				},
			},
		},
	}, a.q.Tools)
}
