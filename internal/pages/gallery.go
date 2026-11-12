package pages

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/xeronith/bubbletea-showcase/internal/ui"
)

type project struct {
	name     string
	author   string
	category string
	desc     string
	url      string
}

var catalog = []project{
	{"chezmoi", "twpayne", "dotfiles", "Manage dotfiles across multiple machines, securely.", "github.com/twpayne/chezmoi"},
	{"Glow", "charmbracelet", "markdown", "Render markdown in the terminal and stash it online.", "github.com/charmbracelet/glow"},
	{"Huh?", "charmbracelet", "forms", "A toolkit for interactive forms and prompts.", "github.com/charmbracelet/huh"},
	{"Mods", "charmbracelet", "ai", "AI on the CLI, built for pipelines.", "github.com/charmbracelet/mods"},
	{"gh-dash", "dlvhdr", "git", "A GitHub CLI extension for pull requests and issues.", "github.com/dlvhdr/gh-dash"},
	{"Tetrigo", "Broderick-Westrope", "games", "Tetris, in your terminal.", "github.com/Broderick-Westrope/tetrigo"},
	{"Superfile", "yorukot", "files", "A modern terminal file manager.", "github.com/yorukot/superfile"},
	{"Circumflex", "bensadeh", "news", "Read Hacker News without leaving the terminal.", "github.com/bensadeh/circumflex"},
	{"Soft Serve", "charmbracelet", "git", "A self-hostable Git server for the command line.", "github.com/charmbracelet/soft-serve"},
	{"Wish", "charmbracelet", "ssh", "Make SSH apps, just like that.", "github.com/charmbracelet/wish"},
	{"Pop", "charmbracelet", "email", "Send emails from your terminal.", "github.com/charmbracelet/pop"},
	{"Slides", "maaslalani", "talks", "Presentations in your terminal.", "github.com/maaslalani/slides"},
	{"Signls", "emprcl", "music", "A generative MIDI sequencer for composition and live performance.", "github.com/emprcl/signls"},
	{"eks-node-viewer", "awslabs", "infra", "Visualize dynamic node usage within an EKS cluster.", "github.com/awslabs/eks-node-viewer"},
	{"container-canary", "NVIDIA", "infra", "A container validator for Kubernetes.", "github.com/NVIDIA/container-canary"},
	{"mc", "minio", "storage", "The official MinIO client.", "github.com/minio/mc"},
	{"aztify", "Azure", "infra", "Bring Azure resources under Terraform.", "github.com/Azure/aztfy"},
	{"Daytona", "daytonaio", "devtools", "An AI infrastructure platform.", "github.com/daytonaio/daytona"},
	{"TruffleHog", "trufflesecurity", "security", "Find leaked credentials.", "github.com/trufflesecurity/trufflehog"},
	{"CockroachDB", "cockroachdb", "database", "A cloud-native, distributed SQL database.", "github.com/cockroachdb/cockroach"},
}

var categoryColors = map[string]color.Color{
	"dotfiles": ui.Yellow,
	"markdown": ui.Orange,
	"forms":    ui.Purple,
	"ai":       ui.Magenta,
	"git":      ui.Orange,
	"games":    ui.Green,
	"files":    ui.Cyan,
	"news":     ui.Yellow,
	"ssh":      ui.Cyan,
	"email":    ui.Pink,
	"talks":    ui.Green,
	"music":    ui.Magenta,
	"infra":    ui.Cyan,
	"storage":  ui.Yellow,
	"devtools": ui.Purple,
	"security": ui.Red,
	"database": ui.Green,
}

// Gallery is a fuzzy-filtered catalog of real applications built with Bubble
// Tea. It demonstrates list state, custom filtering, and keyboard-first UI.
type Gallery struct {
	width  int
	height int

	cursor    int
	query     string
	searching bool
	filtered  []int
	starred   map[int]bool
}

// NewGallery creates the gallery page.
func NewGallery() *Gallery {
	g := &Gallery{starred: make(map[int]bool)}
	g.applyFilter()
	return g
}

// Title implements Page.
func (g *Gallery) Title() string { return "Gallery" }

// Subtitle implements Page.
func (g *Gallery) Subtitle() string {
	return "Twenty real Bubble Tea apps, with fuzzy search and a keyboard-only cursor."
}

// Init implements Page.
func (g *Gallery) Init() tea.Cmd { return nil }

// OnShow implements Page.
func (g *Gallery) OnShow() tea.Cmd { return nil }

// OnHide implements Page.
func (g *Gallery) OnHide() { g.searching = false }

// Editing implements Page.
func (g *Gallery) Editing() bool { return g.searching }

// Update implements Page.
func (g *Gallery) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		g.width, g.height = msg.Width, msg.Height

	case tea.KeyPressMsg:
		g.handleKey(msg)
	}
	return nil
}

func (g *Gallery) handleKey(msg tea.KeyPressMsg) {
	key := msg.String()

	if g.searching {
		switch key {
		case "esc":
			g.searching = false
			g.query = ""
			g.cursor = 0
			g.applyFilter()
		case "enter":
			g.searching = false
		case "up", "ctrl+p":
			g.move(-1)
		case "down", "ctrl+n":
			g.move(1)
		case "backspace":
			if g.query != "" {
				runes := []rune(g.query)
				g.query = string(runes[:len(runes)-1])
				g.applyFilter()
			}
		case "ctrl+u":
			g.query = ""
			g.cursor = 0
			g.applyFilter()
		default:
			if msg.Text != "" {
				g.query += msg.Text
				g.applyFilter()
			}
		}
		return
	}

	switch key {
	case "/":
		g.searching = true
	case "esc":
		if g.query != "" {
			g.query = ""
			g.cursor = 0
			g.applyFilter()
		}
	case "up", "k":
		g.move(-1)
	case "down", "j":
		g.move(1)
	case "pgup":
		g.move(-6)
	case "pgdown", " ":
		g.move(6)
	case "home", "g":
		g.cursor = 0
	case "end", "G":
		if len(g.filtered) > 0 {
			g.cursor = len(g.filtered) - 1
		}
	case "enter":
		if len(g.filtered) > 0 {
			idx := g.filtered[g.cursor]
			g.starred[idx] = !g.starred[idx]
		}
	}
}

func (g *Gallery) move(delta int) {
	if len(g.filtered) == 0 {
		g.cursor = 0
		return
	}
	g.cursor = clampInt(g.cursor+delta, 0, len(g.filtered)-1)
}

func (g *Gallery) applyFilter() {
	g.filtered = g.filtered[:0]
	q := strings.ToLower(strings.TrimSpace(g.query))

	for i, p := range catalog {
		if q == "" ||
			fuzzyMatch(strings.ToLower(p.name), q) ||
			fuzzyMatch(strings.ToLower(p.author), q) ||
			fuzzyMatch(strings.ToLower(p.category), q) {
			g.filtered = append(g.filtered, i)
		}
	}
	g.cursor = clampInt(g.cursor, 0, max(0, len(g.filtered)-1))
}

func fuzzyMatch(haystack, query string) bool {
	if query == "" {
		return true
	}
	needle := []rune(query)
	if len(needle) == 0 {
		return true
	}
	j := 0
	for _, r := range haystack {
		if r == needle[j] {
			j++
			if j == len(needle) {
				return true
			}
		}
	}
	return false
}

// View implements Page.
func (g *Gallery) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	wide := width >= 78 && height >= 16
	if !wide {
		return ui.Fit(g.listCard(width, height), width, height)
	}

	leftW := width * 58 / 100
	rightW := width - leftW - 1
	if rightW < 28 {
		return ui.Fit(g.listCard(width, height), width, height)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top,
		g.listCard(leftW, height), " ", g.detailCard(rightW, height))
}

func (g *Gallery) listCard(width, height int) string {
	bodyW := ui.CardBodyWidth(width)
	innerH := height - 2
	if bodyW < 10 {
		return ""
	}

	search := g.searchLine(bodyW)
	footer := ui.SpaceBetween(
		ui.DimStyle.Render("↑↓ move · enter star · / search"),
		ui.DimStyle.Render(fmt.Sprintf("★ %d", g.starCount())),
		bodyW,
	)

	available := innerH - 4 // title, search rule, footer
	if available < 2 {
		available = 2
	}
	perPage := max(1, (available+1)/3)
	start := g.cursor - perPage/2
	if start < 0 {
		start = 0
	}
	if start > len(g.filtered)-perPage {
		start = len(g.filtered) - perPage
	}
	if start < 0 {
		start = 0
	}

	var rows []string
	if len(g.filtered) == 0 {
		rows = append(rows,
			"",
			"  "+ui.DimStyle.Render("No projects match “"+g.query+"”."),
			"  "+ui.DimStyle.Render("Press esc to clear the filter."),
		)
	} else {
		for i := start; i < len(g.filtered) && i < start+perPage; i++ {
			rows = append(rows, g.itemLines(g.filtered[i], i == g.cursor, bodyW)...)
			if i < start+perPage-1 && i < len(g.filtered)-1 {
				rows = append(rows, "")
			}
		}
	}

	filler := innerH - 4 - len(rows)
	if filler < 0 {
		filler = 0
	}
	body := search + "\n" + ui.HLine(bodyW, ui.Border) + "\n" + strings.Join(rows, "\n")
	if filler > 0 {
		body += strings.Repeat("\n", filler)
	}
	body += "\n" + footer

	return ui.Card(ui.Cyan, "projects", body, width, height)
}

func (g *Gallery) searchLine(width int) string {
	var left string
	switch {
	case g.searching:
		left = lipgloss.NewStyle().Foreground(ui.Pink).Bold(true).Render("/ ") +
			ui.TextStyle.Render(g.query) +
			lipgloss.NewStyle().Foreground(ui.Pink).Render("▌") +
			ui.DimStyle.Render("  esc cancel")
	case g.query != "":
		left = lipgloss.NewStyle().Foreground(ui.Pink).Render("/ ") +
			ui.TextStyle.Render(g.query) +
			ui.DimStyle.Render("  esc clear")
	default:
		left = ui.DimStyle.Render("/  search projects")
	}
	right := ui.DimStyle.Render(fmt.Sprintf("%d results", len(g.filtered)))
	return ui.SpaceBetween(left, right, width)
}

func (g *Gallery) itemLines(index int, selected bool, width int) []string {
	p := catalog[index]

	bar := "  "
	titleStyle := lipgloss.NewStyle().Foreground(ui.Text)
	descStyle := ui.DimStyle
	if selected {
		bar = lipgloss.NewStyle().Foreground(ui.Pink).Render("▌") + " "
		titleStyle = lipgloss.NewStyle().Foreground(ui.Pink).Bold(true)
		descStyle = ui.MutedStyle
	}

	star := ""
	if g.starred[index] {
		star = lipgloss.NewStyle().Foreground(ui.Yellow).Render(" ★")
	}

	title := highlightMatch(p.name, g.query, titleStyle)
	right := ui.Badge(p.category, categoryColor(p.category))
	line1 := bar + ui.SpaceBetween(title+star, right, width-2)
	line2 := "  " + descStyle.Render(p.author+" · "+p.category)
	return []string{line1, ui.Clip(line2, width)}
}

func (g *Gallery) detailCard(width, height int) string {
	bodyW := ui.CardBodyWidth(width)
	innerH := height - 2
	if bodyW < 10 {
		return ""
	}

	if len(g.filtered) == 0 {
		return ui.Card(ui.Purple, "details", ui.DimStyle.Render("Nothing selected."), width, height)
	}

	p := catalog[g.filtered[g.cursor]]
	lines := []string{
		ui.Badge(p.category, categoryColor(p.category)),
		ui.GradientBold(p.name, ui.Pink, ui.Purple, ui.Cyan),
		ui.MutedStyle.Render("by " + p.author),
		ui.HLine(bodyW, ui.Border),
		ui.Wrap(ui.TextStyle.Render(p.desc), bodyW),
		"",
		lipgloss.NewStyle().Foreground(ui.Cyan).Render("↗ " + p.url),
	}

	if innerH >= len(lines)+4 {
		state := "enter to star"
		if g.starred[g.filtered[g.cursor]] {
			state = "★ starred · enter to unstar"
		}
		lines = append(lines, "", ui.DimStyle.Render(state))
	}

	if innerH >= len(lines)+6 {
		lines = append(lines,
			"",
			ui.HLine(bodyW, ui.Border),
			ui.DimStyle.Render("THE FRAMEWORK"),
			ui.MutedStyle.Render("Bubble Tea · 45k+ stars"),
			ui.MutedStyle.Render("18,000+ apps in the wild"),
		)
	}

	return ui.Card(ui.Purple, "details", strings.Join(lines, "\n"), width, height)
}

func (g *Gallery) starCount() int {
	n := 0
	for _, v := range g.starred {
		if v {
			n++
		}
	}
	return n
}

func categoryColor(category string) color.Color {
	if c, ok := categoryColors[category]; ok {
		return c
	}
	return ui.Muted
}

func highlightMatch(text, query string, base lipgloss.Style) string {
	if query == "" {
		return base.Render(text)
	}
	runes := []rune(text)
	needle := []rune(strings.ToLower(query))
	j := 0
	var b strings.Builder
	for _, r := range runes {
		if j < len(needle) && strings.ToLower(string(r)) == string(needle[j]) {
			b.WriteString(base.Underline(true).Foreground(ui.Yellow).Render(string(r)))
			j++
			continue
		}
		b.WriteString(base.Render(string(r)))
	}
	return b.String()
}
