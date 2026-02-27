package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/ARJ2211/cpgrinder/internal/store"

	"github.com/ARJ2211/cpgrinder/tui/problemlist"
)

type sessionState int

const (
	projectView sessionState = iota
	problemlistView
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

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

	case tea.KeyPressMsg:
		switch m.state {

		// When state is the project view (HOME PAGE)
		case projectView:
			switch msg.String() {
			// Exit
			case "ctrl+c", "q":
				return m, tea.Quit

			// Move up
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}

			// Move down
			case "down", "j":
				if m.cursor < len(m.stateChoices)-1 {
					m.cursor++
				}

			// Slection of view
			case "enter", "space":
				if m.cursor == 0 {
					m.state = problemlistView
				} else {
					m.state = notImplemented
				}
				return m, nil
			}

		// When the state is the list problems view
		case problemlistView:
			switch msg.String() {

			// Exit
			case "ctrl+c", "q":
				return m, tea.Quit

			// Go back to the prev screen
			case "esc":
				m.state = projectView
				return m, nil
			}

			updated, cmd := m.promblemListView.Update(msg)
			if lm, ok := updated.(problemlist.ProblemListModel); ok {
				m.promblemListView = lm
			} else {
				m.state = projectView
				return m, nil
			}

			return m, cmd

		//When the state is not yet implemented
		case notImplemented:
			switch msg.String() {

			// Exit
			case "ctrl+c", "q":
				return m, tea.Quit

			// Go back to the prev screen
			case "esc":
				m.state = projectView
				return m, nil
			}
		}
	}

	return m, nil
}

func (m MainModel) View() tea.View {
	switch m.state {

	// View controller given to the home screen
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

		v := tea.NewView(s)
		return v

	// View controller given to the problem list view
	case problemlistView:
		v := m.promblemListView.View()
		return v

	// Default case when some random state is passed
	default:
		msg := fmt.Sprintf(
			"%s is not yet implemented... Coming soon :)",
			m.stateChoices[m.cursor],
		)
		v := tea.NewView(msg)
		v.WindowTitle = "CpGrinder"
		return v
	}
}
