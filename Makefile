SHELL := /bin/bash

# Default hyperdark version number to the shorthand git commit hash if
# not set at the command line.
VER     := $(or $(VER),$(shell git log -1 --format="%h"))
COMMIT  := $(shell git log -1 --format="%h - %ae")
DATE    := $(shell date -u)
VERSION := $(VER) (commit $(COMMIT)) $(DATE)

GOSOURCES := $(shell find . \( -name '*.go' \))
INCLUDES  := $(shell find includes \( -name '*' \))

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
	-rm bin/atomicredteam
	-rm bindata.go

.PHONY: install-build-deps
install-build-deps: bin/go-bindata

.PHONY: remove-build-deps
remove-build-deps:
	$(RM) bin/go-bindata

bin/go-bindata:
	go install github.com/go-bindata/go-bindata/v3/go-bindata

bindata.go: $(INCLUDES) bin/go-bindata
	$(GOBIN)/go-bindata -pkg atomicredteam -prefix include -o bindata.go include/...

bin/atomicredteam: $(GOSOURCES) bindata.go
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-X 'actshad.dev/go-atomicredteam.Version=$(VERSION)' -s -w" -trimpath -o bin/atomicredteam cmd/main.go

bin/atomicredteam-linux: $(GOSOURCES) bindata.go
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-X 'actshad.dev/go-atomicredteam.Version=$(VERSION)' -s -w" -trimpath -o bin/atomicredteam-linux cmd/main.go

bin/atomicredteam-darwin: $(GOSOURCES) bindata.go
	mkdir -p bin
	CGO_ENABLED=0 GOOS=darwin go build -a -ldflags="-X 'actshad.dev/go-atomicredteam.Version=$(VERSION)' -s -w" -trimpath -o bin/atomicredteam-darwin cmd/main.go

bin/atomicredteam-windows: $(GOSOURCES) bindata.go
	mkdir -p bin
	CGO_ENABLED=0 GOOS=windows go build -a -ldflags="-X 'actshad.dev/go-atomicredteam.Version=$(VERSION)' -s -w" -trimpath -o bin/atomicredteam-windows cmd/main.go

release: bin/atomicredteam-linux bin/atomicredteam-darwin bin/atomicredteam-windows