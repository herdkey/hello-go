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
COPY bin/api .

# Copy config file
COPY configs/default.yml configs/default.yml

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./api"]
CMD ["serve"]
