package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/ARJ2211/cpgrinder/internal/store"
)

type ProblemListModel struct {
	dbStore  store.Store     // dbStore
	choices  []store.Problem // List of problems
	cursor   int             // The position of the problem
	selected string          // The ID of the problem
}

// Initialize it with a list of problems in the database
func InitialModel(dbStore *store.Store) (ProblemListModel, error) {
	filters := store.UserFilters{
		Limit: 30,
	}

	problems, err := dbStore.ListProblems(filters)
	if err != nil {
		return ProblemListModel{}, err
	}

	return ProblemListModel{
		choices:  problems,
		selected: "",
	}, nil
}

func (m ProblemListModel) Init() tea.Cmd {
	return nil
}

func (m ProblemListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {

		// Exit
		case "ctrl+c", "q":
			return m, tea.Quit

		// Move up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// Move down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// Select problem
		case "enter", "space":
			for _, c := range m.choices {
				if c.Id == m.selected {
					m.selected = ""
				}
			}
			m.selected = m.choices[m.cursor].Id
		}
	}

	return m, nil
}

func (m ProblemListModel) View() tea.View {
	s := "Select a problem that you would want to solve\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if choice.Id == m.selected {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice.Title)
	}

	s += fmt.Sprintf("\nSelected Problem ID: %s\n", m.selected)
	s += "\nPress q to quit.\n"

	v := tea.NewView(s)
	v.WindowTitle = "Problem List"

	return v
}
