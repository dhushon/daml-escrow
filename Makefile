# Phase 7: Institutional Grade Orchestration (Makefile)
# authoritatively manages tripartite lifecycles across Local and GKE perimeters.

# Architecture Detection
LOCAL_ARCH := $(shell uname -m | sed 's/x86_64/amd64/' | sed 's/arm64/arm64/')
ARCH ?= $(LOCAL_ARCH)

.PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

## -- Local Development (Standalone) --

.PHONY: standalone-up
standalone-up: ## Authoritatively launch local baseline (Ledger: 7575, API: 8081, UX: 4321)
	@echo "Launching accelerated standalone stack..."
	@docker compose up -d
	@echo "Awaiting ledger (60s)..." && sleep 60
	@make bootstrap-local
	@nohup go run ./cmd/escrow-api serve --notls --bypass --port 8081 > log/standalone-api.log 2>&1 &
	@cd frontend && npm run dev -- --port 4321 > ../log/standalone-frontend.log 2>&1 &
	@echo "SUCCESS: Standalone Baseline LIVE on http://localhost:4321"

.PHONY: standalone-down
standalone-up: ## Purge all local standalone processes and containers
	@pkill -f "escrow-api" || true
	@pkill -f "astro" || true
	@docker compose down -v
	@echo "Standalone environment purged."

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

.PHONY: pilot-down
pilot-down: ## Tier 2: Purge tripartite pilot workloads and namespaces
	@./scripts/gke-pilot.sh down

.PHONY: pilot-release
pilot-release: ## Authoritatively build and push AMD64 image to GCP Artifact Registry
	@echo "Releasing GKE Pilot API (Arch: amd64)..."
	@docker build --build-arg TARGETARCH=amd64 -t us-central1-docker.pkg.dev/vdcai-daml/escrow-platform-dev/escrow-api:latest .
	@docker push us-central1-docker.pkg.dev/vdcai-daml/escrow-platform-dev/escrow-api:latest
	@kubectl rollout restart deployment bank-api -n bank
	@kubectl rollout restart deployment buyer-api -n buyer
	@kubectl rollout restart deployment seller-api -n seller

## -- Ledger & State Synchronization --

.PHONY: bootstrap-local
bootstrap-local: ## Synchronize DAR packages and allocate Parties on localhost
	@./bin/ledger-sync -host localhost -port 7575 \
		-impl stablecoin-escrow \
		-iface stablecoin-escrow-interfaces \
		-out ledger-state.json

.PHONY: bootstrap-gke
bootstrap-gke: ## Authoritatively synchronize all tripartite nodes in GKE
	@for ns in bank buyer seller; do \
		echo "Bootstrapping GKE Namespace: $$ns"; \
		kubectl port-forward pod/$$ns-ledger-0 7575:7575 -n $$ns > /dev/null 2>&1 & \
		PID=$$!; sleep 15; \
		./bin/ledger-sync -host localhost -port 7575 -impl stablecoin-escrow -iface stablecoin-escrow-interfaces -out log/gke-state-$$ns.json || true; \
		kill $$PID || true; \
	done
