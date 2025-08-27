package api

import (
	_ "embed"
)

// Embed the OpenAPI specification file

//go:embed openapi.yaml
var OpenAPISpec []byte
