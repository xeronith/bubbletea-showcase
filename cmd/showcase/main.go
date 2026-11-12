// Command showcase is a hands-on tour of Bubble Tea v2. It is intentionally
// small enough to read in one sitting and broad enough to be useful as a
// reference.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/xeronith/bubbletea-showcase/internal/app"
)

// version is injected by GoReleaser or with -ldflags.
var version = "dev"

const usage = `Bubble Tea Showcase — a terminal tour of the Elm Architecture

Usage:
  showcase              launch the interactive tour
  showcase -h, --help   show this help
  showcase -v, --version

Keys:
  1–5        jump to a demo
  tab        next demo
  shift+tab  previous demo
  ?          keyboard shortcuts
  q          quit

Learn more:
  https://github.com/charmbracelet/bubbletea
  https://github.com/xeronith/bubbletea-showcase
`

func main() {
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help":
			fmt.Print(usage)
			return
		case "-v", "--version":
			fmt.Printf("bubbletea-showcase %s\n", version)
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown argument %q\n\n%s", arg, usage)
			os.Exit(2)
		}
	}

	p := tea.NewProgram(app.New())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "bubbletea-showcase: %v\n", err)
		os.Exit(1)
	}
}
