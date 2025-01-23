package main

import (
	"fmt"
	"log"

	"github.com/codefromthecrypt/practical-genai-go/agent/agent"
	"github.com/codefromthecrypt/practical-genai-go/agent/dev"
)

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
	reply, err = a.Request("Add a thank you to GopherCon Singapore to the " +
		"bottom of that file as a new section. Write it in Singlish.")
	if err != nil {
		log.Fatal("😡:", err)
	}
	fmt.Println(reply)
	fmt.Println()
}
