package agent

import (
	"io"
)

// OpencodeAgent implements the Agent interface for the opencode CLI.
type OpencodeAgent struct {
	Model     string
	AgentMode string
	Env       []string
}

// Execute runs opencode with the given prompt.
func (a *OpencodeAgent) Execute(prompt string, output io.Writer) (string, error) {
	// Opencode CLI uses: opencode run [--model <model>] <prompt>
	args := []string{"run"}
	if a.Model != "" {
		args = append(args, "--model", a.Model)
	}
	if a.AgentMode != "" {
		args = append(args, "--agent", a.AgentMode)
	}
	args = append(args, prompt)

	return executeAgentCommand("opencode", args, a.Env, output, "opencode")
}

// Name returns the name of the agent.
func (a *OpencodeAgent) Name() string {
	return "opencode"
}

// IsAvailable checks if opencode is available in PATH.
func (a *OpencodeAgent) IsAvailable() bool {
	return isAgentAvailable("opencode")
}
