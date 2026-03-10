// Package config handles loading and resolving Ralph configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
)

const (
	defaultMaxIterations          = 25
	defaultSpecsDir               = "specs"
	defaultSpecsIndexFile         = "README.md"
	defaultImplementationPlanName = "IMPLEMENTATION_PLAN.md"
	defaultAgentName              = "opencode"
)

type envValues struct {
	maxIterations          string
	specsDir               string
	specsIndexFile         string
	implementationPlanName string
	customPrompt           string
	logFile                string
	logEnabled             string
	logAppend              string
	promptsDir             string
	agentName              string
	model                  string
	agentMode              string
}

// Config holds all Ralph configuration.
type Config struct {
	ConfigFile             string                          `toml:"config-file"`
	MaxIterations          int                             `toml:"max-iterations"`
	PromptFile             string                          `toml:"prompt-file"`
	SpecsDir               string                          `toml:"specs-dir"`
	SpecsIndexFile         string                          `toml:"specs-index-file"`
	NoSpecsIndex           bool                            `toml:"no-specs-index"`
	ImplementationPlanName string                          `toml:"implementation-plan-name"`
	LogFile                string                          `toml:"log-file"`
	NoLog                  bool                            `toml:"no-log"`
	LogTruncate            bool                            `toml:"log-truncate"`
	CustomPrompt           string                          `toml:"custom-prompt"`
	PromptsDir             string                          `toml:"prompts-dir"`
	AgentName              string                          `toml:"agent"`
	Model                  string                          `toml:"model"`
	AgentMode              string                          `toml:"agent-mode"`
	PromptOverrides        map[string]PromptConfigOverride `toml:"prompt-overrides"`

	configLoaded bool
}

// PromptConfigOverride defines per-prompt configuration overrides.
type PromptConfigOverride struct {
	Model     string `toml:"model"`
	AgentMode string `toml:"agent-mode"`
}

// LoadConfig loads configuration with proper precedence: flags > env vars > config file > defaults.
func (c *Config) LoadConfig() error {
	configFromFile, err := c.resolveFileConfig()
	if err != nil {
		return err
	}

	env := readEnv()
	c.applyConfigValues(configFromFile, env)
	c.configLoaded = true

	return nil
}

func (c *Config) resolveFileConfig() (*Config, error) {
	configFromFile := &Config{}

	// Priority 1: Config file path from flag (c.ConfigFile is already set by flag parsing)
	configPath := c.ConfigFile

	// Priority 2: Config file path from environment variable
	if configPath == "" {
		configPath = os.Getenv("RALPH_CONFIG")
	}

	if configPath != "" {
		if !filepath.IsAbs(configPath) {
			cwd, _ := os.Getwd()
			configPath = filepath.Join(cwd, configPath)
		}

		if err := c.loadConfigFile(configPath, configFromFile); err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", configPath, err)
		}

		return configFromFile, nil
	}

	loadDefaultConfig(c, configFromFile)

	return configFromFile, nil
}

func loadDefaultConfig(c *Config, target *Config) {
	cwd, _ := os.Getwd()
	for _, name := range []string{"ralph.toml", ".ralphrc.toml", ".ralphrc"} {
		path := filepath.Join(cwd, name)
		if _, err := os.Stat(path); err == nil {
			_ = c.loadConfigFile(path, target)

			return
		}
	}
}

func readEnv() envValues {
	return envValues{
		maxIterations:          os.Getenv("RALPH_MAX_ITERATIONS"),
		specsDir:               os.Getenv("RALPH_SPECS_DIR"),
		specsIndexFile:         os.Getenv("RALPH_SPECS_INDEX_FILE"),
		implementationPlanName: os.Getenv("RALPH_IMPLEMENTATION_PLAN_NAME"),
		customPrompt:           os.Getenv("RALPH_CUSTOM_PROMPT"),
		logFile:                os.Getenv("RALPH_LOG_FILE"),
		logEnabled:             os.Getenv("RALPH_LOG_ENABLED"),
		logAppend:              os.Getenv("RALPH_LOG_APPEND"),
		promptsDir:             os.Getenv("RALPH_PROMPTS_DIR"),
		agentName:              os.Getenv("RALPH_AGENT"),
		model:                  os.Getenv("RALPH_MODEL"),
		agentMode:              os.Getenv("RALPH_AGENT_MODE"),
	}
}

func (c *Config) applyConfigValues(fileCfg *Config, env envValues) {
	c.MaxIterations = resolveInt(c.MaxIterations, env.maxIterations, fileCfg.MaxIterations, defaultMaxIterations)
	c.SpecsDir = resolveString(c.SpecsDir, env.specsDir, fileCfg.SpecsDir, defaultSpecsDir)
	c.SpecsIndexFile = resolveString(c.SpecsIndexFile, env.specsIndexFile, fileCfg.SpecsIndexFile, defaultSpecsIndexFile)
	c.ImplementationPlanName = resolveString(
		c.ImplementationPlanName,
		env.implementationPlanName,
		fileCfg.ImplementationPlanName,
		defaultImplementationPlanName,
	)
	c.CustomPrompt = resolveString(c.CustomPrompt, env.customPrompt, fileCfg.CustomPrompt, "")
	c.PromptsDir = resolveString(c.PromptsDir, env.promptsDir, fileCfg.PromptsDir, defaultPromptsDir())
	c.LogFile = resolveString(c.LogFile, env.logFile, fileCfg.LogFile, defaultLogFile())
	c.NoLog = resolveBool(c.NoLog, env.logEnabled, fileCfg.NoLog, true)
	c.LogTruncate = resolveBool(c.LogTruncate, env.logAppend, fileCfg.LogTruncate, true)
	c.AgentName = resolveString(c.AgentName, env.agentName, fileCfg.AgentName, defaultAgentName)
	c.Model = resolveString(c.Model, env.model, fileCfg.Model, "")
	c.AgentMode = resolveString(c.AgentMode, env.agentMode, fileCfg.AgentMode, "")

	// Prompt overrides only come from the config file.
	if len(fileCfg.PromptOverrides) > 0 {
		c.PromptOverrides = fileCfg.PromptOverrides
	}
}

func resolveInt(flagValue int, envValue string, fileValue int, defaultValue int) int {
	if flagValue != 0 {
		return flagValue
	}

	if envValue != "" {
		if parsed, err := strconv.Atoi(envValue); err == nil {
			return parsed
		}
	}

	if fileValue != 0 {
		return fileValue
	}

	return defaultValue
}

func resolveString(flagValue, envValue, fileValue, defaultValue string) string {
	if flagValue != "" {
		return flagValue
	}

	if envValue != "" {
		return envValue
	}

	if fileValue != "" {
		return fileValue
	}

	return defaultValue
}

func resolveBool(flagValue bool, envValue string, fileValue bool, envDisableIsZero bool) bool {
	if flagValue {
		return true
	}

	if envDisableIsZero && envValue == "0" {
		return true
	}

	if fileValue {
		return true
	}

	return false
}

func defaultPromptsDir() string {
	return filepath.Join(os.Getenv("HOME"), ".ralph")
}

func defaultLogFile() string {
	cwd, _ := os.Getwd()

	return filepath.Join(cwd, "ralph.log")
}

// loadConfigFile reads a TOML config file and populates the given Config.
func (c *Config) loadConfigFile(path string, target *Config) error {
	_, err := toml.DecodeFile(path, target)

	return err
}
