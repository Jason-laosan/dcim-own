.PHONY: build test clean run docker-build install-deps

# Build variables
BINARY_NAME=collector
BUILD_DIR=./bin
CMD_DIR=./cmd/collector
DOCKER_IMAGE=opc-collector
DOCKER_TAG=latest

# Go build flags
LDFLAGS=-ldflags "-s -w"
GOFLAGS=-trimpath

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)

install-deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

test:
	@echo "Running tests..."
	go test -v -race -cover ./...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

run: build
	@echo "Running $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME) -config configs/config.yaml

docker-build:
	@echo "Building Docker image..."
	docker build -f deploy/docker/Dockerfile.collector -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run:
	@echo "Running Docker container..."
	docker run -v $(PWD)/configs:/etc/opc-collector $(DOCKER_IMAGE):$(DOCKER_TAG)

lint:
	@echo "Running linter..."
	golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
	@echo "Running go vet..."
	go vet ./...

help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  build-linux    - Build for Linux"
	@echo "  install-deps   - Install dependencies"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  clean          - Clean build artifacts"
	@echo "  run            - Build and run"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
