# Stage with certs/zoneinfo
FROM alpine:latest AS assets

RUN apk add --no-cache ca-certificates tzdata

# Final Lambda image
FROM public.ecr.aws/lambda/provided:al2023

# Default directly used by lambda. We're just being explicit.
WORKDIR /var/task

# Copy ca-certificates from builder stage
COPY --from=assets /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=assets /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the Lambda binary
COPY bin/lambda ./main

# Ensure the binary is executable
RUN chmod 755 /var/task/main

# Copy config file
COPY configs/default.yml configs/default.yml

# Set the ENTRYPOINT to the handler
ENTRYPOINT ["/var/task/main"]
