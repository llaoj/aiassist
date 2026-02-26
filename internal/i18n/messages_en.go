package i18n

// EnglishMessages contains all UI messages in English
var EnglishMessages = map[string]string{
	// Config messages
	"config.not_found":      "✗ Configuration file not found",
	"config.hint_run_setup": "Please run: aiassist config",
	"config.title":          "AI Shell Assistant - Configuration Wizard",

	// OpenAI-compatible configuration
	"config.openai_compat.title":         "Configure OpenAI-Compatible LLM Providers",
	"config.openai_compat.model":         "Model",
	"config.openai_compat.provider_name": "Enter provider name (e.g., MyQwen, MyDeepSeek): ",
	"config.openai_compat.name_empty":    "✗ Provider name cannot be empty",
	"config.openai_compat.base_url":      "Enter base URL (e.g., https://api.openai.com/v1): ",
	"config.openai_compat.url_empty":     "✗ Base URL cannot be empty",
	"config.openai_compat.api_key":       "Enter API Key: ",
	"config.openai_compat.key_empty":     "✗ API Key cannot be empty",
	"config.openai_compat.model_name":    "Enter model name (e.g., qwen-plus, gpt-4o, deepseek-chat): ",
	"config.openai_compat.model_empty":   "✗ Model name cannot be empty",
	"config.openai_compat.success":       "✓ Provider '%s' configured successfully",
	"config.openai_compat.add_more":      "Add another model? (yes/no): ",
	"config.openai_compat.no_models":     "✗ No models configured",
	"config.openai_compat.added":         "✓ Provider '%s' added successfully",
	"config.openai_compat.models_list":   "Models: %v",
	"config.openai_compat.order_hint":    "Tip: Model invocation follows the order in the config file. When a model is unavailable, the next one will be tried automatically.",

	// Default model selection
	"config.default_model.title":    "Select Default Model",
	"config.default_model.select":   "Enter the number of your preferred model (default: 1): ",
	"config.default_model.selected": "✓ Default model set to: %s",

	// Proxy configuration
	"config.proxy.title":   "Configure Proxy (Optional)",
	"config.proxy.desc":    "If your network requires a proxy, you can configure it here",
	"config.proxy.example": "Proxy address example: http://127.0.0.1:7890",
	"config.proxy.input":   "Enter proxy address (leave empty for no proxy): ",
	"config.proxy.success": "✓ Proxy configured: %s",
	"config.proxy.empty":   "✓ No proxy configured",
	"config.proxy.error":   "✗ Proxy configuration failed: %v",

	"config.complete":   "✓ Configuration saved to ~/.aiassist/config.yaml",
	"config.save_error": "✗ Failed to save configuration: %v",

	// Interactive mode messages
	"interactive.welcome":              "Welcome to AI Shell Assistant",
	"interactive.input_prompt":         "Please enter your question: ",
	"interactive.goodbye":              "Goodbye!",
	"interactive.commands_found":       "Found suggested commands:",
	"interactive.cancelled":            "Cancelled",
	"interactive.thinking":             "Thinking",
	"interactive.continue_analysis":    "Based on the complete conversation history and the executed command output above, please continue with the next steps of analysis and diagnosis, listing the remaining steps and commands.",
	"interactive.executed_command":     "Executed Command",
	"interactive.execution_output":     "Execution Output",
	"interactive.user_label":           "User",
	"interactive.ai_label":             "AI",
	"interactive.all_commands_skipped": "All commands skipped", "interactive.analysis_complete": "✓ Analysis complete, please continue with questions", "interactive.all_steps_complete": "All analysis steps completed. Continue with more questions? (y/n): ",
	"interactive.pipe_user_question": "User question: ",
	"interactive.pipe_data":          "Pipe output data:",
	"interactive.pipe_source":        "Data source: piped input",

	// Executor messages
	"executor.query_command":        "Query command:",
	"executor.modify_command":       "Modify command (requires confirmation):",
	"executor.unclassified_command": "Unclassified command:",
	"executor.execute_prompt":       "Execute this command?",
	"executor.modify_warning":       "Warning: This command will modify server configuration, are you sure?",
	"executor.executing":            "Executing",
	"executor.execute_success":      "✓ Execution successful",
	"executor.execute_failed":       "✗ Execution failed: %v",
	"executor.no_output":            "(Command executed successfully, but no output)",
	"executor.cancelled":            "Cancelled",
	"executor.read_input_failed":    "Failed to read input: %v",
	"executor.max_depth_reached":    "Warning: Maximum command analysis depth reached. Stopping to prevent infinite recursion.",
	// Output truncation messages
	"output.truncated": "omitted %d lines of output",
	// Error messages
	"error.no_models":      "✗ Error: No models configured",
	"error.hint_no_models": "Please run first: aiassist config",
	"error.unknown_model":  "!!! Warning: Unknown model %s",

	// Version messages
	"version.app_name":   "AI Shell Assistant (aiassist)",
	"version.version":    "Version: %s",
	"version.commit":     "Commit: %s",
	"version.build_date": "Build Date: %s",

	// Model status messages
	"llm.status_title":       "Current Model Status",
	"llm.status_available":   "✓ Available",
	"llm.status_unavailable": "✗ Unavailable",
	"llm.remaining_calls":    "Remaining Calls",
	"llm.priority":           "Priority",

	// UI messages
	"ui.ctrlc_exit_hint":      "Press Ctrl+C again to exit",
	"ui.ctrlc_exit_message":   "User exited via Ctrl+C, goodbye!",
}
