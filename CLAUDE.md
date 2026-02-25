# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AI Shell Assistant is an intelligent command-line tool for server operations and cloud-native operations. It leverages LLM models to analyze problems, suggest commands, and guide execution through natural language interaction.

Key capabilities:
- Interactive mode with conversation history (up to 10 recursion depth for complex troubleshooting)
- Pipe mode for analyzing command output (supports up to 1.6MB/~13k lines of logs)
- Multi-provider LLM support with automatic fallback on failure
- Safety controls: query commands (green) vs modify commands (red, require double confirmation)
- Internationalization: Chinese and English interfaces

## Build Commands

```bash
# Build for current platform
make build
# Or: ./scripts/build.sh

# Build for all platforms (Linux, macOS, Windows, FreeBSD)
make build-all
# Or: ./scripts/build-all.sh

# Clean build artifacts
make clean

# Run tests
go test ./...

# Run specific test
go test -v ./internal/config
```

## Architecture

### Core Components

**Entry Point**: `cmd/aiassist/main.go`
- Initializes configuration and signal handlers
- Delegates to CLI commands via Cobra framework

**CLI Layer**: `internal/cmd/`
- `root.go`: Main command handling interactive and pipe modes
- `config.go`: Configuration management commands (provider/model management)
- `run.go`: Implements interactive and pipe mode execution flows
- `version.go`: Version information command

**Configuration System**: `internal/config/`
- Supports two modes:
  1. **Local mode**: YAML file at `~/.aiassist/config.yaml`
  2. **Consul mode**: Centralized config from Consul KV store
- When `consul.enabled=true`, all provider/model config comes from Consul
- Local mode allows modifications via CLI; Consul mode is read-only
- Thread-safe with mutex protection for concurrent access

**LLM Manager**: `internal/llm/`
- Manages multiple LLM providers (OpenAI, Alibaba Qwen, custom OpenAI-compatible APIs)
- Automatic fallback: tries providers in config file order when one fails
- Tracks model availability (marks unavailable on HTTP 429 quota exhaustion)
- `manager.go`: Provider lifecycle and fallback logic
- `openai_compatible.go`: Generic OpenAI API client implementation

**Interactive Session**: `internal/interactive/session.go`
- Maintains conversation history with system/user/assistant messages
- Two modes:
  1. Interactive: Continuous dialogue with command execution
  2. Pipe: Analyze piped data and exit (no command execution)
- Recursion depth limit (10) prevents infinite command loops
- Truncates large outputs to fit LLM context windows (100k chars for interactive, 400k for pipe)

**Command Executor**: `internal/executor/executor.go`
- Extracts commands from AI responses using `[cmd:query]` and `[cmd:modify]` markers
- Classifies commands by risk level:
  - Query commands (green): Read-only operations, single confirmation
  - Modify commands (red): Write operations, double confirmation
- Executes commands via `sh -c` shell

**Prompt System**: `internal/prompt/`
- Three prompt types for different scenarios:
  1. Interactive: Initial user question
  2. ContinueAnalysis: Analyzing command output from previous step
  3. PipeAnalysis: Analyzing piped command output
- Language-specific prompts (Chinese/English)

**System Info**: `internal/sysinfo/`
- Collects OS, kernel, available tools information
- Cached in `~/.aiassist/sysinfo.json` for performance
- Added as context to LLM prompts for better command suggestions

**Internationalization**: `internal/i18n/`
- Message translations in `messages_zh.go` and `messages_en.go`
- All user-facing messages go through i18n translator

**UI Utilities**: `internal/ui/`
- Spinner animations
- Terminal separators

## Key Implementation Details

### Model Fallback Order

The order of model calls is **determined by the order in the config file**. For example:

```yaml
providers:
  bailian:  # First tried
    models:
      - name: qwen-plus  # 1st choice
      - name: qwen-max   # 2nd choice
  openai:   # Second tried if bailian fails
    models:
      - name: gpt-4      # 3rd choice
```

The manager tries each enabled model in this exact order. If a model returns HTTP 429 (quota exhausted), it's marked unavailable and skipped.

### Command Execution Flow

1. User asks question
2. AI analyzes with system prompt + conversation history
3. AI responds with commands marked as `[cmd:query]` or `[cmd:modify]`
4. Executor extracts commands, displays with color coding
5. User confirms (once for query, twice for modify)
6. Command executes, output added to conversation history
7. AI analyzes output and suggests next steps (recursively, max depth 10)

### Pipe Mode vs Interactive Mode

**Pipe mode** (`cmd | aiassist`):
- Reads up to 1.6MB of data
- Analyzes with specialized prompt
- Shows analysis and exits (no command execution)
- No interactive loop

**Interactive mode** (`aiassist`):
- Continuous dialogue
- Executes commands with user confirmation
- Maintains full conversation history
- Max 10 levels of command recursion

### Configuration Management

When adding/modifying providers or models:
- Check `config.IsConsulMode()` first - if true, reject modifications
- All provider operations go through `config.AddProvider()`, `config.DeleteProvider()`, etc.
- Model enable/disable tracked in LLM manager's `modelEnabled` map
- Default model format: `provider/model-name` (e.g., `bailian/qwen-max`)

### Output Truncation Strategy

For large outputs, the system uses intelligent truncation:
- Interactive mode: 100k character limit
- Pipe mode: 400k character limit (supports ~13k lines of nginx logs)
- Keeps 60% from beginning, 40% from end
- Adds truncation message showing number of omitted lines

## Testing

Tests use standard Go testing framework. Key test files:
- `config_test.go`: Configuration loading/saving
- `executor_test.go`: Command extraction and classification
- `manager_test.go`: LLM provider fallback logic
- `prompts_test.go`: Prompt generation by language

Run all tests: `go test ./...`

## Version Information

Version and commit are injected at build time via ldflags:
```bash
CGO_ENABLED=0 go build -ldflags "-X main.Version=${VERSION} -X main.Commit=${COMMIT} -s -w" ./cmd/aiassist/
```

The `-s -w` flags strip debug information for smaller binaries. `CGO_ENABLED=0` produces static binaries.

## Development Notes

- Binary name: `aiassist`
- Config directory: `~/.aiassist/`
- Config file: `~/.aiassist/config.yaml`
- System info cache: `~/.aiassist/sysinfo.json`
- Go version: 1.21+
- Main dependencies: Cobra (CLI), fatih/color (terminal colors), peterh/liner (line editor), hashicorp/consul/api (config center)
