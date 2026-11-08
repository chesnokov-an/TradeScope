GO ?= go
TYPST ?= $(shell command -v typst 2>/dev/null || echo $(HOME)/.local/bin/typst)
BIN := bin/tradescope
REPORT := report/coursework.pdf

.PHONY: all run build test vet check experiment report clean

all: check build report

run:
	$(GO) run ./cmd/server

build:
	mkdir -p bin
	$(GO) build -o $(BIN) ./cmd/server

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

check: test vet build

experiment:
	$(GO) run ./cmd/experiment

report: experiment
	$(TYPST) compile report/main.typ $(REPORT)

clean:
	rm -rf bin $(REPORT)
