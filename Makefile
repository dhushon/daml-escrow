# Phase 7: Production-Ready Orchestration (Makefile)

.PHONY: all
all: build test

.PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: frontend-build api-build ## Build both frontend and backend

.PHONY: frontend-build
frontend-build: ## Build the Astro production server
	@echo "Building frontend..."
	@cd frontend && npm run build

.PHONY: api-build
api-build: ## Build the Go API server
	@echo "Building backend..."
	@go build -o bin/escrow-api ./cmd/escrow-api

.PHONY: test
test: ## Run all unit tests
	@go test -v ./...

.PHONY: integration-test
integration-test: ## Run local integration tests (requires Docker)
	@echo "Running integration tests..."
	@go test -v -tags integration ./internal/ledger/...

.PHONY: sync
sync: ## Synchronize ledger state (resolve Package and Party IDs)
	@echo "Discovering ledger state on localhost:7575 (timeout: 5m)..."
	@./bin/ledger-sync -host localhost -port 7575 \
		-impl stablecoin-escrow \
		-iface stablecoin-escrow-interfaces \
		-out ledger-state.json

## -- High-Assurance Orchestration --

.PHONY: local-up
local-up: ## Launch the local tripartite tripartite simulation and observability stack
	@echo "Launching local tripartite stack..."
	@docker-compose -f docker-compose.distributed.yml -f docker-compose.otel.yml up -d --build
	@sleep 15
	@$(MAKE) sync

.PHONY: local-down
local-down: ## Definitive purge of all local tripartite and observability containers
	@echo "Purging local tripartite stack..."
	@docker-compose -f docker-compose.distributed.yml -f docker-compose.otel.yml down -v --remove-orphans

.PHONY: pilot-deploy
pilot-deploy: ## Authoritatively deploy the tripartite manifests to the live GKE cluster
	@echo "Deploying to GKE Pilot (api.vdatacloudai.com)..."
	@kubectl apply -f k8s/namespaces.yaml
	@kubectl apply -f k8s/tls-issuer.yaml
	@kubectl apply -f k8s/cas-issuer.yaml
	@kubectl apply -f k8s/canton-configs.yaml
	@kubectl apply -f k8s/bank-ledger.yaml
	@kubectl apply -f k8s/bank-api.yaml
	@kubectl apply -f k8s/buyer-ledger.yaml
	@kubectl apply -f k8s/buyer-api.yaml
	@kubectl apply -f k8s/seller-ledger.yaml
	@kubectl apply -f k8s/seller-api.yaml
	@kubectl apply -f k8s/ingress.yaml

.PHONY: pilot-status
pilot-status: ## Authoritatively audit the health of the live GKE tripartite nodes
	@echo "Auditing GKE Pilot Status..."
	@kubectl get pods --all-namespaces -l "env=dev"
	@kubectl get ingress -n bank
	@kubectl get certificate --all-namespaces

## -- Observability & Dashboards --

.PHONY: otel-up
otel-up: ## Launch the High-Assurance Observability stack (Jaeger, Prometheus, Grafana)
	@docker-compose -f docker-compose.otel.yml up -d

.PHONY: otel-down
otel-down: ## Stop the Observability stack
	@docker-compose -f docker-compose.otel.yml down

.PHONY: otel-logs
otel-logs: ## Follow logs from the OTEL Collector
	@docker logs -f otel-collector-dev

## -- Specialized Verification --

.PHONY: integration-gcp
integration-gcp: ## Run all GCP-specific integration tests (Secret Manager, Cloud SQL)
	@echo "Running GCP integration tests..."
	@go test -v -tags integration_gcp ./internal/config/... ./internal/ledger/gcp_db_test.go

.PHONY: verify-gcp-secrets
verify-gcp-secrets: ## Authoritatively audit Secret Manager connectivity and institutional key vending
	@echo "Auditing cloud secret vending..."
	@go test -v -tags integration_gcp -run TestGCPSecretManagerSpecialty ./internal/config/

## -- Simulations --

.PHONY: oracle-sim
oracle-sim: ## Run the Oracle Simulator CLI
	@go run ./cmd/oracle-simulator/main.go
