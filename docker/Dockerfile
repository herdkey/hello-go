# Build stage
FROM golang:1.24.6-alpine AS builder

# Install git and other build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum config.default.yaml ./

# Download dependencies
RUN go mod download

# Copy source code
COPY "./cmd" "./cmd"
COPY ./internal ./internal
COPY ./api ./api

# Build the application with CGO disabled for a static binary
RUN CGO_ENABLED=0 GOOS=linux \
    go build -a \
    -installsuffix cgo \
    -ldflags '-extldflags "-static"' \
    -o bin/hello-go \
    ./cmd/hello-go

# Final stage
FROM scratch

# Copy ca-certificates from builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/bin/hello-go /hello-go

# Copy config file
COPY --from=builder /app/config.default.yaml /config.default.yaml

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/hello-go", "--help"]

# Run the binary
ENTRYPOINT ["/hello-go"]
CMD ["serve"]
