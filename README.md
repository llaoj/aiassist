<div align="center">
  <img src="logo.svg" alt="AI Shell Assistant Logo" width="100"/>
  
  # AI Shell Assistant
  
  > 面向服务器运维、云原生运维的智能命令行助手
  
  [![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
  [![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
  [![Release](https://img.shields.io/github/v/release/llaoj/aiassist)](https://github.com/llaoj/aiassist/releases)
  
  ---
  
  **🤖 本项目由 AI 全程编写完成 | This project is entirely AI-generated**
  
  ---
  
</div>

**aiassist** 是一个基于大语言模型的智能终端工具，通过自然语言交互为运维人员提供诊断分析、方案建议和命令执行指导，显著提升运维效率。

## ✨ 核心特性

- 🤖 **AI 驱动**：集成主流大语言模型（通义千问、OpenAI等），支持自然语言交互
- 🔄 **智能 Fallback**：多模型自动切换，配置文件顺序决定调用优先级
- 🎯 **上下文感知**：自动关联命令执行结果，支持连续对话
- 📊 **管道分析**：直接分析命令输出流，如 `tail -f access.log | aiassist`
- 🛡️ **安全控制**：查询命令（绿色）和修改命令（红色）差异化展示，修改命令需二次确认
- 🌍 **多语言支持**：中文/英文界面
- ⚙️ **灵活配置**：支持多 Provider、多模型、自定义 API Key 和代理

## 🚀 快速开始

### 一键安装

```bash
curl -fsSL https://raw.githubusercontent.com/llaoj/aiassist/main/scripts/install.sh | bash
```

### 支持平台

| 平台 | 架构 |
|------|------|
| Linux | x86_64, ARM64, ARM, i386 |
| macOS | Intel (x86_64), Apple Silicon (ARM64) |
| Windows | x86_64, ARM64, i386 |
| FreeBSD | x86_64, ARM64 |

详细安装说明请查看 [INSTALL.md](INSTALL.md)

## 📖 使用指南

### 首次配置

首次使用需要配置 LLM Provider：

```bash
aiassist config
```

交互式向导将引导你完成：
1. 选择语言（中文/English）
2. 添加 LLM Provider（支持 OpenAI 兼容接口）
3. 配置 API Key
4. 设置模型列表
5. （可选）配置 HTTP 代理

### 交互式模式

直接运行进入对话模式：

```bash
aiassist
```

示例对话：
```
You> 服务器负载很高，帮我排查原因

AI> 让我们先检查一下系统负载情况：

[查询命令]
top -b -n 1 | head -20

是否执行? (yes/no): yes

[执行结果]
...

AI> 从输出看，CPU 使用率主要是进程 nginx (PID 1234) 占用...
建议执行：

[查询命令]
ps aux | grep nginx
```

**特点：**
- ✅ 自然语言提问
- ✅ AI 自动分析并给出命令建议
- ✅ 命令标注类型（查询/修改）
- ✅ 手动确认后执行
- ✅ 自动读取上一条命令输出，进行连续分析

### 管道分析模式

直接分析命令输出：

```bash
# 分析日志文件
tail -f /var/log/nginx/access.log | aiassist

# 分析系统状态
docker ps -a | aiassist "分析容器状态"

# 分析错误日志
journalctl -u nginx -n 100 | aiassist "找出错误原因"
```

**工作流程：**
1. 管道前的命令输出作为输入
2. AI 自动分析数据，识别异常
3. 给出诊断结论和解决方案
4. 提供可执行的修复命令

### 常用命令

```bash
# 查看版本
aiassist version

# 配置向导
aiassist config

# 添加 Provider
aiassist config provider add

# 列出所有 Provider
aiassist config provider list

# 启用/禁用 Provider
aiassist config provider enable <name>
aiassist config provider disable <name>

# 删除 Provider
aiassist config provider delete <name>

# 查看帮助
aiassist --help
```

## 🔧 配置说明

### 配置文件

配置文件位于 `~/.aiassist/config.yaml`：

```yaml
language: zh  # zh=中文, en=English
http_proxy: ""  # HTTP 代理地址（可选）

providers:
  bailian:  # Provider 名称
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

### 模型调用顺序

**重要：模型调用顺序由配置文件中的排列顺序决定。**

例如上面的配置，调用顺序为：
1. `bailian/qwen-plus` (第一个启用的)
2. `bailian/qwen-turbo` (第二个启用的)
3. `openai/gpt-4` (第三个启用的)

如果当前模型调用失败、超时或不可用，会自动切换到下一个启用的模型。

### Provider 配置

#### 通义千问（阿里云百炼）

```bash
# 申请地址
https://dashscope.console.aliyun.com/apiKey

# 配置示例
Provider Name: bailian
Base URL: https://dashscope.aliyuncs.com/compatible-mode/v1
API Key: sk-xxxxxxxxxxxx
Models: qwen-plus,qwen-turbo,qwen-max
```

#### OpenAI

```bash
# 申请地址
https://platform.openai.com/api-keys

# 配置示例
Provider Name: openai
Base URL: https://api.openai.com/v1
API Key: sk-xxxxxxxxxxxx
Models: gpt-4,gpt-3.5-turbo
HTTP Proxy: http://127.0.0.1:7890  # 国内需要代理
```

#### 其他 OpenAI 兼容 API

任何实现 OpenAI API 标准的服务都可以配置：

```bash
Provider Name: custom
Base URL: https://your-api-endpoint/v1
API Key: your-api-key
Models: model-name-1,model-name-2
```

## 🛡️ 安全设计

### 命令分类

aiassist 将命令分为两类：

| 类型 | 标记 | 颜色 | 确认次数 | 示例 |
|------|------|------|---------|------|
| 查询命令 | `[cmd:query]` | 🟢 绿色 | 1次 | `ps aux`, `cat /etc/config`, `docker ps` |
| 修改命令 | `[cmd:modify]` | 🔴 红色 | 2次 | `systemctl restart`, `rm -rf`, `iptables -A` |

### 执行流程

1. **AI 分析**：理解问题，给出方案
2. **命令展示**：显示建议的命令及类型标记
3. **用户确认**：
   - 查询命令：确认1次即可执行
   - 修改命令：需要确认2次，防止误操作
4. **执行反馈**：显示执行结果，AI 继续分析

### 命令标记规范

AI 模型在生成命令时，会自动添加类型标记：

```
[cmd:query] ps aux | grep nginx    # 查询命令
[cmd:modify] systemctl restart nginx  # 修改命令
```

这些标记由 AI 模型根据系统提示词自动生成，工具会解析这些标记来确定命令类型。

## 🎯 使用场景

### 故障排查

```bash
You> Nginx 无法启动，报错 bind failed

AI> 端口可能被占用，让我们检查：
[cmd:query] netstat -tlnp | grep :80

# 发现端口被占用后...
AI> 发现进程 1234 占用了 80 端口，建议：
[cmd:modify] kill -9 1234
[cmd:modify] systemctl start nginx
```

### 性能分析

```bash
You> 服务器 CPU 负载过高

AI> 先检查进程资源占用：
[cmd:query] top -b -n 1 | head -20
[cmd:query] ps aux --sort=-%cpu | head -10

# 分析后...
AI> 发现 mysql 进程占用 CPU 较高，建议检查慢查询...
```

### 日志分析

```bash
tail -f /var/log/nginx/access.log | aiassist

AI> 检测到异常：
- IP 192.168.1.100 在1分钟内请求 500+ 次
- 大量 404 错误
- 疑似扫描攻击

建议：
[cmd:modify] iptables -A INPUT -s 192.168.1.100 -j DROP
```

### 配置检查

```bash
You> 检查 Nginx 配置是否正确

AI> 让我们检查配置文件：
[cmd:query] nginx -t
[cmd:query] cat /etc/nginx/nginx.conf

# 发现错误后...
AI> 配置文件第 45 行缺少分号，建议修改...
```

## 🔄 工作原理

```
┌──────────────┐
│  用户输入    │
│  或管道数据  │
└──────┬───────┘
       │
       ▼
┌──────────────────┐
│  系统信息收集    │
│  (OS/版本/工具)  │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  构建 Prompt     │
│  (问题+系统信息) │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  LLM 分析        │
│  (按配置顺序)    │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  提取命令        │
│  [cmd:query]     │
│  [cmd:modify]    │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  用户确认        │
│  (分类展示)      │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  执行命令        │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  结果反馈        │
│  (继续分析)      │
└──────────────────┘
```

## 📝 开发

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/llaoj/aiassist.git
cd aiassist

# 安装依赖
go mod download

# 构建（当前平台）
make build
# 或
./build.sh

# 构建所有平台
./scripts/build-all.sh

# 运行测试
go test ./...

# 运行
./aiassist
```

### 项目结构

```
aiassist/
├── cmd/aiassist/          # 主程序入口
├── internal/
│   ├── cmd/               # CLI 命令实现
│   │   ├── root.go       # 根命令
│   │   ├── config.go     # 配置命令
│   │   ├── interactive.go # 交互模式
│   │   └── version.go    # 版本命令
│   ├── config/            # 配置管理
│   ├── executor/          # 命令执行器
│   ├── i18n/              # 国际化
│   ├── interactive/       # 交互会话
│   ├── llm/               # LLM 管理器
│   │   ├── manager.go    # Provider 管理
│   │   └── openai_compatible.go # OpenAI 兼容接口
│   ├── prompt/            # 系统提示词
│   ├── sysinfo/           # 系统信息收集
│   └── ui/                # UI 工具
├── .github/workflows/     # CI/CD
└── scripts/               # 脚本目录
    ├── install.sh        # 一键安装脚本
    ├── build-all.sh      # 多平台构建
    └── test-install.sh   # 安装测试
```

### 技术栈

- **语言**: Go 1.21+
- **CLI 框架**: cobra
- **配置**: YAML
- **HTTP 客户端**: 标准库 net/http
- **测试**: Go testing

## 🤝 贡献

欢迎贡献代码、报告问题或提出建议！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 提交 Pull Request

## 📄 许可证

本项目采用 Apache License 2.0 许可证 - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

- [OpenAI](https://openai.com/) - API 标准参考
- [阿里云百炼](https://www.aliyun.com/product/bailian) - 通义千问支持
- [Cobra](https://github.com/spf13/cobra) - CLI 框架

## 📞 联系方式

- 问题反馈: [GitHub Issues](https://github.com/llaoj/aiassist/issues)
- 功能建议: [GitHub Discussions](https://github.com/llaoj/aiassist/discussions)

---

**⭐ 如果这个项目对你有帮助，欢迎点个 Star！**
