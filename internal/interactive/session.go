package interactive

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/llaoj/aiassist/internal/config"
	"github.com/llaoj/aiassist/internal/executor"
	"github.com/llaoj/aiassist/internal/i18n"
	"github.com/llaoj/aiassist/internal/llm"
	"github.com/llaoj/aiassist/internal/prompt"
	"github.com/llaoj/aiassist/internal/sysinfo"
	"github.com/llaoj/aiassist/internal/ui"
)

// SessionMessage represents a message in the session
type SessionMessage struct {
	Role    string // "user" or "assistant"
	Content string
}

// Session represents an interactive session with user
type Session struct {
	llmManager *llm.Manager
	executor   *executor.CommandExecutor
	history    []SessionMessage
	stdin      io.Reader
	cfg        *config.Config
	translator *i18n.I18n
	stopChan   chan bool
	sysInfo    *sysinfo.SystemInfo
}

// NewSession creates a new interactive session
func NewSession(manager *llm.Manager, cfg *config.Config, translator *i18n.I18n, stdin io.Reader) *Session {
	if stdin == nil {
		stdin = os.Stdin
	}

	// Load or collect system info
	sysInfo, err := sysinfo.LoadOrCollect()
	if err != nil {
		// If failed, just log and continue without system info
		color.Yellow("Warning: failed to load system info: %v\n", err)
	}

	return &Session{
		llmManager: manager,
		executor:   executor.NewCommandExecutor(),
		history:    make([]SessionMessage, 0),
		stdin:      stdin,
		cfg:        cfg,
		translator: translator,
		stopChan:   make(chan bool, 1),
		sysInfo:    sysInfo,
	}
}

// Run starts the interactive session
func (s *Session) Run() error {
	reader := bufio.NewReader(s.stdin)

	// Display welcome message

	color.Cyan(ui.Separator() + "\n")
	color.Cyan(s.translator.T("interactive.welcome") + "\n")
	color.Cyan(s.translator.T("interactive.help_hint") + "\n")
	color.Cyan(ui.Separator() + "\n")

	// Print current model status
	s.llmManager.PrintStatus()

	// Main interaction loop
	for {
		color.Yellow(s.translator.T("interactive.input_prompt"))

		userInput, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}

		userInput = strings.TrimSpace(userInput)

		if userInput == "" {
			continue
		}

		// Handle special commands
		if userInput == "exit" {
			color.Cyan(s.translator.T("interactive.goodbye") + "\n")
			break
		}

		if userInput == "help" {
			s.printHelp()
			continue
		}

		if userInput == "history" {
			s.printHistory()
			continue
		}

		// Add user message to history
		s.history = append(s.history, SessionMessage{Role: "user", Content: userInput})

		// Prepare user input with system context
		userInputWithContext := userInput
		if s.sysInfo != nil {
			userInputWithContext = fmt.Sprintf("%s\n\n%s", s.sysInfo.FormatAsContext(), userInput)
		}

		// Call LLM with loading animation
		ctx := context.Background()
		systemPrompt := prompt.GetInteractivePrompt()
		s.startLoading()
		response, modelUsed, err := s.llmManager.CallWithFallbackSystemPrompt(ctx, systemPrompt, userInputWithContext)
		s.stopLoading()

		if err != nil {
			color.Red("Error: %v\n", err)
			continue
		}

		// Add assistant message to history
		s.history = append(s.history, SessionMessage{Role: "assistant", Content: response})

		// Display response with model name on separate line
		color.Cyan("\n[%s]\n", modelUsed)
		color.Cyan("%s\n\n", response)

		// Extract and execute commands from response
		commands := s.executor.ExtractCommands(response)

		if len(commands) > 0 {
			s.handleCommands(commands)
		}
	}

	return nil
}

// RunWithPipe runs with pipe input
func (s *Session) RunWithPipe(input string) error {
	// Read pipe data from stdin
	pipeData, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	// Merge user input and pipe data with system context
	var fullPrompt string
	sysContext := ""
	if s.sysInfo != nil {
		sysContext = s.sysInfo.FormatAsContext() + "\n"
	}

	if s.cfg.GetLanguage() == config.LanguageChinese {
		fullPrompt = fmt.Sprintf("%s用户问题: %s\n\n管道输出数据:\n%s", sysContext, input, string(pipeData))
	} else {
		fullPrompt = fmt.Sprintf("%sUser question: %s\n\nPipe output data:\n%s", sysContext, input, string(pipeData))
	}

	// Call LLM with pipe analysis prompt for standalone command analysis
	ctx := context.Background()
	systemPrompt := prompt.GetPipeAnalysisPrompt()
	s.startLoading()
	response, modelUsed, err := s.llmManager.CallWithFallbackSystemPrompt(ctx, systemPrompt, fullPrompt)
	s.stopLoading()

	if err != nil {
		return err
	}

	// Display response with model name on separate line
	color.Cyan("\n[%s]\n", modelUsed)
	color.Cyan("%s\n", response)

	// Extract and process commands from response
	commands := s.executor.ExtractCommands(response)
	if len(commands) > 0 {
		s.handleCommands(commands)
	}

	return nil
}

// handleCommands processes extracted commands
func (s *Session) handleCommands(commands []executor.Command) {
	if len(commands) == 0 {
		return
	}

	// Process each command one by one
	for i, cmd := range commands {
		// Execute command with confirmation
		if !s.executor.DisplayCommand(cmd.Text, cmd.Type, s.translator) {
			color.Yellow("Skipped\n")
			continue
		}

		output, err := s.executor.ExecuteCommand(cmd.Text)
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

		// After executing the first command, analyze output and continue
		// Always trigger analysis for the first command, even if output is empty
		if i == 0 {
			s.analyzeCommandOutput(cmd.Text, output)
			// After analysis, we return to allow the next step's commands to be handled
			return
		}
	}
}

// analyzeCommandOutput analyzes command output and continues the investigation
func (s *Session) analyzeCommandOutput(cmd, output string) {
	reader := bufio.NewReader(os.Stdin)

	// Add command execution to history
	executionMsg := fmt.Sprintf("[执行命令]\n%s\n\n[执行输出]\n%s", cmd, output)
	s.history = append(s.history, SessionMessage{Role: "user", Content: executionMsg})

	// Build complete conversation history for context
	var fullContext string

	// Add system context at the beginning
	if s.sysInfo != nil {
		fullContext = s.sysInfo.FormatAsContext() + "\n"
	}

	for _, msg := range s.history {
		if msg.Role == "user" {
			fullContext += fmt.Sprintf("[用户]: %s\n\n", msg.Content)
		} else {
			fullContext += fmt.Sprintf("[AI]: %s\n\n", msg.Content)
		}
	}

	// Add instruction for next steps
	var instructionPrompt string
	if s.cfg.GetLanguage() == config.LanguageChinese {
		instructionPrompt = fmt.Sprintf("%s\n根据以上完整的对话历史和已执行的命令输出，请继续进行接下来的分析和诊断，列出剩余的步骤和命令。", fullContext)
	} else {
		instructionPrompt = fmt.Sprintf("%s\nBased on the complete conversation history and the executed command output above, please continue with the next steps of analysis and diagnosis, listing the remaining steps and commands.", fullContext)
	}

	// Call LLM with continue analysis prompt to handle the output and proceed with next steps
	ctx := context.Background()
	systemPrompt := prompt.GetContinueAnalysisPrompt()
	s.startLoading()
	response, modelUsed, err := s.llmManager.CallWithFallbackSystemPrompt(ctx, systemPrompt, instructionPrompt)
	s.stopLoading()

	if err != nil {
		color.Red("Error: %v\n", err)
		return
	}

	// Add analysis to history
	s.history = append(s.history, SessionMessage{Role: "assistant", Content: response})

	// Display analysis with model name on separate line
	color.Cyan("\n[%s]\n", modelUsed)
	color.Cyan("%s\n\n", response)

	// Extract new commands from analysis
	newCommands := s.executor.ExtractCommands(response)
	if len(newCommands) > 0 {
		// Recursively handle new commands
		s.handleCommands(newCommands)
	} else {
		// No more commands, ask if user wants to continue
		color.Yellow("所有分析步骤已完成。是否继续提问? (y/n): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "y" || input == "yes" {
			color.Yellow("请输入后续问题: ")
			nextQuestion, _ := reader.ReadString('\n')
			nextQuestion = strings.TrimSpace(nextQuestion)

			if nextQuestion != "" {
				// Treat as new user input
				s.history = append(s.history, SessionMessage{Role: "user", Content: nextQuestion})

				ctx := context.Background()
				systemPrompt := prompt.GetInteractivePrompt()
				s.startLoading()
				response, modelUsed, err := s.llmManager.CallWithFallbackSystemPrompt(ctx, systemPrompt, nextQuestion)
				s.stopLoading()

				if err != nil {
					color.Red("Error: %v\n", err)
					return
				}

				s.history = append(s.history, SessionMessage{Role: "assistant", Content: response})
				color.Cyan("\n[%s]\n", modelUsed)
				color.Cyan("%s\n\n", response)
				// Extract and handle new commands from this response
				newCommands := s.executor.ExtractCommands(response)
				if len(newCommands) > 0 {
					s.handleCommands(newCommands)
				}
			}
		}
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
	color.Cyan("───────────────────────────────────────\n")

	// Display all messages in history
	for _, msg := range s.history {
		if msg.Role == "user" {
			color.Yellow("👤 %s: %s\n", s.translator.T("interactive.history_user"), msg.Content)
		} else {
			color.Cyan("🤖 %s: %s\n", s.translator.T("interactive.history_assistant"), msg.Content)
		}
	}

	color.Cyan("───────────────────────────────────────\n\n")
}

// startLoading displays an animated loading message
func (s *Session) startLoading() {
	// Check if stdout is a terminal (TTY)
	// Only show animation in interactive mode, not in pipes
	stat, err := os.Stdout.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) == 0 {
		// Not a TTY (e.g., piped output), skip animation
		return
	}

	// Create a new channel for this loading session
	done := make(chan bool)
	s.stopChan = done

	go func() {
		message := s.translator.T("interactive.thinking")

		dots := -1
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				// Clear the loading line
				fmt.Fprintf(os.Stdout, "\r%s\r", strings.Repeat(" ", 100))
				return
			case <-ticker.C:
				dots = (dots + 1) % 4
				dotStr := strings.Repeat(".", dots)
				// Clear line and reprint with updated dots
				fmt.Fprintf(os.Stdout, "\r")
				// Use color.New to avoid automatic newline
				greenPrinter := color.New(color.FgGreen)
				greenPrinter.Printf("%s%s", message, dotStr)
				os.Stdout.Sync()
			}
		}
	}()

	// Give goroutine time to start
	time.Sleep(100 * time.Millisecond)
}

// stopLoading stops the loading animation
func (s *Session) stopLoading() {
	select {
	case s.stopChan <- true:
	default:
	}
	// Give a moment for the goroutine to finish
	time.Sleep(200 * time.Millisecond)
}
