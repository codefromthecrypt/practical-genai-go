package main

import (
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

	question := "Answer in up to 3 words: Which ocean contains Bouvet Island?"
	q := llm.OpenAIQuery{
		Model:    model,
		Messages: []llm.Message{{Role: "user", Content: question}},
	}

	answer, err := completion.ChatWithOpenAI(url, q)
	if err != nil {
		log.Fatal("😡:", err)
	}
	response := answer.Choices[0].Message
	fmt.Println("Question:", question)
	fmt.Println("Answer:", response.Content)

	fmt.Println()

	secondQuestion := "What’s the capital?"
	q.Messages = append(q.Messages,
		llm.Message{Role: response.Role, Content: response.Content},
		llm.Message{Role: "user", Content: secondQuestion},
	)
	answer, err = completion.ChatWithOpenAI(url, q)
	if err != nil {
		log.Fatal("😡:", err)
	}
	response = answer.Choices[0].Message
	fmt.Println("Follow-up Question:", secondQuestion)
	fmt.Println("Answer:", response.Content)

	fmt.Println()
}
