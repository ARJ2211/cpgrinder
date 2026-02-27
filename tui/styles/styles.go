package styles

import (
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/glamour"
)

// Style sheet for the legend
var LegendStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62")). // Lavender/Purple
	Padding(0, 0).
	MarginTop(1).
	Foreground(lipgloss.Color("248"))

// Style sheet for the markdowns
var GlamourMD, _ = glamour.NewTermRenderer(
	glamour.WithAutoStyle(),
	glamour.WithWordWrap(100),
	glamour.WithChromaFormatter(">"),
)
