package ui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/llaoj/aiassist/internal/i18n"
)

// Common errors
var (
	ErrUserAbort = errors.New("user aborted")
	ErrUserExit  = errors.New("user exit")
	ErrUserDone  = errors.New("user chose to exit") // User chose "No" to exit
)

// inputModel is a custom text input model using bubbletea
type inputModel struct {
	textInput textinput.Model
	prompt    string
	quitting  bool
	err       error
}

func newInputModel(prompt string) inputModel {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Prompt = "> "
	ti.Focus()
	ti.CharLimit = 0
	ti.Width = 80

	return inputModel{
		textInput: ti,
		prompt:    prompt,
	}
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}
func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			m.err = ErrUserAbort
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	if m.quitting {
		return ""
	}
	return m.prompt + "\n" + m.textInput.View()
}

// selectModel is a custom select model using bubbletea
type selectModel struct {
	prompt   string
	options  []string
	selected int
	quitting bool
	err      error
}

func newSelectModel(prompt string, options []string) selectModel {
	return selectModel{
		prompt:   prompt,
		options:  options,
		selected: 0,
	}
}

func (m selectModel) Init() tea.Cmd {
	return nil
}

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			m.err = ErrUserAbort
			return m, tea.Quit
		case tea.KeyUp, tea.KeyLeft:
			if m.selected > 0 {
				m.selected--
			}
		case tea.KeyDown, tea.KeyRight:
			if m.selected < len(m.options)-1 {
				m.selected++
			}
		}
	}

	return m, nil
}

func (m selectModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder
	b.WriteString(m.prompt + "\n\n")

	for i, option := range m.options {
		if i == m.selected {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Render("> "+option) + "\n")
		} else {
			b.WriteString("  " + option + "\n")
		}
	}

	return b.String()
}

// PromptInput displays an input prompt and returns the user's input
func PromptInput(prompt string, translator *i18n.I18n) (string, error) {
	// Use custom bubbletea input for better control
	model := newInputModel(prompt)
	p := tea.NewProgram(model)
	final, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("input error: %w", err)
	}

	m := final.(inputModel)
	if m.err != nil {
		return "", m.err
	}

	return m.textInput.Value(), nil
}

// PromptConfirm displays a confirmation prompt and returns the result
func PromptConfirm(prompt string, translator *i18n.I18n) (bool, error) {
	options := []string{"Yes", "No"}
	model := newSelectModel(prompt, options)
	p := tea.NewProgram(model)
	final, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("confirmation error: %w", err)
	}

	m := final.(selectModel)
	if m.err != nil {
		return false, m.err
	}

	return m.selected == 0, nil
}

// PromptConfirmWithDefault displays a confirmation prompt with a default value
func PromptConfirmWithDefault(prompt string, defaultValue bool, translator *i18n.I18n) (bool, error) {
	options := []string{"Yes", "No"}
	model := newSelectModel(prompt, options)

	// Set default selection
	if defaultValue {
		model.selected = 0 // Yes
	} else {
		model.selected = 1 // No
	}

	p := tea.NewProgram(model)
	final, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("confirmation error: %w", err)
	}

	m := final.(selectModel)
	if m.err != nil {
		return false, m.err
	}

	return m.selected == 0, nil
}

// PromptSelect displays a selection list and returns the selected option
func PromptSelect(prompt string, options []string) (string, error) {
	model := newSelectModel(prompt, options)
	p := tea.NewProgram(model)
	final, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("selection error: %w", err)
	}

	m := final.(selectModel)
	if m.err != nil {
		return "", m.err
	}

	return m.options[m.selected], nil
}

// IsTerminal returns true if stdout is a terminal
func IsTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
