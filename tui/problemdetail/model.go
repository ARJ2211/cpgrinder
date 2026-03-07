package problemdetail

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/glamour"

	"github.com/ARJ2211/cpgrinder/internal/solve"
	"github.com/ARJ2211/cpgrinder/internal/store"
	texlite "github.com/ARJ2211/cpgrinder/internal/textlite"
	"github.com/ARJ2211/cpgrinder/tui/attempts"
)

type ProblemDetailModel struct {
	dbStore *store.Store

	problemID  string
	title      string
	url        string
	difficulty string
	rawMD      string
	samples    []store.Sample

	width  int
	height int

	viewport viewport.Model

	// computed locally
	totalLines int
	viewH      int

	// full problem record (needed for runSamplesCmd)
	problem   store.ProblemID
	runResult samplesOverlay

	currentLang string
	langPick    langPicker

	attemptsShow  bool
	attemptsModel attempts.Model
}

type showAttemptsOKMsg struct {
	attempts []store.Attempt
}

type showAttemptsErrMsg struct {
	text string
}

func showAttemptsCmd(db *store.Store, problemID string) tea.Cmd {
	return func() tea.Msg {
		if db == nil {
			return showAttemptsErrMsg{text: "missing db store"}
		}
		if strings.TrimSpace(problemID) == "" {
			return showAttemptsErrMsg{text: "no problem loaded"}
		}

		items, err := db.ListAttemptsByProblemID(problemID, 20)
		if err != nil {
			return showAttemptsErrMsg{text: err.Error()}
		}

		return showAttemptsOKMsg{attempts: items}
	}
}

func New(dbStore *store.Store) ProblemDetailModel {
	vp := viewport.New()
	vp.YPosition = 0
	vp.SetContent("Select a problem to preview its statement")

	return ProblemDetailModel{
		dbStore:     dbStore,
		viewport:    vp,
		totalLines:  1,
		viewH:       0,
		runResult:   newSamplesOverlay(),
		langPick:    newLangPicker(),
		currentLang: "python3",
	}
}

func (m ProblemDetailModel) IsModalOpen() bool {
	return m.runResult.show || m.langPick.show || m.attemptsShow
}

func (m ProblemDetailModel) Init() tea.Cmd { return nil }

func (m ProblemDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Language picker mode
	if m.langPick.show {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "esc":
				m.langPick.close()
				return m, nil

			case "up", "k":
				if m.langPick.cursor > 0 {
					m.langPick.cursor--
				}
				return m, nil

			case "down", "j":
				if m.langPick.cursor < len(m.langPick.specs)-1 {
					m.langPick.cursor++
				}
				return m, nil

			case "enter", "return", "ctrl+m":
				spec, ok := m.langPick.selected()
				if !ok {
					return m, nil
				}
				m.langPick.running = true
				return m, setLanguageCmd(m.dbStore, m.problem, string(spec.ID))
			}

		case langSetOKMsg:
			m.currentLang = msg.lang
			m.langPick.running = false
			m.langPick.close()
			return m, nil

		case langSetErrMsg:
			m.langPick.running = false
			// reuse the samples overlay to show the error (simple and obvious)
			m.langPick.close()
			m.runResult.setText("language error: " + msg.text + "\n")
			m.runResult.show = true
			return m, nil
		}

		return m, nil
	}

	// Sample results overlay mode
	if m.runResult.show {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			if msg.String() == "esc" {
				m.runResult.close()
				return m, nil
			}
			var cmd tea.Cmd
			m.runResult.vp, cmd = m.runResult.vp.Update(msg)
			return m, cmd

		case runSamplesOKMsg:
			m.runResult.setText(msg.text)
			return m, nil

		case runSamplesErrMsg:
			m.runResult.setText(msg.text + "\n")
			return m, nil

		case editorDoneMsg:
			if msg.err != nil {
				m.runResult.setText("editor error: " + msg.err.Error() + "\n")
				m.runResult.show = true
			}
			return m, nil
		}

		var cmd tea.Cmd
		m.runResult.vp, cmd = m.runResult.vp.Update(msg)
		return m, cmd
	}

	// Attempts overlay mode
	if m.attemptsShow {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "esc", "q":
				m.attemptsShow = false
				return m, nil
			default:
				var cmd tea.Cmd
				m.attemptsModel, cmd = m.attemptsModel.Update(msg)
				return m, cmd
			}

		case tea.WindowSizeMsg:
			var cmd tea.Cmd
			m.attemptsModel, cmd = m.attemptsModel.Update(msg)
			return m, cmd
		}

		return m, nil
	}

	// Normal mode
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {

		// Run the script that you write
		case "r":
			m.attemptsShow = false
			m.runResult.setRunning()
			return m, runSamplesCmd(m.dbStore, m.problem)

		// List all the language and pick one
		case "l":
			m.attemptsShow = false
			m.langPick.open(m.currentLang)
			return m, nil

		// Open the editor so you can write your solution
		case "e":
			m.attemptsShow = false
			return m, openEditorCmd(m.dbStore, m.problem)

		// List all your prev attempts on this problem
		case "a":
			return m, showAttemptsCmd(m.dbStore, m.problemID)
		}

	case showAttemptsOKMsg:
		m.attemptsShow = true
		m.attemptsModel = attempts.New(
			"Attempts: "+m.title,
			msg.attempts,
			m.width,
			m.height,
		)
		return m, nil

	case showAttemptsErrMsg:
		m.runResult.setText("attempts error: " + msg.text + "\n")
		m.runResult.show = true
		return m, nil

	case runSamplesOKMsg:
		m.runResult.setText(msg.text)
		return m, nil

	case runSamplesErrMsg:
		m.runResult.setText(msg.text + "\n")
		return m, nil

	case tea.WindowSizeMsg:
		m = m.SetSize(msg.Width, msg.Height)
		if m.attemptsShow {
			var cmd tea.Cmd
			m.attemptsModel, cmd = m.attemptsModel.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m ProblemDetailModel) View() tea.View {
	if m.langPick.show {
		v := tea.NewView(m.langPick.view(m.width, m.height))
		v.WindowTitle = "Language"
		return v
	}

	if m.runResult.show {
		v := tea.NewView(m.runResult.view(m.width, m.height))
		v.WindowTitle = "Sample Results"
		return v
	}

	if m.attemptsShow {
		v := tea.NewView(m.attemptsModel.View())
		v.WindowTitle = "Attempts"
		return v
	}

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
	m.samples = nil
	m.totalLines = 1

	m.problem = store.ProblemID{}
	m.runResult.close()

	m.currentLang = "python3"
	m.langPick.close()

	m.viewport.SetContent("Select a problem to preview its statement")
	m.viewport.GotoTop()

	m.attemptsShow = false
	m.attemptsModel = attempts.Model{}

	// header/footer removed, expand viewport back
	m = m.SetSize(m.width, m.height)

	return m
}

func (m ProblemDetailModel) IsOverlayOpen() bool {
	return m.runResult.show || m.langPick.show || m.attemptsShow
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
	m.samples = p.Samples
	m.problem = p

	m.currentLang = "python3"
	if m.dbStore != nil {
		if dir, err := solve.EnsureProblemDir(m.dbStore.WorkspacePath(), m.problem); err == nil {
			if cfg, exists, err := (solve.RunConfig{}).ReadRunConfig(dir); err == nil && exists && strings.TrimSpace(cfg.Language) != "" {
				m.currentLang = cfg.Language
			}
		}
	}

	// close any previous results overlay
	m.runResult.close()

	// close previous attempts overlay
	m.attemptsShow = false
	m.attemptsModel = attempts.Model{}

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
	if md == "" {
		md = "(no statement found for this problem)"
	}

	// Append samples as markdown (fenced code blocks)
	md = md + samplesToMarkdown(m.samples)

	// Humanize LaTeX-ish bits
	md = texlite.HumanizeMathInMarkdown(md)

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

func samplesToMarkdown(samples []store.Sample) string {
	if len(samples) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n\n# Samples\n")

	for _, s := range samples {
		name := strings.TrimSpace(s.Name)
		if name == "" {
			name = "sample"
		}

		b.WriteString("\n### ")
		b.WriteString(name)
		b.WriteString("\n\nInput\n\n```")
		b.WriteString("\n")
		b.WriteString(strings.TrimSuffix(s.In, "\n"))
		b.WriteString("\n```\n\nOutput\n\n```")
		b.WriteString("\n")
		b.WriteString(strings.TrimSuffix(s.Out, "\n"))
		b.WriteString("\n```\n")
	}

	return b.String()
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

	line := fmt.Sprintf("%3d%%  %d-%d/%d   r run samples   a attempts   e edit/solve   l language (%s)", pct, top, bot, m.totalLines, m.currentLang)
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
