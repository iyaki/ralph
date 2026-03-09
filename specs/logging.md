# Logging

Status: Implemented

## Overview

### Purpose

- Define how Ralph initializes logging and writes logs to disk.
- Provide a testable description of log enablement, file selection, and headers.

### Goals

- Specify when logging is enabled or disabled.
- Describe log file creation, append/truncate behavior, and headers.
- Document git metadata capture in log headers.

### Non-Goals

- Remote log shipping or structured logging.
- Log rotation or retention policies.

### Scope

- In scope: log enablement, file creation, headers, and lifecycle.
- Out of scope: agent output semantics and prompt content rules.

## Architecture

### Module/package layout (tree format)

```
internal/
  logger/
    logger.go
```

### Component diagram (ASCII)

```
+------------------+
| Logger           |
| internal/logger  |
+---------+--------+
          |
          v
+---------+--------+
| Log File (disk)  |
+------------------+
```

### Data flow summary

1. CLI initializes a logger after configuration is loaded.
2. If logging is enabled, the logger opens/creates a log file.
3. A run header and git metadata are written at startup.
4. CLI writes output to stdout and the log file via a multi-writer.

## Data model

### Core Entities

- Logger
  - Fields: `enabled`, `file`.
  - Responsibilities: manage log file lifecycle and header writing.

- LogFile
  - Fields: `Path`, `Append` (bool).
  - Derived from config fields and environment variables.

### Relationships

- Logger behavior depends on configuration fields `NoLog`, `LogFile`, and `LogTruncate`.
- Environment variables can disable logging or force truncation.

### Persistence Notes

- Log output is stored as plain text on disk.

| Store | Format     | Location                 | Notes                                                      |
| ----- | ---------- | ------------------------ | ---------------------------------------------------------- |
| Logs  | Plain text | `./ralph.log` by default | Header includes timestamp and git metadata when available. |

## Workflows

### Initialize logging (enabled)

1. Evaluate config and env to determine logging enabled/disabled.
2. Determine log file path; if empty, create a temp file.
3. Create log directory if it does not exist.
4. Open file in append or truncate mode.
5. Write header with timestamp and git branch/commit.

### Initialize logging (disabled)

1. If `NoLog` is true or `RALPH_LOG_ENABLED=0`, logging is disabled.
2. Logger returns without a file.

### Close logging

1. On CLI exit, logger closes the file if present.

## APIs

- None. Logging is internal.

## Client SDK Design

- Not applicable.

## Configuration

- See configuration spec for option definitions and precedence.
- Relevant fields:
  - `NoLog`
  - `LogFile`
  - `LogTruncate`
  - Env vars: `RALPH_LOG_ENABLED`, `RALPH_LOG_APPEND`

## Permissions

- Requires write access to log file path.
- Requires directory creation permissions for the log file folder.

## Security Considerations

- Logs may contain prompt text and agent outputs; treat log files as sensitive.
- Log file paths should avoid world-writable directories to reduce tampering risk.

## Dependencies

- Standard library only (`os`, `os/exec`, `path/filepath`, `time`).

## Open Questions / Risks

- Should log header include the config file path to aid debugging?
- Should logging be disabled by default in CI environments?

## Verifications

- With `RALPH_LOG_ENABLED=0`, no log file is created.
- With `RALPH_LOG_APPEND=0`, log file is truncated on start.
- Log header includes timestamp and git branch/commit when available.

## Appendices

- None.
