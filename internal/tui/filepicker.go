package tui

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
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
	boxWidth := defaultBoxWidth
	var lines []string

	lines = append(lines, renderBoxTop("Select Media File", boxWidth))
	lines = append(lines, renderBoxEmpty(boxWidth))
	lines = append(lines, padBoxLine("  "+labelStyle.Render("Directory: ")+valueStyle.Render(m.cwd), boxWidth))
	lines = append(lines, renderBoxEmpty(boxWidth))
	lines = append(lines, renderBoxDivider(boxWidth))
	lines = append(lines, renderBoxEmpty(boxWidth))

	if len(m.files) == 0 {
		lines = append(lines, padBoxLine("  "+normalBulletStyle.Render("○ ")+normalItemStyle.Render("No media files found in current directory"), boxWidth))
	} else {
		for i, f := range m.files {
			name := filepath.Base(f)
			if i == m.cursor {
				lines = append(lines, padBoxLine("  "+activeBulletStyle.Render("● ")+activeItemStyle.Render(name), boxWidth))
			} else {
				lines = append(lines, padBoxLine("  "+normalBulletStyle.Render("○ ")+normalItemStyle.Render(name), boxWidth))
			}
		}
	}

	lines = append(lines, renderBoxEmpty(boxWidth))
	lines = append(lines, renderBoxDivider(boxWidth))
	lines = append(lines, padBoxLine("  "+footerStyle.Render("↑/↓ or j/k to navigate  •  Enter to select  •  q to exit"), boxWidth))
	lines = append(lines, renderBoxBottom(boxWidth))

	return strings.Join(lines, "\n")
}
