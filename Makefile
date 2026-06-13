# Phase 7: Institutional Grade Orchestration (Makefile)
# authoritatively manages tripartite lifecycles across Local and GKE perimeters.

# Architecture Detection
LOCAL_ARCH := $(shell uname -m | sed 's/x86_64/amd64/' | sed 's/arm64/arm64/')
ARCH ?= $(LOCAL_ARCH)

.PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

## -- Local Development (Standalone - Single Node) --

.PHONY: standalone-up
standalone-up: ## Authoritatively launch local baseline (Ledger: 7575, API: 8081, UX: 4321)
	@echo "Launching Standalone (Single-Node) stack..."
	@mkdir -p log
	@docker compose up -d
	@echo "Awaiting ledger (60s)..." && sleep 60
	@make bootstrap-local
	@nohup env LEDGER_HOST=localhost \
		STORAGE_ENDPOINT=http://localhost:9000 \
		STORAGE_ACCESS_KEY=escrow \
		STORAGE_SECRET_KEY=escrow-secret \
		STORAGE_BUCKET=escrow-documents \
		go run ./cmd/escrow-api serve --notls --bypass --config config/config-standalone.yaml --port 8081 > log/standalone-api.log 2>&1 &
	@cd frontend && env PUBLIC_API_URL=http://localhost:8081 npm run dev -- --port 4321 > ../log/standalone-frontend.log 2>&1 &
	@echo "SUCCESS: Standalone Baseline LIVE on http://localhost:4321"

.PHONY: standalone-down
standalone-down: ## Purge all local standalone processes and containers
	@pkill -f "escrow-api" || true
	@pkill -f "astro" || true
	@docker compose down -v
	@echo "Standalone environment purged."

## -- Local Development (Tripartite - Multi Node) --

.PHONY: tri-up
tri-up: ## Authoritatively launch distributed tripartite stack
	@echo "Launching Standalone-Tri (Multi-Node) stack..."
	@mkdir -p log
	@docker compose -f docker-compose.distributed.yml up -d
	@echo "Awaiting distributed ledger (60s)..." && sleep 60
	@./scripts/setup_users.sh localhost 7575
	@make bootstrap-local
	@nohup env LEDGER_HOST=localhost \
		STORAGE_ENDPOINT=http://localhost:9000 \
		STORAGE_ACCESS_KEY=escrow \
		STORAGE_SECRET_KEY=escrow-secret \
		STORAGE_BUCKET=escrow-documents \
		go run ./cmd/escrow-api serve --notls --bypass --config config/config-tri.yaml --port 8081 > log/tri-api.log 2>&1 &
	@cd frontend && env PUBLIC_API_URL=http://localhost:8080 npm run dev -- --port 4321 > ../log/tri-frontend.log 2>&1 &
	@echo "SUCCESS: Tripartite Distributed LIVE on http://localhost:4321"

.PHONY: tri-down
tri-down: ## Purge all tripartite processes and containers
	@pkill -f "escrow-api" || true
	@pkill -f "astro" || true
	@docker compose -f docker-compose.distributed.yml down -v
	@echo "Tripartite environment purged."

## -- GKE Pilot Orchestration (3-Tier Governance) --

.PHONY: admin-up
admin-up: ## Tier 1: Provision Cloud Foundation (Root CA, Audit, Static IP)
	@./scripts/admin-setup.sh vdcai-daml up

.PHONY: admin-status
admin-status: ## Tier 1: Audit Admin Foundation health and CAS bindings
	@./scripts/admin-setup.sh vdcai-daml status

.PHONY: pilot-up
pilot-up: ## Tier 2: Provision Workload (GKE, KMS) and deploy tripartite nodes
	@./scripts/gke-pilot.sh up

.PHONY: pilot-status
pilot-status: ## Tier 2: Audit pod health and mTLS certificate status
	@./scripts/gke-pilot.sh status

.PHONY: pilot-local
pilot-local: ## Authoritatively launch local services pointing to GKE (api.vdatacloudai.com)
	@echo "Launching local services in GCP-Proxy mode..."
	@nohup go run ./cmd/escrow-api serve --notls --bypass --port 8081 --config config/config-gke.yaml > log/pilot-local-api.log 2>&1 &
	@cd frontend && PUBLIC_API_URL=http://localhost:8081 npm run dev -- --port 4321 > ../log/pilot-local-frontend.log 2>&1 &
	@echo "SUCCESS: Local UX pointing to GKE LIVE on http://localhost:4321"

.PHONY: pilot-down
pilot-down: ## Tier 2: authoritatively DESTROY GKE cluster and node pools (Cost Protection)
	@./scripts/gke-pilot.sh down

.PHONY: build-contracts
build-contracts: ## Build authoritative DAML DAR packages
	@cd contracts/stablecoin-escrow-interfaces && daml build
	@cd contracts/stablecoin-escrow && daml build
	@cd contracts/stablecoin-escrow-tests && daml build

.PHONY: codegen
codegen: build-contracts ## Authoritatively regenerate Go bindings from DAR files
	@echo "Building godaml tool..."
	@cd third_party/go-daml && make build
	@echo "Generating institutional bindings..."
	@./third_party/go-daml/bin/godaml \
		--dar ./contracts/stablecoin-escrow/.daml/dist/stablecoin-escrow-0.0.3.dar \
		--output ./internal/ledger/generated \
		--go_package generated
	@echo "Codegen Complete: Go bindings synchronized with institutional vocabulary."

.PHONY: swagger-gen
swagger-gen: ## Regenerate Swagger/OpenAPI documentation
	@echo "Generating API Swagger docs..."
	@swag init -g cmd/escrow-api/main.go -o docs

.PHONY: pilot-release
pilot-release: ## Authoritatively build and push image to GCP Artifact Registry (Defaults to ARCH or TARGETARCH)
	@echo "Releasing GKE Pilot API (Arch: $(ARCH))..."
	@docker build --build-arg TARGETARCH=$(ARCH) -t us-central1-docker.pkg.dev/vdcai-daml/escrow-platform-dev/escrow-api:latest .
	@docker push us-central1-docker.pkg.dev/vdcai-daml/escrow-platform-dev/escrow-api:latest
	@kubectl rollout restart deployment bank-api -n bank
	@kubectl rollout restart deployment depositor-api -n depositor
	@kubectl rollout restart deployment beneficiary-api -n beneficiary

## -- Ledger & State Synchronization --

.PHONY: bootstrap-local
bootstrap-local: ## Synchronize DAR packages and allocate Parties on localhost
	@GOARCH=$(ARCH) ./bin/ledger-sync -host localhost -port 7575 \
		-impl stablecoin-escrow \
		-iface stablecoin-escrow-interfaces \
		-out ledger-state.json

.PHONY: bootstrap-gke
bootstrap-gke: ## Authoritatively synchronize all tripartite nodes in GKE
	@for ns in bank depositor beneficiary; do \
		echo "Bootstrapping GKE Namespace: $$ns"; \
		kubectl port-forward pod/$$ns-ledger-0 7575:7575 -n $$ns > /dev/null 2>&1 & \
		PID=$$!; sleep 15; \
		./bin/ledger-sync -host localhost -port 7575 -impl stablecoin-escrow -iface stablecoin-escrow-interfaces -out log/gke-state-$$ns.json || true; \
		kill $$PID || true; \
	done

## -- Testing & Verification --

.PHONY: test
test: ## Run all backend unit tests
	@source ~/.zprofile && go test ./...

.PHONY: test-storage
test-storage: ## Authoritatively verify storage infrastructure (MinIO/GCS) and Read-Through logic
	@echo "Running Storage Infrastructure & API Integration Tests..."
	@source ~/.zprofile && go test -v -tags=integration -run TestEndToEndStorageMirroring_Infra ./internal/services/...

.PHONY: verify
verify: ## Run all local verification tests (Go, DAML, Astro build, frontend tests)
	@echo "Running local verification..."
	@source ~/.zprofile && go test ./...
	@cd contracts/stablecoin-escrow-tests && source ~/.zprofile && daml test
	@cd frontend && npm run build
	@cd frontend && npm run test

.PHONY: install-hooks
install-hooks: ## Install local git hooks
	@bash scripts/install-git-hooks.sh

.PHONY: test-e2e
test-e2e: ## Run Playwright E2E integration tests locally (boots, tests, and tears down stack)
	@echo "Launching local baseline stack..."
	@make standalone-up
	@echo "Awaiting services health check..."
	@npx wait-on -t 120000 http-get://localhost:8081/api/v1/health http-get://localhost:4321/login
	@echo "Executing Playwright E2E integration tests..."
	@cd frontend && npx playwright test; \
	status=$$?; \
	echo "Tearing down baseline stack..."; \
	make standalone-down; \
	exit $$status
