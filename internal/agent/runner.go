package agent

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

func executeAgentCommand(command string, args []string, output io.Writer, errPrefix string) (string, error) {
	cmd := exec.Command(command, args...) // #nosec G204 -- command and args are controlled by internal agent integrations

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	if output == nil {
		output = io.Discard
	}
	streamOutput := &synchronizedWriter{w: output}

	cmd.Stdout = io.MultiWriter(&outBuf, streamOutput)
	cmd.Stderr = io.MultiWriter(&errBuf, streamOutput)

	err := cmd.Run()
	result := outBuf.String() + errBuf.String()

	if err != nil {
		return result, fmt.Errorf("%s execution failed: %w", errPrefix, err)
	}

	return result, nil
}

func isAgentAvailable(command string) bool {
	_, err := exec.LookPath(command)

	return err == nil
}
