# Implementation Plan (e2e-testing)

**Status:** Pending (0/21)
**Last Updated:** 2026-03-10
**Primary Spec:** [specs/e2e-testing.md](specs/e2e-testing.md)

## Quick Reference

| System             | Spec                                | Package           | Artifacts           | Implemented? |
| :----------------- | :---------------------------------- | :---------------- | :------------------ | :----------- |
| **E2E Harness**    | [E2E Testing](specs/e2e-testing.md) | `test/e2e`        | `harness_test.go`   | [ ]          |
| **Test Agent**     | [E2E Testing](specs/e2e-testing.md) | `test/e2e/agents` | `ralph-test-agent`  | [ ]          |
| **Test Scenarios** | [E2E Testing](specs/e2e-testing.md) | `test/e2e`        | `scenarios_test.go` | [ ]          |

## Phased Plan

### Phase 1: Test Infrastructure

**Goal:** Establish the test harness, build the test-only agent, and define the test execution logic.
**Paths:** `test/e2e/`

#### 1.1 Test Agent Implementation

- [ ] Create `test/e2e/agents/ralph-test-agent/main.go`.
- [ ] Implement `main` to respect `RALPH_TEST_AGENT_MODE`.
- [ ] Implement `complete_once` mode (emit `<promise>COMPLETE</promise>`).
- [ ] Implement `never_complete` mode (no output).
- [ ] Implement `return_error` mode (exit non-zero).
- [ ] Implement `slow_complete` mode (delay + complete).

#### 1.2 Harness & Types

- [ ] Create `test/e2e/types.go` with `E2ETestCase`, `AgentFixture`, `E2ERunResult` structs.
- [ ] Create `test/e2e/harness_test.go`.
- [ ] Implement `TestMain` to:
  - [ ] Build `ralph` binary to a temp location.
  - [ ] Build `ralph-test-agent` binary to a temp location.
  - [ ] Ensure cleanup of temp binaries on exit.
- [ ] Implement `runTestCase(t *testing.T, tc E2ETestCase)` helper:
  - [ ] Create temp test execution directory.
  - [ ] Write fixture files.
  - [ ] Execute `ralph` with correct `PATH` and environment variables.
  - [ ] Capture output and exit code.

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

| Date | Verification Step | Result |
| :--- | :---------------- | :----- |
|      |                   |        |

## Summary

| Phase                        | Status   | Completion |
| :--------------------------- | :------- | :--------- |
| Phase 1: Test Infrastructure | Pending  | 0%         |
| Phase 2: Core Scenarios      | Pending  | 0%         |
| **Remaining Effort**         | **High** | **100%**   |

## Known Existing Work

- None. This is a new subsystem.

## Manual Deployment Tasks

None.
