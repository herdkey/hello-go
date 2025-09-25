package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi/v5"

	"github.com/herdkey/hello-go/internal/config"
	"github.com/herdkey/hello-go/internal/logging"
	"github.com/herdkey/hello-go/internal/router"
	"github.com/herdkey/hello-go/internal/telemetry"
)

var chiLambda *chiadapter.ChiLambda

func init() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := logging.Setup(cfg.Logging)

	telemetryProvider, err := telemetry.Setup(ctx, cfg.Telemetry)
	if err != nil {
		logger.Error("failed to setup telemetry", "error", err)
		os.Exit(1)
	}
	defer func() {
		if shutdownErr := telemetryProvider.Shutdown(ctx); shutdownErr != nil {
			logger.Error("failed to shutdown telemetry", "error", shutdownErr)
		}
	}()

	r := router.BuildRouter(logger)
	chiLambda = chiadapter.New(r.(*chi.Mux))
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return chiLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(handler)
}
