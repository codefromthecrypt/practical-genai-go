package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/parakeet-nest/parakeet/completion"
	"github.com/parakeet-nest/parakeet/embeddings"
	"github.com/parakeet-nest/parakeet/llm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("😡:", err)
	}

	url := "http://localhost:11434/v1"
	embeddingsModel := "mxbai-embed-large"
	model := "qwen2.5:7b"

	elasticStore := embeddings.ElasticsearchStore{}
	err = elasticStore.Initialize(
		[]string{
			os.Getenv("ELASTICSEARCH_URL"),
		},
		os.Getenv("ELASTICSEARCH_USER"),
		os.Getenv("ELASTICSEARCH_PASSWORD"),
		nil,
		"mxbai-golang-index",
	)
	if err != nil {
		log.Fatalln("😡:", err)
	}

	userContent := `Summarize what's new with benchmarks in 3 bullet points. Be succinct`

	// Create an embedding from the question
	embeddingFromQuestion, err := embeddings.CreateEmbeddingWithOpenAI(
		url,
		llm.OpenAIQuery4Embedding{
			Model: embeddingsModel,
			Input: userContent,
		},
		"question",
	)
	if err != nil {
		log.Fatalln("😡:", err)
	}
	fmt.Println("🔎 searching for similarity...")

	similarities, err := elasticStore.SearchTopNSimilarities(embeddingFromQuestion, 5)

	for _, similarity := range similarities {
		fmt.Println("📝 doc:", similarity.Id, "score:", similarity.Score)
	}

	if err != nil {
		log.Fatalln("😡:", err)
	}

	documentsContent := embeddings.GenerateContentFromSimilarities(similarities)
	fmt.Println("Context is now: ", documentsContent)

	systemContent := `You are a Golang expert.
	Using only the below provided context, answer the user's question
	to the best of your ability using only the resources provided.
	`

	queryChat := llm.OpenAIQuery{
		Model: model,
		Messages: []llm.Message{
			{Role: "system", Content: systemContent},
			{Role: "system", Content: documentsContent},
			{Role: "user", Content: userContent},
		},
	}

	fmt.Println()
	fmt.Println("🤖 answer:")

	// Answer the question
	_, err = completion.ChatWithOpenAIStream(url, queryChat,
		func(answer llm.OpenAIAnswer) error {
			fmt.Print(answer.Choices[0].Delta.Content)
			return nil
		})
	if err != nil {
		log.Fatal("😡:", err)
	}

	fmt.Println()
}
