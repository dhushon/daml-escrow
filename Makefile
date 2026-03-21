# Makefile for Stablecoin Escrow Platform

APP_NAME=escrow-api
DPM=/Users/dhushon/.dpm/bin/dpm
JAVA_HOME_17=/opt/homebrew/opt/openjdk@17

# Ledger Config Paths
SANDBOX_CONF=Sandbox/sandbox.conf
SANDBOX_INIT=Sandbox/sandbox_init.canton
SETUP_SCRIPT=scripts/setup_users.sh

.PHONY: build
build:
	@echo "Building Go API..."
	go build -o bin/$(APP_NAME) ./cmd/escrow-api
	@echo "Building Oracle Simulator..."
	go build -o bin/oracle-simulator ./cmd/oracle-simulator

.PHONY: oracle-sim
oracle-sim:
	@go run ./cmd/oracle-simulator/main.go

.PHONY: daml-build
daml-build:
	@echo "Building Daml Contracts..."
	cd contracts && $(DPM) build --all

.PHONY: clean-db
clean-db:
	@echo "Wiping persistent ledger database..."
	docker-compose exec -T postgres psql -U escrow -d postgres -c "DROP DATABASE IF EXISTS escrow;" || true
	docker-compose exec -T postgres psql -U escrow -d postgres -c "CREATE DATABASE escrow;" || true

.PHONY: docker-up
docker-up:
	@echo "Starting Docker Compose stack (Ledger + DB)..."
	docker-compose up -d --build
	@echo "Monitoring ledger logs for readiness..."
	@count=0; until docker-compose logs sandbox | grep -q "Setup complete." || [ $$count -eq 60 ]; do \
		sleep 5; \
		count=$$((count + 1)); \
	done
	@docker-compose logs sandbox | grep -q "Setup complete." && echo "Ledger reports: Setup complete." || (echo "Timed out waiting for ledger setup"; exit 1)

.PHONY: api-up
api-up: build
	@echo "Starting Go API on port 8081 (background)..."
	@nohup bin/$(APP_NAME) > api.log 2>&1 & echo $$! > api.pid
	@echo "API started. PID: $$(cat api.pid). Logs: api.log"

.PHONY: frontend-up
frontend-up:
	@echo "Starting Astro Frontend on port 8080 (background)..."
	@cd frontend && nohup npm run dev > ../frontend.log 2>&1 & echo $$! > frontend.pid
	@echo "Frontend started. PID: $$(cat frontend.pid). Logs: frontend.log"

.PHONY: up
up: docker-up api-up frontend-up
	@echo "---------------------------------------"
	@echo "Full stack is UP."
	@echo "Web Dashboard: http://localhost:8080"
	@echo "Backend API:   http://localhost:8081"
	@echo "Ledger JSON:   http://localhost:7575"
	@echo "---------------------------------------"

.PHONY: down
down:
	@echo "Stopping all components..."
	@-[ -f api.pid ] && kill $$(cat api.pid) && rm api.pid || true
	@-[ -f frontend.pid ] && kill $$(cat frontend.pid) && rm frontend.pid || true
	@pkill -f "npm run dev" || true
	docker-compose down
	@echo "Stack is down."

.PHONY: test
test:
	go test ./...

.PHONY: integration-test
integration-test:
	@echo "Running integration tests..."
	go test -v -tags integration ./internal/ledger/ledger_integration_test.go \
		./internal/ledger/json_base.go \
		./internal/ledger/json_parser.go \
		./internal/ledger/json_parties.go \
		./internal/ledger/json_escrows.go \
		./internal/ledger/json_settlements.go \
		./internal/ledger/daml_client.go \
		./internal/ledger/client.go \
		./internal/ledger/stablecoin.go

.PHONY: clean
clean:
	rm -rf bin *.log *.pid
