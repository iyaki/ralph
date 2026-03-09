package agent

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

func executeAgentCommand(command string, args []string, output io.Writer, errPrefix string) (string, error) {
	cmd := exec.Command(command, args...) // #nosec G204 -- command and args are controlled by internal agent integrations

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer

	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	result := outBuf.String() + errBuf.String()
	_, _ = io.WriteString(output, result)

	if err != nil {
		return result, fmt.Errorf("%s execution failed: %w", errPrefix, err)
	}

	return result, nil
}

func isAgentAvailable(command string) bool {
	_, err := exec.LookPath(command)

	return err == nil
}
