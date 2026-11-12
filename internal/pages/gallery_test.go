package pages

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		haystack string
		query    string
		want     bool
	}{
		{"charmbracelet", "chb", true},
		{"charmbracelet", "zzz", false},
		{"", "", true},
		{"anything", "", true},
		{"git", "git", true},
		{"git", "gti", false},
	}
	for _, tt := range tests {
		if got := fuzzyMatch(tt.haystack, tt.query); got != tt.want {
			t.Errorf("fuzzyMatch(%q, %q) = %v, want %v", tt.haystack, tt.query, got, tt.want)
		}
	}
}

func TestGalleryFilterIsFieldScoped(t *testing.T) {
	g := NewGallery()
	g.query = "git"
	g.applyFilter()

	if len(g.filtered) != 2 {
		t.Fatalf("filtered = %d, want 2 (%v)", len(g.filtered), g.filtered)
	}
	if catalog[g.filtered[0]].name != "gh-dash" || catalog[g.filtered[1]].name != "Soft Serve" {
		t.Fatalf("unexpected matches: %q, %q", catalog[g.filtered[0]].name, catalog[g.filtered[1]].name)
	}
}

func TestGallerySearchAndStar(t *testing.T) {
	g := NewGallery()
	g.width, g.height = 120, 40

	g.handleKey(tea.KeyPressMsg{Code: '/', Text: "/"})
	if !g.searching {
		t.Fatal("expected search mode")
	}
	for _, r := range "gh" {
		g.handleKey(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	if g.query != "gh" {
		t.Fatalf("query = %q, want gh", g.query)
	}
	if len(g.filtered) == 0 {
		t.Fatal("expected at least one match")
	}

	g.handleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !g.searching {
		// Enter commits the search.
	} else {
		t.Fatal("enter should commit the search")
	}

	idx := g.filtered[g.cursor]
	g.handleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !g.starred[idx] {
		t.Fatal("enter should star the selected project")
	}
	if g.starCount() != 1 {
		t.Fatalf("star count = %d, want 1", g.starCount())
	}
}

func TestGalleryCursorClamps(t *testing.T) {
	g := NewGallery()
	g.move(-5)
	if g.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", g.cursor)
	}
	g.move(1000)
	if g.cursor != len(g.filtered)-1 {
		t.Fatalf("cursor = %d, want %d", g.cursor, len(g.filtered)-1)
	}

	g.query = "no-such-project-anywhere"
	g.applyFilter()
	if len(g.filtered) != 0 || g.cursor != 0 {
		t.Fatalf("expected empty filter and cursor 0, got %d/%d", len(g.filtered), g.cursor)
	}
}
