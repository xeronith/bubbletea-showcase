// Package ui contains the shared visual language for the showcase: a small
// palette, gradient helpers, and the building blocks used by every page.
package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// The palette leans into a warm cup of tea at midnight: deep plum surfaces,
// pink and violet accents, and a fresh cyan highlight.
var (
	BG         color.Color = lipgloss.Color("#0D0B14")
	Surface    color.Color = lipgloss.Color("#151020")
	SurfaceAlt color.Color = lipgloss.Color("#1C1430")
	SurfaceHi  color.Color = lipgloss.Color("#251A3D")
	Border     color.Color = lipgloss.Color("#2B2140")
	BorderHi   color.Color = lipgloss.Color("#453563")

	Text  color.Color = lipgloss.Color("#F2EDFA")
	Muted color.Color = lipgloss.Color("#9A8DB2")
	Dim   color.Color = lipgloss.Color("#6B5F80")

	Pink    color.Color = lipgloss.Color("#FF6AC1")
	Magenta color.Color = lipgloss.Color("#E45BFF")
	Purple  color.Color = lipgloss.Color("#B388FF")
	Cyan    color.Color = lipgloss.Color("#69E8FF")
	Green   color.Color = lipgloss.Color("#7CF5B2")
	Yellow  color.Color = lipgloss.Color("#FFD479")
	Orange  color.Color = lipgloss.Color("#FF9E64")
	Red     color.Color = lipgloss.Color("#FF5D73")
)

// GradientStops is the signature pink → violet → cyan sweep used throughout
// the app.
var GradientStops = []color.Color{Pink, Purple, Cyan}

// GradientStopsWarm is a secondary sweep for warm, tea-related accents.
var GradientStopsWarm = []color.Color{Orange, Pink, Magenta}

// Common styles. Creating them once keeps render loops inexpensive.
var (
	TextStyle     = lipgloss.NewStyle().Foreground(Text)
	MutedStyle    = lipgloss.NewStyle().Foreground(Muted)
	DimStyle      = lipgloss.NewStyle().Foreground(Dim)
	BoldStyle     = lipgloss.NewStyle().Foreground(Text).Bold(true)
	AccentStyle   = lipgloss.NewStyle().Foreground(Pink)
	SurfaceStyle  = lipgloss.NewStyle().Background(Surface)
	SelectedStyle = lipgloss.NewStyle().Foreground(Pink).Bold(true)
)
