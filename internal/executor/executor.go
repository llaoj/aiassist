package executor

import (
	"fmt"
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

func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{}
}

func (ce *CommandExecutor) GetCommandTypeInfo(cmdType CommandType, translator *i18n.I18n) (string, *color.Color) {
	if cmdType == QueryCommand {
		return translator.T("executor.query_command"), color.New(color.FgGreen)
	}
	return translator.T("executor.modify_command"), color.New(color.FgRed)
}

// DisplayCommand displays the command (without user confirmation - that's handled by session)
func (ce *CommandExecutor) DisplayCommand(cmdText string, cmdType CommandType, translator *i18n.I18n) {
	label, colorFn := ce.GetCommandTypeInfo(cmdType, translator)
	colorFn.Println(label)
	colorFn.Println(cmdText)
	fmt.Println()
}

// ExecuteCommand executes the command and returns the output
// Note: Output is not printed here, caller should print it after spinner stops
func (ce *CommandExecutor) ExecuteCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	// Caller can decide whether to treat non-zero exit as error
	return string(output), err
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
