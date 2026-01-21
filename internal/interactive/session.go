package interactive

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/llaoj/aiassist/internal/config"
	"github.com/llaoj/aiassist/internal/executor"
	"github.com/llaoj/aiassist/internal/i18n"
	"github.com/llaoj/aiassist/internal/llm"
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
}

// NewSession creates a new interactive session
func NewSession(manager *llm.Manager, cfg *config.Config, translator *i18n.I18n, stdin io.Reader) *Session {
	if stdin == nil {
		stdin = os.Stdin
	}

	return &Session{
		llmManager: manager,
		executor:   executor.NewCommandExecutor(),
		history:    make([]SessionMessage, 0),
		stdin:      stdin,
		cfg:        cfg,
		translator: translator,
	}
}

// Run starts the interactive session
func (s *Session) Run() error {
	reader := bufio.NewReader(s.stdin)

	// Display welcome message
	color.Cyan("───────────────────────────────────────\n\n")
	color.Cyan(s.translator.T("interactive.welcome") + "\n")
	color.Cyan(s.translator.T("interactive.help_hint") + "\n")
	color.Cyan("───────────────────────────────────────\n\n")

	// Print current model status
	s.llmManager.PrintStatus(s.cfg.GetLanguage())

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
		if userInput == "exit" || userInput == "quit" {
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

		// Call LLM
		ctx := context.Background()
		response, modelUsed, err := s.llmManager.CallWithFallback(ctx, userInput, s.cfg.GetLanguage())
		if err != nil {
			color.Red("❌ Error: %v\n", err)
			continue
		}

		// Add assistant message to history
		s.history = append(s.history, SessionMessage{Role: "assistant", Content: response})

		// Display response
		color.Cyan("\n[%s]: %s\n\n", modelUsed, response)

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

	// Merge user input and pipe data
	var fullPrompt string
	if s.cfg.GetLanguage() == config.LanguageChinese {
		fullPrompt = fmt.Sprintf("用户问题: %s\n\n管道输出数据:\n%s", input, string(pipeData))
	} else {
		fullPrompt = fmt.Sprintf("User question: %s\n\nPipe output data:\n%s", input, string(pipeData))
	}

	// Call LLM
	ctx := context.Background()
	response, modelUsed, err := s.llmManager.CallWithFallback(ctx, fullPrompt, s.cfg.GetLanguage())
	if err != nil {
		return err
	}

	// Display response
	fmt.Printf("[%s]: %s\n", modelUsed, response)

	// Extract and process commands from response
	commands := s.executor.ExtractCommands(response)
	if len(commands) > 0 {
		s.handleCommands(commands)
	}

	return nil
}

// handleCommands processes extracted commands
func (s *Session) handleCommands(commands []string) {
	if len(commands) == 0 {
		return
	}

	// Display found commands
	color.Yellow("\n" + s.translator.T("interactive.commands_found") + "\n")
	for i, cmd := range commands {
		fmt.Printf("%d. %s\n", i+1, cmd)
	}

	reader := bufio.NewReader(os.Stdin)
	color.Yellow("\n" + s.translator.T("interactive.execute_prompt"))
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input != "yes" && input != "y" {
		fmt.Println(s.translator.T("interactive.cancelled"))
		return
	}

	// Execute commands
	for _, cmd := range commands {
		if !s.executor.DisplayCommand(cmd, false, s.translator) {
			continue
		}

		output, err := s.executor.ExecuteCommand(cmd)
		if err != nil {
			color.Red(s.translator.T("executor.execute_failed", err) + "\n")
		} else {
			color.Green(s.translator.T("executor.execute_success") + "\n")
			fmt.Println(output)
		}

		// If there is command output, offer to continue analysis
		if output != "" {
			reader := bufio.NewReader(os.Stdin)
			color.Yellow("\n" + s.translator.T("interactive.continue_prompt"))
			continueInput, _ := reader.ReadString('\n')
			continueInput = strings.TrimSpace(continueInput)

			if continueInput == "yes" || continueInput == "y" {
				color.Yellow(s.translator.T("interactive.followup_prompt"))
				nextQuestion, _ := reader.ReadString('\n')
				nextQuestion = strings.TrimSpace(nextQuestion)

				if nextQuestion != "" {
					// Construct new prompt with command output context
					var contextPrompt string
					if s.cfg.GetLanguage() == config.LanguageChinese {
						contextPrompt = fmt.Sprintf("前一条命令的输出: %s\n\n用户问题: %s", output, nextQuestion)
					} else {
						contextPrompt = fmt.Sprintf("Previous command output: %s\n\nUser question: %s", output, nextQuestion)
					}

					ctx := context.Background()
					response, modelUsed, err := s.llmManager.CallWithFallback(ctx, contextPrompt, s.cfg.GetLanguage())
					if err != nil {
						color.Red("❌ Error: %v\n", err)
					} else {
						color.Cyan("\n[%s]: %s\n\n", modelUsed, response)
					}
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
