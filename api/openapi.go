package api

import (
	_ "embed"
)

// Embed the OpenAPI specification file

//go:embed openapi.yml
var OpenAPISpec []byte
