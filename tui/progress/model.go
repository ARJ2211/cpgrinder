package progress

import (
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/ARJ2211/cpgrinder/tui/styles"
)

type BackToProjectMsg struct{}

type ProgressTrackerModel struct {
	width  int
	height int
	table  table.Model
}

func (m ProgressTrackerModel) Init() tea.Cmd { return nil }

func (m ProgressTrackerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.table.SetHeight(msg.Height - 5)
		m.table.SetWidth(msg.Width - 5)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg { return BackToProjectMsg{} }
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m ProgressTrackerModel) View() tea.View {
	v := tea.NewView(
		styles.TableStyle.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n",
	)
	v.WindowTitle = "Progress Tracker"

	return v
}
