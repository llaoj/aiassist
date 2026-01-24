<div align="center">
  <img src="logo.svg" alt="AI Shell Assistant Logo" width="100"/>
  
  # AI Shell Assistant
  
  Intelligent Command-Line Assistant for Server & Cloud-Native Operations
  
  [![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
  [![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
  [![Release](https://img.shields.io/github/v/release/llaoj/aiassist)](https://github.com/llaoj/aiassist/releases)
  
  ---
  
  **🤖 This project is entirely AI-generated | 本项目由 AI 全程编写完成**
  
  ---
  
  [中文](README.md) | English
  
</div>

**aiassist** is an intelligent terminal tool powered by Large Language Models, providing DevOps engineers with diagnostics, solution recommendations, and command execution guidance through natural language interaction, significantly improving operational efficiency.

## ✨ Key Features

- 🤖 **AI-Powered**: Integrates mainstream LLMs (Qwen, OpenAI, etc.) with natural language interaction
- 🔄 **Smart Fallback**: Automatic model switching based on configuration file order
- 🎯 **Context-Aware**: Automatically correlates command execution results for continuous conversation
- 📊 **Pipeline Analysis**: Directly analyzes command output streams, e.g., `tail -f access.log | aiassist`
- 🛡️ **Safety Controls**: Query commands (green) and modify commands (red) with differentiated display; modify commands require double confirmation
- 🌍 **Multilingual**: Chinese/English interface support
- ⚙️ **Flexible Configuration**: Supports multiple Providers, models, custom API Keys, and proxies

## 🚀 Quick Start

### One-Line Installation

```bash
curl -fsSL https://raw.githubusercontent.com/llaoj/aiassist/main/scripts/install.sh | bash
```

Mainland China Network Environment (GitHub Proxy):

```bash
ghproxy=https://ghfast.top; curl -fsSL ${ghproxy}/https://raw.githubusercontent.com/llaoj/aiassist/main/scripts/install.sh | sed "s#https://#${ghproxy}/https://#g" | bash
```

> ⚠️ **Note**: GitHub Proxy services are often unstable and may fail or change domains at any time. If installation fails, replace `ghproxy=https://ghfast.top` in the command with another available proxy address.

### Supported Platforms

| Platform | Architectures |
|----------|---------------|
| Linux | x86_64, ARM64, ARM, i386 |
| macOS | Intel (x86_64), Apple Silicon (ARM64) |
| Windows | x86_64, ARM64, i386 |
| FreeBSD | x86_64, ARM64 |

For detailed installation instructions, see [INSTALL.md](INSTALL.md)

## 📖 User Guide

### Initial Configuration

First-time use requires LLM Provider configuration:

```bash
aiassist config
```

The interactive wizard will guide you through:
1. Select language (中文/English)
2. Add LLM Provider (supports OpenAI-compatible interfaces)
3. Configure API Key
4. Set model list
5. (Optional) Configure HTTP proxy

### Interactive Mode

Run directly to enter conversation mode:

```bash
aiassist
```

Example conversation:
```
You> Server load is very high, help me troubleshoot

AI> Let's check the system load first:

[Query Command]
top -b -n 1 | head -20

Execute? (yes/no): yes

[Execution Result]
...

AI> From the output, CPU usage is mainly consumed by nginx process (PID 1234)...
Recommendation:

[Query Command]
ps aux | grep nginx
```

**Features:**
- ✅ Natural language questions
- ✅ AI automatically analyzes and suggests commands
- ✅ Command type annotation (query/modify)
- ✅ Manual confirmation before execution
- ✅ Automatically reads previous command output for continuous analysis

### Pipeline Analysis Mode

Directly analyze command output:

```bash
# Analyze log files
tail -f /var/log/nginx/access.log | aiassist

# Analyze system state
docker ps -a | aiassist "analyze container status"

# Analyze error logs
journalctl -u nginx -n 100 | aiassist "find error cause"
```

**Workflow:**
1. Command output before the pipe serves as input
2. AI automatically analyzes data and identifies anomalies
3. Provides diagnostic conclusions and solutions
4. Offers executable fix commands

### Common Commands

```bash
# View version
aiassist version

# Configuration wizard
aiassist config

# Add Provider
aiassist config provider add

# List all Providers
aiassist config provider list

# Enable/disable Provider
aiassist config provider enable <name>
aiassist config provider disable <name>

# Delete Provider
aiassist config provider delete <name>

# View help
aiassist --help
```

## 🔧 Configuration

### Configuration File

Configuration file is located at `~/.aiassist/config.yaml`:

```yaml
language: en  # zh=中文, en=English
http_proxy: ""  # HTTP proxy address (optional)

providers:
  bailian:  # Provider name
    name: bailian
    base_url: https://dashscope.aliyuncs.com/compatible-mode/v1
    api_key: sk-xxx
    enabled: true
    models:
      - name: qwen-plus
        enabled: true
      - name: qwen-turbo
        enabled: true
  
  openai:
    name: openai
    base_url: https://api.openai.com/v1
    api_key: sk-xxx
    enabled: true
    models:
      - name: gpt-4
        enabled: true
      - name: gpt-3.5-turbo
        enabled: false
```

### Model Call Order

**Important: Model call order is determined by the order in the configuration file.**

For example, with the above configuration, the call order is:
1. `bailian/qwen-plus` (first enabled)
2. `bailian/qwen-turbo` (second enabled)
3. `openai/gpt-4` (third enabled)

If the current model fails, times out, or is unavailable, it will automatically switch to the next enabled model.

### Provider Configuration

#### Qwen (Alibaba Cloud Bailian)

```bash
# Application URL
https://dashscope.console.aliyun.com/apiKey

# Configuration example
Provider Name: bailian
Base URL: https://dashscope.aliyuncs.com/compatible-mode/v1
API Key: sk-xxxxxxxxxxxx
Models: qwen-plus,qwen-turbo,qwen-max
```

#### OpenAI

```bash
# Application URL
https://platform.openai.com/api-keys

# Configuration example
Provider Name: openai
Base URL: https://api.openai.com/v1
API Key: sk-xxxxxxxxxxxx
Models: gpt-4,gpt-3.5-turbo
HTTP Proxy: http://127.0.0.1:7890  # Required for Mainland China
```

#### Other OpenAI-Compatible APIs

Any service implementing the OpenAI API standard can be configured:

```bash
Provider Name: custom
Base URL: https://your-api-endpoint/v1
API Key: your-api-key
Models: model-name-1,model-name-2
```

## 🛡️ Security Design

### Command Classification

aiassist categorizes commands into two types:

| Type | Marker | Color | Confirmations | Examples |
|------|--------|-------|---------------|----------|
| Query Command | `[cmd:query]` | 🟢 Green | 1 time | `ps aux`, `cat /etc/config`, `docker ps` |
| Modify Command | `[cmd:modify]` | 🔴 Red | 2 times | `systemctl restart`, `rm -rf`, `iptables -A` |

### Execution Flow

1. **AI Analysis**: Understand the problem and provide solutions
2. **Command Display**: Show suggested commands with type markers
3. **User Confirmation**:
   - Query commands: Confirm once to execute
   - Modify commands: Requires double confirmation to prevent mistakes
4. **Execution Feedback**: Display execution results, AI continues analysis

### Command Marker Specification

The AI model automatically adds type markers when generating commands:

```
[cmd:query] ps aux | grep nginx    # Query command
[cmd:modify] systemctl restart nginx  # Modify command
```

These markers are automatically generated by the AI model based on system prompts. The tool parses these markers to determine command types.

## 🎯 Use Cases

### Troubleshooting

```bash
You> Nginx won't start, error says bind failed

AI> Port may be in use, let's check:
[cmd:query] netstat -tlnp | grep :80

# After finding port occupied...
AI> Found process 1234 occupying port 80, suggestion:
[cmd:modify] kill -9 1234
[cmd:modify] systemctl start nginx
```

### Performance Analysis

```bash
You> Server CPU load is too high

AI> Let's check process resource usage first:
[cmd:query] top -b -n 1 | head -20
[cmd:query] ps aux --sort=-%cpu | head -10

# After analysis...
AI> Found mysql process consuming high CPU, suggest checking slow queries...
```

### Log Analysis

```bash
tail -f /var/log/nginx/access.log | aiassist

AI> Anomalies detected:
- IP 192.168.1.100 made 500+ requests in 1 minute
- Large number of 404 errors
- Suspected scanning attack

Recommendation:
[cmd:modify] iptables -A INPUT -s 192.168.1.100 -j DROP
```

### Configuration Check

```bash
You> Check if Nginx configuration is correct

AI> Let's check the configuration file:
[cmd:query] nginx -t
[cmd:query] cat /etc/nginx/nginx.conf

# After finding errors...
AI> Configuration file line 45 is missing a semicolon, suggest modifying...
```

## 🔄 How It Works

```
┌──────────────┐
│  User Input  │
│  or Pipeline │
└──────┬───────┘
       │
       ▼
┌──────────────────┐
│  System Info     │
│  (OS/Ver/Tools)  │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  Build Prompt    │
│  (Q+SysInfo)     │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  LLM Analysis    │
│  (Config Order)  │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  Extract Cmds    │
│  [cmd:query]     │
│  [cmd:modify]    │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  User Confirm    │
│  (By Type)       │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  Execute Command │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  Result Feedback │
│  (Continue)      │
└──────────────────┘
```

## 📝 Development

### Build from Source

```bash
# Clone repository
git clone https://github.com/llaoj/aiassist.git
cd aiassist

# Install dependencies
go mod download

# Build (current platform)
make build
# or
./build.sh

# Build all platforms
./scripts/build-all.sh

# Run tests
go test ./...

# Run
./aiassist
```

### Project Structure

```
aiassist/
├── cmd/aiassist/          # Main program entry
├── internal/
│   ├── cmd/               # CLI command implementation
│   │   ├── root.go       # Root command
│   │   ├── config.go     # Config command
│   │   ├── interactive.go # Interactive mode
│   │   └── version.go    # Version command
│   ├── config/            # Configuration management
│   ├── executor/          # Command executor
│   ├── i18n/              # Internationalization
│   ├── interactive/       # Interactive session
│   ├── llm/               # LLM manager
│   │   ├── manager.go    # Provider management
│   │   └── openai_compatible.go # OpenAI compatible interface
│   ├── prompt/            # System prompts
│   ├── sysinfo/           # System info collection
│   └── ui/                # UI utilities
├── .github/workflows/     # CI/CD
└── scripts/               # Scripts
    ├── install.sh        # One-line installation script
    ├── build-all.sh      # Multi-platform build
    └── test-install.sh   # Install test
```

### Tech Stack

- **Language**: Go 1.21+
- **CLI Framework**: cobra
- **Configuration**: YAML
- **HTTP Client**: net/http (standard library)
- **Testing**: Go testing

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details

## 🙏 Acknowledgments

- [OpenAI](https://openai.com/) - API standard reference
- [Alibaba Cloud Bailian](https://www.aliyun.com/product/bailian) - Qwen support
- [Cobra](https://github.com/spf13/cobra) - CLI framework

## 📞 Contact

- Bug Reports: [GitHub Issues](https://github.com/llaoj/aiassist/issues)
- Feature Requests: [GitHub Discussions](https://github.com/llaoj/aiassist/discussions)

---

**⭐ If this project helps you, please give it a Star!**
