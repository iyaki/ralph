// Package executor runs external commands for Ralph components.
package executor

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

type synchronizedWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (w *synchronizedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.w.Write(p)
}

// ExecuteCommand executes a command and returns its output.
func ExecuteCommand(command string, args []string, output io.Writer) (string, error) {
	cmd := exec.Command(command, args...) // #nosec G204 -- command and args are intentionally caller-provided

	// Create buffers to capture stdout and stderr
	var outBuf, errBuf bytes.Buffer
	if output == nil {
		output = io.Discard
	}
	streamOutput := &synchronizedWriter{w: output}

	cmd.Stdout = io.MultiWriter(&outBuf, streamOutput)
	cmd.Stderr = io.MultiWriter(&errBuf, streamOutput)

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
