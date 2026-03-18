APP_NAME=escrow-api
PORT=8080
DPM=/Users/dhushon/.dpm/bin/dpm
JAVA_HOME_17=/opt/homebrew/opt/openjdk@17

.PHONY: build
build:
	go build -o bin/$(APP_NAME) ./cmd/escrow-api

.PHONY: daml-build
daml-build:
	cd contracts && $(DPM) build

.PHONY: sandbox
sandbox:
	@echo "Starting Canton Sandbox 3.x..."
	cd contracts && export JAVA_HOME="$(JAVA_HOME_17)" && \
	nohup $(DPM) sandbox --config sandbox.conf --dar .daml/dist/stablecoin-escrow-0.0.1.dar > sandbox.log 2>&1 &
	@echo "Sandbox started in background. Use 'make ledger-setup' once ready."

.PHONY: ledger-setup
ledger-setup:
	@echo "Uploading latest DAR and establishing topology..."
	cd contracts && export JAVA_HOME="$(JAVA_HOME_17)" && \
	$(DPM) canton-console --port 6865 --bootstrap init.canton --no-tty

.PHONY: restart-ledger
restart-ledger: clean-ledger daml-build sandbox
	@echo "Waiting for sandbox..."
	@sleep 15
	@$(MAKE) ledger-setup

.PHONY: clean-ledger
clean-ledger:
	@echo "Stopping sandbox..."
	-pkill -f "canton"
	rm -f contracts/sandbox.log

.PHONY: run
run: build
	./bin/$(APP_NAME)

.PHONY: test
test:
	go test ./...

.PHONY: integration-test
integration-test:
	go test -v ./internal/ledger/ledger_integration_test.go ./internal/ledger/json_client.go ./internal/ledger/daml_client.go ./internal/ledger/client.go ./internal/ledger/stablecoin.go

.PHONY: clean
clean:
	rm -rf bin *.log
