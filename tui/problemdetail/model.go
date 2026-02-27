package problemdetail

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/glamour"

	"github.com/ARJ2211/cpgrinder/internal/store"
)

type ProblemDetailModel struct {
	dbStore *store.Store

	problemID  string
	title      string
	url        string
	difficulty string
	rawMD      string

	width  int
	height int

	viewport viewport.Model
}

func New(dbStore *store.Store) ProblemDetailModel {
	vp := viewport.New()
	vp.SetContent("Select a problem to preview its statement")

	return ProblemDetailModel{
		dbStore:  dbStore,
		viewport: vp,
	}
}

func (m ProblemDetailModel) Init() tea.Cmd { return nil }

func (m ProblemDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m ProblemDetailModel) View() tea.View {
	if m.width <= 0 || m.height <= 0 {
		return tea.NewView("")
	}

	header := m.renderHeader()
	content := lipgloss.JoinVertical(lipgloss.Top, header, m.viewport.View())
	return tea.NewView(content)
}

func (m ProblemDetailModel) SetSize(width, height int) ProblemDetailModel {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	prevW := m.width
	m.width = width
	m.height = height

	headerH := m.headerHeight()
	viewportH := height - headerH
	if viewportH < 1 {
		viewportH = 1
	}

	m.viewport.SetWidth(width)
	m.viewport.SetHeight(viewportH)

	if prevW != width && m.problemID != "" {
		m = m.renderAndSetContent()
	}

	return m
}

func (m ProblemDetailModel) Clear() ProblemDetailModel {
	m.problemID = ""
	m.title = ""
	m.url = ""
	m.difficulty = ""
	m.rawMD = ""
	m = m.SetMessage("Select a problem to preview its statement")
	return m
}

func (m ProblemDetailModel) SetMessage(msg string) ProblemDetailModel {
	if strings.TrimSpace(msg) == "" {
		msg = " "
	}
	m.viewport.SetContent(msg)
	m.viewport.GotoTop()
	return m
}

func (m ProblemDetailModel) LoadProblem(id string) (ProblemDetailModel, error) {
	if m.dbStore == nil {
		return m, fmt.Errorf("detail model is missing dbStore")
	}

	p, err := m.dbStore.GetProblemByID(id)
	if err != nil {
		return m, err
	}

	m.problemID = p.Id
	m.title = p.Title
	m.url = p.Url
	m.difficulty = p.Difficulty
	m.rawMD = p.StatementMd

	m = m.renderAndSetContent()
	m.viewport.GotoTop()

	return m, nil
}

func (m ProblemDetailModel) renderAndSetContent() ProblemDetailModel {
	wrapW := m.width
	if wrapW < 20 {
		wrapW = 20
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(wrapW-2),
		glamour.WithChromaFormatter(">"),
	)
	if err != nil {
		m.viewport.SetContent("failed to init markdown renderer")
		return m
	}

	md := strings.TrimSpace(m.rawMD)
	if md == "" {
		md = "(no statement found for this problem)"
	}

	rendered, err := r.Render(md)
	if err != nil {
		m.viewport.SetContent("failed to render markdown")
		return m
	}

	m.viewport.SetContent(rendered)
	return m
}

func (m ProblemDetailModel) headerHeight() int {
	return 3
}

func (m ProblemDetailModel) renderHeader() string {
	if m.problemID == "" {
		return lipgloss.NewStyle().Width(m.width).Render("")
	}

	maxW := m.width
	if maxW < 10 {
		maxW = 10
	}

	title := lipgloss.NewStyle().Width(maxW).Render(m.title)
	meta := lipgloss.NewStyle().Width(maxW).Render(fmt.Sprintf("%s  %s", m.difficulty, m.url))
	sep := lipgloss.NewStyle().Width(maxW).Render(strings.Repeat("─", maxW))

	return lipgloss.JoinVertical(lipgloss.Top, title, meta, sep)
}
