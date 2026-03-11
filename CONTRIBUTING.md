# Contributing to Ralph

Thanks for your interest in contributing.

## Prerequisites

- Go 1.25 (see `go.mod`)
- `make`
- Tooling used by quality gates:
  - `golangci-lint`
  - `govulncheck`
  - `gosec`
  - `go-arch-lint`
  - `gremlins` (mutation testing, final-stage checks)

## Quickstart

```bash
make deps
make build
make test
```

Run from source:

```bash
make run ARGS='--help'
```

## Build and Run

Build with Make:

```bash
make build
```

Build directly with Go:

```bash
go build -o bin/ralph ./cmd/ralph
```

Cross-compile examples:

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o bin/ralph-linux ./cmd/ralph

# macOS
GOOS=darwin GOARCH=amd64 go build -o bin/ralph-darwin ./cmd/ralph

# Windows
GOOS=windows GOARCH=amd64 go build -o bin/ralph.exe ./cmd/ralph
```

Run from source without building a binary:

```bash
make run ARGS='--help'
```

## Project Structure

```text
.
├── cmd/
│   └── ralph/main.go        # CLI entry point
├── internal/
│   ├── agent/               # Agent implementations
│   ├── cli/                 # Cobra commands
│   ├── config/              # Config loading/writing
│   ├── executor/            # Command execution helpers
│   ├── logger/              # Logging
│   └── prompt/              # Prompt rendering
├── specs/                   # Technical specifications
├── test/e2e/                # End-to-end tests
├── go.mod
├── go.sum
└── Makefile
```

### Agent Files

Each agent implementation is in its own file:

- `internal/agent/agent.go`: Agent interface definition and factory function
- `internal/agent/opencode.go`: Opencode CLI agent implementation
- `internal/agent/claude.go`: Claude Code CLI agent implementation
- `internal/agent/cursor.go`: Cursor CLI agent implementation

This modular design makes it easy to add support for additional AI CLI tools in the future.

## Dependency Management

Add a dependency and clean up module files:

```bash
go get package-name
go mod tidy
```

## Development Workflow

1. Create a branch from `main`.
2. Make focused changes.
3. Run checks locally before opening a PR.

Recommended local checks:

```bash
make lint
make test
make test-race
make coverage
make security
make arch
```

Test command shortcuts:

```bash
make test
make test-e2e
go test -v ./...
go test -v ./test/e2e
```

Use mutation testing only in final validation phases:

```bash
make mutation
```

For one-command verification:

```bash
make quality
```

## Spec-Driven Workflow (Specs-First + TDD)

Ralph uses a spec-driven workflow so contributors can align on intent before implementation details.

1. Start with `specs/README.md` and read the relevant specs for your change.
2. Treat specs as the source of intent, then verify current behavior in the codebase and tests.
3. Keep changes aligned with documented patterns and data shapes from the specs.
4. For code changes, follow test-driven development: write a failing test first, then implement the smallest change to make it pass.
5. Only update specs when behavior is intentionally changing and that change is approved by maintainers.

Practical tips:

- In your PR description, list which spec(s) informed your implementation.
- If implementation and spec disagree, prefer opening a small clarification/update to the spec before broad refactors.
- If your contribution is spec-only, write/update the spec first and stop there (no implementation in the same step).

## Pull Requests

- Open a PR using the provided template.
- Include the problem statement, approach, and test evidence.
- Update docs/specs/changelog when behavior changes.
- Keep PRs small and focused.

## Community Standards

- Read and follow `CODE_OF_CONDUCT.md`.
- For vulnerabilities, use private reporting in `SECURITY.md`.
- For usage help and triage expectations, see `SUPPORT.md`.
