# Implementation Plan (e2e-testing)

**Status:** In Progress (5/21)
**Last Updated:** 2026-03-10
**Primary Spec:** [specs/e2e-testing.md](specs/e2e-testing.md)

## Quick Reference

| System             | Spec                                | Package           | Artifacts           | Implemented? |
| :----------------- | :---------------------------------- | :---------------- | :------------------ | :----------- |
| **E2E Harness**    | [E2E Testing](specs/e2e-testing.md) | `test/e2e`        | `harness_test.go`   | [x]          |
| **Test Agent**     | [E2E Testing](specs/e2e-testing.md) | `test/e2e/agents` | `ralph-test-agent`  | [x]          |
| **Test Scenarios** | [E2E Testing](specs/e2e-testing.md) | `test/e2e`        | `scenarios_test.go` | [ ]          |

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

- [ ] Create `test/e2e/scenarios_test.go`.
- [ ] Implement `TestE2ECompletionFlow`:
  - [ ] Configure `complete_once` agent.
  - [ ] Run with valid prompt.
  - [ ] Assert zero exit code and completion signal.

#### 2.2 Failure Paths

- [ ] Implement `TestE2EMaxIterations`:
  - [ ] Configure `never_complete` agent.
  - [ ] Run with low `--max-iterations`.
  - [ ] Assert non-zero exit code.
- [ ] Implement `TestE2EMissingPromptFile`:
  - [ ] Run with non-existent `--prompt-file`.
  - [ ] Assert non-zero exit code and error message.

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

## Summary

| Phase                        | Status     | Completion |
| :--------------------------- | :--------- | :--------- |
| Phase 1: Test Infrastructure | Done       | 100%       |
| Phase 2: Core Scenarios      | Pending    | 0%         |
| **Remaining Effort**         | **Medium** | **50%**    |

## Known Existing Work

- None. This is a new subsystem.

## Manual Deployment Tasks

None.
