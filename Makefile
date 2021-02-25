SHELL := /bin/bash

# Default version number to the shorthand git commit hash if not set at the
# command line.
VER     := $(or $(VER),$(shell git log -1 --format="%h"))
COMMIT  := $(shell git log -1 --format="%h - %ae")
DATE    := $(shell date -u)
VERSION := $(VER) (commit $(COMMIT)) $(DATE)

GOSOURCES := $(shell find . \( -name '*.go' \))
INCLUDES  := $(shell find include \( -name '*' \) | sed 's/ /\\ /g')

# Default atomics repo to redcanaryco/master if not set at the command line.
ATOMICS_REPO := $(or $(ATOMICS_REPO),redcanaryco/master)

THISFILE := $(lastword $(MAKEFILE_LIST))
THISDIR  := $(shell dirname $(realpath $(THISFILE)))
GOBIN    := $(THISDIR)/bin

# Prepend this repo's bin directory to our path since we'll want to
# install some build tools there for use during the build process.
PATH := $(GOBIN):$(PATH)

# Export GOBIN env variable so `go install` picks it up correctly.
export GOBIN

all:

clean:
	-rm bin/atomicredteam*
	-rm -rf include/atomics
	-rm bindata.go

.PHONY: download-atomics
download-atomics:
	./download-atomics.sh $(ATOMICS_REPO)

bin/goart-linux: $(GOSOURCES) download-atomics
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-X 'actshad.dev/go-atomicredteam.Version=$(VERSION)' -X 'actshad.dev/go-atomicredteam.REPO=$(ATOMICS_REPO)' -s -w" -trimpath -o bin/goart-linux cmd/main.go

bin/goart-darwin: $(GOSOURCES) download-atomics
	mkdir -p bin
	CGO_ENABLED=0 GOOS=darwin go build -a -ldflags="-X 'actshad.dev/go-atomicredteam.Version=$(VERSION)' -X 'actshad.dev/go-atomicredteam.REPO=$(ATOMICS_REPO)' -s -w" -trimpath -o bin/goart-darwin cmd/main.go

bin/goart-windows: $(GOSOURCES) download-atomics
	mkdir -p bin
	CGO_ENABLED=0 GOOS=windows go build -a -ldflags="-X 'actshad.dev/go-atomicredteam.Version=$(VERSION)' -X 'actshad.dev/go-atomicredteam.REPO=$(ATOMICS_REPO)' -s -w" -trimpath -o bin/goart-windows cmd/main.go

release: bin/goart-linux bin/goart-darwin bin/goart-windows
