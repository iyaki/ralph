# Configuration

## Overview

### Purpose

- Define how Ralph resolves configuration values across flags, environment variables, config files, and defaults.
- Provide a complete, testable list of configuration options and their sources.

### Goals

- Specify precedence rules and default values.
- Enumerate CLI flags, environment variables, and TOML keys.
- Describe config file discovery and supported filenames.

### Non-Goals

- Defining new configuration options.
- External secret management or vault integrations.

### Scope

- In scope: configuration sources, precedence, defaults, and validation behavior as implemented.
- Out of scope: agent CLI configuration beyond what Ralph passes through.

## Architecture

### Module/package layout (tree format)

```
internal/
  config/
    config.go
internal/
  cli/
    cmd.go
```

### Component diagram (ASCII)

```
+------------------+
|  CLI flags       |
+---------+--------+
          |
          v
+---------+--------+        +------------------+
| Config Loader    |<------>| Env vars         |
| internal/config  |        +------------------+
+---------+--------+        +------------------+
          |---------------->| Config files     |
          |                 +------------------+
          v
+---------+--------+
| Resolved Config  |
+------------------+
```

### Data flow summary

1. CLI flags populate an initial `Config` struct.
2. Loader reads environment variables and config files.
3. Values are resolved by precedence: flags > env vars > config file > defaults.

## Data model

### Core Entities

- Config
  - Fields: `ConfigFile`, `MaxIterations`, `PromptFile`, `SpecsDir`, `SpecsIndexFile`, `NoSpecsIndex`, `ImplementationPlanName`, `LogFile`, `NoLog`, `LogTruncate`, `CustomPrompt`, `PromptsDir`, `AgentName`, `Model`, `AgentMode`.
  - Each field may be set by flag, env var, or config file key.

### Relationships

- `CLI` owns a single `Config` per run.
- `Config` values are resolved once before execution.

### Persistence Notes

| Store | Format | Location | Notes |
| --- | --- | --- | --- |
| Config | TOML | `ralph.toml`, `.ralphrc.toml`, `.ralphrc` | First existing file in this order is used. |

## Workflows

### Load configuration (happy path)

1. CLI parses flags into `Config` fields.
2. Loader reads env vars and config file.
3. Resolver applies precedence rules and defaults.

### Config file resolution

1. If `--config` is provided, that file is used.
2. Otherwise, the loader checks in order: `ralph.toml`, `.ralphrc.toml`, `.ralphrc` (current directory only).
3. The first existing file is parsed as TOML.

## APIs

- None. Configuration is local and file-based.

## Client SDK Design

- Not applicable.

## Configuration

### Precedence

- Flags > environment variables > config file > defaults.

### CLI flags

| Flag | Field | Description |
| --- | --- | --- |
| `--config`, `-c` | `ConfigFile` | Config file to source |
| `--max-iterations`, `-m` | `MaxIterations` | Max iterations |
| `--prompt-file`, `-p` | `PromptFile` | Prompt file path or `-` for stdin |
| `--specs-dir`, `-s` | `SpecsDir` | Specs directory |
| `--specs-index`, `-i` | `SpecsIndexFile` | Specs index file name |
| `--no-specs-index` | `NoSpecsIndex` | Disable specs index file |
| `--implementation-plan-name`, `-n` | `ImplementationPlanName` | Implementation plan file name |
| `--log-file`, `-l` | `LogFile` | Log file path |
| `--no-log` | `NoLog` | Disable logs |
| `--log-truncate` | `LogTruncate` | Truncate log file before writing |
| `--prompt` | `CustomPrompt` | Inline custom prompt |
| `--agent`, `-a` | `AgentName` | Agent name (`opencode`, `claude`, `cursor`) |
| `--model` | `Model` | Model name passed to the agent CLI |
| `--agent-mode` | `AgentMode` | Agent mode/sub-agent passed to the agent CLI |

### Environment variables

| Env var | Field | Notes |
| --- | --- | --- |
| `RALPH_MAX_ITERATIONS` | `MaxIterations` | Integer |
| `RALPH_SPECS_DIR` | `SpecsDir` | String path |
| `RALPH_SPECS_INDEX_FILE` | `SpecsIndexFile` | String file name |
| `RALPH_IMPLEMENTATION_PLAN_NAME` | `ImplementationPlanName` | String file name |
| `RALPH_CUSTOM_PROMPT` | `CustomPrompt` | Inline prompt text |
| `RALPH_LOG_FILE` | `LogFile` | String path |
| `RALPH_LOG_ENABLED` | `NoLog` | `0` disables logging |
| `RALPH_LOG_APPEND` | `LogTruncate` | `0` truncates (disables append) |
| `RALPH_PROMPTS_DIR` | `PromptsDir` | String path |
| `RALPH_AGENT` | `AgentName` | Agent name |
| `RALPH_MODEL` | `Model` | Model name |
| `RALPH_AGENT_MODE` | `AgentMode` | Agent mode |

### Config file keys (TOML)

| Key | Field | Example |
| --- | --- | --- |
| `config-file` | `ConfigFile` | `config-file = "./ralph.toml"` |
| `max-iterations` | `MaxIterations` | `max-iterations = 30` |
| `prompt-file` | `PromptFile` | `prompt-file = "./prompts/plan.md"` |
| `specs-dir` | `SpecsDir` | `specs-dir = "specs"` |
| `specs-index-file` | `SpecsIndexFile` | `specs-index-file = "README.md"` |
| `no-specs-index` | `NoSpecsIndex` | `no-specs-index = true` |
| `implementation-plan-name` | `ImplementationPlanName` | `implementation-plan-name = "IMPLEMENTATION_PLAN.md"` |
| `log-file` | `LogFile` | `log-file = "./ralph.log"` |
| `no-log` | `NoLog` | `no-log = false` |
| `log-truncate` | `LogTruncate` | `log-truncate = false` |
| `custom-prompt` | `CustomPrompt` | `custom-prompt = "..."` |
| `prompts-dir` | `PromptsDir` | `prompts-dir = "./prompts"` |
| `agent` | `AgentName` | `agent = "opencode"` |
| `model` | `Model` | `model = "gpt-4"` |
| `agent-mode` | `AgentMode` | `agent-mode = "planner"` |

### Defaults

| Field | Default |
| --- | --- |
| `MaxIterations` | `25` |
| `SpecsDir` | `specs` |
| `SpecsIndexFile` | `README.md` |
| `ImplementationPlanName` | `IMPLEMENTATION_PLAN.md` |
| `PromptsDir` | `$HOME/.ralph` |
| `LogFile` | `./ralph.log` |
| `AgentName` | `opencode` |
| `Model` | none (optional) |
| `AgentMode` | none (optional) |

## Permissions

- Requires read access for config files and prompt files.
- Requires write access to the log file path when logging is enabled.

## Security Considerations

- Config files may contain sensitive data if inline prompts are used; treat config files as sensitive.
- Environment variables are inherited by child processes; avoid storing secrets in plain env vars when possible.

## Dependencies

| Dependency | Purpose |
| --- | --- |
| `github.com/BurntSushi/toml` | TOML config decoding |

## Open Questions / Risks

- Should config file search include parent directories (currently current directory only)?
- Should invalid `RALPH_MAX_ITERATIONS` values fail fast instead of being ignored?

## Verifications

- `ralph --max-iterations 1 build` uses `1`.
- `RALPH_MAX_ITERATIONS=2 ralph build` uses `2`.
- `ralph --config ./ralph.toml` loads TOML values.
- Default values apply when no flags, env vars, or config files are provided.

## Appendices

- None.
