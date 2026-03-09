package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ARJ2211/cpgrinder/internal/store"

	"github.com/ARJ2211/cpgrinder/tui/problemlist"
	"github.com/ARJ2211/cpgrinder/tui/progress"
	"github.com/ARJ2211/cpgrinder/tui/styles"
)

type sessionState int

const (
	projectView sessionState = iota
	problemlistView
	progressTracker
	notImplemented
)

type MainModel struct {
	dbStore      *store.Store
	prevState    sessionState
	state        sessionState
	stateChoices []string

	index  int
	cursor int
	width  int
	height int

	promblemListView problemlist.ProblemListModel
	progressTracker  progress.ProgressTrackerModel
}

func InitializeModel(dbStore *store.Store) (MainModel, error) {
	promblemListView, err := problemlist.InitializeModel(dbStore)
	if err != nil {
		return MainModel{}, err
	}

	progressTracker, err := progress.InitializeModel(dbStore)
	if err != nil {
		return MainModel{}, err
	}

	return MainModel{
		dbStore:          dbStore,
		prevState:        projectView,
		state:            projectView,
		stateChoices:     []string{"List Problems", "Show Activity", "Some more features..."},
		promblemListView: promblemListView,
		progressTracker:  progressTracker,
	}, nil
}

func (m MainModel) Init() tea.Cmd { return nil }

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

		if m.state == problemlistView {
			updated, cmd := m.promblemListView.Update(msg)
			if lm, ok := updated.(problemlist.ProblemListModel); ok {
				m.promblemListView = lm
			}
			return m, cmd
		}

		if m.state == progressTracker {
			updated, cmd := m.progressTracker.Update(msg)
			if lm, ok := updated.(progress.ProgressTrackerModel); ok {
				m.progressTracker = lm
			}
			return m, cmd
		}
		return m, nil

	case tea.KeyPressMsg:
		switch m.state {

		case projectView:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
				return m, nil
			case "down", "j":
				if m.cursor < len(m.stateChoices)-1 {
					m.cursor++
				}
				return m, nil
			case "enter", "space":
				// When user selects "List Problems"
				if m.cursor == 0 {
					m.state = problemlistView
					ws := tea.WindowSizeMsg{Width: m.width, Height: m.height}
					updated, cmd := m.promblemListView.Update(ws)
					if lm, ok := updated.(problemlist.ProblemListModel); ok {
						m.promblemListView = lm
					}
					return m, cmd
				}

				// When user selects show activity
				if m.cursor == 1 {
					m.state = progressTracker

					ws := tea.WindowSizeMsg{Width: m.width, Height: m.height}
					updated, cmd := m.progressTracker.Update(ws)
					if lm, ok := updated.(progress.ProgressTrackerModel); ok {
						m.progressTracker = lm
					}
					return m, cmd
				}

				m.state = notImplemented
				return m, nil
			}

		case problemlistView:
			// only global quit keys here; everything else must go to the child model
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}
			updated, cmd := m.promblemListView.Update(msg)
			if lm, ok := updated.(problemlist.ProblemListModel); ok {
				m.promblemListView = lm
			}
			return m, cmd

		case progressTracker:
			// only global quit keys here
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}
			updated, cmd := m.progressTracker.Update(msg)
			if lm, ok := updated.(progress.ProgressTrackerModel); ok {
				m.progressTracker = lm
			}
			return m, cmd

		case notImplemented:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				m.state = projectView
				return m, nil
			}
		}

	case problemlist.BackToProjectMsg:
		m.state = projectView
		return m, nil

	case progress.BackToProjectMsg:
		m.state = projectView
		return m, nil

	default:
		// crucial: forward command results to the active child model
		if m.state == problemlistView {
			updated, cmd := m.promblemListView.Update(msg)
			if lm, ok := updated.(problemlist.ProblemListModel); ok {
				m.promblemListView = lm
			}
			return m, cmd
		}
	}

	return m, nil
}

func (m MainModel) renderProjectView() string {
	banner := styles.RenderBanner()

	menuDescriptions := []string{
		"Browse the synced problem set and open something worth solving.",
		"Track streaks, solved counts, and recent terminal activity.",
		"Extra experiments and upcoming features. Free free to reach out or contribute your own :)",
	}

	totalWidth := 100
	if m.width > 0 {
		totalWidth = clamp(m.width-12, 92, 116)
	}

	hero := styles.HeroCardStyle.Width(totalWidth).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			styles.KickerStyle.Render("terminal-first competitive programming\n\n"),
			banner,
			styles.SubtitleStyle.Render("Your terminal-based competitive coding platform                                       "),
		),
	)

	leftWidth := 58
	rightWidth := totalWidth - leftWidth - 2
	if totalWidth < 104 {
		leftWidth = totalWidth
		rightWidth = totalWidth
	}

	menuLines := make([]string, 0, len(m.stateChoices))
	for i, choice := range m.stateChoices {
		desc := ""
		if i < len(menuDescriptions) {
			desc = menuDescriptions[i]
		}

		row := lipgloss.JoinVertical(
			lipgloss.Left,
			styles.MenuLabelStyle.Render(choice),
			styles.MenuDescStyle.Render(desc),
		)

		if m.cursor == i {
			menuLines = append(menuLines, styles.SelectedMenuItemStyle.Width(leftWidth-6).Render(row))
		} else {
			menuLines = append(menuLines, styles.MenuItemStyle.Width(leftWidth-6).Render(row))
		}
	}

	menuCard := styles.CardStyle.Width(leftWidth).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			styles.SectionTitleStyle.Render("What do you want to do today?                       "),
			lipgloss.JoinVertical(lipgloss.Left, menuLines...),
		),
	)

	sideCard := styles.InfoCardStyle.Width(rightWidth).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			styles.SectionTitleStyle.Render("Quick start"),
			renderShortcut("↑/↓ or j/k", "move through the menu"),
			renderShortcut("enter", "open the selected view"),
			renderShortcut("q", "quit the app"),
			"",
		),
	)

	var body string
	if totalWidth >= 104 {
		body = lipgloss.JoinHorizontal(
			lipgloss.Top,
			menuCard,
			"  ",
			sideCard,
		)
	} else {
		body = lipgloss.JoinVertical(
			lipgloss.Left,
			menuCard,
			sideCard,
		)
	}

	footer := styles.HelpStyle.Render("Ready when you are.")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		hero,
		body,
		footer,
	)

	content = styles.AppStyle.Render(content)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, content)
	}

	return content
}

func renderShortcut(key, description string) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		styles.KeyStyle.Render(key),
		" ",
		styles.MutedStyle.Render(description),
	)
}

func clamp(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

func (m MainModel) renderNotImplemented() string {
	msg := fmt.Sprintf("%s is not yet implemented... Coming soon :)", m.stateChoices[m.cursor])
	help := styles.HelpStyle.Render("Press esc to go back • q to quit")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.NotImplStyle.Render(msg),
		help,
	)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	return content
}

func (m MainModel) View() tea.View {
	switch m.state {

	case projectView:
		v := tea.NewView(m.renderProjectView())
		v.WindowTitle = "CPGrinder"
		return v

	case problemlistView:
		return m.promblemListView.View()

	case progressTracker:
		return m.progressTracker.View()

	default:
		v := tea.NewView(m.renderNotImplemented())
		v.WindowTitle = "CPGrinder"
		return v
	}
}
