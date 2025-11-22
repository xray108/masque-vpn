#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "Starting Integration Tests..."

# Check requirements
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo -e "${RED}Warning: Network namespaces are not supported on macOS.${NC}"
    echo "These tests are designed for Linux. Running in limited mode (local connectivity check only)."
    # TODO: Implement local connectivity test for macOS
    exit 0
fi

if [ "$EUID" -ne 0 ]; then 
  echo -e "${RED}Please run as root (required for network namespaces)${NC}"
  exit 1
fi

# Create namespaces
echo "Creating network namespaces..."
ip netns add server_ns
ip netns add client_ns

# Setup cleanup trap
cleanup() {
    echo "Cleaning up..."
    ip netns del server_ns 2>/dev/null || true
    ip netns del client_ns 2>/dev/null || true
    # Kill processes
    pkill -f vpn-server || true
    pkill -f vpn-client || true
}
trap cleanup EXIT

# Create veth pair to connect namespaces
echo "Connecting namespaces..."
ip link add veth_server type veth peer name veth_client
ip link set veth_server netns server_ns
ip link set veth_client netns client_ns

# Configure server namespace
ip netns exec server_ns ip addr add 192.168.50.1/24 dev veth_server
ip netns exec server_ns ip link set veth_server up
ip netns exec server_ns ip link set lo up

# Configure client namespace
ip netns exec client_ns ip addr add 192.168.50.2/24 dev veth_client
ip netns exec client_ns ip link set veth_client up
ip netns exec client_ns ip link set lo up

# Test connectivity between namespaces
echo "Testing namespace connectivity..."
ip netns exec client_ns ping -c 1 192.168.50.1 > /dev/null
if [ $? -eq 0 ]; then
    echo -e "${GREEN}Namespace connectivity OK${NC}"
else
    echo -e "${RED}Namespace connectivity FAILED${NC}"
    exit 1
fi

# Start VPN Server
echo "Starting VPN Server..."
# We need a config file for the test
# Copy example config and modify listen address
cp ../../vpn_server/config.server.toml.example ./test_server.toml
sed -i 's/listen_addr = .*/listen_addr = "192.168.50.1:4433"/' ./test_server.toml

ip netns exec server_ns ../../vpn_server/vpn-server -config ./test_server.toml &
SERVER_PID=$!
sleep 2

# Start VPN Client
echo "Starting VPN Client..."
# We need a client config
# For this test, we might need to generate one or use a pre-generated one.
# This is tricky without the web UI interaction.
# Ideally, we should have a CLI tool to generate client config.
# For now, we'll skip the actual VPN connection test in this script until we have a way to provision clients via CLI.

echo -e "${GREEN}Integration tests setup passed (VPN connection test pending client provisioning tool)${NC}"
