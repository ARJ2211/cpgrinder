package progress

import (
	"fmt"
	"strconv"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ARJ2211/cpgrinder/internal/store"
	"github.com/ARJ2211/cpgrinder/tui/styles"
)

type BackToProjectMsg struct{}

type focusState int

const (
	focusMain focusState = iota
	focusDetail
)

const defaultVisibleRows = 25

type ProgressTrackerModel struct {
	dbStore *store.Store
	width   int
	height  int

	table             table.Model
	detailTable       table.Model
	noToIDMap         map[int]string
	detailProblemName string

	heatmapModel HeatMapModel

	focus focusState
}

func InitializeModel(dbStore *store.Store) (ProgressTrackerModel, error) {
	tbl, noToID, err := buildTable(dbStore)
	if err != nil {
		return ProgressTrackerModel{}, err
	}

	var model ProgressTrackerModel
	model.dbStore = dbStore
	model.table = tbl
	model.noToIDMap = noToID
	model.focus = focusMain

	row1, ok := noToID[1]
	if ok {
		dtlTbl, err := buildDetailTable(dbStore, row1)
		if err != nil {
			return ProgressTrackerModel{}, err
		}

		model.detailTable = dtlTbl
	}

	pName, err := dbStore.GetName(row1)
	model.detailProblemName = pName

	model.sizeTable()
	model.FocusMain()

	heatmap, err := InitializeHeatmapModel(dbStore)
	if err != nil {
		return ProgressTrackerModel{}, err
	}

	model.heatmapModel = heatmap

	return model, nil
}

func (m *ProgressTrackerModel) FocusMain() {
	m.focus = focusMain
	m.detailTable.Blur()
	m.table.Focus()
}

func (m *ProgressTrackerModel) FocusDetail() {
	m.focus = focusDetail
	m.table.Blur()
	m.detailTable.Focus()
}

func (m ProgressTrackerModel) Init() tea.Cmd { return nil }

func (m *ProgressTrackerModel) sizeTable() {
	desiredWidth := tableContentWidth(getTableColumns())
	desiredHeight := defaultVisibleRows

	if m.width == 0 || m.height == 0 {
		m.table.SetWidth(desiredWidth)
		m.table.SetHeight(desiredHeight)

		m.detailTable.SetWidth(desiredWidth)
		m.detailTable.SetHeight(desiredHeight)

		return
	}

	frameW, frameH := styles.TableStyle.GetFrameSize()

	availableWidth := max(20, m.width-frameW)
	availableHeight := max(5, m.height-frameH-2)

	m.table.SetWidth(min(desiredWidth, availableWidth))
	m.table.SetHeight(min(desiredHeight, availableHeight))

	m.detailTable.SetWidth(min(desiredWidth, availableWidth))
	m.detailTable.SetHeight(min(desiredHeight, availableHeight))
}

func (m ProgressTrackerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.focus {

	case focusMain:
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			m.sizeTable()
			return m, nil

		case tea.KeyPressMsg:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit

			case "esc":
				return m, func() tea.Msg { return BackToProjectMsg{} }

			case "r":
				updatedTable, noToID, err := buildTable(m.dbStore)
				if err != nil {
					return m, nil
				}

				m.table = updatedTable
				m.noToIDMap = noToID
				m.table.Focus()
				m.sizeTable()
				return m, nil

			case "enter", "space":
				selectedID := m.table.SelectedRow()
				if len(selectedID) == 0 {
					return m, nil
				}

				id, err := strconv.Atoi(selectedID[0])
				if err != nil {
					return m, nil
				}

				problemID, ok := m.noToIDMap[id]
				if !ok {
					return m, nil
				}

				pName, err := m.dbStore.GetName(problemID)
				if err != nil {
					return ProgressTrackerModel{}, nil
				}

				m.detailProblemName = pName

				dtlTbl, err := buildDetailTable(m.dbStore, problemID)
				if err != nil {
					return m, nil
				}

				m.detailTable = dtlTbl
				m.FocusDetail()

				return m, nil
			}
		}

		m.table, cmd = m.table.Update(msg)
		return m, cmd

	case focusDetail:
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			m.sizeTable()
			return m, nil

		case tea.KeyPressMsg:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit

			case "esc":
				m.FocusMain()
				return m, nil
			}
		}

		m.detailTable, cmd = m.detailTable.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m ProgressTrackerModel) View() tea.View {
	headingStyles := lipgloss.NewStyle().
		Bold(true).
		Blink(true).
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderStyle(lipgloss.ThickBorder())

	mainTableHeading := "LIST OF ALL ATTEMPTED PROBLEMS\n"
	detailTableHeading := fmt.Sprintf(
		"ALL ATTEMPTS FOR : %s\n", m.detailProblemName,
	)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		headingStyles.Render(mainTableHeading),
		styles.TableStyle.Render(m.table.View()),
		headingStyles.Render(detailTableHeading),
		styles.TableStyle.Render(m.detailTable.View()),
	)

	heatmap := m.heatmapModel.View()
	content = lipgloss.JoinHorizontal(
		lipgloss.Center,
		content,
		"                 ",
		heatmap.Content,
	)

	v := tea.NewView(content)
	v.WindowTitle = "Progress Tracker"
	return v
}
