package prompt

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyaki/ralph/internal/config"
)

func TestBuildPromptIncludesConfiguredReferences(t *testing.T) {
	cfg := &config.Config{
		SpecsDir:               "specs",
		SpecsIndexFile:         "README.md",
		ImplementationPlanName: "PLAN.md",
	}
	p := BuildPrompt(cfg)
	if !strings.Contains(p, "specs/README.md") {
		t.Fatalf("expected specs index reference, got %q", p)
	}
	if !strings.Contains(p, "PLAN.md") {
		t.Fatalf("expected implementation plan name, got %q", p)
	}
}

func TestPlanPromptIncludesScopeAndPlanName(t *testing.T) {
	cfg := &config.Config{ImplementationPlanName: "PLAN.md", SpecsDir: "specs"}
	p := PlanPrompt(cfg, "API")
	if !strings.Contains(p, "Scope: API") {
		t.Fatalf("expected scope in prompt, got %q", p)
	}
	if !strings.Contains(p, "PLAN.md") {
		t.Fatalf("expected plan name in prompt, got %q", p)
	}
}

func TestFindFileUpwardsAbsoluteAndRelative(t *testing.T) {
	dir := t.TempDir()
	abs := filepath.Join(dir, "x.md")
	if err := os.WriteFile(abs, []byte("x"), 0644); err != nil {
		t.Fatalf("failed writing absolute file: %v", err)
	}
	if got := findFileUpwards(abs); got != abs {
		t.Fatalf("expected absolute path %q, got %q", abs, got)
	}

	nested := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("failed creating nested dir: %v", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if err := os.Chdir(nested); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	rel := filepath.Join("prompts", "plan.md")
	if err := os.MkdirAll(filepath.Join(dir, "prompts"), 0755); err != nil {
		t.Fatalf("failed creating prompts dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, rel), []byte("content"), 0644); err != nil {
		t.Fatalf("failed writing relative target: %v", err)
	}

	got := findFileUpwards(rel)
	if got != filepath.Join(dir, rel) {
		t.Fatalf("expected %q, got %q", filepath.Join(dir, rel), got)
	}

	if findFileUpwards("missing.md") != "" {
		t.Fatal("expected empty string for missing file")
	}
}

func TestGetPromptBranches(t *testing.T) {
	t.Run("custom prompt", func(t *testing.T) {
		cfg := &config.Config{CustomPrompt: "inline custom"}
		var out bytes.Buffer
		p, err := GetPrompt(cfg, "build", "scope", &out)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p != "inline custom" {
			t.Fatalf("expected custom prompt, got %q", p)
		}
	})

	t.Run("stdin prompt", func(t *testing.T) {
		cfg := &config.Config{PromptFile: "-"}
		var out bytes.Buffer

		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("pipe creation failed: %v", err)
		}
		if _, err := w.Write([]byte("from-stdin")); err != nil {
			t.Fatalf("pipe write failed: %v", err)
		}
		w.Close()

		oldStdin := os.Stdin
		os.Stdin = r
		t.Cleanup(func() {
			os.Stdin = oldStdin
		})

		p, err := GetPrompt(cfg, "build", "scope", &out)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p != "from-stdin" {
			t.Fatalf("expected stdin content, got %q", p)
		}
	})

	t.Run("prompt file", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "custom.md")
		if err := os.WriteFile(file, []byte("from-file"), 0644); err != nil {
			t.Fatalf("failed writing prompt file: %v", err)
		}

		cfg := &config.Config{PromptFile: file}
		p, err := GetPrompt(cfg, "build", "scope", &bytes.Buffer{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p != "from-file" {
			t.Fatalf("expected file content, got %q", p)
		}
	})

	t.Run("prompt from prompts dir", func(t *testing.T) {
		dir := t.TempDir()
		promptsDir := filepath.Join(dir, "prompts")
		if err := os.MkdirAll(promptsDir, 0755); err != nil {
			t.Fatalf("failed to create prompts dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(promptsDir, "custom.md"), []byte("from-prompts-dir"), 0644); err != nil {
			t.Fatalf("failed to write prompt: %v", err)
		}

		cfg := &config.Config{PromptsDir: promptsDir}
		p, err := GetPrompt(cfg, "custom", "scope", &bytes.Buffer{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p != "from-prompts-dir" {
			t.Fatalf("expected prompts dir content, got %q", p)
		}
	})

	t.Run("default build and plan", func(t *testing.T) {
		cfg := &config.Config{SpecsDir: "specs", SpecsIndexFile: "README.md", ImplementationPlanName: "PLAN.md"}

		buildPrompt, err := GetPrompt(cfg, "build", "scope", &bytes.Buffer{})
		if err != nil {
			t.Fatalf("unexpected error for build: %v", err)
		}
		if !strings.Contains(buildPrompt, "Agent Instructions (Build Mode)") {
			t.Fatalf("unexpected build prompt: %q", buildPrompt)
		}

		planPrompt, err := GetPrompt(cfg, "plan", "My Scope", &bytes.Buffer{})
		if err != nil {
			t.Fatalf("unexpected error for plan: %v", err)
		}
		if !strings.Contains(planPrompt, "Scope: My Scope") {
			t.Fatalf("unexpected plan prompt: %q", planPrompt)
		}
	})

	t.Run("unknown prompt returns error", func(t *testing.T) {
		cfg := &config.Config{PromptsDir: t.TempDir()}
		_, err := GetPrompt(cfg, "unknown", "scope", &bytes.Buffer{})
		if err == nil {
			t.Fatal("expected error for unknown prompt")
		}
	})
}
