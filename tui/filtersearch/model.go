package filtersearch

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Field int

const (
	FieldTitle Field = iota
	FieldSource
	FieldDifficulty
	FieldTopic
	FieldTag
)

type Action int

const (
	ActionNone Action = iota
	ActionApply
	ActionClose
	ActionReset
)

type Model struct {
	Open bool

	Title      string
	Source     string // "all" | "codeforces" | "leetcode"
	Difficulty string // "all" | "easy" | "medium" | "hard" | "expert"
	Topic      string
	Tag        string

	field         Field
	MaxQueryRunes int
}

func New() Model {
	return Model{
		Open:          false,
		Title:         "",
		Source:        "all",
		Difficulty:    "all",
		Topic:         "",
		Tag:           "",
		field:         FieldTitle,
		MaxQueryRunes: 64,
	}
}

func (m Model) Toggle() Model {
	m.Open = !m.Open
	if m.Open {
		m.field = FieldTitle
	}
	return m
}

func (m Model) Close() Model {
	m.Open = false
	return m
}

func (m Model) Reset() Model {
	m.Title = ""
	m.Source = "all"
	m.Difficulty = "all"
	m.Topic = ""
	m.Tag = ""
	m.field = FieldTitle
	return m
}

func (m Model) Update(msg tea.Msg) (Model, Action) {
	if !m.Open {
		return m, ActionNone
	}

	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, ActionNone
	}

	switch k.String() {
	case "esc":
		m.Open = false
		return m, ActionClose

	case "enter":
		m.Open = false
		return m, ActionApply

	case "ctrl+r":
		m = m.Reset()
		return m, ActionReset

	case "tab", "ctrl+i":
		m.field = nextField(m.field, 1)
		return m, ActionNone

	case "shift+tab":
		m.field = nextField(m.field, -1)
		return m, ActionNone

	case "up", "k", "left", "h":
		if m.field == FieldSource {
			m.Source = cycleSource(m.Source, -1)
			return m, ActionNone
		}
		if m.field == FieldDifficulty {
			m.Difficulty = cycleDiff(m.Difficulty, -1)
			return m, ActionNone
		}

	case "down", "j", "right", "l":
		if m.field == FieldSource {
			m.Source = cycleSource(m.Source, 1)
			return m, ActionNone
		}
		if m.field == FieldDifficulty {
			m.Difficulty = cycleDiff(m.Difficulty, 1)
			return m, ActionNone
		}

	case "backspace", "delete":
		m = m.backspaceActiveField()
		return m, ActionNone

	case "ctrl+u":
		m = m.clearActiveField()
		return m, ActionNone

	case "space":
		// Bubble Tea v2 reports spacebar as "space"
		m = m.appendToActiveField(" ")
		return m, ActionNone
	}

	// text input for Title/Topic/Tag only
	s := k.String()
	if isPrintableSingleRune(s) {
		m = m.appendToActiveField(s)
	}

	return m, ActionNone
}

func (m Model) View(width int) string {
	if !m.Open || width <= 0 {
		return ""
	}

	w := width
	if w < 24 {
		w = 24
	}

	title := "Filter / Search:"
	hint := "[tab] next  [shift+tab] prev  [↑/↓] cycle options \n[enter] apply  [esc] close  [ctrl+r] reset  [ctrl+u] clear field"

	ph := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	titleVal := m.Title
	if titleVal == "" {
		titleVal = ph.Render("<type title>")
	}
	topicVal := m.Topic
	if topicVal == "" {
		topicVal = ph.Render("<type topic>")
	}
	tagVal := m.Tag
	if tagVal == "" {
		tagVal = ph.Render("<type tag>")
	}

	lineTitle := fmt.Sprintf("Title: %s%s", titleVal, caret(m.field == FieldTitle))
	lineSource := fmt.Sprintf("Source: %s%s", m.Source, caret(m.field == FieldSource))
	lineDiff := fmt.Sprintf("Difficulty: %s%s", m.Difficulty, caret(m.field == FieldDifficulty))
	lineTopic := fmt.Sprintf("Topic: %s%s", topicVal, caret(m.field == FieldTopic))
	lineTag := fmt.Sprintf("Tag: %s%s", tagVal, caret(m.field == FieldTag))

	body := strings.Join([]string{
		title,
		"",
		lineTitle,
		lineSource,
		lineDiff,
		lineTopic,
		lineTag,
		"",
		hint,
	}, "\n")

	box := lipgloss.NewStyle().
		Width(w).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	return box.Render(body)
}

func caret(on bool) string {
	if on {
		return " ▏"
	}
	return ""
}

func nextField(f Field, dir int) Field {
	min := int(FieldTitle)
	max := int(FieldTag)

	n := int(f) + dir
	if n < min {
		n = max
	}
	if n > max {
		n = min
	}
	return Field(n)
}

func cycleSource(cur string, dir int) string {
	opts := []string{"all", "codeforces", "leetcode"}
	return cycle(opts, cur, dir)
}

func cycleDiff(cur string, dir int) string {
	opts := []string{"all", "easy", "medium", "hard", "expert"}
	return cycle(opts, cur, dir)
}

func cycle(opts []string, cur string, dir int) string {
	idx := 0
	for i := range opts {
		if opts[i] == cur {
			idx = i
			break
		}
	}
	idx += dir
	if idx < 0 {
		idx = len(opts) - 1
	}
	if idx >= len(opts) {
		idx = 0
	}
	return opts[idx]
}

func (m Model) clearActiveField() Model {
	switch m.field {
	case FieldTitle:
		m.Title = ""
	case FieldTopic:
		m.Topic = ""
	case FieldTag:
		m.Tag = ""
	}
	return m
}

func (m Model) backspaceActiveField() Model {
	switch m.field {
	case FieldTitle:
		m.Title = popRune(m.Title)
	case FieldTopic:
		m.Topic = popRune(m.Topic)
	case FieldTag:
		m.Tag = popRune(m.Tag)
	}
	return m
}

func (m Model) appendToActiveField(s string) Model {
	switch m.field {
	case FieldTitle:
		if len([]rune(m.Title)) < m.MaxQueryRunes {
			m.Title += s
		}
	case FieldTopic:
		if len([]rune(m.Topic)) < m.MaxQueryRunes {
			m.Topic += s
		}
	case FieldTag:
		if len([]rune(m.Tag)) < m.MaxQueryRunes {
			m.Tag += s
		}
	}
	return m
}

func popRune(s string) string {
	if len(s) == 0 {
		return s
	}
	_, size := utf8.DecodeLastRuneInString(s)
	if size <= 0 || size > len(s) {
		return ""
	}
	return s[:len(s)-size]
}

func isPrintableSingleRune(s string) bool {
	rs := []rune(s)
	if len(rs) != 1 {
		return false
	}
	r := rs[0]
	if r < 32 || r == 127 {
		return false
	}
	if s == "\t" || s == "\n" {
		return false
	}
	return true
}
