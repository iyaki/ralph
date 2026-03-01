# Ralph Agentic Loop

POSIX-compliant AI Agentic Loop shell runner aimed for spec-driven development workflows.  
It loads prompts from files (with optional inline overrides) and comes with build/plan presets.

## Installation

Just download the `ralph.sh` script and make it executable:

```sh
curl -fsSLO https://raw.githubusercontent.com/iyaki/ralph/main/ralph.sh
chmod +x ralph.sh
```

### Spec Creator skill

This repo includes the `spec-creator` [skill](https://agentskills.io/home) (see [.agents/skills/spec-creator/SKILL.md](.agents/skills/spec-creator/SKILL.md)) for usage in the first phase of the Ralph Wiggum methodology (see [below](#about-the-ralph-wiggum-methodology)).


To install it using Vercel's skills CLI, run:

```sh
npm skills add https://github.com/iyaki/ralph/ --skill spec-creator
```

## Usage

If you have a `specs/` directory similar to [this one](https://github.com/ghuntley/loom/tree/trunk/specs), using ralph can be as simple as:

1. Running `./ralph.sh plan my-feature` to generate an implementation plan
2. Executing `./ralph.sh` (defaults to `build`) to start implementing the feature

## About the Ralph Wiggum Methodology

This Ralph implementation is based on the [Ralph Wiggum methodology](https://ghuntley.com/ralph/) pioneered by [Geoffrey Huntley](https://ghuntley.com/).

**Core Principles:**
- **Spec-driven development** - Requirements defined upfront in markdown specs
- **Monolithic operation** - One agent, one task, one loop iteration at a time
- **Fresh context** - Each iteration starts with a clean context window
- **Backpressure** - Tests and validation provide immediate feedback (Architectural constraints of [Harness engineering](https://martinfowler.com/articles/exploring-gen-ai/harness-engineering.html))
- **Let Ralph Ralph** - Trust the agent to self-correct through iteration
- **Disposable plans** - Regenerate implementation plans when they go stale
- **Simple loops** - Minimal bash loops feeding prompts to AI agents

The methodology works in three phases:
  1. Define requirements through human+LLM conversation to create specs
  2. Gap analysis to generate/update implementation plans
  3. Build loops that implement one task at a time, commit, and update the plan

For a comprehensive guide, see the [Ralph Playbook](https://github.com/ClaytonFarr/ralph-playbook).

This script implements what has worked well for me, inspired by the work made by Geoffrey at [Loom](https://github.com/ghuntley/loom).

This script alone is not enough to achieve good results. The quality of the prompts and the specs, and the backpressure you set, will greatly influence the outcomes.

If you don't know where to start to implement backpressure, [lefthook](https://github.com/evilmartians/lefthook) is a great tool for setting pre-commit hooks.

## Features

- POSIX-compliant script for maximum compatibility
- Pre-bundled `build` and `plan` prompts
- Configurable via file (sourced before execution) or environment variables
- Max-iteration loop control

## Requirements

- `opencode` on your `PATH` (or replace with your preferred command-line AI tool, see [Changing the agent implementation](#changing-the-agent-implementation) below)
- git for usage in pre-bundled prompts

## Advanced usage

```sh
./ralph.sh [options] [prompt] [scope]
```

### Positional arguments

- `prompt`: Name of the prompt file (without `.md`), or `build` / `plan`. Defaults to `build`.
- `-`: Read the prompt from standard input (stdin).
- `scope`: Optional scope for plan mode.

### Options

```text
-c, --config FILE                 Config file to source
-m, --max-iterations N            Maximum iterations (default: 25)
-p, --prompt-file FILE            Prompt file path (use '-' to read from stdin)
-s, --specs-dir DIR               Specs directory (default: specs)
-i, --specs-index FILE            Specs index file (default: README.md)
--no-specs-index                  Disable specs index file
-n, --implementation-plan-name N  Implementation plan file name
-l, --log-file FILE               Log file path
--no-log                          Disable logs
--log-truncate                    Truncate log file before writing
--stop-condition CONDITION        Custom stop condition text
--prompt PROMPT                   Inline custom prompt (overrides prompt files)
-h, --help                        Show this help message
```

### Prompt Sources

Ralph resolves prompt content in this order:

1. `--prompt` (inline prompt)
2. `--prompt-file FILE`
3. positional prompt name from `RALPH_PROMPTS_DIR` (default `prompts/`)
4. built-in `build` / `plan` defaults

### Read prompt from stdin

```sh
cat prompts/custom.md | ./ralph.sh --prompt-file -
cat prompts/custom.md | ./ralph.sh -
```

### Prompt Placeholders

Ralph supports placeholders in your prompt files that get automatically substituted:

- **`<COMPLETION_SIGNAL>`** - Replaced with the completion signal (`<promise>COMPLETE</promise>`). Use this in your prompts to tell the agent when to stop. Example: "When done, reply with \`<COMPLETION_SIGNAL>\`"

- **`<SCOPE>`** - Replaced with the scope argument passed on the command line (default: "Whole system"). The scope is the second positional argument. Example: `./ralph.sh plan "user authentication"` sets scope to "user authentication". Use this in planning mode to focus the agent on a specific area.

Both placeholders are automatically substituted before the prompt is sent to the agent. Reference them in your custom prompts.

### Changing the agent implementation

This Ralph implementation relies on [opencode](https://opencode.ai/) to work. You can replace it with any command-line tool. To change the implementation, simply replace the `opencode` command in the script with your desired tool.

## Configuration

Ralph supports environment variables and an optional config file. Flags override the environment variables.

### Environment variables

- `RALPH_CONFIG_FILE`
- `RALPH_MAX_ITERATIONS`
- `RALPH_SPECS_DIR`
- `RALPH_SPECS_INDEX_FILE`
- `RALPH_PROMPTS_DIR`
- `RALPH_IMPLEMENTATION_PLAN_NAME`
- `RALPH_LOG_FILE` - Path to a log file where all Ralph output (stdout/stderr) is mirrored.
- `RALPH_LOG_ENABLED` - Set to `0` to disable logs, `1` to enable (default: `1`).
- `RALPH_LOG_APPEND` - Set to `0` to truncate before writing, `1` to append (default: `1`).
- `DEBUG` - Set to any value to print the prompt instead of executing it. Useful for reviewing what would be sent to the agent without actually running it. Example: `DEBUG=1 ./ralph.sh plan "my-feature"`

### Config file

Create a .ralphrc on project root, use `--config FILE` or set `RALPH_CONFIG_FILE`. The config file is sourced by `sh`, so it can set environment variables. Example:

```sh
RALPH_MAX_ITERATIONS=10
RALPH_PROMPTS_DIR=prompts
RALPH_SPECS_DIR=specs
RALPH_SPECS_INDEX_FILE=README.md
RALPH_IMPLEMENTATION_PLAN_NAME=IMPLEMENTATION_PLAN.md
RALPH_LOG_FILE=logs/ralph.log
RALPH_LOG_ENABLED=1
RALPH_LOG_APPEND=1
```

## Examples

Using a custom prompt file located at `prompts/my_prompt.md`:

```sh
./ralph.sh my_prompt
```

Custom prompt file at custom prmompts directory:


```sh
./ralph.sh --prompt-file custom_dir/prompts/my_prompt.md

# Or with environment variable:

RALPH_PROMPTS_DIR=custom_dir/prompts ./ralph.sh my_prompt
```

Inline prompt override:

```sh
./ralph.sh --prompt "Run a quick audit of the API docs and report issues."
```

Read prompt from stdin:

```sh
cat prompts/custom.md | ./ralph.sh --prompt-file -

# Or using positional '-':
cat prompts/custom.md | ./ralph.sh -
```

Enable logs in a file (append mode):

```sh
./ralph.sh --log-file logs/ralph.log
```

Enable logs via environment variable:

```sh
RALPH_LOG_FILE=logs/ralph.log ./ralph.sh
```

Limit iterations and add a stop condition:

```sh
./ralph.sh -m 5 --specs "my_package/specifications" --no-specs-index custom_build_prompt
```

## Testing

Run locally:

```sh
bash test_ralph.sh
```

CI runs tests on push/PR via GitHub Actions.
