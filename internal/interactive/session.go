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

const (
	// MaxPipeDataBytes limits pipe input to ~1.6MB
	// Based on mainstream LLM context windows (2026):
	// - DeepSeek-V3: 64K tokens
	// - GPT-4/Qwen: 128K tokens
	// - Claude 3.5: 200K tokens
	// - Gemini 1.5: 1M tokens
	// For nginx logs (~3.5 chars/token):
	// 400K chars ≈ 114K tokens ≈ 13,000 lines of nginx access logs
	MaxPipeDataBytes = 400000 * 4

	// MaxPipeDataChars limits pipe data display to ~400k characters
	MaxPipeDataChars = 400000

	// MaxOutputChars limits command output display in interactive mode
	MaxOutputChars = 100000
)

// Common errors
var (
	ErrUserAbort = ui.ErrUserAbort
	ErrUserExit  = ui.ErrUserExit
	ErrUserDone  = ui.ErrUserDone
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

func NewSession(manager *llm.Manager, translator *i18n.I18n) *Session {
	session := &Session{
		llmManager:        manager,
		executor:          executor.NewCommandExecutor(),
		history:           make([]SessionMessage, 0),
		translator:        translator,
		maxRecursionDepth: 10, // Allow deeper analysis for complex troubleshooting scenarios
	}

	sysInfo, err := sysinfo.LoadOrCollect()
	if err != nil {
		color.Yellow("Warning: failed to load system info: %v\n", err)
	} else {
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
	color.Cyan(s.translator.T("interactive.exit_hint") + "\n")
	color.Cyan(ui.Separator() + "\n")

	// Print current model status
	s.llmManager.PrintStatus()

	// If initial question is provided, process it first
	if initialQuestion != "" {
		fmt.Println()
		fmt.Printf("[%s]: %s\n", s.translator.T("interactive.user_label"), initialQuestion)

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
	s.displayResponse(modelUsed, response)

	commands := s.executor.ExtractCommands(response)
	if len(commands) > 0 {
		return s.handleCommands(commands)
	}

	return nil
}

func (s *Session) callLLM(systemPrompt string) (response string, modelUsed string, err error) {
	conversationContext := s.buildConversationContext()
	ctx := context.Background()
	response, modelUsed, err = s.llmManager.CallWithFallbackSystemPrompt(ctx, systemPrompt, conversationContext)
	return
}

func (s *Session) displayResponse(modelUsed, response string) {
	response = strings.TrimSpace(response)
	fmt.Println()
	fmt.Printf("[%s]:\n", modelUsed)
	fmt.Printf("%s\n", response)
	os.Stdout.Sync()
}

func (s *Session) RunWithPipe(initialQuestion string) error {
	limitedReader := io.LimitReader(os.Stdin, MaxPipeDataBytes)
	pipeData, err := io.ReadAll(limitedReader)
	if err != nil {
		return err
	}

	truncatedPipeData := s.truncateOutput(string(pipeData), MaxPipeDataChars)

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

	s.history = append(s.history, SessionMessage{Role: "user", Content: pipeMsg})
	response, modelUsed, err := s.callLLM(prompt.GetPipeAnalysisPrompt())
	if err != nil {
		return err
	}

	s.history = append(s.history, SessionMessage{Role: "assistant", Content: response})
	s.displayResponse(modelUsed, response)

	// In pipe mode, just show the analysis and exit
	// No interactive loop, no command execution
	fmt.Println()
	color.Green(s.translator.T("interactive.analysis_complete") + "\n")
	os.Stdout.Sync()

	return nil
}

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

		fmt.Println(userInput)

		if err := s.processQuestion(userInput); err != nil {
			if errors.Is(err, ErrUserAbort) || errors.Is(err, ErrUserExit) || errors.Is(err, ErrUserDone) {
				return err
			}
			color.Red("Error: %v\n", err)
			continue
		}
	}
}

func (s *Session) handleCommands(commands []executor.Command) error {
	if len(commands) == 0 {
		return nil
	}

	if s.recursionDepth >= s.maxRecursionDepth {
		color.Yellow(s.translator.T("executor.max_depth_reached") + "\n")
		return nil
	}
	s.recursionDepth++
	defer func() { s.recursionDepth-- }()

	executedAny := false

	for _, cmd := range commands {
		fmt.Println()
		s.executor.DisplayCommand(cmd.Text, cmd.Type, s.translator)

		confirmed, err := s.confirmCommandExecution(cmd.Type)
		if err != nil {
			return err
		}
		if !confirmed {
			continue
		}

		execStop := ui.StartSpinner(s.translator.T("executor.executing"))
		output, err := s.executor.ExecuteCommand(cmd.Text)
		if execStop != nil {
			execStop()
		}

		if output == "" {
			output = s.translator.T("executor.no_output")
		}

		fmt.Println()
		fmt.Printf("[%s]:\n", s.translator.T("interactive.execution_output"))
		fmt.Println(output)

		// Build execution result message including error information
		var executionResult string
		if err != nil {
			color.Red(s.translator.T("executor.execute_failed", err))
			// Include error information in the execution result for LLM analysis
			executionResult = fmt.Sprintf("[%s]\n%s\n\n[%s]\n%s\n\n[%s]\n%s",
				s.translator.T("interactive.executed_command"), cmd.Text,
				s.translator.T("interactive.execution_output"), output,
				s.translator.T("interactive.execution_error"), err.Error())
		} else {
			// Show execution success message
			color.Green(s.translator.T("executor.execute_success"))
			executionResult = fmt.Sprintf("[%s]\n%s\n\n[%s]\n%s",
				s.translator.T("interactive.executed_command"), cmd.Text,
				s.translator.T("interactive.execution_output"), output)
		}

		// Record first executed command and analyze its output (whether success or failure)
		if !executedAny {
			executedAny = true
			return s.analyzeCommandOutput(executionResult)
		}
	}

	if !executedAny {
		fmt.Println()
		color.Green(s.translator.T("interactive.analysis_complete") + "\n")
		return nil
	}

	return nil
}

func (s *Session) truncateOutput(output string, maxChars int) string {
	if len(output) <= maxChars {
		return output
	}

	headSize := int(float64(maxChars) * 0.6)
	tailSize := maxChars - headSize - 100

	head := output[:headSize]
	tail := output[len(output)-tailSize:]

	truncatedLines := strings.Count(output[headSize:len(output)-tailSize], "\n")
	truncationMsg := s.translator.T("output.truncated", truncatedLines)
	return fmt.Sprintf("%s\n\n... [%s] ...\n\n%s", head, truncationMsg, tail)
}

func (s *Session) analyzeCommandOutput(executionResult string) error {
	// Truncate the execution result if it's too large
	truncatedResult := s.truncateOutput(executionResult, MaxOutputChars)
	s.history = append(s.history, SessionMessage{Role: "user", Content: truncatedResult})

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

	newCommands := s.executor.ExtractCommands(response)
	if len(newCommands) > 0 {
		return s.handleCommands(newCommands)
	}

	fmt.Println()
	color.Green(s.translator.T("interactive.analysis_complete") + "\n")
	return nil
}
