package executor

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/llaoj/aiassist/internal/i18n"
)

// CommandType represents the classification of a command
type CommandType int

const (
	QueryCommand  CommandType = iota // Query command (read-only, safe)
	ModifyCommand                    // Modify command (write operations, high risk)
)

// Command represents a command with its type
type Command struct {
	Text string
	Type CommandType
}

// CommandExecutor handles command extraction and execution
type CommandExecutor struct{}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{}
}

// readUserInput reads input from terminal (handles both interactive and pipe mode)
func (ce *CommandExecutor) readUserInput() (string, error) {
	// Try to open /dev/tty for reading user input
	// This works even when stdin is a pipe
	tty, err := os.Open("/dev/tty")
	if err != nil {
		// Fallback to stdin if /dev/tty is not available
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(input), nil
	}
	defer tty.Close()

	reader := bufio.NewReader(tty)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// DisplayCommand displays the command and gets user confirmation
func (ce *CommandExecutor) DisplayCommand(cmdText string, cmdType CommandType, translator *i18n.I18n) bool {
	if cmdType == QueryCommand {
		// Query command (display in green, requires confirmation)
		color.Green(translator.T("executor.query_command") + "\n")
		color.Green("%s\n", cmdText)

		// Ask for confirmation (prompt on same line as input)
		fmt.Print("\n")
		color.New(color.FgYellow).Print(translator.T("executor.execute_prompt"))
		input, err := ce.readUserInput()
		if err != nil {
			color.Red("Failed to read input: %v\n", err)
			return false
		}
		input = strings.ToLower(input)

		if input == "exit" {
			color.Yellow("Exiting...\n")
			os.Exit(0)
		}

		if input != "yes" && input != "y" {
			color.Yellow("Cancelled\n")
			return false
		}

		return true
	}

	if cmdType == ModifyCommand {
		// Modify command (display in red, requires confirmation)
		color.Red(translator.T("executor.modify_command") + "\n")
		color.Red("%s\n", cmdText)

		// First confirmation (prompt on same line as input)
		fmt.Print("\n")
		color.New(color.FgYellow).Print(translator.T("executor.execute_prompt"))
		input, err := ce.readUserInput()
		if err != nil {
			color.Red("Failed to read input: %v\n", err)
			return false
		}
		input = strings.ToLower(input)

		if input == "exit" {
			color.Yellow("Exiting...\n")
			os.Exit(0)
		}

		if input != "yes" && input != "y" {
			color.Yellow("Cancelled\n")
			return false
		}

		// Second confirmation for critical operations (only after yes, prompt on same line)
		fmt.Print("\n")
		color.New(color.FgRed).Print(translator.T("executor.modify_warning"))
		input, err = ce.readUserInput()
		if err != nil {
			color.Red("Failed to read input: %v\n", err)
			return false
		}
		input = strings.ToLower(input)

		if input == "exit" {
			color.Yellow("Exiting...\n")
			os.Exit(0)
		}

		if input != "yes" && input != "y" {
			color.Yellow("Cancelled\n")
			return false
		}

		return true
	}

	return false
}

// ExecuteCommand executes the command
func (ce *CommandExecutor) ExecuteCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)

	// Capture output
	output, _ := cmd.CombinedOutput()

	// Print output to console
	if len(output) > 0 {
		fmt.Print(string(output))
	}

	// For commands, non-zero exit status is not always an error
	// (e.g., grep with no matches returns 1, command not found returns 127)
	// We return the output regardless, and only return error for modify commands
	return string(output), nil
}

// ExecuteWithConfirmation displays the command and executes after user confirmation
func (ce *CommandExecutor) ExecuteWithConfirmation(cmd Command, translator *i18n.I18n) (string, error) {
	if !ce.DisplayCommand(cmd.Text, cmd.Type, translator) {
		return "", fmt.Errorf("user cancelled command execution")
	}

	return ce.ExecuteCommand(cmd.Text)
}

// ExtractCommands extracts executable commands from AI response text
// Looks for [cmd:query] and [cmd:modify] markers and returns commands with their types
func (ce *CommandExecutor) ExtractCommands(response string) []Command {
	var commands []Command
	lines := strings.Split(response, "\n")

	for _, line := range lines {
		// Trim leading/trailing spaces for detection, but preserve original for processing
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			continue
		}

		var cmdType CommandType
		var cmdText string

		if strings.HasPrefix(trimmedLine, "[cmd:query]") {
			cmdType = QueryCommand
			cmdText = strings.TrimPrefix(trimmedLine, "[cmd:query]")
		} else if strings.HasPrefix(trimmedLine, "[cmd:modify]") {
			cmdType = ModifyCommand
			cmdText = strings.TrimPrefix(trimmedLine, "[cmd:modify]")
		} else {
			continue
		}

		cmdText = strings.TrimSpace(cmdText)
		// Clean up any markdown formatting
		cmdText = strings.ReplaceAll(cmdText, "**", "")
		cmdText = strings.ReplaceAll(cmdText, "`", "")
		cmdText = strings.TrimSpace(cmdText)

		if cmdText != "" {
			commands = append(commands, Command{
				Text: cmdText,
				Type: cmdType,
			})
		}
	}

	return commands
}

// RunCommand executes command with direct terminal output
func (ce *CommandExecutor) RunCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
