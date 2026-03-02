# Agent Guidelines

## Specifications

**IMPORTANT:** Before implementing any feature, consult the specifications in `specs/README.md`.

- **Assume NOT implemented.** Many specs describe planned features that may not yet exist in the codebase.
- **Check the codebase first.** Before concluding something is or isn't implemented, search the actual code. Specs describe intent; code describes reality.
- **Use specs as guidance.** When implementing a feature, follow the design patterns, types, and architecture defined in the relevant spec.
- **Spec index:** `specs/README.md` lists all specifications organized by category (core, LLM, security, etc.).

## Commands

- Build: `make build` (or `go build -o ralph ./cmd/ralph`).
- Tests: `make test`.
- Coverage gate: `make test-coverage` (fails if coverage < 90%).

## Local Testing

- Run the CLI locally with `./ralph` once built.
- Use `DEBUG=1` to short-circuit the agent loop in `internal/cli` during tests.

## Release Process

- TBD

## Local Testing

## Architecture

- Entry point is `cmd/ralph/main.go` which runs `internal/cli.NewRalphCommand`.
- CLI flow: load config (`internal/config`), init logger (`internal/logger`), load prompt (`internal/prompt`), then `runLoop`.
- The loop replaces `<COMPLETION_SIGNAL>` with `<promise>COMPLETE</promise>` and iterates until the agent output contains it.
- Agents are in `internal/agent`; add new ones by implementing the `Agent` interface and wiring `GetAgent`.

## Prompts and Specs

- Prompt sources are prioritized: inline flag, stdin, explicit prompt file, prompts dir lookup (walks upward), then built-in build/plan prompts.
- Default prompts are generated in `internal/prompt` and reference `specs/` and `IMPLEMENTATION_PLAN.md`.

## Configuration and Logging

- Config precedence: flags > env vars > config file > defaults.
- Config file search order: `ralph.toml` -> `.ralphrc.toml` -> `.ralphrc` in the current directory.
- Logging is controlled by config and `RALPH_LOG_ENABLED` / `RALPH_LOG_APPEND`; log files include git branch/commit headers.

## Code Style

gofmt

## Conventions

- When multiple code paths do similar things with slight variations, create a shared service with a request struct capturing the variations rather than duplicating logic.
- Prefer adding behavior via internal packages (agent, prompt, config, logger) rather than the CLI layer.
