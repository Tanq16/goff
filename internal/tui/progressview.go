package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
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
	boxWidth := defaultBoxWidth
	var lines []string

	lines = append(lines, renderBoxTop("Processing Media", boxWidth))
	lines = append(lines, renderBoxEmpty(boxWidth))

	pct := m.progress.Percent
	if pct > 100 {
		pct = 100
	}

	const progressWidth = 34
	filled := progressWidth * pct / 100
	empty := progressWidth - filled

	barStr := progressFillStyle.Render("●"+strings.Repeat("━", filled)) +
		progressEmptyStyle.Render(strings.Repeat(" ", empty)) +
		progressFillStyle.Render("●")

	speed := m.progress.Speed
	if speed == "" {
		speed = "1.0x"
	}

	fpsStr := ""
	if m.progress.FPS > 0 {
		fpsStr = fmt.Sprintf("  •  %.1f fps", m.progress.FPS)
	}

	sizeStr := ""
	if m.progress.OutBytes > 0 {
		sizeStr = fmt.Sprintf("  •  %s", probe.FormatBytes(m.progress.OutBytes))
	}

	curTime := probe.FormatDuration(m.progress.CurrentSeconds)
	totTime := probe.FormatDuration(m.progress.TotalSeconds)

	lines = append(lines, padBoxLine("  "+activeBulletStyle.Render("● ")+activeItemStyle.Render(m.filename)+footerStyle.Render("  [Encoding]"), boxWidth))
	lines = append(lines, padBoxLine("    "+barStr+"  "+percentStyle.Render(fmt.Sprintf("%3d%%", pct))+"  "+footerStyle.Render(fmt.Sprintf("(%s / %s)", curTime, totTime)), boxWidth))
	lines = append(lines, padBoxLine("    "+branchStyle.Render("└─ ")+footerStyle.Render(fmt.Sprintf("Speed: %s%s%s", speed, fpsStr, sizeStr)), boxWidth))
	lines = append(lines, padBoxLine("    "+branchStyle.Render("└─ ")+footerStyle.Render("Output → "+filepath.Base(m.outputPath)), boxWidth))

	lines = append(lines, renderBoxEmpty(boxWidth))
	lines = append(lines, renderBoxDivider(boxWidth))
	lines = append(lines, padBoxLine("  "+footerStyle.Render("Encoding in background  •  Press Ctrl+C or q to abort"), boxWidth))
	lines = append(lines, renderBoxBottom(boxWidth))

	return strings.Join(lines, "\n")
}
