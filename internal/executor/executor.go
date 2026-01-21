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
	QueryCommand   CommandType = iota // Query command (read-only, safe)
	ModifyCommand                     // Modify command (write operations, high risk)
	UnknownCommand                    // Unknown command classification
)

// CommandExecutor handles command extraction and execution
type CommandExecutor struct{}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{}
}

// ClassifyCommand classifies the command type based on keywords
func (ce *CommandExecutor) ClassifyCommand(command string) CommandType {
	cmd := strings.ToLower(command)

	// Query command keywords (read-only operations)
	queryKeywords := []string{
		"ps ", "grep", "cat ", "ls ", "find ", "head ", "tail ",
		"wc ", "du ", "df ", "free ", "iostat", "vmstat",
		"netstat", "ss ", "lsof", "strace", "tcpdump",
		"curl", "wget", "ping", "nslookup", "dig",
	}

	for _, keyword := range queryKeywords {
		if strings.Contains(cmd, keyword) {
			return QueryCommand
		}
	}

	// Modify command keywords (write operations, dangerous)
	modifyKeywords := []string{
		"rm ", "rmdir", "mv ", "cp ", "chmod ", "chown ",
		"sed -i", "vi ", "vim ", "nano ", "ed ",
		"systemctl restart", "service restart", "systemctl stop", "systemctl start",
		"kill", "killall", "pkill",
		"useradd", "usermod", "userdel",
		"fdisk", "mkfs", "mount",
		"yum install", "apt install", "apk add", "apt-get",
		"reboot", "shutdown", "halt",
	}

	for _, keyword := range modifyKeywords {
		if strings.Contains(cmd, keyword) {
			return ModifyCommand
		}
	}

	return UnknownCommand
}

// DisplayCommand displays the command and gets user confirmation
func (ce *CommandExecutor) DisplayCommand(command string, confirmed bool, translator *i18n.I18n) bool {
	cmdType := ce.ClassifyCommand(command)

	if cmdType == QueryCommand {
		// Query command (display in green, execute directly)
		color.Green(translator.T("executor.query_command") + "\n")
		color.Green("%s\n", command)
		return true
	}

	if cmdType == ModifyCommand {
		// Modify command (display in red, requires confirmation)
		color.Red(translator.T("executor.modify_command") + "\n")
		color.Red("%s\n", command)

		// First confirmation
		var input string
		color.Yellow("\n" + translator.T("executor.execute_prompt"))
		fmt.Scanln(&input)

		if input != "yes" && input != "y" {
			color.Yellow("Cancelled\n")
			return false
		}

		// Second confirmation for critical operations
		color.Red("\n" + translator.T("executor.modify_warning"))
		fmt.Scanln(&input)

		if input != "yes" && input != "y" {
			color.Yellow("Cancelled\n")
			return false
		}

		return true
	}

	// Unknown command classification
	color.Yellow(translator.T("executor.unclassified_command") + "\n")
	color.Yellow("%s\n", command)
	var input string
	color.Yellow(translator.T("executor.execute_prompt"))
	fmt.Scanln(&input)

	return input == "yes" || input == "y"
}

// ExecuteCommand executes the command
func (ce *CommandExecutor) ExecuteCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

// ExecuteWithConfirmation displays the command and executes after user confirmation
func (ce *CommandExecutor) ExecuteWithConfirmation(command string, translator *i18n.I18n) (string, error) {
	if !ce.DisplayCommand(command, false, translator) {
		return "", fmt.Errorf("user cancelled command execution")
	}

	return ce.ExecuteCommand(command)
}

// ExtractCommands extracts executable commands from AI response text
func (ce *CommandExecutor) ExtractCommands(response string) []string {
	var commands []string
	lines := strings.Split(response, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip markdown code blocks
		if strings.HasPrefix(line, "```bash") || strings.HasPrefix(line, "```sh") {
			continue
		}

		if strings.HasPrefix(line, "```") || line == "" {
			continue
		}

		// Extract commands after shell prompt
		if strings.HasPrefix(line, "$") || strings.HasPrefix(line, "#") {
			command := strings.TrimPrefix(line, "$")
			command = strings.TrimPrefix(command, "#")
			command = strings.TrimSpace(command)

			if command != "" {
				commands = append(commands, command)
			}
		} else if isLikelyCommand(line) {
			commands = append(commands, line)
		}
	}

	return commands
}

// isLikelyCommand checks if a line is likely a shell command
func isLikelyCommand(text string) bool {
	text = strings.ToLower(strings.TrimSpace(text))

	commonBinaries := []string{
		"ps", "grep", "cat", "ls", "find", "tail", "head",
		"yum", "apt", "apt-get", "systemctl", "service",
		"curl", "wget", "netstat", "ss", "ping", "dig",
		"mount", "chmod", "chown", "kill", "reboot",
	}

	for _, bin := range commonBinaries {
		if strings.HasPrefix(text, bin+" ") || strings.HasPrefix(text, bin+"|") {
			return true
		}
	}

	return false
}

// RunCommand executes command with direct terminal output
func (ce *CommandExecutor) RunCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
