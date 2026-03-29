# Cursor Agent

Status: Implemented

## Overview

### Purpose

- Define how Ralph executes the Cursor CLI agent.
- Specify command invocation and argument behavior.

### Goals

- Document invocation format and optional flags.
- Describe availability checks and error handling.

### Non-Goals

- Describing Cursor model behavior or API semantics.
- Implementing new agent features.

### Scope

- In scope: CLI invocation shape and runtime behavior in Ralph.
- Out of scope: prompt resolution and config precedence.

## Architecture

### Module/package layout (tree format)

```
internal/
  agent/
    cursor.go
```

### Component diagram (ASCII)

```
+------------------+
| CursorAgent      |
+---------+--------+
          |
          v
+---------+--------+
| cursor CLI       |
+------------------+
```

### Data flow summary

1. Ralph selects `cursor` when `AgentName` is `cursor`.
2. The agent builds CLI args based on `Model`.
3. The agent executes `cursor ... <prompt>` and returns output.

## Data model

### Core Entities

- CursorAgent
  - Fields: `Model`, `AgentMode` (unused by this agent).
  - Implements `Agent` interface.

### Relationships

- Selected by `GetAgent` based on `AgentName`.
- Uses `Model` configuration field; `AgentMode` is ignored.

### Persistence Notes

- None.

## Workflows

### Execute cursor (happy path)

1. Build args: optional `--model <model>`, `<prompt>`.
2. Execute `cursor` CLI.
3. Stream stdout/stderr and return combined output.

### Execute cursor (error)

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

- Requires OS permission to execute `cursor`.

## Security Considerations

- Prompt text is passed as a CLI argument; sensitive data may appear in process lists.
- Executable resolved from PATH; ensure trusted `cursor` binary.

## Dependencies

- Standard library only (`os/exec`, `bytes`, `io`).

## Open Questions / Risks

- Should `AgentMode` be rejected when using `cursor` to avoid confusion?

## Verifications

- `ralph --agent cursor build` invokes `cursor`.
- `ralph --agent cursor --model gpt-4 build` includes `--model gpt-4`.

## Appendices

### Invocation

```
cursor [--model <model>] <prompt>
```
