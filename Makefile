# MASQUE VPN Makefile

# Go source directories
CLIENT_DIR=./vpn_client
SERVER_DIR=./vpn_server
WEBUI_DIR=./admin_webui
COMMON_DIR=./common

# Output binary names
CLIENT_BIN=vpn-client
SERVER_BIN=vpn-server

# Docker configuration
DOCKER_REGISTRY ?= localhost:5000
IMAGE_TAG ?= latest
SERVER_IMAGE = $(DOCKER_REGISTRY)/masque-vpn-server:$(IMAGE_TAG)
CLIENT_IMAGE = $(DOCKER_REGISTRY)/masque-vpn-client:$(IMAGE_TAG)
WEBUI_IMAGE = $(DOCKER_REGISTRY)/masque-admin-webui:$(IMAGE_TAG)

# Default target
.PHONY: all
all: build-client build-server build-webui

# Build targets
.PHONY: build-client
build-client: build-client-win build-client-linux build-client-darwin

.PHONY: build-client-win
build-client-win:
	cd $(CLIENT_DIR) && GOOS=windows GOARCH=amd64 go build -o $(CLIENT_BIN).exe

.PHONY: build-client-linux
build-client-linux:
	cd $(CLIENT_DIR) && GOOS=linux GOARCH=amd64 go build -o $(CLIENT_BIN)

.PHONY: build-client-darwin
build-client-darwin:
	cd $(CLIENT_DIR) && GOOS=darwin GOARCH=amd64 go build -o $(CLIENT_BIN)

.PHONY: build-server
build-server: build-server-linux build-server-darwin

.PHONY: build-server-win
build-server-win:
	cd $(SERVER_DIR) && CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o $(SERVER_BIN).exe

.PHONY: build-server-linux
build-server-linux:
	cd $(SERVER_DIR) && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o $(SERVER_BIN)

.PHONY: build-server-darwin
build-server-darwin:
	cd $(SERVER_DIR) && CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o $(SERVER_BIN)

# Build Web UI
.PHONY: build-webui
build-webui:
	cd $(WEBUI_DIR) && npm install && npm run build

# Development targets
.PHONY: dev-server
dev-server:
	cd $(SERVER_DIR) && go run main.go -c config.server.toml

.PHONY: dev-client
dev-client:
	cd $(CLIENT_DIR) && go run main.go -c config.client.toml

.PHONY: dev-webui
dev-webui:
	cd $(WEBUI_DIR) && npm run dev

# Testing targets
.PHONY: test
test: test-unit test-integration

.PHONY: test-unit
test-unit:
	cd $(COMMON_DIR) && go test -v ./...
	cd $(SERVER_DIR) && go test -v ./...
	cd $(CLIENT_DIR) && go test -v ./...

.PHONY: test-integration
test-integration:
	cd tests/integration && go test -v ./...

.PHONY: test-load
test-load:
	cd tests/load && go test -v ./...

.PHONY: test-chaos
test-chaos:
	cd tests/chaos && ./chaos_monkey.sh

.PHONY: test-coverage
test-coverage:
	cd $(COMMON_DIR) && go test -coverprofile=coverage.out ./...
	cd $(SERVER_DIR) && go test -coverprofile=coverage.out ./...
	cd $(CLIENT_DIR) && go test -coverprofile=coverage.out ./...
	go tool cover -html=$(COMMON_DIR)/coverage.out -o coverage.html

.PHONY: test-race
test-race:
	cd $(COMMON_DIR) && go test -race -v ./...
	cd $(SERVER_DIR) && go test -race -v ./...
	cd $(CLIENT_DIR) && go test -race -v ./...

.PHONY: test-bench
test-bench:
	cd $(COMMON_DIR) && go test -bench=. -benchmem ./...
	cd $(SERVER_DIR) && go test -bench=. -benchmem ./...
	cd $(CLIENT_DIR) && go test -bench=. -benchmem ./...
	cd tests/load && go test -bench=. -benchmem ./...

# Docker targets
.PHONY: docker-build
docker-build: docker-build-server docker-build-client docker-build-webui

.PHONY: docker-build-server
docker-build-server:
	docker build -t $(SERVER_IMAGE) -f $(SERVER_DIR)/Dockerfile .

.PHONY: docker-build-client
docker-build-client:
	docker build -t $(CLIENT_IMAGE) -f $(CLIENT_DIR)/Dockerfile .

.PHONY: docker-build-webui
docker-build-webui:
	docker build -t $(WEBUI_IMAGE) -f $(WEBUI_DIR)/Dockerfile $(WEBUI_DIR)

.PHONY: docker-push
docker-push:
	docker push $(SERVER_IMAGE)
	docker push $(CLIENT_IMAGE)
	docker push $(WEBUI_IMAGE)

.PHONY: docker-run
docker-run:
	docker-compose up -d

.PHONY: docker-stop
docker-stop:
	docker-compose down

.PHONY: docker-logs
docker-logs:
	docker-compose logs -f

# Certificate generation
.PHONY: certs
certs:
	cd $(SERVER_DIR)/cert && ./gen_ca.sh && ./gen_server_keypair.sh

.PHONY: client-cert
client-cert:
	cd $(SERVER_DIR)/cert && ./gen_client_keypair.sh

# Dependency management
.PHONY: deps
deps:
	cd $(COMMON_DIR) && go mod tidy
	cd $(SERVER_DIR) && go mod tidy
	cd $(CLIENT_DIR) && go mod tidy
	cd $(WEBUI_DIR) && npm install

.PHONY: deps-update
deps-update:
	cd $(COMMON_DIR) && go get -u ./... && go mod tidy
	cd $(SERVER_DIR) && go get -u ./... && go mod tidy
	cd $(CLIENT_DIR) && go get -u ./... && go mod tidy
	cd $(WEBUI_DIR) && npm update

# Code quality
.PHONY: lint
lint:
	golangci-lint run ./$(COMMON_DIR)/...
	cd $(SERVER_DIR) && golangci-lint run ./...
	cd $(CLIENT_DIR) && golangci-lint run ./...
	cd $(WEBUI_DIR) && npm run lint

.PHONY: fmt
fmt:
	go fmt ./$(COMMON_DIR)/...
	cd $(SERVER_DIR) && go fmt ./...
	cd $(CLIENT_DIR) && go fmt ./...
	cd $(WEBUI_DIR) && npm run format

# Security scanning
.PHONY: security-scan
security-scan:
	gosec ./$(COMMON_DIR)/...
	cd $(SERVER_DIR) && gosec ./...
	cd $(CLIENT_DIR) && gosec ./...
	cd $(WEBUI_DIR) && npm audit

# Monitoring
.PHONY: monitoring-up
monitoring-up:
	docker-compose --profile monitoring up -d prometheus grafana

.PHONY: monitoring-down
monitoring-down:
	docker-compose --profile monitoring down

# Cleanup
.PHONY: clean
clean:
	rm -f $(CLIENT_DIR)/$(CLIENT_BIN) $(CLIENT_DIR)/$(CLIENT_BIN).exe
	rm -f $(SERVER_DIR)/$(SERVER_BIN) $(SERVER_DIR)/$(SERVER_BIN).exe
	cd $(WEBUI_DIR) && rm -rf dist node_modules
	docker-compose down --volumes --remove-orphans
	docker system prune -f

.PHONY: clean-all
clean-all: clean
	docker rmi -f $(SERVER_IMAGE) $(CLIENT_IMAGE) $(WEBUI_IMAGE) 2>/dev/null || true

# Help
.PHONY: help
help:
	@echo "MASQUE VPN Build System"
	@echo ""
	@echo "Build targets:"
	@echo "  all                 Build all components"
	@echo "  build-client        Build client for all platforms"
	@echo "  build-server        Build server for all platforms"
	@echo "  build-webui         Build web UI"
	@echo ""
	@echo "Development targets:"
	@echo "  dev-server          Run server in development mode"
	@echo "  dev-client          Run client in development mode"
	@echo "  dev-webui           Run web UI in development mode"
	@echo ""
	@echo "Testing targets:"
	@echo "  test                Run all tests"
	@echo "  test-unit           Run unit tests"
	@echo "  test-integration    Run integration tests"
	@echo "  test-load           Run load tests"
	@echo "  test-chaos          Run chaos tests"
	@echo ""
	@echo "Docker targets:"
	@echo "  docker-build        Build all Docker images"
	@echo "  docker-run          Start all services with docker-compose"
	@echo "  docker-stop         Stop all services"
	@echo "  docker-logs         Show logs from all services"
	@echo ""
	@echo "Utility targets:"
	@echo "  certs               Generate CA and server certificates"
	@echo "  client-cert         Generate client certificate"
	@echo "  deps                Install/update dependencies"
	@echo "  lint                Run code linters"
	@echo "  fmt                 Format code"
	@echo "  security-scan       Run security scanners"
	@echo "  clean               Clean build artifacts"
	@echo "  help                Show this help message"