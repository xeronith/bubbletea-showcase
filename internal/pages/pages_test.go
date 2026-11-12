package pages

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func TestKitchenRequiresAName(t *testing.T) {
	k := NewKitchen()
	k.input.SetValue("")
	k.startBrew()

	if k.brewing {
		t.Fatal("should not start brewing without a name")
	}
	if k.err == "" {
		t.Fatal("expected a validation error")
	}
	if k.focus != fieldName {
		t.Fatalf("focus = %d, want name field", k.focus)
	}
}

func TestKitchenBrewLifecycle(t *testing.T) {
	k := NewKitchen()
	k.active = true
	k.input.SetValue("Luna")
	k.focus = fieldBrew

	if cmd := k.startBrew(); cmd == nil {
		t.Fatal("expected a brew command")
	}
	if !k.brewing {
		t.Fatal("expected brewing state")
	}
	run := k.run

	k.Update(kitchenDoneMsg{run: run})
	if k.brewing || !k.done {
		t.Fatalf("expected finished order, got brewing=%v done=%v", k.brewing, k.done)
	}
	if k.orderName() != "Luna" {
		t.Fatalf("order name = %q", k.orderName())
	}

	k.reset()
	if k.done || k.brewing || k.input.Value() != "" {
		t.Fatal("reset should clear the order")
	}
	if k.strength != 3 {
		t.Fatalf("strength = %d, want 3", k.strength)
	}
}

func TestKitchenFocusSkipsInputWhenLeaving(t *testing.T) {
	k := NewKitchen()
	k.active = true
	k.moveFocus(-1) // Tea -> Name
	if !k.Editing() {
		t.Fatal("name field should capture input")
	}
	k.moveFocus(1) // Name -> Tea
	if k.Editing() {
		t.Fatal("tea field should not capture input")
	}
}

func TestKitchenRestoresInputFocusOnShow(t *testing.T) {
	k := NewKitchen()
	k.focus = fieldName
	k.input.Blur()
	k.OnShow()
	if !k.Editing() {
		t.Fatal("returning to the name field should restore input focus")
	}
}

func TestQueueRunAndReset(t *testing.T) {
	q := NewQueue()
	q.active = true
	if cmd := q.startAll(); cmd == nil {
		t.Fatal("expected commands")
	}
	if !q.running {
		t.Fatal("queue should be running")
	}

	run := q.run
	q.Update(jobTickMsg{run: run, index: 0, elapsed: 5 * time.Second})
	if !q.jobs[0].done || q.jobs[0].progress < 1 {
		t.Fatalf("job 0 should be done, got %+v", q.jobs[0])
	}

	for i := 1; i < len(q.jobs); i++ {
		q.jobs[i].done = true
		q.jobs[i].progress = 1
	}
	q.Update(queueClockMsg{run: run, at: q.started.Add(10 * time.Second)})
	if q.running {
		t.Fatal("queue should stop when every job is done")
	}

	q.run++
	q.elapsed = time.Second
	q.reset()
	if q.running || q.elapsed != 0 || q.doneCount() != 0 {
		t.Fatal("reset should clear the queue")
	}
}

func TestMotionWaveStaysInRange(t *testing.T) {
	m := NewMotion()
	for mode := 0; mode < 3; mode++ {
		m.mode = mode
		for _, phase := range []float64{0, 0.5, 1.7, 9.3} {
			m.phase = phase
			for i := 0; i <= 20; i++ {
				v := m.waveAt(float64(i) / 20)
				if v < 0 || v > 1.2 {
					t.Fatalf("mode %d phase %.1f produced %f", mode, phase, v)
				}
			}
		}
	}
}

func TestMotionKeys(t *testing.T) {
	m := NewMotion()
	m.active = true

	m.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	if !m.paused {
		t.Fatal("space should pause")
	}
	m.Update(tea.KeyPressMsg{Code: 'm', Text: "m"})
	if m.mode != 1 {
		t.Fatalf("mode = %d, want 1", m.mode)
	}
	m.Update(tea.KeyPressMsg{Code: '+', Text: "+"})
	if m.speed <= 1 {
		t.Fatalf("speed = %f, want > 1", m.speed)
	}
}

func TestPagesRenderAtEverySize(t *testing.T) {
	sizes := [][2]int{
		{76, 14},
		{82, 22},
		{94, 30},
		{120, 40},
	}
	for _, p := range All() {
		for _, size := range sizes {
			out := p.View(size[0], size[1])
			if out == "" {
				t.Fatalf("%s rendered nothing at %dx%d", p.Title(), size[0], size[1])
			}
			if w := lipgloss.Width(out); w > size[0] {
				t.Fatalf("%s is %d wide at %dx%d", p.Title(), w, size[0], size[1])
			}
			if h := lipgloss.Height(out); h > size[1] {
				t.Fatalf("%s is %d tall at %dx%d", p.Title(), h, size[0], size[1])
			}
		}
	}
}
