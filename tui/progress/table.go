package progress

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"github.com/ARJ2211/cpgrinder/internal/store"
)

/*
Get all the columns required.
*/
func getTableColumns() []table.Column {
	return []table.Column{
		{Title: "No", Width: 4},
		{Title: "Name", Width: 28},
		{Title: "Started At", Width: 20},
		{Title: "Finished At", Width: 20},
		{Title: "Created At", Width: 20},
		{Title: "Verdict", Width: 10},
		{Title: "Language", Width: 12},
		{Title: "Time Spent", Width: 12},
	}
}

func tableContentWidth(cols []table.Column) int {
	total := 0
	for _, col := range cols {
		total += col.Width
	}

	return total + len(cols) + 2
}

/*
Function to build the table and its rows.
*/
func buildTable(db *store.Store) (table.Model, error) {
	cols := getTableColumns()

	attemptsData, err := db.ListAllAttempts()
	if err != nil {
		return table.Model{}, err
	}

	rows := []table.Row{}

	for i, attempt := range attemptsData {
		problem, err := db.GetProblemByID(attempt.ProblemID)
		if err != nil {
			return table.Model{}, err
		}

		startedAt := "-"
		if attempt.StartedAt != nil {
			startedAt = time.Unix(*attempt.StartedAt, 0).Format("02 Jan 2006 03:04 PM")
		}

		finishedAt := "-"
		if attempt.FinishedAt != nil {
			finishedAt = time.Unix(*attempt.FinishedAt, 0).Format("02 Jan 2006 03:04 PM")
		}

		createdAt := time.Unix(attempt.CreatedAt, 0).Format("02 Jan 2006 03:04 PM")

		row := table.Row{
			fmt.Sprintf("%d", i+1),
			problem.Title,
			startedAt,
			finishedAt,
			createdAt,
			attempt.Verdict,
			attempt.Language,
			fmt.Sprintf("%d", attempt.TimeSpentSeconds),
		}

		rows = append(rows, row)
	}

	tableModel := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	tableModel.SetStyles(s)
	tableModel.SetWidth(tableContentWidth(cols))
	tableModel.SetHeight(defaultVisibleRows)

	return tableModel, nil
}
