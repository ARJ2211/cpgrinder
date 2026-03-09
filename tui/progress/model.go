package progress

import (
	"strconv"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/ARJ2211/cpgrinder/internal/store"
	"github.com/ARJ2211/cpgrinder/tui/styles"
)

type BackToProjectMsg struct{}

const defaultVisibleRows = 25

type ProgressTrackerModel struct {
	dbStore *store.Store
	width   int
	height  int

	table       table.Model    // This is to show the main table
	detailTable table.Model    // This is to show the detail table
	noToIDMap   map[int]string // This is to map the row num to the problem ID
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

	row1, ok := noToID[1]
	if ok {
		dtlTbl, err := buildDetailTable(dbStore, row1)
		if err != nil {
			return ProgressTrackerModel{}, err
		}

		model.detailTable = dtlTbl
	}

	model.sizeTable()

	return model, nil
}

func (m ProgressTrackerModel) Init() tea.Cmd { return nil }

func (m *ProgressTrackerModel) sizeTable() {
	desiredWidth := tableContentWidth(getTableColumns())
	desiredHeight := defaultVisibleRows

	// Before the first WindowSizeMsg arrives, keep the table compact.
	if m.width == 0 || m.height == 0 {
		m.table.SetWidth(desiredWidth)
		m.table.SetHeight(desiredHeight)
		return
	}

	frameW, frameH := styles.TableStyle.GetFrameSize()

	availableWidth := max(20, m.width-frameW)
	availableHeight := max(5, m.height-frameH-2)

	m.table.SetWidth(min(desiredWidth, availableWidth))
	m.table.SetHeight(min(desiredHeight, availableHeight))
}

func (m ProgressTrackerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.sizeTable()

	case tea.KeyPressMsg:
		switch msg.String() {

		// Close the program globally
		case "q", "ctrl+c":
			return m, tea.Quit

		// Go to the previous screen
		case "esc":
			return m, func() tea.Msg { return BackToProjectMsg{} }

		// Refresh the rows in the table (refetch)
		case "r":
			updatedTable, noToID, err := buildTable(m.dbStore)
			if err != nil {
				return ProgressTrackerModel{}, nil
			}

			m.table = updatedTable
			m.noToIDMap = noToID
			return m, nil

		// When something is selected using enter or space
		case "enter", "space":
			selectedID := m.table.SelectedRow()

			id, err := strconv.Atoi(selectedID[0])
			if err != nil {
				return ProgressTrackerModel{}, nil
			}

			problemID := m.noToIDMap[id]

			// TODO: Initialize the problem detail table here
			dtlTbl, err := buildDetailTable(m.dbStore, problemID)
			if err != nil {
				return ProgressTrackerModel{}, nil
			}

			// Pass updated to the new table now
			updated, cmd := dtlTbl.Update(msg)
			m.detailTable = updated

			return m, cmd
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m ProgressTrackerModel) View() tea.View {
	v := tea.NewView(
		styles.TableStyle.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n" + styles.TableStyle.Render(m.detailTable.View()),
	)
	v.WindowTitle = "Progress Tracker"
	return v
}
