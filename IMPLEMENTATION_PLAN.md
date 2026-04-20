# Implementation Plan (Whole system)

**Status:** Core runtime is stable; 9/9 phases are complete.
**Last Updated:** 2026-04-20
**Primary Specs:** `specs/core-architecture.md`, `specs/configuration.md`, `specs/prompts.md`, `specs/agents.md`, `specs/init-command.md`, `specs/e2e-testing.md`, `specs/release-workflow.md`

## Quick Reference

| System/Subsystem | Specs | Modules/Packages | Web Packages | Migrations/Artifacts | Current State |
| --- | --- | --- | --- | --- | --- |
| CLI routing and loop control | `specs/core-architecture.md`, `specs/run-command.md` | `cmd/ralph/main.go` ✅, `internal/cli/cmd.go` ✅, `internal/cli/run.go` ✅ | None | CLI entrypoint + `run` command behavior ✅ | Implemented; default/alias/subcommand routing works and loop completion is enforced |
| Configuration and overlays | `specs/configuration.md`, `specs/config-local-overlay.md`, `specs/agent-env-overrides.md` | `internal/config/config.go` ✅, `internal/cli/run.go` ✅, `internal/config/config_local_test.go` ✅ | None | `ralph.toml` ✅, `ralph-local.toml` overlay ✅ | Implemented in runtime; `prompt-file` and `no-specs-index` file precedence are wired and TOML `config-file` now fails fast as unsupported |
| Prompt resolution and prompt-level overrides | `specs/prompts.md`, `specs/config-by-prompt.md` | `internal/prompt/prompts.go` ✅, `internal/prompt/frontmatter.go` ✅, `internal/cli/run.go` ✅ | None | Prompt markdown files (`<prompts-dir>/*.md`) ✅ | Implemented; front matter parsing/stripping and precedence are in place |
| Agent adapters and subprocess env wiring | `specs/agents.md`, `specs/agents/opencode.md`, `specs/agents/claude.md`, `specs/agents/cursor.md`, `specs/agent-env-overrides.md` | `internal/agent/agent.go` ✅, `internal/agent/runner.go` ✅, `internal/agent/opencode.go` ✅, `internal/agent/claude.go` ✅, `internal/agent/cursor.go` ✅ | None | e2e fixture symlinks for `opencode`/`claude`/`cursor` ✅ | Implemented; deterministic `cmd.Env` is passed and unknown agents fail fast |
| Logging | `specs/logging.md` | `internal/logger/logger.go` ✅, `internal/cli/run.go` ✅ | None | `ralph.log` creation/truncate/append semantics ✅ | Implemented; disabled-by-default logging with secure file permissions and git metadata headers |
| Init command bootstrap UX | `specs/init-command.md` | `internal/cli/init.go` ✅, `internal/config/writer.go` ✅ | None | Generated `ralph.toml` ✅ | Implemented in runtime; ordered questionnaire, retries, overwrite/preview confirmations, existing-config defaults, and robust stdin/stdout TTY validation are now in place |
| End-to-end suite and deterministic harness | `specs/e2e-testing.md` | `test/e2e/harness_test.go` ✅, `test/e2e/*.go` ✅, `test/e2e/agents/ralph-test-agent/main.go` ✅ | None | Test-only agent fixture binary ✅ | Implemented; broad scenario coverage exists, coverage matrix completeness is test-enforced, and required `return_error`/`slow_complete` agent-mode scenarios are covered |
| Quality, security, and release automation | `specs/development-testing.md`, `specs/release-workflow.md` | `Makefile` ✅, `.github/workflows/quality.yml` ✅, `.github/workflows/security.yml` ✅, `.github/workflows/release.yml` ✅ | None | Release binaries + `checksums.txt` ✅ | Implemented in automation; manual repo/org setup still required for production |
| Documentation and examples | `specs/README.md`, `specs/configuration.md`, `specs/init-command.md` | `README.md` ✅, `examples/ralph.toml` ✅, `IMPLEMENTATION_PLAN.md` (updated) ✅ | None | README regression checks in `cmd/ralph/main_test.go` ✅ | Implemented; docs/spec statuses are synchronized with runtime behavior |

## Phased Plan

### Phase 1: Command Routing and Loop Lifecycle

**Goal:** Keep CLI dispatch deterministic and preserve loop completion semantics across root and `run` invocations.
**Status:** Complete (1.1 and 1.2 verified in code)
**Paths:**

- `cmd/ralph/main.go`
- `internal/cli/cmd.go`
- `internal/cli/run.go`
- `internal/cli/cmd_test.go`
- `internal/cli/run_test.go`
- `test/e2e/run_command_test.go`

#### 1.1 Root dispatch, alias behavior, and collision policy

**Paths:**

- `internal/cli/cmd.go`
- `internal/cli/run.go`
- `test/e2e/run_command_test.go`

**Reference pattern:** `internal/cli/cmd.go` (single root command with explicit subcommand registration and shared run path)

**Checklist:**

- [x] Root command registers `init` and `run` subcommands.
- [x] Root command and `run` command both delegate to `runCommandLogic`.
- [x] Alias behavior (`ralph <prompt> [scope]`) is preserved for non-subcommand prompt names.
- [x] Collision behavior is verified: `ralph init` executes subcommand; `ralph run init` treats `init` as prompt.

#### 1.2 Loop control and completion signal behavior

**Paths:**

- `internal/cli/run.go`
- `internal/cli/cmd_test.go`

**Reference pattern:** `internal/cli/run.go` (`RunLoop` with placeholder replacement + exact completion token detection)

**Checklist:**

- [x] No-arg invocation defaults to `build` and scope defaults to `Whole system`.
- [x] `<COMPLETION_SIGNAL>` is replaced with `<promise>COMPLETE</promise>` before execution.
- [x] Completion is detected by trimmed line match of the completion token.
- [x] Max-iteration exhaustion returns a non-zero error path.
- [x] `DEBUG=1` short-circuits the loop after first iteration.

**Definition of Done:**

- `go test ./internal/cli -run 'TestNewRalphCommand|TestNewRunCommand|TestRunLoop' -count=1`
- `go test ./test/e2e -run TestE2ERunCommandRouting -count=1`
- Files touched: `internal/cli/*`, `cmd/ralph/main.go`, `test/e2e/run_command_test.go`

**Risks/Dependencies:**

- Routing behavior depends on Cobra subcommand precedence; adding future subcommands must preserve reserved-name collision behavior.

### Phase 2: Configuration Resolution and Overlay Semantics

**Goal:** Keep config precedence deterministic while closing schema/runtime parity gaps in file-backed fields.
**Status:** Complete (2.1 and 2.2 verified in code)
**Paths:**

- `internal/config/config.go`
- `internal/config/config_test.go`
- `internal/config/config_local_test.go`
- `internal/cli/run.go`
- `test/e2e/config_precedence_test.go`
- `test/e2e/config_local_test.go`

#### 2.1 Verified precedence and overlay behavior

**Paths:**

- `internal/config/config.go`
- `internal/config/config_local_test.go`
- `internal/config/config_test.go`

**Reference pattern:** `internal/config/config.go` (`resolveFileConfig` + `applyConfigValues` + `mergeConfig`)

**Checklist:**

- [x] Base config selection order is implemented as `--config` > `RALPH_CONFIG` > `./ralph.toml`.
- [x] Core precedence is implemented for most fields as `flags > env > config file > defaults`.
- [x] Sibling `ralph-local.toml` lookup is anchored to the selected base config directory.
- [x] Overlay deep merge is implemented for `[prompt-overrides.<prompt>]`.
- [x] Overlay deep merge is implemented for `[env]` map keys.

#### 2.2 Remaining config-schema parity gaps

**Paths:**

- `internal/config/config.go`
- `internal/config/config_test.go`
- `test/e2e/config_precedence_test.go`

**Reference pattern:** `internal/config/config.go:161` (`applyConfigValues` as the single precedence sink)

**Checklist:**

- [x] Wire config-file `prompt-file` into effective runtime config resolution (`Config.PromptFile` is now applied in `applyConfigValues`).
- [x] Wire config-file `no-specs-index` into effective runtime config resolution (`Config.NoSpecsIndex` is now applied in `applyConfigValues`).
- [x] Implement deterministic behavior for TOML `config-file` key by rejecting it as unsupported in base/overlay configs with fail-fast startup errors.
- [x] Add unit/e2e coverage for file-sourced `prompt-file` and `no-specs-index` precedence.

**Definition of Done:**

- `go test ./internal/config -run 'TestLoadConfig.*' -count=1`
- `go test ./internal/cli -run 'TestConfigPrecedence.*' -count=1`
- `go test ./test/e2e -run 'TestE2EConfigPrecedence|TestE2EConfigLocalOverlay' -count=1`
- Files touched: `internal/config/config.go`, `internal/config/config_test.go`, `test/e2e/config_precedence_test.go`

**Risks/Dependencies:**

- Precedence changes can regress existing boolean override behavior (`--no-log=false`, `--log-truncate=false`) if not isolated.

### Phase 3: Prompt Resolution and Prompt-Level Overrides

**Goal:** Preserve deterministic prompt-source precedence and safe front matter override behavior.
**Status:** Complete (3.1 and 3.2 verified in code)
**Paths:**

- `internal/prompt/prompts.go`
- `internal/prompt/frontmatter.go`
- `internal/prompt/prompts_test.go`
- `internal/prompt/frontmatter_test.go`
- `internal/cli/run.go`
- `test/e2e/prompt_test.go`
- `test/e2e/config_by_prompt_test.go`

#### 3.1 Prompt source precedence and fallback behavior

**Paths:**

- `internal/prompt/prompts.go`
- `internal/prompt/prompts_test.go`

**Reference pattern:** `internal/prompt/prompts.go` (`GetPrompt` source chain)

**Checklist:**

- [x] Inline prompt (`--prompt`) has top priority.
- [x] Stdin prompt mode works for `--prompt-file -` and `ralph -`.
- [x] Explicit `--prompt-file` is read before prompts-dir lookup.
- [x] Relative prompts-dir paths are searched upward; absolute paths are checked directly.
- [x] Built-in `build`/`plan` prompts are used as fallback.
- [x] Unknown prompt names fail with a clear error.

#### 3.2 Front matter extraction and effective settings merge

**Paths:**

- `internal/prompt/frontmatter.go`
- `internal/cli/run.go`
- `test/e2e/config_by_prompt_test.go`

**Reference pattern:** `internal/cli/run.go` (`applyModelSettings` / `applyAgentModeSettings`)

**Checklist:**

- [x] YAML front matter keys `model` and `agent-mode` are parsed from file-based prompts.
- [x] Front matter is stripped from prompt body before dispatch to the agent.
- [x] Invalid front matter fails before agent execution starts.
- [x] Effective precedence is honored as flags > env > front matter > `[prompt-overrides.<prompt>]` > global config.
- [x] Unknown front matter keys are ignored without affecting supported keys.

**Definition of Done:**

- `go test ./internal/prompt -count=1`
- `go test ./internal/cli -run 'TestConfigPrecedence_' -count=1`
- `go test ./test/e2e -run 'TestE2E(ConfigByPrompt|InlinePrompt|StdinPrompt)' -count=1`
- Files touched: `internal/prompt/*`, `internal/cli/run.go`, `test/e2e/config_by_prompt_test.go`

**Risks/Dependencies:**

- Front matter parser edge cases around leading `---` content need to stay deterministic and non-executable.

### Phase 4: Agent Adapters and Child Process Environment

**Goal:** Keep all supported agents behaviorally consistent and pass deterministic env to subprocesses.
**Status:** Complete (4.1 and 4.2 verified in code)
**Paths:**

- `internal/agent/agent.go`
- `internal/agent/runner.go`
- `internal/agent/opencode.go`
- `internal/agent/claude.go`
- `internal/agent/cursor.go`
- `internal/agent/agent_test.go`
- `test/e2e/agent_selection_test.go`
- `test/e2e/agent_env_overrides_test.go`

#### 4.1 Agent selection and invocation contract

**Paths:**

- `internal/agent/agent.go`
- `internal/agent/opencode.go`
- `internal/agent/claude.go`
- `internal/agent/cursor.go`

**Reference pattern:** `internal/agent/agent.go` (`GetAgent` single factory)

**Checklist:**

- [x] Supported agents are `opencode`, `claude`, and `cursor`.
- [x] Unknown configured agent names fail fast with clear error text.
- [x] Availability checks use `exec.LookPath`.
- [x] Agent CLI argument construction for model/mode is implemented per adapter spec.

#### 4.2 Effective environment construction and propagation

**Paths:**

- `internal/agent/runner.go`
- `internal/cli/run.go`
- `internal/agent/agent_test.go`
- `test/e2e/agent_env_overrides_test.go`

**Reference pattern:** `internal/agent/runner.go` (`BuildEffectiveEnv` + explicit `cmd.Env` assignment)

**Checklist:**

- [x] Effective env is built from inherited process env plus validated overrides.
- [x] Invalid env keys fail without leaking values.
- [x] `cmd.Env` is explicitly set for all agent subprocesses.
- [x] `GetAgent` snapshots env slices to avoid caller-side mutation affecting subprocess state.
- [x] E2E env override matrix covers config-only, flag-only, precedence, duplicate-key, and invalid-entry paths.

**Definition of Done:**

- `go test ./internal/agent -count=1`
- `go test ./internal/cli -run 'TestRunLoop(AppliesEffectiveEnvOverridesToAgentProcess|RejectsInvalidAgentEnvKeyBeforeExecution)' -count=1`
- `go test ./test/e2e -run TestE2EEnvOverrides -count=1`
- Files touched: `internal/agent/*`, `internal/cli/run.go`, `test/e2e/agent_env_overrides_test.go`

**Risks/Dependencies:**

- Child-process env behavior is security-sensitive; error paths must remain value-redacted.

### Phase 5: Logging and File-Safety Guarantees

**Goal:** Preserve secure, deterministic logging behavior with explicit enablement semantics.
**Status:** Complete (5.1 and 5.2 verified in code)
**Paths:**

- `internal/logger/logger.go`
- `internal/logger/logger_test.go`
- `internal/cli/run.go`
- `test/e2e/logging_flags_test.go`

#### 5.1 Enablement precedence and file behavior

**Paths:**

- `internal/config/config.go`
- `internal/cli/run.go`
- `internal/logger/logger.go`

**Reference pattern:** `internal/cli/run.go` (`readBoolFlagOverride` + `applyBoolFlagOverrides`)

**Checklist:**

- [x] Logging defaults to disabled (`NoLog=true`).
- [x] Explicit `--no-log=false` is honored and can override config/env disablement.
- [x] Log file supports append and truncate behavior.
- [x] Log directory/file permissions are restrictive (`0750` dir, `0600` file).

#### 5.2 Header metadata and stream parity

**Paths:**

- `internal/logger/logger.go`
- `test/e2e/logging_flags_test.go`

**Reference pattern:** `internal/logger/logger.go` (header + git metadata write path)

**Checklist:**

- [x] Log header includes timestamp and git metadata (`N/A` fallback when unresolved).
- [x] Output is streamed to stdout and log file through a multi-writer path.
- [x] Empty log path creates a temp file when logging is enabled.

**Definition of Done:**

- `go test ./internal/logger -count=1`
- `go test ./test/e2e -run TestE2ELoggingFlags -count=1`
- Files touched: `internal/logger/logger.go`, `internal/cli/run.go`, `test/e2e/logging_flags_test.go`

**Risks/Dependencies:**

- Logging still captures prompt/agent output verbatim; operational guidance must treat log paths as sensitive.

### Phase 6: Init Command Interactive Workflow

**Goal:** Move `ralph init` from starter-file bootstrap to the full interactive questionnaire behavior described in spec.
**Status:** Complete (6.1 and 6.2 verified in code)
**Paths:**

- `internal/cli/init.go`
- `internal/cli/init_test.go`
- `internal/cli/init_internal_test.go`
- `internal/config/writer.go`
- `specs/init-command.md`
- `test/e2e/init_command_test.go`

#### 6.1 Current implemented baseline

**Paths:**

- `internal/cli/init.go`
- `internal/config/writer.go`

**Reference pattern:** `internal/cli/init.go` (guard checks + default config write)

**Checklist:**

- [x] `init` subcommand exists with `--output` and `--force` flags.
- [x] Non-interactive invocation fails fast with a terminal requirement error.
- [x] Starter `ralph.toml` is written atomically.
- [x] Existing file path prompts for overwrite confirmation unless `--force` is set.

#### 6.2 Interactive target behavior from spec

**Paths:**

- `internal/cli/init.go`
- `internal/cli/init_internal_test.go`
- `specs/init-command.md`

**Reference pattern:** `specs/init-command.md` interactive question set and workflow table

**Checklist:**

- [x] Implement ordered interactive questionnaire with validation/re-prompt loops.
- [x] Seed defaults from existing config file values when present.
- [x] Implement overwrite confirmation flow when file exists and `--force` is not set.
- [x] Add final preview/confirmation step before write.
- [x] Expand TTY check to robustly validate interactive input/output expectations.
- [x] Add tests for invalid answer retry paths.
- [x] Add tests for declined-overwrite no-op behavior.
- [x] Add tests for preview-declined write cancellation behavior.

**Definition of Done:**

- `go test ./internal/cli -run 'TestInit' -count=1`
- `go test ./test/e2e -run TestE2EInitCommand -count=1`
- Files touched: `internal/cli/init.go`, `internal/cli/init_internal_test.go`, `test/e2e/init_command_test.go`

**Risks/Dependencies:**

- Interactive flows are hard to test without deterministic I/O seams; command design should keep prompt logic isolated from filesystem writes.

### Phase 7: End-to-End Coverage Matrix and Governance

**Goal:** Close the gap between existing broad e2e scenarios and the spec requirement for explicit full-surface traceability.
**Status:** Complete (7.1 and 7.2 verified in code)
**Paths:**

- `test/e2e/harness_test.go`
- `test/e2e/types_test.go`
- `test/e2e/*.go`
- `test/e2e/COVERAGE_MATRIX.md`
- `test/e2e/agents/ralph-test-agent/main.go`
- `specs/e2e-testing.md`

#### 7.1 Existing deterministic harness and scenario coverage

**Paths:**

- `test/e2e/harness_test.go`
- `test/e2e/scenarios_test.go`
- `test/e2e/config_precedence_test.go`
- `test/e2e/logging_flags_test.go`

**Reference pattern:** `test/e2e/harness_test.go` (single process harness for args/env/files/assertions)

**Checklist:**

- [x] Harness builds one test binary and reuses one fixture-agent implementation via symlinks.
- [x] Scenario assertions cover stdout/stderr/exit code/files/forbidden output.
- [x] Core flows are covered (completion, max-iterations, prompt failures, logging, routing, agent selection).
- [x] Env override and local overlay scenarios are present.

#### 7.2 Remaining spec-required completeness work

**Paths:**

- `test/e2e/*.go`
- `specs/e2e-testing.md`

**Reference pattern:** `specs/e2e-testing.md` coverage requirements and traceability rules

**Checklist:**

- [x] Add a maintained coverage matrix artifact mapping every supported option/config/output behavior to concrete e2e test names.
- [x] Add CI enforcement for matrix completeness (fail when a required mapping is missing/stale).
- [x] Add e2e invalid-config parse failure scenario (`ralph.toml` malformed).
- [x] Add e2e scenario validating `RALPH_TEST_AGENT_MODE=return_error` path.
- [x] Add e2e scenario validating `RALPH_TEST_AGENT_MODE=slow_complete` deterministic delay path.
- [x] Add e2e scenarios for file-sourced `prompt-file` and `no-specs-index` once Phase 2 parity gaps are fixed.

**Definition of Done:**

- `go test ./test/e2e -count=1`
- `go test ./test/e2e -run 'TestE2E(ConfigPrecedence|ConfigLocalOverlay|EnvOverrides|Logging|RunCommandRouting|InitCommand)' -count=1`
- Files touched: `test/e2e/*.go`, coverage matrix artifact path, CI workflow files

**Risks/Dependencies:**

- Enforcing matrix completeness adds process overhead; rule design must stay low-friction and deterministic.

### Phase 8: Quality, Security, and Release Automation

**Goal:** Keep local and CI quality gates aligned with release publishing behavior.
**Status:** Complete (8.1 and 8.2 verified in code)
**Paths:**

- `Makefile`
- `.github/workflows/quality.yml`
- `.github/workflows/security.yml`
- `.github/workflows/release.yml`
- `specs/development-testing.md`
- `specs/release-workflow.md`

#### 8.1 Local/CI quality and security gates

**Paths:**

- `Makefile`
- `.github/workflows/quality.yml`
- `.github/workflows/security.yml`

**Reference pattern:** `Makefile` (`quality`, `test-e2e`, `coverage`, `mutation`, `security`, `arch` targets)

**Checklist:**

- [x] Local quality targets cover lint, tests, race, coverage gate, mutation, security, and architecture.
- [x] CI quality workflow runs lint/test/coverage/arch/mutation jobs.
- [x] CI security workflow runs `govulncheck`, `gosec`, and Semgrep job.

#### 8.2 Release publication workflow

**Paths:**

- `.github/workflows/release.yml`

**Reference pattern:** `.github/workflows/release.yml` (prepare -> build matrix -> publish with checksums)

**Checklist:**

- [x] Releases trigger on semver tags (`v*`) and manual `workflow_dispatch`.
- [x] Manual flow supports optional tag creation.
- [x] Matrix builds generate cross-platform binaries and `checksums.txt`.
- [x] Release assets are published with `softprops/action-gh-release`.

**Definition of Done:**

- `make quality`
- `gh workflow run release.yml -f tag=vX.Y.Z -f create_tag=false` (manual validation path in a release-capable environment)
- Files touched: `Makefile`, `.github/workflows/*.yml`

**Risks/Dependencies:**

- Release/publish success depends on repo permissions, protected tag policy, and GitHub token scopes.

### Phase 9: Documentation and Spec Status Alignment

**Goal:** Keep docs/spec statuses synchronized with actual runtime behavior and avoid stale implementation-plan scope.
**Status:** Complete (9.1 and 9.2 verified)
**Paths:**

- `README.md`
- `examples/ralph.toml`
- `specs/README.md`
- `specs/configuration.md`
- `specs/init-command.md`
- `IMPLEMENTATION_PLAN.md`

#### 9.1 Confirmed existing documentation alignment

**Paths:**

- `README.md`
- `examples/ralph.toml`
- `specs/README.md`
- `cmd/ralph/main_test.go`

**Reference pattern:** `cmd/ralph/main_test.go` README assertions for canonical repo links/naming

**Checklist:**

- [x] README documents `iyaki/ralphex` repo naming with `ralph` CLI command naming split.
- [x] README documents release URLs, local overlay behavior, and child env precedence.
- [x] `examples/ralph.toml` includes `[env]` override examples.
- [x] Specs index includes current feature specs.
- [x] Whole-system plan regenerated from stale env-only phase history.

#### 9.2 Remaining status synchronization work

**Paths:**

- `specs/configuration.md`
- `specs/init-command.md`
- `IMPLEMENTATION_PLAN.md`

**Reference pattern:** `specs/init-command.md` status header (`Implemented`) and `specs/configuration.md` key semantics for unsupported TOML `config-file`

**Checklist:**

- [x] After completed Phase 2 parity fixes, update `specs/configuration.md` status and verification bullets to fully implemented behavior.
- [x] After Phase 6 interactive init work, update `specs/init-command.md` status and remove the temporary "current implementation note" caveat.
- [x] Keep this plan synchronized after each merged feature to prevent drift between specs and code.

**Definition of Done:**

- `go test ./cmd/ralph -run TestReadmeDocumentsRalphexRepoAndRalphCli -count=1`
- docs/spec review pass over `README.md`, `specs/*.md`, and `IMPLEMENTATION_PLAN.md`
- Files touched: docs/spec markdown files listed above

**Risks/Dependencies:**

- Spec/documentation drift can reintroduce false-positive "implemented" claims and create duplicate feature work.

## Verification Log

- 2026-04-20: `git status --short` - verified working tree is clean for planning update; tests run: none (planning mode); bug fixes discovered: none; files touched: none.
- 2026-04-20: `git log --oneline --decorate -n 30 -- specs` - identified latest scope-shaping spec commits (`7e485bc`, `585a05c`, `328fc39`, `4422684`, `8965973`, `1321859`); tests run: none; bug fixes discovered: none; files touched: `specs/*.md`.
- 2026-04-20: `git log --oneline --decorate -n 20 -- IMPLEMENTATION_PLAN.md` - confirmed existing plan history is scoped to older env/logging phases and stale for whole-system planning; tests run: none; bug fixes discovered: none; files touched: `IMPLEMENTATION_PLAN.md`.
- 2026-04-20: `git show --oneline --name-only 7e485bc -- specs` - verified broad docs refresh touched configuration/logging/init/core/e2e specs; tests run: none; bug fixes discovered: none; files touched: `specs/agents/opencode.md`, `specs/configuration.md`, `specs/core-architecture.md`, `specs/development-testing.md`, `specs/e2e-testing.md`, `specs/init-command.md`, `specs/logging.md`, `specs/release-workflow.md`.
- 2026-04-20: `git show --oneline --name-only 585a05c -- specs` and `git show --oneline --name-only 328fc39 -- specs` - verified spec updates for config-file discovery simplification and unknown-agent failure behavior; tests run: none; bug fixes discovered: none; files touched: `specs/configuration.md`, `specs/config-local-overlay.md`, `specs/agents.md`.
- 2026-04-20: `git show --oneline --name-only 4422684 -- specs`, `git show --oneline --name-only 8965973 -- specs`, `git show --oneline --name-only 1321859 -- specs` - verified status updates for config-by-prompt/init, expanded e2e requirements, and agent-env-overrides spec introduction; tests run: none; bug fixes discovered: none; files touched: `specs/config-by-prompt.md`, `specs/init-command.md`, `specs/e2e-testing.md`, `specs/README.md`, `specs/agent-env-overrides.md`.
- 2026-04-20: `grep pattern="Status:\s*Partially Implemented|Status:\s*Implemented|Status:\s*Proposed|Status:\s*Draft" include="*.md" path="/workspaces/ralph"` - confirmed only `specs/configuration.md` and `specs/init-command.md` remain partial among active runtime specs; tests run: none; bug fixes discovered: none; files touched: `specs/configuration.md`, `specs/init-command.md`.
- 2026-04-20: `grep pattern="TODO|FIXME|XXX|HACK" include="*.go" path="/workspaces/ralph"` and `grep pattern="t\.Skip|Skip\(|Flaky|flaky" include="*.go" path="/workspaces/ralph"` - found no actionable TODOs or skip/flaky markers in runtime/test code (only prompt text mentions); tests run: none; bug fixes discovered: none; files touched: `internal/prompt/prompts.go`.
- 2026-04-20: `grep pattern="NoSpecsIndex|no-specs-index" include="*.go" path="/workspaces/ralph"` - verified flag and merge hooks exist, but no file-precedence assignment path in `applyConfigValues`; tests run: none; bug fixes discovered: configuration parity gap identified; files touched: `internal/config/config.go`, `internal/cli/run.go`, `internal/prompt/prompts.go`.
- 2026-04-20: `grep pattern="PromptFile|prompt-file" include="*.go" path="/workspaces/ralph"` - verified `PromptFile` struct tags and flag wiring exist, but no config-file precedence assignment in `applyConfigValues`; tests run: none; bug fixes discovered: configuration parity gap identified; files touched: `internal/config/config.go`, `internal/cli/run.go`, `internal/prompt/prompts.go`.
- 2026-04-20: `grep pattern="config-file" include="*.go" path="/workspaces/ralph"` - verified TOML field exists (`Config.ConfigFile`) but is not resolved as an effective runtime key; tests run: none; bug fixes discovered: schema parity gap identified; files touched: `internal/config/config.go`.
- 2026-04-20: `grep pattern="ExecuteCommand\(|internal/executor|executor\." include="*.go" path="/workspaces/ralph"` - verified `internal/executor` is currently test-only/unused by runtime path while agent runner has duplicate streaming helper pattern; tests run: none; bug fixes discovered: none (captured as consistency risk); files touched: `internal/executor/executor.go`, `internal/agent/runner.go`.
- 2026-04-20: read pass over `.github/workflows/release.yml`, `.github/workflows/quality.yml`, `.github/workflows/security.yml`, `Makefile`, and `cmd/ralph/main_test.go` - verified release and quality automation wiring and README regression checks are present; tests run: none; bug fixes discovered: none; files touched: workflow files, `Makefile`, `cmd/ralph/main_test.go`.
- 2026-04-20: `go test ./internal/config -run 'TestLoadConfig(PromptFileFromConfigFile|PromptFileFlagWinsOverConfigFile|NoSpecsIndexFromConfigFile|NoSpecsIndexFlagWinsOverConfigFile)$' -count=1` - failed as expected before implementation (`PromptFile` and `NoSpecsIndex` from config file were not applied).
- 2026-04-20: `go test ./internal/config -count=1` - passed after wiring file-sourced `prompt-file`/`no-specs-index` precedence and adding overlay merge support for `prompt-file`.
- 2026-04-20: `go test ./internal/cli -run 'TestConfigPrecedence_.*' -count=1` - passed; prompt/front matter precedence behavior remained stable after config precedence changes.
- 2026-04-20: `go test ./test/e2e -run 'TestE2EConfigPrecedence|TestE2EConfigLocalOverlay' -count=1` - passed including new e2e coverage for config-file `prompt-file` and `no-specs-index` behavior.
- 2026-04-20: `go test ./internal/config -run 'TestLoadConfigRejectsConfigFileKey|TestLoadConfigRejectsConfigFileKeyInDefaultConfig' -count=1` - failed as expected before implementation because unsupported TOML `config-file` entries were accepted.
- 2026-04-20: `go test ./internal/config -run 'TestLoadConfigRejectsConfigFileKey|TestLoadConfigRejectsConfigFileKeyInDefaultConfig|TestLoadConfigRejectsConfigFileKeyInOverlay' -count=1` - passed after adding fail-fast validation for unsupported `config-file` keys in base/default/overlay config files.
- 2026-04-20: `go test ./test/e2e -run 'TestE2EConfigPrecedence_ConfigFileKeyInBaseConfigFails|TestE2EConfigPrecedence_ConfigFileKeyInOverlayFails' -count=1` - passed with deterministic non-zero exits and no agent execution when unsupported `config-file` keys are present.
- 2026-04-20: `go test ./internal/config -run 'TestLoadConfig.*' -count=1 && go test ./internal/cli -run 'TestConfigPrecedence.*' -count=1 && go test ./test/e2e -run 'TestE2EConfigPrecedence|TestE2EConfigLocalOverlay' -count=1` - passed full Phase 2 Definition of Done validation after implementing unsupported `config-file` fail-fast behavior.
- 2026-04-20: `go test ./internal/cli -run 'TestInitCommand(AsksQuestionsInSpecifiedOrder|RePromptsForInvalidAnswers)$' -count=1` - failed before implementation (questionnaire prompts/validation retries were not present).
- 2026-04-20: `go test ./internal/cli -run 'TestInit' -count=1` - passed after implementing ordered init questionnaire, conditional logging questions, and retry-on-invalid-input behavior.
- 2026-04-20: `go test ./test/e2e -run TestE2EInitCommand -count=1` - passed; non-TTY guard and run-subcommand collision behavior remained stable.
- 2026-04-20: `make lint` - passed (0 issues).
- 2026-04-20: `make test-coverage` - passed (overall coverage gate >= 90%).
- 2026-04-20: `make security && make arch` - passed (no gosec findings; architecture lint clean).
- 2026-04-20: `go test ./internal/cli -run 'TestInitCommand(DeclinedOverwriteLeavesExistingFileUnchanged|ConfirmedOverwriteRewritesExistingFile)$' -count=1` - failed as expected before implementation because init returned the existing-file `--force` error path.
- 2026-04-20: `go test ./internal/cli -run 'TestInitCommand(DeclinedOverwriteLeavesExistingFileUnchanged|ConfirmedOverwriteRewritesExistingFile)$' -count=1` - passed after adding interactive overwrite confirmation and no-op-on-decline behavior.
- 2026-04-20: `go test ./internal/cli -run 'TestInit' -count=1` - passed after refactoring `NewInitCommand` into helper functions to satisfy lint constraints.
- 2026-04-20: `go test ./test/e2e -run TestE2EInitCommand -count=1` - passed; non-interactive and `run init` routing scenarios remained stable.
- 2026-04-20: `make lint` - passed (0 issues) after init command refactor.
- 2026-04-20: `go test ./internal/cli -run TestInitCommandSeedsQuestionDefaultsFromExistingConfig -count=1` - failed as expected before implementation (questionnaire kept hardcoded defaults instead of reading existing `ralph.toml` values).
- 2026-04-20: `go test ./internal/cli -run TestInitCommandSeedsQuestionDefaultsFromExistingConfig -count=1` - passed after adding existing-config default seeding for supported questionnaire fields.
- 2026-04-20: `go test ./internal/cli -run 'TestInit' -count=1` - passed after integrating existing-config default seeding and helper refactors.
- 2026-04-20: `go test ./test/e2e -run TestE2EInitCommand -count=1` - passed; init non-TTY and `run init` routing behavior remained stable after default seeding changes.
- 2026-04-20: `make lint` - passed (0 issues) after seeding implementation and test helper extraction.
- 2026-04-20: `go test ./internal/cli -run 'TestInitCommand(AsksQuestionsInSpecifiedOrder|PreviewDeclinedSkipsWrite|RePromptsForInvalidAnswers|SeedsQuestionDefaultsFromExistingConfig)$' -count=1` - failed as expected before implementation because final preview confirmation and write-gate behavior were missing.
- 2026-04-20: `go test ./internal/cli -run 'TestInitCommand(AsksQuestionsInSpecifiedOrder|PreviewDeclinedSkipsWrite|RePromptsForInvalidAnswers|SeedsQuestionDefaultsFromExistingConfig)$' -count=1` - passed after adding a final preview summary and confirmation before write.
- 2026-04-20: `go test ./internal/cli -run 'TestInit' -count=1` - passed after preview confirmation implementation.
- 2026-04-20: `go test ./test/e2e -run TestE2EInitCommand -count=1` - passed; init non-TTY guard and `run init` routing behavior remained stable after preview confirmation changes.
- 2026-04-20: `go test ./internal/cli -run TestIsInteractiveTerminalRejectsDevNullStreams -count=1` - failed as expected before implementation because `isInteractiveTerminal` accepted `/dev/null` streams as interactive.
- 2026-04-20: `go test ./internal/cli -run 'TestInit|TestIsInteractiveTerminalRejectsDevNullStreams' -count=1` - passed after requiring both stdin/stdout character-device checks and terminal FD checks.
- 2026-04-20: `go test ./test/e2e -run TestE2EInitCommand -count=1` - passed; init non-TTY guard and `run init` routing behavior remained stable after TTY hardening.
- 2026-04-20: `go test ./test/e2e -count=1` - passed after adding `test/e2e/COVERAGE_MATRIX.md` to map supported option/config/output behavior to concrete e2e tests.
- 2026-04-20: `go test ./test/e2e -run TestCoverageMatrixCompleteness -count=1` - failed as expected before implementation because matrix-enforcement helpers were not yet implemented.
- 2026-04-20: `go test ./test/e2e -run TestCoverageMatrixCompleteness -count=1` - passed after adding coverage-matrix completeness enforcement over discovered `TestE2E*` cases and matrix references.
- 2026-04-20: `make lint` - failed initially on cyclomatic complexity in coverage-matrix helper, then passed after refactoring into smaller helper functions.
- 2026-04-20: `make test` - passed full Go suite including `test/e2e` after matrix-enforcement test and helper refactor.
- 2026-04-20: `go test ./test/e2e -run TestE2EConfigPrecedence_InvalidBaseConfigFailsBeforeAgentExecution -count=1` - passed after adding malformed base `ralph.toml` parse-failure coverage; scenario validates deterministic non-zero exit before agent execution.
- 2026-04-20: `go test ./test/e2e -run 'TestE2EConfigPrecedence_InvalidBaseConfigFailsBeforeAgentExecution|TestCoverageMatrixCompleteness' -count=1` - passed; coverage matrix mapping is updated and enforcement remains green.
- 2026-04-20: `go test ./test/e2e -run TestE2EReturnErrorPath -count=1` - passed after adding the `RALPH_TEST_AGENT_MODE=return_error` scenario to validate deterministic warning+max-iteration failure behavior without completion output.
- 2026-04-20: `go test ./test/e2e -run 'TestE2EReturnErrorPath|TestCoverageMatrixCompleteness' -count=1` - passed; coverage matrix completeness enforcement remains green after adding the new scenario mapping.
- 2026-04-20: `go test ./test/e2e -count=1` - passed full e2e suite after adding return-error scenario coverage and matrix updates.
- 2026-04-20: `go test ./test/e2e -run TestE2ESlowCompletePath -count=1` - failed as expected before implementation because deterministic delay assertions were not available in the shared e2e test case schema.
- 2026-04-20: `go test ./test/e2e -run TestE2ESlowCompletePath -count=1` - passed after adding `slow_complete` scenario coverage and minimum-duration assertions in the harness.
- 2026-04-20: `go test ./test/e2e -run TestCoverageMatrixCompleteness -count=1` - failed as expected before implementation because `COVERAGE_MATRIX.md` did not yet map `TestE2ESlowCompletePath`.
- 2026-04-20: `go test ./test/e2e -run 'TestE2ESlowCompletePath|TestCoverageMatrixCompleteness' -count=1 && go test ./test/e2e -count=1` - passed; slow-complete scenario coverage and full e2e suite remain deterministic and green.
- 2026-04-20: `go test ./internal/config -run 'TestLoadConfig.*' -count=1` - passed; config runtime behavior remains green while synchronizing `specs/configuration.md` status and key semantics.
- 2026-04-20: `go test ./cmd/ralph -run TestReadmeDocumentsRalphexRepoAndRalphCli -count=1` - passed; docs regression guard remains green after spec synchronization updates.
- 2026-04-20: `grep -n '^Status: Partially Implemented' specs/*.md` - confirmed only `specs/init-command.md` remains partial after completing the configuration-spec synchronization task.
- 2026-04-20: `go test ./internal/cli -run TestInit -count=1` - passed after synchronizing `specs/init-command.md` status/content with implemented interactive workflow behavior.
- 2026-04-20: `go test ./cmd/ralph -run TestReadmeDocumentsRalphexRepoAndRalphCli -count=1` - passed; docs regression guard remains green after init-spec synchronization updates.
- 2026-04-20: `grep pattern="^Status:\s*Partially Implemented" include="*.md" path="/workspaces/ralph/specs"` - no matches; active specs now report implemented status.

## Summary

| Phase | Goal | Status |
| --- | --- | --- |
| Phase 1 | Command routing and loop lifecycle | Complete |
| Phase 2 | Configuration resolution and overlay semantics | Complete |
| Phase 3 | Prompt resolution and prompt-level overrides | Complete |
| Phase 4 | Agent adapters and child process environment | Complete |
| Phase 5 | Logging and file-safety guarantees | Complete |
| Phase 6 | Init command interactive workflow | Complete |
| Phase 7 | End-to-end coverage matrix and governance | Complete |
| Phase 8 | Quality, security, and release automation | Complete |
| Phase 9 | Documentation and spec status alignment | Complete |

**Remaining effort:** None; continue routine docs/plan maintenance as features evolve.

## Known Existing Work

- Root and `run` command paths already converge through `runCommandLogic`, preserving deterministic routing.
- Built-in `build` and `plan` prompts already include planning/build-mode instructions with completion-signal placeholders.
- Prompt front matter parsing/stripping and precedence merge with `[prompt-overrides]` already exist.
- File-sourced `prompt-file` and `no-specs-index` config precedence now resolve correctly and are covered by unit/e2e tests.
- TOML `config-file` is treated as an unsupported key and now fails fast in base/default/overlay config resolution paths.
- `specs/configuration.md` now marks implemented status and documents `RALPH_CONFIG` selection plus unsupported TOML `config-file` behavior.
- `specs/init-command.md` now marks implemented status and reflects the shipped interactive questionnaire/preview workflow (without stale partial-implementation caveats).
- Child-process env overrides via `[env]` and repeatable `--env` are already implemented with validation and deterministic merge order.
- All supported agents (`opencode`, `claude`, `cursor`) already use a shared subprocess runner with explicit `cmd.Env`.
- Unknown agent names already fail fast before loop execution.
- Logging defaults, secure file permissions, and stdout-log parity are already implemented and covered by tests.
- `ralph init` now runs an ordered interactive questionnaire with per-question validation/re-prompt behavior, conditional logging follow-up prompts, overwrite confirmation/no-op behavior, existing-config default seeding for supported fields, a final preview confirmation, and robust stdin/stdout TTY validation before writing config.
- E2E harness already compiles one fixture agent and symlinks all supported agent names to it.
- E2E coverage matrix artifact now exists at `test/e2e/COVERAGE_MATRIX.md` and maps supported option/config/output behaviors to concrete e2e tests.
- E2E suite now enforces coverage matrix completeness via `TestCoverageMatrixCompleteness`, which fails on missing or stale `COVERAGE_MATRIX.md` test mappings.
- E2E config precedence coverage now includes malformed base `ralph.toml` parse failures, validated to fail before any agent execution.
- E2E scenario coverage now includes both `RALPH_TEST_AGENT_MODE=return_error` (`TestE2EReturnErrorPath`) and `RALPH_TEST_AGENT_MODE=slow_complete` (`TestE2ESlowCompletePath`), and the matrix marks both paths complete.
- Release workflow already builds cross-platform artifacts and publishes checksums.
- README regression checks already guard canonical `iyaki/ralphex` links and CLI naming text.

## Manual Deployment Tasks

- None
