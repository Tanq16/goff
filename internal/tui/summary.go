package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Tanq16/goff/internal/probe"
)

type SummaryModel struct {
	inputPath  string
	outputPath string
	duration   time.Duration
	origSize   int64
	outSize    int64
	err        error
}

func NewSummaryModel(inputPath string, outputPath string, duration time.Duration, origSize int64, err error) SummaryModel {
	var outSize int64
	if err == nil {
		if fi, statErr := os.Stat(outputPath); statErr == nil {
			outSize = fi.Size()
		}
	}
	return SummaryModel{
		inputPath:  inputPath,
		outputPath: outputPath,
		duration:   duration,
		origSize:   origSize,
		outSize:    outSize,
		err:        err,
	}
}

func (m SummaryModel) Init() tea.Cmd {
	return nil
}

func (m SummaryModel) Update(msg tea.Msg) (SummaryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter", "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m SummaryModel) View() string {
	boxWidth := defaultBoxWidth
	var lines []string

	if m.err != nil {
		lines = append(lines, renderBoxTop("Operation Failed", boxWidth))
		lines = append(lines, renderBoxEmpty(boxWidth))
		lines = append(lines, padBoxLine("  "+errorStyle.Render("✗ Processing Failed"), boxWidth))
		lines = append(lines, padBoxLine("    "+branchStyle.Render("└─ ")+normalItemStyle.Render(fmt.Sprintf("Error: %v", m.err)), boxWidth))
		lines = append(lines, renderBoxEmpty(boxWidth))
		lines = append(lines, renderBoxDivider(boxWidth))
		lines = append(lines, padBoxLine("  "+footerStyle.Render("Press Enter, Esc, or q to exit"), boxWidth))
		lines = append(lines, renderBoxBottom(boxWidth))
		return strings.Join(lines, "\n")
	}

	lines = append(lines, renderBoxTop("Operation Complete", boxWidth))
	lines = append(lines, renderBoxEmpty(boxWidth))
	lines = append(lines, padBoxLine("  "+successStyle.Render("✓ Operation Completed Successfully!"), boxWidth))
	lines = append(lines, renderBoxEmpty(boxWidth))

	lines = append(lines, padBoxLine("  "+labelStyle.Render("Output File: ")+valueStyle.Render(m.outputPath), boxWidth))
	lines = append(lines, padBoxLine("  "+labelStyle.Render("Time Taken:  ")+valueStyle.Render(m.duration.Round(time.Millisecond*100).String()), boxWidth))

	if m.origSize > 0 && m.outSize > 0 {
		origStr := probe.FormatBytes(m.origSize)
		outStr := probe.FormatBytes(m.outSize)
		lines = append(lines, padBoxLine("  "+labelStyle.Render("File Size:   ")+valueStyle.Render(fmt.Sprintf("%s → %s", origStr, outStr)), boxWidth))

		if m.outSize < m.origSize {
			savedBytes := m.origSize - m.outSize
			ratio := float64(savedBytes) / float64(m.origSize) * 100.0
			const savingsBarWidth = 20
			filled := int(ratio * float64(savingsBarWidth) / 100.0)
			if filled > savingsBarWidth {
				filled = savingsBarWidth
			}
			barStr := progressFillStyle.Render("●"+strings.Repeat("━", filled)) +
				progressEmptyStyle.Render(strings.Repeat(" ", savingsBarWidth-filled)) +
				progressFillStyle.Render("●")
			lines = append(lines, padBoxLine("  "+labelStyle.Render("Savings:     ")+barStr+"  "+percentStyle.Render(fmt.Sprintf("%.1f%%", ratio))+footerStyle.Render(fmt.Sprintf(" (saved %s)", probe.FormatBytes(savedBytes))), boxWidth))
		} else if m.outSize > m.origSize {
			diff := m.outSize - m.origSize
			lines = append(lines, padBoxLine("  "+labelStyle.Render("Size Delta:  ")+valueStyle.Render(fmt.Sprintf("+%s", probe.FormatBytes(diff))), boxWidth))
		}
	}

	lines = append(lines, renderBoxEmpty(boxWidth))
	lines = append(lines, renderBoxDivider(boxWidth))
	lines = append(lines, padBoxLine("  "+footerStyle.Render("Press Enter, Esc, or q to exit"), boxWidth))
	lines = append(lines, renderBoxBottom(boxWidth))

	return strings.Join(lines, "\n")
}
