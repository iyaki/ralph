# Opencode Agent

Status: Proposed

## Overview

### Purpose

- Define how Ralph executes the `opencode` CLI agent.
- Specify command invocation and argument behavior.

### Goals

- Document invocation format and optional flags.
- Describe availability checks and error handling.

### Non-Goals

- Describing `opencode` internal model behavior.
- Implementing new agent features.

### Scope

- In scope: CLI invocation shape and runtime behavior in Ralph.
- Out of scope: prompt resolution and config precedence.

## Architecture

### Module/package layout (tree format)

```
internal/
  agent/
    opencode.go
```

### Component diagram (ASCII)

```
+------------------+
| OpencodeAgent    |
+---------+--------+
          |
          v
+---------+--------+
| opencode CLI     |
+------------------+
```

### Data flow summary

1. Ralph selects `opencode` when `AgentName` is `opencode` or unknown.
2. The agent builds CLI args based on `Model` and `AgentMode`.
3. The agent executes `opencode run ... <prompt>` and returns output.

## Data model

### Core Entities

- OpencodeAgent
  - Fields: `Model`, `AgentMode`.
  - Implements `Agent` interface.

### Relationships

- Selected by `GetAgent` based on `AgentName`.
- Uses `Model` and `AgentMode` configuration fields.

### Persistence Notes

- None.

## Workflows

### Execute opencode (happy path)

1. Build args: `run`, optional `--model <model>`, optional `--agent <mode>`, `<prompt>`.
2. Execute `opencode` CLI.
3. Stream stdout/stderr and return combined output.

### Execute opencode (error)

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

- Requires OS permission to execute `opencode`.

## Security Considerations

- Prompt text is passed as a CLI argument; sensitive data may appear in process lists.
- Executable resolved from PATH; ensure trusted `opencode` binary.

## Dependencies

- Standard library only (`os/exec`, `bytes`, `io`).

## Open Questions / Risks

- Should missing `opencode` binary fail fast instead of warning?

## Verifications

- `ralph --agent opencode build` invokes `opencode run`.
- `ralph --agent opencode --model gpt-4 build` includes `--model gpt-4`.
- `ralph --agent opencode --agent-mode reviewer build` includes `--agent reviewer`.

## Appendices

### Invocation

```
opencode run [--model <model>] [--agent <mode>] <prompt>
```
