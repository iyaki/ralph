# Implementation Plan (run-command)

**Status:** Proposed
**Last Updated:** 2026-03-10
**Primary Spec:** [specs/run-command.md](specs/run-command.md)

## Quick Reference

| System             | Spec                                | Package        | Artifacts | Implemented? |
| :----------------- | :---------------------------------- | :------------- | :-------- | :----------- |
| **Run Command**    | [Run Command](specs/run-command.md) | `internal/cli` | `run.go`  | ⬜           |
| **Command Router** | [Run Command](specs/run-command.md) | `internal/cli` | `cmd.go`  | ⬜           |
| **Legacy Support** | [Run Command](specs/run-command.md) | `internal/cli` | `cmd.go`  | ✅           |

## Phased Plan

### Phase 1: Explicit Run Command

**Goal:** Implement the explicit `ralph run` command and extract the execution loop logic.
**Paths:** `internal/cli/`

#### 1.1 Refactor Execution Loop

- [ ] Move `RunLoop` and `hasCompletionSignal` from `internal/cli/cmd.go` to `internal/cli/run.go`.
- [ ] Move `applyEffectiveSettings`, `applyModelSettings`, `applyAgentModeSettings` to `internal/cli/run.go` (or a shared utility if needed, but likely `run.go` is fine as they are specific to the run loop).
- [ ] Ensure `RunLoop` signature remains compatible or update call sites (`cmd.go`, tests).

#### 1.2 Implement Run Command

- [ ] Create `NewRunCommand()` in `internal/cli/run.go`.
- [ ] Configure `Use: "run [prompt] [scope]"`.
- [ ] Register all flags currently on the root command to the `run` command (they must be available to both).
- [ ] Implement `RunE` for `run` command to:
  - Parse args (prompt, scope).
  - Load config.
  - Initialize logger.
  - Get prompt.
  - Apply settings.
  - Call `RunLoop`.

**Definition of Done:**

- `ralph run` exists and works identical to current `ralph` command.
- Unit tests for `RunLoop` pass in their new location.

### Phase 2: Root Command Routing

**Goal:** Update the root command to act as a router/dispatcher while preserving backward compatibility.
**Paths:** `internal/cli/cmd.go`

#### 2.1 Update Root Command

- [ ] Register `NewRunCommand()` as a subcommand of the root command.
- [ ] Modify `NewRalphCommand`'s `RunE` to:
  - If no args: Default to `run build` (invoke run logic).
  - If args present (and not caught by subcommands): Treat as `run <args>` (invoke run logic).
- [ ] Ensure `ralph init` still works (handled by Cobra automatically).
- [ ] Ensure `ralph run init` works (handled by `run` subcommand, treats "init" as prompt name).

#### 2.2 Shared Flags

- [ ] Refactor flag setup so common flags are available to both root (for alias usage) and `run` command.

**Definition of Done:**

- `ralph` (no args) executes `run build`.
- `ralph my-prompt` executes `run my-prompt`.
- `ralph run my-prompt` executes `run my-prompt`.
- `ralph init` executes init command.

### Phase 3: Verification & Cleanup

**Goal:** Verify all routing scenarios and collision rules.
**Paths:** `internal/cli/`, `test/e2e/`

#### 3.1 Unit Tests

- [ ] Update `internal/cli/cmd_test.go` to test routing logic.
- [ ] Add `internal/cli/run_test.go` for specific `run` command tests.

#### 3.2 E2E Verification

- [ ] Verify `ralph run build` works.
- [ ] Verify `ralph` works (defaults to build).
- [ ] Verify `ralph init` works.
- [ ] Verify `ralph run init` works (runs prompt "init", does not trigger init command).

**Definition of Done:**

- All specs scenarios verified.
- Tests pass.

## Verification Log

| Date | Verification Step | Result |
| :--- | :---------------- | :----- |
|      |                   |        |

## Summary

| Phase                           | Status     | Completion |
| :------------------------------ | :--------- | :--------- |
| Phase 1: Explicit Run Command   | Pending    | 0%         |
| Phase 2: Root Command Routing   | Pending    | 0%         |
| Phase 3: Verification & Cleanup | Pending    | 0%         |
| **Remaining Effort**            | **Medium** | **0%**     |

## Known Existing Work

- `RunLoop` logic exists in `internal/cli/cmd.go` (needs refactoring).
- `NewInitCommand` exists and works.
- `NewRalphCommand` exists but implements the logic directly.

## Manual Deployment Tasks

None.
