package problemlist

import (
	"fmt"
	"strings"

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
	dbStore  *store.Store
	choices  []store.Problem
	cursor   int
	page     int
	selected string
	count    int

	problemStmt pdm.ProblemDetailModel

	width  int
	height int
	leftW  int
	rightW int
	focus  focusPane
}

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

func (m ProblemListModel) Init() tea.Cmd { return nil }

func (m ProblemListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Make right pane a bit narrower by giving list more width
		m.leftW = clamp(int(float64(m.width)*0.45), 38, 72)

		sepW := 1
		m.rightW = m.width - m.leftW - sepW
		if m.rightW < 24 {
			m.rightW = 24
			m.leftW = m.width - m.rightW - sepW
			if m.leftW < 20 {
				m.leftW = 20
			}
		}

		// Right pane wrapper uses Padding(0,1) so inner content width shrinks by 2
		innerW := m.rightW - 2
		if innerW < 20 {
			innerW = 20
		}

		m.problemStmt = m.problemStmt.SetSize(innerW, m.height)
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab", "ctrl+i", "shift+tab":
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
			return m, nil

		case "down", "j":
			if m.focus == focusDetail {
				updated, cmd := m.problemStmt.Update(msg)
				m.problemStmt = updated.(pdm.ProblemDetailModel)
				return m, cmd
			}
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
			return m, nil

		case "pgup", "pgdown", "home", "end", "g", "G":
			if m.focus == focusDetail {
				updated, cmd := m.problemStmt.Update(msg)
				m.problemStmt = updated.(pdm.ProblemDetailModel)
				return m, cmd
			}
			return m, nil

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
				m.focus = focusList
				return m, nil
			}

			m.selected = id
			detail, err := m.problemStmt.LoadProblem(id)
			if err != nil {
				m.problemStmt = m.problemStmt.
					Clear().
					SetMessage("failed to load problem: "+err.Error()).
					SetSize(max(20, m.rightW-2), m.height)
				m.focus = focusList
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

			if (m.page+1)*PAGE_LIMIT < m.count {
				m.page++
			}

			uf := store.UserFilters{
				Limit:  PAGE_LIMIT,
				Offset: PAGE_LIMIT * m.page,
			}

			var err error
			m.choices, err = m.dbStore.ListProblems(uf)
			if err != nil {
				return m, nil
			}

			m.cursor = 0
			m.selected = ""
			m.problemStmt = m.problemStmt.Clear()
			m.focus = focusList
			return m, nil

		case "b", "left":
			if m.focus == focusDetail {
				updated, cmd := m.problemStmt.Update(msg)
				m.problemStmt = updated.(pdm.ProblemDetailModel)
				return m, cmd
			}

			if m.page > 0 {
				m.page--
			}

			uf := store.UserFilters{
				Limit:  PAGE_LIMIT,
				Offset: PAGE_LIMIT * m.page,
			}

			var err error
			m.choices, err = m.dbStore.ListProblems(uf)
			if err != nil {
				return m, nil
			}

			m.cursor = 0
			m.selected = ""
			m.problemStmt = m.problemStmt.Clear()
			m.focus = focusList
			return m, nil
		}
	}

	return m, nil
}

func (m ProblemListModel) View() tea.View {
	left := m.renderLeftPane()
	sep := m.renderSeparator()
	right := m.renderRightPane()

	content := lipgloss.JoinHorizontal(lipgloss.Top, left, sep, right)
	v := tea.NewView(content)
	v.WindowTitle = "Problem List"
	return v
}

func (m ProblemListModel) renderLeftPane() string {
	var b strings.Builder

	focusTxt := "list"
	if m.focus == focusDetail {
		focusTxt = "detail"
	}

	b.WriteString("Practice\n\n")

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if choice.Id == m.selected {
			checked = "x"
		}

		b.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice.Title))
	}

	if m.count > 0 {
		b.WriteString(fmt.Sprintf("\n(%d/%d)", m.cursor+(m.page*PAGE_LIMIT)+1, m.count))
	}

	legend := fmt.Sprintf("focus=%s | [tab] switch | [↑/↓] move/scroll | [enter] select | [n/b] page | [esc] back | [q] quit", focusTxt)
	b.WriteString("\n\n")
	b.WriteString(styles.LegendStyle.Render(legend))

	style := lipgloss.NewStyle().
		Width(m.leftW).
		Height(m.height).
		Padding(0, 1)

	return style.Render(b.String())
}

func (m ProblemListModel) renderSeparator() string {
	var b strings.Builder
	for i := 0; i < m.height; i++ {
		b.WriteString("│")
		if i < m.height-1 {
			b.WriteByte('\n')
		}
	}

	// Just tint the separator based on focus (no layout changes)
	col := lipgloss.Color("240")
	if m.focus == focusDetail {
		col = lipgloss.Color("62")
	}

	return lipgloss.NewStyle().Foreground(col).Render(b.String())
}

func (m ProblemListModel) renderRightPane() string {
	style := lipgloss.NewStyle().
		Width(m.rightW).
		Height(m.height).
		Padding(0, 1)

	return style.Render(m.problemStmt.View().Content)
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
