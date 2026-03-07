package problemlist

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/ARJ2211/cpgrinder/internal/store"
	"github.com/ARJ2211/cpgrinder/tui/filtersearch"
	pdm "github.com/ARJ2211/cpgrinder/tui/problemdetail"
	"github.com/ARJ2211/cpgrinder/tui/styles"
)

var PAGE_LIMIT int = 30

type focusPane int
type BackToProjectMsg struct{}

const (
	focusList focusPane = iota
	focusDetail
)

type ProblemListModel struct {
	dbStore *store.Store

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

	filter filtersearch.Model

	didInitialLayout bool
}

func InitializeModel(dbStore *store.Store) (ProblemListModel, error) {
	m := ProblemListModel{
		dbStore:     dbStore,
		choices:     nil,
		cursor:      0,
		page:        0,
		selected:    "",
		count:       0,
		problemStmt: pdm.New(dbStore),
		focus:       focusList,
		filter:      filtersearch.New(),
	}
	// initial fetch (no size yet, so we don't load detail until WindowSize)
	if err := m.fetchPage(); err != nil {
		return ProblemListModel{}, err
	}
	if len(m.choices) > 0 {
		m.selected = m.choices[0].Id
	}
	return m, nil
}

func (m ProblemListModel) Init() tea.Cmd { return nil }

func (m ProblemListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		usableH := max(1, m.height-1)
		m.computeWidths()

		innerW := max(20, m.rightW-2)
		m.problemStmt = m.problemStmt.SetSize(innerW, usableH)

		if !m.didInitialLayout {
			m.didInitialLayout = true

			if m.selected == "" && len(m.choices) > 0 {
				m.cursor = 0
				m.selected = m.choices[0].Id
			}

			if m.selected != "" {
				detail, err := m.problemStmt.LoadProblem(m.selected)
				if err != nil {
					m.problemStmt = m.problemStmt.Clear().
						SetMessage("failed to load problem: "+err.Error()).
						SetSize(innerW, usableH)
					return m, nil
				}
				m.problemStmt = detail.SetSize(innerW, usableH)
			}

			return m, nil
		}

		return m, nil

	case tea.KeyPressMsg:
		// modal owns input
		if m.filter.Open {
			var action filtersearch.Action
			m.filter, action = m.filter.Update(msg)

			if action == filtersearch.ActionApply || action == filtersearch.ActionReset {
				m.page = 0
				if err := m.fetchPage(); err != nil {
					// show error in detail pane
					usableH := max(1, m.height-1)
					innerW := max(20, m.rightW-2)
					m.problemStmt = m.problemStmt.Clear().
						SetMessage("failed to fetch problems: "+err.Error()).
						SetSize(innerW, usableH)
					return m, nil
				}

				m.cursor = 0
				m.selected = ""
				usableH := max(1, m.height-1)
				innerW := max(20, m.rightW-2)
				m.problemStmt = m.problemStmt.Clear().SetSize(innerW, usableH)
				m = m.selectFirstAndRender(innerW, usableH)
			}
			return m, nil
		}

		// If the detail pane has any modal open (samples overlay or language picker),
		// it must receive keypresses first (esc/enter/up/down/etc).
		if m.problemStmt.IsModalOpen() {
			// keep global quit keys
			if msg.String() == "ctrl+c" || msg.String() == "q" {
				return m, tea.Quit
			}

			updated, cmd := m.problemStmt.Update(msg)
			if dm, ok := updated.(pdm.ProblemDetailModel); ok {
				m.problemStmt = dm
			}
			m.focus = focusDetail
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "f", "F":
			m.filter = m.filter.Toggle()
			m.focus = focusList
			return m, nil

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
			m.selected = id

			usableH := max(1, m.height-1)
			innerW := max(20, m.rightW-2)
			m.problemStmt = m.problemStmt.SetSize(innerW, usableH)

			detail, err := m.problemStmt.LoadProblem(id)
			if err != nil {
				m.problemStmt = m.problemStmt.Clear().
					SetMessage("failed to load problem: "+err.Error()).
					SetSize(innerW, usableH)
				m.focus = focusList
				return m, nil
			}

			m.problemStmt = detail.SetSize(innerW, usableH)
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
			if err := m.fetchPage(); err != nil {
				return m, nil
			}

			m.cursor = 0
			m.selected = ""
			m.focus = focusList

			usableH := max(1, m.height-1)
			innerW := max(20, m.rightW-2)
			m.problemStmt = m.problemStmt.Clear().SetSize(innerW, usableH)
			m = m.selectFirstAndRender(innerW, usableH)
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
			if err := m.fetchPage(); err != nil {
				return m, nil
			}

			m.cursor = 0
			m.selected = ""
			m.focus = focusList

			usableH := max(1, m.height-1)
			innerW := max(20, m.rightW-2)
			m.problemStmt = m.problemStmt.Clear().SetSize(innerW, usableH)
			m = m.selectFirstAndRender(innerW, usableH)
			return m, nil

		case "e":
			updated, cmd := m.problemStmt.Update(msg)
			if dm, ok := updated.(pdm.ProblemDetailModel); ok {
				m.problemStmt = dm
			}
			m.focus = focusDetail
			return m, cmd

		case "r":
			updated, cmd := m.problemStmt.Update(msg)
			if dm, ok := updated.(pdm.ProblemDetailModel); ok {
				m.problemStmt = dm
			}
			m.focus = focusDetail
			return m, cmd

		case "l":
			updated, cmd := m.problemStmt.Update(msg)
			if dm, ok := updated.(pdm.ProblemDetailModel); ok {
				m.problemStmt = dm
			}
			m.focus = focusDetail
			return m, cmd

		case "a":
			updated, cmd := m.problemStmt.Update(msg)
			if dm, ok := updated.(pdm.ProblemDetailModel); ok {
				m.problemStmt = dm
			}
			m.focus = focusDetail
			return m, cmd

		case "esc":
			// if sample overlay is open, close it first
			if m.problemStmt.IsOverlayOpen() {
				updated, cmd := m.problemStmt.Update(msg)
				if dm, ok := updated.(pdm.ProblemDetailModel); ok {
					m.problemStmt = dm
				}
				return m, cmd
			}
			// otherwise go back to the main menu
			return m, func() tea.Msg { return BackToProjectMsg{} }
		}

	default:
		updated, cmd := m.problemStmt.Update(msg)
		if dm, ok := updated.(pdm.ProblemDetailModel); ok {
			m.problemStmt = dm
			return m, cmd
		}
		return m, nil
	}

	return m, nil
}

func (m ProblemListModel) View() tea.View {
	left := m.renderLeftPane()
	sep := m.renderSeparator()
	right := m.renderRightPane()

	content := "\n" + lipgloss.JoinHorizontal(lipgloss.Top, left, sep, right)
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

	b.WriteString("Question Bank:\n\n")

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

	// modal above legend
	if m.filter.Open {
		b.WriteString("\n\n")
		b.WriteString(m.filter.View(max(24, m.leftW-2)))
	}

	legend := fmt.Sprintf(
		"focus=%s | [tab] switch | [f] filter\n[↑/↓] move/scroll | [enter] open\n[n/b] page | [esc] back | [q] quit",
		focusTxt,
	)
	b.WriteString("\n\n")
	b.WriteString(styles.LegendStyle.Render(legend))

	usableH := max(1, m.height-1)
	style := lipgloss.NewStyle().
		Width(m.leftW).
		Height(usableH).
		Padding(0, 1)

	return style.Render(b.String())
}

func (m ProblemListModel) renderSeparator() string {
	usableH := max(1, m.height-1)

	var b strings.Builder
	for i := 0; i < usableH; i++ {
		b.WriteString("|")
		if i < usableH-1 {
			b.WriteByte('\n')
		}
	}

	col := lipgloss.Color("240")
	if m.focus == focusDetail {
		col = lipgloss.Color("62")
	}

	return lipgloss.NewStyle().Foreground(col).Render(b.String())
}

func (m ProblemListModel) renderRightPane() string {
	usableH := max(1, m.height-1)

	style := lipgloss.NewStyle().
		Width(m.rightW).
		Height(usableH).
		Padding(0, 1)

	return style.Render(m.problemStmt.View().Content)
}

func (m *ProblemListModel) computeWidths() {
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
}

func (m ProblemListModel) currentFilters() store.UserFilters {
	uf := store.UserFilters{
		Title: m.filter.Title,
		Topic: m.filter.Topic,
		Tag:   m.filter.Tag,
	}

	if m.filter.Source != "" && m.filter.Source != "all" {
		uf.Source = m.filter.Source
	}
	if m.filter.Difficulty != "" && m.filter.Difficulty != "all" {
		uf.Difficulty = m.filter.Difficulty
	}

	return uf
}

func (m *ProblemListModel) fetchPage() error {
	uf := m.currentFilters()
	uf.Limit = PAGE_LIMIT
	uf.Offset = PAGE_LIMIT * m.page

	problems, err := m.dbStore.ListProblems(uf)
	if err != nil {
		return err
	}
	m.choices = problems

	// if you add CountProblemsWithFilters (below), this becomes correct for filtered paging
	if c, err := m.dbStore.CountProblemsWithFilters(m.currentFilters()); err == nil {
		m.count = c
	}

	// fallback (if you don't add the store method yet)
	if m.count == 0 {
		if c2, err := m.dbStore.CountProblems(); err == nil && (m.filter.Title == "" && m.filter.Source == "all" && m.filter.Difficulty == "all" && m.filter.Topic == "" && m.filter.Tag == "") {
			m.count = c2
		} else {
			// unknown filtered total, but prevents divide-by-zero / weird display
			m.count = len(m.choices)
		}
	}

	return nil
}

func (m ProblemListModel) selectFirstAndRender(innerW, usableH int) ProblemListModel {
	if len(m.choices) == 0 {
		m.selected = ""
		m.cursor = 0
		m.problemStmt = m.problemStmt.Clear().SetSize(innerW, usableH)
		return m
	}

	m.cursor = 0
	m.selected = m.choices[0].Id

	detail, err := m.problemStmt.LoadProblem(m.selected)
	if err != nil {
		m.problemStmt = m.problemStmt.Clear().
			SetMessage("failed to load problem: "+err.Error()).
			SetSize(innerW, usableH)
		return m
	}

	m.problemStmt = detail.SetSize(innerW, usableH)
	return m
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
