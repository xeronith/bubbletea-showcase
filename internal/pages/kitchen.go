package pages

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/xeronith/bubbletea-showcase/internal/ui"
)

const (
	fieldName = iota
	fieldTea
	fieldStrength
	fieldExtras
	fieldBrew
	fieldCount
)

var teaNames = []string{"Oolong", "Matcha", "Chai"}

// Kitchen is an interactive order form built from Bubbles components. It
// demonstrates focus management, validation, a spinner, and a spring-animated
// progress bar.
type Kitchen struct {
	active bool

	input    textinput.Model
	spinner  spinner.Model
	progress progress.Model

	focus    int
	tea      int
	strength int
	oat      bool

	brewing bool
	done    bool
	run     int
	err     string
}

type kitchenDoneMsg struct{ run int }

// NewKitchen creates the order form.
func NewKitchen() *Kitchen {
	in := textinput.New()
	in.Placeholder = "Luna's midnight blend"
	in.Prompt = "› "
	in.CharLimit = 32
	in.SetWidth(26)

	styles := textinput.DefaultDarkStyles()
	styles.Focused.Prompt = lipgloss.NewStyle().Foreground(ui.Pink)
	styles.Focused.Text = lipgloss.NewStyle().Foreground(ui.Text)
	styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(ui.Dim)
	styles.Cursor.Color = ui.Pink
	in.SetStyles(styles)

	sp := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(ui.Cyan)),
	)

	pr := progress.New(
		progress.WithColors(ui.Pink, ui.Purple, ui.Cyan),
		progress.WithWidth(26),
		progress.WithoutPercentage(),
	)

	return &Kitchen{
		input:    in,
		spinner:  sp,
		progress: pr,
		focus:    fieldTea,
		strength: 3,
	}
}

// Title implements Page.
func (k *Kitchen) Title() string { return "Kitchen" }

// Subtitle implements Page.
func (k *Kitchen) Subtitle() string {
	return "A tiny order form assembled from Bubbles: input, spinner, and spring progress."
}

// Init implements Page.
func (k *Kitchen) Init() tea.Cmd { return nil }

// OnShow implements Page.
func (k *Kitchen) OnShow() tea.Cmd {
	k.active = true
	if k.focus == fieldName {
		return tea.Batch(k.spinner.Tick, k.input.Focus())
	}
	return k.spinner.Tick
}

// OnHide implements Page.
func (k *Kitchen) OnHide() {
	k.active = false
	k.input.Blur()
	if k.brewing {
		k.finishBrew()
	}
}

// Editing implements Page.
func (k *Kitchen) Editing() bool { return k.input.Focused() }

// UsesTab tells the shell that Tab is part of this page's form navigation.
func (k *Kitchen) UsesTab() bool { return true }

// Update implements Page.
func (k *Kitchen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		pageW := msg.Width - 4
		if msg.Width >= 96 {
			pageW = msg.Width - 30
		}
		leftW := pageW * 56 / 100
		if leftW < 42 {
			leftW = 42
		}
		k.input.SetWidth(max(12, leftW-21))
		return nil

	case kitchenDoneMsg:
		if !k.active || !k.brewing || msg.run != k.run {
			return nil
		}
		k.finishBrew()
		return nil

	case spinner.TickMsg:
		if !k.active {
			return nil
		}
		var cmd tea.Cmd
		k.spinner, cmd = k.spinner.Update(msg)
		return cmd

	case progress.FrameMsg:
		if !k.active || !k.brewing {
			return nil
		}
		var cmd tea.Cmd
		k.progress, cmd = k.progress.Update(msg)
		return cmd

	case tea.KeyPressMsg:
		return k.handleKey(msg)
	}

	if k.input.Focused() {
		var cmd tea.Cmd
		k.input, cmd = k.input.Update(msg)
		return cmd
	}
	return nil
}

func (k *Kitchen) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	key := msg.String()

	switch key {
	case "tab":
		return k.moveFocus(1)
	case "shift+tab":
		return k.moveFocus(-1)
	case "esc":
		if k.focus == 0 {
			k.focus = fieldTea
			k.input.Blur()
		}
		return nil
	case "up":
		if k.focus != 0 {
			return k.moveFocus(-1)
		}
	case "down":
		if k.focus != 0 {
			return k.moveFocus(1)
		}
	}

	if k.focus == 0 {
		if key == "enter" {
			k.input.Blur()
			k.focus = fieldTea
			return nil
		}
		if key == "ctrl+u" {
			k.input.SetValue("")
			k.err = ""
		}
		var cmd tea.Cmd
		k.input, cmd = k.input.Update(msg)
		k.err = ""
		return cmd
	}

	switch key {
	case "left", "h":
		k.adjust(-1)
	case "right", "l":
		k.adjust(1)
	case "space", "x":
		if k.focus == fieldExtras {
			k.oat = !k.oat
		}
	case "enter":
		if k.focus == fieldBrew {
			return k.startBrew()
		}
		return k.moveFocus(1)
	case "r":
		return k.reset()
	}
	return nil
}

func (k *Kitchen) moveFocus(delta int) tea.Cmd {
	k.focus = (k.focus + delta + fieldCount) % fieldCount
	if k.focus == 0 {
		k.err = ""
		return k.input.Focus()
	}
	k.input.Blur()
	return nil
}

func (k *Kitchen) adjust(delta int) {
	switch k.focus {
	case fieldTea:
		k.tea = (k.tea + delta + len(teaNames)) % len(teaNames)
	case fieldStrength:
		k.strength = clampInt(k.strength+delta, 1, 5)
	}
}

func (k *Kitchen) startBrew() tea.Cmd {
	if k.brewing {
		return nil
	}
	if strings.TrimSpace(k.input.Value()) == "" {
		k.err = "give your blend a name first"
		k.focus = 0
		return k.input.Focus()
	}

	k.brewing = true
	k.done = false
	k.err = ""
	k.run++
	run := k.run
	k.input.Blur()

	return tea.Batch(
		k.progress.SetPercent(1),
		tea.Tick(1900*time.Millisecond, func(time.Time) tea.Msg {
			return kitchenDoneMsg{run: run}
		}),
	)
}

func (k *Kitchen) finishBrew() {
	k.brewing = false
	k.done = true
}

func (k *Kitchen) reset() tea.Cmd {
	k.brewing = false
	k.done = false
	k.err = ""
	k.run++
	k.tea = 0
	k.strength = 3
	k.oat = false
	k.focus = fieldTea
	k.input.SetValue("")
	k.input.Blur()
	return k.progress.SetPercent(0)
}

func (k *Kitchen) orderName() string {
	name := strings.TrimSpace(k.input.Value())
	if name == "" {
		return "House blend"
	}
	return name
}

func (k *Kitchen) price() float64 {
	p := 4.50 + float64(k.strength)*0.35
	if k.oat {
		p += 0.60
	}
	return p
}

// View implements Page.
func (k *Kitchen) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	wide := width >= 76 && height >= 14
	if !wide {
		return ui.Fit(k.formCard(width, height), width, height)
	}

	leftW := width * 56 / 100
	if leftW < 42 {
		leftW = 42
	}
	rightW := width - leftW - 1
	if rightW < 28 {
		return ui.Fit(k.formCard(width, height), width, height)
	}

	if height >= 28 {
		topH := 20
		top := lipgloss.JoinHorizontal(lipgloss.Top,
			k.formCard(leftW, topH), " ", k.sideColumn(rightW, topH))
		return ui.Fit(lipgloss.JoinVertical(lipgloss.Left,
			top, "", k.conceptCard(width, height-topH-1)), width, height)
	}

	left := k.formCard(leftW, height)
	right := k.sideColumn(rightW, height)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
}

func (k *Kitchen) conceptCard(width, height int) string {
	bodyW := ui.CardBodyWidth(width)
	row := func(state, typ, keys string) string {
		return "  " +
			ui.MutedStyle.Render(fmt.Sprintf("%-10s", state)) +
			ui.DimStyle.Render(fmt.Sprintf("%-24s", typ)) +
			lipgloss.NewStyle().Foreground(ui.Pink).Render(keys)
	}
	rows := []string{
		row("name", "textinput.Model", "tab"),
		row("tea", "int · 3 options", "← →"),
		row("strength", "int · 1..5", "← →"),
		row("extras", "bool", "space"),
	}
	footer := ui.DimStyle.Render("  enter fires a tea.Batch of spinner, spring progress, and timer commands.")
	body := strings.Join(rows, "\n") + "\n\n" + ui.Clip(footer, bodyW)
	return ui.Card(ui.Cyan, "state under the hood", body, width, height)
}

func (k *Kitchen) formCard(width, height int) string {
	bodyW := ui.CardBodyWidth(width)

	lines := []string{}
	if height >= 20 {
		lines = append(lines,
			ui.DimStyle.Render("Build a cup. Every field is a Bubble."),
			"",
		)
	}

	lines = append(lines,
		k.fieldLine(fieldName, "Name", k.input.View(), bodyW),
		"",
		k.fieldLine(fieldTea, "Tea", k.teaOptions(bodyW), bodyW),
		"",
		k.fieldLine(fieldStrength, "Strength", k.strengthView(), bodyW),
		"",
		k.fieldLine(fieldExtras, "Extras", k.extrasView(), bodyW),
		"",
		k.actionView(),
	)

	if k.brewing {
		lines = append(lines, "  "+k.progress.View())
	}
	if height >= 22 {
		lines = append(lines, "", ui.DimStyle.Render("  Cursor by Bubbles, loop by Tea."))
	}

	return ui.Card(ui.Pink, "order desk", strings.Join(lines, "\n"), width, height)
}

func (k *Kitchen) fieldLine(field int, label, value string, width int) string {
	focused := k.focus == field && !k.brewing

	gutter := "  "
	labelStyle := ui.MutedStyle
	if focused {
		gutter = lipgloss.NewStyle().Foreground(ui.Pink).Render("▸ ")
		labelStyle = lipgloss.NewStyle().Foreground(ui.Text).Bold(true)
	}

	labelText := fmt.Sprintf("%-10s", label)
	prefix := gutter + labelStyle.Render(labelText)
	return ui.Clip(prefix+value, width)
}

func (k *Kitchen) teaOptions(width int) string {
	parts := make([]string, 0, len(teaNames))
	for i, name := range teaNames {
		if i == k.tea {
			parts = append(parts, lipgloss.NewStyle().Foreground(ui.Pink).Bold(true).Render("● "+name))
		} else {
			parts = append(parts, ui.DimStyle.Render("○ ")+ui.MutedStyle.Render(name))
		}
	}
	return ui.Clip(strings.Join(parts, "  "), width-13)
}

func (k *Kitchen) strengthView() string {
	return ui.Bar(float64(k.strength)/5, 10, ui.Pink, ui.Purple, ui.Cyan) +
		ui.MutedStyle.Render(fmt.Sprintf("  %d/5", k.strength))
}

func (k *Kitchen) extrasView() string {
	box, label := "[ ]", ui.MutedStyle.Render("oat milk")
	if k.oat {
		box = lipgloss.NewStyle().Foreground(ui.Green).Bold(true).Render("[x]")
		label = ui.TextStyle.Render("oat milk")
	}
	return box + " " + label + ui.DimStyle.Render("   +$0.60")
}

func (k *Kitchen) actionView() string {
	if k.brewing {
		return "  " + k.spinner.View() + " " +
			lipgloss.NewStyle().Foreground(ui.Cyan).Render("steeping your cup…")
	}
	if k.done {
		return "  " + lipgloss.NewStyle().Foreground(ui.Green).Bold(true).Render("✓ ready") +
			ui.DimStyle.Render("   press r to reset")
	}
	if k.err != "" {
		return "  " + lipgloss.NewStyle().Foreground(ui.Red).Render("! "+k.err)
	}

	button := ui.Keycap("enter") + " " + ui.MutedStyle.Render("brew")
	if k.focus == fieldBrew {
		button = ui.Keycap("enter") + " " + lipgloss.NewStyle().Foreground(ui.Pink).Bold(true).Render("brew")
	}
	return "  " + button + ui.DimStyle.Render("   tab · space")
}

func (k *Kitchen) sideColumn(width, height int) string {
	receiptH := 11
	if height < 13 {
		receiptH = max(6, height)
	}
	receipt := k.receiptCard(width, receiptH)

	if height-receiptH-1 < 7 {
		return ui.Fit(receipt, width, height)
	}
	specimen := k.specimenCard(width, height-receiptH-1)
	return lipgloss.JoinVertical(lipgloss.Left, receipt, "", specimen)
}

func (k *Kitchen) receiptCard(width, height int) string {
	bodyW := ui.CardBodyWidth(width)
	var lines []string

	lines = append(lines, ui.SpaceBetween(
		ui.DimStyle.Render("ORDER"),
		ui.DimStyle.Render(fmt.Sprintf("#%03d", 41+k.run)),
		bodyW,
	))
	lines = append(lines, ui.HLine(bodyW, ui.Border), "")

	switch {
	case k.done:
		lines = append(lines,
			ui.TextStyle.Bold(true).Render("1× "+k.orderName()),
			ui.MutedStyle.Render(teaNames[k.tea]+fmt.Sprintf(" · strength %d/5", k.strength)),
			ui.MutedStyle.Render(k.milkLabel()),
			"",
			lipgloss.NewStyle().Foreground(ui.Green).Bold(true).Render(fmt.Sprintf("$%.2f", k.price())),
		)
	case k.brewing:
		lines = append(lines,
			k.spinner.View()+" "+lipgloss.NewStyle().Foreground(ui.Cyan).Render("brewing…"),
			"",
			k.progress.View(),
			"",
			ui.DimStyle.Render(teaNames[k.tea]+fmt.Sprintf(" · %d/5 · %s", k.strength, k.milkLabel())),
		)
	default:
		lines = append(lines,
			ui.DimStyle.Render("Nothing brewing yet."),
			"",
			ui.MutedStyle.Render("Pick a tea, dial in the"),
			ui.MutedStyle.Render("strength, then press enter."),
			"",
			ui.DimStyle.Render("Estimated "+fmt.Sprintf("$%.2f", k.price())),
		)
	}

	return ui.Card(ui.Yellow, "receipt", strings.Join(lines, "\n"), width, height)
}

func (k *Kitchen) specimenCard(width, height int) string {
	bodyW := ui.CardBodyWidth(width)
	lines := []string{
		k.spinner.View() + " " + ui.MutedStyle.Render("spinner.Dot"),
		"",
		ui.Bar(0.64, max(6, bodyW-9), ui.Pink, ui.Purple, ui.Cyan) +
			ui.DimStyle.Render("  64%"),
		"",
		ui.Keycap("tab") + " " + ui.Keycap("space") + " " + ui.DimStyle.Render("input primitives"),
	}
	return ui.Card(ui.Purple, "component lab", strings.Join(lines, "\n"), width, height)
}

func (k *Kitchen) milkLabel() string {
	if k.oat {
		return "oat milk"
	}
	return "whole milk"
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
