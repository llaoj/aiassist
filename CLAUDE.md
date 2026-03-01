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
```

## Architecture

### Core Components

**Entry Point**: `cmd/aiassist/main.go`
- Initializes configuration and signal handlers
- Delegates to CLI commands via Cobra framework

**CLI Layer**: `internal/cmd/`
- `root.go`: Main command handling interactive and pipe modes
- `config.go`: Interactive configuration wizard and configuration viewer
- `run.go`: Implements interactive and pipe mode execution flows
- `version.go`: Version information command

**Configuration System**: `internal/config/`
- Supports two modes:
  1. **Local mode**: YAML file at `~/.aiassist/config.yaml`
  2. **Consul mode**: Centralized config from Consul KV store
- When `consul.enabled=true`, all provider/model config comes from Consul
- Local mode: config modified via direct file editing
- Consul mode: read-only from CLI, modified via Consul UI/API/CLI
- Thread-safe with mutex protection for concurrent access

**LLM Manager**: `internal/llm/`
- Manages multiple LLM models across different providers (OpenAI, Alibaba Qwen, custom OpenAI-compatible APIs)
- Automatic fallback: tries models in config file order when one fails
- Tracks model availability (marks unavailable on HTTP errors)
- `manager.go`: Model lifecycle and fallback logic
- `openai_compatible.go`: Generic OpenAI API model implementation

**Interactive Session**: `internal/interactive/session.go`
- Maintains conversation history with system/user/assistant messages
- Two modes:
  1. Interactive: Continuous dialogue with command execution
  2. Pipe: Analyze piped data and exit (no command execution)
- Recursion depth limit (10) prevents infinite command loops
- Truncates large outputs to fit LLM context windows (400k chars limit for both modes)

**Command Executor**: `internal/executor/executor.go`
- Extracts commands from AI responses using `[cmd:query]` and `[cmd:modify]` markers
- Classifies commands by risk level:
  - Query commands (green): Read-only operations, single confirmation (selection list)
  - Modify commands (red): Write operations, double confirmation (selection list)
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
  - name: bailian  # First tried
    models:
      - name: qwen-plus  # 1st choice
        enabled: true
      - name: qwen-max   # 2nd choice
        enabled: true
  - name: openai   # Second tried if bailian fails
    models:
      - name: gpt-4      # 3rd choice
        enabled: true
```

The manager tries each enabled model in this exact order. If a model returns an error, it's marked unavailable and skipped.

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
- Default model format: `provider/model-name` (e.g., `bailian/qwen-max`)

### Output Truncation Strategy

For large outputs, the system uses intelligent truncation:
- Both interactive and pipe mode: 400k character limit (supports ~13k lines of nginx logs)
- Keeps 60% from beginning, 40% from end
- Adds truncation message showing number of omitted lines

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
- Main dependencies: Cobra (CLI), fatih/color (terminal colors), charmbracelet/bubbletea (modern terminal UI), hashicorp/consul/api (config center)

## Documentation Synchronization

**CRITICAL**: Code and documentation must always stay in sync.

### Synchronization Rules

1. **When modifying code logic**:
   - Check if the modified logic is documented in `docs/` or README files
   - Update all relevant documentation to reflect the changes
   - Ensure examples, flow descriptions, and behavior descriptions match the new code

2. **When removing code logic**:
   - Search for and remove all documentation references to the removed feature
   - Remove related examples, usage instructions, and descriptions
   - Clean up any screenshots, diagrams, or references that mention the removed logic

3. **When adding new features**:
   - Document the new feature in relevant documentation files
   - Add usage examples and behavior descriptions
   - Update architecture diagrams if needed

### Documentation Files to Check

When making changes, search these locations for documentation that may need updates:
- `docs/ARTICLE.md` - Main article with examples and use cases
- `README.md` - Project overview and quick start
- `CLAUDE.md` - This file (architecture and implementation details)
- Code comments and godoc strings

### Verification Checklist

Before completing any code change:
- [ ] Searched documentation files for references to modified/removed logic
- [ ] Updated all found references with new behavior
- [ ] Removed documentation for deleted features
- [ ] Added documentation for new features
- [ ] Examples still work as documented
- [ ] Screenshots/diagrams are still accurate

## Code Comments Guidelines

**CRITICAL**: All code comments must be written in English.

### Implementation Rules

1. **Never write comments in Chinese** or any other non-English language
2. **All inline comments, block comments, and documentation comments** must be in English
3. **Variable names, function names, and other identifiers** should also be in English
4. **Exception**: Translation files (like `messages_zh.go`, `prompts_zh.go`) contain Chinese text by design - these are not comments but actual content

### Why English Comments

- Code is read by developers worldwide
- Consistency with Go community standards
- Better collaboration and code review experience
- Prevents encoding issues in different environments

## Internationalization (i18n) Guidelines

**CRITICAL**: All user-facing messages must support both Chinese and English.

### Implementation Rules

1. **Never hardcode user-facing strings** in the codebase
2. **Always use the i18n system** for all messages displayed to users
3. **Add translations to both** `internal/i18n/messages_zh.go` and `internal/i18n/messages_en.go`
4. **Use consistent message keys** across both language files

### How to Add New Messages

When adding a new user-facing message:

1. Add the message key to both language files:
   ```go
   // messages_zh.go
   "ui.ctrlc_exit_hint": "再按一次 Ctrl+C 退出程序",

   // messages_en.go
   "ui.ctrlc_exit_hint": "Press Ctrl+C again to exit",
   ```

2. Use the translator in your code:
   ```go
   translator := i18n.New(language)
   message := translator.T("ui.ctrlc_exit_hint")
   fmt.Println(message)
   ```

### Message Categories

Organize message keys by category using dot notation:
- `config.*` - Configuration related messages
- `interactive.*` - Interactive session messages
- `executor.*` - Command execution messages
- `error.*` - Error messages
- `ui.*` - UI component messages
- `llm.*` - LLM related messages

### Where to Find i18n Implementation

- **i18n package**: `internal/i18n/`
- **Chinese messages**: `internal/i18n/messages_zh.go`
- **English messages**: `internal/i18n/messages_en.go`
- **Translator initialization**: Via `i18n.New()` in each command handler
