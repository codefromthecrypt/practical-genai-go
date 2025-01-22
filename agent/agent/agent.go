// Package agent includes a naive agent which can be used to understand the
// mechanics of LLM agents with Ollama and Parakeet.
package agent

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"regexp"
	"strings"
	"unicode"

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

// New creates a new agent that will use Ollama with a specific model for
// requests (Agent.Request).
//
// toolSource must have the following.
//   - package godoc with the system prompt which overviews all functions, and
//     can give hints on which tools to use for a purpose.
//   - godoc on each exported function, written in a way an LLM can understand.
//     For example, you need to describe the parameters and the output.
//
// As there's no way to lookup functions by name in golang, we pass a fns map
// of lower_snake_case function name to reflect.Value of the exported function
// that exists in toolsSource. All functions must return string and an error
// result.
func New(url, model string, toolSource string, fns map[string]reflect.Value) (*Agent, error) {
	a := &Agent{
		ollamaURL: url,
		q: &llm.Query{
			Model: model,
		},
		goFuncs: map[string]struct {
			fn         reflect.Value
			paramNames []string
		}{},
	}
	return a, a.parseFunctions(toolSource, fns)
}

// parseFunctions initializes the agent with metdata the LLM uses to call
// tools, as well fields used to ensure parameters are passed in the correct
// order when invoking.
func (a *Agent) parseFunctions(toolSource string, fns map[string]reflect.Value) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", toolSource, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// take the system prompt from tools.go package doc
	if node.Doc == nil {
		return fmt.Errorf("missing package godoc on tool source file")
	}
	a.q.Messages = []llm.Message{
		{Role: "system", Content: node.Doc.Text()},
	}

	for _, decl := range node.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || !fd.Name.IsExported() {
			continue
		}

		tName := fd.Name.Name
		toolName := toLowerSnakeCase(tName)

		fn, exists := fns[toolName]
		if !exists {
			continue
		}

		doc := ""
		if fd.Doc != nil {
			doc = fd.Doc.Text()
		}

		params := llm.Parameters{
			Type:       "object",
			Properties: make(map[string]llm.Property),
			Required:   []string{},
		}

		for _, field := range fd.Type.Params.List {
			if len(field.Names) != 1 {
				continue
			}
			pName := field.Names[0].Name
			paramName := toLowerSnakeCase(pName)

			doc = replaceWholeWord(doc, pName, paramName)

			params.Properties[paramName] = llm.Property{
				Type:        fmt.Sprintf("%v", field.Type),
				Description: paramName,
			}
			params.Required = append(params.Required, paramName)
		}

		doc = replaceWholeWord(doc, tName, toolName)

		a.q.Tools = append(a.q.Tools, llm.Tool{
			Type: "function",
			Function: llm.Function{
				Name:        toolName,
				Description: doc,
				Parameters:  params,
			},
		})

		a.goFuncs[toolName] = struct {
			fn         reflect.Value
			paramNames []string
		}{
			fn:         fn,
			paramNames: params.Required,
		}
	}

	if len(a.q.Tools) != len(fns) {
		return fmt.Errorf("missing tools")
	}

	return nil
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

// toLowerSnakeCase converts a string to snake_case.
func toLowerSnakeCase(s string) string {
	if len(s) == 0 {
		return s
	}
	var result strings.Builder
	for i, r := range s {
		if i > 0 && (unicode.IsUpper(r) || unicode.IsNumber(r)) && (!unicode.IsUpper(rune(s[i-1])) &&
			!unicode.IsDigit(rune(s[i-1]))) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// replaceWholeWord replaces all occurrences of a whole word in the text with the replacement.
func replaceWholeWord(text, old, new string) string {
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(old) + `\b`)
	return re.ReplaceAllString(text, new)
}

// callFunction invokes the tool call, taking care to order parameters
// identified by name in the correct order.
//
// LLMs can understand problems, and attempt to resolve them. Hence, we encode the error into a string instead of returning two values.
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
