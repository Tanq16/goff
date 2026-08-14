package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	colorBlue    = lipgloss.ANSIColor(12)
	colorGreen   = lipgloss.ANSIColor(10)
	colorRed     = lipgloss.ANSIColor(9)
	colorMagenta = lipgloss.ANSIColor(13)
	colorCyan    = lipgloss.ANSIColor(14)
	colorFg      = lipgloss.ANSIColor(15)
	colorMuted   = lipgloss.ANSIColor(7)
	colorChrome  = lipgloss.ANSIColor(8)

	borderStyle        = lipgloss.NewStyle().Foreground(colorChrome)
	boxTitleStyle      = lipgloss.NewStyle().Foreground(colorBlue).Bold(true)
	activeItemStyle    = lipgloss.NewStyle().Foreground(colorFg).Bold(true)
	activeBulletStyle  = lipgloss.NewStyle().Foreground(colorGreen)
	normalItemStyle    = lipgloss.NewStyle().Foreground(colorMuted)
	normalBulletStyle  = lipgloss.NewStyle().Foreground(colorChrome)
	branchStyle        = lipgloss.NewStyle().Foreground(colorMagenta)
	descStyle          = lipgloss.NewStyle().Foreground(colorChrome)
	progressFillStyle  = lipgloss.NewStyle().Foreground(colorBlue)
	progressEmptyStyle = lipgloss.NewStyle().Foreground(colorChrome)
	percentStyle       = lipgloss.NewStyle().Foreground(colorCyan)
	footerStyle        = lipgloss.NewStyle().Foreground(colorChrome)
	successStyle       = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	errorStyle         = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	labelStyle         = lipgloss.NewStyle().Foreground(colorBlue)
	valueStyle         = lipgloss.NewStyle().Foreground(colorFg)
)

const defaultBoxWidth = 74

func renderBoxTop(title string, boxWidth int) string {
	titleText := " " + title + " "
	leftPad := 2
	rightPad := boxWidth - leftPad - len(titleText) - 2
	if rightPad < 0 {
		rightPad = 0
	}
	return borderStyle.Render("┌"+strings.Repeat("─", leftPad)) +
		boxTitleStyle.Render(titleText) +
		borderStyle.Render(strings.Repeat("─", rightPad)+"┐")
}

func renderBoxBottom(boxWidth int) string {
	return borderStyle.Render("└" + strings.Repeat("─", boxWidth-2) + "┘")
}

func renderBoxDivider(boxWidth int) string {
	innerWidth := boxWidth - 4
	return borderStyle.Render("│  " + strings.Repeat("─", innerWidth-2) + "  │")
}

func renderBoxEmpty(boxWidth int) string {
	return borderStyle.Render("│") + strings.Repeat(" ", boxWidth-2) + borderStyle.Render("│")
}

func padBoxLine(content string, boxWidth int) string {
	visibleLen := lipgloss.Width(content)
	padding := boxWidth - visibleLen - 2
	if padding < 0 {
		padding = 0
	}
	return borderStyle.Render("│") + content + strings.Repeat(" ", padding) + borderStyle.Render("│")
}
