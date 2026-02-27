package styles

import "charm.land/lipgloss/v2"

var LegendStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62")). // Lavender/Purple
	Padding(0, 0).
	MarginTop(1).
	Foreground(lipgloss.Color("248"))
