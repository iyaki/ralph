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
}

// Find and load config file
configFile := c.ConfigFile
if configFile == "" {
configFile = getEnvWithDefault("RALPH_CONFIG_FILE", ".ralphrc")
}

configPath := findFileUpwards(configFile)
if configPath != "" {
if err := c.loadConfigFile(configPath); err != nil {
return fmt.Errorf("failed to load config file %s: %w", configPath, err)
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
if c.NoSpecsIndex {
c.SpecsIndexFile = ""
} else if c.SpecsIndexFile == "" {
if origEnv["RALPH_SPECS_INDEX_FILE"] != "" {
c.SpecsIndexFile = origEnv["RALPH_SPECS_INDEX_FILE"]
}
}
if c.SpecsIndexFile == "" && !c.NoSpecsIndex {
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

// Log File
if c.LogFile == "" {
if origEnv["RALPH_LOG_FILE"] != "" {
c.LogFile = origEnv["RALPH_LOG_FILE"]
}
}
if c.LogFile == "" {
c.LogFile = os.Getenv("RALPH_LOG_FILE")
}

// Prompts Dir
if c.PromptsDir == "" {
c.PromptsDir = getEnvWithDefault("RALPH_PROMPTS_DIR", "prompts")
}

c.configLoaded = true
return nil
}

// loadConfigFile sources a config file (simple key=value format)
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

// Parse key=value (simple shell variable format)
parts := strings.SplitN(line, "=", 2)
if len(parts) != 2 {
continue
}

key := strings.TrimSpace(parts[0])
value := strings.TrimSpace(parts[1])

// Remove quotes if present
value = strings.Trim(value, "\"'")

// Set environment variable
os.Setenv(key, value)
}

return scanner.Err()
}

// findFileUpwards searches for a file recursively upwards from current directory
func findFileUpwards(filename string) string {
// If it's an absolute path, return as-is
if filepath.IsAbs(filename) {
if _, err := os.Stat(filename); err == nil {
return filename
}
return ""
}

// Get current working directory
currentDir, err := os.Getwd()
if err != nil {
return ""
}

// Search upwards
for {
testPath := filepath.Join(currentDir, filename)
if _, err := os.Stat(testPath); err == nil {
return testPath
}

// Move up one directory
parentDir := filepath.Dir(currentDir)
if parentDir == currentDir {
// Reached root
break
}
currentDir = parentDir
}

return ""
}

// getEnvWithDefault returns environment variable value or default
func getEnvWithDefault(key, defaultValue string) string {
if val := os.Getenv(key); val != "" {
return val
}
return defaultValue
}
