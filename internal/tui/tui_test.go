package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/Tanq16/goff/internal/probe"
)

func TestThemeBoxFraming(t *testing.T) {
	top := renderBoxTop("Test Title", 60)
	if !strings.Contains(top, "Test Title") || !strings.HasPrefix(top, "\x1b[") && !strings.Contains(top, "┌") {
		t.Errorf("renderBoxTop malformed: %s", top)
	}

	bot := renderBoxBottom(60)
	if !strings.Contains(bot, "└") || !strings.Contains(bot, "┘") {
		t.Errorf("renderBoxBottom malformed: %s", bot)
	}

	div := renderBoxDivider(60)
	if !strings.Contains(div, "│") || !strings.Contains(div, "─") {
		t.Errorf("renderBoxDivider malformed: %s", div)
	}
}

func TestTUIViewsRendering(t *testing.T) {
	p := &probe.ProbeResult{
		Format: probe.FormatInfo{
			Filename:       "sample.mp4",
			FormatLongName: "QuickTime / MOV",
			DurationStr:    "60.0",
			SizeStr:        "10485760",
		},
		Streams: []probe.StreamInfo{
			{
				CodecType: "video",
				Width:     1920,
				Height:    1080,
			},
		},
	}

	menu := NewMenuModel(p, 1)
	menuView := menu.View()
	if !strings.Contains(menuView, "Media Suite: sample") {
		t.Errorf("MenuModel.View() missing title: %s", menuView)
	}

	options := NewOptionsModel("video_opt", "sample.mp4", p)
	optionsView := options.View()
	if !strings.Contains(optionsView, "Configure: Video Optimization") {
		t.Errorf("OptionsModel.View() missing title: %s", optionsView)
	}

	prog := NewProgressViewModel("sample.mp4", "sample.optimized.mp4", 60.0)
	progView := prog.View()
	if !strings.Contains(progView, "Processing Media") {
		t.Errorf("ProgressViewModel.View() missing title: %s", progView)
	}

	sum := NewSummaryModel("sample.mp4", "sample.optimized.mp4", 2*time.Second, 10485760, nil)
	sumView := sum.View()
	if !strings.Contains(sumView, "Operation Complete") {
		t.Errorf("SummaryModel.View() missing title: %s", sumView)
	}
}
