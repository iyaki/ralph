# Implementation Plan (command-init)

**Status:** Implementation In Progress (2/2)
**Last Updated:** 2026-03-09
**Primary Spec:** [specs/init-command.md](specs/init-command.md)

## Quick Reference

| System            | Spec                                    | Package           | Artifacts   | Implemented? |
| :---------------- | :-------------------------------------- | :---------------- | :---------- | :----------- | --- | -------------- | ------------------------------------- | -------------- | -------- | --- |
| **Config Writer** | [Configuration](specs/configuration.md) | `internal/config` | `writer.go` | [x]          |
| **Init Command**  | [Init Command](specs/init-command.md)   | `internal/cli`    | `init.go`   | [~]          |     | **CLI Wiring** | [Init Command](specs/init-command.md) | `internal/cli` | `cmd.go` | [x] |

## Phased Plan

### Phase 1: Config Writer Implementation

**Goal:** Enable programmatic writing of `ralph.toml` configuration files.
**Paths:** `internal/config/`

#### 1.1 TOML Writer

- [x] Create `internal/config/writer.go`
- [x] Implement `WriteConfig(path string, cfg *Config) error`
- [x] Implement atomic write pattern (write temp -> rename)
- [x] Add `internal/config/writer_test.go` for verification

**Definition of Done:**

- `WriteConfig` correctly serializes `Config` struct to TOML.
- Unit tests pass for writing and overwriting files.

### Phase 2: Init Command Implementation

**Goal:** Implement the interactive `ralph init` command.
**Paths:** `internal/cli/`

#### 2.1 Init Command Structure

- [x] Create `internal/cli/init.go`
- [x] Implement `NewInitCommand() *cobra.Command`
- [x] Define `InitSession`, `InitQuestion`, `InitAnswers` structs (internal to `init.go` or private)

#### 2.2 Interactive Logic & Validation

- [ ] Implement TTY detection
- [ ] Implement question loop with `bufio`
- [ ] Implement validators (Enum, Integer, Non-empty)
- [ ] Implement default seeding from existing config/defaults
- [ ] Implement overwrite confirmation logic

#### 2.3 CLI Wiring

- [x] Register `init` subcommand in `internal/cli/cmd.go` (Update `NewRalphCommand`)

**Definition of Done:**

- `ralph init` runs interactively.
- Input validation works as specified.
- `ralph.toml` is generated correctly.
- Tests in `internal/cli/init_test.go` (if feasible) or manual verification passes.

## Verification Log

| Date       | Verification Step                  | Result |
| :--------- | :--------------------------------- | :----- |
| 2026-03-09 | `go test -v ./internal/config/...` | PASS   |
| 2026-03-09 | `go test -v ./internal/cli/...`    | PASS   |

## Summary

| Phase                  | Status      | Completion |
| :--------------------- | :---------- | :--------- |
| Phase 1: Config Writer | Complete    | 100%       |
| Phase 2: Init Command  | In Progress | 50%        |
| **Remaining Effort**   | **Medium**  | **50%**    |

## Known Existing Work

- `internal/config/config.go`: Existing configuration data model and loader.
- `internal/cli/cmd.go`: Existing root command structure.

## Manual Deployment Tasks

None.
