package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/ARJ2211/cpgrinder/internal/store"
)

type model struct {
	dbStore  store.Store
	choices  []string
	cursor   int
	selected map[int]struct{}
}

// Initialize it with a list of problems in the database
func InitialModel(dbStore *store.Store) (model, error) {
	filters := store.UserFilters{
		Limit: 10,
	}

	problems, err := dbStore.ListProblems(filters)
	if err != nil {
		return model{}, err
	}

	var choices []string
	for _, p := range problems {
		choices = append(choices, p.Title)
	}

	return model{
		choices:  choices,
		selected: make(map[int]struct{}),
	}, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyPressMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The enter key and the space bar toggle the selected state for the
		// item that the cursor is pointing at.
		case "enter", "space":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() tea.View {
	s := "What should we buy at the market?\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	s += "\nPress q to quit.\n"

	v := tea.NewView(s)
	v.WindowTitle = "Grocery List"

	return v
}
