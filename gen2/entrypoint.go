package gen2

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	googlepropagator "github.com/GoogleCloudPlatform/opentelemetry-operations-go/propagator"
	"github.com/ebi-yade/cloud-functions-samples/gen2/app"
	"github.com/ebi-yade/cloud-functions-samples/gen2/app/handlers"
	"github.com/ebi-yade/cloud-functions-samples/gen2/infra/topic"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func init() {
	// https://cloud.google.com/blog/ja/products/application-development/graceful-shutdowns-cloud-run-deep-dive
	const gracePeriod = 5 * time.Second // shorter than Cloud Run's grace period
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	// ==============================================================
	// Setup observability solutions
	// ==============================================================
	logger := slog.New(NewLogHandler(os.Stderr, slog.LevelInfo))
	slog.SetDefault(logger)
	// maybe you want to get the project ID from the metadata server
	projectID := mustEnv("GOOGLE_CLOUD_PROJECT")
	slog.SetDefault(logger.With(slog.String("project_id", projectID)))

	propagators := []propagation.TextMapPropagator{
		googlepropagator.CloudTraceOneWayPropagator{},
		propagation.TraceContext{},
		propagation.Baggage{},
	}
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagators...))
	tp, err := NewTracerProvider(ctx, projectID, 0.1)
	if err != nil {
		fatal(ctx, errors.Wrap(err, "error NewTracerProvider"))
	}
	otel.SetTracerProvider(tp)

	// ==============================================================
	// Initialize infrastructure dependencies
	// ==============================================================
	topicID := mustEnv("PUBSUB_TOPIC_ID")
	pubsubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		fatal(ctx, errors.Wrap(err, "error pubsub.NewClient"))
	}
	googleTopic := topic.NewGoogleTopic(pubsubClient.Topic(topicID))
	slog.InfoContext(ctx, "initialized pubsub topic", slog.String("topic", topicID))

	// ==============================================================
	// Register HTTP / Event-driven handlers
	// ==============================================================
	h := handlers.New(googleTopic)
	functionsHTTP("functions-samples-start", app.WrapHTTP(nil, h.Start))

	// ==============================================================
	// Start an asynchronous routine to handle shutdown signals
	// ==============================================================
	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), gracePeriod)
		defer cancel()

		slog.InfoContext(ctx, "shutting down...")
		if err := tp.ForceFlush(ctx); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("error ForceFlush: %+v", err))
		}
		if err := pubsubClient.Close(); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("error pubsubClient.Close: %+v", err))
		}
		slog.InfoContext(ctx, "shutdown completed. bye!")
	}()
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("missing env: %s", key))
	}
	slog.Info("detected value from environment", slog.String("key", key), slog.String("value", v))
	return v
}

func fatal(ctx context.Context, err error) {
	slog.ErrorContext(ctx, fmt.Sprintf("exit with: %+v", err))
	os.Exit(1)
}

// functionsHTTP は HTTP 関数を登録するための functions.HTTP をラップして otel に対応させたものです。
func functionsHTTP(entrypoint string, stdHandler http.HandlerFunc) {
	otelHandler := otelhttp.NewHandler(stdHandler, entrypoint)
	functions.HTTP(entrypoint, otelHandler.ServeHTTP)
}
