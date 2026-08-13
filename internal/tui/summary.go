package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	var b strings.Builder

	if m.err != nil {
		errStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(9))
		b.WriteString(errStyle.Render("✗ Processing Failed"))
		b.WriteString("\n\n")
		fmt.Fprintf(&b, "Error: %v\n\n", m.err)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(8)).Render("(Press Enter or q to exit)"))
		return b.String()
	}

	successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(10))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(12))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(15))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(8))

	b.WriteString(successStyle.Render("✓ Operation Completed Successfully!"))
	b.WriteString("\n\n")

	fmt.Fprintf(&b, "  %s %s\n", labelStyle.Render("Output File:"), valueStyle.Render(m.outputPath))
	fmt.Fprintf(&b, "  %s %s\n", labelStyle.Render("Time Taken: "), valueStyle.Render(m.duration.Round(time.Millisecond*100).String()))

	if m.origSize > 0 && m.outSize > 0 {
		origStr := probe.FormatBytes(m.origSize)
		outStr := probe.FormatBytes(m.outSize)
		fmt.Fprintf(&b, "  %s %s → %s\n", labelStyle.Render("File Size:  "), origStr, valueStyle.Render(outStr))

		if m.outSize < m.origSize {
			savedBytes := m.origSize - m.outSize
			ratio := float64(savedBytes) / float64(m.origSize) * 100.0
			savingsStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(10))
			fmt.Fprintf(&b, "  %s %s (saved %s)\n", labelStyle.Render("Savings:    "), savingsStyle.Render(fmt.Sprintf("%.1f%%", ratio)), probe.FormatBytes(savedBytes))
		} else if m.outSize > m.origSize {
			diff := m.outSize - m.origSize
			fmt.Fprintf(&b, "  %s +%s\n", labelStyle.Render("Size Delta: "), probe.FormatBytes(diff))
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("(Press Enter, q, or Esc to exit)"))
	return b.String()
}
