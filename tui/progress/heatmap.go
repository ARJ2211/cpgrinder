package progress

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ARJ2211/cpgrinder/internal/store"
)

type HeatMapCell struct {
	Date     time.Time
	Count    int
	IsFuture bool
	Symbol   string
}

type HeatMapModel struct {
	width  int
	height int

	cellR int
	cellC int

	grid     [][]HeatMapCell // [53][7]
	maxCount int
}

func symbolForHeatmap(count int, isFuture bool, maxCount int) string {
	if isFuture {
		return "•"
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
	const weeks = 53

	rawGrid, maxCount, err := dbStore.GetAttemptHeatmapData(weeks)
	if err != nil {
		return HeatMapModel{}, err
	}

	model := HeatMapModel{
		cellR:    7,
		cellC:    weeks,
		maxCount: maxCount,
		grid:     make([][]HeatMapCell, weeks),
	}

	for col := 0; col < weeks; col++ {
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

func (m HeatMapModel) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m HeatMapModel) View() tea.View {
	var b strings.Builder

	for row := 0; row < m.cellR; row++ {
		for col := 0; col < m.cellC; col++ {
			b.WriteString(m.grid[col][row].Symbol)
			b.WriteString(" ")
		}
		b.WriteString("\n")
	}

	return tea.NewView(b.String())
}
