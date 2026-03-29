package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyaki/ralph/internal/cli"
)

func TestNewRunCommandBasicProperties(t *testing.T) {
	cmd := cli.NewRunCommand()
	if !strings.Contains(cmd.Use, "run") {
		t.Fatalf("unexpected use string: %q", cmd.Use)
	}
	if cmd.Flags().Lookup("max-iterations") == nil {
		t.Fatal("expected max-iterations flag to exist")
	}

	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help should execute successfully: %v", err)
	}
}

func TestRunCommandExecuteDebugHappyPath(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	t.Setenv("HOME", t.TempDir())
	t.Setenv("DEBUG", "1")

	binDir := t.TempDir()
	writeExecutable(t, binDir, "opencode", "#!/bin/sh\necho \"ok\"\n")
	t.Setenv("PATH", binDir)

	cmd := cli.NewRunCommand()
	cmd.SetArgs([]string{"build"})

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success in debug mode, got: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "[build]") {
		t.Errorf("expected output to contain [build], got %q", output)
	}
}

func TestRunCommandExecuteInitAsPrompt(t *testing.T) {
	// This tests that `ralph run init` treats "init" as a prompt name, not the init command
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	t.Setenv("HOME", t.TempDir())
	t.Setenv("DEBUG", "1")

	binDir := t.TempDir()
	writeExecutable(t, binDir, "opencode", "#!/bin/sh\necho \"ok\"\n")
	t.Setenv("PATH", binDir)

	cmd := cli.NewRunCommand()
	cmd.SetArgs([]string{"init"}) // "init" as prompt name

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing 'init' prompt")
	}
	if !strings.Contains(err.Error(), "prompt file not found for 'init'") {
		t.Fatalf("expected prompt not found error, got: %v", err)
	}
}

func TestRunCommandNoLogFalseFlagOverridesConfig(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	t.Setenv("HOME", t.TempDir())
	t.Setenv("DEBUG", "1")

	binaryDir := t.TempDir()
	writeExecutable(t, binaryDir, "opencode", "#!/bin/sh\necho \"ok\"\n")
	t.Setenv("PATH", binaryDir)

	if err := os.WriteFile("ralph.toml", []byte("no-log = true\n"), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	if err := os.WriteFile("prompt.txt", []byte("Just a simple prompt"), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	cmd := cli.NewRunCommand()
	cmd.SetArgs([]string{
		"--config", "ralph.toml",
		"--prompt-file", "prompt.txt",
		"--log-file", "ralph.log",
		"--no-log=false",
		"build",
	})

	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success in debug mode, got: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "ralph.log")); err != nil {
		t.Fatalf("expected log file to exist when --no-log=false is set, got error: %v", err)
	}
}

func TestRunCommandNoLogFalseFlagOverridesEnv(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	t.Setenv("HOME", t.TempDir())
	t.Setenv("DEBUG", "1")
	t.Setenv("RALPH_LOG_ENABLED", "0")

	binaryDir := t.TempDir()
	writeExecutable(t, binaryDir, "opencode", "#!/bin/sh\necho \"ok\"\n")
	t.Setenv("PATH", binaryDir)

	if err := os.WriteFile("prompt.txt", []byte("Just a simple prompt"), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	cmd := cli.NewRunCommand()
	cmd.SetArgs([]string{
		"--prompt-file", "prompt.txt",
		"--log-file", "ralph.log",
		"--no-log=false",
		"build",
	})

	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success in debug mode, got: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "ralph.log")); err != nil {
		t.Fatalf("expected log file to exist when --no-log=false overrides env, got error: %v", err)
	}
}

func TestRunCommandNoLogFlagTracksExplicitFalse(t *testing.T) {
	cmd := cli.NewRunCommand()
	if err := cmd.ParseFlags([]string{"--no-log=false"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	if !cmd.Flags().Changed("no-log") {
		t.Fatal("expected --no-log=false to mark no-log flag as changed")
	}

	value, err := cmd.Flags().GetBool("no-log")
	if err != nil {
		t.Fatalf("failed to read no-log flag value: %v", err)
	}
	if value {
		t.Fatalf("expected --no-log=false value to be false, got %v", value)
	}
}
