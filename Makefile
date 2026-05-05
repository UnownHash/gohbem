.PHONY: all build test race vet lint bench fuzz-query fuzz-cp tidy clean help

GO       ?= go
PKG      ?= ./...
FUZZTIME ?= 30s

all: vet test

build:
	$(GO) build $(PKG)

test:
	$(GO) test $(PKG)

race:
	$(GO) test -race $(PKG)

vet:
	$(GO) vet $(PKG)

lint:
	golangci-lint run

bench:
	$(GO) test -bench=. -benchmem -run=^$$ $(PKG)

fuzz-query:
	$(GO) test -run=^$$ -fuzz=FuzzQueryPvPRankBounds -fuzztime=$(FUZZTIME) .

fuzz-cp:
	$(GO) test -run=^$$ -fuzz=FuzzCalculateCpBounds -fuzztime=$(FUZZTIME) .

tidy:
	$(GO) mod tidy

clean:
	$(GO) clean -testcache
	rm -rf testdata/fuzz

help:
	@echo "Targets:"
	@echo "  build       compile package"
	@echo "  test        run unit tests"
	@echo "  race        run unit tests with -race"
	@echo "  vet         go vet"
	@echo "  lint        golangci-lint run"
	@echo "  bench       run benchmarks"
	@echo "  fuzz-query  fuzz QueryPvPRank bounds (FUZZTIME=$(FUZZTIME))"
	@echo "  fuzz-cp     fuzz CalculateCp bounds (FUZZTIME=$(FUZZTIME))"
	@echo "  tidy        go mod tidy"
	@echo "  clean       drop test cache + fuzz corpus"
