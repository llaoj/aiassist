# AI Shell Assistant - Go 实现

这是基于设计文档的完整 Go 语言实现。

## 项目结构

```
aiassist/
├── cmd/
│   └── aiassist/
│       └── main.go              # 程序入口
├── internal/
│   ├── cmd/
│   │   ├── root.go              # CLI 主命令
│   │   ├── config.go            # config 命令 - 配置管理
│   │   ├── version.go           # version 命令 - 版本查看
│   │   └── interactive.go       # 交互模式入口
│   ├── config/
│   │   └── config.go            # 配置管理和持久化
│   ├── llm/
│   │   ├── provider.go          # LLM 提供商接口和实现
│   │   └── manager.go           # LLM 管理器 - 多模型切换
│   ├── executor/
│   │   └── executor.go          # 命令执行和分类
│   └── interactive/
│       └── session.go           # 交互会话管理
├── go.mod                       # Go 模块定义
├── go.sum                       # Go 模块锁定
└── README.md
```

## 核心功能模块

### 1. **配置管理** (`internal/config/config.go`)
- 配置文件存储在 `~/.aiassist/config.yaml`
- 支持多模型配置、API Key 管理、语言偏好、代理配置
- 自动校验 API Key 有效性
- 每日调用次数追踪和自动重置

### 2. **LLM 提供商** (`internal/llm/`)
- `provider.go` - 支持三个模型：通义千问、ChatGPT、DeepSeek
- `manager.go` - 管理多个提供商的生命周期和自动切换
  - 优先级管理
  - 自动故障转移
  - 额度监控
  - 每日配额重置

### 3. **命令执行器** (`internal/executor/executor.go`)
- 命令分类：查询类（绿色）vs 修改类（红色）
- 智能提取 AI 响应中的命令
- 二次确认机制防止误操作
- 支持与管道输入集成

### 4. **交互会话** (`internal/interactive/session.go`)
- 直接 AI 模式 - 连续对话
- 管道联动模式 - 处理 stdin 数据（最多1.6MB，~13k行nginx日志）
- 上下文联动 - 自动传递命令执行结果，最多10层递归分析
- 会话历史管理
- 使用 liner 库支持中文输入、光标移动、Ctrl+C退出
- 严格输入验证：确认提示仅接受 y/n/exit 输入

### 5. **CLI 框架** (`internal/cmd/`)
- 使用 Cobra 框架实现命令行
- 基本命令：
  - `aiassist` - 进入交互模式
  - `aiassist config view` - 查看配置
  - `aiassist version` - 显示版本

## 使用流程

### 1. 配置
直接编辑配置文件 `~/.aiassist/config.yaml`，配置 Provider 和 API Key。

### 2. 直接交互模式
```bash
aiassist
```
进入交互式会话，支持连续对话和上下文联动。

### 3. 管道联动模式
```bash
tail -f /var/log/nginx/access.log | aiassist "请分析这些日志是否有异常"
```

## 实现亮点

### ✅ 多模型动态切换
- 按优先级尝试模型
- 自动故障转移
- 终端提示模型变更原因

### ✅ 智能上下文管理
- 递归深度限制：最多10层命令分析防止无限递归
- 内存保护：管道输入限制400K字符（~1.6MB，支持13,000行nginx日志）
- 命令输出截断：100K字符限制，保留头尾关键信息
- 自动读取上一条命令的输出并传递给 AI
- 支持递进式问题排查

### ✅ 命令执行风险管控
- 查询命令 → 绿色展示 → 一次确认
- 修改命令 → 红色展示 → 二次确认
- 严格输入验证：仅接受 y/n/exit，无效输入循环提示
- 最大限度规避误操作

### ✅ 输入处理增强
- 使用 liner 库替代 readline，完美支持中文输入
- 支持光标移动、删除字符等编辑操作
- Ctrl+C 处理：所有输入点统一支持退出
- 管道模式简化：非交互式，仅显示分析结果

### ✅ 额度管理
- 实时显示剩余额度：`[模型][剩余/总数]`
- 每日自动重置配额
- 多模型并行额度跟踪

### ✅ 配置安全
- API Key 本地加密存储
- 无需第三方服务器验证
- 用户完全控制

## 依赖

```go
- github.com/spf13/cobra - CLI 框架
- github.com/spf13/viper - 配置管理
- github.com/fatih/color - 终端彩色输出
- github.com/peterh/liner - 行编辑，支持中文输入和光标移动
- gopkg.in/yaml.v3 - YAML 解析
```

## 编译和运行

```bash
# 编译
go build -o aiassist ./cmd/aiassist

# 运行
./aiassist config
./aiassist
echo "some data" | ./aiassist "analyze this"
```

## 后续完善方向

1. **性能优化**
   - 异步命令执行
   - 缓存 LLM 响应
   - 连接池管理

2. **持久化历史**
   - 将会话历史保存到数据库
   - 支持历史查询和重放

3. **监控和日志**
   - 详细的操作日志
   - 模型调用统计
   - 性能指标收集

4. **上下文窗口优化**
   - 自适应模型上下文限制
   - 智能历史截断
   - Token 使用估算和控制
