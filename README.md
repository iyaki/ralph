# Ralph Wiggum Agentic Loop Runner

A cross-platform AI agentic loop runner for spec-driven development workflows.

## What Ralph Does

- Runs iterative prompt loops against supported AI CLIs until a completion signal is produced.
- Loads prompts from built-ins, prompt files, stdin, or inline text.
- Applies deterministic configuration precedence across flags, environment variables, config files, local overlays, and prompt front matter.

## Supported Agents

- `opencode` (default)
- `claude`
- `cursor`
- **Codex, Copilot, Gemini, and more agents, coming soon**

Agent selection fallback is deterministic: unknown agent names resolve to `opencode`.

### Adding Support for New Agents

Agent support Pull Requests are always welcomed. To add or update agent integrations, follow the workflow in [`CONTRIBUTING.md` ("Adding Support for a New Agent")](CONTRIBUTING.md#adding-support-for-a-new-agent):

- `agent-spec-creation` for spec definition
- `agent-implementation` for TDD-based code changes

## Installation

### Prerequisites

- A supported agent CLI available in `PATH` (`opencode`, `claude`, or `cursor`)

### Pre-built Binaries

Coming soon.

### From Source

Requires Go `1.25` (see `go.mod`).

```bash
make build
```

Binary output path defaults to `bin/ralph`.

### Install System-Wide

```bash
make install
```

Or manually:

```bash
sudo install -m 0755 bin/ralph /usr/local/bin/ralph
```

### Initialize Config File

```bash
ralph init
```

## Quick Start

If you are already doing spec-driven development:

1. Run `ralph plan my-feature` to generate an implementation plan.
2. Run `ralph` (defaults to `run build`) to start implementing the feature.

## About the Ralph Wiggum Methodology

This implementation is based on the [Ralph Wiggum methodology](https://ghuntley.com/ralph/) pioneered by [Geoffrey Huntley](https://ghuntley.com/).

**Core Principles:**

- **Spec-driven development** - Requirements defined upfront in markdown specs
- **Monolithic operation** - One agent, one task, one loop iteration at a time
- **Fresh context** - Each iteration starts with a clean context window
- **Backpressure** - Tests and validation provide immediate feedback (Architectural constraints of [Harness engineering](https://martinfowler.com/articles/exploring-gen-ai/harness-engineering.html))
- **Let Ralph Ralph** - Trust the agent to self-correct through iteration
- **Disposable plans** - Regenerate implementation plans when they go stale
- **Simple loops** - Minimal Bash loops feeding prompts to AI agents

The methodology works in three phases:

1. Define requirements through human and LLM conversations to create specs
2. Gap analysis to generate/update implementation plans
3. Build loops that implement one task at a time, commit, update the plan and repeat until completion

For a comprehensive guide, see the [Ralph Playbook](https://github.com/ClaytonFarr/ralph-playbook).

This tool implements what has worked well for me, inspired by how Geoffrey worked on [Loom](https://github.com/ghuntley/loom).

This CLI tool alone is not enough to achieve good results. The quality of the prompts and specs, and the backpressure you set, will greatly influence the outcomes.

If you don't know where to start implementing backpressure, [lefthook](https://github.com/evilmartians/lefthook) is a great tool for setting pre-commit hooks.

## Command Reference

Ralph exposes the root command and a `run` subcommand with shared behavior.

```bash
ralph [subcommand] [options] [prompt] [scope]
```

Examples:

```bash
# Run with default build prompt
ralph # Equivalent to `ralph run build`

# Run with plan prompt
ralph plan my-feature

# Use custom max iterations
ralph --max-iterations 10 build

# Use inline prompt
ralph --prompt "Custom prompt text"

# Read prompt from stdin
echo "prompt from stdin" | ralph -

# Use a specific agent and agent mode
ralph --agent claude --agent-mode planner

# Override child agent environment variables
ralph --env OPENAI_API_KEY=<redacted> --env HTTP_PROXY=http://127.0.0.1:8080 build

# Show help
ralph --help
```

## Prompt Sources

Ralph resolves prompt content in this precedence order:

- `--prompt` inline text (highest for prompt content)
- `--prompt-file <path>` (or `-` to read from stdin)
- named prompt resolution (for example `build`, `plan`) from prompt directories

For markdown prompt files, YAML front matter supports runtime overrides for:

- `model`
- `agent-mode`

Front matter is stripped from the prompt body before sending text to the agent process.

## Configuration

Ralph supports flags, environment variables, and TOML config files.

### General Precedence

`flags > env vars > config file values > defaults`

### `model` / `agent-mode` Effective Precedence

1. `--model` / `--agent-mode`
2. `RALPH_MODEL` / `RALPH_AGENT_MODE`
3. prompt file front matter (`model`, `agent-mode`)
4. `[prompt-overrides.<prompt>]` from config
5. global `model` / `agent-mode` in config
6. defaults (empty)

### Agent Process Environment Precedence

1. inherited process environment (`os.Environ()`)
2. config `[env]`
3. repeated `--env KEY=VALUE` flags (highest)

Notes:

- `--env` splits on the first `=`.
- Empty values are valid (`KEY=`).
- Duplicate keys resolve by command-line order (last value wins).

### Config File Selection and Local Overlay

Base config selection order:

1. `--config <path>`
2. `RALPH_CONFIG=<path>`
3. auto-discovery of `ralph.toml` in the current directory.

If a base config is selected and a sibling `ralph-local.toml` exists, it is merged over the base config.

## Flags, Env Vars, and TOML Keys

| Setting                  | Flag                               | Env var                          | TOML key                   | Default                               |
| ------------------------ | ---------------------------------- | -------------------------------- | -------------------------- | ------------------------------------- |
| Config path              | `--config`, `-c`                   | `RALPH_CONFIG`                   | n/a                        | Auto-discover `ralph.toml` in cwd     |
| Max iterations           | `--max-iterations`, `-m`           | `RALPH_MAX_ITERATIONS`           | `max-iterations`           | `25`                                  |
| Prompt file path         | `--prompt-file`, `-p`              | n/a                              | `prompt-file`              | unset                                 |
| Specs dir                | `--specs-dir`, `-s`                | `RALPH_SPECS_DIR`                | `specs-dir`                | `specs`                               |
| Specs index file         | `--specs-index`, `-i`              | `RALPH_SPECS_INDEX_FILE`         | `specs-index-file`         | `README.md`                           |
| Disable specs index      | `--no-specs-index`                 | n/a                              | `no-specs-index`           | `false`                               |
| Implementation plan name | `--implementation-plan-name`, `-n` | `RALPH_IMPLEMENTATION_PLAN_NAME` | `implementation-plan-name` | `IMPLEMENTATION_PLAN.md`              |
| Inline custom prompt     | `--prompt`                         | `RALPH_CUSTOM_PROMPT`            | `custom-prompt`            | unset                                 |
| Log file path            | `--log-file`, `-l`                 | `RALPH_LOG_FILE`                 | `log-file`                 | `<cwd>/ralph.log`                     |
| Disable logging          | `--no-log`                         | `RALPH_LOG_ENABLED=0`            | `no-log`                   | disabled by default (`no-log = true`) |
| Truncate log file        | `--log-truncate`                   | `RALPH_LOG_APPEND=0`             | `log-truncate`             | append mode (`log-truncate = false`)  |
| Prompt templates dir     | none                               | `RALPH_PROMPTS_DIR`              | `prompts-dir`              | `$HOME/.ralph`                        |
| Agent                    | `--agent`, `-a`                    | `RALPH_AGENT`                    | `agent`                    | `opencode`                            |
| Model                    | `--model`                          | `RALPH_MODEL`                    | `model`                    | unset                                 |
| Agent mode               | `--agent-mode`                     | `RALPH_AGENT_MODE`               | `agent-mode`               | unset                                 |
| Agent env overrides      | `--env KEY=VALUE` (repeatable)     | n/a                              | `[env]`                    | inherited process env only            |

## Configuration Examples

Repository baseline:

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

Local overlay (keep untracked):

```toml
# ralph-local.toml
[prompt-overrides.build]
agent-mode = "architect"
```

Child agent env overrides:

```toml
# ralph.toml or ralph-local.toml
[env]
OPENAI_API_KEY = "<redacted>"
ANTHROPIC_API_KEY = "<redacted>"
HTTP_PROXY = "http://127.0.0.1:8080"
```

```bash
ralph --config ./ralph.toml --env OPENAI_API_KEY=<redacted> --env HTTP_PROXY=http://127.0.0.1:8080 build
```

Prompt front matter override:

```md
---
model: claude-sonnet-4
agent-mode: planner
---
```

## Spec Creator Skill

This repo includes the `spec-creator` [skill](https://agentskills.io/home) (see [.agents/skills/spec-creator/SKILL.md](.agents/skills/spec-creator/SKILL.md)) for use in the first phase of the Ralph Wiggum methodology (see [Ralph Methodology section](#about-the-ralph-wiggum-methodology)).

To install it using Vercel's skills CLI, run:

```sh
npx skills add https://github.com/iyaki/ralph/ --skill spec-creator
```

## License

[MIT License](LICENSE.md)
