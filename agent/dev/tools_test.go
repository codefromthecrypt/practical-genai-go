package dev

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var logBuffer bytes.Buffer

func init() {
	log.SetOutput(&logBuffer)
}

func TestShell(t *testing.T) {
	logBuffer.Reset()
	output, err := Shell("echo Hello, World!")
	require.NoError(t, err)
	expected := "Hello, World!\n"
	require.Equal(t, expected, output)

	logContent := logBuffer.String()
	require.Contains(t, logContent, "Shell Command:\n```bash\necho Hello, World!\n```")
	require.Contains(t, logContent, "Command succeeded:\nHello, World!\n")
}

func TestReadFile(t *testing.T) {
	logBuffer.Reset()
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.txt")
	content := "Hello, World!"
	err := os.WriteFile(filePath, []byte(content), 0o644)
	require.NoError(t, err)

	out, err := ReadFile(filePath)
	require.NoError(t, err)
	expected := "```plaintext\nHello, World!\n```"
	require.Equal(t, expected, out)

	logContent := logBuffer.String()
	require.Contains(t, logContent, "Reading file: "+filePath)
	require.Contains(t, logContent, "Successfully read file:\n```plaintext\nHello, World!\n```")
}

func TestWriteFile(t *testing.T) {
	logBuffer.Reset()
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.txt")
	content := "Hello, World!"

	out, err := WriteFile(filePath, content)
	require.NoError(t, err)
	require.Equal(t, "Successfully wrote to "+filePath, out)

	writtenContent, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, content, string(writtenContent))

	logContent := logBuffer.String()
	require.Contains(t, logContent, "Writing file: "+filePath)
	require.Contains(t, logContent, "```plaintext\nHello, World!\n```")
}

func TestPatchFile(t *testing.T) {
	logBuffer.Reset()
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.txt")
	initialContent := "Hello, World!"
	err := os.WriteFile(filePath, []byte(initialContent), 0o644)
	require.NoError(t, err)

	before := "World"
	after := "Gopher"
	out, err := PatchFile(filePath, before, after)
	require.NoError(t, err)
	require.Contains(t, "Successfully replaced before with after.", out)

	patchedContent, err := os.ReadFile(filePath)
	require.NoError(t, err)
	expected := "Hello, Gopher!"
	require.Equal(t, expected, string(patchedContent))

	logContent := logBuffer.String()
	require.Contains(t, logContent, "Patching file: "+filePath)
	require.Contains(t, logContent, "```plaintext\nWorld\n```\n->\n```plaintext\nGopher\n```")
}
