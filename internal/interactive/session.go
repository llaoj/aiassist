package interactive

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/llaoj/aiassist/internal/executor"
	"github.com/llaoj/aiassist/internal/i18n"
	"github.com/llaoj/aiassist/internal/llm"
	"github.com/llaoj/aiassist/internal/prompt"
	"github.com/llaoj/aiassist/internal/sysinfo"
	"github.com/llaoj/aiassist/internal/ui"
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
	rl         *readline.Instance
	tty        io.Writer // /dev/tty for manual prompt output
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

	// Initialize readline for interactive input
	// Check if stdin is a terminal or a pipe
	stdinStat, _ := os.Stdin.Stat()
	isPipe := (stdinStat.Mode() & os.ModeCharDevice) == 0

	var rlConfig *readline.Config
	if isPipe {
		// stdin is a pipe, use /dev/tty for both input and output
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err != nil {
			color.Yellow("Warning: failed to open /dev/tty: %v\n", err)
			// Fallback to default
			rlConfig = &readline.Config{
				Prompt:          "",
				InterruptPrompt: "^C",
				EOFPrompt:       "exit",
			}
		} else {
			session.tty = tty
			rlConfig = &readline.Config{
				Prompt:          "",
				Stdin:           readline.NewCancelableStdin(tty),
				Stdout:          tty,
				Stderr:          os.Stderr,
				InterruptPrompt: "^C",
				EOFPrompt:       "exit",
			}
		}
	} else {
		// stdin is a terminal, use default stdin/stdout
		session.tty = os.Stdout
		rlConfig = &readline.Config{
			Prompt:          "",
			InterruptPrompt: "^C",
			EOFPrompt:       "exit",
		}
	}

	rl, err := readline.NewEx(rlConfig)
	if err != nil {
		color.Yellow("Warning: failed to initialize readline: %v\n", err)
	}
	session.rl = rl

	return session
}

// Run starts the interactive session
// If initialQuestion is provided, it will be processed and ask if user wants to continue
func (s *Session) Run(initialQuestion string) error {
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

// readUserInput reads input from terminal using readline
func (s *Session) readUserInput(prompt string) (string, error) {
	if s.rl == nil {
		return "", fmt.Errorf("readline not initialized")
	}

	// In pipe mode, manually write prompt to /dev/tty
	// In interactive mode, use readline's SetPrompt
	if s.tty != os.Stdout {
		// Pipe mode: s.tty is /dev/tty file, manually print prompt
		fmt.Fprint(s.tty, prompt)
	} else {
		// Interactive mode: use readline's SetPrompt
		s.rl.SetPrompt(prompt)
	}

	line, err := s.rl.Readline()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(line), nil
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
	// Add user message to history
	s.history = append(s.history, SessionMessage{Role: "user", Content: userInput})

	// Build conversation context from complete history
	conversationContext := s.buildConversationContext()

	// Call LLM with loading animation
	ctx := context.Background()
	systemPrompt := prompt.GetInteractivePrompt()
	stopSpinner := ui.StartSpinner(s.translator.T("interactive.thinking"))
	response, modelUsed, err := s.llmManager.CallWithFallbackSystemPrompt(ctx, systemPrompt, conversationContext)
	if stopSpinner != nil {
		stopSpinner()
	}

	if err != nil {
		return err
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

	return nil
}

// RunWithPipe runs with pipe input
// initialQuestion is the user's question to be included as context
func (s *Session) RunWithPipe(initialQuestion string) error {
	// Read pipe data from stdin
	pipeData, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	// Truncate pipe data if too long
	// Pipe mode is typically used for log analysis, so allow more data
	// Keep within ~150k chars to leave room for system context and question
	const maxPipeDataChars = 150000
	truncatedPipeData := s.truncateOutput(string(pipeData), maxPipeDataChars)

	// Build pipe input message with labels
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

	// Add pipe data to history
	s.history = append(s.history, SessionMessage{Role: "user", Content: pipeMsg})

	// Build context from history (includes system info)
	conversationContext := s.buildConversationContext()

	// Call LLM with pipe analysis prompt for standalone command analysis
	ctx := context.Background()
	systemPrompt := prompt.GetPipeAnalysisPrompt()
	stopSpinner := ui.StartSpinner(s.translator.T("interactive.thinking"))
	response, modelUsed, err := s.llmManager.CallWithFallbackSystemPrompt(ctx, systemPrompt, conversationContext)
	if stopSpinner != nil {
		stopSpinner()
	}

	if err != nil {
		return err
	}

	// Add assistant response to history
	s.history = append(s.history, SessionMessage{Role: "assistant", Content: response})

	// Display response with model name on separate line
	color.Cyan("\n[%s]\n", modelUsed)
	color.Cyan("%s\n\n", response)

	// Extract and process commands from response
	commands := s.executor.ExtractCommands(response)
	if len(commands) > 0 {
		s.handleCommands(commands)
	} else {
		// No commands found, show completion message
		color.Cyan(ui.Separator() + "\n")
		color.Green(s.translator.T("interactive.analysis_complete") + "\n")
		color.Cyan(ui.Separator() + "\n")
	}

	// Enter interactive loop to allow user to ask more questions
	return s.runInteractiveLoop()
}

// runInteractiveLoop provides a unified interactive prompt loop for both Run and RunWithPipe
func (s *Session) runInteractiveLoop() error {
	for {
		prompt := s.translator.T("interactive.input_prompt")
		userInput, err := s.readUserInput(prompt)
		if err != nil {
			if err == readline.ErrInterrupt {
				// Ctrl+C pressed
				color.Cyan("\n" + s.translator.T("interactive.goodbye") + "\n")
				os.Exit(0)
			}
			if err == io.EOF {
				// EOF (Ctrl+D)
				color.Cyan("\n" + s.translator.T("interactive.goodbye") + "\n")
				return nil
			}
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
		// Execute command with confirmation
		if !s.executor.DisplayCommand(cmd.Text, cmd.Type, s.translator) {
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
		color.Cyan(ui.Separator() + "\n")
		color.Yellow(s.translator.T("interactive.all_commands_skipped") + "\n")
		color.Green(s.translator.T("interactive.analysis_complete") + "\n")
		color.Cyan(ui.Separator() + "\n")
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

// analyzeCommandOutput analyzes command output and continues the investigation
func (s *Session) analyzeCommandOutput(cmd, output string) {
	// Truncate output if too long (reserve space for context and history)
	// Approximate: 1 char ≈ 0.3 tokens for Chinese/English mix, model max ≈ 128k tokens
	// Keep output within ~100k chars to leave room for system context and history
	const maxOutputChars = 100000
	truncatedOutput := s.truncateOutput(output, maxOutputChars)

	// Add command execution to history
	executionMsg := fmt.Sprintf("[%s]\n%s\n\n[%s]\n%s",
		s.translator.T("interactive.executed_command"), cmd,
		s.translator.T("interactive.execution_output"), truncatedOutput)
	s.history = append(s.history, SessionMessage{Role: "user", Content: executionMsg})

	// Build complete conversation context
	fullContext := s.buildConversationContext()

	// Add instruction for next steps
	instructionPrompt := fullContext + s.translator.T("interactive.continue_analysis")

	// Call LLM with continue analysis prompt to handle the output and proceed with next steps
	ctx := context.Background()
	systemPrompt := prompt.GetContinueAnalysisPrompt()
	stopSpinner := ui.StartSpinner(s.translator.T("interactive.thinking"))
	response, modelUsed, err := s.llmManager.CallWithFallbackSystemPrompt(ctx, systemPrompt, instructionPrompt)
	if stopSpinner != nil {
		stopSpinner()
	}

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
		prompt := s.translator.T("interactive.all_steps_complete")
		input, err := s.readUserInput(prompt)
		if err != nil {
			if err == readline.ErrInterrupt || err == io.EOF {
				// User interrupted, return to main loop
				return
			}
			color.Red("Failed to read input: %v\n", err)
			return
		}

		input = strings.TrimSpace(strings.ToLower(input))
		if input == "y" || input == "yes" {
			// User wants to continue, return to main loop for next question
			return
		}

		// User chose not to continue, exit program
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
