# 命令黑名单功能实现总结

## 实现概述

成功实现了命令执行的黑名单机制，该机制在两个层面工作：
1. **AI提示层**：告知AI黑名单内容，尽量避免生成黑名单命令
2. **执行拦截层**：执行前检查并拒绝黑名单命令

## 核心功能

### 1. 配置文件支持

**文件**: `internal/config/config.go`

在配置结构体中添加了 `Blacklist` 字段：

```go
type Config struct {
    Language     string            `yaml:"language"`
    DefaultModel string            `yaml:"default_model"`
    Consul       *ConsulConfig     `yaml:"consul,omitempty"`
    Providers    []*ProviderConfig `yaml:"providers"`
    Blacklist    []string          `yaml:"blacklist,omitempty"` // Command blacklist
    // ...
}
```

添加了获取命令黑名单的方法：

```go
func (c *Config) GetBlacklist() []string
```

### 2. 命令黑名单检查模块

**文件**: `internal/blacklist/blacklist.go`

核心功能：
- `IsBlacklisted(command string) (bool, string)`: 检查命令是否在黑名单中
- `FormatBlacklistForPrompt() string`: 格式化黑名单用于AI提示

模式匹配规则：
1. 词级匹配：pattern 和命令都按空格分词后逐词比较
2. 尾部 `*` 匹配剩余所有参数（至少一个）：
   - `rm *` 匹配 `rm -rf /` 和 `rm file.txt`，但不匹配 `rm`
   - `kubectl delete *` 匹配 `kubectl delete pod nginx`，但不匹配 `kubectl delete`
3. 无 `*` 时只需前缀词匹配（允许额外参数）：
   - `shutdown` 匹配 `shutdown` 和 `shutdown -h now`
   - `rm -rf` 匹配 `rm -rf /tmp`，但不匹配 `rm -r /tmp`
4. 首词取 base name：`/usr/bin/rm` 等同于 `rm`

### 3. 执行器集成

**文件**: `internal/executor/executor.go`

修改内容：
- 添加 `blacklistChecker` 字段
- `DisplayCommand()` 方法显示黑名单警告
- 新增 `IsBlacklisted()` 方法

**文件**: `internal/interactive/session.go`

修改内容：
- `handleCommands()` 方法中添加黑名单检查
- 黑名单命令被拒绝后，将拒绝信息添加到对话历史
- AI收到拒绝信息后可以提供替代方案

### 4. AI提示集成

**文件**: `internal/prompt/prompts.go` 和 `internal/prompt/prompts_base.go`

修改内容：
- 修改了 `GetInteractivePrompt()`、`GetContinueAnalysisPrompt()`、`GetPipeAnalysisPrompt()`
- 通过 `injectBlacklist()` 函数将黑名单信息注入系统提示

AI被告知的黑名单规则：
```
[Command Blacklist]:
- rm *
- dd *
- kubectl delete *

The above commands are blacklisted and forbidden to execute. You should:
1. Avoid generating these commands - use alternatives when possible
2. If a blacklisted command is absolutely necessary, clearly inform the user:
   - State that the command is in the blacklist
   - Explain why permission is needed
3. Never assume blacklisted commands will execute successfully
```

### 5. 国际化支持

**文件**: `internal/i18n/messages_zh.go` 和 `messages_en.go`

新增消息键：
- `executor.blacklisted`: 黑名单拒绝消息
- `executor.blacklist_hint`: 黑名单提示信息
- `executor.blacklist_required`: 黑名单命令警告（显示在命令旁）

### 6. 配置示例更新

**文件**: `config.example.yaml`

添加了黑名单配置示例和说明：

```yaml
blacklist:
  - "rm *"               # 禁止所有带参数的 rm 命令
  - "dd *"               # 禁止 dd 命令（危险磁盘操作）
  - "kubectl delete *"   # 禁止 kubectl delete 操作
  - "shutdown"           # 禁止 shutdown 命令（无论是否有参数）
```

### 7. 文档更新

**文件**: `README.md` 和 `README_EN.md`

新增"命令黑名单"章节，包括：
- 功能说明
- 配置示例
- 模式匹配规则
- 工作流程图
- 使用场景示例

### 8. 单元测试

**文件**: `internal/blacklist/blacklist_test.go`

测试覆盖：
- `TestIsBlacklisted`: 测试黑名单匹配逻辑
  - 通配符匹配测试（`rm *` 匹配有参数的，不匹配无参数的）
  - 多词 pattern 测试
  - 精确匹配测试
  - 无 `*` 时前缀匹配测试
  - base name 归一化测试
  - 空黑名单测试
- `TestFormatBlacklistForPrompt`: 测试提示格式化

所有测试通过：✅

## 工作流程

```
用户提问
    ↓
AI 分析（被告知黑名单）
    ↓
生成命令建议（可能包含黑名单命令）
    ↓
显示命令（如果匹配黑名单，显示警告）
    ↓
用户确认执行
    ↓
系统检查黑名单
    ├─ 匹配 → 拒绝执行，返回拒绝信息给 AI
    └─ 不匹配 → 执行命令
```

## 使用示例

### 配置示例

```yaml
language: zh
default_model: bailian/qwen-max

blacklist:
  - "rm *"
  - "dd *"
  - "kubectl delete *"

providers:
  - name: bailian
    base_url: https://dashscope.aliyuncs.com/compatible-mode/v1
    api_key: sk-xxx
    enabled: true
    models:
      - name: qwen-max
        enabled: true
```

### 运行示例

```
You> 删除日志文件

AI> 建议执行以下命令：
[修改命令]
rm /var/log/app.log
注意: 该命令匹配黑名单规则 'rm *'，属于禁止执行的命令。
如必须使用，请先向用户申请权限

是否执行? (yes/no): yes

✗ 命令被拒绝: 该命令匹配黑名单规则 'rm *'，禁止执行
如需执行此命令，请联系管理员申请权限或修改黑名单配置

AI> 由于 rm 命令在黑名单中，建议使用其他方式：
1. 使用 truncate 清空文件内容
[修改命令]
truncate -s 0 /var/log/app.log
```

## 技术细节

### 模式匹配实现

使用词级匹配实现：

```go
func matchPattern(pattern string, cmdParts []string) bool {
    patParts := strings.Fields(pattern)
    // ... 首词归一化为 base name ...

    hasTrailingWildcard := patParts[len(patParts)-1] == "*"
    if hasTrailingWildcard {
        patParts = patParts[:len(patParts)-1]
    }

    // 带 * 时命令词数必须大于 pattern 词数
    // 不带 * 时命令词数只需 >= pattern 词数
    if hasTrailingWildcard {
        if len(cmdParts) <= len(patParts) {
            return false
        }
    } else {
        if len(cmdParts) < len(patParts) {
            return false
        }
    }

    // 逐词比较
    for i, p := range patParts {
        if cmdParts[i] != p {
            return false
        }
    }
    return true
}
```

### 黑名单信息传递

黑名单信息通过两个渠道传递：
1. **系统提示**：添加到所有场景的system prompt中
2. **拒绝反馈**：执行拒绝时，拒绝信息作为user message添加到对话历史

这确保AI能够：
- 提前知道哪些命令被禁止
- 看到拒绝结果后提供替代方案

## 测试结果

- ✅ 单元测试全部通过（17 个测试用例）
- ✅ 编译成功无错误
- ✅ 代码符合Go规范

## 文件清单

### 核心文件
- `internal/blacklist/blacklist.go`: 命令黑名单检查模块
- `internal/blacklist/blacklist_test.go`: 单元测试

### 集成文件
- `internal/config/config.go`: 添加命令黑名单配置支持
- `internal/executor/executor.go`: 集成命令黑名单检查
- `internal/interactive/session.go`: 执行前命令黑名单拦截
- `internal/prompt/prompts.go`: 命令黑名单提示注入
- `internal/prompt/prompts_base.go`: 黑名单提示模板
- `internal/i18n/messages_zh.go`: 中文消息
- `internal/i18n/messages_en.go`: 英文消息
- `config.example.yaml`: 配置示例
- `README.md`: 中文文档
- `README_EN.md`: 英文文档

## 符合项目规范

✅ 代码注释全部使用英文
✅ 用户界面支持中英文双语
✅ 文档与代码同步更新
✅ 遵循现有架构模式
✅ 添加完整单元测试
