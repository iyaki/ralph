# End-to-End Testing Suite

## Overview

### Purpose

- Define an extensive, deterministic, end-to-end (e2e) test suite for Ralph CLI behavior.
- Ensure user-visible CLI workflows are validated against real process execution boundaries.

### Goals

- Validate complete CLI flows from process startup through exit code and output.
- Cover happy path, failure path, and edge-case behavior across prompt, config, logging, and agent execution.
- Keep tests reproducible in local and CI environments.
- Standardize all e2e scenarios on a single custom test-only agent.

### Non-Goals

- Re-testing internal unit-level logic already covered by package tests.
- Validating third-party agent quality or model response correctness.
- Running destructive or network-dependent scenarios as part of the core e2e suite.
- Running e2e cases separately against each production agent implementation.

### Scope

- In scope: CLI invocation behavior, file-based config/prompt loading, completion loop behavior, logging side effects, and command exit semantics.
- In scope: one dedicated test-only agent used by every e2e scenario.
- Out of scope: upstream agent provider reliability, remote API latency profiling, and long-running soak/perf tests.

## Architecture

### Module/package layout (tree format)

```
cmd/
  ralph/
internal/
  cli/
test/
  e2e/
    fixtures/
    agents/
    testdata/
```

### Component diagram (ASCII)

```
+---------------------+
| e2e test harness    |
| go test ./test/e2e  |
+----------+----------+
           |
           v
+----------+----------+
| ralph CLI process   |
| cmd/ralph main      |
+---+-------------+---+
    |             |
    v             v
+---+---+      +--+----------------+
| temp  |      | test-only agent    |
| files |      | (single fixture)   |
+-------+      +--------------------+
```

### Data flow summary

1. Test harness creates an isolated temporary workspace and fixture files.
2. Harness sets deterministic environment variables and PATH entry for one test-only agent binary.
3. Harness executes Ralph as an external process with explicit args.
4. Harness captures stdout, stderr, exit code, and filesystem side effects.
5. Assertions validate behavior against expected e2e outcomes.

## Data model

### Core Entities

- E2ETestCase
  - `Name` (string): human-readable scenario id.
  - `Args` ([]string): CLI arguments passed to Ralph.
  - `Env` (map[string]string): environment overrides for the process.
  - `Files` (map[string]string): fixture files to create before execution.
  - `ExpectedExitCode` (int): process exit code expectation.
  - `ExpectedStdoutContains` ([]string): required stdout fragments.
  - `ExpectedStderrContains` ([]string): required stderr fragments.
  - `ExpectedFiles` ([]string): files expected to exist after execution.
  - `ForbiddenOutput` ([]string): strings that must not appear.

- AgentFixture
  - `Name` (string): fixed test-only agent identifier.
  - `Behavior` (enum): `complete_once`, `never_complete`, `return_error`, `slow_complete`.
  - `ScriptPath` (string): executable test-only script path added to PATH.

- E2ERunResult
  - `ExitCode` (int).
  - `Stdout` (string).
  - `Stderr` (string).
  - `DurationMs` (int64).
  - `CreatedFiles` ([]string).

### Relationships

- One `E2ETestCase` executes one Ralph process.
- All `E2ETestCase` entries bind the same `AgentFixture` binary.
- One process run yields one `E2ERunResult` used for assertions.

### Persistence Notes

- e2e artifacts are temporary and must be created under OS temp directories.
- On failure, harness may preserve artifacts for inspection under a known `test/e2e/.artifacts` location.
- No artifact is committed to git.

## Workflows

### Suite bootstrap

1. Build Ralph binary once for the suite or run from source using `go test` subprocess strategy.
2. Verify the single test-only agent executable is present and executable.
3. Create per-test temporary directory and fixture file set.

### Happy path: completion detected

1. Configure the test-only agent in `complete_once` mode.
2. Execute Ralph with a known prompt source.
3. Assert zero exit code.
4. Assert completion signal is observed in streamed output.

### Failure path: max iterations reached

1. Configure the test-only agent in `never_complete` mode.
2. Run Ralph with low `MaxIterations` for fast execution.
3. Assert non-zero exit code.
4. Assert warning/error indicates max-iteration exhaustion.

### Failure path: prompt file missing

1. Invoke Ralph with `--prompt-file` pointing to a non-existent path.
2. Assert non-zero exit code.
3. Assert stderr includes prompt file read failure.

### Logging validation

1. Run with logging enabled and explicit log path.
2. Assert process completes according to fixture behavior.
3. Assert log file exists and includes expected high-level entries.
4. Assert sensitive match text is not printed when scenarios include scan outputs.

## Configuration

| Key/Flag                | Purpose                        | Default in e2e suite                      |
| ----------------------- | ------------------------------ | ----------------------------------------- |
| `DEBUG`                 | Optional single-iteration mode | `0` unless explicitly tested              |
| `RALPH_CONFIG`          | Config file path override      | Per-test temp file                        |
| `--max-iterations`      | Bound loop attempts            | Scenario-specific (often low)             |
| `--prompt-file`         | Explicit prompt source         | Scenario-specific                         |
| `--agent`               | Select test-only agent name    | Fixed for all e2e scenarios               |
| `RALPH_TEST_AGENT_MODE` | Control test-only behavior     | Scenario-specific (`complete_once`, etc.) |

## Permissions

- Requires permission to execute subprocesses.
- Requires permission to create temporary files/directories.
- Requires permission to mark fixture scripts as executable.

## Security Considerations

- Fixture inputs must avoid real secrets.
- Tests must avoid printing secret-like content in logs or snapshots.
- PATH manipulation is test-local and must not leak to global environment.

## Dependencies

| Dependency       | Purpose                    |
| ---------------- | -------------------------- |
| Go `testing`     | Test runner and assertions |
| Go `os/exec`     | CLI subprocess execution   |
| Go `t.TempDir()` | Isolated filesystem state  |

## Open Questions / Risks

- Should nightly CI include slower long-loop e2e stress scenarios?
- Do we want golden-output snapshots, or substring-based assertions only?

## Verifications

- `go test ./test/e2e -run TestE2ECompletionFlow` passes with deterministic output.
- `go test ./test/e2e -run TestE2EMaxIterations` fails correctly when completion is absent.
- `go test ./test/e2e -run TestE2EMissingPromptFile` returns non-zero exit behavior as specified.
- `go test ./test/e2e -run TestE2ELogging` confirms log file creation and expected entries.
- Full suite execution is stable across repeated runs with no flaky failures.

## Appendices

### Scenario matrix (minimum coverage)

| Area              | Scenario                      | Expected Result                   |
| ----------------- | ----------------------------- | --------------------------------- |
| Prompt resolution | Inline prompt                 | Prompt used and process runs      |
| Prompt resolution | Stdin prompt                  | Stdin consumed and process runs   |
| Prompt resolution | Prompt file                   | File loaded and process runs      |
| Prompt resolution | Missing prompt file           | Non-zero exit + clear error       |
| Config loading    | Valid config file             | Config applied                    |
| Config loading    | Invalid config file           | Non-zero exit + parse error       |
| Agent execution   | Test-only agent + completion  | Success exit                      |
| Agent execution   | Test-only agent returns error | Failure path validated            |
| Loop behavior     | Completion on first iteration | Early success                     |
| Loop behavior     | Completion on later iteration | Success before max                |
| Loop behavior     | Never completes               | Max-iteration failure             |
| Logging           | Logging enabled               | Log file contains expected events |
| Logging           | Logging disabled              | No log file created               |
| Debug mode        | `DEBUG=1` enabled             | Single-iteration behavior         |

### Determinism rules

- No external network dependencies in required e2e scenarios.
- No time-sensitive assertions without fixed tolerances.
- No dependence on global user config or home directory files.
- Every scenario sets explicit env and files needed for execution.
- Every scenario uses the same custom test-only agent; no per-agent matrix.

### Test-only agent contract

- Binary name: stable and reserved for tests (for example, `ralph-test-agent`).
- Input interface: accepts prompt text from Ralph exactly as production agents do.
- Output interface: prints deterministic fixture output and optional completion signal.
- Mode selection: controlled by `RALPH_TEST_AGENT_MODE` with values:
  - `complete_once`: emits `<promise>COMPLETE</promise>` on first run.
  - `never_complete`: emits no completion signal.
  - `return_error`: exits non-zero with deterministic stderr.
  - `slow_complete`: delays deterministically, then emits completion signal.
- Stability requirement: no network calls, no randomness, no dependency on host user configuration.
