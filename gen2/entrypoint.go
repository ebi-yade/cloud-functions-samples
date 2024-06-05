package gen2

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ebi-yade/cloud-functions-samples/gen2/app"
	"github.com/ebi-yade/cloud-functions-samples/gen2/app/handlers"
	"github.com/ebi-yade/cloud-functions-samples/gen2/infra/pubsub"
	"github.com/pkg/errors"
)

func init() {
	// https://cloud.google.com/blog/ja/products/application-development/graceful-shutdowns-cloud-run-deep-dive
	const gracePeriod = 5 * time.Second // shorter than Cloud Run's grace period
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	// ==============================================================
	// Initialize structured logging
	// ==============================================================
	logger := slog.New(NewLogHandler(os.Stderr, slog.LevelInfo))
	slog.SetDefault(logger)
	// you may want to get the project ID from the metadata server
	projectID := mustEnv("GOOGLE_CLOUD_PROJECT")
	slog.SetDefault(logger.With(slog.String("project_id", projectID)))

	// ==============================================================
	// Initialize Pub/Sub topic
	// ==============================================================
	topicID := mustEnv("PUBSUB_TOPIC_ID")
	topic, err := pubsub.NewGoogleTopic(ctx, projectID, topicID)
	if err != nil {
		fatal(ctx, errors.Wrap(err, "error pubsub.NewGoogleTopic"))
	}
	slog.InfoContext(ctx, "initialized pubsub topic", slog.String("topic", topicID))

	// ==============================================================
	// Register HTTP / Event-driven handlers
	// ==============================================================
	h := handlers.New(topic)
	app.RegisterHTTP("start", nil, h.Start)

	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), gracePeriod)
		defer cancel()
		slog.InfoContext(ctx, "shutting down...")
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
