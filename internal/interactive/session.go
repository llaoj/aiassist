package interactive

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/llaoj/aiassist/internal/executor"
	"github.com/llaoj/aiassist/internal/i18n"
	"github.com/llaoj/aiassist/internal/llm"
	"github.com/llaoj/aiassist/internal/prompt"
	"github.com/llaoj/aiassist/internal/sysinfo"
	"github.com/llaoj/aiassist/internal/ui"
	"github.com/peterh/liner"
)

// SessionMessage represents a message in the session
type SessionMessage struct {
	Role    string // "system", "user", or "assistant"
	Content string
}

// Session represents an interactive session with user
type Session struct {
	llmManager *llm.Manager
	executor   *executor.CommandExecutor
	history    []SessionMessage
	translator *i18n.I18n
	line       *liner.State
}

// NewSession creates a new interactive session
func NewSession(manager *llm.Manager, translator *i18n.I18n) *Session {
	// Initialize session
	session := &Session{
		llmManager: manager,
		executor:   executor.NewCommandExecutor(),
		history:    make([]SessionMessage, 0),
		translator: translator,
	}

	// Load or collect system info and add to history
	sysInfo, err := sysinfo.LoadOrCollect()
	if err != nil {
		// If failed, just log and continue without system info
		color.Yellow("Warning: failed to load system info: %v\n", err)
	} else {
		// Add system info as the first message in history
		session.history = append(session.history, SessionMessage{
			Role:    "system",
			Content: sysInfo.FormatAsContext(),
		})
	}

	// Initialize liner for interactive input (only for normal mode)
	stdinStat, _ := os.Stdin.Stat()
	isPipe := (stdinStat.Mode() & os.ModeCharDevice) == 0

	if !isPipe {
		// Create liner only for normal interactive mode
		session.line = liner.NewLiner()
		session.line.SetCtrlCAborts(true)
	}

	return session
}

// Run starts the interactive session
// If initialQuestion is provided, it will be processed and ask if user wants to continue
func (s *Session) Run(initialQuestion string) error {
	// Ensure liner is properly closed on exit
	defer s.line.Close()

	// Display welcome message
	color.Cyan(ui.Separator() + "\n")
	color.Cyan(s.translator.T("interactive.welcome") + "\n")
	color.Cyan(s.translator.T("interactive.help_hint") + "\n")
	color.Cyan(ui.Separator() + "\n")

	// Print current model status
	s.llmManager.PrintStatus()

	// If initial question is provided, process it first
	if initialQuestion != "" {
		if err := s.processQuestion(initialQuestion); err != nil {
			return err
		}
		// Note: analyzeCommandOutput will ask if user wants to continue after processing
		// No need to ask again here, just fall through to interactive loop
	}

	// Enter unified interactive loop
	return s.runInteractiveLoop()
}

// readUserInput reads input from terminal using liner
func (s *Session) readUserInput(prompt string) (string, error) {
	if s.line == nil {
		return "", fmt.Errorf("liner not initialized")
	}

	// Flush any pending output before reading
	os.Stdout.Sync()

	// Read line with liner (supports Chinese, editing, cursor movement)
	line, err := s.line.Prompt(prompt)
	if err != nil {
		// Handle Ctrl+C - exit gracefully
		if err == liner.ErrPromptAborted {
			color.Cyan("\n" + s.translator.T("interactive.goodbye") + "\n")
			os.Exit(0)
		}
		return "", err
	}

	return strings.TrimSpace(line), nil
}

// confirmCommandExecution asks user to confirm command execution
func (s *Session) confirmCommandExecution(cmdType executor.CommandType) bool {
	// First confirmation
	if !s.askConfirmation(s.translator.T("executor.execute_prompt")) {
		return false
	}

	// Second confirmation for modify commands
	if cmdType == executor.ModifyCommand {
		fmt.Println()
		if !s.askConfirmation(s.translator.T("executor.modify_warning")) {
			return false
		}
	}

	return true
}

// askConfirmation prompts user for yes/no confirmation
// Only accepts: y, n, exit (+ Enter), or Ctrl+C
func (s *Session) askConfirmation(prompt string) bool {
	for {
		// Print colored prompt
		color.New(color.FgYellow).Fprint(os.Stdout, prompt)
		os.Stdout.Sync()

		// Use liner with empty prompt to read input
		line, err := s.line.Prompt("")
		if err != nil {
			if err == liner.ErrPromptAborted {
				// Ctrl+C pressed
				color.Cyan("\n" + s.translator.T("interactive.goodbye") + "\n")
				os.Exit(0)
			}
			if err == io.EOF {
				color.Cyan("\n" + s.translator.T("interactive.goodbye") + "\n")
				os.Exit(0)
			}
			color.Red(s.translator.T("executor.read_input_failed", err) + "\n")
			return false
		}

		input := strings.TrimSpace(line)
		input = strings.ToLower(input)

		// Handle valid inputs
		if input == "exit" {
			color.Yellow(s.translator.T("executor.exiting") + "\n")
			os.Exit(0)
		}

		if input == "y" || input == "yes" {
			return true
		}

		if input == "n" || input == "no" {
			color.Yellow(s.translator.T("executor.cancelled") + "\n")
			return false
		}

		// Invalid input - show error and re-prompt
		color.Red("Invalid input. Please enter: y, n, or exit\n")
	}
}

// buildConversationContext builds conversation context from history
func (s *Session) buildConversationContext() string {
	var context string

	// Add all conversation history in chronological order
	for _, msg := range s.history {
		switch msg.Role {
		case "system":
			// System info doesn't need label
			context += msg.Content + "\n\n"
		case "user":
			context += fmt.Sprintf("[%s]: %s\n\n", s.translator.T("interactive.user_label"), msg.Content)
		case "assistant":
			context += fmt.Sprintf("[%s]: %s\n\n", s.translator.T("interactive.ai_label"), msg.Content)
		}
	}

	return context
}

// processQuestion handles a single question and its response
func (s *Session) processQuestion(userInput string) error {
	s.history = append(s.history, SessionMessage{Role: "user", Content: userInput})

	response, modelUsed, err := s.callLLM(prompt.GetInteractivePrompt())
	if err != nil {
		return err
	}

	s.history = append(s.history, SessionMessage{Role: "assistant", Content: response})
	s.displayResponse(modelUsed, response)

	commands := s.executor.ExtractCommands(response)
	if len(commands) > 0 {
		s.handleCommands(commands)
	}

	return nil
}

// callLLM calls the LLM with current conversation context
func (s *Session) callLLM(systemPrompt string) (response string, modelUsed string, err error) {
	conversationContext := s.buildConversationContext()
	ctx := context.Background()
	stopSpinner := ui.StartSpinner(s.translator.T("interactive.thinking"))
	response, modelUsed, err = s.llmManager.CallWithFallbackSystemPrompt(ctx, systemPrompt, conversationContext)
	if stopSpinner != nil {
		stopSpinner()
	}
	return
}

// displayResponse displays AI response with model name
func (s *Session) displayResponse(modelUsed, response string) {
	color.Cyan("[%s]\n", modelUsed)
	color.Cyan("%s\n", response)
}

// RunWithPipe runs with pipe input
func (s *Session) RunWithPipe(initialQuestion string) error {
	// Read pipe data directly from stdin
	pipeData, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	// Truncate if too long (~150k chars)
	const maxPipeDataChars = 150000
	truncatedPipeData := s.truncateOutput(string(pipeData), maxPipeDataChars)

	// Build pipe message
	var pipeMsg string
	if initialQuestion != "" {
		pipeMsg = fmt.Sprintf("%s\n%s%s\n\n%s\n%s",
			s.translator.T("interactive.pipe_source"),
			s.translator.T("interactive.pipe_user_question"), initialQuestion,
			s.translator.T("interactive.pipe_data"), truncatedPipeData)
	} else {
		pipeMsg = fmt.Sprintf("%s\n\n%s\n%s",
			s.translator.T("interactive.pipe_source"),
			s.translator.T("interactive.pipe_data"), truncatedPipeData)
	}

	// Process as a question with pipe analysis prompt
	s.history = append(s.history, SessionMessage{Role: "user", Content: pipeMsg})
	response, modelUsed, err := s.callLLM(prompt.GetPipeAnalysisPrompt())
	if err != nil {
		return err
	}

	s.history = append(s.history, SessionMessage{Role: "assistant", Content: response})
	s.displayResponse(modelUsed, response)

	// In pipe mode, just show the analysis and exit
	// No interactive loop, no command execution
	color.Green(s.translator.T("interactive.analysis_complete") + "\n")

	return nil
}

// runInteractiveLoop provides a unified interactive prompt loop for both Run and RunWithPipe
func (s *Session) runInteractiveLoop() error {
	for {
		prompt := s.translator.T("interactive.input_prompt")
		userInput, err := s.readUserInput(prompt)
		if err != nil {
			if err == io.EOF {
				// EOF (Ctrl+D)
				color.Cyan("\n" + s.translator.T("interactive.goodbye") + "\n")
				return nil
			}
			// Note: Ctrl+C (liner.ErrPromptAborted) is handled in readUserInput
			color.Red("Error: %v\n", err)
			continue
		}

		userInput = strings.TrimSpace(userInput)
		if userInput == "" {
			continue
		}

		// Handle special commands
		switch strings.ToLower(userInput) {
		case "exit":
			color.Cyan(s.translator.T("interactive.goodbye") + "\n")
			return nil
		case "help":
			s.printHelp()
			continue
		case "history":
			s.printHistory()
			continue
		}

		// Process the question
		if err := s.processQuestion(userInput); err != nil {
			color.Red("Error: %v\n", err)
			continue
		}
	}
}

// handleCommands processes extracted commands
func (s *Session) handleCommands(commands []executor.Command) {
	if len(commands) == 0 {
		return
	}

	// Track if we executed at least one command
	executedAny := false
	var firstExecutedCmd string
	var firstExecutedOutput string

	// Process each command one by one
	for _, cmd := range commands {
		// Display command
		s.executor.DisplayCommand(cmd.Text, cmd.Type, s.translator)

		// Get user confirmation
		if !s.confirmCommandExecution(cmd.Type) {
			continue
		}

		// Show executing spinner using shared utility
		execStop := ui.StartSpinner(s.translator.T("executor.executing"))
		output, err := s.executor.ExecuteCommand(cmd.Text)
		if execStop != nil {
			execStop()
		}
		if err != nil {
			color.Red(s.translator.T("executor.execute_failed", err) + "\n")
			continue
		}

		// Show execution success message
		color.Green(s.translator.T("executor.execute_success") + "\n")

		// Handle empty output - tell AI explicitly
		if output == "" {
			output = s.translator.T("executor.no_output")
		}

		// Record first executed command
		if !executedAny {
			executedAny = true
			firstExecutedCmd = cmd.Text
			firstExecutedOutput = output
			// After executing the first command, analyze output and continue
			s.analyzeCommandOutput(firstExecutedCmd, firstExecutedOutput)
			// After analysis, we return to allow the next step's commands to be handled
			return
		}
	}

	// If no command was executed (all skipped), show message and return to let caller continue
	if !executedAny {
		color.Yellow(s.translator.T("interactive.all_commands_skipped") + "\n")
		color.Green(s.translator.T("interactive.analysis_complete") + "\n")
		return
	}
}

// truncateOutput intelligently truncates command output to fit within model limits
// Keeps the beginning and end of output for better context
func (s *Session) truncateOutput(output string, maxChars int) string {
	if len(output) <= maxChars {
		return output
	}

	// Keep first 60% and last 40% of the allowed content
	headSize := int(float64(maxChars) * 0.6)
	tailSize := maxChars - headSize - 100 // Reserve space for truncation message

	head := output[:headSize]
	tail := output[len(output)-tailSize:]

	truncatedLines := strings.Count(output[headSize:len(output)-tailSize], "\n")
	truncationMsg := s.translator.T("output.truncated", truncatedLines)
	return fmt.Sprintf("%s\n\n... [%s] ...\n\n%s", head, truncationMsg, tail)
}

// analyzeCommandOutput analyzes command output and continues investigation
func (s *Session) analyzeCommandOutput(cmd, output string) {
	// Truncate output if too long (~100k chars)
	const maxOutputChars = 100000
	truncatedOutput := s.truncateOutput(output, maxOutputChars)

	// Add execution result to history
	executionMsg := fmt.Sprintf("[%s]\n%s\n\n[%s]\n%s",
		s.translator.T("interactive.executed_command"), cmd,
		s.translator.T("interactive.execution_output"), truncatedOutput)
	s.history = append(s.history, SessionMessage{Role: "user", Content: executionMsg})

	// Call LLM for analysis
	fullContext := s.buildConversationContext() + s.translator.T("interactive.continue_analysis")
	ctx := context.Background()
	stopSpinner := ui.StartSpinner(s.translator.T("interactive.thinking"))
	response, modelUsed, err := s.llmManager.CallWithFallbackSystemPrompt(ctx, prompt.GetContinueAnalysisPrompt(), fullContext)
	if stopSpinner != nil {
		stopSpinner()
	}

	if err != nil {
		color.Red("Error: %v\n", err)
		return
	}

	s.history = append(s.history, SessionMessage{Role: "assistant", Content: response})
	s.displayResponse(modelUsed, response)

	// Handle new commands or ask to continue
	newCommands := s.executor.ExtractCommands(response)
	if len(newCommands) > 0 {
		s.handleCommands(newCommands)
	} else {
		// Ask if user wants to continue
		input, err := s.readUserInput(s.translator.T("interactive.all_steps_complete"))
		if err != nil {
			return // Return to main loop on error or Ctrl+C
		}

		if strings.ToLower(strings.TrimSpace(input)) == "y" || strings.ToLower(strings.TrimSpace(input)) == "yes" {
			return // Continue to main loop
		}

		color.Cyan(s.translator.T("interactive.goodbye") + "\n")
		os.Exit(0)
	}
}

// printHelp displays help information
func (s *Session) printHelp() {
	fmt.Println("\n" + s.translator.T("interactive.help_title"))
	fmt.Println(s.translator.T("interactive.help_command"))
	fmt.Println(s.translator.T("interactive.help_history"))
	fmt.Println(s.translator.T("interactive.help_exit"))
	fmt.Println("\n" + s.translator.T("interactive.help_examples"))
	fmt.Println(s.translator.T("interactive.help_ex1"))
	fmt.Println(s.translator.T("interactive.help_ex2"))
	fmt.Println(s.translator.T("interactive.help_ex3"))
	fmt.Println()
}

// printHistory displays session history
func (s *Session) printHistory() {
	if len(s.history) == 0 {
		color.Yellow(s.translator.T("interactive.history_empty") + "\n")
		return
	}

	color.Cyan("\n" + s.translator.T("interactive.history_title") + "\n")
	color.Cyan(ui.Separator() + "\n")

	// Display all messages in history
	for _, msg := range s.history {
		switch msg.Role {
		case "system":
			color.Green("[System Info]\n")
		case "user":
			color.Yellow("[%s]: %s\n", s.translator.T("interactive.user_label"), msg.Content)
		case "assistant":
			color.Cyan("[%s]: %s\n", s.translator.T("interactive.ai_label"), msg.Content)
		}
		color.Cyan(ui.Separator() + "\n")
	}
}
