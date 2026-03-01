package i18n

// EnglishMessages contains all UI messages in English
var EnglishMessages = map[string]string{
	// Config messages
	"config.not_found":      "✗ Configuration file not found",
	"config.hint_run_setup": "Please edit config file: ~/.aiassist/config.yaml",

	// Interactive mode messages
	"interactive.welcome":            "Welcome to AI Shell Assistant",
	"interactive.exit_hint":          "Tip: Press Ctrl+C anytime to exit",
	"interactive.input_prompt":       "Please enter your question: ",
	"interactive.goodbye":            "Goodbye!",
	"interactive.thinking":           "Thinking",
	"interactive.continue_analysis":  "Based on the complete conversation history and the executed command output above, please continue with the next steps of analysis and diagnosis, listing the remaining steps and commands.",
	"interactive.executed_command":   "Executed Command",
	"interactive.execution_output":   "Execution Output",
	"interactive.execution_error":    "Execution Error",
	"interactive.user_label":         "User",
	"interactive.ai_label":           "AI",
	"interactive.analysis_complete":  "✓ Analysis complete, please continue with questions",
	"interactive.pipe_user_question": "User question: ",
	"interactive.pipe_data":          "Pipe output data:",
	"interactive.pipe_source":        "Data source: piped input",

	// Executor messages
	"executor.query_command":     "Query command:",
	"executor.modify_command":    "Modify command (requires confirmation):",
	"executor.execute_prompt":    "Execute this command?",
	"executor.modify_warning":    "Warning: This command will modify server configuration, are you sure?",
	"executor.executing":         "Executing",
	"executor.execute_success":   "✓ Execution successful",
	"executor.execute_failed":    "✗ Execution failed: %v",
	"executor.no_output":         "(Command executed successfully, but no output)",
	"executor.max_depth_reached": "Warning: Maximum command analysis depth reached. Stopping to prevent infinite recursion.",

	// Blacklist messages
	"executor.blacklisted":        "✗ Command rejected: This command matches blacklist rule '%s', execution forbidden",
	"executor.blacklist_hint":     "To execute this command, please contact the administrator for permission or modify the blacklist configuration",
	"executor.blacklist_required": "Note: This command matches blacklist rule '%s' and is forbidden. If you must use it, please request permission from the user first",

	// Output truncation messages
	"output.truncated": "omitted %d lines of output",

	// Error messages
	"error.no_models":      "✗ Error: No models configured",
	"error.hint_no_models": "Please edit config file first: ~/.aiassist/config.yaml",
	"error.general":        "✗ Error: %v",

	// Version messages
	"version.app_name":   "AI Shell Assistant (aiassist)",
	"version.version":    "Version: %s",
	"version.commit":     "Commit: %s",
	"version.build_date": "Build Date: %s",

	// Model status messages
	"llm.status_title":   "Current Model",
	"llm.status_default": "(Default)",
}
