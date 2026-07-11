GO ?= go
BINARY ?= manly
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin

.PHONY: build install test coverage check clean

build:
	$(GO) build -o "$(BINARY)" ./cmd/manly

install:
	mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/$(BINARY)" ./cmd/manly

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
