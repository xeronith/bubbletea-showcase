package app

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/xeronith/bubbletea-showcase/internal/ui"
)

type helpGroup struct {
	title string
	rows  [][2]string
}

func (m *Model) helpArea(width, height int) string {
	if height < 20 {
		return m.helpCompact(width, height)
	}

	navigate := helpGroup{"Navigate", [][2]string{
		{"1 – 5", "jump to a demo"},
		{"tab", "next demo"},
		{"shift+tab", "previous demo"},
		{"?", "toggle this card"},
		{"q", "quit"},
	}}
	kitchen := helpGroup{"Kitchen", [][2]string{
		{"tab", "next field"},
		{"← →", "adjust a choice"},
		{"enter", "brew the order"},
		{"r", "reset the form"},
	}}
	gallery := helpGroup{"Gallery", [][2]string{
		{"↑ ↓ / j k", "move the cursor"},
		{"/", "search projects"},
		{"esc", "clear the search"},
		{"enter", "star a project"},
	}}
	motion := helpGroup{"Motion", [][2]string{
		{"space", "pause the wave"},
		{"m", "change the wave"},
		{"+ / -", "change speed"},
	}}
	queue := helpGroup{"Brew queue", [][2]string{
		{"enter", "run every job"},
		{"r", "reset the queue"},
	}}

	col1 := lipgloss.JoinVertical(lipgloss.Left, renderHelpGroup(navigate), "", renderHelpGroup(kitchen))
	col2 := lipgloss.JoinVertical(lipgloss.Left,
		renderHelpGroup(gallery), "", renderHelpGroup(motion), "", renderHelpGroup(queue))

	var body string
	cardW := min(width-2, 80)
	if cardW >= 66 {
		body = lipgloss.JoinHorizontal(lipgloss.Top, col1, "     ", col2)
	} else {
		body = lipgloss.JoinVertical(lipgloss.Left,
			renderHelpGroup(navigate), "", renderHelpGroup(kitchen), "",
			renderHelpGroup(gallery), "", renderHelpGroup(motion), "", renderHelpGroup(queue))
	}

	hint := ui.DimStyle.Render("press any key to close")

	cardH := lipgloss.Height(body) + 5
	if cardH > height {
		cardH = height
	}
	card := ui.GlowCard("keyboard shortcuts", body+"\n\n"+hint, cardW, cardH)

	panel := lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, card)
	return ui.Fit(panel, width, height)
}

func (m *Model) helpCompact(width, height int) string {
	row := func(pairs ...string) string {
		var b strings.Builder
		for i := 0; i+1 < len(pairs); i += 2 {
			if i > 0 {
				b.WriteString("   ")
			}
			b.WriteString(lipgloss.NewStyle().Foreground(ui.Pink).Render(pairs[i]))
			b.WriteString(" ")
			b.WriteString(ui.MutedStyle.Render(pairs[i+1]))
		}
		return " " + b.String()
	}

	body := strings.Join([]string{
		ui.GradientBold("Navigate"),
		row("1-5", "demos", "tab", "next", "?", "help", "q", "quit"),
		"",
		ui.GradientBold("Kitchen"),
		row("tab", "field", "←→", "adjust", "enter", "brew"),
		"",
		ui.GradientBold("Gallery"),
		row("↑↓", "move", "/", "search", "enter", "star", "esc", "clear"),
		"",
		ui.GradientBold("Motion + Queue"),
		row("space", "pause", "m", "wave", "+-", "speed", "enter", "run"),
	}, "\n")

	cardW := min(width-2, 78)
	cardH := min(height, lipgloss.Height(body)+4)
	card := ui.GlowCard("keyboard shortcuts", body, cardW, cardH)
	panel := lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, card)
	return ui.Fit(panel, width, height)
}

func renderHelpGroup(g helpGroup) string {
	lines := []string{ui.GradientBold(g.title)}
	for _, row := range g.rows {
		key := fmt.Sprintf("%-11s", row[0])
		lines = append(lines,
			"  "+lipgloss.NewStyle().Foreground(ui.Pink).Render(key)+
				ui.MutedStyle.Render(row[1]),
		)
	}
	return strings.Join(lines, "\n")
}
