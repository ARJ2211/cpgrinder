package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/ARJ2211/cpgrinder/internal/store"

	"github.com/ARJ2211/cpgrinder/tui/problemlist"
	"github.com/ARJ2211/cpgrinder/tui/progress"
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

	return MainModel{
		dbStore:          dbStore,
		prevState:        projectView,
		state:            projectView,
		stateChoices:     []string{"List Problems", "Show Activity", "Some more features..."},
		promblemListView: promblemListView,
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

func (m MainModel) View() tea.View {
	switch m.state {

	case projectView:
		s := "Welcome to CpGrinder - Your terminal based competitive coding platform\n\n"
		s += "Please select what you would like to do today! \n\n"

		for i, choice := range m.stateChoices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			checked := " "
			if m.cursor == i {
				checked = "x"
			}

			s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
		}

		return tea.NewView(s)

	case problemlistView:
		return m.promblemListView.View()

	case progressTracker:
		return m.progressTracker.View()

	default:
		msg := fmt.Sprintf("%s is not yet implemented... Coming soon :)", m.stateChoices[m.cursor])
		v := tea.NewView(msg)
		v.WindowTitle = "CpGrinder"
		return v
	}
}
