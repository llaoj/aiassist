package ui

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
)

// Common errors
var (
	ErrUserAbort = errors.New("user aborted")
	ErrUserExit  = errors.New("user exit")
)

// Track Ctrl+C presses for double-press to exit
var (
	ctrlCCount   int
	ctrlCMutex   sync.Mutex
	ctrlCMessage = "Press Ctrl+C again to exit"
)

// getDefaultKeyMap returns a custom keymap with Tab for accepting suggestions
func getDefaultKeyMap() *huh.KeyMap {
	keymap := huh.NewDefaultKeyMap()
	// Override the AcceptSuggestion key to use Tab instead of Ctrl+E
	keymap.Input.AcceptSuggestion = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "complete"),
	)
	return keymap
}

// PromptInput displays an input prompt and returns the user's input
func PromptInput(prompt string) (string, error) {
	var input string

	err := huh.NewInput().
		Title(prompt).
		Value(&input).
		WithKeyMap(getDefaultKeyMap()).
		Run()

	if err != nil {
		// Check if user pressed Ctrl+C or Ctrl+D
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrUserAbort
		}
		return "", fmt.Errorf("input error: %w", err)
	}

	return input, nil
}

// PromptInputWithHistory displays an input prompt with history support
func PromptInputWithHistory(prompt string, suggestions []string) (string, error) {
	var input string

	// Create input field with title and value
	inputField := huh.NewInput().
		Title(prompt).
		Value(&input)

	// Add suggestions if provided (must be called before WithKeyMap)
	if len(suggestions) > 0 {
		inputField = inputField.Suggestions(suggestions)
	}

	// Apply custom keymap (must be last)
	err := inputField.WithKeyMap(getDefaultKeyMap()).Run()

	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrUserAbort
		}
		return "", fmt.Errorf("input error: %w", err)
	}

	return input, nil
}

// PromptConfirm displays a confirmation prompt and returns the result
func PromptConfirm(prompt string) (bool, error) {
	var selected string

	for {
		err := huh.NewSelect[string]().
			Title(prompt).
			Options(
				huh.NewOption("Yes", "yes"),
				huh.NewOption("No", "no"),
			).
			Value(&selected).
			Run()

		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				// Check if this is the second Ctrl+C press
				ctrlCMutex.Lock()
				ctrlCCount++
				if ctrlCCount >= 2 {
					ctrlCCount = 0
					ctrlCMutex.Unlock()
					return false, ErrUserAbort
				}
				// First Ctrl+C - show message and continue
				fmt.Println("\n" + ctrlCMessage)
				ctrlCMutex.Unlock()
				continue
			}
			return false, fmt.Errorf("confirmation error: %w", err)
		}

		// Reset counter on successful selection
		ctrlCMutex.Lock()
		ctrlCCount = 0
		ctrlCMutex.Unlock()

		return selected == "yes", nil
	}
}

// PromptConfirmWithDefault displays a confirmation prompt with a default value
func PromptConfirmWithDefault(prompt string, defaultValue bool) (bool, error) {
	var selected string

	defaultOption := "no"
	if defaultValue {
		defaultOption = "yes"
	}

	for {
		err := huh.NewSelect[string]().
			Title(prompt).
			Options(
				huh.NewOption("Yes", "yes"),
				huh.NewOption("No", "no"),
			).
			Value(&selected).
			Run()

		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				// Check if this is the second Ctrl+C press
				ctrlCMutex.Lock()
				ctrlCCount++
				if ctrlCCount >= 2 {
					ctrlCCount = 0
					ctrlCMutex.Unlock()
					return false, ErrUserAbort
				}
				// First Ctrl+C - show message and continue
				fmt.Println("\n" + ctrlCMessage)
				ctrlCMutex.Unlock()
				continue
			}
			return false, fmt.Errorf("confirmation error: %w", err)
		}

		// Reset counter on successful selection
		ctrlCMutex.Lock()
		ctrlCCount = 0
		ctrlCMutex.Unlock()

		_ = defaultOption // Avoid unused variable warning

		return selected == "yes", nil
	}
}

// PromptSelect displays a selection list and returns the selected option
func PromptSelect(prompt string, options []string) (string, error) {
	var selected string

	// Convert options to huh options
	huhOptions := make([]huh.Option[string], len(options))
	for i, opt := range options {
		huhOptions[i] = huh.NewOption(opt, opt)
	}

	err := huh.NewSelect[string]().
		Title(prompt).
		Options(huhOptions...).
		Value(&selected).
		Run()

	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrUserAbort
		}
		return "", fmt.Errorf("selection error: %w", err)
	}

	return selected, nil
}

// PromptMultiSelect displays a multi-select list and returns the selected options
func PromptMultiSelect(prompt string, options []string) ([]string, error) {
	var selected []string

	// Convert options to huh options
	huhOptions := make([]huh.Option[string], len(options))
	for i, opt := range options {
		huhOptions[i] = huh.NewOption(opt, opt)
	}

	err := huh.NewMultiSelect[string]().
		Title(prompt).
		Options(huhOptions...).
		Value(&selected).
		Run()

	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, ErrUserAbort
		}
		return nil, fmt.Errorf("multi-selection error: %w", err)
	}

	return selected, nil
}

// PromptText displays a multi-line text input
func PromptText(prompt string) (string, error) {
	var text string

	err := huh.NewText().
		Title(prompt).
		Value(&text).
		Run()

	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrUserAbort
		}
		return "", fmt.Errorf("text input error: %w", err)
	}

	return text, nil
}

// IsTerminal returns true if stdout is a terminal
func IsTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
