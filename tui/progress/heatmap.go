package progress

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	zone "github.com/lrstanley/bubblezone/v2"

	"github.com/ARJ2211/cpgrinder/internal/store"
)

type HeatMapCell struct {
	Date     time.Time
	Count    int
	IsFuture bool
	Symbol   string
}

type HeatMapModel struct {
	cellR int
	cellC int

	width  int
	height int

	grid     [][]HeatMapCell // [week][day]
	maxCount int

	status string
}

const (
	heatmapWeeks   = 53
	leftLabelWidth = 5
	cellWidth      = 2 // one symbol + one trailing space
)

func symbolForHeatmap(count int, isFuture bool, maxCount int) string {
	if isFuture {
		return "·"
	}
	if count <= 0 {
		return "·"
	}

	if maxCount <= 3 {
		switch count {
		case 1:
			return "░"
		case 2:
			return "▒"
		default:
			return "▓"
		}
	}

	ratio := float64(count) / float64(maxCount)

	switch {
	case ratio <= 0.34:
		return "░"
	case ratio <= 0.67:
		return "▒"
	default:
		return "▓"
	}
}

func InitializeHeatmapModel(dbStore *store.Store) (HeatMapModel, error) {
	rawGrid, maxCount, err := dbStore.GetAttemptHeatmapData(heatmapWeeks)
	if err != nil {
		return HeatMapModel{}, err
	}

	model := HeatMapModel{
		cellR:    7,
		cellC:    heatmapWeeks,
		maxCount: maxCount,
		grid:     make([][]HeatMapCell, heatmapWeeks),
		status:   "Click a day to inspect activity",
	}

	for col := 0; col < heatmapWeeks; col++ {
		model.grid[col] = make([]HeatMapCell, 7)
		for row := 0; row < 7; row++ {
			day := rawGrid[col][row]
			model.grid[col][row] = HeatMapCell{
				Date:     day.Date,
				Count:    day.Count,
				IsFuture: day.IsFuture,
				Symbol:   symbolForHeatmap(day.Count, day.IsFuture, maxCount),
			}
		}
	}

	return model, nil
}

func (m HeatMapModel) Init() tea.Cmd {
	return nil
}

func (m HeatMapModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft {
			return m, nil
		}

		for col := 0; col < m.cellC; col++ {
			for row := 0; row < m.cellR; row++ {
				if zone.Get(m.zoneID(col, row)).InBounds(msg) {
					cell := m.grid[col][row]

					if cell.IsFuture {
						m.status = fmt.Sprintf("Future date: %s", cell.Date.Format("2006-01-02"))
					} else {
						m.status = fmt.Sprintf(
							"%d attempt%s on %s",
							cell.Count,
							plural(cell.Count),
							cell.Date.Format("2006-01-02"),
						)
					}

					return m, nil
				}
			}
		}
	}

	return m, nil
}

func (m HeatMapModel) Render() string {
	var b strings.Builder

	b.WriteString(m.renderMonthHeader())
	b.WriteString("\n")

	for row := 0; row < m.cellR; row++ {
		b.WriteString(m.renderRow(row))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.status)
	b.WriteString("\n")

	return b.String()
}

func (m HeatMapModel) View() tea.View {
	return tea.NewView(m.Render())
}

func (m HeatMapModel) zoneID(col, row int) string {
	return fmt.Sprintf("heatmap-%d-%d", col, row)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func (m HeatMapModel) dayLabel(row int) string {
	switch row {
	case 1:
		return "Mon"
	case 3:
		return "Wed"
	case 5:
		return "Fri"
	default:
		return ""
	}
}

func (m HeatMapModel) monthStarts() map[int]string {
	out := make(map[int]string)

	if len(m.grid) == 0 || len(m.grid[0]) == 0 {
		return out
	}

	out[0] = m.grid[0][0].Date.Format("Jan")

	for col := 0; col < m.cellC; col++ {
		for row := 0; row < m.cellR; row++ {
			d := m.grid[col][row].Date
			if d.Day() == 1 {
				out[col] = d.Format("Jan")
				break
			}
		}
	}

	return out
}

func (m HeatMapModel) renderMonthHeader() string {
	totalGridWidth := m.cellC * cellWidth
	runes := []rune(strings.Repeat(" ", totalGridWidth))

	monthCols := m.monthStarts()
	lastWrittenEnd := -999

	for col := 0; col < m.cellC; col++ {
		label, ok := monthCols[col]
		if !ok {
			continue
		}

		pos := col * cellWidth
		if pos <= lastWrittenEnd {
			continue
		}

		labelRunes := []rune(label)
		for i, r := range labelRunes {
			if pos+i >= len(runes) {
				break
			}
			runes[pos+i] = r
		}

		lastWrittenEnd = pos + len(labelRunes) - 1
	}

	return fmt.Sprintf("%-*s%s", leftLabelWidth, "", string(runes))
}

func (m HeatMapModel) renderRow(row int) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("%-*s", leftLabelWidth, m.dayLabel(row)))

	for col := 0; col < m.cellC; col++ {
		cell := m.grid[col][row]
		renderedCell := zone.Mark(m.zoneID(col, row), cell.Symbol+" ")
		b.WriteString(renderedCell)
	}

	return b.String()
}
