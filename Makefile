# Makefile for Stablecoin Escrow Platform

APP_NAME=escrow-api
DPM=/Users/dhushon/.dpm/bin/dpm
JAVA_HOME_17=/opt/homebrew/opt/openjdk@17

# Ledger Config Paths
SANDBOX_CONF=Sandbox/sandbox.conf
SANDBOX_INIT=Sandbox/sandbox_init.canton
SETUP_SCRIPT=scripts/setup_users.sh

.DEFAULT_GOAL := help

## -- Help & Documentation --

.PHONY: help
help: ## Show this help message
	@echo "Stablecoin Escrow Platform - Management Console"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Strategies:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

## -- Full-Stack Orchestration --

.PHONY: up
up: docker-up api-up frontend-up ## START everything (Docker Ledger + Background API + Background Frontend)
	@echo "---------------------------------------"
	@echo "Full stack is UP."
	@echo "Web Dashboard: http://localhost:8080"
	@echo "Backend API:   http://localhost:8081"
	@echo "Ledger JSON:   http://localhost:7575"
	@echo "---------------------------------------"

.PHONY: down
down: ## STOP everything (Docker Ledger + Background processes)
	@echo "Stopping all components..."
	@-[ -f api.pid ] && kill $$(cat api.pid) && rm api.pid || true
	@-[ -f frontend.pid ] && kill $$(cat frontend.pid) && rm frontend.pid || true
	@pkill -f "npm run dev" || true
	docker-compose down
	@echo "Stack is down."

## -- Component Lifecycle --

.PHONY: docker-up
docker-up: ## Start only the Ledger and Database (Docker)
	@echo "Starting Docker Compose stack (Ledger + DB)..."
	docker-compose up -d --build
	@echo "Monitoring ledger logs for readiness..."
	@count=0; until docker-compose logs sandbox | grep -q "Setup complete." || [ $$count -eq 60 ]; do \
		sleep 5; \
		count=$$((count + 1)); \
	done
	@docker-compose logs sandbox | grep -q "Setup complete." && echo "Ledger reports: Setup complete." || (echo "Timed out waiting for ledger setup"; exit 1)

.PHONY: api-up
api-up: build ## Start only the Go API (background)
	@echo "Starting Go API on port 8081 (background)..."
	@nohup bin/$(APP_NAME) > api.log 2>&1 & echo $$! > api.pid
	@echo "API started. PID: $$(cat api.pid). Logs: api.log"

.PHONY: frontend-up
frontend-up: ## Start only the Astro Frontend (background)
	@echo "Starting Astro Frontend on port 8080 (background)..."
	@cd frontend && nohup npm run dev > ../frontend.log 2>&1 & echo $$! > frontend.pid
	@echo "Frontend started. PID: $$(cat frontend.pid). Logs: frontend.log"

## -- Build & Development --

.PHONY: build
build: ## Build Go binaries (API and Simulator)
	@echo "Building Go API..."
	go build -o bin/$(APP_NAME) ./cmd/escrow-api
	@echo "Building Oracle Simulator..."
	go build -o bin/oracle-simulator ./cmd/oracle-simulator
	@echo "Building Ledger Sync Tool..."
	go build -o bin/ledger-sync ./cmd/ledger-sync

.PHONY: sync
sync: build ## DISCOVER and EXPORT ledger state (Package IDs, Party IDs) to ledger-state.json
	@echo "Exporting ledger state..."
	FORCE_DISCOVERY=true ./bin/ledger-sync -host localhost -port 7575 \
		-impl stablecoin-escrow \
		-iface stablecoin-escrow-interfaces \
		-out ledger-state.json

.PHONY: daml-build
daml-build: ## Build all Daml packages using DPM
	@echo "Building Daml Contracts..."
	cd contracts && $(DPM) build --all

.PHONY: clean-db
clean-db: ## Wipe the persistent ledger database (Postgres)
	@echo "Wiping persistent ledger database..."
	docker-compose exec -T postgres psql -U escrow -d postgres -c "DROP DATABASE IF EXISTS escrow;" || true
	docker-compose exec -T postgres psql -U escrow -d postgres -c "CREATE DATABASE escrow;" || true

.PHONY: clean
clean: ## Remove binaries, logs, and PID files
	rm -rf bin *.log *.pid

.PHONY: clean-frontend
clean-frontend: ## Clean Astro frontend build artifacts
	cd frontend && rm -rf dist .astro node_modules

## -- GCP Identity Management --

.PHONY: gcp-identity-up
gcp-identity-up: ## Initialize Google Cloud Identity Platform and enable APIs
	@chmod +x scripts/setup_gcp_identity.sh
	./scripts/setup_gcp_identity.sh

.PHONY: gcp-identity-down
gcp-identity-down: ## Disable Identity Platform APIs (Clean)
	@echo "Disabling Identity Platform APIs..."
	gcloud services disable identitytoolkit.googleapis.com identityplatform.googleapis.com
	@echo "GCP Identity services disabled."

## -- Testing --

.PHONY: test
test: ## Run Go unit tests (fast, no infra)
	go test -v ./...

.PHONY: integration-test
integration-test: ## Run local integration tests (single-node sandbox)
	@echo "Running local integration tests..."
	@go test -v -tags integration ./internal/ledger/ledger_integration_test.go \
		./internal/ledger/json_base.go \
		./internal/ledger/json_parser.go \
		./internal/ledger/json_parties.go \
		./internal/ledger/json_escrows.go \
		./internal/ledger/json_stablecoin.go \
		./internal/ledger/multi_client.go \
		./internal/ledger/client.go \
		./internal/services/config_service_int_test.go \
		./internal/services/config_service.go

.PHONY: distributed-test
distributed-test: ## Run multi-node distributed tests (full topology)
	@echo "Running multi-node distributed tests..."
	@go test -v -tags distributed ./internal/ledger/multi_node_test.go \
		./internal/ledger/json_base.go \
		./internal/ledger/json_parser.go \
		./internal/ledger/json_parties.go \
		./internal/ledger/json_escrows.go \
		./internal/ledger/json_stablecoin.go \
		./internal/ledger/multi_client.go \
		./internal/ledger/client.go



## -- Simulations --

.PHONY: oracle-sim
oracle-sim: ## Run the Oracle Simulator CLI
	@go run ./cmd/oracle-simulator/main.go
