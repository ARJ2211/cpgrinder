package attempts

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ARJ2211/cpgrinder/internal/store"
)

type Model struct {
	Title    string
	Attempts []store.Attempt

	Width  int
	Height int
	Scroll int
}

func New(title string, attempts []store.Attempt, width, height int) Model {
	return Model{
		Title:    title,
		Attempts: attempts,
		Width:    width,
		Height:   height,
		Scroll:   0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.Scroll > 0 {
				m.Scroll--
			}
		case "down", "j":
			maxScroll := m.maxScroll()
			if m.Scroll < maxScroll {
				m.Scroll++
			}
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

func (m Model) View() string {
	body := m.renderBody()

	lines := strings.Split(body, "\n")
	start := m.Scroll
	if start < 0 {
		start = 0
	}
	if start > len(lines) {
		start = len(lines)
	}

	end := start + m.bodyHeight()
	if end > len(lines) {
		end = len(lines)
	}

	visible := strings.Join(lines[start:end], "\n")

	return fmt.Sprintf(
		"%s\n\n%s\n\n[esc] close  [j/k] scroll",
		m.Title,
		visible,
	)
}

func (m Model) renderBody() string {
	if len(m.Attempts) == 0 {
		return "No attempts yet for this problem."
	}

	var b strings.Builder
	for i, a := range m.Attempts {
		ts := time.Unix(a.CreatedAt, 0).Format("02 Jan 2006 03:04 PM")
		fmt.Fprintf(&b, "%d) %s  %s\n", i+1, a.Verdict, a.Status)
		fmt.Fprintf(&b, "   lang: %s\n", a.Language)
		fmt.Fprintf(&b, "   at:   %s\n", ts)
		fmt.Fprintf(&b, "   time: %ds\n", a.TimeSpentSeconds)

		if strings.TrimSpace(a.Notes) != "" {
			fmt.Fprintf(&b, "   notes: %s\n", a.Notes)
		}

		if i != len(m.Attempts)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func (m Model) bodyHeight() int {
	h := m.Height - 6
	if h < 5 {
		h = 5
	}
	return h
}

func (m Model) maxScroll() int {
	totalLines := len(strings.Split(m.renderBody(), "\n"))
	maxScroll := totalLines - m.bodyHeight()
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}
