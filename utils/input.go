package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var stdinScanner *bufio.Scanner

func getStdinScanner() *bufio.Scanner {
	if stdinScanner == nil {
		stdinScanner = bufio.NewScanner(os.Stdin)
	}
	return stdinScanner
}

func ReadPipedInput() string {
	fi, err := os.Stdin.Stat()
	if err != nil || fi.Mode()&os.ModeCharDevice != 0 {
		return ""
	}
	scanner := getStdinScanner()
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return strings.TrimSpace(strings.Join(lines, "\n"))
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func ReadPipedLine() string {
	fi, err := os.Stdin.Stat()
	if err != nil || fi.Mode()&os.ModeCharDevice != 0 {
		return ""
	}
	scanner := getStdinScanner()
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

type inputModel struct {
	textInput textinput.Model
	done      bool
	value     string
	initCmd   tea.Cmd
}

func (m inputModel) Init() tea.Cmd {
	return m.initCmd
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			m.value = m.textInput.Value()
			m.done = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.done = true
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m inputModel) View() tea.View {
	if m.done {
		return tea.NewView("")
	}
	return tea.NewView(m.textInput.View())
}

func PromptInput(prompt string, placeholder string) (string, error) {
	if GlobalForAIFlag {
		return ReadPipedLine(), nil
	}

	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = prompt + " "
	focusCmd := ti.Focus()

	m := inputModel{textInput: ti, initCmd: focusCmd}
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	result := finalModel.(inputModel)
	return strings.TrimSpace(result.value), nil
}

func PromptPassword(prompt string) (string, error) {
	if GlobalForAIFlag {
		return ReadPipedLine(), nil
	}

	ti := textinput.New()
	ti.Placeholder = "••••••••"
	ti.Prompt = prompt + " "
	ti.EchoMode = textinput.EchoPassword
	focusCmd := ti.Focus()

	m := inputModel{textInput: ti, initCmd: focusCmd}
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	result := finalModel.(inputModel)
	return result.value, nil
}

type textAreaModel struct {
	textarea textarea.Model
	done     bool
	value    string
	initCmd  tea.Cmd
}

func (m textAreaModel) Init() tea.Cmd {
	return m.initCmd
}

func (m textAreaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+d":
			m.value = m.textarea.Value()
			m.done = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.done = true
			return m, tea.Quit
		}
	}

	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m textAreaModel) View() tea.View {
	if m.done {
		return tea.NewView("")
	}
	return tea.NewView(m.textarea.View() + "\n(Ctrl+D to submit, Esc to cancel)")
}

func PromptTextArea(prompt string, placeholder string) (string, error) {
	if GlobalForAIFlag {
		return ReadPipedInput(), nil
	}

	PrintInfo(prompt)

	ta := textarea.New()
	ta.Placeholder = placeholder
	focusCmd := ta.Focus()

	m := textAreaModel{textarea: ta, initCmd: focusCmd}
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	result := finalModel.(textAreaModel)
	return strings.TrimSpace(result.value), nil
}

type selectModel struct {
	label    string
	options  []string
	cursor   int
	selected int
	done     bool
}

func (m selectModel) Init() tea.Cmd {
	return nil
}

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.cursor
			m.done = true
			return m, tea.Quit
		case "esc", "ctrl+c", "q":
			m.selected = -1
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m selectModel) View() tea.View {
	if m.done {
		return tea.NewView("")
	}
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(12))
	cursorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(10))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(7))

	fmt.Fprintf(&b, "%s\n\n", titleStyle.Render(m.label))
	for i, opt := range m.options {
		if i == m.cursor {
			fmt.Fprintf(&b, "%s\n", cursorStyle.Render(" > "+opt))
		} else {
			fmt.Fprintf(&b, "%s\n", normalStyle.Render("   "+opt))
		}
	}
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(8)).Render("(Use ↑/↓ or j/k to navigate, Enter to select, Esc to cancel)"))
	return tea.NewView(b.String())
}

func PromptSelect(label string, options []string) (int, error) {
	if len(options) == 0 {
		return -1, nil
	}
	if GlobalForAIFlag {
		line := ReadPipedLine()
		if line == "" {
			return -1, nil
		}
		idx, err := strconv.Atoi(line)
		if err != nil || idx < 1 || idx > len(options) {
			return -1, nil
		}
		return idx - 1, nil
	}

	m := selectModel{
		label:    label,
		options:  options,
		selected: -1,
	}
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return -1, err
	}
	res := finalModel.(selectModel)
	return res.selected, nil
}

type multiSelectModel struct {
	label    string
	options  []string
	cursor   int
	selected map[int]bool
	done     bool
	aborted  bool
}

func (m multiSelectModel) Init() tea.Cmd {
	return nil
}

func (m multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case " ":
			if m.selected[m.cursor] {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = true
			}
		case "enter":
			m.done = true
			return m, tea.Quit
		case "esc", "ctrl+c", "q":
			m.aborted = true
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m multiSelectModel) View() tea.View {
	if m.done {
		return tea.NewView("")
	}
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(12))
	cursorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(10))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(7))

	fmt.Fprintf(&b, "%s\n\n", titleStyle.Render(m.label))
	for i, opt := range m.options {
		check := "[ ]"
		if m.selected[i] {
			check = "[✓]"
		}
		if i == m.cursor {
			fmt.Fprintf(&b, "%s\n", cursorStyle.Render(fmt.Sprintf(" > %s %s", check, opt)))
		} else {
			fmt.Fprintf(&b, "%s\n", normalStyle.Render(fmt.Sprintf("   %s %s", check, opt)))
		}
	}
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(8)).Render("(Space to toggle, Enter to confirm, Esc to cancel)"))
	return tea.NewView(b.String())
}

func PromptMultiSelect(label string, options []string) (map[int]bool, error) {
	if len(options) == 0 {
		return nil, nil
	}
	if GlobalForAIFlag {
		line := ReadPipedLine()
		if line == "" || strings.EqualFold(line, "none") {
			return nil, nil
		}
		selected := make(map[int]bool)
		parts := strings.Split(line, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if idx, err := strconv.Atoi(p); err == nil && idx >= 1 && idx <= len(options) {
				selected[idx-1] = true
			}
		}
		return selected, nil
	}

	m := multiSelectModel{
		label:    label,
		options:  options,
		selected: make(map[int]bool),
	}
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}
	res := finalModel.(multiSelectModel)
	if res.aborted {
		return nil, nil
	}
	return res.selected, nil
}
