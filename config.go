package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
)

// Config holds all Ralph configuration
type Config struct {
	// Command-line flags
	ConfigFile             string `toml:"config-file"`
	MaxIterations          int    `toml:"max-iterations"`
	PromptFile             string `toml:"prompt-file"`
	SpecsDir               string `toml:"specs-dir"`
	SpecsIndexFile         string `toml:"specs-index-file"`
	NoSpecsIndex           bool   `toml:"no-specs-index"`
	ImplementationPlanName string `toml:"implementation-plan-name"`
	LogFile                string `toml:"log-file"`
	NoLog                  bool   `toml:"no-log"`
	LogTruncate            bool   `toml:"log-truncate"`
	CustomPrompt           string `toml:"custom-prompt"`
	PromptsDir             string `toml:"prompts-dir"`
	AgentName              string `toml:"agent"`
	Model                  string `toml:"model"`
	AgentMode              string `toml:"agent-mode"`
	// Internal state
	configLoaded bool
}

// LoadConfig loads configuration with proper precedence: flags > env vars > config file > defaults
func (c *Config) LoadConfig() error {
	// Save original environment variables
	origEnv := map[string]string{
		"RALPH_MAX_ITERATIONS":           os.Getenv("RALPH_MAX_ITERATIONS"),
		"RALPH_SPECS_DIR":                os.Getenv("RALPH_SPECS_DIR"),
		"RALPH_SPECS_INDEX_FILE":         os.Getenv("RALPH_SPECS_INDEX_FILE"),
		"RALPH_IMPLEMENTATION_PLAN_NAME": os.Getenv("RALPH_IMPLEMENTATION_PLAN_NAME"),
		"RALPH_CUSTOM_PROMPT":            os.Getenv("RALPH_CUSTOM_PROMPT"),
		"RALPH_LOG_FILE":                 os.Getenv("RALPH_LOG_FILE"),
		"RALPH_LOG_ENABLED":              os.Getenv("RALPH_LOG_ENABLED"),
		"RALPH_LOG_APPEND":               os.Getenv("RALPH_LOG_APPEND"),
		"RALPH_PROMPTS_DIR":              os.Getenv("RALPH_PROMPTS_DIR"),
		"RALPH_AGENT":                    os.Getenv("RALPH_AGENT"),
		"RALPH_MODEL":                    os.Getenv("RALPH_MODEL"),
		"RALPH_AGENT_MODE":               os.Getenv("RALPH_AGENT_MODE"),
	}

	// Load config file if specified
	configFromFile := &Config{}
	if c.ConfigFile != "" {
		configPath := c.ConfigFile
		if !filepath.IsAbs(configPath) {
			cwd, _ := os.Getwd()
			configPath = filepath.Join(cwd, configPath)
		}
		if err := c.loadConfigFile(configPath, configFromFile); err != nil {
			return fmt.Errorf("failed to load config file %s: %w", configPath, err)
		}
	} else {
		// Try default config files in current directory (in order of preference)
		cwd, _ := os.Getwd()

		// Primary: ralph.toml
		defaultConfig := filepath.Join(cwd, "ralph.toml")
		if _, err := os.Stat(defaultConfig); err == nil {
			_ = c.loadConfigFile(defaultConfig, configFromFile)
		} else {
			// Secondary: .ralphrc.toml (backward compatibility)
			oldDefault := filepath.Join(cwd, ".ralphrc.toml")
			if _, err := os.Stat(oldDefault); err == nil {
				_ = c.loadConfigFile(oldDefault, configFromFile)
			} else {
				// Tertiary: .ralphrc (old shell format, for legacy compatibility)
				legacyDefault := filepath.Join(cwd, ".ralphrc")
				if _, err := os.Stat(legacyDefault); err == nil {
					_ = c.loadConfigFile(legacyDefault, configFromFile)
				}
			}
		}
	}

	// Apply precedence: flags > original env vars > config file values > defaults

	// Max Iterations
	if c.MaxIterations == 0 { // Flag not set
		if origEnv["RALPH_MAX_ITERATIONS"] != "" {
			if val, err := strconv.Atoi(origEnv["RALPH_MAX_ITERATIONS"]); err == nil {
				c.MaxIterations = val
			}
		}
	}
	if c.MaxIterations == 0 && configFromFile.MaxIterations != 0 {
		c.MaxIterations = configFromFile.MaxIterations
	}
	if c.MaxIterations == 0 {
		c.MaxIterations = 25
	}

	// Specs Dir
	if c.SpecsDir == "" {
		if origEnv["RALPH_SPECS_DIR"] != "" {
			c.SpecsDir = origEnv["RALPH_SPECS_DIR"]
		}
	}
	if c.SpecsDir == "" && configFromFile.SpecsDir != "" {
		c.SpecsDir = configFromFile.SpecsDir
	}
	if c.SpecsDir == "" {
		c.SpecsDir = "specs"
	}

	// Specs Index File
	if c.SpecsIndexFile == "" {
		if origEnv["RALPH_SPECS_INDEX_FILE"] != "" {
			c.SpecsIndexFile = origEnv["RALPH_SPECS_INDEX_FILE"]
		}
	}
	if c.SpecsIndexFile == "" && configFromFile.SpecsIndexFile != "" {
		c.SpecsIndexFile = configFromFile.SpecsIndexFile
	}
	if c.SpecsIndexFile == "" {
		c.SpecsIndexFile = "README.md"
	}

	// Implementation Plan Name
	if c.ImplementationPlanName == "" {
		if origEnv["RALPH_IMPLEMENTATION_PLAN_NAME"] != "" {
			c.ImplementationPlanName = origEnv["RALPH_IMPLEMENTATION_PLAN_NAME"]
		}
	}
	if c.ImplementationPlanName == "" && configFromFile.ImplementationPlanName != "" {
		c.ImplementationPlanName = configFromFile.ImplementationPlanName
	}
	if c.ImplementationPlanName == "" {
		c.ImplementationPlanName = "IMPLEMENTATION_PLAN.md"
	}

	// Custom Prompt
	if c.CustomPrompt == "" {
		if origEnv["RALPH_CUSTOM_PROMPT"] != "" {
			c.CustomPrompt = origEnv["RALPH_CUSTOM_PROMPT"]
		}
	}
	if c.CustomPrompt == "" && configFromFile.CustomPrompt != "" {
		c.CustomPrompt = configFromFile.CustomPrompt
	}

	// Prompts Dir
	if c.PromptsDir == "" {
		if origEnv["RALPH_PROMPTS_DIR"] != "" {
			c.PromptsDir = origEnv["RALPH_PROMPTS_DIR"]
		}
	}
	if c.PromptsDir == "" && configFromFile.PromptsDir != "" {
		c.PromptsDir = configFromFile.PromptsDir
	}
	if c.PromptsDir == "" {
		c.PromptsDir = filepath.Join(os.Getenv("HOME"), ".ralph")
	}

	// Log File
	if c.LogFile == "" {
		if origEnv["RALPH_LOG_FILE"] != "" {
			c.LogFile = origEnv["RALPH_LOG_FILE"]
		}
	}
	if c.LogFile == "" && configFromFile.LogFile != "" {
		c.LogFile = configFromFile.LogFile
	}
	if c.LogFile == "" {
		cwd, _ := os.Getwd()
		c.LogFile = filepath.Join(cwd, "ralph.log")
	}

	// Log Enabled
	if !c.NoLog {
		if origEnv["RALPH_LOG_ENABLED"] == "0" || os.Getenv("RALPH_LOG_ENABLED") == "0" {
			c.NoLog = true
		}
	}
	if !c.NoLog && configFromFile.NoLog {
		c.NoLog = configFromFile.NoLog
	}

	// Log Truncate (append by default)
	if !c.LogTruncate {
		if origEnv["RALPH_LOG_APPEND"] == "0" || os.Getenv("RALPH_LOG_APPEND") == "0" {
			c.LogTruncate = true
		}
	}
	if !c.LogTruncate && configFromFile.LogTruncate {
		c.LogTruncate = configFromFile.LogTruncate
	}

	// Agent Name
	if c.AgentName == "" {
		if origEnv["RALPH_AGENT"] != "" {
			c.AgentName = origEnv["RALPH_AGENT"]
		}
	}
	if c.AgentName == "" && configFromFile.AgentName != "" {
		c.AgentName = configFromFile.AgentName
	}
	if c.AgentName == "" {
		c.AgentName = "opencode"
	}

	// Model
	if c.Model == "" {
		if origEnv["RALPH_MODEL"] != "" {
			c.Model = origEnv["RALPH_MODEL"]
		}
	}
	if c.Model == "" && configFromFile.Model != "" {
		c.Model = configFromFile.Model
	}
	// Note: Model is optional, so we don't set a default

	// Agent Mode
	if c.AgentMode == "" {
		if origEnv["RALPH_AGENT_MODE"] != "" {
			c.AgentMode = origEnv["RALPH_AGENT_MODE"]
		}
	}
	if c.AgentMode == "" && configFromFile.AgentMode != "" {
		c.AgentMode = configFromFile.AgentMode
	}
	// Note: Agent mode is optional, so we don't set a default

	c.configLoaded = true
	return nil
}

// loadConfigFile reads a TOML config file and populates the given Config
func (c *Config) loadConfigFile(path string, target *Config) error {
	_, err := toml.DecodeFile(path, target)
	return err
}
