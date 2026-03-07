package problemdetail

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/ARJ2211/cpgrinder/internal/solve"
	"github.com/ARJ2211/cpgrinder/internal/store"
)

type langSetOKMsg struct {
	lang        string
	displayName string
	sourceFile  string
}

type langSetErrMsg struct {
	text string
}

type langPicker struct {
	show    bool
	running bool

	cursor int
	specs  []solve.LanguageSpec

	currentLang string
}

func newLangPicker() langPicker {
	return langPicker{
		specs: solve.ListLanguageSpecs(),
	}
}

func (p *langPicker) open(currentLang string) {
	p.show = true
	p.running = false
	p.currentLang = strings.TrimSpace(currentLang)

	// try to place cursor on current language
	if p.currentLang != "" {
		cur := solve.NormalizeLanguageID(p.currentLang)
		for i := range p.specs {
			if p.specs[i].ID == cur {
				p.cursor = i
				return
			}
		}
	}
	p.cursor = 0
}

func (p *langPicker) close() {
	p.show = false
	p.running = false
}

func (p *langPicker) selected() (solve.LanguageSpec, bool) {
	if p.cursor < 0 || p.cursor >= len(p.specs) {
		return solve.LanguageSpec{}, false
	}
	return p.specs[p.cursor], true
}

func setLanguageCmd(db *store.Store, prob store.ProblemID, lang string) tea.Cmd {
	return func() tea.Msg {
		if db == nil {
			return langSetErrMsg{text: "missing db store"}
		}
		if strings.TrimSpace(prob.Id) == "" {
			return langSetErrMsg{text: "no problem loaded"}
		}

		wsRoot := db.WorkspacePath()

		dir, err := solve.EnsureProblemDir(wsRoot, prob)
		if err != nil {
			return langSetErrMsg{text: err.Error()}
		}

		id := solve.NormalizeLanguageID(lang)
		spec, ok := solve.GetLanguageSpec(id)
		if !ok {
			return langSetErrMsg{text: fmt.Sprintf("unknown language %q", lang)}
		}

		cfg := solve.RunConfig{Language: string(spec.ID)}
		if err := cfg.WriteRunConfig(dir); err != nil {
			return langSetErrMsg{text: err.Error()}
		}

		if _, _, err := solve.EnsureStarterFile(dir, string(spec.ID)); err != nil {
			return langSetErrMsg{text: err.Error()}
		}

		return langSetOKMsg{
			lang:        string(spec.ID),
			displayName: spec.DisplayName,
			sourceFile:  spec.SourceFile,
		}
	}
}

func (p *langPicker) view(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	boxW := width - 4
	if boxW < 34 {
		boxW = 34
	}
	if boxW > 70 {
		boxW = 70
	}

	boxH := height - 4
	if boxH < 10 {
		boxH = 10
	}
	if boxH > 18 {
		boxH = 18
	}

	title := "\n\nSelect language\n\n"
	if p.running {
		title = "Setting language..."
	}

	header := lipgloss.NewStyle().Padding(0, 1).Render(title)
	footer := lipgloss.NewStyle().Padding(0, 1).Render("\n\n\nenter select | esc cancel | ↑↓ move")

	var b strings.Builder
	for i, s := range p.specs {
		cursor := " "
		if i == p.cursor {
			cursor = ">"
		}

		mark := " "
		if p.currentLang != "" && solve.NormalizeLanguageID(p.currentLang) == s.ID {
			mark = "*"
		}

		b.WriteString(fmt.Sprintf("%s [%s] %s (%s)\n", cursor, mark, s.DisplayName, s.ID))
	}

	body := lipgloss.NewStyle().Padding(0, 1).Render(strings.TrimRight(b.String(), "\n"))

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(boxW).
		Height(boxH).
		Render(header + "\n" + body + "\n" + footer)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, card)
}
