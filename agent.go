package main

import "io"

// Agent represents an AI agent CLI that can execute prompts
type Agent interface {
// Execute runs the agent with the given prompt and returns the output
Execute(prompt string, output io.Writer) (string, error)

// Name returns the name of the agent
Name() string

// IsAvailable checks if the agent CLI is available on the system
IsAvailable() bool
}

// GetAgent returns the appropriate agent based on configuration
func GetAgent(agentName, model string) Agent {
	switch agentName {
	case "claude":
		return &ClaudeAgent{Model: model}
	case "opencode":
		return &OpencodeAgent{Model: model}
	default:
		// Default to opencode for backward compatibility
		return &OpencodeAgent{Model: model}
	}
}
