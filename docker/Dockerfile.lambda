# Stage with certs/zoneinfo
FROM alpine:latest AS assets

RUN apk add --no-cache ca-certificates tzdata

# Final Lambda image
FROM public.ecr.aws/lambda/provided:al2023

WORKDIR /var/task

# Copy ca-certificates from builder stage
COPY --from=assets /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=assets /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the Lambda binary
COPY bin/lambda ./main

# Copy config file
COPY configs/default.yaml configs/default.yaml

# Set the CMD to the handler
CMD ["./main"]