package agent

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

// CursorAgent implements the Agent interface for the cursor CLI
type CursorAgent struct {
	Model     string
	AgentMode string
}

// Execute runs cursor with the given prompt
func (a *CursorAgent) Execute(prompt string, output io.Writer) (string, error) {
	// Cursor CLI uses: cursor [--model <model>] <prompt>
	args := []string{}
	if a.Model != "" {
		args = append(args, "--model", a.Model)
	}
	args = append(args, prompt)
	cmd := exec.Command("cursor", args...)

	// Create buffers to capture stdout and stderr
	var outBuf, errBuf bytes.Buffer

	// Use MultiWriter to write to both buffer and output
	cmd.Stdout = io.MultiWriter(&outBuf, output)
	cmd.Stderr = io.MultiWriter(&errBuf, output)

	// Run the command
	err := cmd.Run()

	// Combine stdout and stderr for result
	result := outBuf.String() + errBuf.String()

	if err != nil {
		return result, fmt.Errorf("cursor execution failed: %w", err)
	}

	return result, nil
}

// Name returns the name of the agent
func (a *CursorAgent) Name() string {
	return "cursor"
}

// IsAvailable checks if cursor is available in PATH
func (a *CursorAgent) IsAvailable() bool {
	_, err := exec.LookPath("cursor")
	return err == nil
}
