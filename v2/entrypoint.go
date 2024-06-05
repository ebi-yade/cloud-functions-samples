package v2

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	// https://cloud.google.com/blog/ja/products/application-development/graceful-shutdowns-cloud-run-deep-dive
	const gracePeriod = 5 * time.Second // shorter than Cloud Run's grace period
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	appLogger := slog.New(NewLogHandler(os.Stderr, slog.LevelInfo))
	slog.SetDefault(appLogger)
	// you may want to get the project ID from the metadata server
	projectID := mustEnv("GOOGLE_CLOUD_PROJECT")
	slog.SetDefault(appLogger.With(slog.String("project_id", projectID)))

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
