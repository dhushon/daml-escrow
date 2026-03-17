APP_NAME=escrow-api
PORT=8080
DPM=/Users/dhushon/.dpm/bin/dpm

.PHONY: build
build:
	go build -o bin/$(APP_NAME) ./cmd/escrow-api

.PHONY: daml-build
daml-build:
	cd contracts && $(DPM) build

.PHONY: sandbox
sandbox:
	cd contracts && export JAVA_HOME="/opt/homebrew/opt/openjdk@17" && \
	nohup $(DPM) sandbox --config sandbox.conf --dar .daml/dist/stablecoin-escrow-0.0.1.dar > sandbox.log 2>&1 &

.PHONY: ledger-setup
ledger-setup:
	cd contracts && export JAVA_HOME="/opt/homebrew/opt/openjdk@17" && \
	$(DPM) canton-console --port 6865 --bootstrap init.canton --no-tty

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
	rm -rf bin api.log api_new.log api_v2_local.log api_v2_local_final.log api_v2_final.log