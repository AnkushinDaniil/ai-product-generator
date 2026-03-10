# Memex MCP Server - Makefile
# Go-based Model Context Protocol server for persistent memory

.PHONY: all build run test lint clean help vuln trivy trivy-image security fmt-check deps deps-check deps-verify ci ci-quick

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
COVERAGE_THRESHOLD ?= 80

# Default target
all: lint test build

# Build the memex binary (with FTS5 support)
build:
	@echo "Building memex MCP server..."
	CGO_ENABLED=1 go build -tags "fts5" $(LDFLAGS) -o bin/memex ./cmd/memex

# Run the memex server
run:
	@echo "Running memex MCP server..."
	go run $(LDFLAGS) ./cmd/memex

# Run with hot reload (requires air)
dev:
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	air

# Run tests (with FTS5 support)
test:
	@echo "Running tests..."
	CGO_ENABLED=1 go test -tags "fts5" -v -race ./...

# Run tests with coverage and threshold enforcement (with FTS5 support)
test-coverage:
	@echo "Running tests with coverage..."
	CGO_ENABLED=1 go test -tags "fts5" -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "Checking coverage threshold ($(COVERAGE_THRESHOLD)%)..."
	@go tool cover -func=coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}' | \
		awk -v threshold=$(COVERAGE_THRESHOLD) '{if ($$1 < threshold) {print "Coverage " $$1 "% is below threshold " threshold "%"; exit 1} else {print "Coverage " $$1 "% meets threshold " threshold "%"}}'

# Generate HTML coverage report
coverage-html: test-coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ============================================
# SECURITY SCANNING
# ============================================

# Vulnerability scanning (Go dependencies)
vuln:
	@echo "Running govulncheck..."
	@which govulncheck > /dev/null || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

# Container security scanning
trivy:
	@echo "Running Trivy filesystem scan..."
	@which trivy > /dev/null || (echo "Install trivy: brew install trivy" && exit 1)
	trivy fs --severity HIGH,CRITICAL .

# Scan Docker image
trivy-image: docker-build
	@echo "Running Trivy image scan..."
	trivy image --severity HIGH,CRITICAL memex:$(VERSION)

# Full security audit
security: vuln trivy
	@echo "Security scan complete"

# ============================================
# DEPENDENCY MANAGEMENT
# ============================================

# Check for outdated dependencies
deps-check:
	@echo "Checking for outdated dependencies..."
	go list -u -m all

# Verify dependencies
deps-verify:
	@echo "Verifying dependencies..."
	go mod verify

# Full dependency audit
deps: tidy deps-verify vuln
	@echo "Dependency audit complete"

# Run linter
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

# Format code with gofumpt (stricter than gofmt)
fmt:
	@echo "Formatting code with gofumpt..."
	@which gofumpt > /dev/null || go install mvdan.cc/gofumpt@latest
	gofumpt -w .
	@which gci > /dev/null || go install github.com/daixiang0/gci@latest
	gci write --skip-generated -s standard -s default -s "prefix(github.com/AnkushinDaniil/memex)" .

# Check formatting (CI mode - no write)
fmt-check:
	@echo "Checking formatting..."
	@which gofumpt > /dev/null || go install mvdan.cc/gofumpt@latest
	@test -z "$$(gofumpt -l .)" || (echo "Files need formatting:" && gofumpt -l . && exit 1)

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/ coverage.out coverage.html

# ============================================
# CI TARGETS
# ============================================

# Full CI pipeline
ci: deps-verify fmt-check lint test-coverage security
	@echo "CI pipeline complete"

# Quick CI (skip security for speed)
ci-quick: fmt-check lint test
	@echo "Quick CI complete"

# Install development tools
tools:
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/daixiang0/gci@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "Note: Install trivy separately (brew install trivy)"

# Docker build
docker-build:
	docker build -t memex:$(VERSION) .

# Docker run
docker-run:
	docker run memex:$(VERSION)

# Help
help:
	@echo "Available targets:"
	@echo ""
	@echo "BUILD & RUN:"
	@echo "  build          - Build the memex MCP server binary"
	@echo "  run            - Run the memex server"
	@echo "  dev            - Run with hot reload"
	@echo ""
	@echo "TESTING:"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage threshold (default: 80%)"
	@echo "  coverage-html  - Generate HTML coverage report"
	@echo ""
	@echo "CODE QUALITY:"
	@echo "  lint           - Run golangci-lint"
	@echo "  fmt            - Format code (gofumpt + gci)"
	@echo "  fmt-check      - Check formatting (CI mode)"
	@echo ""
	@echo "SECURITY:"
	@echo "  vuln           - Run govulncheck"
	@echo "  trivy          - Run Trivy filesystem scan"
	@echo "  trivy-image    - Run Trivy on Docker image"
	@echo "  security       - Full security audit"
	@echo ""
	@echo "DEPENDENCIES:"
	@echo "  tidy           - Tidy go.mod"
	@echo "  deps           - Full dependency audit"
	@echo "  deps-check     - Check for outdated deps"
	@echo "  deps-verify    - Verify checksums"
	@echo ""
	@echo "CI/CD:"
	@echo "  ci             - Full CI pipeline"
	@echo "  ci-quick       - Quick CI (no security)"
	@echo ""
	@echo "DOCKER:"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo ""
	@echo "MISC:"
	@echo "  tools          - Install dev tools"
	@echo "  clean          - Clean build artifacts"
