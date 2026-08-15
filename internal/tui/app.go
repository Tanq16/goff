package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Tanq16/goff/internal/engine"
	"github.com/Tanq16/goff/internal/probe"
)

type AppState int

const (
	StateFilePicker AppState = iota
	StateActionMenu
	StateOptionsMenu
	StateExecuting
	StateSummary
)

type AppModel struct {
	state          AppState
	files          []string
	activeFile     string
	probeResult    *probe.ProbeResult
	filePicker     FilePickerModel
	menu           MenuModel
	options        OptionsModel
	progressView   ProgressViewModel
	summary        SummaryModel
	progressChan   chan ProgressMsg
	finishChan     chan ExecutionFinishedMsg
	cancelFunc     context.CancelFunc
	startTime      time.Time
	outputTarget   string
	allowOverwrite bool
}

func NewAppModel(files []string, allowOverwrite bool) (AppModel, error) {
	app := AppModel{
		files:          files,
		allowOverwrite: allowOverwrite,
	}

	if len(files) == 0 {
		fp, err := NewFilePickerModel(".")
		if err != nil {
			return app, err
		}
		app.filePicker = fp
		app.state = StateFilePicker
		return app, nil
	}

	if len(files) == 1 {
		app.activeFile = files[0]
		p, err := probe.RunProbe(context.Background(), files[0])
		if err != nil {
			return app, fmt.Errorf("failed to probe %q: %w", files[0], err)
		}
		app.probeResult = p
		app.menu = NewMenuModel(p, 1)
		app.state = StateActionMenu
		return app, nil
	}

	app.menu = NewMenuModel(nil, len(files))
	app.state = StateActionMenu
	return app, nil
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func waitForProgress(ch <-chan ProgressMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func waitForFinish(ch <-chan ExecutionFinishedMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "esc" {
			switch m.state {
			case StateOptionsMenu:
				m.state = StateActionMenu
				return m, nil
			case StateActionMenu:
				if len(m.files) == 0 {
					m.state = StateFilePicker
					return m, nil
				}
				return m, tea.Quit
			}
		}

	case ProgressMsg:
		var cmd tea.Cmd
		m.progressView, cmd = m.progressView.Update(msg)
		return m, tea.Batch(cmd, waitForProgress(m.progressChan))

	case ExecutionFinishedMsg:
		elapsed := time.Since(m.startTime)
		var origSize int64
		if m.probeResult != nil {
			origSize = m.probeResult.Format.Size()
		}
		m.summary = NewSummaryModel(m.activeFile, msg.OutputPath, elapsed, origSize, msg.Err)
		m.state = StateSummary
		return m, nil
	}

	switch m.state {
	case StateFilePicker:
		var picked []string
		var cmd tea.Cmd
		m.filePicker, cmd, picked = m.filePicker.Update(msg)
		if len(picked) > 0 {
			m.files = picked
			m.activeFile = picked[0]
			p, err := probe.RunProbe(context.Background(), m.activeFile)
			if err != nil {
				m.summary = NewSummaryModel(m.activeFile, "", 0, 0, err)
				m.state = StateSummary
				return m, nil
			}
			m.probeResult = p
			m.menu = NewMenuModel(p, 1)
			m.state = StateActionMenu
		}
		return m, cmd

	case StateActionMenu:
		var item *ActionItem
		var cmd tea.Cmd
		m.menu, cmd, item = m.menu.Update(msg)
		if item != nil {
			if item.ID == "inspect" {
				return m, tea.Quit
			}
			m.options = NewOptionsModel(item.ID, m.activeFile, m.probeResult, m.files)
			m.state = StateOptionsMenu
		}
		return m, cmd

	case StateOptionsMenu:
		var choice *OptionChoice
		var cmd tea.Cmd
		m.options, cmd, choice = m.options.Update(msg)
		if choice != nil {
			res, err := choice.Build(m.activeFile, m.probeResult)
			if err != nil {
				m.summary = NewSummaryModel(m.activeFile, "", 0, 0, err)
				m.state = StateSummary
				return m, nil
			}

			outPath, err := engine.ResolveOutputName(m.activeFile, res.Suffix, res.TargetExt, "", m.allowOverwrite)
			if err != nil {
				m.summary = NewSummaryModel(m.activeFile, "", 0, 0, err)
				m.state = StateSummary
				return m, nil
			}

			fullArgs := append(res.Args, outPath)
			totalSec := 0.0
			if m.probeResult != nil {
				totalSec = m.probeResult.TotalDuration()
			}

			m.progressView = NewProgressViewModel(filepath.Base(m.activeFile), outPath, totalSec)
			m.outputTarget = outPath
			m.progressChan = make(chan ProgressMsg, 50)
			m.finishChan = make(chan ExecutionFinishedMsg, 1)
			m.startTime = time.Now()
			m.state = StateExecuting

			ctx, cancel := context.WithCancel(context.Background())
			m.cancelFunc = cancel

			go func() {
				defer close(m.progressChan)
				defer close(m.finishChan)

				err := engine.RunFFmpeg(ctx, fullArgs, totalSec, func(p engine.ProgressUpdate) {
					select {
					case m.progressChan <- ProgressMsg(p):
					default:
					}
				})

				m.finishChan <- ExecutionFinishedMsg{
					Err:        err,
					OutputPath: outPath,
				}
			}()

			return m, tea.Batch(cmd, waitForProgress(m.progressChan), waitForFinish(m.finishChan))
		}
		return m, cmd

	case StateExecuting:
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			if msg.String() == "ctrl+c" || msg.String() == "q" {
				if m.cancelFunc != nil {
					m.cancelFunc()
				}
				return m, tea.Quit
			}
		}

	case StateSummary:
		var cmd tea.Cmd
		m.summary, cmd = m.summary.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m AppModel) View() tea.View {
	var v tea.View
	switch m.state {
	case StateFilePicker:
		v = tea.NewView(m.filePicker.View())
	case StateActionMenu:
		v = tea.NewView(m.menu.View())
	case StateOptionsMenu:
		v = tea.NewView(m.options.View())
	case StateExecuting:
		v = tea.NewView(m.progressView.View())
	case StateSummary:
		v = tea.NewView(m.summary.View())
	default:
		v = tea.NewView("")
	}
	v.AltScreen = true
	return v
}
