# AI Shell Assistant：让运维人员用自然语言"聊"服务器

> 🤖 一个完全由 AI 编写的开源运维工具，支持中文交互，多模型自动切换

![AI Shell Assistant Logo](../logo.svg)

## 写在前面

作为运维工程师，你是否遇到过这些场景：

- 凌晨3点被告警吵醒，脑子一片混乱不知道从哪里开始排查
- 面对陌生的系统，不知道该用什么命令诊断问题  
- 需要查手册、搜 StackOverflow，才能找到合适的诊断命令
- nginx 日志几万行，不知道如何快速定位异常

现在，你可以直接问 AI："**服务器负载很高，帮我排查原因**"

## 项目简介

**AI Shell Assistant (aiassist)** 是一个基于大语言模型的智能终端工具，通过自然语言交互为运维人员提供诊断分析、方案建议和命令执行指导。

**GitHub**: https://github.com/llaoj/aiassist

![项目架构图](images/architecture.png)
*▲ 系统架构：从用户输入到命令执行的完整流程*

### 核心特点

✅ **自然语言交互** - 用人话提问，AI 给出专业诊断  
✅ **智能命令建议** - 自动分析并给出最合适的 Shell 命令  
✅ **连续对话分析** - 自动传递上一条命令的输出，支持最多10层递归诊断  
✅ **管道模式** - 直接分析日志输出，支持13,000行nginx日志  
✅ **多模型支持** - 通义千问、GPT-4、DeepSeek等，自动切换  
✅ **安全可控** - 查询命令和修改命令差异化展示，严格确认机制  
✅ **完美中文** - 支持中文输入、光标编辑、Ctrl+C退出

## 实际使用场景

### 场景1：服务器负载异常排查

![服务器排查演示](images/scenario-1.gif)
*▲ 真实场景演示：AI 一步步引导排查问题*

```bash
$ aiassist "服务器负载很高，帮我排查原因"

[AI] 让我们先检查系统负载和CPU使用情况：

[查询命令]
top -b -n 1 | head -20

是否执行此命令? (y/n, exit退出): y

[执行成功]
top - 03:24:15 up 127 days,  5:32,  2 users,  load average: 8.45, 7.23, 6.89
Tasks: 256 total,   3 running, 253 sleeping...
  PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
 1234 nginx     20   0 2.156g 1.234g  12.3m R  98.3  15.6   124:56 nginx

[AI] 发现 nginx 进程 (PID 1234) CPU占用98.3%，内存使用1.2GB。
让我们查看nginx的连接状态：

[查询命令]
ss -ant | grep :80 | wc -l

是否执行此命令? (y/n, exit退出): y

[执行成功]
15234

[AI] 发现当前有15,234个活跃连接到80端口。建议检查是否遭受DDoS攻击...
```

### 场景2：管道模式分析日志

![管道模式演示](images/scenario-2.gif)
*▲ 管道模式：直接分析日志输出，支持13,000行数据*

```bash
# 直接分析 nginx 访问日志
$ tail -1000 /var/log/nginx/access.log | aiassist "分析访问日志"

[AI] 分析结果：
1. 发现来自 192.168.1.100 的异常请求：1分钟内发起了523次请求
2. 路径 /api/login 被频繁访问，可能存在暴力破解
3. 建议执行：
   - 检查该IP的请求模式
   - 临时封禁该IP
   - 启用限流策略

# 分析 Docker 容器状态
$ docker ps -a | aiassist "帮我看看容器有什么问题"

[AI] 发现3个容器处于Exited状态：
- redis-cache (退出码: 137) - 内存不足被OOM Killer杀死
- app-worker (退出码: 1) - 启动失败，建议查看日志
- nginx-proxy (退出码: 0) - 正常退出

```

### 场景3：交互式问题诊断

```bash
$ aiassist

??? 请输入问题: 为什么我的磁盘满了

[AI] 让我们先查看磁盘使用情况...
[查询命令] df -h

??? 请输入问题: /var 目录占用最大

[AI] 让我们查看/var目录下哪些子目录占用空间最大...
[查询命令] du -sh /var/* | sort -hr | head -10

??? 请输入问题: /var/log占了80GB

[AI] 让我们查看具体是哪些日志文件...
```

## 技术亮点

### 1. 智能递归分析

支持最多**10层递归命令分析**，AI会：
- 自动读取每条命令的输出
- 将输出作为上下文传递给下一轮分析
- 逐步深入定位问题根源
![递归分析流程](images/recursive-analysis.png)
*▲ 递归分析流程：从用户提问到最终诊断的完整路径（最多10层）*
### 2. 内存保护机制

![安全控制机制](docs/images/security-control.png)
*▲ 双重确认机制：查询命令绿色，修改命令红色+二次确认*
- 防止内存耗尽和无限递归

### 3. 严格的安全控制

```
查询命令（绿色） → 一次确认 → 执行
![中文输入演示](docs/images/chinese-input.gif)
*▲ 完美支持中文输入、光标编辑、删除等操作*

![安全控制机制](images/security-control.png)
*▲ 安全控制流程：查询命令（绿色）一次确认，修改命令（红色）二次确认*

```
查询命令（绿色） → 一次确认 → 执行- 修改类命令（rm、chmod等）强制二次确认
- 最大限度防止误操作

### 4. 完美的中文支持
![多模型切换](docs/images/model-fallback.png)
*▲ 自动切换流程：Qwen失败 → GPT-4 → DeepSeek*
- ✅ 支持中文输入编辑
- ✅ 光标左右移动
- ✅ 删除字符时提示符不消失
- ✅ Ctrl+C 优雅退出

### 5. 多模型自动切换

```yaml
# 配置文件按优先级排列
providers:
  qwen:      # 优先使用
    enabled: true
  gpt4:      # Qwen失败后切换
![多模型切换](images/model-fallback.png)
*▲ 多模型自动切换：优先级顺序调用，失败自动切换到下一个*

```yaml
# 配置文件按优先级排列
providers:
  qwen:      # 优先使用
    enabled: true
  gpt4:      # Qwen失败后切换
    enabled: true
  deepseek:  # 最后备选## 快速开始

### 一键安装

![配置向导](docs/images/config-wizard.gif)
*▲ 交互式配置向导：step-by-step 引导*

```bash
curl -fsSL https://raw.githubusercontent.com/llaoj/aiassist/main/scripts/install.sh | bash
```
- macOS (Intel, Apple Silicon)
- Windows (x86_64, ARM64, i386)
- FreeBSD (x86_64, ARM64)

### 配置大模型

```bash
aiassist config
```

按向导提示配置：
1. 选择语言（中文/English）
2. 添加 LLM Provider（支持 OpenAI 兼容接口）
3. 输入 API Key
4. 配置模型列表
5. （可选）设置 HTTP 代理

支持的模型：
- 通义千问（Qwen）
- OpenAI GPT-4/3.5
- DeepSeek
- 其他 OpenAI 兼容接口

### 开始使用

```bash
# 交互式对话
aiassist

# 单次问答
aiassist "服务器负载很高，帮我排查"

# 管道模式
tail -f /var/log/nginx/access.log | aiassist "分析日志"
docker ps -a | aiassist "检查容器状态"
```

## 项目背景

这个项目有个特别之处：**完全由 AI 编写**。

从需求分析、架构设计、代码实现、文档编写，全部由 AI（Claude）完成。这也从侧面证明了 AI 在代码生成领域的能力已经相当成熟。

技术栈：
- Go 1.21+
- Cobra (CLI框架)
- Liner (行编辑)
- Viper (配置管理)

## 适用人群

✅ **运维工程师** - 快速诊断生产环境问题  
✅ **SRE** - 自动化故障排查流程  
✅ **DevOps** - CI/CD 环境调试  
✅ **新手运维** - 学习 Shell 命令最佳实践  
✅ **开发者** - 快速解决服务器问题

## 开源协议

Apache 2.0 - 完全开源免费

## 后续计划

- [ ] 持久化会话历史
- [ ] 性能优化（异步执行、响应缓存）
- [ ] 监控和日志系统
- [ ] 自适应模型上下文窗口
- [ ] 更多模型支持

## 参与贡献

欢迎提交 Issue 和 Pull Request！

项目地址：https://github.com/llaoj/aiassist

---

**如果觉得有用，请给个 ⭐️ Star 支持一下！**

## 结语

在 AI 时代，我们不再需要记住成百上千条 Shell 命令，只需要用自然语言描述问题，AI 会给出最合适的解决方案。

AI Shell Assistant 让运维工作变得更简单、更高效、更智能。

立即体验：
```bash
curl -fsSL https://raw.githubusercontent.com/llaoj/aiassist/main/scripts/install.sh | bash
```
