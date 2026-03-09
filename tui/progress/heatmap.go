package progress

import (
	"math/rand"

	tea "charm.land/bubbletea/v2"
	"github.com/ARJ2211/cpgrinder/internal/store"
)

type HeatMapModel struct {
	width  int
	height int

	cellR int
	cellC int

	cellContents []string
}

func InitializeHeatmapModel(dbStore *store.Store) (HeatMapModel, error) {
	model := HeatMapModel{
		cellR:        7,
		cellC:        53,
		cellContents: []string{"▓", "▒", "░"},
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
	content := ""

	for i := 0; i < m.cellR; i++ {
		for j := 0; j < m.cellC; j++ {
			randomIndex := rand.Intn(len(m.cellContents))
			content += m.cellContents[randomIndex] + " "
		}
		content += "\n"
	}

	v := tea.NewView(content)
	return v
}
