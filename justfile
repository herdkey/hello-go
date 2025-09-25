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

# Build Lambda Docker image
[group('docker')]
docker-build-lambda: build-lambda
    docker buildx build --platform linux/amd64 --provenance=false -f docker/Dockerfile.lambda -t hello-go-lambda:local .
