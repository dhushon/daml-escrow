APP_NAME=escrow-api
PORT=8080

.PHONY: build
build:
	go build -o bin/$(APP_NAME) ./cmd/escrow-api

.PHONY: run
run:
	go run ./cmd/escrow-api

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: docker-build
docker-build:
	docker build -t $(APP_NAME):latest .

.PHONY: docker-run
docker-run:
	docker run -p $(PORT):8080 $(APP_NAME):latest

.PHONY: compose-up
compose-up:
	docker compose up --build

.PHONY: compose-down
compose-down:
	docker compose down

.PHONY: clean
clean:
	rm -rf bin