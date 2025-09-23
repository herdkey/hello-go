import '.justfiles/go/base.just'

# Build for distribution
[group('build')]
build: (build_cmd "api")
