package problemdetail

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/ARJ2211/cpgrinder/internal/solve"
	"github.com/ARJ2211/cpgrinder/internal/store"
)

type runSamplesOKMsg struct {
	text string
}

type runSamplesErrMsg struct {
	text string
}

func runSamplesCmd(db *store.Store, p store.ProblemID) tea.Cmd {
	return func() tea.Msg {
		if db == nil {
			return runSamplesErrMsg{text: "missing db store"}
		}
		if strings.TrimSpace(p.Id) == "" {
			return runSamplesErrMsg{text: "no problem loaded"}
		}

		wsRoot := db.WorkspacePath()

		dir, err := solve.EnsureProblemDir(wsRoot, p)
		if err != nil {
			return runSamplesErrMsg{text: err.Error()}
		}

		cfg, exists, err := (solve.RunConfig{}).ReadRunConfig(dir)
		if err != nil {
			return runSamplesErrMsg{text: err.Error()}
		}
		if !exists {
			cfg = solve.RunConfig{Language: "python3"}
			if err := cfg.WriteRunConfig(dir); err != nil {
				return runSamplesErrMsg{text: err.Error()}
			}
		}

		lang := solve.NormalizeLanguageID(cfg.Language)
		spec, ok := solve.GetLanguageSpec(lang)
		if !ok {
			return runSamplesErrMsg{text: fmt.Sprintf("unknown language %q", cfg.Language)}
		}

		if _, _, err := solve.EnsureStarterFile(dir, cfg.Language); err != nil {
			return runSamplesErrMsg{text: err.Error()}
		}

		res, err := solve.RunSamples(context.Background(), spec, dir, p.Samples, 0)
		if err != nil {
			return runSamplesErrMsg{text: err.Error()}
		}

		return runSamplesOKMsg{text: formatSamplesResult(spec, dir, res)}
	}
}

func formatSamplesResult(spec solve.LanguageSpec, dir string, r solve.SamplesResult) string {
	var b strings.Builder

	b.WriteString("Samples\n")
	b.WriteString(fmt.Sprintf("Language: %s\n", spec.DisplayName))
	b.WriteString(fmt.Sprintf("Workspace: %s\n\n", dir))

	if r.Compile != nil {
		b.WriteString("Compile failed\n\n")
		if strings.TrimSpace(r.Compile.Stderr) != "" {
			b.WriteString(r.Compile.Stderr)
			if !strings.HasSuffix(r.Compile.Stderr, "\n") {
				b.WriteString("\n")
			}
		} else {
			b.WriteString("compile failed\n")
		}
		return b.String()
	}

	passed := 0
	for i, c := range r.Cases {
		name := strings.TrimSpace(c.Name)
		if name == "" {
			name = fmt.Sprintf("sample %d", i+1)
		}

		status := "PASS"
		if c.Run.TimedOut {
			status = "TLE"
		} else if c.Run.ExitCode != 0 {
			status = "RE"
		} else if !c.OK {
			status = "WA"
		}

		if c.OK {
			passed++
		}

		b.WriteString(fmt.Sprintf("%s: %s\n", name, status))
		b.WriteString(fmt.Sprintf("  time: %s\n", c.Run.DurationMS))

		if c.Run.TimedOut {
			b.WriteString("  timed out\n\n")
			continue
		}
		if c.Run.ExitCode != 0 {
			b.WriteString(fmt.Sprintf("  exit: %d\n", c.Run.ExitCode))
			if strings.TrimSpace(c.Run.Stderr) != "" {
				b.WriteString("  stderr:\n")
				b.WriteString(indentBlock(c.Run.Stderr, "    "))
				if !strings.HasSuffix(c.Run.Stderr, "\n") {
					b.WriteString("\n")
				}
			}
			b.WriteString("\n")
			continue
		}

		if !c.OK {
			b.WriteString(fmt.Sprintf("  first diff line: %d\n", c.Diff.FirstDiffLine))
			b.WriteString(fmt.Sprintf("  expected: %s\n", c.Diff.ExpectedLine))
			b.WriteString(fmt.Sprintf("  got:      %s\n\n", c.Diff.GotLine))
		}
	}

	b.WriteString(fmt.Sprintf("Summary: %d/%d passed\n", passed, len(r.Cases)))
	return b.String()
}

func indentBlock(s, prefix string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i := range lines {
		if lines[i] == "" && i == len(lines)-1 {
			continue
		}
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}

type samplesOverlay struct {
	show    bool
	running bool
	vp      viewport.Model
}

func newSamplesOverlay() samplesOverlay {
	vp := viewport.New()
	vp.SetContent("")
	return samplesOverlay{vp: vp}
}

func (o *samplesOverlay) setRunning() {
	o.show = true
	o.running = true
	o.vp.SetContent("Running...\n")
	o.vp.GotoTop()
}

func (o *samplesOverlay) setText(text string) {
	o.show = true
	o.running = false
	o.vp.SetContent(text)
	o.vp.GotoTop()
}

func (o *samplesOverlay) close() {
	o.show = false
	o.running = false
	o.vp.SetContent("")
	o.vp.GotoTop()
}

func (o *samplesOverlay) view(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	boxW := width - 4
	if boxW < 30 {
		boxW = 30
	}
	if boxW > 110 {
		boxW = 110
	}

	boxH := height - 4
	if boxH < 10 {
		boxH = 10
	}

	o.vp.SetWidth(boxW - 2)
	o.vp.SetHeight(boxH - 4)

	title := "Sample results\n\n"
	if o.running {
		title = "Running samples..."
	}

	header := lipgloss.NewStyle().Padding(0, 1).Render(title)
	footer := lipgloss.NewStyle().Padding(0, 1).Render("esc close  |  ↑/↓ scroll")

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(boxW).
		Height(boxH).
		Render(header + "\n" + o.vp.View() + "\n" + footer)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, card)
}
