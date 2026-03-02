# Ralph Go CLI Implementation

This is a Go implementation of the Ralph shell script, providing a native binary with identical functionality.

## Features

- **Native Binary**: Compiled Go application for better performance and portability
- **Cross-Platform**: Can be built for Linux, macOS, Windows, and other platforms
- **Feature Parity**: Implements all features from the original `ralph.sh` script
- **Modern CLI**: Uses cobra framework for robust command-line interface

## Building

### Using Make

```bash
make build
```

### Using Go Directly

```bash
go build -o ralph .
```

### Cross-Compilation

Build for different platforms:

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o ralph-linux .

# macOS
GOOS=darwin GOARCH=amd64 go build -o ralph-darwin .

# Windows
GOOS=windows GOARCH=amd64 go build -o ralph.exe .
```

## Installation

### System-Wide Installation

```bash
make install
```

Or manually:

```bash
sudo install -m 0755 ralph /usr/local/bin/ralph
```

### Local Installation

Simply copy the `ralph` binary to a directory in your PATH.

## Usage

The Go implementation provides identical command-line interface as the shell script:

```bash
# Run with default build prompt
./ralph

# Run with plan prompt
./ralph plan my-feature

# Use custom max iterations
./ralph --max-iterations 10 build

# Use inline prompt
./ralph --prompt "Custom prompt text"

# Read prompt from stdin
echo "prompt from stdin" | ./ralph -

# Use Claude Code CLI agent instead of opencode
./ralph --agent claude

# Use Cursor CLI agent
./ralph --agent cursor

# Use a specific model with Claude
./ralph --agent claude --model claude-sonnet-4

# Use a sub-agent/agent mode
./ralph --agent opencode --agent-mode reviewer
./ralph --agent claude --agent-mode planner

# Show help
./ralph --help
```

## AI Agent Support

Ralph supports multiple AI CLI agents. Each agent has its own implementation in a separate file:

- **opencode** (default): Uses the `opencode` CLI tool
- **claude**: Uses the `claude` CLI tool (Claude Code CLI)
- **cursor**: Uses the `cursor` CLI tool

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

### Agent Files

Each agent implementation is in its own file:

- `agent.go`: Agent interface definition and factory function
- `agent_opencode.go`: Opencode CLI agent implementation
- `agent_claude.go`: Claude Code CLI agent implementation
- `agent_cursor.go`: Cursor CLI agent implementation

This modular design makes it easy to add support for additional AI CLI tools in the future.

## Configuration

Configuration works identically to the shell script:

1. **Command-line flags** (highest priority)
2. **Environment variables** 
3. **Config file** (`.ralphrc`)
4. **Defaults** (lowest priority)

### Environment Variables

- `RALPH_MAX_ITERATIONS`: Maximum iterations (default: 25)
- `RALPH_SPECS_DIR`: Specs directory (default: `specs`)
- `RALPH_SPECS_INDEX_FILE`: Specs index file (default: `README.md`)
- `RALPH_IMPLEMENTATION_PLAN_NAME`: Implementation plan file name (default: `IMPLEMENTATION_PLAN.md`)
- `RALPH_CUSTOM_PROMPT`: Custom prompt text
- `RALPH_LOG_FILE`: Log file path
- `RALPH_LOG_ENABLED`: Enable/disable logging (`1` or `0`)
- `RALPH_LOG_APPEND`: Append to log file (`1` or `0`)
- `RALPH_PROMPTS_DIR`: Prompts directory (default: `prompts`)
- `RALPH_CONFIG_FILE`: Config file path (default: `ralph.toml`, `.ralphrc.toml`, or `.ralphrc`)
- `RALPH_AGENT`: AI agent to use: `opencode`, `claude`, or `cursor` (default: `opencode`)
- `RALPH_MODEL`: AI model to use (optional, e.g., `claude-sonnet-4`, `gpt-4`)
- `RALPH_AGENT_MODE`: Agent mode/sub-agent name (optional, e.g., `reviewer`, `planner`)

### Config File Format

Create a `ralph.toml` file in your project root or parent directories (legacy `.ralphrc.toml` and `.ralphrc` files are also supported):

```toml
# AI Agent Configuration
agent = "claude"
model = "claude-sonnet-4"
agent-mode = "planner"

# Iteration Settings
max-iterations = 30

# Directory Settings
specs-dir = "specifications"
specs-index-file = "README.md"
implementation-plan-name = "IMPLEMENTATION_PLAN.md"
prompts-dir = ".ralph/prompts"

# Logging Configuration
log-file = "logs/ralph.log"
no-log = false
log-truncate = false
```

All configuration keys in the TOML file correspond to their command-line flags (with hyphens instead of underscores).

**Note:** The config file search order is: `ralph.toml` → `.ralphrc.toml` → `.ralphrc`

## Development

### Project Structure

```
.
├── main.go              # Entry point
├── cmd.go               # Cobra command setup and main loop
├── config.go            # Configuration management
├── prompts.go           # Prompt generation and handling
├── logger.go            # Logging functionality
├── executor.go          # Command execution
├── agent.go             # Agent interface and factory
├── agent_opencode.go    # Opencode CLI agent implementation
├── agent_claude.go      # Claude Code CLI agent implementation
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
└── Makefile             # Build automation
```

### Adding Dependencies

```bash
go get package-name
go mod tidy
```

### Running Tests

```bash
make test
```

Or:

```bash
go test -v ./...
```

## Advantages Over Shell Script

1. **Performance**: Faster startup and execution
2. **Portability**: Single binary, no shell dependencies
3. **Type Safety**: Compile-time error checking
4. **Maintenance**: Easier to refactor and extend
5. **Testing**: Better unit testing support
6. **Cross-Platform**: Native support for Windows, macOS, Linux

## Compatibility

The Go implementation maintains 100% compatibility with the shell script:

- All command-line flags work identically
- Configuration precedence is the same
- Pre-bundled prompts generate identical output
- Log format is identical
- Environment variable handling is the same

You can switch between the shell script and Go binary without any changes to your workflow.
