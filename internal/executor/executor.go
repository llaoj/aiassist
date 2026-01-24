package executor

import (
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

// DisplayCommand displays the command and gets user confirmation
func (ce *CommandExecutor) DisplayCommand(cmdText string, cmdType CommandType, translator *i18n.I18n) bool {
	if cmdType == QueryCommand {
		// Query command (display in green, requires confirmation)
		color.Green(translator.T("executor.query_command") + "\n")
		color.Green("%s\n", cmdText)

		// Ask for confirmation
		var input string
		color.Yellow("\n" + translator.T("executor.execute_prompt"))
		fmt.Scanln(&input)
		input = strings.ToLower(strings.TrimSpace(input))

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

		// First confirmation
		var input string
		color.Yellow("\n" + translator.T("executor.execute_prompt"))
		fmt.Scanln(&input)
		input = strings.ToLower(strings.TrimSpace(input))

		if input == "exit" {
			color.Yellow("Exiting...\n")
			os.Exit(0)
		}

		if input != "yes" && input != "y" {
			color.Yellow("Cancelled\n")
			return false
		}

		// Second confirmation for critical operations (only after yes)
		color.Red("\n" + translator.T("executor.modify_warning"))
		fmt.Scanln(&input)
		input = strings.ToLower(strings.TrimSpace(input))

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
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		var cmdType CommandType
		var cmdText string

		if strings.HasPrefix(line, "[cmd:query]") {
			cmdType = QueryCommand
			cmdText = strings.TrimPrefix(line, "[cmd:query]")
		} else if strings.HasPrefix(line, "[cmd:modify]") {
			cmdType = ModifyCommand
			cmdText = strings.TrimPrefix(line, "[cmd:modify]")
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
