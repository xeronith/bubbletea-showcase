package pages

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/xeronith/bubbletea-showcase/internal/ui"
)

type jobSpec struct {
	name     string
	duration time.Duration
	startAt  time.Duration
	accent   color.Color
}

type jobState struct {
	progress float64
	started  bool
	done     bool
}

type queueEvent struct {
	at     time.Duration
	text   string
	accent color.Color
}

type jobTickMsg struct {
	run     int
	index   int
	elapsed time.Duration
}

type queueClockMsg struct {
	run int
	at  time.Time
}

var queueJobs = []jobSpec{
	{"Warm the kettle", 1100 * time.Millisecond, 0, ui.Orange},
	{"Steep the oolong", 2200 * time.Millisecond, 250 * time.Millisecond, ui.Yellow},
	{"Froth the oat milk", 1500 * time.Millisecond, 450 * time.Millisecond, ui.Pink},
	{"Add tapioca pearls", 1800 * time.Millisecond, 700 * time.Millisecond, ui.Purple},
	{"Seal and serve", 1000 * time.Millisecond, 950 * time.Millisecond, ui.Green},
}

var queueSpinner = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Queue demonstrates concurrent commands: five jobs, five tick streams, one
// tea.Batch. Nothing blocks the UI thread.
type Queue struct {
	active bool

	jobs    []jobState
	running bool

	run     int
	started time.Time
	elapsed time.Duration
	everRun bool
	events  []queueEvent
}

// NewQueue creates the job queue page.
func NewQueue() *Queue {
	return &Queue{jobs: make([]jobState, len(queueJobs))}
}

// Title implements Page.
func (q *Queue) Title() string { return "Brew Queue" }

// Subtitle implements Page.
func (q *Queue) Subtitle() string {
	return "Five async jobs, five tick streams, one tea.Batch. The UI never blocks."
}

// Init implements Page.
func (q *Queue) Init() tea.Cmd { return nil }

// OnShow implements Page.
func (q *Queue) OnShow() tea.Cmd {
	q.active = true
	if !q.everRun {
		return q.startAll()
	}
	return nil
}

// OnHide implements Page.
func (q *Queue) OnHide() {
	q.active = false
	if q.running {
		q.reset()
	}
}

// Editing implements Page.
func (q *Queue) Editing() bool { return false }

func (q *Queue) allDone() bool {
	for _, j := range q.jobs {
		if !j.done {
			return false
		}
	}
	return true
}

func (q *Queue) doneCount() int {
	n := 0
	for _, j := range q.jobs {
		if j.done {
			n++
		}
	}
	return n
}

func (q *Queue) overall() float64 {
	if len(q.jobs) == 0 {
		return 0
	}
	total := 0.0
	for _, j := range q.jobs {
		total += j.progress
	}
	return total / float64(len(q.jobs))
}

func (q *Queue) startAll() tea.Cmd {
	q.run++
	run := q.run
	q.running = true
	q.everRun = true
	q.started = time.Now()
	q.elapsed = 0
	for i := range q.jobs {
		q.jobs[i] = jobState{}
	}
	q.events = nil

	cmds := []tea.Cmd{q.clockTick(run)}
	for i := range queueJobs {
		cmds = append(cmds, q.jobStep(run, i, 0))
	}
	return tea.Batch(cmds...)
}

func (q *Queue) clockTick(run int) tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return queueClockMsg{run: run, at: t}
	})
}

func (q *Queue) jobStep(run, index int, elapsed time.Duration) tea.Cmd {
	return tea.Tick(90*time.Millisecond, func(time.Time) tea.Msg {
		return jobTickMsg{run: run, index: index, elapsed: elapsed + 90*time.Millisecond}
	})
}

// Update implements Page.
func (q *Queue) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case jobTickMsg:
		if !q.active || msg.run != q.run || msg.index < 0 || msg.index >= len(q.jobs) {
			return nil
		}
		spec := queueJobs[msg.index]
		job := &q.jobs[msg.index]
		if job.done {
			return nil
		}
		if msg.elapsed <= spec.startAt {
			return q.jobStep(msg.run, msg.index, msg.elapsed)
		}
		if !job.started {
			job.started = true
			q.logEvent(msg.elapsed, "▸ "+spec.name+" started", spec.accent)
		}
		job.progress = mathMin(1, float64(msg.elapsed-spec.startAt)/float64(spec.duration))
		if job.progress >= 1 {
			job.done = true
			q.logEvent(msg.elapsed, "✓ "+spec.name, ui.Green)
			if q.allDone() {
				q.running = false
			}
			return nil
		}
		return q.jobStep(msg.run, msg.index, msg.elapsed)

	case queueClockMsg:
		if !q.active || msg.run != q.run {
			return nil
		}
		q.elapsed = msg.at.Sub(q.started)
		if q.allDone() {
			q.running = false
			return nil
		}
		return q.clockTick(msg.run)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if !q.running {
				return q.startAll()
			}
		case "r":
			q.reset()
			return nil
		}
	}
	return nil
}

func (q *Queue) reset() {
	q.run++
	q.running = false
	q.elapsed = 0
	q.events = nil
	for i := range q.jobs {
		q.jobs[i] = jobState{}
	}
}

func (q *Queue) logEvent(at time.Duration, text string, accent color.Color) {
	q.events = append(q.events, queueEvent{at: at, text: text, accent: accent})
	if len(q.events) > 40 {
		q.events = q.events[len(q.events)-40:]
	}
}

// View implements Page.
func (q *Queue) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	jobsH := height
	if height >= 20 {
		stats := q.stats(width, 5)
		if height >= 30 {
			logH := height - 23
			if logH >= 7 {
				top := lipgloss.JoinVertical(lipgloss.Left, stats, "", q.jobsCard(width, 16))
				return ui.Fit(lipgloss.JoinVertical(lipgloss.Left, top, "", q.logCard(width, logH)), width, height)
			}
		}
		jobsH = height - 6
		return ui.Fit(lipgloss.JoinVertical(lipgloss.Left, stats, "", q.jobsCard(width, jobsH)), width, height)
	}
	return ui.Fit(q.jobsCard(width, jobsH), width, height)
}

func (q *Queue) stats(width, height int) string {
	gap := 1
	cw := (width - 2*gap) / 3
	widths := []int{cw, cw, width - cw*2 - 2*gap}

	elapsed := fmt.Sprintf("%.1fs", q.elapsed.Seconds())
	finished := fmt.Sprintf("%d/%d", q.doneCount(), len(q.jobs))
	overall := fmt.Sprintf("%.0f%%", q.overall()*100)

	cards := []string{
		ui.Card(ui.Orange, "elapsed",
			lipgloss.NewStyle().Foreground(ui.Orange).Bold(true).Render(elapsed)+"\n"+
				ui.DimStyle.Render("wall clock"), widths[0], height),
		ui.Card(ui.Green, "finished",
			lipgloss.NewStyle().Foreground(ui.Green).Bold(true).Render(finished)+"\n"+
				ui.DimStyle.Render("jobs complete"), widths[1], height),
		ui.Card(ui.Cyan, "overall",
			lipgloss.NewStyle().Foreground(ui.Cyan).Bold(true).Render(overall)+"\n"+
				ui.DimStyle.Render("weighted progress"), widths[2], height),
	}
	return lipgloss.JoinHorizontal(lipgloss.Top,
		cards[0], strings.Repeat(" ", gap), cards[1], strings.Repeat(" ", gap), cards[2])
}

func (q *Queue) jobsCard(width, height int) string {
	bodyW := ui.CardBodyWidth(width)
	innerH := height - 2
	if bodyW < 16 {
		return ""
	}

	status := ui.SpaceBetween(
		q.statusLine(),
		ui.DimStyle.Render(fmt.Sprintf("%d concurrent commands", len(queueJobs))),
		bodyW,
	)

	spacious := innerH >= 12
	rows := make([]string, 0, len(queueJobs)*2)
	for i, spec := range queueJobs {
		rows = append(rows, q.jobRows(i, spec, bodyW, spacious)...)
	}

	footer := ui.SpaceBetween(
		ui.DimStyle.Render("enter run · r reset"),
		ui.DimStyle.Render(fmt.Sprintf("★ %.0f%% complete", q.overall()*100)),
		bodyW,
	)

	body := status + "\n\n" + strings.Join(rows, "\n")
	if spacious {
		// The card title occupies one inner row; keep the footer when it fits.
		switch {
		case innerH >= 1+lipgloss.Height(body)+2:
			body += "\n\n" + footer
		case innerH >= 1+lipgloss.Height(body)+1:
			body += "\n" + footer
		}
	}

	return ui.Card(ui.Orange, "brew queue", body, width, height)
}

func (q *Queue) logCard(width, height int) string {
	bodyW := ui.CardBodyWidth(width)
	innerH := height - 2
	if innerH < 3 {
		return ""
	}

	limit := max(1, innerH-2) // title + footer
	lines := make([]string, 0, limit)
	if len(q.events) == 0 {
		lines = append(lines, ui.DimStyle.Render("Waiting for commands…"))
	} else {
		start := max(0, len(q.events)-limit)
		for _, e := range q.events[start:] {
			stamp := fmt.Sprintf("%5.1fs", e.at.Seconds())
			lines = append(lines, ui.DimStyle.Render(stamp)+"  "+
				lipgloss.NewStyle().Foreground(e.accent).Render(e.text))
		}
	}

	footer := ui.SpaceBetween(
		ui.DimStyle.Render("every job is a Cmd · every tick is a Msg"),
		ui.DimStyle.Render(fmt.Sprintf("%d events", len(q.events))),
		bodyW,
	)
	body := strings.Join(lines, "\n") + "\n" + footer
	return ui.Card(ui.Cyan, "message log", body, width, height)
}

func (q *Queue) statusLine() string {
	switch {
	case q.running:
		spin := queueSpinner[int(q.elapsed/(100*time.Millisecond))%len(queueSpinner)]
		return lipgloss.NewStyle().Foreground(ui.Cyan).Render(spin) + " " +
			ui.TextStyle.Render("brewing") + ui.DimStyle.Render(" · tea is happening")
	case q.allDone():
		return lipgloss.NewStyle().Foreground(ui.Green).Bold(true).Render("✓ all served") +
			ui.DimStyle.Render(" · every cup made it")
	default:
		return ui.DimStyle.Render("○ idle · press enter to start the queue")
	}
}

func (q *Queue) jobRows(index int, spec jobSpec, width int, spacious bool) []string {
	job := q.jobs[index]
	status, statusColor := q.jobStatus(index, spec)
	prefix := "  "
	if !job.done && q.running && q.jobStarted(index, spec) {
		prefix = lipgloss.NewStyle().Foreground(statusColor).Render("▌") + " "
	}

	nameStyle := ui.TextStyle
	if job.done {
		nameStyle = lipgloss.NewStyle().Foreground(ui.Muted)
	}
	name := nameStyle.Render(spec.name)
	statusText := lipgloss.NewStyle().Foreground(statusColor).Render(status)

	if !spacious {
		barW := max(6, width-34)
		return []string{
			prefix + ui.SpaceBetween(name, statusText, width-2),
			"    " + ui.Bar(job.progress, barW, spec.accent, ui.Cyan) +
				ui.DimStyle.Render(fmt.Sprintf(" %3.0f%%", job.progress*100)),
		}
	}

	line1 := prefix + ui.SpaceBetween(name, statusText, width-2)
	barW := max(8, width-12)
	line2 := "    " + ui.Bar(job.progress, barW, spec.accent, ui.Cyan) +
		ui.DimStyle.Render(fmt.Sprintf(" %3.0f%%", job.progress*100))
	return []string{line1, line2}
}

func (q *Queue) jobStatus(index int, spec jobSpec) (string, color.Color) {
	job := q.jobs[index]
	switch {
	case job.done:
		return "✓ done", ui.Green
	case !q.running && q.elapsed == 0:
		return "○ waiting", ui.Dim
	case q.elapsed > 0 && q.elapsed < spec.startAt:
		return "○ queued", ui.Dim
	default:
		return "● running", spec.accent
	}
}

func (q *Queue) jobStarted(index int, spec jobSpec) bool {
	return q.jobs[index].progress > 0
}

func mathMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
