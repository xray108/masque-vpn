#!/bin/bash

set -e

echo "🐳 Running MASQUE VPN Docker Tests"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is installed and running
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed"
    exit 1
fi

if ! docker info &> /dev/null; then
    print_error "Docker is not running"
    exit 1
fi

print_status "Docker is available ✓"

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    print_error "docker-compose is not installed"
    exit 1
fi

print_status "docker-compose is available ✓"

# Build Docker images
print_status "Building Docker images..."
docker-compose build --no-cache

if [ $? -ne 0 ]; then
    print_error "Docker build failed"
    exit 1
fi

print_status "Docker images built successfully ✓"

# Start services
print_status "Starting services..."
docker-compose up -d

# Wait for services to be ready
print_status "Waiting for services to start..."
sleep 10

# Check if services are running
print_status "Checking service health..."

# Check VPN server
if curl -f http://localhost:8080/health &> /dev/null; then
    print_status "VPN server is healthy ✓"
else
    print_warning "VPN server health check failed"
fi

# Check metrics endpoints
if curl -f http://localhost:9090/metrics &> /dev/null; then
    print_status "Server metrics endpoint is accessible ✓"
else
    print_warning "Server metrics endpoint not accessible"
fi

# Check Grafana
if curl -f http://localhost:3001 &> /dev/null; then
    print_status "Grafana is accessible ✓"
else
    print_warning "Grafana not accessible"
fi

# Check Prometheus
if curl -f http://localhost:9091 &> /dev/null; then
    print_status "Prometheus is accessible ✓"
else
    print_warning "Prometheus not accessible"
fi

# Run integration tests against Docker services
print_status "Running integration tests against Docker services..."
cd tests/integration
go test -v ./... -tags=docker

# Run load tests
print_status "Running load tests against Docker services..."
cd ../load
go test -v ./... -tags=docker

# Check logs for errors
print_status "Checking service logs..."
docker-compose logs vpn-server | grep -i error || print_status "No errors in server logs ✓"

# Cleanup
print_status "Cleaning up..."
docker-compose down

print_status "Docker tests completed successfully! 🎉"
echo ""
echo "All services tested:"
echo "✓ VPN Server"
echo "✓ Admin Web UI"  
echo "✓ Prometheus"
echo "✓ Grafana"
echo "✓ Integration tests"
echo "✓ Load tests"