# Stage with certs/zoneinfo
FROM alpine:latest AS assets

RUN apk add --no-cache ca-certificates tzdata

# Final minimal image
FROM scratch

WORKDIR /app

# Copy ca-certificates from builder stage
COPY --from=assets /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=assets /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY bin/hello-go-api .

# Copy config file
COPY configs/default.yaml configs/default.yaml

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["./hello-go-api", "--help"]

# Run the binary
ENTRYPOINT ["./hello-go-api"]
CMD ["serve"]
