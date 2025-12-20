#!/bin/bash

set -e

echo "🧪 Running MASQUE VPN Local Tests"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.25 or later."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.25"

if ! printf '%s\n%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V -C; then
    print_error "Go version $GO_VERSION is too old. Required: $REQUIRED_VERSION or later."
    exit 1
fi

print_status "Go version $GO_VERSION detected ✓"

# Update dependencies
print_status "Updating dependencies..."
cd common && go mod tidy
cd ../vpn_server && go mod tidy  
cd ../vpn_client && go mod tidy
cd ../tests/integration && go mod tidy
cd ../load && go mod tidy
cd ../..

# Run unit tests
print_status "Running unit tests..."
echo "Testing common package..."
cd common && go test -v ./...
if [ $? -ne 0 ]; then
    print_error "Common package tests failed"
    exit 1
fi

echo "Testing VPN client..."
cd ../vpn_client && go test -v ./...
if [ $? -ne 0 ]; then
    print_error "VPN client tests failed"
    exit 1
fi

# Test with race detection
print_status "Running race condition tests..."
cd ../common && go test -race -v ./...
cd ../vpn_client && go test -race -v ./...

# Generate coverage report
print_status "Generating coverage report..."
cd ../common && go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
print_status "Coverage report generated: common/coverage.html"

# Build binaries to ensure they compile
print_status "Building binaries..."
cd ../vpn_client && go build -o vpn-client .
if [ $? -eq 0 ]; then
    print_status "VPN client built successfully ✓"
    rm -f vpn-client vpn-client.exe
else
    print_error "VPN client build failed"
    exit 1
fi

# Test configuration files
print_status "Validating configuration files..."
if [ -f "config.client.toml.example" ]; then
    print_status "Client config example found ✓"
else
    print_warning "Client config example not found"
fi

# Run integration tests (if services are running)
print_status "Running integration tests..."
cd ../tests/integration && go test -v ./...
if [ $? -ne 0 ]; then
    print_warning "Integration tests failed (services may not be running)"
else
    print_status "Integration tests passed ✓"
fi

# Run load tests
print_status "Running load tests..."
cd ../load && go test -short -v ./...
if [ $? -ne 0 ]; then
    print_warning "Load tests failed (services may not be running)"
else
    print_status "Load tests passed ✓"
fi

# Run benchmarks
print_status "Running benchmarks..."
cd ../../common && go test -bench=. -benchmem ./...
cd ../vpn_client && go test -bench=. -benchmem ./...

print_status "All local tests completed successfully! 🎉"
echo ""
echo "Next steps:"
echo "1. Start the VPN server: cd vpn_server && sudo ./vpn-server"
echo "2. Start the VPN client: cd vpn_client && sudo ./vpn-client"  
echo "3. Run full integration tests: make test-integration"
echo "4. Check metrics: curl http://localhost:9092/metrics"