package problemlist

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/ARJ2211/cpgrinder/internal/store"
	pdm "github.com/ARJ2211/cpgrinder/tui/problemdetail"
	"github.com/ARJ2211/cpgrinder/tui/styles"
)

var PAGE_LIMIT int = 40

type focusPane int

const (
	focusList focusPane = iota
	focusDetail
)

type ProblemListModel struct {
	dbStore  *store.Store    // dbStore
	choices  []store.Problem // List of problems
	cursor   int             // The position of the problem
	page     int             // Pagination offset
	selected string          // The ID of the problem
	count    int             // Count of the total problems

	problemStmt pdm.ProblemDetailModel // Problem statement

	width  int
	height int
	leftW  int
	rightW int
	focus  focusPane
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
		problemStmt: pdm.New(dbStore),
		focus:       focusList,
	}, nil
}

func (m ProblemListModel) Init() tea.Cmd {
	return nil
}

func (m ProblemListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.leftW = clamp(m.width/3, 32, 56)
		m.rightW = m.width - m.leftW - 1
		if m.rightW < 20 {
			m.rightW = 20
		}
		m.problemStmt = m.problemStmt.SetSize(m.rightW, m.height)
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			if m.focus == focusList {
				m.focus = focusDetail
			} else {
				m.focus = focusList
			}
			return m, nil

		case "up", "k":
			if m.focus == focusDetail {
				updated, cmd := m.problemStmt.Update(msg)
				m.problemStmt = updated.(pdm.ProblemDetailModel)
				return m, cmd
			}
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.focus == focusDetail {
				updated, cmd := m.problemStmt.Update(msg)
				m.problemStmt = updated.(pdm.ProblemDetailModel)
				return m, cmd
			}
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "pgup", "pgdown", "home", "end", "g", "G":
			if m.focus == focusDetail {
				updated, cmd := m.problemStmt.Update(msg)
				m.problemStmt = updated.(pdm.ProblemDetailModel)
				return m, cmd
			}

		case "enter", "space":
			if m.focus == focusDetail {
				return m, nil
			}
			if len(m.choices) == 0 {
				return m, nil
			}

			id := m.choices[m.cursor].Id
			if m.selected == id {
				m.selected = ""
				m.problemStmt = m.problemStmt.Clear()
				return m, nil
			}

			m.selected = id
			detail, err := m.problemStmt.LoadProblem(id)
			if err != nil {
				m.problemStmt = m.problemStmt.
					Clear().
					SetMessage("failed to load problem: "+err.Error()).
					SetSize(m.rightW, m.height)
				return m, nil
			}
			m.problemStmt = detail
			m.focus = focusDetail
			return m, nil

		case "n", "right":
			if m.focus == focusDetail {
				updated, cmd := m.problemStmt.Update(msg)
				m.problemStmt = updated.(pdm.ProblemDetailModel)
				return m, cmd
			}

			for _, c := range m.choices {
				if c.Id == m.selected {
					m.selected = ""
					m.problemStmt = m.problemStmt.Clear()
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
			m.focus = focusList

		case "b", "left":
			if m.focus == focusDetail {
				updated, cmd := m.problemStmt.Update(msg)
				m.problemStmt = updated.(pdm.ProblemDetailModel)
				return m, cmd
			}

			for _, c := range m.choices {
				if c.Id == m.selected {
					m.selected = ""
					m.problemStmt = m.problemStmt.Clear()
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
			m.focus = focusList
		}
	}

	return m, nil
}

func (m ProblemListModel) View() tea.View {
	left := m.renderLeftPane()
	right := m.renderRightPane()

	content := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	v := tea.NewView(content)
	v.WindowTitle = "Problem List"
	return v
}

func (m ProblemListModel) renderLeftPane() string {
	var b string
	b = "List of questions currently in the question bank: \n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if choice.Id == m.selected {
			checked = "x"
		}

		b += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice.Title)
	}

	if m.count > 0 {
		b += fmt.Sprintf("\n(%d/%d)", m.cursor+(m.page*PAGE_LIMIT)+1, m.count)
	}

	legend := "[tab] focus | [↑/↓] move/scroll | [enter] select | [n/b] page | [esc] back | [q] quit"

	b += "\n\n"
	b += styles.LegendStyle.Render(legend)

	border := lipgloss.NewStyle().
		Width(m.leftW).
		Height(m.height).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, true, false, false)

	if m.focus == focusList {
		border = border.BorderForeground(lipgloss.Color("62"))
	}

	return border.Render(b)
}

func (m ProblemListModel) renderRightPane() string {
	right := lipgloss.NewStyle().
		Width(m.rightW).
		Height(m.height).
		Padding(0, 1)

	if m.focus == focusDetail {
		right = right.Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("62"))
	}

	return right.Render(m.problemStmt.View().Content)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
