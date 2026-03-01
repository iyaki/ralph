# Ralph Agentic Loop

Ralph is a POSIX-compliant shell runner aimed for spec-driven development workflows.  
It loads prompts from files (with optional inline overrides) and comes with build/plan presets.

## About the Ralph Wiggum Methodology

Ralph is based on the [Ralph Wiggum methodology](https://ghuntley.com/loop/) pioneered by [Geoffrey Huntley](https://ghuntley.com/).

**Core Principles:**
- **Spec-driven development** - Requirements defined upfront in markdown specs
- **Monolithic operation** - One agent, one task, one loop iteration at a time
- **Fresh context** - Each iteration starts with a clean context window
- **Backpressure** - Tests and validation provide immediate feedback (The Architectural constraints of [Harness engineering](https://martinfowler.com/articles/exploring-gen-ai/harness-engineering.html))
- **Let Ralph Ralph** - Trust the agent to self-correct through iteration
- **Disposable plans** - Regenerate implementation plans when they go stale
- **Simple loops** - Minimal bash loops feeding prompts to AI agents

The methodology works in three phases:
  1. Define requirements through human+LLM conversation to create specs
  2. Gap analysis to generate/update implementation plans
  3. Build loops that implement one task at a time, commit, and update the plan

For a comprehensive guide, see the [Ralph Playbook](https://github.com/ClaytonFarr/ralph-playbook).

This script implements what has worked well for me, based on the work made by Geoffrey at [Loom](https://github.com/ghuntley/loom).

This script alone is not enough to achieve good results. The quality of the prompts and specs you provide, and the backpressure you set, will greatly influence the outcomes.

For setting pre-commit hooks as a kind of backpressure, I like pretty much [lefthook](https://github.com/evilmartians/lefthook).

## Features

- POSIX-compliant `sh` script
- Pre-bundled `build` and `plan` prompts
- Prompt file lookup that searches upward from the current directory
- Inline custom prompt support
- Config file support (sourced before execution)
- Custom stop condition text and stop signal handling
- Specs directory and optional specs index file integration
- Max-iteration loop control

## Requirements

- POSIX `sh`
- `opencode` on your `PATH`
- Git available on your system

### Changing the agent implementation

This Ralph implementation relies on `opencode` to work. You can replace it with any command-line tool that accepts code input and returns output. To change the implementation, simply replace the `opencode` command in the script with your desired tool.

## Usage

```sh
./ralph.sh [options] [prompt] [scope]
```

### Positional arguments

- `prompt`: Name of the prompt file (without `.md`), or `build` / `plan`. Defaults to `build`.
- `scope`: Optional scope for plan mode.

### Options

```text
-c, --config FILE                         Config file to source
-m, --max-iterations N                    Maximum iterations (default: 25)
-p, --prompt-file FILE                    Prompt file path
-s, --specs-dir DIR                       Specs directory (default: specs)
-i, --specs-index FILE                    Specs index file (default: README.md)
	--no-specs-index                        Disable specs index file
-n, --implementation-plan-name FILENAME   Implementation plan file name
	-l, --log-file FILE                      Log all output to a file
	--no-log                                 Disable logging
	--log-truncate                           Truncate log file before writing
	--stop-condition CONDITION              Custom stop condition text
	--prompt PROMPT                         Inline custom prompt (overrides prompt files)
-h, --help                                Show help
```

## Examples

If you have a `specs/` directory similar to [this one](https://github.com/ghuntley/loom/tree/trunk/specs), using ralph can be as simple as running `./ralph.sh plan my-feature` to generate an implementation plan and `./ralph.sh` (defaults to `build`) to start implementing the feature.

### Advanced usage

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

Create a .ralphrc on project root or use `--config FILE`, set `RALPH_CONFIG_FILE`. The config file is sourced by `sh`, so it can set environment variables. Example:

```sh
RALPH_MAX_ITERATIONS=10
RALPH_PROMPTS_DIR=prompts
RALPH_SPECS_DIR=specs
RALPH_SPECS_INDEX_FILE=README.md
RALPH_IMPLEMENTATION_PLAN_NAME=IMPLEMENTATION_PLAN.md
```

## Stop condition

The script expects the stop signal `<promise>COMPLETE</promise>` to appear in the tool output. You can also add extra stop guidance with `--stop-condition` or `RALPH_STOP_CONDITION`, which is appended to the prompt.

## Prompt Placeholders

Ralph supports placeholders in your prompt files that get automatically substituted:

- **`<COMPLETION_SIGNAL>`** - Replaced with the completion signal (`<promise>COMPLETE</promise>`). Use this in your prompts to tell the agent when to stop. Example: "When done, reply with \`<COMPLETION_SIGNAL>\`"

- **`<SCOPE>`** - Replaced with the scope argument passed on the command line (default: "Whole system"). The scope is the second positional argument. Example: `./ralph.sh plan "user authentication"` sets scope to "user authentication". Use this in planning mode to focus the agent on a specific area.

Both placeholders are automatically substituted before the prompt is sent to the agent. You can reference them in your custom prompt files.

## Notes

- The script is intentionally POSIX-compliant and should run with `/bin/sh`.
- Prompt files are expected to be Markdown (`.md`).
