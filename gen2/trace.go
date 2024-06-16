package gen2

import (
	"context"
	"fmt"
	"log/slog"

	googletrace "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	googleoption "google.golang.org/api/option"
)

func NewTracerProvider(ctx context.Context, projectID string, defaultSampleRatio float64) (*sdktrace.TracerProvider, error) {
	exporter, err := googletrace.New(
		googletrace.WithProjectID(projectID),
		googletrace.WithTraceClientOptions([]googleoption.ClientOption{
			googleoption.WithTelemetryDisabled(),
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error NewExporter")
	}

	otelResource, err := resource.New(ctx,
		resource.WithDetectors(gcp.NewDetector()),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error resource.New")
	}

	if defaultSampleRatio >= 0.5 {
		slog.WarnContext(ctx, "defaultSampleRatio is higher than 0.5, it may cause high-rate loops in sending traces")
		if defaultSampleRatio == 1 {
			return nil, fmt.Errorf("defaultSampleRatio must be less than 1 or it causes infinite loops in sending traces")
		}
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(otelResource),
		sdktrace.WithSampler(
			sdktrace.ParentBased(
				sdktrace.TraceIDRatioBased(defaultSampleRatio),
			),
		),
	)

	return tp, nil
}
