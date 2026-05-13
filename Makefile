MODULE  := github.com/EtienneMR/tbx
BINARY  := tbx
VERSION ?= $(shell git describe --tags --always --dirty=-dev)
LDFLAGS := -ldflags "-X $(MODULE)/cli.Version=$(VERSION) -s -w"

.PHONY: build install clean lint test

## build: compile the binary into ./bin/tbx
build:
	go build $(LDFLAGS) -o dist/$(BINARY) .

## install: install the binary into $GOPATH/bin (or ~/go/bin)
install:
	go install $(LDFLAGS) .

## run: build and run with arguments, e.g.  make run ARGS="git start"
run: build
	./dist/$(BINARY) $(ARGS)

## test: run all tests
test:
	go test ./...

## clean: remove compiled artifacts
clean:
	go clean
	rm -rf dist/

## help: list targets
help:
	@grep -E '^## ' Makefile | sed 's/## //'
