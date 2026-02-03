.PHONY: build build-dev clean help

# Get version from git tag or use dev version
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Build variables
BINARY_NAME = aiassist
LDFLAGS = -ldflags "\
	-X 'main.Version=$(VERSION)' \
	-X 'main.Commit=$(COMMIT)' \
	"

# Default target
build: ## Build the binary with version from git
	@echo "Building $(BINARY_NAME) version=$(VERSION) commit=$(COMMIT)"
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/aiassist/

build-dev: ## Build development binary with dev version
	@echo "Building $(BINARY_NAME) (development)"
	CGO_ENABLED=0 go build -o $(BINARY_NAME) ./cmd/aiassist/

clean: ## Remove built binary
	rm -f $(BINARY_NAME)

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Example commands:
# make build                          # Build with git version
# make build VERSION=v1.0.0           # Build with specific version
# make build-dev                      # Build development version
# make clean                          # Remove binary
