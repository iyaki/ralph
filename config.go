package main

import (
"bufio"
"fmt"
"os"
"path/filepath"
"strconv"
"strings"
)

// Config holds all Ralph configuration
type Config struct {
// Command-line flags
ConfigFile             string
MaxIterations          int
PromptFile             string
SpecsDir               string
SpecsIndexFile         string
NoSpecsIndex           bool
ImplementationPlanName string
LogFile                string
NoLog                  bool
LogTruncate            bool
CustomPrompt           string
PromptsDir             string
AgentName              string
Model                  string
// Internal state
configLoaded bool
}

// LoadConfig loads configuration with proper precedence: flags > env vars > config file > defaults
func (c *Config) LoadConfig() error {
// Save original environment variables
origEnv := map[string]string{
"RALPH_MAX_ITERATIONS":            os.Getenv("RALPH_MAX_ITERATIONS"),
"RALPH_SPECS_DIR":                 os.Getenv("RALPH_SPECS_DIR"),
"RALPH_SPECS_INDEX_FILE":          os.Getenv("RALPH_SPECS_INDEX_FILE"),
"RALPH_IMPLEMENTATION_PLAN_NAME":  os.Getenv("RALPH_IMPLEMENTATION_PLAN_NAME"),
"RALPH_CUSTOM_PROMPT":             os.Getenv("RALPH_CUSTOM_PROMPT"),
"RALPH_LOG_FILE":                  os.Getenv("RALPH_LOG_FILE"),
"RALPH_LOG_ENABLED":               os.Getenv("RALPH_LOG_ENABLED"),
"RALPH_LOG_APPEND":                os.Getenv("RALPH_LOG_APPEND"),
"RALPH_PROMPTS_DIR":               os.Getenv("RALPH_PROMPTS_DIR"),
"RALPH_AGENT":                     os.Getenv("RALPH_AGENT"),
"RALPH_MODEL":                     os.Getenv("RALPH_MODEL"),
}

// Load config file if specified
if c.ConfigFile != "" {
configPath := c.ConfigFile
if !filepath.IsAbs(configPath) {
cwd, _ := os.Getwd()
configPath = filepath.Join(cwd, configPath)
}
if err := c.loadConfigFile(configPath); err != nil {
return fmt.Errorf("failed to load config file %s: %w", configPath, err)
}
} else {
// Try default .ralphrc in current directory
cwd, _ := os.Getwd()
defaultConfig := filepath.Join(cwd, ".ralphrc")
if _, err := os.Stat(defaultConfig); err == nil {
_ = c.loadConfigFile(defaultConfig)
}
}

// Apply precedence: flags > original env vars > config file env vars > defaults

// Max Iterations
if c.MaxIterations == 0 { // Flag not set
if origEnv["RALPH_MAX_ITERATIONS"] != "" {
if val, err := strconv.Atoi(origEnv["RALPH_MAX_ITERATIONS"]); err == nil {
c.MaxIterations = val
}
}
}
if c.MaxIterations == 0 {
if val := os.Getenv("RALPH_MAX_ITERATIONS"); val != "" {
if intVal, err := strconv.Atoi(val); err == nil {
c.MaxIterations = intVal
}
}
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
if c.SpecsDir == "" {
c.SpecsDir = getEnvWithDefault("RALPH_SPECS_DIR", "specs")
}

// Specs Index File
if c.SpecsIndexFile == "" {
if origEnv["RALPH_SPECS_INDEX_FILE"] != "" {
c.SpecsIndexFile = origEnv["RALPH_SPECS_INDEX_FILE"]
}
}
if c.SpecsIndexFile == "" {
c.SpecsIndexFile = getEnvWithDefault("RALPH_SPECS_INDEX_FILE", "README.md")
}

// Implementation Plan Name
if c.ImplementationPlanName == "" {
if origEnv["RALPH_IMPLEMENTATION_PLAN_NAME"] != "" {
c.ImplementationPlanName = origEnv["RALPH_IMPLEMENTATION_PLAN_NAME"]
}
}
if c.ImplementationPlanName == "" {
c.ImplementationPlanName = getEnvWithDefault("RALPH_IMPLEMENTATION_PLAN_NAME", "IMPLEMENTATION_PLAN.md")
}

// Custom Prompt
if c.CustomPrompt == "" {
if origEnv["RALPH_CUSTOM_PROMPT"] != "" {
c.CustomPrompt = origEnv["RALPH_CUSTOM_PROMPT"]
}
}
if c.CustomPrompt == "" {
c.CustomPrompt = os.Getenv("RALPH_CUSTOM_PROMPT")
}

// Prompts Dir
if c.PromptsDir == "" {
if origEnv["RALPH_PROMPTS_DIR"] != "" {
c.PromptsDir = origEnv["RALPH_PROMPTS_DIR"]
}
}
if c.PromptsDir == "" {
c.PromptsDir = getEnvWithDefault("RALPH_PROMPTS_DIR", filepath.Join(os.Getenv("HOME"), ".ralph"))
}

// Log File
if c.LogFile == "" {
if origEnv["RALPH_LOG_FILE"] != "" {
c.LogFile = origEnv["RALPH_LOG_FILE"]
}
}
if c.LogFile == "" {
cwd, _ := os.Getwd()
c.LogFile = getEnvWithDefault("RALPH_LOG_FILE", filepath.Join(cwd, "ralph.log"))
}

// Log Enabled
if !c.NoLog {
if origEnv["RALPH_LOG_ENABLED"] == "0" || os.Getenv("RALPH_LOG_ENABLED") == "0" {
c.NoLog = true
}
}

// Log Truncate (append by default)
if !c.LogTruncate {
if origEnv["RALPH_LOG_APPEND"] == "0" || os.Getenv("RALPH_LOG_APPEND") == "0" {
c.LogTruncate = true
}
}

// Agent Name
if c.AgentName == "" {
if origEnv["RALPH_AGENT"] != "" {
c.AgentName = origEnv["RALPH_AGENT"]
}
}
if c.AgentName == "" {
c.AgentName = getEnvWithDefault("RALPH_AGENT", "opencode")
}

// Model
if c.Model == "" {
if origEnv["RALPH_MODEL"] != "" {
c.Model = origEnv["RALPH_MODEL"]
}
}
if c.Model == "" {
c.Model = getEnvWithDefault("RALPH_MODEL", "")
}
// Note: Model is optional, so we don't set a default

c.configLoaded = true
return nil
}

// loadConfigFile reads a shell-style config file and sets environment variables
func (c *Config) loadConfigFile(path string) error {
file, err := os.Open(path)
if err != nil {
return err
}
defer file.Close()

scanner := bufio.NewScanner(file)
for scanner.Scan() {
line := strings.TrimSpace(scanner.Text())

// Skip empty lines and comments
if line == "" || strings.HasPrefix(line, "#") {
continue
}

// Parse KEY=VALUE or export KEY=VALUE
line = strings.TrimPrefix(line, "export ")
parts := strings.SplitN(line, "=", 2)
if len(parts) != 2 {
continue
}

key := strings.TrimSpace(parts[0])
value := strings.TrimSpace(parts[1])

// Remove quotes if present
value = strings.Trim(value, "\"'")

// Set environment variable (only if not already set by original env)
if os.Getenv(key) == "" {
os.Setenv(key, value)
}
}

return scanner.Err()
}

// getEnvWithDefault returns environment variable value or default
func getEnvWithDefault(key, defaultValue string) string {
if val := os.Getenv(key); val != "" {
return val
}
return defaultValue
}
