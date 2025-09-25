import '.justfiles/go/base.just'

# Build for distribution
[group('build')]
build: (build_cmd "api")

# Build Lambda binary
[group('build')]
build-lambda: (build_cmd "lambda")

[group('docker')]
docker-build: build
    "{{ justfile_dir() }}/scripts/docker_build.sh"

# Build Lambda docker image
[group('docker')]
docker-build-lambda: build-lambda
    "{{ justfile_dir() }}/scripts/docker_build_lambda.sh"

# Start API docker image
[group('docker')]
docker-up: docker-build
    docker compose -f docker/compose.yml up -d

# Start Lambda docker image
[group('docker')]
docker-up-lambda: docker-build-lambda
    docker compose -f docker/compose-lambda.yml up -d

# Run Lambda integration tests (assumes Lambda is already running)
[group('test')]
test-lambda:
    go test -v ./tests/integration/lambda_test.go
