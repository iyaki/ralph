# Makefile for Ralph Go CLI

.PHONY: all build clean install test test-coverage deps help

BINARY_NAME=ralph
GO=go
INSTALL_PATH=/usr/local/bin

all: build

# Build the binary
build:
	$(GO) build -o $(BINARY_NAME) ./cmd/ralph

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	$(GO) clean

# Install the binary to system path
install: build
	install -m 0755 $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)

# Run tests (if any)
test:
	$(GO) test -v ./...

# Run tests with coverage and enforce minimum threshold
test-coverage:
	$(GO) test ./... -coverprofile=coverage.out -covermode=atomic
	@total=`$(GO) tool cover -func=coverage.out | tail -n 1 | awk '{print $$3}' | sed 's/%//'`; \
	echo "Total coverage: $$total%"; \
	awk 'BEGIN { exit !('"$$total"' >= 90) }' || (echo "Coverage check failed: require >= 90%" && exit 1)

# Download dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

# Show help
help:
	@echo "Ralph Go CLI Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build       - Build the ralph binary"
	@echo "  clean       - Remove built artifacts"
	@echo "  install     - Install ralph to $(INSTALL_PATH)"
	@echo "  test        - Run tests"
	@echo "  test-coverage - Run tests with coverage (fails if < 90%)"
	@echo "  deps        - Download and tidy dependencies"
	@echo "  help        - Show this help message"
