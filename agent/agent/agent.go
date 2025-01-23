// Package agent includes a naive agent which can be used to understand the
// mechanics of LLM agents with Ollama and Parakeet.
package agent

import (
	"fmt"
	"reflect"

	"github.com/parakeet-nest/parakeet/completion"
	"github.com/parakeet-nest/parakeet/llm"
)

// Agent is an LLM combined with tools (go functions) it can use to solve a
// problem. Technically, the LLM response requests to invoke tools, but it is
// the agent that actually does it. In other words, there is no RPC connection
// to the LLM.
type Agent struct {
	// ollamaURL is the endpoint of Ollama. This isn't yet ported to OpenAI, so only works with Ollama.
	ollamaURL string
	// q includes the message history, which is naively managed. It is never
	// summarized or truncated.
	q *llm.Query
	// goFuncs are the allowed functions that the LLM can request us to invoke.
	goFuncs map[string]struct {
		fn         reflect.Value
		paramNames []string
	}
}

type Config struct {
	// SystemPrompt overviews all functions, and can give hints on which tools
	// to use for a purpose.
	SystemPrompt string
	// ToolSource must have godoc on each exported function, written in a way
	// an LLM can understand. For example, you need to describe the parameters
	// and the output.
	ToolSource string
	// Tools are a map of lower_snake_case function name to reflect.Value of
	// the Go function representing the tool. All functions must be defined in
	// ToolSource and return string and an error  result.
	Tools map[string]reflect.Value
}

// New creates a new agent that will use Ollama with a specific model for
// requests (Agent.Request).
func New(url, model string, config *Config) (*Agent, error) {
	a := &Agent{
		ollamaURL: url,
		q: &llm.Query{
			Model:    model,
			Messages: []llm.Message{{Role: "system", Content: config.SystemPrompt}},
		},
		goFuncs: map[string]struct {
			fn         reflect.Value
			paramNames []string
		}{},
	}
	return a, a.parseFunctions(config)
}

// Request a task for the agent to perform. The result will only use tools if
// the LLM determines it necessary. For example, if the message asks a question
// that can be answered without side effects, it won't likely use tools.
func (a *Agent) Request(message string) (string, error) {
	a.q.Messages = append(a.q.Messages, llm.Message{Role: "user", Content: message})

	// Ask the agent to solve our request goal
	answer, err := completion.Chat(a.ollamaURL, *a.q)
	if err != nil {
		return "", fmt.Errorf("failed to get chat response: %w", err)
	}

	// Loop until the agent is done asking to invoke tools. When certain LLMs
	// hallucinate, they accidentally write tools into Message.Content, instead
	// of Message.ToolCalls. Handling this is tricky in real Agent frameworks.
	// Tool hallucination happens, but is less frequent in large models.
	for len(answer.Message.ToolCalls) == 1 {
		toolCall := answer.Message.ToolCalls[0]

		// A tool call may fail, but the assistant may still be able to resolve
		// it. That's why we don't handle errors like usual.
		result := a.callFunction(toolCall.Function)

		a.q.Messages = append(a.q.Messages,
			answer.Message,
			llm.Message{Role: "tool", Content: fmt.Sprintf("%v", result)},
		)

		if answer, err = completion.Chat(a.ollamaURL, *a.q); err != nil {
			return "", fmt.Errorf("failed to get chat response after tool call: %w", err)
		}
	}

	a.q.Messages = append(a.q.Messages, answer.Message)
	return answer.Message.Content, nil
}

// callFunction invokes the tool call, taking care to order parameters
// identified by name in the correct order.
//
// LLMs can understand problems, and attempt to resolve them. Hence, we encode
// the error into a string instead of returning two values.
func (a *Agent) callFunction(toolCall llm.FunctionTool) string {
	fn, ok := a.goFuncs[toolCall.Name]
	if !ok {
		return toolCall.Name + " is not a registered tool"
	}

	// Get the type of the function
	funcType := fn.fn.Type()

	// Prepare a slice to hold the arguments in the correct order
	args := make([]reflect.Value, funcType.NumIn())

	// Iterate over the parameters and set them in the correct order
	for i := 0; i < funcType.NumIn(); i++ {
		paramName := fn.paramNames[i]
		if value, ok := toolCall.Arguments[paramName]; ok {
			args[i] = reflect.ValueOf(value)
		} else {
			return fmt.Sprintf("Missing parameter: %s", paramName)
		}
	}

	// Invoke the function with the ordered arguments
	results := fn.fn.Call(args)

	// Extract and handle the results
	if len(results) != 2 {
		return fmt.Sprintf("unexpected number of results")
	}

	// Extract the success message
	result := results[0].Interface().(string)

	// Extract the error
	if !results[1].IsNil() {
		err, _ := results[1].Interface().(error)
		return fmt.Sprintf("%v\n\nError:\n%v", result, err)
	}
	return result
}
