package prompt

import (
	"strings"
	"testing"
)

func TestGetInteractivePrompt_ReturnsNonEmpty(t *testing.T) {
	prompt := GetInteractivePrompt()

	if prompt == "" {
		t.Error("Expected interactive prompt to not be empty")
	}

	// Should contain command markers
	if !strings.Contains(prompt, "[cmd:query]") {
		t.Error("Expected prompt to contain [cmd:query] marker")
	}

	if !strings.Contains(prompt, "[cmd:modify]") {
		t.Error("Expected prompt to contain [cmd:modify] marker")
	}
}

func TestGetContinueAnalysisPrompt_ReturnsNonEmpty(t *testing.T) {
	prompt := GetContinueAnalysisPrompt()

	if prompt == "" {
		t.Error("Expected continue analysis prompt to not be empty")
	}
}

func TestGetPipeAnalysisPrompt_ReturnsNonEmpty(t *testing.T) {
	prompt := GetPipeAnalysisPrompt()

	if prompt == "" {
		t.Error("Expected pipe analysis prompt to not be empty")
	}
}

func TestGetSystemPrompts_ReturnsAllPrompts(t *testing.T) {
	prompts := GetSystemPrompts()

	if prompts.Interactive == "" {
		t.Error("Expected Interactive prompt to not be empty")
	}

	if prompts.ContinueAnalysis == "" {
		t.Error("Expected ContinueAnalysis prompt to not be empty")
	}

	if prompts.PipeAnalysis == "" {
		t.Error("Expected PipeAnalysis prompt to not be empty")
	}
}

func TestPrompts_ContainCommandMarkers(t *testing.T) {
	interactive := GetInteractivePrompt()

	// Interactive prompt should contain both command types
	if !strings.Contains(interactive, "[cmd:query]") {
		t.Error("Interactive prompt should contain [cmd:query] marker")
	}

	if !strings.Contains(interactive, "[cmd:modify]") {
		t.Error("Interactive prompt should contain [cmd:modify] marker")
	}
}
