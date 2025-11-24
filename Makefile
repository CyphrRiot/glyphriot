.PHONY: all build install clean build-linux-amd64 build-linux-arm64 build-windows-amd64 release

BINARY := glyphriot
OUTDIR := ./bin
OUTBIN := $(OUTDIR)/$(BINARY)
PREFIX ?= $(HOME)/.local
BINDIR := $(PREFIX)/bin
GOFLAGS ?=
LDFLAGS ?=
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)

all: build

build:
	mkdir -p $(OUTDIR)
	go build -o $(OUTBIN) .
	@echo "Built $(OUTBIN)"

install: build
	install -d $(BINDIR)
	install -m 0755 $(OUTBIN) $(BINDIR)/$(BINARY)
	@echo "Installed to $(BINDIR)/$(BINARY)"

clean:
	rm -f $(OUTBIN)

build-linux-amd64:
	mkdir -p $(OUTDIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(OUTDIR)/$(BINARY)-linux-amd64 .

build-linux-arm64:
	mkdir -p $(OUTDIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(OUTDIR)/$(BINARY)-linux-arm64 .

build-windows-amd64:
	mkdir -p $(OUTDIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(OUTDIR)/$(BINARY)-windows-amd64.exe .

release: clean build build-linux-amd64 build-linux-arm64 build-windows-amd64
	@echo "Release artifacts in $(OUTDIR):"
	@ls -1 $(OUTDIR) | sed 's/^/ - /'
