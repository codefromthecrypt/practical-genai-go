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

	"github.com/parakeet-nest/parakeet/llm"
)

// parseFunctions initializes the agent with metadata the LLM uses to call
// tools, as well fields used to ensure parameters are passed in the correct
// order when invoking.
func (a *Agent) parseFunctions(config *Config) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", config.ToolSource, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	for _, decl := range node.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || !fd.Name.IsExported() {
			continue
		}

		tName := fd.Name.Name
		toolName := toLowerSnakeCase(tName)

		fn, exists := config.Tools[toolName]
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
			for _, name := range field.Names {
				pName := name.Name
				paramName := toLowerSnakeCase(pName)

				doc = replaceWholeWord(doc, pName, paramName)

				params.Properties[paramName] = llm.Property{
					Type:        fmt.Sprintf("%v", field.Type),
					Description: paramName,
				}
				params.Required = append(params.Required, paramName)
			}
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

	if len(a.q.Tools) != len(config.Tools) {
		return fmt.Errorf("missing tools")
	}

	return nil
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
