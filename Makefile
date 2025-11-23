.PHONY: all build install clean

BINARY := glyphriot
OUTDIR := ./bin
OUTBIN := $(OUTDIR)/$(BINARY)
PREFIX ?= $(HOME)/.local
BINDIR := $(PREFIX)/bin

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
