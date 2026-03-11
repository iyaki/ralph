# Implementation Plan (Logging)

**Status:** In Progress (Logging defaults alignment underway)
**Last Updated:** 2026-03-11
**Reference:** `specs/logging.md`

## Quick Reference

| System             | Spec                             | Package           | Tests                            |
| :----------------- | :------------------------------- | :---------------- | :------------------------------- |
| Logger Logic       | [logging.md](./specs/logging.md) | `internal/logger` | `internal/logger/logger_test.go` |
| Config Integration | [logging.md](./specs/logging.md) | `internal/config` | `internal/config/config_test.go` |
| CLI Integration    | [logging.md](./specs/logging.md) | `internal/cli`    | `test/e2e`                       |

## Phased Plan

### Selected Task (This Run)

**Task:** Preserve explicit boolean logging flag overrides during config loading.

**Why this is most important now:**

- The Phase 1 definition of done requires explicit CLI enablement behavior to work.
- Current failures showed explicit boolean false values were being dropped during config resolution.
- This breaks deterministic precedence for the logging controls in real CLI usage.

### Phase 1: Configuration Defaults Alignment

**Goal:** Ensure logging is disabled by default as per spec, and configuration overrides work correctly.

**Paths:**

- `internal/config/config.go`
- `internal/cli/run.go`
- `internal/logger/logger.go`

**Checklist:**

- [x] Preserve explicit boolean logging flag values by reading changed bool flags before config load and re-applying after config resolution.
- [ ] Verify `NoLog` default value logic (Currently enabled by default, Spec says disabled)
- [ ] Update `resolveBool` or defaults in `internal/config/config.go` if necessary to match "Disabled by default"
- [ ] Verify `RALPH_LOG_ENABLED` env var precedence overrides config defaults
- [ ] Verify `LogTruncate` defaults to `false` (Append mode)

**Definition of Done:**

- Running `ralph` without flags/config does NOT create `ralph.log`.
- Running `ralph --no-log=false` (or equivalent enablement) creates `ralph.log`.
- `RALPH_LOG_ENABLED=1` enables logging.

### Phase 2: End-to-End Verification

**Goal:** Verify file creation, permissions, headers, and content.

**Paths:**

- `test/e2e/logging_test.go` (New or existing)

**Checklist:**

- [ ] Create/Update E2E test for logging lifecycle:
  - [ ] Default state (No log file)
  - [ ] Enabled via Env (`RALPH_LOG_ENABLED=1`)
  - [ ] Enabled via Config (`no-log = false`)
  - [ ] File creation at `ralph.log` (default) or custom path
  - [ ] Header presence (Timestamp, Git metadata)
  - [ ] File content matches stdout (via MultiWriter)
  - [ ] File permissions (`0600`)
- [ ] Verify Truncate vs Append behavior (`RALPH_LOG_APPEND=0`)

**Definition of Done:**

- `make test-e2e` passes.
- Log file behaviors confirmed on disk.

## Verification Log

2026-03-11: `go test ./internal/cli -run 'TestRunCommandNoLogFalseFlagOverridesConfig|TestRunCommandNoLogFalseFlagOverridesEnv' -count=1` - failed initially, confirming explicit false flag values were not preserved through config load.
2026-03-11: `go test ./internal/cli -run 'TestRunCommandNoLogFalseFlagOverridesConfig|TestRunCommandNoLogFalseFlagOverridesEnv|TestRunCommandNoLogFlagTracksExplicitFalse' -count=1` - pass.
2026-03-11: `go test ./test/e2e -run 'TestE2ELoggingFlags/(NoLogFalseOverridesConfig|NoLogFalseOverridesEnv)' -count=1` - pass.

## Summary

| Phase   | Goal                             | Status      |
| :------ | :------------------------------- | :---------- |
| Phase 1 | Configuration Defaults Alignment | In Progress |
| Phase 2 | End-to-End Verification          | Pending     |

**Remaining effort:** Fix default `NoLog` value and add E2E tests.

## Known Existing Work

- `internal/logger` package is fully implemented (`NewLogger`, `openLogFile`, `getGitBranch`).
- `internal/config` includes `LogFile`, `NoLog`, `LogTruncate` fields.
- `internal/cli` integrates logger via `io.MultiWriter`.
- Unit tests exist in `internal/logger/logger_test.go`.

## Manual Deployment Tasks

None
