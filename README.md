# Ralph Go CLI Implementation

Multi-platform AI Agentic Loop runner aimed for Spec-Driven Development workflows.

## Features

- **Native Binary**: Compiled Go application for better performance and portability
- **Cross-Platform**: Can be built for Linux, macOS, Windows, and other platforms

## Workflow

If you have are already making Spec-Driven Development:

1. Running `ralph plan my-feature` to generate an implementation plan
2. Executing `ralph` (defaults to `build`) to start implementing the feature

## Installation

### System-Wide Installation

```bash
make install
```

Or manually:

```bash
sudo install -m 0755 bin/ralph /usr/local/bin/ralph
```

### Local Installation

Simply copy the built `bin/ralph` binary to a directory in your PATH.

### Spec Creator skill

This repo includes the `spec-creator` [skill](https://agentskills.io/home) (see [.agents/skills/spec-creator/SKILL.md](.agents/skills/spec-creator/SKILL.md)) for usage in the first phase of the Ralph Wiggum methodology (see [below](#about-the-ralph-wiggum-methodology)).


To install it using Vercel's skills CLI, run:

```sh
npx skills add https://github.com/iyaki/ralph/ --skill spec-creator
```

## Usage

```bash
# Run with default build prompt
ralph

# Run with plan prompt
ralph plan my-feature

# Use custom max iterations
ralph --max-iterations 10 build

# Use inline prompt
ralph --prompt "Custom prompt text"

# Read prompt from stdin
echo "prompt from stdin" | ralph -

ralph --agent claude --agent-mode planner

# Show help
ralph --help
```

## AI Agent Support

Ralph supports multiple AI CLI agents. Each agent has its own implementation in a separate file:

- **opencode** (default): Uses the `opencode` CLI tool
- **claude**: Uses the `claude` CLI tool (Claude Code CLI)
- **cursor**: Uses the `cursor` CLI tool
- **Codex, Copilot, Gemini and agents, coming soon**

### Selecting an Agent

You can select the agent in three ways:

1. **Command-line flag** (highest priority):

   ```bash
   ralph --agent claude
   ralph --agent cursor
   ```

2. **Environment variable**:

   ```bash
   export RALPH_AGENT=claude
   ralph
   ```

3. **Config file** (`ralph.toml`):
   ```toml
   agent = "claude"
   ```

### Selecting a Model

You can optionally specify which AI model to use with the `--model` flag or `RALPH_MODEL` environment variable:

1. **Command-line flag** (highest priority):

   ```bash
   ralph --agent claude --model claude-sonnet-4
   ralph --agent opencode --model gpt-4
   ```

2. **Environment variable**:

   ```bash
   export RALPH_MODEL=claude-sonnet-4
   ralph --agent claude
   ```

3. **Config file** (`ralph.toml`):
   ```toml
   agent = "claude"
   model = "claude-sonnet-4"
   ```

If no model is specified, the agent will use its default model.

### Selecting a Sub-Agent / Agent Mode

You can optionally select a custom agent mode (sub-agent) for tools that support it:

1. **Command-line flag** (highest priority):

   ```bash
   ralph --agent opencode --agent-mode reviewer
   ralph --agent claude --agent-mode planner
   ```

2. **Environment variable**:

   ```bash
   export RALPH_AGENT_MODE=reviewer
   ralph --agent opencode
   ```

3. **Config file** (`ralph.toml`):
   ```toml
   agent = "claude"
   agent-mode = "planner"
   ```

If no agent mode is specified, the tool's default behavior is used.

Implementation details for agent integrations and internal package layout are documented in `CONTRIBUTING.md`.

## Configuration

Ralph can be configured through command-line flags, environment variables, and TOML files.

### Precedence Rules

General precedence is:

1. command-line flags
2. environment variables
3. config file values
4. built-in defaults

For `model` and `agent-mode`, effective precedence is:

1. `--model` / `--agent-mode`
2. `RALPH_MODEL` / `RALPH_AGENT_MODE`
3. prompt file front matter (`model`, `agent-mode`)
4. `[prompt-overrides.<prompt>]` in config
5. global `model` / `agent-mode` in config
6. defaults (empty)

### Config File Selection

Ralph picks a single base config file in this order:

1. `--config <path>`
2. `RALPH_CONFIG=<path>`

If `ralph-local.toml` exists in the same directory as the selected base file, it is merged on top of the base config.

### Flags, Environment Variables, and TOML Keys

| Setting                  | Flag                               | Env var                          | TOML key                   | Default                                                             |
| ------------------------ | ---------------------------------- | -------------------------------- | -------------------------- | ------------------------------------------------------------------- |
| Config path              | `--config`, `-c`                   | `RALPH_CONFIG`                   | n/a                        | Auto-discover in cwd: `ralph.toml` -> `.ralphrc.toml` -> `.ralphrc` |
| Max iterations           | `--max-iterations`, `-m`           | `RALPH_MAX_ITERATIONS`           | `max-iterations`           | `25`                                                                |
| Prompt file path         | `--prompt-file`, `-p`              | n/a                              | n/a                        | unset                                                               |
| Specs dir                | `--specs-dir`, `-s`                | `RALPH_SPECS_DIR`                | `specs-dir`                | `specs`                                                             |
| Specs index file         | `--specs-index`, `-i`              | `RALPH_SPECS_INDEX_FILE`         | `specs-index-file`         | `README.md`                                                         |
| Disable specs index      | `--no-specs-index`                 | n/a                              | n/a                        | `false`                                                             |
| Implementation plan name | `--implementation-plan-name`, `-n` | `RALPH_IMPLEMENTATION_PLAN_NAME` | `implementation-plan-name` | `IMPLEMENTATION_PLAN.md`                                            |
| Inline custom prompt     | `--prompt`                         | `RALPH_CUSTOM_PROMPT`            | `custom-prompt`            | unset                                                               |
| Log file path            | `--log-file`, `-l`                 | `RALPH_LOG_FILE`                 | `log-file`                 | `<cwd>/ralph.log`                                                   |
| Disable logging          | `--no-log`                         | `RALPH_LOG_ENABLED=0`            | `no-log = true`            | logging enabled (`no-log = false`)                                  |
| Truncate log file        | `--log-truncate`                   | `RALPH_LOG_APPEND=0`             | `log-truncate = true`      | append mode (`log-truncate = false`)                                |
| Prompt templates dir     | none                               | `RALPH_PROMPTS_DIR`              | `prompts-dir`              | `$HOME/.ralph`                                                      |
| Agent                    | `--agent`, `-a`                    | `RALPH_AGENT`                    | `agent`                    | `opencode`                                                          |
| Model                    | `--model`                          | `RALPH_MODEL`                    | `model`                    | unset                                                               |
| Agent mode               | `--agent-mode`                     | `RALPH_AGENT_MODE`               | `agent-mode`               | unset                                                               |

### TOML Examples

Repository baseline config:

```toml
# ralph.toml
agent = "opencode"
model = "gpt-5"
agent-mode = "builder"

max-iterations = 30
specs-dir = "specs"
specs-index-file = "README.md"
implementation-plan-name = "IMPLEMENTATION_PLAN.md"

prompts-dir = ".ralph/prompts"

log-file = "logs/ralph.log"
no-log = false
log-truncate = false
```

Per-prompt overrides:

```toml
# ralph.toml
model = "gpt-5"

[prompt-overrides.plan]
agent-mode = "planner"

[prompt-overrides.build]
model = "gpt-5.3-codex"
agent-mode = "reviewer"
```

Local personal overrides (kept untracked):

```toml
# ralph-local.toml
[prompt-overrides.build]
agent-mode = "architect"
```

Prompt front matter override:

```md
---
model: claude-sonnet-4
agent-mode: planner
---

# Task

...
```

## Contributing

For contributor setup and all development workflows (prerequisites, build/test commands, project structure, implementation details, and spec-driven workflow), see `CONTRIBUTING.md`.

## License

[MIT License](LICENSE.md)
