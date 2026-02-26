package ui

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/llaoj/aiassist/internal/i18n"
)

// Common errors
var (
	ErrUserAbort = errors.New("user aborted")
	ErrUserExit  = errors.New("user exit")
)

// getInputTheme returns a minimal theme without the left vertical bar and indentation
func getInputTheme() *huh.Theme {
	t := huh.ThemeBase()
	t.Focused.Base = lipgloss.NewStyle()
	t.Blurred.Base = lipgloss.NewStyle()
	return t
}

// getSelectTheme returns a minimal theme for select without the left vertical bar and indentation
func getSelectTheme() *huh.Theme {
	t := huh.ThemeBase()
	t.Focused.Base = lipgloss.NewStyle()
	t.Blurred.Base = lipgloss.NewStyle()
	t.Focused.Title = t.Focused.Title.PaddingLeft(0)
	t.Blurred.Title = t.Blurred.Title.PaddingLeft(0)
	return t
}


// PromptInput displays an input prompt and returns the user's input
func PromptInput(prompt string, translator *i18n.I18n) (string, error) {
	var input string

	form := huh.NewInput().
		Title(prompt).
		Value(&input).
		WithTheme(getInputTheme())

	// Use RunAccessible to prevent redraw issues on terminal resize
	err := form.RunAccessible(os.Stdout, os.Stdin)

	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrUserAbort
		}
		return "", fmt.Errorf("input error: %w", err)
	}

	return input, nil
}

// PromptInputWithHistory displays an input prompt with history support
func PromptInputWithHistory(prompt string, suggestions []string, translator *i18n.I18n) (string, error) {
	var input string

	// Create input field with title and value
	inputField := huh.NewInput().
		Title(prompt).
		Value(&input)

	// Add suggestions if provided
	if len(suggestions) > 0 {
		inputField = inputField.Suggestions(suggestions)
	}

	// Apply theme and use RunAccessible to prevent redraw issues
	err := inputField.WithTheme(getInputTheme()).RunAccessible(os.Stdout, os.Stdin)

	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrUserAbort
		}
		return "", fmt.Errorf("input error: %w", err)
	}

	return input, nil
}

// PromptConfirm displays a confirmation prompt and returns the result
func PromptConfirm(prompt string, translator *i18n.I18n) (bool, error) {
	var selected string

	form := huh.NewSelect[string]().
		Title(prompt).
		Options(
			huh.NewOption("Yes", "yes"),
			huh.NewOption("No", "no"),
		).
		Value(&selected).
		WithTheme(getSelectTheme())

	// Use RunAccessible to prevent redraw issues on terminal resize
	err := form.RunAccessible(os.Stdout, os.Stdin)

	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, ErrUserAbort
		}
		return false, fmt.Errorf("confirmation error: %w", err)
	}

	return selected == "yes", nil
}

// PromptConfirmWithDefault displays a confirmation prompt with a default value
func PromptConfirmWithDefault(prompt string, defaultValue bool, translator *i18n.I18n) (bool, error) {
	var selected string

	defaultOption := "no"
	if defaultValue {
		defaultOption = "yes"
	}

	form := huh.NewSelect[string]().
		Title(prompt).
		Options(
			huh.NewOption("Yes", "yes"),
			huh.NewOption("No", "no"),
		).
		Value(&selected).
		WithTheme(getSelectTheme())

	// Use RunAccessible to prevent redraw issues on terminal resize
	err := form.RunAccessible(os.Stdout, os.Stdin)

	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, ErrUserAbort
		}
		return false, fmt.Errorf("confirmation error: %w", err)
	}

	_ = defaultOption // Avoid unused variable warning

	return selected == "yes", nil
}

// PromptSelect displays a selection list and returns the selected option
func PromptSelect(prompt string, options []string) (string, error) {
	var selected string

	// Convert options to huh options
	huhOptions := make([]huh.Option[string], len(options))
	for i, opt := range options {
		huhOptions[i] = huh.NewOption(opt, opt)
	}

	form := huh.NewSelect[string]().
		Title(prompt).
		Options(huhOptions...).
		Value(&selected)

	// Use RunAccessible to prevent redraw issues on terminal resize
	err := form.RunAccessible(os.Stdout, os.Stdin)

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

	form := huh.NewMultiSelect[string]().
		Title(prompt).
		Options(huhOptions...).
		Value(&selected)

	// Use RunAccessible to prevent redraw issues on terminal resize
	err := form.RunAccessible(os.Stdout, os.Stdin)

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

	form := huh.NewText().
		Title(prompt).
		Value(&text)

	// Use RunAccessible to prevent redraw issues on terminal resize
	err := form.RunAccessible(os.Stdout, os.Stdin)

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
