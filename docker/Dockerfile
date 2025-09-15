# Set by .envrc via direnv
ARG GO_VERSION

# Build stage
FROM golang:${GO_VERSION}-alpine AS builder

# Install git and other build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY configs/default.yaml configs/default.yaml
COPY "./cmd" "./cmd"
COPY ./internal ./internal
COPY ./api ./api
RUN mkdir bin

# Build the application with CGO disabled for a static binary
# NOTE: -s -w remove debugging info so will make attaching a debugger impossible
RUN CGO_ENABLED=0 GOOS=linux \
    go build -a \
    -ldflags '-s -w -extldflags=-static' \
    -o bin/hello-go \
    "./cmd/hello-go"

# Final stage
FROM scratch

# Copy ca-certificates from builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/bin/hello-go /hello-go

# Copy config file
COPY --from=builder /app/configs/default.yaml /configs/default.yaml

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/hello-go", "--help"]

# Run the binary
ENTRYPOINT ["/hello-go"]
CMD ["serve"]
