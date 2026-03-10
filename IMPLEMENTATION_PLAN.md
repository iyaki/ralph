# Implementation Plan (config-by-prompt)

**Status:** In Progress
**Last Updated:** 2026-03-10
**Primary Spec:** [specs/config-by-prompt.md](specs/config-by-prompt.md)

## Quick Reference

| System                  | Spec                                          | Package           | Artifacts        | Implemented? |
| :---------------------- | :-------------------------------------------- | :---------------- | :--------------- | :----------- |
| **Config Data Model**   | [Config by Prompt](specs/config-by-prompt.md) | `internal/config` | `config.go`      | [x]          |
| **Front Matter Parser** | [Config by Prompt](specs/config-by-prompt.md) | `internal/prompt` | `frontmatter.go` | [x]          |
| **Prompt Resolution**   | [Config by Prompt](specs/config-by-prompt.md) | `internal/prompt` | `prompts.go`     | [ ]          |
| **CLI Integration**     | [Config by Prompt](specs/config-by-prompt.md) | `internal/cli`    | `cmd.go`         | [ ]          |

## Phased Plan

### Phase 1: Configuration Data Model

**Goal:** Update the configuration structure to support per-prompt overrides.
**Paths:** `internal/config/`

#### 1.1 Add Override Structures

- [x] Define `PromptConfigOverride` struct (Model, AgentMode).
- [x] Add `PromptOverrides` map to `Config` struct (`[prompt-overrides.<name>]`).
- [x] Update `config_test.go` to verify TOML parsing of the new section.

**Definition of Done:**

- `Config` struct can hold `prompt-overrides` data loaded from TOML.
- Unit tests pass.

### Phase 2: Front Matter Parsing

**Goal:** Implement parsing of YAML front matter from markdown prompts.
**Paths:** `internal/prompt/`

#### 2.1 YAML Parser Dependency

- [x] Add `gopkg.in/yaml.v3` dependency.

#### 2.2 Front Matter Extractor

- [x] Create `internal/prompt/frontmatter.go`.
- [x] Implement `ParseFrontMatter(content string) (*PromptFrontMatterSettings, string, error)`.
- [x] Ensure `ParseFrontMatter` returns the body with front matter stripped.
- [x] Handle invalid YAML (fail fast).
- [x] Handle unknown keys (ignore).
- [x] Add unit tests for various front matter scenarios (valid, invalid, missing, unknown keys).

**Definition of Done:**

- Reliable extraction of `model` and `agent-mode` from markdown content.
- Robust error handling and stripping logic.

### Phase 3: Integration & Precedence

**Goal:** Integrate front matter and config overrides into the CLI execution flow with correct precedence.
**Paths:** `internal/prompt/`, `internal/cli/`

#### 3.1 Update Prompt Resolver

- [ ] Update `GetPrompt` signature to return `(string, *config.PromptConfigOverride, error)`.
- [ ] Update `explicitPromptFile` and `promptFromDir` to use `ParseFrontMatter`.
- [ ] Ensure `bundledPrompt`, `customPrompt`, `stdinPrompt` return nil/empty overrides or handle accordingly.

#### 3.2 CLI Command Logic

- [ ] Update `RunE` in `internal/cli/cmd.go`.
- [ ] Implement precedence logic:
  1. CLI Flags (`cmd.Flags().Changed`)
  2. Env Vars (`os.Getenv`)
  3. Front Matter (from `GetPrompt`)
  4. Config Override (`cfg.PromptOverrides[name]`)
  5. Global Config (already in `cfg`)
- [ ] Apply the effective `Model` and `AgentMode` to the `Config` object before `RunLoop`.

**Definition of Done:**

- `ralph` command respects the precedence rules defined in the spec.
- `RunLoop` receives the correct Model and AgentMode.
- Manual verification with sample prompts and configs.

## Verification Log

| Date       | Verification Step              | Result |
| :--------- | :----------------------------- | :----- |
| 2026-03-10 | `go test -v ./internal/config` | PASS   |
| 2026-03-10 | `go test -v ./internal/prompt` | PASS   |

## Summary

| Phase                             | Status    | Completion |
| :-------------------------------- | :-------- | :--------- |
| Phase 1: Config Data Model        | Completed | 100%       |
| Phase 2: Front Matter Parsing     | Completed | 100%       |
| Phase 3: Integration & Precedence | Pending   | 0%         |
| **Remaining Effort**              | **Low**   | **33%**    |

## Known Existing Work

- `internal/config/config.go`: Existing configuration loading logic.
- `internal/prompt/prompts.go`: Existing prompt file reading logic (needs modification).
- `internal/cli/cmd.go`: Existing CLI entry point and flag setup.

## Manual Deployment Tasks

None.
