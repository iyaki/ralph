# Claude Agent

Status: Implemented

## Overview

### Purpose

- Define how Ralph executes the Claude Code CLI agent.
- Specify command invocation and argument behavior.

### Goals

- Document invocation format and optional flags.
- Describe availability checks and error handling.

### Non-Goals

- Describing Claude model behavior or API semantics.
- Implementing new agent features.

### Scope

- In scope: CLI invocation shape and runtime behavior in Ralph.
- Out of scope: prompt resolution and config precedence.

## Architecture

### Module/package layout (tree format)

```
internal/
  agent/
    claude.go
```

### Component diagram (ASCII)

```
+------------------+
| ClaudeAgent      |
+---------+--------+
          |
          v
+---------+--------+
| claude CLI       |
+------------------+
```

### Data flow summary

1. Ralph selects `claude` when `AgentName` is `claude`.
2. The agent builds CLI args based on `Model` and `AgentMode`.
3. The agent executes `claude --dangerously-skip-permissions ... <prompt>` and returns output.

## Data model

### Core Entities

- ClaudeAgent
  - Fields: `Model`, `AgentMode`.
  - Implements `Agent` interface.

### Relationships

- Selected by `GetAgent` based on `AgentName`.
- Uses `Model` and `AgentMode` configuration fields.

### Persistence Notes

- None.

## Workflows

### Execute claude (happy path)

1. Build args: `--dangerously-skip-permissions`, optional `--model <model>`, optional `--agent <mode>`, `<prompt>`.
2. Execute `claude` CLI.
3. Stream stdout/stderr and return combined output.

### Execute claude (error)

1. CLI exits non-zero.
2. Error is returned along with output.

## APIs

- None. This is a local CLI integration.

## Client SDK Design

- Not applicable.

## Configuration

- Relevant fields: `AgentName`, `Model`, `AgentMode`.
- See configuration spec for definitions and precedence.

## Permissions

- Requires OS permission to execute `claude`.

## Security Considerations

- Prompt text is passed as a CLI argument; sensitive data may appear in process lists.
- Executable resolved from PATH; ensure trusted `claude` binary.
- `--dangerously-skip-permissions` is always used; ensure environment is trusted.

## Dependencies

- Standard library only (`os/exec`, `bytes`, `io`).

## Open Questions / Risks

- Should the `--dangerously-skip-permissions` flag be configurable?

## Verifications

- `ralph --agent claude build` invokes `claude --dangerously-skip-permissions`.
- `ralph --agent claude --model claude-sonnet-4 build` includes `--model claude-sonnet-4`.
- `ralph --agent claude --agent-mode planner build` includes `--agent planner`.

## Appendices

### Invocation

```
claude --dangerously-skip-permissions [--model <model>] [--agent <mode>] <prompt>
```
