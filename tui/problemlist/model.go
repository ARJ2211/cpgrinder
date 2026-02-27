package problemlist

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/ARJ2211/cpgrinder/internal/store"
	"github.com/ARJ2211/cpgrinder/tui/styles"
)

var PAGE_LIMIT int = 40

type ProblemListModel struct {
	dbStore  *store.Store    // dbStore
	choices  []store.Problem // List of problems
	cursor   int             // The position of the problem
	page     int             // Pagination offset
	selected string          // The ID of the problem
	count    int             // Count of the total problems

	problemStmt string // Problem statement
}

// Initialize it with a list of problems in the database
func InitializeModel(dbStore *store.Store) (ProblemListModel, error) {
	filters := store.UserFilters{
		Limit: PAGE_LIMIT,
	}

	problems, err := dbStore.ListProblems(filters)
	if err != nil {
		return ProblemListModel{}, err
	}

	count, err := dbStore.CountProblems()
	if err != nil {
		return ProblemListModel{}, err
	}

	return ProblemListModel{
		dbStore:     dbStore,
		choices:     problems,
		selected:    "",
		page:        0,
		count:       count,
		problemStmt: "",
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
			var prevSelection string
			for _, c := range m.choices {
				if c.Id == m.selected {
					prevSelection = m.selected
					m.selected = ""
					m.problemStmt = ""
				}
			}

			if prevSelection != m.choices[m.cursor].Id {
				m.selected = m.choices[m.cursor].Id

				p, err := m.dbStore.GetProblemByID(m.selected)
				if err != nil {
					return ProblemListModel{}, nil
				}

				m.problemStmt = p.StatementMd
			}

		// Increment page by 1
		case "n", "right":
			for _, c := range m.choices {
				if c.Id == m.selected {
					m.selected = ""
				}
			}
			if (m.page+1)*PAGE_LIMIT < m.count {
				m.page = m.page + 1
			}

			uf := store.UserFilters{
				Limit:  PAGE_LIMIT,
				Offset: PAGE_LIMIT * m.page,
			}

			var err error
			m.choices, err = m.dbStore.ListProblems(uf)

			if err != nil {
				return ProblemListModel{}, nil
			}

			m.cursor = 0

		// Decrement page by 1
		case "b", "left":
			for _, c := range m.choices {
				if c.Id == m.selected {
					m.selected = ""
				}
			}

			if m.page > 0 {
				m.page = m.page - 1
			}

			uf := store.UserFilters{
				Limit:  PAGE_LIMIT,
				Offset: PAGE_LIMIT * m.page,
			}

			var err error
			m.choices, err = m.dbStore.ListProblems(uf)

			if err != nil {
				return ProblemListModel{}, nil
			}

			m.cursor = 0
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

	s += fmt.Sprintf("\nProblem: (%d/%d)", m.cursor+(m.page*PAGE_LIMIT)+1, m.count)
	s += fmt.Sprintf("\nSelected Problem ID: %s\n", m.selected)

	legend := styles.LegendStyle.Render("Legend: [↑/↓] Navigate | [Enter] Select | [q] Quit | [n/→] Next | [b/←] Back")
	s += legend

	p, err := styles.GlamourMD.Render(m.problemStmt)
	if err != nil {
		s = "FAILED TO RENDER MARKDOWN"
	}

	s = lipgloss.JoinHorizontal(lipgloss.Left, s, p)

	v := tea.NewView(s)
	v.WindowTitle = "Problem List"

	return v
}
