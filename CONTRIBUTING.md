# Contributing

Thanks for taking a look! This is a small demo project, so the process is intentionally light.

## Development

```bash
git clone https://github.com/xeronith/bubbletea-showcase.git
cd bubbletea-showcase
make run
```

Before opening a pull request, run:

```bash
make check
```

That checks formatting, runs `go vet`, runs the tests, and builds the binary.

## Commits

Use [Conventional Commits](https://www.conventionalcommits.org/):

```text
feat: add a new demo page
fix: keep the receipt card border when content overflows
docs: describe the motion easing
test: cover cursor clamping in the gallery
chore: bump bubbletea
```

## Code style

- Wrap errors and include context.
- Keep page state in the page model; keep side effects in `tea.Cmd`s.
- Prefer small helpers in `internal/ui` over copy-pasted Lip Gloss styles.
- Every behavior that can regress should have a test.
- A page must still render at `76 × 24` and must not panic at `82 × 22`.

## Demo assets

The animated GIF and stills are recorded with [VHS](https://github.com/charmbracelet/vhs):

```bash
make demo
```

This requires `vhs`, `ttyd`, and `ffmpeg`. The screenshots under `docs/screenshots/` were captured from the same tape at 1400 × 1000. If you change the UI in a visible way, re-record the GIF.

## Pull requests

- Keep the change focused.
- Include a short description of the behavior before and after.
- Add a test when the behavior is not purely visual.
- Follow the existing visual language (see `internal/ui/theme.go`).
