package interactive

import (
	"context"
	"errors"
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
)

// Common errors
var (
	ErrUserAbort = ui.ErrUserAbort
	ErrUserExit  = ui.ErrUserExit
)

// SessionMessage represents a message in the session
type SessionMessage struct {
	Role    string // "system", "user", or "assistant"
	Content string
}

// Session represents an interactive session with user
type Session struct {
	llmManager        *llm.Manager
	executor          *executor.CommandExecutor
	history           []SessionMessage
	translator        *i18n.I18n
	recursionDepth    int // Current recursion depth for command handling
	maxRecursionDepth int // Maximum allowed recursion depth
}

// NewSession creates a new interactive session
func NewSession(manager *llm.Manager, translator *i18n.I18n) *Session {
	// Initialize session
	session := &Session{
		llmManager:        manager,
		executor:          executor.NewCommandExecutor(),
		history:           make([]SessionMessage, 0),
		translator:        translator,
		maxRecursionDepth: 10, // Allow deeper analysis for complex troubleshooting scenarios
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

	return session
}

// Run starts the interactive session
// If initialQuestion is provided, it will be processed and ask if user wants to continue
func (s *Session) Run(initialQuestion string) (err error) {
	// Add panic recovery to ensure terminal is restored
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()

	// Display welcome message
	color.Cyan(ui.Separator() + "\n")
	color.Cyan(s.translator.T("interactive.welcome") + "\n")
	color.Cyan(ui.Separator() + "\n")

	// Print current model status
	s.llmManager.PrintStatus()

	// If initial question is provided, process it first
	if initialQuestion != "" {
		color.Yellow("[%s]: %s\n", s.translator.T("interactive.user_label"), initialQuestion)

		if err := s.processQuestion(initialQuestion); err != nil {
			return err
		}
		// Note: analyzeCommandOutput will ask if user wants to continue after processing
		// No need to ask again here, just fall through to interactive loop
	}

	// Enter unified interactive loop
	return s.runInteractiveLoop()
}

// readUserInput reads input from terminal
func (s *Session) readUserInput(prompt string) (string, error) {
	input, err := ui.PromptInput(prompt, s.translator)
	if err != nil {
		if errors.Is(err, ui.ErrUserAbort) {
			// User pressed Ctrl+C - return error to trigger exit
			return "", err
		}
		return "", err
	}

	return input, nil
}

// confirmCommandExecution asks user to confirm command execution
func (s *Session) confirmCommandExecution(cmdType executor.CommandType) (bool, error) {
	// First confirmation
	confirmed, err := s.askConfirmation(s.translator.T("executor.execute_prompt"))
	if err != nil || !confirmed {
		return false, err
	}

	// Second confirmation for modify commands
	if cmdType == executor.ModifyCommand {
		fmt.Println()
		confirmed, err = s.askConfirmation(s.translator.T("executor.modify_warning"))
		if err != nil || !confirmed {
			return false, err
		}
	}

	return true, nil
}

// askConfirmation prompts user for yes/no confirmation
func (s *Session) askConfirmation(prompt string) (bool, error) {
	confirmed, err := ui.PromptConfirm(prompt, s.translator)
	if err != nil {
		if errors.Is(err, ui.ErrUserAbort) {
			// User pressed Ctrl+C - return error to trigger exit
			return false, err
		}
		return false, err
	}

	return confirmed, nil
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
	// Print empty line before displaying response
	fmt.Println()
	s.displayResponse(modelUsed, response)

	commands := s.executor.ExtractCommands(response)
	if len(commands) > 0 {
		return s.handleCommands(commands)
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
	color.Cyan("[%s]: \n", modelUsed)
	color.Cyan("%s\n\n", response)
	// Flush to ensure output appears immediately in pipe mode
	os.Stdout.Sync()
}

// RunWithPipe runs with pipe input
func (s *Session) RunWithPipe(initialQuestion string) error {
	// Read pipe data with memory limit to prevent exhaustion
	// Based on mainstream LLM context windows (2026):
	// - DeepSeek-V3: 64K tokens
	// - GPT-4/Qwen: 128K tokens
	// - Claude 3.5: 200K tokens
	// - Gemini 1.5: 1M tokens
	//
	// For nginx logs (~3.5 chars/token):
	// 400K chars ≈ 114K tokens ≈ 13,000 lines of nginx access logs
	// Fits comfortably in most models with room for system prompt & history
	const maxPipeDataBytes = 400000 * 4 // ~1.6MB, supports ~13k lines of nginx logs
	limitedReader := io.LimitReader(os.Stdin, maxPipeDataBytes)
	pipeData, err := io.ReadAll(limitedReader)
	if err != nil {
		return err
	}

	// Truncate if too long (~400k chars)
	const maxPipeDataChars = 400000
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
	os.Stdout.Sync() // Ensure all output is flushed

	return nil
}

// runInteractiveLoop provides a unified interactive prompt loop for both Run and RunWithPipe
func (s *Session) runInteractiveLoop() error {
	for {
		// Print empty line before showing input prompt
		fmt.Println()
		prompt := s.translator.T("interactive.input_prompt")
		userInput, err := s.readUserInput(prompt)
		if err != nil {
			if errors.Is(err, ErrUserAbort) {
				// User pressed Ctrl+C
				return ErrUserExit
			}
			if err == io.EOF {
				// EOF (Ctrl+D)
				color.Cyan("\n" + s.translator.T("interactive.goodbye") + "\n")
				return ErrUserExit
			}
			color.Red("Error: %v\n", err)
			continue
		}

		userInput = strings.TrimSpace(userInput)
		if userInput == "" {
			continue
		}

		// Print user input after it's entered (bubbletea clears the input prompt)
		fmt.Println(userInput)

		// Process the question
		if err := s.processQuestion(userInput); err != nil {
			if errors.Is(err, ErrUserAbort) || errors.Is(err, ErrUserExit) {
				return err
			}
			color.Red("Error: %v\n", err)
			continue
		}
	}
}

// handleCommands processes extracted commands
func (s *Session) handleCommands(commands []executor.Command) error {
	if len(commands) == 0 {
		return nil
	}

	// Check recursion depth to prevent stack overflow
	if s.recursionDepth >= s.maxRecursionDepth {
		color.Yellow(s.translator.T("executor.max_depth_reached") + "\n")
		return nil
	}
	s.recursionDepth++
	defer func() { s.recursionDepth-- }()

	// Track if we executed at least one command
	executedAny := false
	var firstExecutedCmd string
	var firstExecutedOutput string

	// Process each command one by one
	for _, cmd := range commands {
		// Display command
		s.executor.DisplayCommand(cmd.Text, cmd.Type, s.translator)

		// Get user confirmation
		confirmed, err := s.confirmCommandExecution(cmd.Type)
		if err != nil {
			return err
		}
		if !confirmed {
			continue
		}

		// Show executing spinner using shared utility
		execStop := ui.StartSpinner(s.translator.T("executor.executing"))
		output, err := s.executor.ExecuteCommand(cmd.Text)
		if execStop != nil {
			execStop()
		}

		// Handle empty output - tell AI explicitly
		if output == "" {
			output = s.translator.T("executor.no_output")
		}

		// Print command output after spinner stops
		// Print empty line before output
		fmt.Println()
		fmt.Printf("[%s]:\n", s.translator.T("interactive.execution_output"))
		fmt.Print(output)

		if err != nil {
			color.Red(s.translator.T("executor.execute_failed", err) + "\n")
			continue
		}

		// Show execution success message
		color.Green(s.translator.T("executor.execute_success") + "\n")
		fmt.Println()

		// Record first executed command
		if !executedAny {
			executedAny = true
			firstExecutedCmd = cmd.Text
			firstExecutedOutput = output
			// After executing the first command, analyze output and continue
			return s.analyzeCommandOutput(firstExecutedCmd, firstExecutedOutput)
		}
	}

	// If no command was executed (all skipped), show message and return to let caller continue
	if !executedAny {
		// color.Yellow(s.translator.T("interactive.all_commands_skipped") + "\n")
		color.Green(s.translator.T("interactive.analysis_complete") + "\n")
		return nil
	}

	return nil
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
func (s *Session) analyzeCommandOutput(cmd, output string) error {
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
		return err
	}

	s.history = append(s.history, SessionMessage{Role: "assistant", Content: response})
	s.displayResponse(modelUsed, response)

	// Handle new commands or ask to continue
	newCommands := s.executor.ExtractCommands(response)
	if len(newCommands) > 0 {
		return s.handleCommands(newCommands)
	}

	// Ask if user wants to continue
	confirmed, err := s.askConfirmation(s.translator.T("interactive.all_steps_complete"))
	if err != nil {
		return err
	}
	if confirmed {
		return nil // Continue to main loop
	}
	// User chose 'n', exit program
	color.Cyan(s.translator.T("interactive.goodbye") + "\n")
	return ErrUserExit
}
