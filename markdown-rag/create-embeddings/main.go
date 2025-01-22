package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/parakeet-nest/parakeet/content"
	"github.com/parakeet-nest/parakeet/embeddings"
	"github.com/parakeet-nest/parakeet/llm"
)

//go:embed go1.24.md
var goReleaseNotes string

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("😡:", err)
	}

	url := "http://localhost:11434/v1"
	embeddingsModel := "mxbai-embed-large"

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

	chunks := content.ParseMarkdown(goReleaseNotes)

	// Create embeddings from documents and save them in the store
	for idx, doc := range chunks {
		fmt.Println("📝 Creating embedding from document ", idx)
		embedding, err := embeddings.CreateEmbeddingWithOpenAI(
			url,
			llm.OpenAIQuery4Embedding{
				Model: embeddingsModel,
				Input: fmt.Sprintf("## %s\n\n%s\n\n", doc.Header, doc.Content),
			},
			strconv.Itoa(idx),
		)
		if err != nil {
			log.Fatalln("😡:", err)
		}

		_, err = elasticStore.Save(embedding)
		if _, err = elasticStore.Save(embedding); err != nil {
			log.Fatalln("😡:", err)
		}
		fmt.Println("Document", embedding.Id, "indexed successfully")
	}
}
