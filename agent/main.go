package main

import (
	"fmt"
	"log"

	"github.com/codefromthecrypt/practical-genai-go/agent/agent"
	"github.com/codefromthecrypt/practical-genai-go/agent/dev"
)

// main shows an agent can perform tasks for you, including figuring out which
// tools to use. Notably, an agent isn't just retrieving information, it is
// performing operations so you don't have to.
//
// This agent was written to be easy to explain in a conference, so it doesn't
// do anything fancy. Real agents are more than just an LLM, a good system
// prompt and a few tools.
//
// For example, https://github.com/block/goose includes a plugin system,
// context summarization, robust LLM handling, and work in progress towards
// Model Context Protocol (MCP), which decouples tool, prompt and resource
// definitions from the agents that use them.
func main() {
	url := "http://localhost:11434"
	model := "qwen2.5:7b"

	// Initialize the agent and give it access to certain functions.
	a, err := agent.New(url, model, dev.AgentConfig)
	if err != nil {
		log.Panicln("😡:", err)
	}

	// Ask the agent to do something that requires poking around the machine.
	// This could be solved multiple ways given the functions we've allowed.
	reply, err := a.Request(
		"Analyze each top-level directory in the current working directory." +
			"Make a new file named README.md which describes each under the " +
			"heading 'Parakeet examples'.")
	if err != nil {
		log.Fatal("😡:", err)
	}
	fmt.Println(reply)
	fmt.Println()

	// Since the agent is stateful, it will remember the last thing it did. It
	// can revise or do something related to it without restating context.
	reply, err = a.Request("Append a thank you to GopherCon Singapore to the " +
		"bottom of that file, in Singlish.")
	if err != nil {
		log.Fatal("😡:", err)
	}
	fmt.Println(reply)
	fmt.Println()
}
