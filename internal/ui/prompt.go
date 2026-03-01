package ui

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/llaoj/aiassist/internal/i18n"
)

// inputModel is a custom text input model using bubbletea
type inputModel struct {
	textInput textinput.Model
	prompt    string
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
		case tea.KeyCtrlC:
			// Send SIGINT to trigger global interrupt handler
			// This unifies all exit logic in one place
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	return m.prompt + "\n" + m.textInput.View()
}

// selectModel is a custom select model using bubbletea
type selectModel struct {
	prompt   string
	options  []string
	selected int
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
		case tea.KeyCtrlC:
			// Send SIGINT to trigger global interrupt handler
			// This unifies all exit logic in one place
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
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
	return m.selected == 0, nil
}
