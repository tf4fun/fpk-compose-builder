# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o fpk-compose-builder ./cmd/fpk-compose-builder

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache bash docker-cli curl

# Download and install fnpack CLI
ARG FNPACK_VERSION=1.0.4
RUN curl -L -o /usr/local/bin/fnpack \
    "https://static2.fnnas.com/fnpack/fnpack-${FNPACK_VERSION}-linux-amd64" && \
    chmod +x /usr/local/bin/fnpack

# Copy the built binary from builder stage
COPY --from=builder /app/fpk-compose-builder /usr/local/bin/fpk-compose-builder

# Set entrypoint
ENTRYPOINT ["fpk-compose-builder"]
