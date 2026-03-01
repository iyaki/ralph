package main

import (
"fmt"
"os"
"os/exec"
"path/filepath"
"time"
)

// Logger handles logging to file
type Logger struct {
file    *os.File
enabled bool
}

// NewLogger creates a new logger based on configuration
func NewLogger(cfg *Config) (*Logger, error) {
logger := &Logger{
enabled: !cfg.NoLog,
}

// Check environment variable for log enabled
if logEnabled := os.Getenv("RALPH_LOG_ENABLED"); logEnabled == "0" {
logger.enabled = false
}

if !logger.enabled {
return logger, nil
}

// Determine log file path
logFile := cfg.LogFile
if logFile == "" {
// Create temporary log file
tmpFile, err := os.CreateTemp("", "ralph-*.log")
if err != nil {
return nil, fmt.Errorf("failed to create temp log file: %w", err)
}
logFile = tmpFile.Name()
tmpFile.Close()
}

// Create log directory if it doesn't exist
logDir := filepath.Dir(logFile)
if err := os.MkdirAll(logDir, 0755); err != nil {
return nil, fmt.Errorf("failed to create log directory: %w", err)
}

// Determine open mode
var file *os.File
var err error

logAppend := true
if cfg.LogTruncate {
logAppend = false
}
if logAppendEnv := os.Getenv("RALPH_LOG_APPEND"); logAppendEnv == "0" {
logAppend = false
}

if logAppend {
file, err = os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
} else {
file, err = os.Create(logFile)
}

if err != nil {
return nil, fmt.Errorf("failed to open log file %s: %w", logFile, err)
}

logger.file = file

// Write log header
fmt.Fprintf(file, "===== Ralph run started at %s =====\n", time.Now().Format("2006-01-02 15:04:05 -0700"))

// Add git info if available
if gitBranch := getGitBranch(); gitBranch != "" {
fmt.Fprintf(file, "Git branch: %s\n", gitBranch)
}
if gitCommit := getGitCommit(); gitCommit != "" {
fmt.Fprintf(file, "Git commit: %s\n", gitCommit)
}

fmt.Fprintf(file, "===== Ralph run started at %s =====\n", time.Now().Format("2006-01-02 15:04:05 -0700"))

return logger, nil
}

// Close closes the logger
func (l *Logger) Close() error {
if l.file != nil {
return l.file.Close()
}
return nil
}

// getGitBranch returns the current git branch name
func getGitBranch() string {
cmd := exec.Command("git", "symbolic-ref", "HEAD")
output, err := cmd.Output()
if err != nil {
return "N/A"
}
branch := string(output)
// Remove "refs/heads/" prefix
if len(branch) > 11 {
branch = branch[11:]
}
// Trim newline
if len(branch) > 0 && branch[len(branch)-1] == '\n' {
branch = branch[:len(branch)-1]
}
return branch
}

// getGitCommit returns the current git commit hash
func getGitCommit() string {
cmd := exec.Command("git", "rev-parse", "HEAD")
output, err := cmd.Output()
if err != nil {
return "N/A"
}
commit := string(output)
// Trim newline
if len(commit) > 0 && commit[len(commit)-1] == '\n' {
commit = commit[:len(commit)-1]
}
return commit
}
