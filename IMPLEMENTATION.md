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
- 管道联动模式 - 处理 stdin 数据
- 上下文联动 - 自动传递命令执行结果
- 会话历史管理

### 5. **CLI 框架** (`internal/cmd/`)
- 使用 Cobra 框架实现命令行
- 基本命令：
  - `aiassist` - 进入交互模式
  - `aiassist config` - 配置管理
  - `aiassist version` - 显示版本

## 使用流程

### 1. 初始化配置
```bash
aiassist config
```
按提示完成模型选择、API Key 配置、语言偏好设置。

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

### ✅ 智能上下文联动
- 自动读取上一条命令的输出
- 作为下一次查询的输入参数
- 支持递进式问题排查

### ✅ 命令执行风险管控
- 查询命令 → 绿色展示 → 自动执行
- 修改命令 → 红色展示 → 二次确认
- 最大限度规避误操作

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

1. **实现实际的 API 调用**
   - 集成通义千问、ChatGPT、DeepSeek 的官方 SDK
   - 处理 API 响应和错误

2. **增强命令分类**
   - 更全面的命令关键词库
   - 基于模式的命令风险评估

3. **持久化历史**
   - 将会话历史保存到数据库
   - 支持历史查询和重放

4. **性能优化**
   - 异步命令执行
   - 缓存 LLM 响应
   - 连接池管理

5. **监控和日志**
   - 详细的操作日志
   - 模型调用统计
   - 性能指标收集
