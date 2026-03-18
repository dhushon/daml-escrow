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

.PHONY: daml-build
daml-build:
	@echo "Building Daml Contracts..."
	cd contracts && $(DPM) build

.PHONY: clean-db
clean-db:
	@echo "Wiping persistent ledger database..."
	docker-compose exec -T postgres psql -U escrow -d postgres -c "DROP DATABASE IF EXISTS escrow;"
	docker-compose exec -T postgres psql -U escrow -d postgres -c "CREATE DATABASE escrow;"

.PHONY: clean-ledger
clean-ledger:
	@echo "Stopping local Canton processes..."
	-pkill -f "canton"
	rm -f contracts/sandbox.log

.PHONY: sandbox-local
sandbox-local: clean-ledger daml-build
	@echo "Starting Canton Sandbox locally (background)..."
	cd contracts && export JAVA_HOME="$(JAVA_HOME_17)" && \
	nohup $(DPM) sandbox --config $(SANDBOX_CONF) > sandbox.log 2>&1 &

.PHONY: sandbox-wait
sandbox-wait:
	@echo "Waiting for sandbox port 6865..."
	@until lsof -i :6865 > /dev/null 2>&1; do sleep 1; done
	@echo "Sandbox port is open. Waiting 5s for internal stability..."
	@sleep 5

.PHONY: sandbox-setup
sandbox-setup:
	@echo "Establishing topology and uploading DAR..."
	cd contracts && export JAVA_HOME="$(JAVA_HOME_17)" && \
	$(DPM) canton-console --port 6866 --bootstrap $(SANDBOX_INIT) --no-tty

.PHONY: sandbox-up
sandbox-up: clean-db sandbox-local sandbox-wait sandbox-setup
	@echo "---------------------------------------"
	@echo "Sandbox is UP and fully initialized."
	@echo "JSON API: http://localhost:7575"
	@echo "gRPC API: localhost:6865"
	@echo "---------------------------------------"

.PHONY: docker-up
docker-up:
	@echo "Starting Docker Compose stack..."
	docker-compose up -d --build
	@echo "Monitoring ledger logs for readiness..."
	@count=0; until docker-compose logs sandbox | grep -q "Setup complete." || [ $$count -eq 60 ]; do \
		sleep 5; \
		count=$$((count + 1)); \
	done
	@docker-compose logs sandbox | grep -q "Setup complete." && echo "Ledger reports: Setup complete." || (echo "Timed out waiting for ledger setup"; exit 1)

.PHONY: docker-down
docker-down:
	@echo "Stopping Docker Compose stack..."
	docker-compose down -v

.PHONY: test
test:
	go test ./...

.PHONY: integration-test
integration-test:
	@echo "Running integration tests..."
	go test -v -tags integration ./internal/ledger/ledger_integration_test.go \
		./internal/ledger/json_client.go \
		./internal/ledger/daml_client.go \
		./internal/ledger/client.go \
		./internal/ledger/stablecoin.go

.PHONY: clean
clean:
	rm -rf bin *.log
