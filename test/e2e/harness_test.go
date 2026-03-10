// Package e2e provides end-to-end testing infrastructure for the Ralph CLI.
package e2e_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var (
	ralphPath string
	agentPath string
)

func TestMain(m *testing.M) {
	// 1. Create temporary directory for test binaries
	tmpDir, err := os.MkdirTemp("", "ralph-e2e-bins-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			fmt.Fprintf(os.Stderr, "failed to remove temp dir: %v\n", err)
		}
	}()

	// 2. Build ralph binary
	ralphPath = filepath.Join(tmpDir, "ralph")
	if err := buildBinary("../../cmd/ralph", ralphPath); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build ralph: %v\n", err)
		os.Exit(1)
	}

	// 3. Build test agent binary
	agentPath = filepath.Join(tmpDir, "ralph-test-agent")
	if err := buildBinary("./agents/ralph-test-agent", agentPath); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build test agent: %v\n", err)
		os.Exit(1)
	}

	// Create symlink for "opencode" to point to our test agent so ralph picks it up by default
	if err := os.Symlink(agentPath, filepath.Join(tmpDir, "opencode")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to symlink opencode to test agent: %v\n", err)
		os.Exit(1)
	}

	// 4. Run tests
	os.Exit(m.Run())
}

func buildBinary(srcPath, destPath string) error {
	cmd := exec.Command("go", "build", "-o", destPath, srcPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// runTestCase executes a single E2E test scenario.
func runTestCase(t *testing.T, tc TestCase) {
	t.Helper()

	workDir := prepareTestEnv(t, tc)
	res := executeRalph(t, workDir, tc)
	verifyResult(t, workDir, tc, res)
}

func prepareTestEnv(t *testing.T, tc TestCase) string {
	t.Helper()
	workDir := t.TempDir()

	for name, content := range tc.Files {
		path := filepath.Join(workDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("failed to create directory for fixture %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write fixture %s: %v", name, err)
		}
	}

	return workDir
}

func executeRalph(t *testing.T, workDir string, tc TestCase) RunResult {
	t.Helper()

	cmd := exec.Command(ralphPath, tc.Args...)
	cmd.Dir = workDir

	env := os.Environ()
	agentDir := filepath.Dir(agentPath)
	pathEnv := fmt.Sprintf("PATH=%s%c%s", agentDir, os.PathListSeparator, os.Getenv("PATH"))
	env = append(env, pathEnv)

	for k, v := range tc.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	if tc.Stdin != "" {
		cmd.Stdin = strings.NewReader(tc.Stdin)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run ralph: %v", err)
		}
	}

	return RunResult{
		ExitCode:   exitCode,
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		DurationMs: duration.Milliseconds(),
	}
}

func verifyResult(t *testing.T, workDir string, tc TestCase, res RunResult) {
	t.Helper()

	if res.ExitCode != tc.ExpectedExitCode {
		t.Errorf("exit code mismatch: got %d, want %d", res.ExitCode, tc.ExpectedExitCode)
		t.Logf("stdout:\n%s", res.Stdout)
		t.Logf("stderr:\n%s", res.Stderr)
	}

	for _, want := range tc.ExpectedStdoutContains {
		if !strings.Contains(res.Stdout, want) {
			t.Errorf("stdout missing expected content: %q", want)
		}
	}

	for _, want := range tc.ExpectedStderrContains {
		if !strings.Contains(res.Stderr, want) {
			t.Errorf("stderr missing expected content: %q", want)
		}
	}

	verifyForbidden(t, tc, res)
	verifyFiles(t, workDir, tc)

	t.Logf("Test duration: %dms", res.DurationMs)
}

func verifyForbidden(t *testing.T, tc TestCase, res RunResult) {
	t.Helper()
	for _, forbidden := range tc.ForbiddenOutput {
		if strings.Contains(res.Stdout, forbidden) {
			t.Errorf("stdout contains forbidden content: %q", forbidden)
		}
		if strings.Contains(res.Stderr, forbidden) {
			t.Errorf("stderr contains forbidden content: %q", forbidden)
		}
	}
}

func verifyFiles(t *testing.T, workDir string, tc TestCase) {
	t.Helper()
	for _, filename := range tc.ExpectedFiles {
		path := filepath.Join(workDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file missing: %s", filename)
		}
	}

	for filename, content := range tc.ExpectedFileContent {
		path := filepath.Join(workDir, filename)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("failed to read expected file %s: %v", filename, err)

			continue
		}

		fileContent := string(data)
		for _, substr := range content {
			if !strings.Contains(fileContent, substr) {
				t.Errorf("file %s missing expected content: %q", filename, substr)
			}
		}
	}
}
