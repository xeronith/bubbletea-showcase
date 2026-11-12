// Package pages contains the five interactive demos that make up the
// showcase. Every page is an independent Bubble Tea model that the shell
// mounts and unmounts.
package pages

import tea "charm.land/bubbletea/v2"

// Page is a self-contained demo. The shell owns the chrome (header, sidebar,
// footer) and hands the page a rectangle to render into.
type Page interface {
	// Title is the short name shown in navigation.
	Title() string
	// Subtitle is the one-line description shown above the page.
	Subtitle() string
	// Init performs one-time setup. It runs once, when the app starts.
	Init() tea.Cmd
	// OnShow starts any animation loops. It runs whenever the page becomes
	// visible.
	OnShow() tea.Cmd
	// OnHide pauses animation loops. It runs whenever the page is hidden.
	OnHide()
	// Update handles a message and mutates the page in place.
	Update(tea.Msg) tea.Cmd
	// View renders the page body for the given rectangle.
	View(width, height int) string
	// Editing reports whether the page is currently capturing free-form text
	// input. The shell uses it to avoid stealing q, digits, and other keys.
	Editing() bool
}

// All returns the pages in navigation order.
func All() []Page {
	return []Page{
		NewOverview(),
		NewKitchen(),
		NewGallery(),
		NewMotion(),
		NewQueue(),
	}
}
