import '.justfiles/go/base.just'


################################################################################
#                                 Standard API                                 #
################################################################################

# Build for distribution
[group('build')]
build-api goos="linux": (build-cmd "api" goos)

[group('docker')]
docker-build-api: build-api
    "{{ justfile_dir() }}/docker/api/build.sh"

# Start API docker image
[group('docker')]
docker-up-api: docker-build-api
    docker compose -f docker/api/compose.yml up -d


################################################################################
#                                Lambda Handler                                #
################################################################################

# Build Lambda binary
[group('build')]
build-lambda goos="linux": (build-cmd "lambda" goos)

# Build Lambda docker image
[group('docker')]
docker-build-lambda: build-lambda
    "{{ justfile_dir() }}/docker/lambda/build.sh"

# Start Lambda docker image
[group('docker')]
docker-up-lambda: docker-build-lambda
    docker compose -f docker/lambda/compose.yml up -d

################################################################################
#                                Combined Build                                #
################################################################################

# Build all binaries
[group('build')]
build goos="linux": (build-api goos) (build-lambda goos)
