package pages

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/xeronith/bubbletea-showcase/internal/ui"
)

const motionBars = 160

// Motion is a 60 fps spectrum driven entirely by Bubble Tea messages. Move the
// mouse to push the wave around; no imperative paint loop in sight.
type Motion struct {
	active bool
	paused bool
	mode   int
	speed  float64
	phase  float64

	bars   []float64
	fps    float64
	frames int
	last   time.Time

	mouseX  float64
	pageW   int
	originX int
}

type motionTickMsg struct{ at time.Time }

// NewMotion creates the animation page.
func NewMotion() *Motion {
	m := &Motion{
		speed:  1,
		bars:   make([]float64, motionBars),
		mouseX: -1,
	}
	for i := range m.bars {
		m.bars[i] = 0.3
	}
	return m
}

// Title implements Page.
func (m *Motion) Title() string { return "Motion" }

// Subtitle implements Page.
func (m *Motion) Subtitle() string {
	return "A 60 fps spectrum pushed around by commands, keyboard, and the mouse."
}

// Init implements Page.
func (m *Motion) Init() tea.Cmd { return nil }

// OnShow implements Page.
func (m *Motion) OnShow() tea.Cmd {
	m.active = true
	m.last = time.Now()
	return m.tick()
}

// OnHide implements Page.
func (m *Motion) OnHide() { m.active = false }

// Editing implements Page.
func (m *Motion) Editing() bool { return false }

func (m *Motion) tick() tea.Cmd {
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
		return motionTickMsg{at: t}
	})
}

// Update implements Page.
func (m *Motion) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.pageW = msg.Width - 4
		m.originX = 2
		if msg.Width >= 96 {
			m.pageW = msg.Width - 30
			m.originX = 28
		}
		return nil

	case motionTickMsg:
		if !m.active {
			return nil
		}
		now := msg.at
		dt := now.Sub(m.last).Seconds()
		if dt <= 0 || dt > 0.25 {
			dt = 1.0 / 60.0
		}
		m.last = now
		m.frames++

		instant := 1 / dt
		if m.fps == 0 {
			m.fps = instant
		} else {
			m.fps = m.fps*0.92 + instant*0.08
		}

		if !m.paused {
			m.phase += dt * m.speed
		}
		m.updateBars(dt)
		return m.tick()

	case tea.MouseMsg:
		if m.pageW > 0 {
			mouse := msg.Mouse()
			rel := float64(mouse.X-m.originX) / float64(m.pageW)
			m.mouseX = clamp(math.Max(0, math.Min(1, rel)), 0, 1)
		}
		return nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "space":
			m.paused = !m.paused
		case "m":
			m.mode = (m.mode + 1) % 3
		case "+", "=":
			m.speed = math.Min(2.5, m.speed+0.1)
		case "-", "_":
			m.speed = math.Max(0.1, m.speed-0.1)
		case "r":
			m.speed = 1
			m.mode = 0
			m.paused = false
		}
		return nil
	}
	return nil
}

func (m *Motion) updateBars(dt float64) {
	blend := math.Min(1, dt*14)
	for i := range m.bars {
		x := float64(i) / float64(len(m.bars)-1)
		target := m.waveAt(x)
		if m.mouseX >= 0 {
			d := x - m.mouseX
			target += 0.3 * math.Exp(-(d*d)/0.006)
		}
		target = math.Max(0.04, math.Min(1, target))
		m.bars[i] += (target - m.bars[i]) * blend
	}
}

func (m *Motion) waveAt(x float64) float64 {
	switch m.mode {
	case 1: // ripple
		d := math.Hypot((x-0.5)*1.6, 0.35)
		return 0.5 + 0.38*math.Sin(m.phase*2.2-d*9)
	case 2: // noise
		return 0.5 +
			0.22*math.Sin(x*31+m.phase*2.4) +
			0.18*math.Sin(x*17-m.phase*1.7) +
			0.1*math.Sin(x*67+m.phase*3.1)
	default: // sine
		return 0.48 +
			0.25*math.Sin(m.phase*2.1+x*7.5) +
			0.12*math.Sin(m.phase*3.4+x*17)
	}
}

func (m *Motion) modeName() string {
	switch m.mode {
	case 1:
		return "ripple"
	case 2:
		return "noise"
	default:
		return "sine"
	}
}

// View implements Page.
func (m *Motion) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	if height < 18 {
		return ui.Fit(m.spectrumCard(width, height), width, height)
	}

	controlsH := 7
	spectrum := m.spectrumCard(width, height-controlsH-1)
	controls := m.controls(width, controlsH)
	return ui.Fit(lipgloss.JoinVertical(lipgloss.Left, spectrum, "", controls), width, height)
}

func (m *Motion) spectrumCard(width, height int) string {
	bodyW := ui.CardBodyWidth(width)
	// title + status line + borders
	chartH := height - 4
	if chartH < 2 {
		chartH = 2
	}
	if bodyW < 4 {
		return ""
	}

	status := ""
	if m.paused {
		status = lipgloss.NewStyle().Foreground(ui.Yellow).Bold(true).Render("▌▌ paused")
	} else {
		status = ui.DimStyle.Render("space to pause · m to change the wave")
	}
	head := ui.SpaceBetween(
		ui.DimStyle.Render(fmt.Sprintf("%d fps · %s · %.1f×", int(m.fps+0.5), m.modeName(), m.speed)),
		status,
		bodyW,
	)

	chart := m.renderSpectrum(bodyW, chartH)
	return ui.GlowCard("live spectrum", head+"\n"+chart, width, height)
}

func (m *Motion) renderSpectrum(width, height int) string {
	colors := lipgloss.Blend1D(width, ui.Pink, ui.Purple, ui.Cyan)
	blocks := make([]string, width)
	for x := 0; x < width; x++ {
		blocks[x] = lipgloss.NewStyle().Foreground(colors[x]).Render("█")
	}

	heights := make([]int, width)
	for x := 0; x < width; x++ {
		pos := float64(x) / float64(max(1, width-1)) * float64(len(m.bars)-1)
		lo := int(math.Floor(pos))
		hi := min(lo+1, len(m.bars)-1)
		frac := pos - float64(lo)
		v := m.bars[lo]*(1-frac) + m.bars[hi]*frac
		heights[x] = int(math.Round(v * float64(height)))
	}

	rows := make([]string, height)
	for row := 0; row < height; row++ {
		threshold := height - row
		var b strings.Builder
		b.Grow(width * 8)
		for x := 0; x < width; x++ {
			if heights[x] >= threshold {
				b.WriteString(blocks[x])
			} else {
				b.WriteByte(' ')
			}
		}
		rows[row] = b.String()
	}
	return strings.Join(rows, "\n")
}

func (m *Motion) controls(width, height int) string {
	gap := 1
	cw := (width - 2*gap) / 3
	widths := []int{cw, cw, width - cw*2 - 2*gap}

	fpsValue := fmt.Sprintf("%d", int(m.fps+0.5))
	fps := ui.Card(ui.Green, "refresh",
		lipgloss.NewStyle().Foreground(ui.Green).Bold(true).Render(fpsValue)+" fps\n"+
			ui.DimStyle.Render(fmt.Sprintf("%d frames rendered", m.frames)),
		widths[0], height)

	wave := ui.Card(ui.Purple, "waveform",
		lipgloss.NewStyle().Foreground(ui.Purple).Bold(true).Render(m.modeName())+"\n"+
			ui.DimStyle.Render("m to cycle"),
		widths[1], height)

	keys := ui.Keycap("space") + " " + ui.DimStyle.Render("pause") + "\n" +
		ui.Keycap("+ -") + " " + ui.DimStyle.Render(fmt.Sprintf("%.1f× speed", m.speed))
	controls := ui.Card(ui.Cyan, "controls", keys, widths[2], height)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		fps, strings.Repeat(" ", gap), wave, strings.Repeat(" ", gap), controls)
}
