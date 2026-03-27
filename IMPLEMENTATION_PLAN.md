# Implementation Plan (agent-env-overrides)

**Status:** In Progress (Phase 10 complete; Phases 11-12 pending)
**Last Updated:** 2026-03-27
**Primary Specs:** `specs/agent-env-overrides.md` (scope), `specs/configuration.md`, `specs/agents.md`, `specs/e2e-testing.md`

## Quick Reference

| System/Subsystem                 | Specs                                                                                     | Modules/Packages                                                                                                                                           | Web Packages | Migrations/Artifacts                                 | Current State                                                       |
| -------------------------------- | ----------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | ---------------------------------------------------- | ------------------------------------------------------------------- |
| Agent command execution          | `specs/agent-env-overrides.md`, `specs/agents.md`                                         | `internal/agent/runner.go` ✅, `internal/agent/agent.go` ✅, `internal/agent/opencode.go` ✅, `internal/agent/claude.go` ✅, `internal/agent/cursor.go` ✅ | None         | `test/e2e/agents/ralph-test-agent/main.go` ✅        | Shared runner now injects deterministic effective env via `cmd.Env` |
| Config loading and merge         | `specs/agent-env-overrides.md`, `specs/configuration.md`, `specs/config-local-overlay.md` | `internal/config/config.go` ✅, `internal/config/config_local_test.go` ✅                                                                                  | None         | `ralph.toml`, `ralph-local.toml` overlay behavior ✅ | `[env]` table decode + deterministic overlay merge implemented      |
| CLI flag plumbing                | `specs/agent-env-overrides.md`, `specs/run-command.md`                                    | `internal/cli/run.go` ✅ (`setupSharedFlags`)                                                                                                              | None         | CLI root and `run` command share flags ✅            | Repeatable `--env` implemented with validation and override merge   |
| E2E harness and precedence tests | `specs/e2e-testing.md`                                                                    | `test/e2e/harness_test.go` ✅, `test/e2e/config_precedence_test.go` ✅                                                                                     | None         | deterministic fixture agent symlink setup ✅         | Harness ready; agent-env override scenarios missing                 |
| Scope spec artifact              | `specs/agent-env-overrides.md` ✅                                                         | n/a                                                                                                                                                        | None         | Spec commit `d3461d1` ✅                             | Proposed spec exists; implementation gap confirmed                  |

## Phased Plan

### Phase 9: Config and CLI Input Surfaces

**Goal:** Add config and flag inputs for child-process environment overrides without changing existing non-env precedence rules.
**Status:** Complete (9.1 and 9.2 complete)
**Paths:**

- `internal/config/config.go`
- `internal/cli/run.go`
- `internal/cli/cmd.go`
- `internal/config/config_test.go`
- `internal/config/config_local_test.go`
- `internal/cli/run_internal_test.go`

#### 9.1 Config schema and overlay behavior

**Paths:**

- `internal/config/config.go`
- `internal/config/config_test.go`
- `internal/config/config_local_test.go`

**Reference pattern:** `internal/config/config_local_test.go` (map-like deep merge behavior using TOML metadata checks)

**Checklist:**

- [x] Verified existing config precedence engine (`flags > env > file > defaults`) in `internal/config/config.go`.
- [x] Verified existing overlay merge mechanism for map-like `prompt-overrides` in `internal/config/config.go` and `internal/config/config_local_test.go`.
- [x] Add `Config` support for `[env]` table (`map[string]string` or equivalent) with TOML decode support.
- [x] Merge `[env]` entries from base and `ralph-local.toml` deterministically (local over base per key).
- [x] Keep existing `RALPH_*` precedence unchanged for non-scope fields.

#### 9.2 CLI `--env` input parsing and validation

**Paths:**

- `internal/cli/run.go`
- `internal/cli/run_internal_test.go`

**Reference pattern:** `internal/cli/run_internal_test.go` (changed-flag tracking and override plumbing)

**Checklist:**

- [x] Verified single shared flag registration point in `setupSharedFlags` (`internal/cli/run.go`) used by both root and `run` commands.
- [x] Add repeatable `--env` flag (`string[]`) with raw `KEY=VALUE` entries.
- [x] Parse each entry with split-on-first-`=` semantics and allow empty values (`KEY=`).
- [x] Validate keys against `^[A-Za-z_][A-Za-z0-9_]*$` and fail before agent execution.
- [x] Preserve command-line order for duplicate keys (last value wins).

**Definition of Done:**

- `go test ./internal/config -run 'TestLoadConfig.*Env|TestLoadConfigWithOverlay.*' -count=1` passes.
- `go test ./internal/cli -run 'TestRead.*Env|TestApply.*Env' -count=1` passes.
- Files touched are limited to config/CLI parsing and tests listed in this phase.

**Risks/Dependencies:**

- Boolean-flag precedence fixes in `run.go` must remain intact while adding new repeatable string flags.
- Overlay merge logic for maps must stay deterministic across TOML metadata edge cases.

### Phase 10: Effective Environment Construction and Agent Wiring

**Goal:** Build deterministic effective environment and pass it explicitly to agent subprocesses.
**Status:** Complete (10.1 and 10.2 complete)
**Paths:**

- `internal/cli/run.go`
- `internal/agent/runner.go`
- `internal/agent/agent.go`
- `internal/agent/agent_test.go`
- `internal/agent/opencode.go`
- `internal/agent/claude.go`
- `internal/agent/cursor.go`
- `internal/cli/cmd_test.go`

#### 10.1 Effective env merge service

**Paths:**

- `internal/cli/run.go`
- `internal/agent/runner.go`

**Reference pattern:** `internal/config/config.go` resolver helpers (deterministic precedence ordering)

**Checklist:**

- [x] Verified runner choke point exists in `internal/agent/runner.go` (`executeAgentCommand`).
- [x] Add an effective env builder that starts from `os.Environ()`, applies config `[env]`, then applies CLI `--env`.
- [x] Enforce precedence exactly as spec: inherited env < config table < CLI flags.
- [x] Handle values containing additional `=` characters without truncation.
- [x] Return redacted/entry-level errors that do not leak sensitive values.

#### 10.2 Pass env through all agents consistently

**Paths:**

- `internal/agent/runner.go`
- `internal/agent/opencode.go`
- `internal/agent/claude.go`
- `internal/agent/cursor.go`
- `internal/agent/agent_test.go`

**Reference pattern:** `internal/agent/agent_test.go` (cross-agent behavior assertions using fixture executables)

**Checklist:**

- [x] Verified all supported agents route process execution through `executeAgentCommand`.
- [x] Extend runner signature to accept effective environment and set `cmd.Env` explicitly.
- [x] Thread effective env from CLI layer to each agent execution call without changing prompt resolution behavior.
- [x] Add/extend tests proving consistent override behavior for `opencode`, `claude`, and `cursor`.

**Definition of Done:**

- `go test ./internal/agent -count=1` passes with env override coverage.
- `go test ./internal/cli -run 'TestRunCommand.*Env.*' -count=1` passes.
- No regressions in existing model/agent-mode routing behavior.

**Risks/Dependencies:**

- Existing runner signatures may need coordinated updates in all agent implementations and tests.
- Secret-safe error messaging must be validated against both stdout/stderr and log outputs.

### Phase 11: End-to-End Coverage and Safety Validation

**Goal:** Add deterministic E2E scenarios for all spec verification paths, including precedence and invalid input behavior.
**Status:** Not started
**Paths:**

- `test/e2e/harness_test.go`
- `test/e2e/config_precedence_test.go`
- `test/e2e/agent_selection_test.go`
- `test/e2e/types_test.go`
- `test/e2e/*env*` (new/updated scenario files)

#### 11.1 Unit and integration-level safety checks

**Paths:**

- `internal/config/config_test.go`
- `internal/cli/run_internal_test.go`
- `internal/agent/agent_test.go`

**Reference pattern:** `internal/config/config_test.go` and `internal/cli/run_internal_test.go` (focused precedence tests)

**Checklist:**

- [x] Verified existing test structure supports focused precedence tests.
- [x] Add parsing/validation tests for valid, invalid, and duplicate `--env` entries.
- [ ] Add config tests for `[env]` TOML decode and overlay merge behavior.
- [ ] Add runner/agent tests asserting effective env propagation to subprocess.
- [ ] Add regression tests ensuring env-override logic does not alter non-env config precedence.

#### 11.2 E2E verification matrix for agent env overrides

**Paths:**

- `test/e2e/` (new and existing test files)

**Reference pattern:** `test/e2e/config_precedence_test.go` and `test/e2e/logging_flags_test.go` (scenario matrix style)

**Checklist:**

- [x] Verified harness supports per-test environment injection and deterministic fixture agent behavior.
- [ ] Add `--env` flag-only scenario (`ralph --env FOO=bar ...`) proving child process receives value.
- [ ] Add config-only `[env]` scenario (`ralph --config ...`) proving table values are applied.
- [ ] Add combined precedence scenario proving flag value overrides config value.
- [ ] Add repeated flag key scenario proving last value wins.
- [ ] Add invalid entry scenario proving failure before agent execution and no value leak.

**Definition of Done:**

- `go test ./test/e2e -run 'TestE2E.*Env.*' -count=1` passes.
- `go test ./test/e2e -count=1` passes without flakiness.
- Verification output confirms no secret-like values are printed in failure messages.

**Risks/Dependencies:**

- Fixture agent may require controlled environment echo behavior to assert child-process env without leaking sensitive data.
- E2E assertions must remain deterministic across platforms and shells.

### Phase 12: Documentation Alignment and Final Quality Gates

**Goal:** Align user-facing docs/spec status with implemented behavior and close quality gates.
**Status:** Not started
**Paths:**

- `specs/agent-env-overrides.md`
- `specs/configuration.md`
- `README.md`
- `examples/ralph.toml`

#### 12.1 Spec and docs sync

**Paths:**

- `specs/agent-env-overrides.md`
- `specs/configuration.md`
- `README.md`
- `examples/ralph.toml`

**Reference pattern:** Existing configuration tables in `README.md` and `specs/configuration.md`

**Checklist:**

- [x] Verified scope spec exists and describes required precedence/validation behaviors.
- [ ] Update configuration tables/docs to include `--env` and `[env]` semantics.
- [ ] Add redacted examples for CLI and TOML usage in docs.
- [ ] Mark scope spec status and verification section as implemented only after code/tests pass.

#### 12.2 Quality gates and release-readiness verification

**Paths:**

- repository-wide (`internal/`, `test/e2e/`, `specs/`, `README.md`)

**Checklist:**

- [ ] Run `make lint` and address findings.
- [ ] Run `make test` and address failures.
- [ ] Run `make test-e2e` and address failures.
- [ ] Run `make quality` as final gate and record result.

**Definition of Done:**

- Required commands complete successfully with results added to the Verification Log.
- Implementation and docs reflect the same precedence and validation behavior.

**Risks/Dependencies:**

- Documentation drift risk if behavior lands without updating tables/examples.
- Need to avoid exposing real secret material in examples and tests.

## Verification Log

- 2026-03-27: `git log --oneline --decorate -n 25 -- specs/agent-env-overrides.md specs/configuration.md specs/run-command.md specs/agents.md specs/e2e-testing.md` - confirmed scope spec was introduced in commit `d3461d1`; related specs updated earlier for config/run/e2e domains; tests run: none (planning mode); bug fixes discovered: none; files reviewed: `specs/agent-env-overrides.md`, `specs/configuration.md`, `specs/run-command.md`, `specs/agents.md`, `specs/e2e-testing.md`.
- 2026-03-27: `git show --oneline --no-color d3461d1 -- specs/agent-env-overrides.md` - verified spec is proposed-only documentation addition (no accompanying code changes); tests run: none; bug fixes discovered: none; files reviewed: `specs/agent-env-overrides.md`.
- 2026-03-27: code search for `--env|[env]|Config.Env|cmd.Env` across `internal/**/*.go` and `test/**/*.go` - no implementation found for repeatable `--env`, config `[env]` field, or explicit `cmd.Env` assignment in agent runner; tests run: none; bug fixes discovered: implementation gap confirmed; files reviewed: `internal/cli/run.go`, `internal/config/config.go`, `internal/agent/runner.go`, `test/e2e/harness_test.go`.
- 2026-03-27: `grep`/read pass on `internal/config/config.go` and `internal/config/config_local_test.go` - confirmed deterministic overlay merge pattern exists for `prompt-overrides` and can be reused for `[env]` merge strategy; tests run: none; bug fixes discovered: none; files reviewed: `internal/config/config.go`, `internal/config/config_local_test.go`.
- 2026-03-27: read pass on `internal/agent/{opencode,claude,cursor}.go` + `internal/agent/runner.go` - verified all agents call a single runner execution path, enabling one-point env propagation; tests run: none; bug fixes discovered: none; files reviewed: `internal/agent/opencode.go`, `internal/agent/claude.go`, `internal/agent/cursor.go`, `internal/agent/runner.go`.
- 2026-03-27: read pass on `test/e2e/harness_test.go`, `test/e2e/config_precedence_test.go`, `test/e2e/agent_selection_test.go` - verified deterministic harness and precedence-test style exist but no agent-env override scenarios are present; tests run: none; bug fixes discovered: none; files reviewed: `test/e2e/harness_test.go`, `test/e2e/config_precedence_test.go`, `test/e2e/agent_selection_test.go`.
- 2026-03-27: `go test ./internal/config -run 'TestLoadConfigEnvTable|TestLoadConfigWithOverlayEnvDeepMerge' -count=1` - failed first due missing `Config.Env` field, confirming scope gap before implementation.
- 2026-03-27: `go test ./internal/config -run 'TestLoadConfig.*Env|TestLoadConfigWithOverlay.*' -count=1` - pass after adding `[env]` decode support and deterministic local overlay merge.
- 2026-03-27: `go test ./internal/config -count=1` - pass; existing config precedence/logging/prompt-override tests remain green.
- 2026-03-27: `go test ./internal/cli -run 'TestReadEnvFlagOverrides' -count=1` - failed first due missing `readEnvFlagOverrides` implementation, confirming Phase 9.2 gap before implementation.
- 2026-03-27: `go test ./internal/cli -run 'TestRead.*Env|TestApply.*Env' -count=1` - pass after adding repeatable `--env` flag parsing/validation and `Config.Env` override application.
- 2026-03-27: `go test ./internal/cli -count=1` - pass; existing run-command behavior remains green after env flag plumbing changes.
- 2026-03-27: `go test ./internal/config -run 'TestLoadConfig.*Env|TestLoadConfigWithOverlay.*' -count=1` - pass; config env decoding/overlay behavior remains stable with CLI updates.
- 2026-03-27: `go test ./internal/config -count=1` - pass; full config package remains green.
- 2026-03-27: `git commit -m "feat(cli): support validated --env overrides for agent config"` - committed Phase 9.2 implementation as `298990d`.
- 2026-03-27: `go test ./internal/cli -run 'TestRunLoop(AppliesEffectiveEnvOverridesToAgentProcess|RejectsInvalidAgentEnvKeyBeforeExecution)' -count=1` - failed first (red) due missing effective env wiring and missing fail-fast key validation in loop path.
- 2026-03-27: `go test ./internal/agent -count=1` - pass after adding effective env builder, deterministic env materialization, and runner `cmd.Env` wiring.
- 2026-03-27: `go test ./internal/cli -run 'TestRunLoop(AppliesEffectiveEnvOverridesToAgentProcess|RejectsInvalidAgentEnvKeyBeforeExecution)' -count=1` - pass after wiring effective env through `RunLoop` and agent factory.
- 2026-03-27: `go test ./internal/cli -run 'TestRunCommand.*Env.*' -count=1` - pass; env-focused run command behavior remains green.
- 2026-03-27: `go test ./internal/cli -count=1` - pass; no regressions in run-loop/config/logging flag behavior.
- 2026-03-27: `go test ./internal/config -count=1` - pass; non-env precedence and config behavior remain stable.
- 2026-03-27: `git commit -m "feat(agent): apply deterministic env overrides to agent subprocesses"` - committed Phase 10 implementation as `9e58d1d`.

## Summary

| Phase    | Goal                                                | Status      |
| -------- | --------------------------------------------------- | ----------- |
| Phase 9  | Config and CLI input surfaces                       | Complete    |
| Phase 10 | Effective environment construction and agent wiring | Complete    |
| Phase 11 | End-to-end coverage and safety validation           | Not started |
| Phase 12 | Documentation alignment and final quality gates     | Not started |

**Remaining effort:** Add Phase 11 e2e env-override coverage matrix and complete Phase 12 docs/quality gates.

## Known Existing Work

- `setupSharedFlags` in `internal/cli/run.go` already centralizes flag registration for root and `run` commands.
- `internal/cli/run.go` now includes repeatable `--env` parsing/validation (`KEY=VALUE`, split-on-first-`=`, key regex guard, last value wins) and applies CLI env overrides to `Config.Env`.
- `internal/agent/runner.go` now centralizes effective env construction from `os.Environ()` plus validated overrides and injects deterministic `cmd.Env`.
- `internal/agent/agent.go` and all concrete agents now receive and pass explicit effective env slices during execution.
- `internal/cli/run.go` now builds effective agent env once per run and fails fast on invalid env keys before starting agent subprocess execution.
- `internal/agent/agent_test.go` and `internal/cli/cmd_test.go` now cover precedence, value preservation (`=` in values), cross-agent env propagation, and redacted invalid-key failures.
- `internal/config/config.go` already implements deterministic precedence and local overlay merge for existing fields.
- `test/e2e/harness_test.go` already builds a deterministic fixture agent and supports per-test environment setup.
- `specs/agent-env-overrides.md` already defines exact precedence and validation expectations for this scope.

## Manual Deployment Tasks

None
