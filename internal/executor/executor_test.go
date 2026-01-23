package executor

import (
	"testing"
)

func TestNewCommandExecutor(t *testing.T) {
	ce := NewCommandExecutor()
	if ce == nil {
		t.Fatal("Expected CommandExecutor to be created")
	}
}

func TestCommandType(t *testing.T) {
	if QueryCommand != 0 {
		t.Errorf("Expected QueryCommand to be 0, got %d", QueryCommand)
	}

	if ModifyCommand != 1 {
		t.Errorf("Expected ModifyCommand to be 1, got %d", ModifyCommand)
	}
}

func TestCommand(t *testing.T) {
	cmd := Command{
		Text: "ls -la",
		Type: QueryCommand,
	}

	if cmd.Text != "ls -la" {
		t.Errorf("Expected Text to be 'ls -la', got '%s'", cmd.Text)
	}

	if cmd.Type != QueryCommand {
		t.Errorf("Expected Type to be QueryCommand, got %d", cmd.Type)
	}
}

func TestExtractCommands_WithQueryCommands(t *testing.T) {
	ce := NewCommandExecutor()

	response := `Here are the commands to run:
1. Check disk space
   [cmd:query] df -h

2. Check memory
   [cmd:query] free -m
`

	commands := ce.ExtractCommands(response)

	if len(commands) != 2 {
		t.Fatalf("Expected 2 commands, got %d", len(commands))
	}

	if commands[0].Text != "df -h" {
		t.Errorf("Expected first command to be 'df -h', got '%s'", commands[0].Text)
	}

	if commands[0].Type != QueryCommand {
		t.Errorf("Expected first command type to be QueryCommand, got %d", commands[0].Type)
	}

	if commands[1].Text != "free -m" {
		t.Errorf("Expected second command to be 'free -m', got '%s'", commands[1].Text)
	}
}

func TestExtractCommands_WithModifyCommands(t *testing.T) {
	ce := NewCommandExecutor()

	response := `To fix this:
[cmd:modify] sudo systemctl restart nginx
`

	commands := ce.ExtractCommands(response)

	if len(commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(commands))
	}

	if commands[0].Type != ModifyCommand {
		t.Errorf("Expected command type to be ModifyCommand, got %d", commands[0].Type)
	}

	if commands[0].Text != "sudo systemctl restart nginx" {
		t.Errorf("Expected command to be 'sudo systemctl restart nginx', got '%s'", commands[0].Text)
	}
}

func TestExtractCommands_WithMarkdownFormatting(t *testing.T) {
	ce := NewCommandExecutor()

	response := `Run this command:
[cmd:query] **ps aux** | grep nginx
`

	commands := ce.ExtractCommands(response)

	if len(commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(commands))
	}

	// Should strip markdown formatting
	if commands[0].Text != "ps aux | grep nginx" {
		t.Errorf("Expected markdown to be stripped, got '%s'", commands[0].Text)
	}
}

func TestExtractCommands_NoCommands(t *testing.T) {
	ce := NewCommandExecutor()

	response := `This is just explanatory text without any commands.`

	commands := ce.ExtractCommands(response)

	if len(commands) != 0 {
		t.Errorf("Expected 0 commands, got %d", len(commands))
	}
}

func TestExtractCommands_EmptyCommand(t *testing.T) {
	ce := NewCommandExecutor()

	response := `[cmd:query]    
[cmd:modify]
`

	commands := ce.ExtractCommands(response)

	// Empty commands should be filtered out
	if len(commands) != 0 {
		t.Errorf("Expected 0 commands (empty commands filtered), got %d", len(commands))
	}
}

func TestExecuteCommand_Success(t *testing.T) {
	ce := NewCommandExecutor()

	// Simple command that should succeed
	output, err := ce.ExecuteCommand("echo test")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Output might have newline, just check it contains "test"
	if output != "test\n" && output != "test" {
		t.Errorf("Expected output to contain 'test', got '%s'", output)
	}
}

func TestExecuteCommand_WithOutput(t *testing.T) {
	ce := NewCommandExecutor()

	output, err := ce.ExecuteCommand("echo 'hello world'")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check output contains our text
	if output == "" {
		t.Error("Expected output, got empty string")
	}
}

func TestDisplayCommand_QueryCommand(t *testing.T) {
	// This test would require mocking user input, skip for now
	t.Skip("Skipping interactive test")
}

func TestDisplayCommand_ModifyCommand(t *testing.T) {
	// This test would require mocking user input, skip for now
	t.Skip("Skipping interactive test")
}

func TestExecuteWithConfirmation(t *testing.T) {
	// This test would require mocking user input, skip for now
	t.Skip("Skipping interactive test")
}
