.PHONY: build build-all clean help

build: ## Build the binary for current platform
	@./scripts/build.sh

build-all: ## Build for all platforms
	@./scripts/build-all.sh

clean: ## Remove built binaries
	@rm -f aiassist
	@rm -rf dist
	@echo "✓ Cleaned"

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
