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

## 📖 Configuration

### API Key Information

Each LLM Provider requires an API Key for access. API Keys are credentials for accessing model services - please keep them secure.

**Model Status:**
- `✓ Enabled` - Model is enabled and available
- `✗ Disabled` - Model has been disabled by user
- `✗ Unavailable` - **Model is temporarily unavailable** (quota exhausted or rate limited)

When a model returns HTTP 429 status code, it indicates the API call limit has been reached or quota has been exhausted. The model will be marked as `Unavailable` and the system will automatically switch to the next available model.

**Recommendations:**
- Configure multiple Providers or models for fallback
- Check model status regularly: `aiassist config provider list`
- If a model remains unavailable, check your API Key's quota and billing status

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
