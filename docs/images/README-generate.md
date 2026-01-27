# 如何生成图片

## 已生成的Mermaid图表

我已经生成了以下 Mermaid 图表文件：

1. ✅ **architecture.mmd** - 系统架构图
2. ✅ **recursive-analysis.mmd** - 递归分析流程
3. ✅ **security-control.mmd** - 安全控制机制
4. ✅ **model-fallback.mmd** - 模型切换流程

## 生成PNG图片的3种方法

### 方法1: 使用 Mermaid Live Editor（推荐）

1. 打开 https://mermaid.live/
2. 复制对应的 `.mmd` 文件内容
3. 粘贴到左侧编辑器
4. 右侧会实时预览
5. 点击 "Actions" → "PNG" 下载PNG图片
6. 保存为对应的文件名（如 `architecture.png`）

### 方法2: 使用 Mermaid CLI

```bash
# 安装 Mermaid CLI
npm install -g @mermaid-js/mermaid-cli

# 批量生成PNG
mmdc -i architecture.mmd -o architecture.png
mmdc -i recursive-analysis.mmd -o recursive-analysis.png
mmdc -i security-control.mmd -o security-control.png
mmdc -i model-fallback.mmd -o model-fallback.png

# 或者一键生成所有
for file in *.mmd; do mmdc -i "$file" -o "${file%.mmd}.png"; done
```

### 方法3: 在 GitHub 上渲染

1. 将 `.mmd` 文件重命名为 `.md`
2. 在内容外包裹 mermaid 代码块：
   ````markdown
   ```mermaid
   graph TB
   ...
   ```
   ````
3. 在 GitHub 上查看文件会自动渲染
4. 截图保存

## 图片质量建议

- **分辨率**: 至少 1200x800
- **格式**: PNG（支持透明背景）
- **DPI**: 150-300（用于高清显示）
- **背景**: 建议使用浅色背景或透明背景

## 自定义样式

如果需要调整颜色、字体等，可以在 Mermaid Live Editor 中：
1. 点击 "Configuration"
2. 修改主题或自定义样式
3. 重新导出

## 已设置的颜色说明

- 🟢 绿色 (#90EE90) - 查询命令/成功状态
- 🔴 红色 (#FF6B6B) - 修改命令/警告
- 🔵 蓝色 (#87CEEB) - GPT-4/正常流程
- 🟡 黄色 (#FFD700) - 递归检查/确认提示
- 🟠 橙色 (#FFA500) - 错误/无效输入
- 🟣 粉色 (#FFB6C1) - DeepSeek

## 快速开始

```bash
# 1. 安装工具
npm install -g @mermaid-js/mermaid-cli

# 2. 进入图片目录
cd docs/images

# 3. 生成所有PNG
for file in *.mmd; do 
  mmdc -i "$file" -o "${file%.mmd}.png" -b transparent
done

# 4. 检查生成的文件
ls -lh *.png
```

完成后，你会得到：
- architecture.png
- recursive-analysis.png
- security-control.png
- model-fallback.png

这4个静态图就可以用在文章中了！
