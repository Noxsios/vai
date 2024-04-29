package vai

import tea "github.com/charmbracelet/bubbletea"

type model struct {
	yes bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "y" {
			m.yes = true
			return m, tea.Quit
		}
		if msg.String() == "n" {
			m.yes = false
			return m, tea.Quit
		}
		if msg.String() == "q" {
			return m, tea.Quit
		}
		if msg.String() == "esc" {
			return m, tea.Quit
		}
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
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

func ConfirmSHAOverwrite() (bool, error) {
	p := tea.NewProgram(model{})
	m, err := p.Run()
	if err != nil {
		return false, err
	}
	return m.(model).Value(), nil
}
