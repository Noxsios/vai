package vai

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

// TODO: not really sure how to unit test this

type model struct {
	yes bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		switch msg.String() {
		case "y":
			m.yes = true
			return m, tea.Quit
		case "n":
			m.yes = false
			return m, tea.Quit
		case "q", "esc":
			return m, tea.Quit
		default:
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	return "There is a SHA mismatch, do you want to overwrite? (y/n)"
}

func (m model) Value() bool {
	return m.yes
}

// ConfirmSHAOverwrite asks the user if they want to overwrite the SHA
// and bypass the check.
func ConfirmSHAOverwrite(ctx context.Context) (bool, error) {
	p := tea.NewProgram(model{}, tea.WithContext(ctx))
	m, err := p.Run()
	if err != nil {
		return false, err
	}
	return m.(model).Value(), nil
}
