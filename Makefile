MODULE  := github.com/EtienneMR/tbx
BINARY  := tbx
VERSION ?= $(shell git describe --tags --always --dirty=-dev)
LDFLAGS := -ldflags "-X $(MODULE)/internal/cli.Version=$(VERSION) -s -w"

.PHONY: build install clean lint test

build:
	go build $(LDFLAGS) -o dist/$(BINARY) .

install:
	go install $(LDFLAGS) .

clean:
	go clean
	rm -rf dist/
