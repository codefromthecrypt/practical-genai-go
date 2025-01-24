package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/parakeet-nest/parakeet/completion"
	"github.com/parakeet-nest/parakeet/llm"
)

// llama-server --log-disable --hf-repo Qwen/Qwen2.5-7B-Instruct-GGUF --hf-file qwen2.5-7b-instruct-q4_k_m.gguf
var llamaServer = struct {
	url, model string
}{"http://localhost:8080/v1", "ignored"}

// ollama serve; ollama pull qwen2.5:14b
var ollama = struct {
	url, model string
}{"http://localhost:11434/v1", "qwen2.5:14b"}

func main() {
	url := ollama.url
	model := ollama.model

	systemContent := `You are a Golang expert.
	Using only the below provided context, answer the user's question
	to the best of your ability using only the resources provided.
	`

	contextContent := `<context>
		<doc>

### New benchmark function

Benchmarks may now use the faster and less error-prone [testing.B.Loop](/pkg/testing#B.Loop) method to perform benchmark iterations like for b.Loop() { ... } in place of the typical loop structures involving b.N like for range b.N. This offers two significant advantages:
- The benchmark function will execute exactly once per -count, so expensive setup and cleanup steps execute only once.
- Function call parameters and results are kept alive, preventing the compiler from fully optimizing away the loop body.

#### [testing](/pkg/testing/)

The new [T.Context](/pkg/testing#T.Context) and [B.Context](/pkg/testing#B.Context) methods return a context that's canceled
after the test completes and before test cleanup functions run.

<!-- testing.B.Loop mentioned in 6-stdlib/6-testing-bloop.md. -->

The new [T.Chdir](/pkg/testing#T.Chdir) and [B.Chdir](/pkg/testing#B.Chdir) methods can be used to change the working
directory for the duration of a test or benchmark.

		</doc>
	</context>`

	userContent := `Summarize what's new with benchmarks in 3 bullet points. Be succinct`

	query := llm.OpenAIQuery{
		Model: model,
		Messages: []llm.Message{
			{Role: "system", Content: systemContent},
			{Role: "system", Content: contextContent},
			{Role: "user", Content: userContent},
		},
		Stream: false,
	}

	// Answer the question
	_, err := completion.ChatWithOpenAIStream(url, query,
		func(answer llm.OpenAIAnswer) error {
			fmt.Print(answer.Choices[0].Delta.Content)
			return nil
		})
	if err != nil {
		log.Fatal("😡:", err)
	}

	fmt.Println()
}
