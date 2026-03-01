package main

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

// OpencodeAgent implements the Agent interface for the opencode CLI
type OpencodeAgent struct {
	Model string
}

// Execute runs opencode with the given prompt
func (a *OpencodeAgent) Execute(prompt string, output io.Writer) (string, error) {
	// Opencode CLI uses: opencode run [--model <model>] <prompt>
	args := []string{"run"}
	if a.Model != "" {
		args = append(args, "--model", a.Model)
	}
	args = append(args, prompt)
	cmd := exec.Command("opencode", args...)

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
		return result, fmt.Errorf("opencode execution failed: %w", err)
	}

	return result, nil
}

// Name returns the name of the agent
func (a *OpencodeAgent) Name() string {
	return "opencode"
}

// IsAvailable checks if opencode is available in PATH
func (a *OpencodeAgent) IsAvailable() bool {
	_, err := exec.LookPath("opencode")
	return err == nil
}
