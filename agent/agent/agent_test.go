package agent

import (
	"reflect"
	"testing"

	"github.com/parakeet-nest/parakeet/llm"
	"github.com/stretchr/testify/require"
)

func TestNewAgent(t *testing.T) {
	agent, err := New("http://localhost:8080", "test-model", testConfig)
	require.NoError(t, err)
	require.NotNil(t, agent)

	require.Equal(t, "http://localhost:8080", agent.ollamaURL)
	require.Equal(t, &llm.Query{
		Model:    "test-model",
		Messages: []llm.Message{{Role: "system", Content: testConfig.SystemPrompt}},
		Options:  llm.Options{},
		Tools: []llm.Tool{
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
					Name: "patch_file",
					Description: `patch_file patches the file at the specified path by replacing before with
after.

Parameters:
  - path: The path to the file, in the format "path/to/file.txt"
  - before: The content that will be replaced
  - after: The content it will be replaced with
`,
					Parameters: llm.Parameters{
						Type: "object",
						Properties: map[string]llm.Property{
							"path":   {Type: "string", Description: "path"},
							"before": {Type: "string", Description: "before"},
							"after":  {Type: "string", Description: "after"},
						},
						Required: []string{"path", "before", "after"},
					},
				},
			},
		},
	}, agent.q)
	require.Equal(t, map[string]struct {
		fn         reflect.Value
		paramNames []string
	}{
		"patch_file": {fn: reflect.ValueOf(PatchFile), paramNames: []string{"path", "before", "after"}},
		"shell":      {fn: reflect.ValueOf(Shell), paramNames: []string{"command"}},
	}, agent.goFuncs)
}

func TestCallFunction(t *testing.T) {
	agent, err := New("http://localhost:8080", "test-model", testConfig)
	require.NoError(t, err)

	result := agent.callFunction(llm.FunctionTool{
		Name: "shell",
		Arguments: map[string]interface{}{
			"command": "echo Hello, World!",
		},
	})
	require.Equal(t, "hello world", result)

	t.Run("not found", func(t *testing.T) {
		result := agent.callFunction(llm.FunctionTool{
			Name: "shell2",
			Arguments: map[string]interface{}{
				"command": "echo Hello, World!",
			},
		})
		require.Equal(t, "shell2 is not a registered tool", result)
	})
}
