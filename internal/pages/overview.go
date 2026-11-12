package pages

import (
	"fmt"
	"image/color"
	"math"
	"math/rand/v2"
	"runtime"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/xeronith/bubbletea-showcase/internal/ui"
)

// Overview is the landing page: a live wordmark, runtime stats, and a signal
// chart rendered from real process data.
type Overview struct {
	active bool

	started time.Time
	now     time.Time
	steam   int
	gorouts int
	heapMB  float64
	sysMB   float64
	gcCount uint32
	signal  []float64
}

type overviewTickMsg struct{ at time.Time }

type overviewSampleMsg struct {
	at         time.Time
	goroutines int
	heapMB     float64
	sysMB      float64
	gcCount    uint32
}

// NewOverview creates the landing page.
func NewOverview() *Overview {
	now := time.Now()
	o := &Overview{
		started: now,
		now:     now,
		signal:  make([]float64, 0, 64),
	}
	for i := 0; i < 48; i++ {
		v := 0.48 + 0.22*math.Sin(float64(i)*0.31) + 0.1*math.Sin(float64(i)*0.83)
		o.signal = append(o.signal, clamp(v, 0.08, 0.96))
	}
	return o
}

// Title implements Page.
func (o *Overview) Title() string { return "Overview" }

// Subtitle implements Page.
func (o *Overview) Subtitle() string {
	return "A live tour of the Elm Architecture, rendered one cell at a time."
}

// Init implements Page.
func (o *Overview) Init() tea.Cmd { return nil }

// OnShow implements Page.
func (o *Overview) OnShow() tea.Cmd {
	o.active = true
	return tea.Batch(o.clockTick(), o.sampleTick())
}

// OnHide implements Page.
func (o *Overview) OnHide() { o.active = false }

// Editing implements Page.
func (o *Overview) Editing() bool { return false }

func (o *Overview) clockTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return overviewTickMsg{at: t}
	})
}

func (o *Overview) sampleTick() tea.Cmd {
	return tea.Tick(900*time.Millisecond, func(t time.Time) tea.Msg {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return overviewSampleMsg{
			at:         t,
			goroutines: runtime.NumGoroutine(),
			heapMB:     float64(m.HeapAlloc) / (1024 * 1024),
			sysMB:      float64(m.Sys) / (1024 * 1024),
			gcCount:    m.NumGC,
		}
	})
}

// Update implements Page.
func (o *Overview) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case overviewTickMsg:
		if !o.active {
			return nil
		}
		o.now = msg.at
		o.steam++
		return o.clockTick()

	case overviewSampleMsg:
		if !o.active {
			return nil
		}
		o.gorouts = msg.goroutines
		o.heapMB = msg.heapMB
		o.sysMB = msg.sysMB
		o.gcCount = msg.gcCount
		o.pushSignal()
		return o.sampleTick()
	}
	return nil
}

func (o *Overview) pushSignal() {
	prev := 0.5
	if n := len(o.signal); n > 0 {
		prev = o.signal[n-1]
	}
	drift := 0.5 + 0.24*math.Sin(float64(o.steam)*0.17)
	noise := (rand.Float64() - 0.5) * 0.18
	next := clamp(prev*0.72+drift*0.28+noise, 0.06, 0.97)
	o.signal = append(o.signal, next)
	if len(o.signal) > 64 {
		o.signal = o.signal[len(o.signal)-64:]
	}
}

// View implements Page.
func (o *Overview) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	heroH := 8
	useCards := height >= 26
	if height >= 28 {
		heroH = 10
	}
	if height < heroH+2 {
		return ui.Fit(o.hero(width, height), width, height)
	}

	hero := o.hero(width, heroH)
	if useCards {
		mainH := height - heroH - 7
		if mainH >= 8 {
			return ui.Fit(lipgloss.JoinVertical(lipgloss.Left,
				hero, "", o.stats(width, 5), "", o.main(width, mainH),
			), width, height)
		}
	}

	mainH := height - heroH - 3
	strip := o.statStrip(width)
	if mainH < 7 {
		return ui.Fit(lipgloss.JoinVertical(lipgloss.Left, hero, "", strip), width, height)
	}
	return ui.Fit(lipgloss.JoinVertical(lipgloss.Left,
		hero, "", strip, "", o.main(width, mainH),
	), width, height)
}

func (o *Overview) hero(width, height int) string {
	innerH := height - 2
	bodyW := width - 6
	if bodyW < 10 {
		return ""
	}

	tagline := ui.TextStyle.Render("A fun, functional, stateful way to build terminal apps.")
	subtitle := ui.MutedStyle.Render("Five live demos. One Elm loop. Rendered one cell at a time.")
	badges := strings.Join([]string{
		ui.Badge("v"+ui.DepVersion("charm.land/bubbletea/v2", "2.1.1"), ui.Cyan),
		ui.Badge("Elm architecture", ui.Purple),
		ui.Badge("Go "+shortGoVersion(), ui.Green),
	}, " ")

	withCup := bodyW >= ui.LogoWidth+3+ui.TeacupWidth && innerH >= 8

	var lines []string
	if withCup {
		top := lipgloss.JoinHorizontal(lipgloss.Top, ui.Logo(), "   ", ui.Teacup(o.steam))
		lines = append(lines, strings.Split(top, "\n")...)
		lines = append(lines, tagline)
		if innerH >= 10 {
			lines = append(lines, subtitle, "")
		}
		lines = append(lines, badges)
	} else {
		if bodyW >= ui.LogoWidth {
			lines = append(lines, strings.Split(ui.Logo(), "\n")...)
		} else {
			lines = append(lines, ui.GradientBold("Bubble Tea", ui.Pink, ui.Purple, ui.Cyan))
		}
		if innerH >= 8 {
			lines = append(lines, "")
		}
		lines = append(lines, tagline)
		if innerH >= 7 {
			lines = append(lines, subtitle)
		}
		if innerH >= 6 {
			lines = append(lines, "")
		}
		lines = append(lines, badges)
	}

	if len(lines) > innerH {
		lines = lines[:innerH]
	}
	return ui.GlowCard("", strings.Join(lines, "\n"), width, height)
}

func (o *Overview) statStrip(width int) string {
	up := o.now.Sub(o.started)
	if up < 0 {
		up = 0
	}
	uptime := fmt.Sprintf("%02d:%02d", int(up.Minutes()), int(up.Seconds())%60)

	item := func(icon string, accent color.Color, value, label string) string {
		return lipgloss.NewStyle().Foreground(accent).Render(icon) + " " +
			lipgloss.NewStyle().Foreground(ui.Text).Bold(true).Render(value) + " " +
			ui.DimStyle.Render(label)
	}
	sep := ui.DimStyle.Render("   ·   ")

	line := item("●", ui.Purple, fmt.Sprintf("%d", o.gorouts), "goroutines") + sep +
		item("▤", ui.Pink, fmt.Sprintf("%.1f MB", o.heapMB), "heap") + sep +
		item("◷", ui.Cyan, uptime, "uptime")
	return ui.Fit("  "+line, width, 1)
}

func (o *Overview) stats(width, height int) string {
	gap := 1
	cw := (width - 2*gap) / 3
	widths := []int{cw, cw, width - cw*2 - 2*gap}

	up := o.now.Sub(o.started)
	if up < 0 {
		up = 0
	}
	uptime := fmt.Sprintf("%02d:%02d", int(up.Minutes()), int(up.Seconds())%60)

	values := []string{
		fmt.Sprintf("%d", o.gorouts),
		fmt.Sprintf("%.1f MB", o.heapMB),
		uptime,
	}
	labels := []string{
		"live goroutines",
		fmt.Sprintf("%d GC · %.0f MB sys", o.gcCount, o.sysMB),
		"this session",
	}
	titles := []string{"runtime", "memory", "uptime"}
	accents := []color.Color{ui.Purple, ui.Pink, ui.Cyan}

	cards := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		body := lipgloss.NewStyle().
			Foreground(accents[i]).
			Bold(true).
			Render(values[i]) + "\n" + ui.DimStyle.Render(labels[i])
		cards = append(cards, ui.Card(accents[i], titles[i], body, widths[i], height))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, withGap(cards, gap)...)
}

func withGap(blocks []string, gap int) []string {
	out := make([]string, 0, len(blocks)*2)
	for i, b := range blocks {
		if i > 0 {
			out = append(out, strings.Repeat(" ", gap))
		}
		out = append(out, b)
	}
	return out
}

func (o *Overview) main(width, height int) string {
	leftW := width * 3 / 5
	rightW := width - leftW - 1
	if rightW < 30 {
		return o.signalCard(width, height)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top,
		o.signalCard(leftW, height), " ", o.elmCard(rightW, height))
}

func (o *Overview) signalCard(width, height int) string {
	bodyW := ui.CardBodyWidth(width)
	innerH := height - 2
	chartH := innerH - 3
	if chartH < 2 {
		chartH = 2
	}
	if bodyW < 4 {
		return ""
	}

	chart := ui.Area(o.signal, bodyW, chartH)
	latest := 0.0
	if n := len(o.signal); n > 0 {
		latest = o.signal[n-1]
	}
	footer := ui.SpaceBetween(
		ui.DimStyle.Render(fmt.Sprintf("45s window · sample %d", o.steam)),
		lipgloss.NewStyle().Foreground(ui.Cyan).Bold(true).Render(fmt.Sprintf("%3.0f%%", latest*100)),
		bodyW,
	)
	body := chart + "\n" + footer
	return ui.Card(ui.Purple, "live signal", body, width, height)
}

func (o *Overview) elmCard(width, height int) string {
	bodyW := ui.CardBodyWidth(width)
	innerH := height - 2
	if bodyW < 10 {
		return ""
	}

	arrow := lipgloss.NewStyle().Foreground(ui.Pink).Render("──▶")
	key := func(s string) string {
		return lipgloss.NewStyle().Foreground(ui.Text).Bold(true).Render(s)
	}
	value := func(s string) string { return ui.MutedStyle.Render(s) }

	rows := [][2]string{
		{"Msg", "Update"},
		{"Update", "Model + Cmd"},
		{"Model", "View"},
		{"View", "Terminal"},
		{"Cmd", "Msg (async)"},
	}
	lines := make([]string, 0, len(rows)+2)
	for _, r := range rows {
		lines = append(lines, key(r[0])+" "+arrow+" "+value(r[1]))
	}

	body := strings.Join(lines, "\n")
	if innerH >= len(lines)+3 {
		body += "\n\n" + ui.Gradient("one model · one view")
	}
	return ui.Card(ui.Cyan, "the elm loop", body, width, height)
}

func shortGoVersion() string {
	parts := strings.Split(ui.GoVersion(), ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return ui.GoVersion()
}

func clamp(v, lo, hi float64) float64 {
	return math.Max(lo, math.Min(hi, v))
}
