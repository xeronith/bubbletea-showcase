SHELL   := /bin/bash
GO      ?= go
GOFMT   ?= $(shell $(GO) env GOROOT)/bin/gofmt
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BINARY  := bin/showcase

.PHONY: all run build test vet fmt lint check demo clean

all: check

run:
	$(GO) run ./cmd/showcase

build:
	mkdir -p bin
	$(GO) build -trimpath \
		-ldflags "-s -w -X main.version=$(VERSION)" \
		-o $(BINARY) ./cmd/showcase

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	$(GOFMT) -w ./cmd ./internal

lint: vet
	@files="$$($(GOFMT) -l ./cmd ./internal)"; \
	if [ -n "$$files" ]; then \
		echo "gofmt needed for:"; echo "$$files"; \
		echo "run 'make fmt'"; \
		exit 1; \
	fi

check: lint test build

demo: build
	vhs demo/showcase.tape
	ffmpeg -y -i demo/showcase.mp4 \
		-vf "fps=20,scale=1100:-1:flags=lanczos,palettegen=stats_mode=diff" \
		demo/.palette.png
	ffmpeg -y -i demo/showcase.mp4 -i demo/.palette.png \
		-lavfi "fps=20,scale=1100:-1:flags=lanczos[x];[x][1:v]paletteuse=dither=bayer:bayer_scale=3" \
		demo/showcase.gif
	rm -f demo/.palette.png

clean:
	rm -rf bin
	rm -f demo/showcase.mp4 demo/.palette.png
