package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var mediaExtensions = []string{
	".mp4", ".mkv", ".mov", ".avi", ".webm", ".flv", ".wmv", ".m4v", ".ts",
	".mp3", ".m4a", ".aac", ".opus", ".flac", ".wav", ".ogg", ".wma",
}

type FilePickerModel struct {
	cwd      string
	files    []string
	cursor   int
	selected []string
}

func NewFilePickerModel(startDir string) (FilePickerModel, error) {
	if startDir == "" {
		startDir = "."
	}
	abs, err := filepath.Abs(startDir)
	if err != nil {
		abs = startDir
	}
	m := FilePickerModel{
		cwd: abs,
	}
	if err := m.refreshFiles(); err != nil {
		return m, err
	}
	return m, nil
}

func (m *FilePickerModel) refreshFiles() error {
	entries, err := os.ReadDir(m.cwd)
	if err != nil {
		return err
	}

	var found []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if slices.Contains(mediaExtensions, ext) {
			found = append(found, filepath.Join(m.cwd, e.Name()))
		}
	}
	m.files = found
	m.cursor = 0
	return nil
}

func (m FilePickerModel) Init() tea.Cmd {
	return nil
}

func (m FilePickerModel) Update(msg tea.Msg) (FilePickerModel, tea.Cmd, []string) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.files) > 0 {
				return m, nil, []string{m.files[m.cursor]}
			}
		case "ctrl+c", "q", "esc":
			return m, tea.Quit, nil
		}
	}
	return m, nil, nil
}

func (m FilePickerModel) View() string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(12))
	subStyle := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(8))
	cursorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(10))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(7))

	b.WriteString(titleStyle.Render("📂 Select Media File"))
	b.WriteByte('\n')
	fmt.Fprintf(&b, "%s\n\n", subStyle.Render(fmt.Sprintf("Directory: %s", m.cwd)))

	if len(m.files) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(11)).Render("No media files found in current directory."))
		b.WriteString("\n")
		b.WriteString(subStyle.Render("(Press q or Esc to exit)"))
		return b.String()
	}

	for i, f := range m.files {
		name := filepath.Base(f)
		if i == m.cursor {
			fmt.Fprintf(&b, "%s\n", cursorStyle.Render(" > "+name))
		} else {
			fmt.Fprintf(&b, "%s\n", normalStyle.Render("   "+name))
		}
	}

	b.WriteString("\n")
	b.WriteString(subStyle.Render("(↑/↓ or j/k to navigate, Enter to select, q/Esc to exit)"))
	return b.String()
}
