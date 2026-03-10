# Implementation Plan (e2e-testing)

**Status:** Core Scenarios Complete (Phase 1 & 2 done), Coverage Expansion In Progress (Phase 3 ~35% done)
**Last Updated:** 2026-03-10
**Primary Spec:** [specs/e2e-testing.md](specs/e2e-testing.md)

## Quick Reference

| System             | Spec                                | Package           | Artifacts           | Implemented? |
| :----------------- | :---------------------------------- | :---------------- | :------------------ | :----------- |
| **E2E Harness**    | [E2E Testing](specs/e2e-testing.md) | `test/e2e`        | `harness_test.go`   | ✅           |
| **Test Agent**     | [E2E Testing](specs/e2e-testing.md) | `test/e2e/agents` | `ralph-test-agent`  | ✅           |
| **Core Scenarios** | [E2E Testing](specs/e2e-testing.md) | `test/e2e`        | `scenarios_test.go` | ✅           |
| **Full Coverage**  | [E2E Testing](specs/e2e-testing.md) | `test/e2e`        | `test/*.go`         | [ ]          |

## Phased Plan

### Phase 1: Test Infrastructure

**Goal:** Establish the test harness, build the test-only agent, and define the test execution logic.
**Paths:** `test/e2e/`

#### 1.1 Test Agent Implementation

- [x] Create `test/e2e/agents/ralph-test-agent/main.go`.
- [x] Implement `main` to respect `RALPH_TEST_AGENT_MODE`.
- [x] Implement `complete_once` mode (emit `<promise>COMPLETE</promise>`).
- [x] Implement `never_complete` mode (no output).
- [x] Implement `return_error` mode (exit non-zero).
- [x] Implement `slow_complete` mode (delay + complete).

#### 1.2 Harness & Types

- [x] Create `test/e2e/types.go` with `E2ETestCase`, `AgentFixture`, `E2ERunResult` structs.
- [x] Create `test/e2e/harness_test.go`.
- [x] Implement `TestMain` to:
  - [x] Build `ralph` binary to a temp location.
  - [x] Build `ralph-test-agent` binary to a temp location.
  - [x] Ensure cleanup of temp binaries on exit.
- [x] Implement `runTestCase(t *testing.T, tc E2ETestCase)` helper:
  - [x] Create temp test execution directory.
  - [x] Write fixture files.
  - [x] Execute `ralph` with correct `PATH` and environment variables.
  - [x] Capture output and exit code.

**Definition of Done:**

- `go test ./test/e2e` can compile and run (even if empty tests).
- Test agent and Ralph binaries are successfully built during test setup.

### Phase 2: Core Scenarios

**Goal:** Implement the primary happy/failure paths defined in the spec.
**Paths:** `test/e2e/`

#### 2.1 Happy Path

- [x] Create `test/e2e/scenarios_test.go`.
- [x] Implement `TestE2ECompletionFlow`:
  - [x] Configure `complete_once` agent.
  - [x] Run with valid prompt.
  - [x] Assert zero exit code and completion signal.

#### 2.2 Failure Paths

- [x] Implement `TestE2EMaxIterations`:
  - [x] Configure `never_complete` agent.
  - [x] Run with low `--max-iterations`.
  - [x] Assert non-zero exit code.
- [x] Implement `TestE2EMissingPromptFile`:
  - [x] Run with non-existent `--prompt-file`.
  - [x] Assert non-zero exit code and error message.

#### 2.3 Logging

- [x] Implement `TestE2ELogging`:
  - [x] Enable logging via flag/env.
  - [x] Assert log file creation.
  - [x] Assert expected log entries exist.

**Definition of Done:**

- All scenarios pass with `go test -v ./test/e2e`.
- Tests are deterministic and clean up artifacts.

### Phase 3: Comprehensive Coverage

**Goal:** Extend suite to cover ALL supported CLI options, config keys, and output channels as required by spec.
**Paths:** `test/e2e/`

#### 3.1 Prompt Resolution Coverage

- [x] Implement `TestE2EInlinePrompt`:
  - [x] Use `--prompt "custom prompt"`.
  - [x] Assert agent receives the inline prompt.
- [x] Implement `TestE2EStdinPrompt`:
  - [x] Pipe prompt via stdin (`-` argument or implicit).
  - [x] Assert agent receives the stdin prompt.

#### 3.2 Extended CLI Flags Coverage

- [x] Implement `TestE2ESpecsFlags`:
  - [x] Test `--specs-dir` and `--specs-index`.
  - [x] Test `--no-specs-index`.
- [x] Implement `TestE2EPlanFlags`:
  - [x] Test `--implementation-plan-name`.
- [ ] Implement `TestE2ELoggingFlags`:
  - [ ] Test `--no-log`.
  - [ ] Test `--log-truncate`.
- [ ] Implement `TestE2EModelFlags`:
  - [ ] Test `--model` override.
  - [ ] Test `--agent-mode` override.

#### 3.3 Configuration Precedence

- [ ] Implement `TestE2EConfigPrecedence`:
  - [ ] Set conflicting values in Config File, Env Var, and CLI Flag.
  - [ ] Assert CLI Flag wins.
  - [ ] Assert Env Var wins over Config File.

**Definition of Done:**

- Every flag in `ralph --help` has a corresponding e2e test case.
- Prompt loading from all sources (file, inline, stdin) is verified.
- Configuration precedence rules are verified.

## Verification Log

| Date       | Verification Step                                                | Result |
| :--------- | :--------------------------------------------------------------- | :----- |
| 2026-03-10 | `go build ... && RALPH_TEST_AGENT_MODE=complete_once ...`        | Passed |
| 2026-03-10 | `go build ... && RALPH_TEST_AGENT_MODE=never_complete ...`       | Passed |
| 2026-03-10 | `go build ... && RALPH_TEST_AGENT_MODE=return_error ...`         | Passed |
| 2026-03-10 | `go build ... && RALPH_TEST_AGENT_MODE=slow_complete ...`        | Passed |
| 2026-03-10 | `go test ./test/e2e/agents/ralph-test-agent/... && lint && arch` | Passed |
| 2026-03-10 | `go test ./test/e2e` (validates TestMain & build process)        | Passed |
| 2026-03-10 | `go test ./test/e2e -run TestE2ECompletionFlow`                  | Passed |
| 2026-03-10 | `go test ./test/e2e -run TestE2EMaxIterations`                   | Passed |
| 2026-03-10 | `go test ./test/e2e -run TestE2EMissingPromptFile`               | Passed |
| 2026-03-10 | `go test ./test/e2e -run TestE2ELogging`                         | Passed |
| 2026-03-10 | `go test -v ./test/e2e` (all scenarios)                          | Passed |
| 2026-03-10 | `go test ./test/e2e -run TestE2E.*Prompt`                        | Passed |
| 2026-03-10 | `go test ./test/e2e -run TestE2ESpecsFlags`                      | Passed |
| 2026-03-10 | `go test ./test/e2e -run TestE2EPlanFlags`                       | Passed |

## Summary

| Phase                           | Status      | Completion |
| :------------------------------ | :---------- | :--------- |
| Phase 1: Test Infrastructure    | Done        | 100%       |
| Phase 2: Core Scenarios         | Done        | 100%       |
| Phase 3: Comprehensive Coverage | In Progress | 35%        |
| **Remaining Effort**            | **Medium**  | **15%**    |

## Known Existing Work

- `test/e2e/harness_test.go`: Complete harness implementation.
- `test/e2e/scenarios_test.go`: Basic scenarios implemented.
- `test/e2e/prompt_test.go`: Prompt resolution scenarios.
- `test/e2e/agents/ralph-test-agent/`: Fully functional test agent.

## Manual Deployment Tasks

None.
