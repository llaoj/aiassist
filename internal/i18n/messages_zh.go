package i18n

// ChineseMessages contains all UI messages in Chinese
var ChineseMessages = map[string]string{
	// Config messages
	"config.not_found":      "✗ 配置文件不存在",
	"config.hint_run_setup": "请编辑配置文件: ~/.aiassist/config.yaml",
	"config.title":          "AI Shell Assistant - 配置查看",

	// OpenAI-compatible configuration
	"config.openai_compat.title":         "配置 OpenAI 兼容的大模型提供商",
	"config.openai_compat.model":         "模型",
	"config.openai_compat.provider_name": "输入提供商名称 (例如: 我的Qwen、我的DeepSeek): ",
	"config.openai_compat.name_empty":    "✗ 提供商名称不能为空",
	"config.openai_compat.base_url":      "输入 Base URL (例如: https://api.openai.com/v1): ",
	"config.openai_compat.url_empty":     "✗ Base URL 不能为空",
	"config.openai_compat.api_key":       "输入 API Key: ",
	"config.openai_compat.key_empty":     "✗ API Key 不能为空",
	"config.openai_compat.model_name":    "输入模型名称 (例如: qwen-plus, gpt-4o, deepseek-chat): ",
	"config.openai_compat.model_empty":   "✗ 模型名称不能为空",
	"config.openai_compat.success":       "✓ 提供商 '%s' 配置成功",
	"config.openai_compat.add_more":      "继续添加其他模型吗? (yes/no): ",
	"config.openai_compat.no_models":     "✗ 未配置任何模型",
	"config.openai_compat.added":         "✓ 提供商 '%s' 添加成功",
	"config.openai_compat.models_list":   "模型: %v",
	"config.openai_compat.order_hint":    "提示: 模型的调用顺序按照配置文件中的顺序。当一个模型不可用时，将自动尝试下一个模型。",

	// Default model selection
	"config.default_model.title":    "选择默认模型",
	"config.default_model.select":   "输入您偏好的模型序号 (默认: 1): ",
	"config.default_model.selected": "✓ 默认模型已设置为: %s",

	// Proxy configuration
	"config.proxy.title":   "配置代理 (可选)",
	"config.proxy.desc":    "如果您的网络需要代理，可以在这里配置",
	"config.proxy.example": "代理地址示例: http://127.0.0.1:7890",
	"config.proxy.input":   "输入代理地址 (留空则不使用代理): ",
	"config.proxy.success": "✓ 代理已配置: %s",
	"config.proxy.empty":   "✓ 未配置代理",
	"config.proxy.error":   "✗ 代理配置失败: %v",

	"config.complete":   "✓ 配置已保存到 ~/.aiassist/config.yaml",
	"config.save_error": "✗ 保存配置失败: %v",

	// Interactive mode messages
	"interactive.welcome":              "欢迎使用 AI Shell Assistant",
	"interactive.input_prompt":         "请输入问题: ",
	"interactive.goodbye":              "再见！",
	"interactive.commands_found":       "发现以下建议命令:",
	"interactive.cancelled":            "已取消",
	"interactive.thinking":             "思考中",
	"interactive.continue_analysis":    "根据以上完整的对话历史和已执行的命令输出，请继续进行接下来的分析和诊断，列出剩余的步骤和命令。",
	"interactive.executed_command":     "执行命令",
	"interactive.execution_output":     "执行输出",
	"interactive.user_label":           "用户",
	"interactive.ai_label":             "AI",
	"interactive.all_commands_skipped": "所有命令均已跳过",
	"interactive.analysis_complete":    "✓ 所有分析已完成",
	"interactive.all_steps_complete":   "所有分析步骤已完成。是否继续提问?: ",
	"interactive.pipe_user_question":   "用户问题: ",
	"interactive.pipe_data":            "管道输出数据:",
	"interactive.pipe_source":          "数据来源: 通过管道输入",

	// Executor messages
	"executor.query_command":        "查询命令:",
	"executor.modify_command":       "修改命令 (需要确认):",
	"executor.unclassified_command": "未分类命令:",
	"executor.execute_prompt":       "是否执行此命令?",
	"executor.modify_warning":       "警告: 该命令将修改服务器配置，是否确定执行?",
	"executor.executing":            "执行中",
	"executor.execute_success":      "✓ 执行成功",
	"executor.execute_failed":       "✗ 执行失败: %v",
	"executor.no_output":            "(命令执行成功，但没有输出)",
	"executor.cancelled":            "已取消",
	"executor.read_input_failed":    "读取输入失败: %v",
	"executor.max_depth_reached":    "警告: 已达到最大命令分析深度。停止以防止无限递归。",

	// Output truncation messages
	"output.truncated": "省略 %d 行输出",

	// Error messages
	"error.no_models":      "✗ 错误: 未配置任何模型",
	"error.hint_no_models": "请先编辑配置文件: ~/.aiassist/config.yaml",
	"error.unknown_model":  "!!! 警告: 未知模型 %s",

	// Version messages
	"version.app_name":   "AI Shell Assistant (aiassist)",
	"version.version":    "版本: %s",
	"version.commit":     "提交: %s",
	"version.build_date": "构建日期: %s",

	// Model status messages
	"llm.status_title":       "当前模型状态",
	"llm.status_available":   "✓ 可用",
	"llm.status_unavailable": "✗ 不可用",
	"llm.remaining_calls":    "剩余额度",
	"llm.priority":           "优先级",

	// UI messages
	"ui.ctrlc_exit_hint":    "再按一次 Ctrl+C 退出程序",
	"ui.ctrlc_exit_message": "用户通过 Ctrl+C 退出，再见！",
}
