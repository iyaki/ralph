// Package executor runs external commands for Ralph components.
package executor

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

// ExecuteCommand executes a command and returns its output.
func ExecuteCommand(command string, args []string, output io.Writer) (string, error) {
	cmd := exec.Command(command, args...) // #nosec G204 -- command and args are intentionally caller-provided

	// Create buffers to capture stdout and stderr
	var outBuf, errBuf bytes.Buffer

	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	// Run the command
	err := cmd.Run()

	// Combine stdout and stderr for result
	result := outBuf.String() + errBuf.String()
	_, _ = io.WriteString(output, result)

	if err != nil {
		// Return the output even on error (non-fatal)
		return result, fmt.Errorf("command execution failed: %w", err)
	}

	return result, nil
}
