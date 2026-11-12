package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestGradientPreservesWidth(t *testing.T) {
	in := "bubble tea"
	out := Gradient(in)
	if got, want := lipgloss.Width(out), len(in); got != want {
		t.Fatalf("width = %d, want %d", got, want)
	}
}

func TestCardExactSizeAndBorder(t *testing.T) {
	out := Card(Pink, "hello", "world", 24, 7)
	if got, want := lipgloss.Width(out), 24; got != want {
		t.Fatalf("width = %d, want %d", got, want)
	}
	if got, want := lipgloss.Height(out), 7; got != want {
		t.Fatalf("height = %d, want %d", got, want)
	}
	if !strings.Contains(out, "╰") || !strings.Contains(out, "╯") {
		t.Fatalf("card lost its bottom border:\n%s", out)
	}
}

func TestCardTruncatesOverflow(t *testing.T) {
	body := strings.Repeat("overflowing content\n", 20)
	out := Card(Pink, "title", body, 30, 6)
	if got := lipgloss.Height(out); got != 6 {
		t.Fatalf("height = %d, want 6", got)
	}
	if !strings.Contains(out, "╰") {
		t.Fatalf("card lost its bottom border:\n%s", out)
	}
}

func TestBarExactWidth(t *testing.T) {
	for _, f := range []float64{-1, 0, 0.33, 0.5, 1, 2} {
		if got := lipgloss.Width(Bar(f, 17)); got != 17 {
			t.Fatalf("Bar(%v) width = %d, want 17", f, got)
		}
	}
}

func TestAreaExactSize(t *testing.T) {
	values := []float64{0.1, 0.4, 0.8, 0.2, 0.9}
	out := Area(values, 20, 5)
	if got, want := lipgloss.Width(out), 20; got != want {
		t.Fatalf("width = %d, want %d", got, want)
	}
	if got, want := lipgloss.Height(out), 5; got != want {
		t.Fatalf("height = %d, want %d", got, want)
	}
}

func TestTruncateLines(t *testing.T) {
	out := TruncateLines("a\nb\nc\nd", 2)
	if got, want := strings.Count(out, "\n")+1, 2; got != want {
		t.Fatalf("lines = %d, want %d", got, want)
	}
}

func TestSpaceBetweenFitsWidth(t *testing.T) {
	out := SpaceBetween("left", "right", 20)
	if got := lipgloss.Width(out); got != 20 {
		t.Fatalf("width = %d, want 20", got)
	}
	clipped := SpaceBetween("a very long left side", "and a long right side", 12)
	if got := lipgloss.Width(clipped); got > 12 {
		t.Fatalf("width = %d, want <= 12", got)
	}
}
