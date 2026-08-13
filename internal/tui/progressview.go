package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Tanq16/goff/internal/engine"
	"github.com/Tanq16/goff/internal/probe"
)

type ProgressMsg engine.ProgressUpdate
type ExecutionFinishedMsg struct {
	Err        error
	OutputPath string
}

type ProgressViewModel struct {
	filename   string
	outputPath string
	progress   engine.ProgressUpdate
	startTime  time.Time
}

func NewProgressViewModel(filename string, outputPath string, totalSec float64) ProgressViewModel {
	return ProgressViewModel{
		filename:   filename,
		outputPath: outputPath,
		progress: engine.ProgressUpdate{
			TotalSeconds: totalSec,
		},
		startTime: time.Now(),
	}
}

func (m ProgressViewModel) Init() tea.Cmd {
	return nil
}

func (m ProgressViewModel) Update(msg tea.Msg) (ProgressViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ProgressMsg:
		m.progress = engine.ProgressUpdate(msg)
	}
	return m, nil
}

func (m ProgressViewModel) View() string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(12))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(10))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(8))

	fmt.Fprintf(&b, "%s\n", titleStyle.Render(fmt.Sprintf("⚡ Processing: %s", m.filename)))
	fmt.Fprintf(&b, "%s\n\n", dimStyle.Render(fmt.Sprintf("Output → %s", m.outputPath)))

	pct := m.progress.Percent
	if pct > 100 {
		pct = 100
	}

	const barWidth = 30
	filled := barWidth * pct / 100
	empty := barWidth - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	speed := m.progress.Speed
	if speed == "" {
		speed = "1.0x"
	}

	fpsStr := ""
	if m.progress.FPS > 0 {
		fpsStr = fmt.Sprintf(" | %.1f fps", m.progress.FPS)
	}

	sizeStr := ""
	if m.progress.OutBytes > 0 {
		sizeStr = fmt.Sprintf(" | %s", probe.FormatBytes(m.progress.OutBytes))
	}

	curTime := probe.FormatDuration(m.progress.CurrentSeconds)
	totTime := probe.FormatDuration(m.progress.TotalSeconds)

	fmt.Fprintf(&b, "%s\n\n", infoStyle.Render(fmt.Sprintf(" [%s] %3d%%", bar, pct)))
	fmt.Fprintf(&b, "  Time: %s / %s | Speed: %s%s%s\n\n", curTime, totTime, speed, fpsStr, sizeStr)
	b.WriteString(dimStyle.Render("(Encoding in background. Press Ctrl+C to abort)"))

	return b.String()
}
