# Multi-stage build for MASQUE VPN Server
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    gcc \
    musl-dev \
    sqlite-dev \
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
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o vpn-server .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    iptables \
    ip6tables \
    iproute2 \
    sqlite \
    tzdata

# Create non-root user
RUN addgroup -g 1001 vpn && \
    adduser -D -u 1001 -G vpn vpn

# Create directories
RUN mkdir -p /app/cert /app/data /app/logs && \
    chown -R vpn:vpn /app

# Copy binary from builder
COPY --from=builder /app/vpn-server /app/vpn-server
COPY --from=builder /app/config.server.toml.example /app/config.server.toml

# Copy certificate generation scripts
COPY cert/ /app/cert/

# Set permissions
RUN chmod +x /app/vpn-server /app/cert/*.sh

# Switch to non-root user for security
USER vpn

# Set working directory
WORKDIR /app

# Expose ports
EXPOSE 4433/udp 8080/tcp 9090/tcp

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default command
CMD ["./vpn-server", "-c", "config.server.toml"]