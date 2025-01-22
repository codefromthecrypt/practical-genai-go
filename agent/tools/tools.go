// Your role is a developer agent. You build software and solve problems by
// editing files and running commands on the shell.
//
// You can use the shell tool to run any command that would work on the
// relevant operating system.
//
// You are an expert with ripgrep - `rg`. When you need to locate content in
// the code base, use `rg` exclusively. It will respect ignored files for
// efficiency.
//
// # To locate files by name, use
//
// ```bash
// rg --files | rg example.py
// ```
//
// To locate content inside files, use
// ```bash
// rg 'class Example'
// ```
//
// If you need to edit files, use either the write_file tool or the patch tool.
// Make sure to read existing content before attempting to edit.
//
// The write file tool will do a full overwrite of the existing file, while the
// patch tool will edit it using a find and replace. Choose the tool which will
// make the edit as simple as possible to execute.
//
// # Instructions
//
// You'll receive a summary and a plan, and can immediately start using your
// tools and can directly reply to the user as needed.
package tools

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// This source copies and ports a few aspects from block/goose, which is a
// robust system agent written in Python and Rust, by some pretty awesome
// people. https://github.com/square/goose

// getLanguage determines the language type from the file path.
func getLanguage(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".go":
		return "go"
	default:
		return "plaintext"
	}
}

// Shell executes a command on the shell.
//
// This will return the output and error concatenated into a single string, as
// you would see from running on the command line. There will also be an
// indication of if the command succeeded or failed.
//
// Parameters:
//   - command: The Shell command to run. It can support multiline
//     statements, if you need to run more than one at a time.
func Shell(command string) (string, error) {
	log.Printf("Shell Command:\n```bash\n%s\n```", command)
	output, err := exec.Command("sh", "-c", command).CombinedOutput()
	if err != nil {
		log.Printf("Command failed: %s", err)
		return "", fmt.Errorf("command failed: %w", err)
	}
	log.Printf("Command succeeded:\n%s", string(output))
	return string(output), nil
}

// ReadFile reads the content of the file at path.
//
// Parameters:
//   - path: The path to the file, in the format "path/to/file.txt"
func ReadFile(path string) (string, error) {
	log.Printf("Reading file: %s", path)

	expandedPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to expand path: %w", err)
	}

	content, err := os.ReadFile(expandedPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	language := getLanguage(path)
	md := fmt.Sprintf("```%s\n%s\n```", language, string(content))
	log.Printf("Successfully read file:\n%s\n", md)

	return md, nil
}

// WriteFile writes a file at the specified path with the provided content.
// This will create any directories if they do not exist. The content will
// fully overwrite the existing file.
//
// Parameters:
//   - path: The destination file path, in the format "path/to/file.txt"
//   - content: The raw file content.
func WriteFile(path string, content string) (string, error) {
	log.Printf("Writing file: %s", path)

	// Get the programming language for syntax highlighting in logs
	language := getLanguage(path)
	md := fmt.Sprintf("```%s\n%s\n```", language, content)

	// Log the content that will be written to the file
	log.Println(md)

	// Prepare the path and create any necessary parent directories
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the content to the file
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote to %s", path), nil
}
