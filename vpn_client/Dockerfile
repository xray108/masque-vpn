# Multi-stage build for MASQUE VPN Client
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    gcc \
    musl-dev \
    ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
COPY ../common/go.mod ../common/go.sum ../common/

# Download dependencies
RUN go mod download

# Copy source code
COPY . .
COPY ../common ../common/

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o vpn-client .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    iptables \
    ip6tables \
    iproute2 \
    tzdata

# Create non-root user
RUN addgroup -g 1001 vpn && \
    adduser -D -u 1001 -G vpn vpn

# Create directories
RUN mkdir -p /app/cert /app/logs && \
    chown -R vpn:vpn /app

# Copy binary from builder
COPY --from=builder /app/vpn-client /app/vpn-client
COPY --from=builder /app/config.client.toml.example /app/config.client.toml

# Set permissions
RUN chmod +x /app/vpn-client

# Note: Client needs NET_ADMIN capability to create TUN device
# This will be set in docker-compose or kubernetes deployment

# Switch to non-root user
USER vpn

# Set working directory
WORKDIR /app

# Expose metrics port
EXPOSE 9092/tcp

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9092/health || exit 1

# Default command
CMD ["./vpn-client", "-c", "config.client.toml"]