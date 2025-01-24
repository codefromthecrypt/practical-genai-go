# Practical GenAI with Go demo code

These are examples for my [Practical GenAI with Go][talk] talk, which uses
[Parakeet][parakeet] to access [Ollama][ollama] both via its native and
OpenAI-compatible endpoints.

Note: Many examples here are the same as those in the [Parakeet repository][parakeet-examples].

## Setup

All the demos use [Ollama][ollama], so start that before running any.

```bash
ollama serve
```

Then, in another tab, pull the models. We use a medium size chat model to reduce hallucinations:

```bash
ollama pull qwen2.5:14b
ollama pull mxbai-embed-large
```

## Agent

[agent](chat/main.go) writes a new file READMUAH.md for you.

```bash
go run agent/main.go
```

## Chat

[chat](chat/main.go) completes a chat, then a follow-up message.

```bash
go run chat/main.go 
```

## Markdown Context

[markdown-context](markdown-context/main.go) adds markdown into the system
context. This allows the LLM to consider new information when generating a
response.

```bash
go run markdown-context/main.go 
```

## Markdown RAG

[markdown-rag](markdown-rag) uses a VectorDB, Elasticsearch to store markdown
fragments. Later, those similar to the user's prompt are retrieved into the
system context. This allows the LLM to consider new information when generating a response. 

### Setup

First you need to have Elasticsearch running. You can use docker to start it: `docker compose up -d`. When done, stop it via `docker compose down`.

### Create embeddings

[create-embeddings](markdown-rag/create-embeddings/main.go) loads [go1.24.md](markdown-rag/create-embeddings/go1.24.md) and stores the embeddings in Elasticsearch.

```bash
go run markdown-rag/create-embeddings/main.go
```

### Use Embeddings

[use-embeddings](markdown-rag/use-embeddings/main.go)
does a similarity search in Elasticsearch based on the user's prompt to get relevant text fragments. Then, it places those in context to complete the prompt.

```bash
go run markdown-rag/use-embeddings/main.go
```

---
[talk]: https://speakerdeck.com/adriancole/practical-genai-with-go-gophercon-singapore
[ollama]: https://github.com/ollama/ollama
[parakeet]: https://github.com/parakeet-nest/parakeet
[parakeet-examples]: https://github.com/parakeet-nest/parakeet/tree/main/examples