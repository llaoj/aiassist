package i18n

// ChineseMessages contains all UI messages in Chinese
var ChineseMessages = map[string]string{
	// Config messages
	"config.not_found":      "✗ 配置文件不存在",
	"config.hint_run_setup": "请编辑配置文件: ~/.aiassist/config.yaml",

	// Interactive mode messages
	"interactive.welcome":            "欢迎使用 AI Shell Assistant",
	"interactive.exit_hint":          "提示: 随时按 Ctrl+C 退出",
	"interactive.input_prompt":       "请输入问题: ",
	"interactive.goodbye":            "再见！",
	"interactive.thinking":           "思考中",
	"interactive.continue_analysis":  "根据以上完整的对话历史和已执行的命令输出，请继续进行接下来的分析和诊断，列出剩余的步骤和命令。",
	"interactive.executed_command":   "执行命令",
	"interactive.execution_output":   "执行输出",
	"interactive.execution_error":    "执行错误",
	"interactive.user_label":         "用户",
	"interactive.ai_label":           "AI",
	"interactive.analysis_complete":  "✓ 所有分析已完成",
	"interactive.pipe_user_question": "用户问题: ",
	"interactive.pipe_data":          "管道输出数据:",
	"interactive.pipe_source":        "数据来源: 通过管道输入",

	// Executor messages
	"executor.query_command":     "查询命令:",
	"executor.modify_command":    "修改命令 (需要确认):",
	"executor.execute_prompt":    "是否执行此命令?",
	"executor.modify_warning":    "警告: 该命令将修改服务器配置，是否确定执行?",
	"executor.executing":         "执行中",
	"executor.execute_success":   "✓ 执行成功",
	"executor.execute_failed":    "✗ 执行失败: %v",
	"executor.no_output":         "(命令执行成功，但没有输出)",
	"executor.max_depth_reached": "警告: 已达到最大命令分析深度。停止以防止无限递归。",

	// Blacklist messages
	"executor.blacklisted":        "✗ 命令被拒绝: 该命令匹配黑名单规则 '%s'，禁止执行",
	"executor.blacklist_hint":     "如需执行此命令，请联系管理员申请权限或修改黑名单配置",
	"executor.blacklist_required": "注意: 该命令匹配黑名单规则 '%s'，属于禁止执行的命令。如必须使用，请先向用户申请权限",

	// Output truncation messages
	"output.truncated": "省略 %d 行输出",

	// Error messages
	"error.no_models":      "✗ 错误: 未配置任何模型",
	"error.hint_no_models": "请先编辑配置文件: ~/.aiassist/config.yaml",
	"error.general":        "✗ 错误: %v",

	// Version messages
	"version.app_name":   "AI Shell Assistant (aiassist)",
	"version.version":    "版本: %s",
	"version.commit":     "提交: %s",
	"version.build_date": "构建日期: %s",

	// Model status messages
	"llm.status_title":   "当前模型",
	"llm.status_default": "(默认)",
}
