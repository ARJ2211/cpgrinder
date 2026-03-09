package styles

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/glamour"
)

// Style sheet for the legend
var LegendStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62")).
	Padding(0, 0).
	MarginTop(1).
	Foreground(lipgloss.Color("248"))

// Style sheet for the markdowns
var GlamourMD, _ = glamour.NewTermRenderer(
	glamour.WithAutoStyle(),
	glamour.WithWordWrap(100),
	glamour.WithChromaFormatter(">"),
)

// Styles for table
var TableStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var (
	cpOrange     = lipgloss.Color("#FB923C")
	grinderBlue  = lipgloss.Color("#38BDF8")
	accentCyan   = lipgloss.Color("#67E8F9")
	textMain     = lipgloss.Color("#E2E8F0")
	textMuted    = lipgloss.Color("#94A3B8")
	textDim      = lipgloss.Color("#64748B")
	borderStrong = lipgloss.Color("#1E293B")
	borderSoft   = lipgloss.Color("#334155")
	bgPanel      = lipgloss.Color("#0B1220")
	bgPanelAlt   = lipgloss.Color("#0F172A")

	AppStyle = lipgloss.NewStyle().
			Padding(2, 2)

	HeroCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderStrong).
			Background(bgPanelAlt).
			Padding(1, 2)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderSoft).
			Background(bgPanel).
			Padding(1, 2)

	InfoCardStyle = CardStyle.Copy()

	KickerStyle = lipgloss.NewStyle().
			Foreground(accentCyan).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(textMuted).
			MarginTop(1)

	SectionTitleStyle = lipgloss.NewStyle().
				Foreground(textMain).
				Bold(true).
				MarginBottom(1)

	MenuItemStyle = lipgloss.NewStyle().
			Foreground(textMain).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(borderStrong).
			Padding(0, 1).
			MarginBottom(1)

	SelectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(textMain).
				Background(bgPanelAlt).
				Border(lipgloss.ThickBorder(), false, false, false, true).
				BorderForeground(grinderBlue).
				Bold(true).
				Padding(0, 1).
				MarginBottom(1)

	MenuLabelStyle = lipgloss.NewStyle().
			Foreground(textMain).
			Bold(true)

	MenuDescStyle = lipgloss.NewStyle().
			Foreground(textDim)

	KeyStyle = lipgloss.NewStyle().
			Foreground(textMain).
			Background(bgPanelAlt).
			Padding(0, 1)

	MutedStyle = lipgloss.NewStyle().
			Foreground(textMuted)

	HelpStyle = lipgloss.NewStyle().
			Foreground(textDim).
			MarginTop(1)

	NotImplStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(cpOrange).
			Foreground(textMain).
			Background(bgPanel).
			Padding(1, 2)
)

func RenderBanner() string {
	leftArt := strings.Trim(`
  /$$$$$$  /$$$$$$$
 /$$__  $$| $$__  $$
| $$  \__/| $$  \ $$
| $$      | $$$$$$$/
| $$      | $$____/
| $$    $$| $$
|  $$$$$$/| $$
 \______/ |__/
`, "\n")

	rightArt := strings.Trim(`
  /$$$$$$           /$$                 /$$
 /$$__  $$         |__/                | $$
| $$  \__/ /$$$$$$  /$$ /$$$$$$$   /$$$$$$$  /$$$$$$   /$$$$$$
| $$ /$$$$/$$__  $$| $$| $$__  $$ /$$__  $$ /$$__  $$ /$$__  $$
| $$|_  $$| $$  \__/| $$| $$  \ $$| $$  | $$| $$$$$$$$| $$  \__/
| $$  \ $$| $$      | $$| $$  | $$| $$  | $$| $$_____/| $$
|  $$$$$$/| $$      | $$| $$  | $$|  $$$$$$$|  $$$$$$$| $$
 \______/ |__/      |__/|__/  |__/ \_______/ \_______/|__/
`, "\n")

	leftStyle := lipgloss.NewStyle().
		Foreground(cpOrange).
		Bold(true)

	rightStyle := lipgloss.NewStyle().
		Foreground(grinderBlue).
		Bold(true)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyle.Render(leftArt),
		"  ",
		rightStyle.Render(rightArt),
	)
}
