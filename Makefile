.PHONY: build test typecheck lint clean install release-snapshot help

GO ?= go
DIST ?= dist
PKG := github.com/RefuseHQ/refuse-cli
LDFLAGS := -s -w \
  -X "$(PKG)/internal/version.Version=$$(git describe --tags --dirty --always 2>/dev/null || echo dev)" \
  -X "$(PKG)/internal/version.Commit=$$(git rev-parse --short HEAD 2>/dev/null || echo none)" \
  -X "$(PKG)/internal/version.Date=$$(date -u +%Y-%m-%dT%H:%M:%SZ)"

help:
	@echo "Common targets:"
	@echo "  build              build a single binary for the host into $(DIST)/refuse"
	@echo "  test               run unit tests"
	@echo "  install            go install into \$$GOBIN"
	@echo "  release-snapshot   build all release artifacts via goreleaser (no publish)"
	@echo "  clean              remove $(DIST)/"

build:
	@mkdir -p $(DIST)
	$(GO) build -trimpath -ldflags '$(LDFLAGS)' -o $(DIST)/refuse ./cmd/refuse

test:
	$(GO) test ./...

typecheck:
	$(GO) vet ./...

install:
	$(GO) install -trimpath -ldflags '$(LDFLAGS)' ./cmd/refuse

release-snapshot:
	goreleaser release --snapshot --clean

clean:
	rm -rf $(DIST)
