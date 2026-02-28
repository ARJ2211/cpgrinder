package problemdetail

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/glamour"

	"github.com/ARJ2211/cpgrinder/internal/store"
	texlite "github.com/ARJ2211/cpgrinder/internal/textlite"
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

	// computed locally so we don't depend on viewport internals
	totalLines int
	viewH      int
}

func New(dbStore *store.Store) ProblemDetailModel {
	vp := viewport.New()
	vp.YPosition = 0
	vp.SetContent("Select a problem to preview its statement")

	return ProblemDetailModel{
		dbStore:    dbStore,
		viewport:   vp,
		totalLines: 1,
		viewH:      0,
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
	footer := m.renderFooter()
	content := lipgloss.JoinVertical(lipgloss.Top, header, m.viewport.View(), footer)

	v := tea.NewView(content)
	v.WindowTitle = "Problem Detail"

	return v
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

	headerH := 0
	footerH := 0
	if m.problemID != "" {
		headerH = 3
		footerH = 2
	}

	vh := height - headerH - footerH
	if vh < 1 {
		vh = 1
	}
	m.viewH = vh

	m.viewport.SetWidth(width)
	m.viewport.SetHeight(vh)

	// re-wrap markdown when width changes
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
	m.totalLines = 1

	m.viewport.SetContent("Select a problem to preview its statement")
	m.viewport.GotoTop()

	// header/footer removed, expand viewport back
	m = m.SetSize(m.width, m.height)

	return m
}

func (m ProblemDetailModel) SetMessage(msg string) ProblemDetailModel {
	if strings.TrimSpace(msg) == "" {
		msg = " "
	}
	m.viewport.SetContent(msg)
	m.totalLines = countLines(msg)
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

	// header/footer now exist, so resize viewport accordingly
	m = m.SetSize(m.width, m.height)

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
		msg := "failed to init markdown renderer"
		m.viewport.SetContent(msg)
		m.totalLines = countLines(msg)
		return m
	}

	md := strings.TrimSpace(m.rawMD)
	md = texlite.HumanizeMathInMarkdown(md)
	if md == "" {
		md = "(no statement found for this problem)"
	}

	rendered, err := r.Render(md)
	if err != nil {
		msg := "failed to render markdown"
		m.viewport.SetContent(msg)
		m.totalLines = countLines(msg)
		return m
	}

	m.viewport.SetContent(rendered)
	m.totalLines = countLines(rendered)
	return m
}

func (m ProblemDetailModel) renderHeader() string {
	if m.problemID == "" || m.width <= 0 {
		return ""
	}

	w := m.width
	title := fitLine(m.title, w)
	meta := fitLine(fmt.Sprintf("%s  %s", m.difficulty, m.url), w)
	sep := strings.Repeat("─", w)

	return lipgloss.JoinVertical(lipgloss.Top, title, meta, sep)
}

func (m ProblemDetailModel) renderFooter() string {
	if m.problemID == "" || m.width <= 0 {
		return ""
	}

	w := m.width
	sep := strings.Repeat("─", w)

	pct := int(m.viewport.ScrollPercent()*100 + 0.5)

	maxTop := m.totalLines - m.viewH
	if maxTop < 0 {
		maxTop = 0
	}
	top0 := int(m.viewport.ScrollPercent() * float64(maxTop))
	top := top0 + 1
	bot := top0 + m.viewH
	if bot > m.totalLines {
		bot = m.totalLines
	}

	line := fmt.Sprintf("%3d%%  %d-%d/%d", pct, top, bot, m.totalLines)
	line = fitLine(line, w)

	return sep + "\n" + line
}

func countLines(s string) int {
	if s == "" {
		return 1
	}
	return strings.Count(s, "\n") + 1
}

func fitLine(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if len(s) > w {
		return s[:w]
	}
	return s + strings.Repeat(" ", w-len(s))
}
