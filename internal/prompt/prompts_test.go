package prompt_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyaki/ralph/internal/config"
	"github.com/iyaki/ralph/internal/prompt"
)

func TestBuildPromptIncludesConfiguredReferences(t *testing.T) {
	cfg := &config.Config{
		SpecsDir:               "specs",
		SpecsIndexFile:         "README.md",
		ImplementationPlanName: "PLAN.md",
	}
	p := prompt.BuildPrompt(cfg)
	if !strings.Contains(p, "specs/README.md") {
		t.Fatalf("expected specs index reference, got %q", p)
	}
	if !strings.Contains(p, "PLAN.md") {
		t.Fatalf("expected implementation plan name, got %q", p)
	}
}

func TestPlanPromptIncludesScopeAndPlanName(t *testing.T) {
	cfg := &config.Config{ImplementationPlanName: "PLAN.md", SpecsDir: "specs"}
	p := prompt.PlanPrompt(cfg, "API")
	if !strings.Contains(p, "Scope: API") {
		t.Fatalf("expected scope in prompt, got %q", p)
	}
	if !strings.Contains(p, "PLAN.md") {
		t.Fatalf("expected plan name in prompt, got %q", p)
	}
}

func TestGetPromptCustomPrompt(t *testing.T) {
	cfg := &config.Config{CustomPrompt: "inline custom"}
	var out bytes.Buffer
	p, err := prompt.GetPrompt(cfg, "build", "scope", &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != "inline custom" {
		t.Fatalf("expected custom prompt, got %q", p)
	}
}

func TestGetPromptFromStdin(t *testing.T) {
	cfg := &config.Config{PromptFile: "-"}
	var out bytes.Buffer

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe creation failed: %v", err)
	}
	if _, err := w.Write([]byte("from-stdin")); err != nil {
		t.Fatalf("pipe write failed: %v", err)
	}
	_ = w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = oldStdin
	})

	p, err := prompt.GetPrompt(cfg, "build", "scope", &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != "from-stdin" {
		t.Fatalf("expected stdin content, got %q", p)
	}
}

func TestGetPromptFromFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "custom.md")
	if err := os.WriteFile(file, []byte("from-file"), 0o644); err != nil {
		t.Fatalf("failed writing prompt file: %v", err)
	}

	cfg := &config.Config{PromptFile: file}
	p, err := prompt.GetPrompt(cfg, "build", "scope", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != "from-file" {
		t.Fatalf("expected file content, got %q", p)
	}
}

func TestGetPromptFromPromptsDir(t *testing.T) {
	dir := t.TempDir()
	promptsDir := filepath.Join(dir, "prompts")
	if err := os.MkdirAll(promptsDir, 0o755); err != nil {
		t.Fatalf("failed to create prompts dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(promptsDir, "custom.md"), []byte("from-prompts-dir"), 0o644); err != nil {
		t.Fatalf("failed to write prompt: %v", err)
	}

	cfg := &config.Config{PromptsDir: promptsDir}
	p, err := prompt.GetPrompt(cfg, "custom", "scope", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != "from-prompts-dir" {
		t.Fatalf("expected prompts dir content, got %q", p)
	}
}

func TestGetPromptDefaultBuildAndPlan(t *testing.T) {
	cfg := &config.Config{SpecsDir: "specs", SpecsIndexFile: "README.md", ImplementationPlanName: "PLAN.md"}

	buildPrompt, err := prompt.GetPrompt(cfg, "build", "scope", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error for build: %v", err)
	}
	if !strings.Contains(buildPrompt, "Agent Instructions (Build Mode)") {
		t.Fatalf("unexpected build prompt: %q", buildPrompt)
	}

	planPrompt, err := prompt.GetPrompt(cfg, "plan", "My Scope", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error for plan: %v", err)
	}
	if !strings.Contains(planPrompt, "Scope: My Scope") {
		t.Fatalf("unexpected plan prompt: %q", planPrompt)
	}
}

func TestGetPromptUnknownReturnsError(t *testing.T) {
	cfg := &config.Config{PromptsDir: t.TempDir()}
	_, err := prompt.GetPrompt(cfg, "unknown", "scope", &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for unknown prompt")
	}
}
