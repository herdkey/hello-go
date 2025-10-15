package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi/v5"

	"github.com/savisec/hello-go/internal/config"
	"github.com/savisec/hello-go/internal/logging"
	"github.com/savisec/hello-go/internal/router"
	"github.com/savisec/hello-go/internal/telemetry"
)

var (
	chiLambda         *chiadapter.ChiLambdaV2
	telemetryProvider *telemetry.Provider
	logger            *slog.Logger
)

func init() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger = logging.Setup(cfg.Logging)

	telemetryProvider, err = telemetry.Setup(ctx, cfg.Telemetry)
	if err != nil {
		logger.Error("failed to setup telemetry", "error", err)
		os.Exit(1)
	}

	r := router.BuildRouter(logger)
	chiLambda = chiadapter.NewV2(r.(*chi.Mux))

	// Set up signal handler for graceful shutdown
	go shutdownHook()
}

func shutdownHook() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	if shutdownErr := telemetryProvider.Shutdown(context.Background()); shutdownErr != nil {
		logger.Error("failed to shutdown telemetry", "error", shutdownErr)
	}
	os.Exit(0)
}

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return chiLambda.ProxyWithContextV2(ctx, req)
}

func main() {
	lambda.Start(handler)
}
