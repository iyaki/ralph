# Implementation Plan (e2e-testing)

**Status:** In Progress (8/21)
**Last Updated:** 2026-03-10
**Primary Spec:** [specs/e2e-testing.md](specs/e2e-testing.md)

## Quick Reference

| System             | Spec                                | Package           | Artifacts           | Implemented? |
| :----------------- | :---------------------------------- | :---------------- | :------------------ | :----------- |
| **E2E Harness**    | [E2E Testing](specs/e2e-testing.md) | `test/e2e`        | `harness_test.go`   | [x]          |
| **Test Agent**     | [E2E Testing](specs/e2e-testing.md) | `test/e2e/agents` | `ralph-test-agent`  | [x]          |
| **Test Scenarios** | [E2E Testing](specs/e2e-testing.md) | `test/e2e`        | `scenarios_test.go` | [x]          |

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

**Goal:** Implement the specific E2E test cases defined in the spec.
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

- [ ] Implement `TestE2ELogging`:
  - [ ] Enable logging via flag/env.
  - [ ] Assert log file creation.
  - [ ] Assert expected log entries exist.

**Definition of Done:**

- All scenarios pass with `go test -v ./test/e2e`.
- Tests are deterministic and clean up artifacts.

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

## Summary

| Phase                        | Status      | Completion |
| :--------------------------- | :---------- | :--------- |
| Phase 1: Test Infrastructure | Done        | 100%       |
| Phase 2: Core Scenarios      | In Progress | 71%        |
| **Remaining Effort**         | **Low**     | **29%**    |

## Known Existing Work

- None. This is a new subsystem.

## Manual Deployment Tasks

None.
