package utils

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
	"github.com/rs/zerolog/log"
)

type Unit string

const UnitBytes Unit = ""

const (
	minWidth      = 24
	defaultWidth  = 80
	settledPrefix = 4
	minBarWidth   = 8
	maxBarWidth   = 30
	sweepWidth    = 3
)

const (
	percentField     = 4
	transferredField = 14
	rateField        = 11
	etaField         = 11
	averageField     = 15
)

var (
	mutedStyle   = lipgloss.NewStyle().Foreground(ColorMuted)
	barFillStyle = lipgloss.NewStyle().Foreground(ColorBlue)
	barRestStyle = lipgloss.NewStyle().Foreground(ColorChrome)
)

var render struct {
	sync.Mutex
	live   *Meter
	hidden bool
}

var cursorGuard sync.Once

func termWidth() int {
	if w, _, err := term.GetSize(os.Stdout.Fd()); err == nil && w > 0 {
		return w
	}
	if n, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && n >= minWidth {
		return n
	}
	return defaultWidth
}

func hideCursorLocked() {
	if render.hidden {
		return
	}
	cursorGuard.Do(func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
		go func() {
			for range signals {
				ShowCursor()
			}
		}()
	})
	render.hidden = true
	lipgloss.Print("\033[?25l")
}

func showCursorLocked() {
	if !render.hidden {
		return
	}
	render.hidden = false
	lipgloss.Print("\033[?25h")
}

func ShowCursor() {
	render.Lock()
	defer render.Unlock()
	showCursorLocked()
}

func emit(line string) {
	render.Lock()
	defer render.Unlock()
	if render.live == nil {
		lipgloss.Println(line)
		return
	}
	var b strings.Builder
	b.WriteString(strings.Repeat("\033[1A\033[2K", render.live.drawn))
	b.WriteString("\r\033[2K" + line + "\n")
	frame := render.live.frameLocked()
	for _, l := range frame {
		b.WriteString("\r\033[2K" + l + "\n")
	}
	lipgloss.Print(b.String())
	render.live.drawn = len(frame)
}

type sample struct {
	at    time.Time
	value int64
}

type rateWindow struct {
	samples []sample
}

func (r *rateWindow) add(value int64, at time.Time) {
	r.samples = append(r.samples, sample{at: at, value: value})
	cutoff := at.Add(-800 * time.Millisecond)
	drop := 0
	for drop < len(r.samples)-2 && r.samples[drop].at.Before(cutoff) {
		drop++
	}
	r.samples = r.samples[drop:]
}

func (r *rateWindow) current() float64 {
	if len(r.samples) < 2 {
		return 0
	}
	first, last := r.samples[0], r.samples[len(r.samples)-1]
	span := last.at.Sub(first.at)
	if span < 200*time.Millisecond {
		return 0
	}
	delta := float64(last.value - first.value)
	if delta <= 0 {
		return 0
	}
	return delta / span.Seconds()
}

type Meter struct {
	verb    string
	name    string
	context string
	unit    Unit
	total   int64
	current int64
	start   time.Time
	window  rateWindow
	frames  int
	drawn   int
	elapsed time.Duration
	group   *Group
	stop    chan struct{}
	ticker  sync.WaitGroup
	settled bool
}

func NewMeter(verb, name string, total int64, unit Unit) *Meter {
	m := &Meter{
		verb:  verb,
		name:  name,
		unit:  unit,
		total: total,
		start: time.Now(),
		stop:  make(chan struct{}),
	}
	render.Lock()
	if !OutputPersists() {
		render.live = m
		hideCursorLocked()
		m.paintLocked()
	}
	render.Unlock()
	if OutputPersists() {
		m.report()
	}
	m.ticker.Go(m.run)
	return m
}

func (m *Meter) Context(context string) {
	render.Lock()
	defer render.Unlock()
	m.context = context
}

func (m *Meter) Set(value int64) {
	render.Lock()
	defer render.Unlock()
	m.current = value
}

func (m *Meter) Add(delta int64) {
	render.Lock()
	defer render.Unlock()
	m.current += delta
}

func (m *Meter) Write(p []byte) (int, error) {
	m.Add(int64(len(p)))
	return len(p), nil
}

func (m *Meter) run() {
	interval := 100 * time.Millisecond
	if OutputPersists() {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-m.stop:
			return
		case <-ticker.C:
			m.tick()
		}
	}
}

func (m *Meter) tick() {
	if OutputPersists() {
		m.report()
		return
	}
	render.Lock()
	defer render.Unlock()
	m.frames++
	m.paintLocked()
}

func (m *Meter) report() {
	render.Lock()
	m.window.add(m.current, time.Now())
	percent, hasPercent := m.percentLocked()
	label := m.name
	if m.context != "" {
		label += " " + m.context
	}
	stats := strings.Join(m.statsLocked(true, true, true), "  ")
	current, total, rate := m.current, m.total, m.window.current()
	eta := m.etaLocked()
	render.Unlock()

	if GlobalDebugFlag {
		entry := log.Info().
			Int64("current", current).
			Int64("total", total).
			Float64("rate", rate).
			Str("unit", string(m.unit)).
			Str("eta", eta)
		if hasPercent {
			entry = entry.Int("percent", percent)
		}
		entry.Msg(strings.TrimSpace(m.verb + " " + label))
		return
	}
	lipgloss.Println(fmt.Sprintf("  ↻ %s  %s", label, stats))
}

func (m *Meter) percentLocked() (int, bool) {
	if m.total <= 0 {
		return 0, false
	}
	percent := int(m.current * 100 / m.total)
	return min(max(percent, 0), 100), true
}

func (m *Meter) averageLocked() float64 {
	elapsed := m.elapsed
	if !m.settled {
		elapsed = time.Since(m.start)
		if elapsed < 200*time.Millisecond {
			return 0
		}
	}
	if elapsed <= 0 {
		return 0
	}
	return float64(m.current) / elapsed.Seconds()
}

func (m *Meter) etaLocked() string {
	if m.total <= 0 {
		return "eta unknown"
	}
	rate := m.window.current()
	if rate <= 0 {
		return "eta unknown"
	}
	remaining := float64(m.total - m.current)
	if remaining <= 0 {
		return "eta 0s"
	}
	seconds := remaining / rate
	if seconds > 100*3600 {
		return "eta unknown"
	}
	return "eta " + FormatEstimate(time.Duration(seconds*float64(time.Second)))
}

func (m *Meter) transferredLocked() string {
	if m.unit == UnitBytes {
		if m.total > 0 {
			return bytePair(m.current, m.total)
		}
		return FormatAmount(m.current, UnitBytes)
	}
	if m.total > 0 {
		return fmt.Sprintf("%d / %d %s", m.current, m.total, m.unit)
	}
	return FormatAmount(m.current, m.unit)
}

func (m *Meter) statsLocked(rate, eta, average bool) []string {
	var stats []string
	if percent, ok := m.percentLocked(); ok {
		stats = append(stats, fmt.Sprintf("%3d%%", percent))
	}
	stats = append(stats, m.transferredLocked())
	if rate {
		stats = append(stats, FormatRate(m.window.current(), m.unit))
	}
	if eta {
		stats = append(stats, m.etaLocked())
	}
	if average {
		stats = append(stats, "avg "+FormatRate(m.averageLocked(), m.unit))
	}
	return stats
}

func (m *Meter) barLocked(width int) string {
	if percent, ok := m.percentLocked(); ok {
		filled := width * percent / 100
		return barFillStyle.Render(strings.Repeat("─", filled)) +
			barRestStyle.Render(strings.Repeat("─", width-filled))
	}
	cells := make([]bool, width)
	for offset := range sweepWidth {
		cells[(m.frames+offset)%width] = true
	}
	var b strings.Builder
	for _, lit := range cells {
		if lit {
			b.WriteString(barFillStyle.Render("─"))
			continue
		}
		b.WriteString(barRestStyle.Render("─"))
	}
	return b.String()
}

func (m *Meter) reservedLocked(rate, eta, average bool) int {
	width := transferredField
	if _, ok := m.percentLocked(); ok {
		width += percentField + 2
	}
	if rate {
		width += rateField + 2
	}
	if eta {
		width += etaField + 2
	}
	if average {
		width += averageField + 2
	}
	return width
}

func (m *Meter) meterLineLocked(width int) string {
	rate, eta, average := true, true, true
	for {
		room := width - 2 - 2 - m.reservedLocked(rate, eta, average)
		if room >= minBarWidth {
			bar := min(room, maxBarWidth)
			return "  " + m.barLocked(bar) + "  " + strings.Join(m.statsLocked(rate, eta, average), "  ")
		}
		switch {
		case average:
			average = false
		case eta:
			eta = false
		case rate:
			rate = false
		default:
			return "  " + strings.Join(m.statsLocked(false, false, false), "  ")
		}
	}
}

func (m *Meter) headerLineLocked(width int) string {
	prefix := "↻ " + m.verb + " "
	room := width - lipgloss.Width(prefix)
	context := m.context
	if context != "" && lipgloss.Width(m.name)+2+lipgloss.Width(context) > room {
		context = ""
	}
	if context != "" {
		room -= 2 + lipgloss.Width(context)
	}
	line := infoStyle.Render(prefix + clip(m.name, room))
	if context != "" {
		line += "  " + mutedStyle.Render(context)
	}
	return line
}

func (m *Meter) frameLocked() []string {
	width := termWidth()
	return []string{m.headerLineLocked(width), m.meterLineLocked(width)}
}

func (m *Meter) paintLocked() {
	m.window.add(m.current, time.Now())
	frame := m.frameLocked()
	var b strings.Builder
	b.WriteString(strings.Repeat("\033[1A\033[2K", m.drawn))
	for _, line := range frame {
		b.WriteString("\r\033[2K" + line + "\n")
	}
	lipgloss.Print(b.String())
	m.drawn = len(frame)
}

func (m *Meter) closeLocked() {
	if m.drawn > 0 {
		lipgloss.Print(strings.Repeat("\033[1A\033[2K", m.drawn))
		m.drawn = 0
	}
	if render.live == m {
		render.live = nil
		showCursorLocked()
	}
}

func (m *Meter) halt() time.Duration {
	close(m.stop)
	m.ticker.Wait()
	render.Lock()
	defer render.Unlock()
	m.elapsed = time.Since(m.start)
	m.settled = true
	m.closeLocked()
	return m.elapsed
}

func (m *Meter) Close() time.Duration {
	return m.halt()
}

func (m *Meter) Done() {
	elapsed := m.halt()
	render.Lock()
	amount := FormatAmount(m.current, m.unit)
	average := FormatRate(m.averageLocked(), m.unit)
	render.Unlock()

	line := SettledLine(m.name, amount, elapsed, average)
	if m.group != nil {
		m.group.settle(line, "", nil)
		return
	}
	PrintSuccess(line)
}

func (m *Meter) Fail(err error) {
	m.halt()
	if m.group != nil {
		m.group.settle("", FailureLine(m.name, err), err)
		return
	}
	PrintIndentedError(FailureLine(m.name, err), err)
}

func FailureLine(name string, err error) string {
	reason := strings.Join(strings.Fields(fmt.Sprint(err)), " ")
	return fitName(name, 2) + ": " + reason
}

func fitName(name string, tail int) string {
	if OutputPersists() {
		return name
	}
	return clip(name, termWidth()-settledPrefix-tail)
}

func SettledLine(name, amount string, elapsed time.Duration, average string) string {
	var tail []string
	if amount != "" {
		tail = append(tail, amount)
	}
	tail = append(tail, FormatElapsed(elapsed))
	if average != "" {
		tail = append(tail, "avg "+average)
	}
	rest := strings.Join(tail, "  ")
	return fitName(name, lipgloss.Width(rest)+2) + "  " + rest
}

type Group struct {
	mu       sync.Mutex
	label    string
	noun     string
	start    time.Time
	ok       int
	failed   int
	collapse bool
	counter  *Meter
}

func NewGroup(label, noun string) *Group {
	return &Group{label: label, noun: noun, start: time.Now()}
}

func (g *Group) Collapse() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.collapse = true
}

func (g *Group) Meter(verb, name string, total int64, unit Unit) *Meter {
	m := NewMeter(verb, name, total, unit)
	m.group = g
	return m
}

func (g *Group) Count(verb string, total int64) *Meter {
	m := NewMeter(verb, fmt.Sprintf("%d %s", total, g.noun), total, Unit(g.noun))
	g.mu.Lock()
	g.counter = m
	g.mu.Unlock()
	return m
}

func (g *Group) settle(success, failure string, err error) {
	g.mu.Lock()
	if failure != "" {
		g.failed++
	} else {
		g.ok++
	}
	counter := g.counter
	collapse := g.collapse
	g.mu.Unlock()

	if counter != nil {
		counter.Add(1)
	}
	if failure != "" {
		PrintIndentedError(failure, err)
		return
	}
	if !collapse {
		PrintSuccess(success)
	}
}

func (g *Group) OK(name, amount string, elapsed time.Duration, average string) {
	g.settle(SettledLine(name, amount, elapsed, average), "", nil)
}

func (g *Group) Fail(name string, err error) {
	g.settle("", FailureLine(name, err), err)
}

func (g *Group) Done(moved string) {
	g.mu.Lock()
	counter := g.counter
	g.counter = nil
	ok, failed := g.ok, g.failed
	g.mu.Unlock()

	if counter != nil {
		counter.Close()
	}
	if ok+failed <= 1 && failed == 0 {
		return
	}

	elapsed := time.Since(g.start)
	average := FormatRate(float64(ok+failed)/max(elapsed.Seconds(), 0.001), Unit(g.noun))
	count := fmt.Sprintf("%d %s", ok, g.noun)
	if failed > 0 {
		count = fmt.Sprintf("%d ok, %d failed", ok, failed)
	}
	if moved != "" {
		count += "  " + moved
	}
	if failed > 0 {
		PrintError(SettledLine(g.label, count, elapsed, average), nil)
		return
	}
	PrintInfo(SettledLine(g.label, count, elapsed, average))
}

var byteUnits = []string{"B", "KB", "MB", "GB", "TB", "PB"}

func byteScale(value int64) (float64, string) {
	divisor := 1.0
	index := 0
	for float64(value)/divisor >= 1024 && index < len(byteUnits)-1 {
		divisor *= 1024
		index++
	}
	return divisor, byteUnits[index]
}

func formatNumber(value float64) string {
	if value < 100 {
		return strconv.FormatFloat(value, 'f', 1, 64)
	}
	return strconv.FormatFloat(value, 'f', 0, 64)
}

func bytePair(current, total int64) string {
	divisor, suffix := byteScale(total)
	return fmt.Sprintf("%s / %s %s", formatNumber(float64(current)/divisor), formatNumber(float64(total)/divisor), suffix)
}

func FormatAmount(value int64, unit Unit) string {
	if unit != UnitBytes {
		return fmt.Sprintf("%d %s", value, unit)
	}
	divisor, suffix := byteScale(value)
	return formatNumber(float64(value)/divisor) + " " + suffix
}

func FormatRate(value float64, unit Unit) string {
	if unit != UnitBytes {
		return formatNumber(value) + " " + string(unit) + "/s"
	}
	divisor, suffix := byteScale(int64(value))
	return formatNumber(value/divisor) + " " + suffix + "/s"
}

func FormatElapsed(d time.Duration) string {
	if d < time.Minute {
		return strconv.FormatFloat(d.Seconds(), 'f', 1, 64) + "s"
	}
	return FormatEstimate(d)
}

func FormatEstimate(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm%02ds", int(d.Minutes()), int(d.Seconds())%60)
	default:
		return fmt.Sprintf("%dh%02dm", int(d.Hours()), int(d.Minutes())%60)
	}
}

func clip(s string, room int) string {
	if room <= 1 || lipgloss.Width(s) <= room {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)+"…") > room {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}
