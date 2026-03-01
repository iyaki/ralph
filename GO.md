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

# Show help
./ralph --help
```

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
- `RALPH_CONFIG_FILE`: Config file path (default: `.ralphrc`)

### Config File Format

Create a `.ralphrc` file in your project root or parent directories:

```bash
RALPH_MAX_ITERATIONS=30
RALPH_SPECS_DIR=specifications
RALPH_LOG_FILE=logs/ralph.log
```

## Development

### Project Structure

```
.
├── main.go         # Entry point
├── cmd.go          # Cobra command setup and main loop
├── config.go       # Configuration management
├── prompts.go      # Prompt generation and handling
├── logger.go       # Logging functionality
├── executor.go     # Command execution
├── go.mod          # Go module definition
├── go.sum          # Dependency checksums
└── Makefile        # Build automation
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
