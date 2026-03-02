package prompt

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/iyaki/ralph/internal/config"
)

// GetPrompt returns the prompt to use based on configuration and arguments
func GetPrompt(cfg *config.Config, promptName, scope string, output io.Writer) (string, error) {
	// If custom prompt is provided via flag, use it directly
	if cfg.CustomPrompt != "" {
		fmt.Fprintln(output, "")
		fmt.Fprintln(output, "===============================================================")
		fmt.Fprintln(output, "               USING INLINE CUSTOM PROMPT")
		fmt.Fprintln(output, "===============================================================")
		fmt.Fprintln(output, "")
		return cfg.CustomPrompt, nil
	}

	// Check if prompt should come from stdin
	if cfg.PromptFile == "-" || promptName == "-" {
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		fmt.Fprintln(output, "")
		fmt.Fprintln(output, "===============================================================")
		fmt.Fprintln(output, "               USING PROMPT FROM STDIN")
		fmt.Fprintln(output, "===============================================================")
		fmt.Fprintln(output, "")
		return string(content), nil
	}

	// Check if prompt file is specified
	if cfg.PromptFile != "" {
		content, err := os.ReadFile(cfg.PromptFile)
		if err != nil {
			return "", fmt.Errorf("failed to read prompt file %s: %w", cfg.PromptFile, err)
		}
		fmt.Fprintln(output, "")
		fmt.Fprintln(output, "===============================================================")
		fmt.Fprintf(output, " USING PROMPT FILE: %s\n", cfg.PromptFile)
		fmt.Fprintln(output, "===============================================================")
		fmt.Fprintln(output, "")
		return string(content), nil
	}

	// Try to find prompt file in prompts directory
	promptFilePath := filepath.Join(cfg.PromptsDir, promptName+".md")
	foundPath := findFileUpwards(promptFilePath)

	if foundPath != "" {
		content, err := os.ReadFile(foundPath)
		if err != nil {
			return "", fmt.Errorf("failed to read prompt file %s: %w", foundPath, err)
		}
		fmt.Fprintln(output, "")
		fmt.Fprintln(output, "===============================================================")
		fmt.Fprintf(output, " USING PROMPT FILE: %s\n", foundPath)
		fmt.Fprintln(output, "===============================================================")
		fmt.Fprintln(output, "")
		return string(content), nil
	}

	// Use pre-bundled prompts
	switch promptName {
	case "build":
		fmt.Fprintln(output, "")
		fmt.Fprintln(output, "===============================================================")
		fmt.Fprintln(output, "               USING DEFAULT 'BUILD' PROMPT")
		fmt.Fprintln(output, "===============================================================")
		fmt.Fprintln(output, "")
		return BuildPrompt(cfg), nil
	case "plan":
		fmt.Fprintln(output, "")
		fmt.Fprintln(output, "===============================================================")
		fmt.Fprintln(output, "               USING DEFAULT 'PLAN' PROMPT")
		fmt.Fprintln(output, "===============================================================")
		fmt.Fprintln(output, "")
		return PlanPrompt(cfg, scope), nil
	default:
		return "", fmt.Errorf("prompt file not found for '%s'. Use a valid prompt file or one of the pre-bundled prompts (build, plan)", promptName)
	}
}

// BuildPrompt generates the default build prompt
func BuildPrompt(cfg *config.Config) string {
	specsIndexFileReference := ""
	if cfg.SpecsIndexFile != "" {
		specsIndexFileReference = filepath.Join(cfg.SpecsDir, cfg.SpecsIndexFile)
	}

	specsIndexFileReferenceText := ""
	if specsIndexFileReference != "" {
		specsIndexFileReferenceText = fmt.Sprintf(" (including `%s` and related specs)", specsIndexFileReference)
	}

	prompt := "# Agent Instructions (Build Mode)\n\n"
	prompt += fmt.Sprintf("- Study `%s/*`%s.\n", cfg.SpecsDir, specsIndexFileReferenceText)
	prompt += fmt.Sprintf("- Study `%s` and pick the single most important task.\n", cfg.ImplementationPlanName)
	prompt += "- Implement the task\n"
	prompt += "- Validate the implementation\n"
	prompt += "- Update the plan\n"
	prompt += "- Commit the changes\n"
	prompt += "- Stop after the commit\n\n"

	prompt += "## Stop Condition\n\n"
	prompt += "- After completing the selected task, stop. Do NOT start another task in the same run.\n"
	prompt += "- If ALL stories are complete and passing, reply with:\n"
	prompt += "  `<COMPLETION_SIGNAL>`\n\n"

	prompt += "## IMPORTANT\n\n"
	prompt += "- Before changes, search the codebase. Do NOT assume functionality is missing.\n"
	prompt += "- Implement ONLY one task. Stop after committing.\n"
	prompt += fmt.Sprintf("- Update `%s` when the task is done.\n", cfg.ImplementationPlanName)
	prompt += "- Use the verification log format: `YYYY-MM-DD: <command or URL> - <result>`.\n"
	prompt += "- Keep a `Manual Deployment Tasks` section in implementation the plan and use `None` when there are no tasks.\n"
	prompt += fmt.Sprintf("- You may implement missing functionality if required, but study relevant `%s/*` first.\n", cfg.SpecsDir)
	prompt += "- You may add temporary logging as needed and remove if no longer needed.\n\n"

	return prompt
}

// PlanPrompt generates the default plan prompt
func PlanPrompt(cfg *config.Config, scope string) string {
	prompt := "# Agent Instructions (Planning Mode)\n\n"
	prompt += fmt.Sprintf("Scope: %s\n\n", scope)

	prompt += "## Objective\n\n"
	prompt += fmt.Sprintf("Generate or update `%s` in a structured, phase-based format with:\n\n", cfg.ImplementationPlanName)
	prompt += "- Clear status metadata\n"
	prompt += "- Quick reference tables\n"
	prompt += "- Phase sections with paths and checklists\n"
	prompt += "- Verification log entries\n"
	prompt += "- Summary tables and remaining effort\n\n"
	prompt += "Plan only. Do NOT implement anything.\n\n"

	prompt += "## Study and Gap Analysis\n\n"
	prompt += fmt.Sprintf("- Study `%s/*` to learn application requirements.\n", cfg.SpecsDir)
	prompt += fmt.Sprintf("- Study `%s` (if present; it may be incorrect).\n", cfg.ImplementationPlanName)
	prompt += "- Study relevant source code to compare against specs.\n"
	prompt += "- Use `git` to study recent changes on the specs related to the specified current scope.\n\n"

	prompt += "Rules:\n\n"
	prompt += "- Do NOT assume missing; confirm via code search first.\n"
	prompt += "- Identify where work already exists, partial implementations, TODOs, placeholders, skipped/flaky tests, or inconsistent patterns.\n"
	prompt += "- Keep the plan concise but complete; prefer lists and tables over paragraphs.\n"
	prompt += "- Use `[x]` only when verified in code. Use `[ ]` if missing or unverified.\n"
	prompt += "- Regenerate the plan if it becomes stale, contradictory, or significantly out of sync with code.\n"
	prompt += "- If the specified scope has relationships with other domain areas, implementation may be needed in those areas as well (always study the related specs and code). Include this in the plan.\n\n"

	prompt += "## Output Format Requirements\n\n"
	prompt += fmt.Sprintf("Write `%s` using this structure and level of detail:\n\n", cfg.ImplementationPlanName)

	prompt += "Header\n\n"
	prompt += "- Title: `Implementation Plan (<Scope>)`\n"
	prompt += "- Status line: `**Status:** <summary (e.g., \"UI Components Complete (39/39)\")>`\n"
	prompt += "- Last Updated date: `YYYY-MM-DD`\n"
	prompt += "- Reference to primary spec(s)\n\n"

	prompt += "Quick Reference\n\n"
	prompt += "- A table mapping systems/subsystems to:\n"
	prompt += "  - Specs\n"
	prompt += "  - Modules/packages\n"
	prompt += "  - Web packages\n"
	prompt += "  - Migrations or other artifacts\n"
	prompt += "- Use `✅` to mark items already implemented.\n\n"

	prompt += "Phased Plan\n\n"
	prompt += "- Use numbered phases ( e.g., Phase 9, Phase 10) aligned to the spec's domain.\n"
	prompt += "- Each phase includes:\n"
	prompt += "  - Goal\n"
	prompt += "  - Status (if applicable)\n"
	prompt += "  - Paths (directories or file patterns)\n"
	prompt += "  - Checklist with `[x]` for verified complete and `[ ]` for missing\n"
	prompt += "  - Definition of Done (tests run, commands/URLs, files touched)\n"
	prompt += "  - Risks/Dependencies (brief)\n"
	prompt += "- Break phases into subsections (e.g., 9.1, 9.2) with scope-specific paths and item lists.\n"
	prompt += "- Include \"Reference pattern\" links when there's a canonical directory or file to follow.\n\n"

	prompt += "Verification Log\n\n"
	prompt += "- A chronological log of verification steps with dates.\n"
	prompt += "- Each entry includes:\n"
	prompt += "  - What was verified (endpoints, commands, builds, tests, UI routes)\n"
	prompt += "  - Exact commands or URLs used\n"
	prompt += "  - Tests run and results\n"
	prompt += "  - Bug fixes discovered (if any)\n"
	prompt += "  - Files touched (if known from code search)\n"
	prompt += "  - Use format: `YYYY-MM-DD: <command or URL> - <result>`\n\n"

	prompt += "Summary\n\n"
	prompt += "- Table of phases with completion status\n"
	prompt += "  - \"Remaining effort\" line summarizing unfinished sections\n\n"

	prompt += "Known Existing Work\n\n"
	prompt += "- Brief section listing confirmed existing implementations to prevent duplicate work\n\n"

	prompt += "Manual Deployment Tasks\n\n"
	prompt += "- Required section to document manual steps needed before or during production deployment (manual configuration, third-party service setup, API key acquisition, etc).\n"
	prompt += "- If not applicable, write exactly: `None`.\n\n"

	prompt += "## Stop Condition\n\n"
	prompt += fmt.Sprintf("**IMPORTANT**: After writing/updating if `%s` already reflects the current gaps, reply with:\n", cfg.ImplementationPlanName)
	prompt += "`<COMPLETION_SIGNAL>`\n\n"

	return prompt
}

// findFileUpwards searches for a file starting from the current directory
// and moving up through parent directories until found or root is reached
func findFileUpwards(path string) string {
	// Check if the path is absolute and exists
	if filepath.IsAbs(path) {
		if _, err := os.Stat(path); err == nil {
			return path
		}
		return ""
	}

	// Start from current directory and search upwards
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	dir := cwd
	for {
		testPath := filepath.Join(dir, path)
		if _, err := os.Stat(testPath); err == nil {
			return testPath
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return ""
}
