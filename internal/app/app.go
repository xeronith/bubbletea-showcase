package app

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/xeronith/bubbletea-showcase/internal/pages"
	"github.com/xeronith/bubbletea-showcase/internal/ui"
)

const (
	// MinWidth and MinHeight are the smallest terminal the full layout is
	// designed for.
	MinWidth  = 76
	MinHeight = 24

	padX        = 2
	sidebarW    = 24
	sidebarGap  = 2
	headerRows  = 2
	topPad      = 1
	bottomPad   = 1
	footerRows  = 1
	mainTop     = topPad + headerRows + 1
	sidebarItem = 4 // row of the first navigation item inside the sidebar
)

type clockMsg struct{ at time.Time }

// Model is the application shell. It owns navigation, input routing, and the
// persistent chrome around whatever page is currently mounted.
type Model struct {
	width  int
	height int

	pages   []pages.Page
	current int

	clock    time.Time
	showHelp bool
	hover    int
}

// New builds the showcase shell.
func New() *Model {
	return &Model{
		pages: pages.All(),
		clock: time.Now(),
		hover: -1,
	}
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.clockTick()}
	for _, p := range m.pages {
		cmds = append(cmds, p.Init())
	}
	cmds = append(cmds, m.pages[m.current].OnShow())
	return tea.Batch(cmds...)
}

func (m *Model) clockTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return clockMsg{at: t}
	})
}

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		cmds := make([]tea.Cmd, 0, len(m.pages))
		for _, p := range m.pages {
			cmds = append(cmds, p.Update(msg))
		}
		return m, tea.Batch(cmds...)

	case clockMsg:
		m.clock = msg.at
		return m, m.clockTick()

	case tea.MouseMsg:
		return m, m.handleMouse(msg)

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, m.pages[m.current].Update(msg)
}

func (m *Model) handleMouse(msg tea.MouseMsg) tea.Cmd {
	mouse := msg.Mouse()

	index := -1
	if m.hasSidebar() && mouse.X >= padX && mouse.X < padX+sidebarW {
		rel := mouse.Y - mainTop
		if rel >= sidebarItem && rel < sidebarItem+len(m.pages) {
			index = rel - sidebarItem
		}
	}
	m.hover = index

	if _, ok := msg.(tea.MouseClickMsg); ok && index >= 0 {
		return m.switchTo(index)
	}

	return m.pages[m.current].Update(msg)
}

func (m *Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if m.showHelp {
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		m.showHelp = false
		return m, nil
	}

	if key == "ctrl+c" {
		return m, tea.Quit
	}

	page := m.pages[m.current]
	if !page.Editing() {
		switch key {
		case "q":
			return m, tea.Quit
		case "?":
			m.showHelp = true
			return m, nil
		case "tab":
			if !m.pageUsesTab() {
				return m, m.switchTo(m.current + 1)
			}
		case "shift+tab":
			if !m.pageUsesTab() {
				return m, m.switchTo(m.current - 1)
			}
		case "1":
			return m, m.switchTo(0)
		case "2":
			return m, m.switchTo(1)
		case "3":
			return m, m.switchTo(2)
		case "4":
			return m, m.switchTo(3)
		case "5":
			return m, m.switchTo(4)
		}
	}

	return m, page.Update(msg)
}

func (m *Model) pageUsesTab() bool {
	p, ok := m.pages[m.current].(interface{ UsesTab() bool })
	return ok && p.UsesTab()
}

func (m *Model) switchTo(index int) tea.Cmd {
	n := len(m.pages)
	if n == 0 {
		return nil
	}
	index = ((index % n) + n) % n
	if index == m.current {
		return nil
	}

	m.pages[m.current].OnHide()
	m.current = index
	m.showHelp = false
	m.hover = -1
	if m.width > 0 && m.height > 0 {
		m.pages[m.current].Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	}
	return m.pages[m.current].OnShow()
}

func (m *Model) hasSidebar() bool {
	return m.width >= 96
}

func (m *Model) mainHeight() int {
	return m.height - topPad - headerRows - 1 - footerRows - bottomPad
}

// View implements tea.Model.
func (m *Model) View() tea.View {
	if m.width <= 0 || m.height <= 0 {
		v := tea.NewView("Brewing something nice…")
		v.AltScreen = true
		v.BackgroundColor = ui.BG
		return v
	}

	var content string
	if m.width < MinWidth || m.height < MinHeight {
		content = ui.TooSmall(m.width, m.height, MinWidth, MinHeight)
	} else {
		content = m.layout()
	}

	content = ui.Fit(content, m.width, m.height)
	content = lipgloss.NewStyle().
		Background(ui.BG).
		Width(m.width).
		Height(m.height).
		Render(content)

	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeAllMotion
	v.BackgroundColor = ui.BG
	v.WindowTitle = "Bubble Tea Showcase"
	return v
}

func (m *Model) layout() string {
	rows := []string{
		"",
		m.header(),
		"",
		m.main(),
		"",
		m.footer(),
	}
	return strings.Join(rows, "\n")
}

func (m *Model) header() string {
	left := "  " +
		lipgloss.NewStyle().Foreground(ui.Orange).Render("☕") + "  " +
		ui.GradientBold("Bubble Tea Showcase", ui.Pink, ui.Purple, ui.Cyan) + "  " +
		ui.Badge("live", ui.Green)

	right := ui.DimStyle.Render(m.clock.Format("Mon 02 Jan")) +
		ui.MutedStyle.Render("  ·  ") +
		ui.TextStyle.Render(m.clock.Format("15:04:05"))

	line1 := ui.SpaceBetween(left, right, m.width)
	line2 := ui.DimStyle.Render("  A hands-on tour of the Elm Architecture · use 1–5 to explore")
	return line1 + "\n" + line2
}

func (m *Model) main() string {
	h := m.mainHeight()
	if m.hasSidebar() {
		pageW := m.width - 2*padX - sidebarW - sidebarGap
		sidebar := ui.Fit(m.sidebar(h), sidebarW, h)
		page := ui.Fit(m.pageArea(pageW, h), pageW, h)
		return "  " + lipgloss.JoinHorizontal(lipgloss.Top,
			sidebar, strings.Repeat(" ", sidebarGap), page) + "  "
	}

	pageW := m.width - 2*padX
	page := ui.Fit(m.pageArea(pageW, h), pageW, h)
	return "  " + page + "  "
}

func (m *Model) sidebar(height int) string {
	w := sidebarW
	lines := []string{
		"",
		ui.DimStyle.Render(" EXPLORE"),
		ui.HLine(w, ui.Border),
		"",
	}

	for i, p := range m.pages {
		selected := i == m.current
		hovered := i == m.hover

		num := fmt.Sprintf("%02d", i+1)
		label := p.Title()
		if len(label) > 14 {
			label = label[:14]
		}

		prefix := "  "
		style := lipgloss.NewStyle().Foreground(ui.Muted).Width(w)
		switch {
		case selected:
			prefix = lipgloss.NewStyle().Foreground(ui.Pink).Render("▌") + " "
			style = style.
				Foreground(ui.Text).
				Background(ui.SurfaceAlt).
				Bold(true)
		case hovered:
			prefix = "  "
			style = style.Foreground(ui.Text).Background(ui.Surface)
		}

		row := prefix + ui.DimStyle.Render(num) + "   " + label
		lines = append(lines, style.Render(row))
	}

	lines = append(lines,
		"",
		ui.DimStyle.Render(" STACK"),
		ui.HLine(w, ui.Border),
		"",
		stackRow("bubbletea", ui.DepVersion("charm.land/bubbletea/v2", "2.1.1")),
		stackRow("bubbles", ui.DepVersion("charm.land/bubbles/v2", "2.2.1")),
		stackRow("lipgloss", ui.DepVersion("charm.land/lipgloss/v2", "2.0.6")),
	)

	for len(lines) < height-3 {
		lines = append(lines, "")
	}
	lines = append(lines,
		ui.HLine(w, ui.Border),
		ui.SpaceBetween(
			ui.DimStyle.Render("GO "+shortGo(ui.GoVersion())),
			ui.DimStyle.Render(fmt.Sprintf("%d×%d", m.width, m.height)),
			w,
		),
	)

	return ui.Fit(strings.Join(lines, "\n"), w, height)
}

func shortGo(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return version
}

func stackRow(name, version string) string {
	return ui.MutedStyle.Render("  "+name) + strings.Repeat(" ", max(1, sidebarW-2-len(name)-len(version))) +
		ui.DimStyle.Render(version)
}

func (m *Model) pageArea(width, height int) string {
	if m.showHelp {
		return m.helpArea(width, height)
	}

	page := m.pages[m.current]
	title := ui.GradientBold(page.Title(), ui.Pink, ui.Purple, ui.Cyan)
	counter := ui.DimStyle.Render(fmt.Sprintf("%02d / %02d", m.current+1, len(m.pages)))
	header := ui.SpaceBetween(title, counter, width)
	subtitle := ui.MutedStyle.Render(page.Subtitle())
	rule := ui.HLine(width, ui.Border)

	bodyH := height - 4
	if bodyH < 1 {
		bodyH = 1
	}
	body := page.View(width, bodyH)

	return ui.Fit(strings.Join([]string{header, subtitle, rule, "", body}, "\n"), width, height)
}

func (m *Model) footer() string {
	left := "  " +
		ui.Keycap("1-5") + " " + ui.MutedStyle.Render("demos") +
		"   " + ui.Keycap("tab") + " " + ui.MutedStyle.Render("next") +
		"   " + ui.Keycap("?") + " " + ui.MutedStyle.Render("help") +
		"   " + ui.Keycap("q") + " " + ui.MutedStyle.Render("quit")

	right := ""
	if m.width >= 100 {
		right = ui.DimStyle.Render("built with ") +
			ui.Gradient("Bubble Tea", ui.Pink, ui.Purple, ui.Cyan) +
			ui.DimStyle.Render(" + ") +
			ui.Gradient("Lip Gloss", ui.Cyan, ui.Purple, ui.Pink) +
			"  "
	}

	return ui.SpaceBetween(left, right, m.width)
}
