package progress

import (
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/ARJ2211/cpgrinder/tui/styles"
)

type ProgressTracker struct {
	table table.Model
}

func (m ProgressTracker) Init() tea.Cmd { return nil }

func (m ProgressTracker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "crtl+c":
			return m, tea.Quit
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m ProgressTracker) View() tea.View {
	return tea.NewView(
		styles.TableStyle.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n",
	)
}
