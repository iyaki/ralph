# Implementation Plan (config-local)

**Status:** In Progress (5/10)
**Last Updated:** 2026-03-10
**Reference:** `specs/config-local-overlay.md`

## Quick Reference

| System          | Spec                                                       | Package           | Tests                            |
| :-------------- | :--------------------------------------------------------- | :---------------- | :------------------------------- |
| Config Loader   | [config-local-overlay.md](./specs/config-local-overlay.md) | `internal/config` | `internal/config/config_test.go` |
| CLI Integration | [config-local-overlay.md](./specs/config-local-overlay.md) | `internal/cli`    | `test/e2e`                       |

## Phased Plan

### Phase 1: Config Loading & Overlay Logic

**Goal:** Implement discovery, loading, and merging of `ralph-local.toml`.

**Paths:**

- `internal/config/config.go`
- `internal/config/config_test.go`

**Checklist:**

- [x] Refactor `resolveFileConfig` to return base config path alongside the config object
- [x] Implement `resolveLocalOverlayPath(baseConfigPath string) string`
- [x] Implement `mergeConfig(base *Config, overlay *Config) *Config` - [x] Scalar overrides - [x] Map/Table deep merge - [x] Array/List full replacement - [x] `prompt-overrides` deep merge
- [x] Update `LoadConfig` to apply the overlay if present
- [x] Add unit tests for: - [x] Overlay discovery (same dir as base) - [x] Merge semantics (scalars, arrays, tables) - [x] Missing overlay (no error) - [x] Invalid overlay (fail fast)

**Definition of Done:**

- `go test ./internal/config/...` passes.
- Unit tests cover all merge scenarios defined in the spec.

### Phase 2: End-to-End Verification

**Goal:** Verify complete precedence chain and CLI integration.

**Paths:**

- `test/e2e/config_local_test.go` (New)

**Checklist:**

- [ ] Create E2E test: `ralph.toml` + `ralph-local.toml` (happy path)
- [ ] Create E2E test: `RALPH_CONFIG` env var pointing to a dir with both files
- [ ] Create E2E test: Precedence check (Flag > Env > Local > Base)
- [ ] Create E2E test: Array replacement verification
- [ ] Create E2E test: `prompt-overrides` merging verification

**Definition of Done:**

- `make test-e2e` passes.

## Verification Log

| Date       | Verification Step               | Result |
| :--------- | :------------------------------ | :----- |
| 2026-03-10 | `go test ./internal/config/...` | PASS   |

## Summary

| Phase   | Goal                           | Status    |
| :------ | :----------------------------- | :-------- |
| Phase 1 | Config Loading & Overlay Logic | Completed |
| Phase 2 | End-to-End Verification        | Pending   |

**Remaining effort:** Implement E2E tests.

## Known Existing Work

- Basic config loading from single file exists in `internal/config/config.go`.
- `BurntSushi/toml` is already used for parsing.

## Manual Deployment Tasks

None
