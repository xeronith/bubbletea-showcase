package ui

import (
	"image/color"
	"math"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
)

// Gradient renders each rune of text with a color sampled from the given
// stops. It is intended for short, plain-text labels and headings.
func Gradient(text string, stops ...color.Color) string {
	if text == "" {
		return ""
	}
	if len(stops) == 0 {
		stops = GradientStops
	}
	runes := []rune(text)
	colors := lipgloss.Blend1D(len(runes), stops...)
	var b strings.Builder
	b.Grow(len(text) * 8)
	for i, r := range runes {
		b.WriteString(lipgloss.NewStyle().Foreground(colors[i]).Render(string(r)))
	}
	return b.String()
}

// GradientBold is Gradient with weight.
func GradientBold(text string, stops ...color.Color) string {
	if text == "" {
		return ""
	}
	if len(stops) == 0 {
		stops = GradientStops
	}
	runes := []rune(text)
	colors := lipgloss.Blend1D(len(runes), stops...)
	var b strings.Builder
	b.Grow(len(text) * 8)
	for i, r := range runes {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colors[i]).Render(string(r)))
	}
	return b.String()
}

// Card renders a rounded panel with a small accent-tinted title. Width and
// height include borders and padding.
func Card(accent color.Color, title, body string, width, height int) string {
	if width < 8 || height < 3 {
		return ""
	}

	content := body
	if title != "" {
		bar := lipgloss.NewStyle().Foreground(accent).Render("▌")
		label := lipgloss.NewStyle().
			Foreground(accent).
			Bold(true).
			Render(strings.ToUpper(title))
		heading := bar + " " + label
		if content == "" {
			content = heading
		} else {
			content = heading + "\n" + content
		}
	}

	inner := panelBody(content, width, height)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(0, 2).
		Width(width).
		Height(height).
		Render(inner)
}

// GlowCard is a Card whose border carries the signature gradient. Use it
// sparingly for hero content.
func GlowCard(title, body string, width, height int) string {
	if width < 8 || height < 3 {
		return ""
	}

	content := body
	if title != "" {
		label := GradientBold(strings.ToUpper(title))
		bar := Gradient("▌")
		heading := bar + " " + label
		if content == "" {
			content = heading
		} else {
			content = heading + "\n" + content
		}
	}

	inner := panelBody(content, width, height)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForegroundBlend(GradientStops...).
		Padding(0, 2).
		Width(width).
		Height(height).
		Render(inner)
}

// panelBody wraps content to the inner width and trims it to the inner height
// so that a card's bottom border is never lost to overflow.
func panelBody(content string, width, height int) string {
	innerW := width - 6 // border + horizontal padding
	innerH := height - 2
	if innerW < 1 || innerH < 1 {
		return ""
	}
	wrapped := lipgloss.NewStyle().Width(innerW).Render(content)
	return TruncateLines(wrapped, innerH)
}

// TruncateLines keeps at most n lines of a block.
func TruncateLines(s string, n int) string {
	if n <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[:n], "\n")
}

// CardBodyWidth returns the usable inner width of a Card of the given width.
func CardBodyWidth(width int) int {
	return width - 6 // 2 border + 4 horizontal padding
}

// Bar renders a smooth, gradient-filled meter.
func Bar(fraction float64, width int, stops ...color.Color) string {
	if width <= 0 {
		return ""
	}
	if len(stops) == 0 {
		stops = GradientStops
	}
	fraction = clamp01(fraction)
	filled := int(math.Round(fraction * float64(width)))
	if filled > width {
		filled = width
	}

	colors := lipgloss.Blend1D(width, stops...)
	var b strings.Builder
	b.Grow(width * 12)
	for i := 0; i < width; i++ {
		if i < filled {
			b.WriteString(lipgloss.NewStyle().Foreground(colors[i]).Render("█"))
		} else {
			b.WriteString(DimStyle.Render("░"))
		}
	}
	return b.String()
}

// Area renders a filled area chart. Values are expected in the range 0..1 and
// are sampled to the requested width.
func Area(values []float64, width, height int, stops ...color.Color) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	if len(stops) == 0 {
		stops = GradientStops
	}

	colors := lipgloss.Blend1D(width, stops...)
	blocks := make([]string, width)
	for x := 0; x < width; x++ {
		blocks[x] = lipgloss.NewStyle().Foreground(colors[x]).Render("█")
	}

	heights := make([]int, width)
	for x := 0; x < width; x++ {
		v := sample(values, x, width)
		heights[x] = int(math.Round(clamp01(v) * float64(height)))
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

func sample(values []float64, x, width int) float64 {
	if len(values) == 0 || width <= 0 {
		return 0
	}
	if len(values) == 1 {
		return values[0]
	}
	pos := float64(x) / float64(max(1, width-1)) * float64(len(values)-1)
	lo := int(math.Floor(pos))
	hi := min(lo+1, len(values)-1)
	frac := pos - float64(lo)
	return values[lo]*(1-frac) + values[hi]*frac
}

// Keycap renders a key hint with a subtle filled background.
func Keycap(text string) string {
	return lipgloss.NewStyle().
		Foreground(Text).
		Background(SurfaceHi).
		Bold(true).
		Padding(0, 1).
		Render(text)
}

// Badge renders a small, quiet label.
func Badge(text string, fg color.Color) string {
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(SurfaceAlt).
		Padding(0, 1).
		Render(text)
}

// SpaceBetween pins left and right to the edges of a line of the given width.
func SpaceBetween(left, right string, width int) string {
	lw := lipgloss.Width(left)
	rw := lipgloss.Width(right)
	gap := width - lw - rw
	if gap >= 1 {
		return left + strings.Repeat(" ", gap) + right
	}

	// Not enough room for both: keep the left side and clip the right.
	if lw+1 >= width {
		return Clip(left, width)
	}
	return left + " " + Clip(right, width-lw-1)
}

// Clip truncates a single line to width columns, preserving ANSI styling.
func Clip(s string, width int) string {
	if width <= 0 {
		return ""
	}
	return lipgloss.NewStyle().MaxWidth(width).Render(s)
}

// Wrap wraps text to the given width.
func Wrap(s string, width int) string {
	if width <= 0 {
		return ""
	}
	return lipgloss.NewStyle().Width(width).Render(s)
}

// Fit pads or clips a block to exactly width × height.
func Fit(s string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		MaxWidth(width).
		MaxHeight(height).
		Render(s)
}

// HLine renders a horizontal rule.
func HLine(width int, c color.Color) string {
	if width <= 0 {
		return ""
	}
	return lipgloss.NewStyle().Foreground(c).Render(strings.Repeat("─", width))
}

// LogoLines is the compact, three-row wordmark.
var LogoLines = []string{
	"█▀▀▄ █  █ █▀▀▄ █▀▀▄ █    █▀▀▀  ▀█▀ █▀▀▀ █▀▀█",
	"█▀▀▄ █  █ █▀▀▄ █▀▀▄ █    █▀▀    █  █▀▀  █▄▄█",
	"█▄▄▀ █▄▄█ █▄▄▀ █▄▄▀ █▄▄▄ █▄▄▄   █  █▄▄▄ █  █",
}

// LogoWidth is the display width of LogoLines.
const LogoWidth = 45

// Logo renders the wordmark with the signature gradient.
func Logo() string {
	lines := make([]string, len(LogoLines))
	for i, line := range LogoLines {
		lines[i] = GradientBold(line)
	}
	return strings.Join(lines, "\n")
}

// Teacup frames keep the steam gently drifting.
var teacupFrames = [][]string{
	{
		"     ( (",
		"      ) )",
	},
	{
		"      ( (",
		"       ) )",
	},
	{
		"     ) (",
		"    ( (",
	},
	{
		"      ( (",
		"      ) )",
	},
}

// Teacup renders a small cup with animated steam.
func Teacup(frame int) string {
	if len(teacupFrames) == 0 {
		return ""
	}
	steam := teacupFrames[frame%len(teacupFrames)]
	cup := []string{
		"  ╭─────────╮",
		"  │ " + lipgloss.NewStyle().Foreground(Cyan).Render("≈≈≈≈≈≈≈") + " │",
		"  ╰──┬───┬──╯",
		"     ╰───╯",
	}
	lines := make([]string, 0, len(steam)+len(cup))
	steamStyle := lipgloss.NewStyle().Foreground(Cyan)
	for _, s := range steam {
		lines = append(lines, steamStyle.Render(s))
	}
	cupStyle := lipgloss.NewStyle().Foreground(Pink)
	for i, c := range cup {
		if i == 1 {
			// The tea surface carries its own color.
			lines = append(lines, c)
			continue
		}
		lines = append(lines, cupStyle.Render(c))
	}
	return strings.Join(lines, "\n")
}

// TeacupWidth is the display width of Teacup output.
const TeacupWidth = 15

func clamp01(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}

// TooSmall renders a friendly message for terminals below the minimum size.
func TooSmall(width, height, minWidth, minHeight int) string {
	card := Card(Pink, "One moment",
		" This demo needs a little more room.\n\n"+
			" Resize the window to at least\n"+
			" "+Badge(strconv.Itoa(minWidth)+" × "+strconv.Itoa(minHeight), Cyan)+" for the full experience.\n\n"+
			" Current: "+MutedStyle.Render(strconv.Itoa(width)+" × "+strconv.Itoa(height)),
		46, 11)
	return lipgloss.Place(max(width, 1), max(height, 1), lipgloss.Center, lipgloss.Center, card)
}
