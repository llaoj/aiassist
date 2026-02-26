# Consul 配置中心使用指南

## 快速开始

### 1. 启动 Consul

**使用 Docker（推荐）：**
```bash
docker run -d \
  --name consul \
  -p 8500:8500 \
  consul:latest agent -dev -ui -client=0.0.0.0
```

**或下载二进制：**
```bash
# macOS
brew install consul
consul agent -dev

# Linux
wget https://releases.hashicorp.com/consul/1.18.0/consul_1.18.0_linux_amd64.zip
unzip consul_1.18.0_linux_amd64.zip
sudo mv consul /usr/local/bin/
consul agent -dev
```

访问 Consul UI: http://localhost:8500/ui

### 2. 配置 aiassist 使用 Consul

编辑配置文件 `~/.aiassist/config.yaml`：

**配置中心模式（推荐）：**
```yaml
# 启用 Consul 配置中心
# 所有配置（language、http_proxy、default_model、providers）都从 Consul 加载
consul:
  enabled: true
  address: "127.0.0.1:8500"
  key: "aiassist/config"
  token: ""  # 可选，ACL token
```

**本地配置模式：**
```yaml
language: zh
http_proxy: ""
default_model: bailian/qwen-max

# 不配置 consul 或设置 enabled: false
# consul:
#   enabled: false

# 直接在本地配置 providers
providers:
  bailian:
    name: bailian
    base_url: https://dashscope.aliyuncs.com/compatible-mode/v1
    api_key: sk-xxxxxxxxxxxxxxxxxxxxxxxx
    enabled: true
    models:
      - name: qwen-max
        enabled: true
```

### 3. 写入配置到 Consul

**方法一：使用 Consul UI**
1. 访问 http://localhost:8500/ui/dc1/kv
2. 点击 "Create"
3. Key: `aiassist/config`
4. Value: 粘贴下面的 YAML 配置
5. 点击 "Save"

**配置内容示例：**
```yaml
language: zh
http_proxy: ""
default_model: bailian/qwen-max
providers:
  bailian:
    name: bailian
    base_url: https://dashscope.aliyuncs.com/compatible-mode/v1
    api_key: sk-xxxxxxxxxxxxxxxxxxxxxxxx
    enabled: true
    models:
      - name: qwen-max
        enabled: true
      - name: qwen-plus
        enabled: true
  deepseek:
    name: deepseek
    base_url: https://api.deepseek.com/v1
    api_key: sk-xxxxxxxxxxxxxxxxxxxxxxxx
    enabled: true
    models:
      - name: deepseek-chat
        enabled: true
```

**方法二：使用命令行**
```bash
# 创建配置文件
cat > /tmp/consul-providers.yaml << 'EOF'
language: zh
http_proxy: ""
default_model: bailian/qwen-max
providers:
  bailian:
    name: bailian
    base_url: https://dashscope.aliyuncs.com/compatible-mode/v1
    api_key: sk-xxxxxxxxxxxxxxxxxxxxxxxx
    enabled: true
    models:
      - name: qwen-max
        enabled: true
EOF

# 写入 Consul
consul kv put aiassist/config @/tmp/consul-providers.yaml

# 或使用 curl
curl -X PUT \
  --data-binary @/tmp/consul-providers.yaml \
  http://127.0.0.1:8500/v1/kv/aiassist/config
```

### 4. 使用 aiassist

```bash
# 直接使用（自动从 Consul 加载配置）
aiassist "检查服务器负载"

# 或交互模式
aiassist
```

## 配置模式说明

### 配置中心模式

在 `~/.aiassist/config.yaml` 中仅配置 Consul 连接信息：

```yaml
consul:
  enabled: true
  address: "127.0.0.1:8500"
  key: "aiassist/config"
  token: ""  # 可选
```

**工作流程：**
1. aiassist 读取本地配置文件
2. 检测到 `consul.enabled: true`
3. 连接 Consul，从 `key` 指定的位置加载**完整配置**
   - language（界面语言）
   - http_proxy（HTTP 代理）
   - default_model（默认模型）
   - providers（模型提供商配置）
4. 使用 Consul 中的配置运行

**优势：**
- ✅ **完全集中管理** - 所有配置在 Consul 中统一维护
- ✅ **统一更新** - 修改配置无需登录每台主机
- ✅ **适合企业批量部署** - 一次配置，全局生效

### 本地配置模式

在 `~/.aiassist/config.yaml` 中直接配置所有内容：

```yaml
language: zh
default_model: bailian/qwen-max
providers:
  bailian:
    name: bailian
    base_url: https://dashscope.aliyuncs.com/compatible-mode/v1
    api_key: sk-xxxxxxxxxxxxxxxxxxxxxxxx
    enabled: true
    models:
      - name: qwen-max
        enabled: true
```

**工作流程：**
1. aiassist 读取本地配置文件
2. 没有 consul 配置或 `consul.enabled: false`
3. 直接使用本地的完整配置

**优势：**
- ✅ 简单直接，无需额外服务
- ✅ 适合个人使用或单机部署

## 工作机制

### 配置加载流程

```
启动 aiassist
    ↓
读取 ~/.aiassist/config.yaml
    ↓
检查 consul.enabled?
    ↓
   是 → 连接 Consul → 加载完整配置
    ↓           ↓       (language, http_proxy, default_model, providers)
   否      连接失败 → 使用本地配置
```

### 配置修改方式

**配置中心模式：**

1. **通过 Consul UI 修改（推荐）**
   - 访问 http://consul-addr:8500/ui/dc1/kv
   - 找到 key（如 `aiassist/config`）
   - 点击 "Edit" 修改配置
   - 点击 "Save"
   - 重启 aiassist 生效

2. **通过命令行修改**
   ```bash
   # 导出当前配置
   consul kv get aiassist/config > config.yaml
   
   # 编辑配置
   vim config.yaml
   
   # 更新到 Consul
   consul kv put aiassist/config @config.yaml
   ```

3. **通过 API 修改**
   ```bash
   curl -X PUT \
     --data-binary @config.yaml \
     http://consul-addr:8500/v1/kv/aiassist/config
   ```

**本地模式：**
- 直接编辑 `~/.aiassist/config.yaml` 配置文件
- 使用 `aiassist config view` 查看当前配置

### 配置优先级

**配置中心模式（`consul.enabled: true`）：**
- language、http_proxy、default_model、providers **全部从 Consul 加载**
- 本地配置文件只需要 consul 连接信息
- ❌ **禁止本地修改** - 本地配置文件只读

**本地模式（无 consul 或 `enabled: false`）：**
- language、http_proxy、default_model、providers **全部从本地文件读取**
- ✅ **允许本地修改** - 直接编辑 `~/.aiassist/config.yaml`

### 配置保存

**配置中心模式（`consul.enabled: true`）：**
- ❌ **禁止本地修改** - 本地配置文件只读
- ✅ **统一在 Consul 修改** - 通过 Consul UI 或 `consul kv put` 命令修改
- 📝 **只读模式** - 本地只能查看配置，不能修改

**本地模式：**
- ✅ 直接编辑 `~/.aiassist/config.yaml`
- ✅ 使用 `aiassist config view` 查看当前配置

**示例：**

```bash
# 配置中心模式下查看配置
$ aiassist config view

# 正确做法：在 Consul UI 中修改或使用命令
$ consul kv put aiassist/config @updated-config.yaml
```

## 企业部署场景

### 场景 1：统一配置管理

**步骤：**

1. **在 Consul 中创建配置**（仅一次）

```bash
# 准备配置文件
cat > /tmp/company-aiassist.yaml << 'EOF'
language: zh
default_model: company/llm-v1
providers:
  company:
    name: company
    base_url: https://ai-internal.company.com/v1
    api_key: ${COMPANY_AI_KEY}
    enabled: true
    models:
      - name: llm-v1
        enabled: true
EOF

# 上传到 Consul
consul kv put production/aiassist/config @/tmp/company-aiassist.yaml
```

2. **在所有主机上部署统一配置**

```bash
# 创建 ~/.aiassist/config.yaml（所有主机配置相同）
cat > ~/.aiassist/config.yaml << 'EOF'
consul:
  enabled: true
  address: "consul.company.com:8500"
  key: "production/aiassist/config"
  token: ""
EOF
```

**说明：**
- 本地配置文件**极简**，只有 Consul 连接信息
- language、default_model、providers 等**全部从 Consul 加载**
- 所有主机使用相同的本地配置文件

3. **使用**

```bash
# 所有主机自动从 Consul 加载相同配置
aiassist "检查服务器负载"
```

**优势：**
- ✅ 配置文件统一，只需部署一次
- ✅ 修改配置在 Consul UI 操作，实时生效所有主机
- ✅ 便于审计和权限控制

### 场景 2：多环境隔离

不同环境使用不同的 Consul Key：

**开发环境 (~/.aiassist/config.yaml)：**
```yaml
consul:
  enabled: true
  address: "consul.company.com:8500"
  key: "dev/aiassist/config"
```

**测试环境 (~/.aiassist/config.yaml)：**
```yaml
consul:
  enabled: true
  address: "consul.company.com:8500"
  key: "test/aiassist/config"
```

**生产环境 (~/.aiassist/config.yaml)：**
```yaml
consul:
  enabled: true
  address: "consul.company.com:8500"
  key: "prod/aiassist/config"
```

**说明：**
- 本地文件只区分 `key`，指向不同环境的配置
- 各环境的完整配置（language、providers 等）在各自的 Consul Key 中维护

### 场景 3：容器化部署

**方式一：通过 ConfigMap 挂载配置文件**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: aiassist-config
data:
  config.yaml: |
    consul:
      enabled: true
      address: "consul-service:8500"
      key: "k8s/aiassist/config"
---
apiVersion: v1
kind: Pod
metadata:
  name: debug-pod
spec:
  containers:
  - name: shell
    image: ubuntu:latest
    volumeMounts:
    - name: config
      mountPath: /root/.aiassist
  volumes:
  - name: config
    configMap:
      name: aiassist-config
```

**方式二：使用 Init Container 生成配置**

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: debug-pod
spec:
  initContainers:
  - name: init-config
    image: busybox
    command:
    - sh
    - -c
    - |
      cat > /config/config.yaml << 'EOF'
      consul:
        enabled: true
        address: "consul-service:8500"
        key: "k8s/aiassist/config"
      EOF
    volumeMounts:
    - name: config
      mountPath: /config
  containers:
  - name: shell
    image: ubuntu:latest
    volumeMounts:
    - name: config
      mountPath: /root/.aiassist
  volumes:
  - name: config
    emptyDir: {}
```

**说明：**
- ConfigMap 中只包含 Consul 连接配置
- 实际的 language、providers 等配置在 Consul 的 `k8s/aiassist/config` 中
- 所有 Pod 使用统一的 ConfigMap

## ACL 安全配置

生产环境建议启用 Consul ACL：

### 1. 创建 Policy

```hcl
# aiassist-policy.hcl
key_prefix "aiassist/" {
  policy = "write"
}
```

### 2. 创建 Token

```bash
# 应用 Policy
consul acl policy create \
  -name aiassist-policy \
  -rules @aiassist-policy.hcl

# 创建 Token
consul acl token create \
  -description "aiassist config access" \
  -policy-name aiassist-policy
```

### 3. 使用 Token

```bash
export AIASSIST_CONSUL_TOKEN=your-generated-token
```

## 故障排查

### Consul 连接失败

如果 Consul 不可用，aiassist 会自动使用本地配置：

```bash
$ aiassist
# 自动降级到本地 providers 配置（如果有）
```

### 检查配置

```bash
# 查看 Consul 中的配置
consul kv get aiassist/config

# 或使用 curl
curl http://127.0.0.1:8500/v1/kv/aiassist/config?raw

# 查看本地配置文件
cat ~/.aiassist/config.yaml
```

### 调试技巧

**测试 Consul 连接：**
```bash
# 检查 Consul 是否可访问
curl http://127.0.0.1:8500/v1/status/leader

# 列出所有 key
consul kv get -recurse
```

**验证配置加载：**
```bash
# 查看当前使用的配置
aiassist config view
```

### 常见问题

**Q: 配置中心模式下，如何添加新的 Provider？**

A: 需要在 Consul UI 或通过 `consul kv put` 命令修改配置。

```bash
# 1. 导出当前配置
consul kv get aiassist/config > config.yaml

# 2. 编辑添加 provider
vim config.yaml

# 3. 更新到 Consul
consul kv put aiassist/config @config.yaml
```

**Q: 本地模式下如何修改配置？**

A: 直接编辑 `~/.aiassist/config.yaml` 文件即可。可使用 `aiassist config view` 查看当前配置。

**Q: 修改了 Consul 配置，为什么 aiassist 还是用旧配置？**

A: aiassist 每次启动时才从 Consul 加载配置。修改 Consul 配置后，重新运行 aiassist 即可。

**Q: 配置中心模式下，本地配置文件需要配置 language 和 providers 吗？**

A: **不需要**。只需要配置 consul 连接信息，language、http_proxy、default_model、providers 全部从 Consul 加载。

**Q: Consul 宕机了怎么办？**

A: aiassist 会自动降级使用本地配置文件中的配置（如果有）。建议保持本地文件作为备份。

**Q: 能否在本地配置 language，Consul 只管理 providers？**

A: 不建议。要么全部用 Consul（配置中心模式），要么全部用本地（本地模式），避免混淆。

**Q: 如何从配置中心模式切换到本地模式？**

A: 编辑 `~/.aiassist/config.yaml`，删除 `consul` 配置或设置 `consul.enabled: false`，然后添加 `providers` 配置。

## 最佳实践

1. **生产环境** - 使用 Consul 集群保证高可用
2. **ACL 控制** - 启用 ACL，限制配置访问权限
3. **配置备份** - 定期备份 Consul 数据目录
4. **监控告警** - 监控 Consul 服务状态
5. **版本控制** - 配置变更记录在 Git，通过 CI/CD 同步到 Consul
6. **配置修改流程**（配置中心模式）：
   - ✅ 在 Consul UI 或通过 `consul kv put` 修改
   - 📝 建立配置审批流程，避免误操作
   - 💡 本地只能通过 `aiassist config view` 查看配置
7. **模式选择**：
   - 企业批量部署 → 配置中心模式
   - 个人使用/单机 → 本地模式
   - 不要混用两种模式
