.DEFAULT_GOAL := help

CHAIN_DIR := chain
OUTPUT_DIR ?= $(HOME)/.lcv-devnet
NUM_VALIDATORS ?= 4

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the lcvd binary into chain/build/
	@mkdir -p $(CHAIN_DIR)/build
	cd $(CHAIN_DIR) && go build -o build/lcvd ./cmd/lcvd
	cd $(CHAIN_DIR) && go build -o build/gengenesis ./cmd/gengenesis

.PHONY: test
test: ## Run unit tests for the chain module
	cd $(CHAIN_DIR) && go vet ./...
	cd $(CHAIN_DIR) && go test ./...

.PHONY: lint
lint: ## Lint the chain module (golangci-lint if available, else go vet + gofmt check)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		cd $(CHAIN_DIR) && golangci-lint run ./... --timeout 15m; \
	else \
		echo "golangci-lint not found; falling back to go vet + gofmt -l"; \
		cd $(CHAIN_DIR) && go vet ./...; \
		unformatted=$$(cd $(CHAIN_DIR) && gofmt -l .); \
		if [ -n "$$unformatted" ]; then \
			echo "gofmt would reformat:"; echo "$$unformatted"; exit 1; \
		fi; \
	fi

.PHONY: genesis
genesis: build ## Regenerate the N-validator devnet genesis (wipes $OUTPUT_DIR)
	OUTPUT_DIR=$(OUTPUT_DIR) NUM_VALIDATORS=$(NUM_VALIDATORS) ./scripts/genesis/init-genesis.sh

.PHONY: devnet
devnet: build ## Start the local N-validator devnet (generates genesis first if missing)
	OUTPUT_DIR=$(OUTPUT_DIR) NUM_VALIDATORS=$(NUM_VALIDATORS) ./scripts/devnet/start.sh

.PHONY: localnet-stop
localnet-stop: ## Stop all running devnet validator processes
	OUTPUT_DIR=$(OUTPUT_DIR) ./scripts/devnet/stop.sh

.PHONY: localnet-reset
localnet-reset: localnet-stop ## Stop the devnet and wipe all its state
	rm -rf $(OUTPUT_DIR)

.PHONY: devnet-docker
devnet-docker: ## (Phase H) docker-compose devnet — not yet implemented, see DECISIONS.md
	@echo "devnet-docker is not implemented yet; use 'make devnet' for the local process-based devnet." && exit 1
