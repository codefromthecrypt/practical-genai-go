package main

import (
	_ "embed"
	"fmt"
	"github.com/codefromthecrypt/practical-genai-go/agent/agent"
	"github.com/codefromthecrypt/practical-genai-go/agent/tools"
	"log"
	"reflect"
)

//go:embed tools/tools.go
var toolSource string

func main() {
	url := "http://localhost:11434"
	model := "qwen2.5:7b"

	// Initialize the agent and give it access to certain functions.
	a, err := agent.New(url, model, toolSource, map[string]reflect.Value{
		"shell":      reflect.ValueOf(tools.Shell),
		"read_file":  reflect.ValueOf(tools.ReadFile),
		"write_file": reflect.ValueOf(tools.WriteFile),
	})
	if err != nil {
		log.Panicln("😡:", err)
	}

	// Ask the agent to do something that requires poking around the machine.
	// This could be solved multiple ways given the functions we've allowed.
	reply, err := a.Request(
		"Write a file in the current directory named README.md. " +
			"In that, make a heading called 'Parakeet examples' which includes a markdown link to and short description of each top-level directory.")
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
