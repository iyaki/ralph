package main

import (
"bytes"
"fmt"
"io"
"os/exec"
)

// ClaudeAgent implements the Agent interface for the claude CLI
type ClaudeAgent struct{}

// Execute runs claude with the given prompt
func (a *ClaudeAgent) Execute(prompt string, output io.Writer) (string, error) {
	// Claude Code CLI uses: claude --dangerously-skip-permissions <prompt>
	cmd := exec.Command("claude", "--dangerously-skip-permissions", prompt)

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
return result, fmt.Errorf("claude execution failed: %w", err)
}

return result, nil
}

// Name returns the name of the agent
func (a *ClaudeAgent) Name() string {
return "claude"
}

// IsAvailable checks if claude is available in PATH
func (a *ClaudeAgent) IsAvailable() bool {
_, err := exec.LookPath("claude")
return err == nil
}
