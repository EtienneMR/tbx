BINARY  ?= tbx-dev
DIST    ?= dist
MODULE  := github.com/EtienneMR/tbx
VERSION ?= $(shell git describe --tags --always --dirty=-dev)
LDFLAGS := -ldflags "-X $(MODULE)/internal/cli.Version=$(VERSION) -s -w"

.PHONY: all build install docs clean

all: install build docs

build:
	go build $(LDFLAGS) -o $(DIST)/$(BINARY) .

install: build
	cp $(DIST)/$(BINARY) $(HOME)/.local/bin

docs:
	rm -rf docs
	mkdir -p docs
	for f in ./internal/*; do \
		go doc -all "$$f" > "docs/$$(basename "$$f").txt"; \
		go doc "$$f" >> docs/all.txt; \
	done

clean:
	go clean
	rm -rf dist/
