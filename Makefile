# Phase 7: Production-Ready Orchestration (Makefile)

# Architecture Detection
LOCAL_ARCH := $(shell uname -m | sed 's/x86_64/amd64/' | sed 's/arm64/arm64/')
ARCH ?= $(LOCAL_ARCH)

.PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

## -- Basline Orchestration (Single Node) --

.PHONY: standalone-up
standalone-up: ## Launch the foundational standalone ledger and API (Bypass Enabled)
	@echo "Bringing down any existing processes..."
	@$(MAKE) standalone-down
	@echo "Launching standalone ledger..."
	@docker-compose up -d
	@echo "Awaiting ledger bootstrapping (60s)..."
	@sleep 60
	@$(MAKE) sync
	@echo "Launching Go API (8081)..."
	@screen -d -m -S escrow-api bash -c "USER_CONFIG_DSN=\"postgres://escrow:escrow@localhost:5432/user_config?sslmode=disable\" PARTICIPANT_ID=\"bank\" LEDGER_NODES_BANK_HOST=\"localhost\" go run ./cmd/escrow-api serve --notls --bypass > log/standalone-api.log 2>&1"
	@echo "Launching Astro UX (4321)..."
	@cd frontend && screen -d -m -S escrow-ux bash -c "PUBLIC_API_URL=http://localhost:8081 ./node_modules/.bin/astro dev --host 127.0.0.1 --port 4321 > ../log/standalone-frontend.log 2>&1"
	@echo "Verifying services..."
	@for i in {1..20}; do \
		if lsof -i:4321 > /dev/null && curl -s http://localhost:8081/api/v1/health > /dev/null; then \
			echo "SUCCESS: Standalone Baseline LIVE on http://localhost:4321"; \
			exit 0; \
		fi; \
		echo "Waiting for services... ($$i/20)"; \
		sleep 2; \
	done; \
	echo "ERROR: Services failed to start properly. Check logs in log/ directory."; \
	exit 1

.PHONY: standalone-down
standalone-down: ## Definitive purge of all standalone containers and local processes
	@docker-compose down -v --remove-orphans > /dev/null 2>&1 || true
	@screen -S escrow-api -X quit > /dev/null 2>&1 || true
	@screen -S escrow-ux -X quit > /dev/null 2>&1 || true
	@lsof -ti:8081 | xargs kill -9 > /dev/null 2>&1 || true
	@lsof -ti:4321 | xargs kill -9 > /dev/null 2>&1 || true
	@pkill -f "astro dev" > /dev/null 2>&1 || true
	@pkill -f "escrow-api" > /dev/null 2>&1 || true
	@echo "Standalone environment purged."

## -- High-Assurance Orchestration (Distributed) --

.PHONY: local-up
local-up: ## Launch the local tripartite stack (Optimized for ARM64)
	@echo "Launching local tripartite stack (Arch: $(ARCH))..."
	@TARGETARCH=$(ARCH) docker-compose -f docker-compose.distributed.yml -f docker-compose.otel.yml up -d --build
	@sleep 15
	@$(MAKE) sync

.PHONY: local-down
local-down: ## Definitive purge of all local tripartite and observability containers
	@docker-compose -f docker-compose.distributed.yml -f docker-compose.otel.yml down -v --remove-orphans

## -- GKE Pilot Orchestration --

.PHONY: pilot-up
pilot-up: ## Authoritatively launch the full GKE Pilot (api.vdatacloudai.com)
	@./scripts/gke-pilot.sh up

.PHONY: pilot-down
pilot-down: ## Definitive purge of all GKE pilot workloads and namespaces
	@./scripts/gke-pilot.sh down

.PHONY: pilot-release
pilot-release: ## Build and push the high-assurance AMD64 image to GCP
	@echo "Releasing GKE Pilot API (Arch: amd64)..."
	@docker build --build-arg TARGETARCH=amd64 -t us-central1-docker.pkg.dev/vdcai-daml/escrow-platform-dev/escrow-api:latest .
	@docker push us-central1-docker.pkg.dev/vdcai-daml/escrow-platform-dev/escrow-api:latest
	@kubectl rollout restart deployment bank-api -n bank
	@kubectl rollout restart deployment buyer-api -n buyer
	@kubectl rollout restart deployment seller-api -n seller

## -- Core Utilities --

.PHONY: sync
sync: ## Synchronize ledger state (resolve Package and Party IDs)
	@echo "Discovering ledger state on localhost:7575..."
	@./bin/ledger-sync -host localhost -port 7575 \
		-impl stablecoin-escrow \
		-iface stablecoin-escrow-interfaces \
		-out ledger-state.json
