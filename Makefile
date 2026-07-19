GO ?= go
BINARY ?= manly
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

GO_BUILD = $(GO) build -ldflags "-X main.version=$(VERSION)"

.PHONY: build install test coverage check clean

build:
	$(GO_BUILD) -o "$(BINARY)" ./cmd/manly

install:
	mkdir -p "$(BINDIR)"
	$(GO_BUILD) -o "$(BINDIR)/$(BINARY)" ./cmd/manly

# Run the full Go toolchain checks used by this project.
test:
	$(GO) test ./...

coverage:
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -func=coverage.out

check: test
	$(GO) build ./cmd/manly

clean:
	rm -f "$(BINARY)"
