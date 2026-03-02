package executor

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

// ExecuteCommand executes a command and returns its output
func ExecuteCommand(command string, args []string, output io.Writer) (string, error) {
	cmd := exec.Command(command, args...)

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
		// Return the output even on error (non-fatal)
		return result, fmt.Errorf("command execution failed: %w", err)
	}

	return result, nil
}
