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

**aiassist** is an intelligent terminal tool powered by Large Language Models, providing DevOps engineers with diagnostics, solution recommendations, and command execution guidance through natural language interaction.

## ✨ Key Features

- 🤖 **AI-Powered**: Integrates mainstream LLMs (Qwen, OpenAI, etc.)
- 🔄 **Smart Fallback**: Automatic model switching based on configuration order
- 🎯 **Context-Aware**: Correlates command execution results for continuous conversation
- 📊 **Pipeline Analysis**: Directly analyzes command output streams
- 🛡️ **Safety Controls**: Query (green) vs modify (red) commands, supports y/n shortcuts, modify commands require double confirmation after initial approval
- 🌍 **Multilingual**: Chinese/English interface
- ⚙️ **Flexible Configuration**: Multiple Providers, models, custom API Keys, and proxies

## 🚀 Quick Start

### One-Line Installation

```bash
curl -fsSL https://raw.githubusercontent.com/llaoj/aiassist/main/scripts/install.sh | bash
```

For detailed installation instructions, see [INSTALL.md](INSTALL.md)

## 📖 Common Commands

```bash
# View version
aiassist version

# View current configuration
aiassist config view

# View help
aiassist --help
```

## � Usage Modes

### Interactive Mode

**Mode 1: Enter Interactive Conversation**

```bash
aiassist
```

**Mode 2: Single Question & Answer (No Interactive Loop)**

```bash
aiassist "Why is the server load high?"
```

### Pipe Analysis Mode

Directly analyze command output:

```bash
# Analyze piped data only (AI infers the question)
tail -f /var/log/nginx/access.log | aiassist

# Analyze with context question (Recommended)
docker ps -a | aiassist "Analyze container status"
cat go.sum | aiassist "Analyze output of cat go.sum"

# Analyze error logs with specific question
journalctl -u nginx -n 100 | aiassist "Find the cause of errors"
```

**Workflow:**
1. Piped command output serves as input
2. AI automatically analyzes data and identifies issues
3. Provides diagnostic conclusions and solutions
4. Offers executable remediation commands

## �🔧 Configuration
### Configuration Modes

aiassist supports two configuration modes:

#### 🏢 Configuration Center Mode (Recommended for Enterprise)

Use **Consul** for centralized configuration management, with all hosts loading config from the center in real-time.

**Advantages:**
- ✅ Unified configuration management, one change applies globally
- ✅ Multi-host configuration sync, no need to configure each host
- ✅ Configuration version control
- ✅ Team collaboration friendly

**Setup Steps:**

1. **Start Consul** (Optional, skip if you already have Consul):
   ```bash
   # Docker way
   docker run -d -p 8500:8500 --name=consul consul agent -server -ui -bootstrap-expect=1 -client=0.0.0.0
   ```

2. **Create configuration in Consul KV**:
   ```bash
   # Access Consul UI: http://localhost:8500
   # Create Key: aiassist/config
   # Content:
   language: en
   default_model: bailian/qwen-max
   providers:
     bailian:
       name: bailian
       base_url: https://dashscope.aliyuncs.com/compatible-mode/v1
       api_key: sk-xxxxxxxxxxxx
       enabled: true
       models:
         - name: qwen-max
           enabled: true
   ```

3. **Configure local file** (`~/.aiassist/config.yaml`):
   ```yaml
   # Only Consul connection info needed
   consul:
     enabled: true
     address: "127.0.0.1:8500"
     key: "aiassist/config"
     token: ""  # ACL Token (optional)
   
   # language, providers, etc. are all loaded from Consul
   ```

**Notes:**
- ⚠️ In configuration center mode, `aiassist config` command is read-only
- 💡 All configuration changes must be done in Consul KV
- 🔄 Configuration changes take effect immediately without restart

#### 💻 Local Configuration Mode (Personal Use)

Configure directly in local file `~/.aiassist/config.yaml`, simple and straightforward.

**Advantages:**
- ✅ No additional services required
- ✅ Simple and intuitive configuration
- ✅ Perfect for personal use

Configuration example:

```yaml
language: en
default_model: bailian/qwen-max

providers:
  bailian:
    name: bailian
    base_url: https://dashscope.aliyuncs.com/compatible-mode/v1
    api_key: sk-xxxxxxxxxxxx
    enabled: true
    models:
      - name: qwen-max
        enabled: true
      - name: qwen-plus
        enabled: true
```

---
### API Key Information

Each LLM Provider requires an API Key for access. API Keys are credentials for accessing model services - please keep them secure.

**Model Status:**
- `✓ Enabled` - Model is enabled and available
- `✗ Disabled` - Model has been disabled by user
- `✗ Unavailable` - **Model is temporarily unavailable** (quota exhausted or rate limited)

When a model returns HTTP 429 status code, it indicates the API call limit has been reached or quota has been exhausted. The model will be marked as `Unavailable` and the system will automatically switch to the next available model.

**Recommendations:**
- Configure multiple Providers or models for fallback
- If a model remains unavailable, check your API Key's quota and billing status
- Run `aiassist config view` to check current configuration and model status

**💡 About Paid Models:**

We encourage paying for quality AI services. Paid models typically offer:
- 🚀 Faster response times
- 🎯 Higher accuracy and intelligence
- 💪 More stable service quality and higher rate limits
- ⭐ Better technical support

### Provider Configuration Examples

#### Qwen (Alibaba Cloud Bailian)

```bash
# Apply for API Key
https://dashscope.console.aliyun.com/apiKey

# Configuration
Provider Name: bailian
Base URL: https://dashscope.aliyuncs.com/compatible-mode/v1
API Key: sk-xxxxxxxxxxxx
Models: qwen-plus,qwen-turbo,qwen-max
```

#### OpenAI

```bash
# Apply for API Key
https://platform.openai.com/api-keys

# Configuration
Provider Name: openai
Base URL: https://api.openai.com/v1
API Key: sk-xxxxxxxxxxxx
Models: gpt-4,gpt-3.5-turbo
HTTP Proxy: http://127.0.0.1:7890  # Required in China
```

#### Other OpenAI-Compatible APIs

Any service implementing the OpenAI API standard can be configured:

```bash
Provider Name: custom
Base URL: https://your-api-endpoint/v1
API Key: your-api-key
Models: model-name-1,model-name-2
```
