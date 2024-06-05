package gen2

import (
	"context"

	googletrace "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func NewTracerProvider(ctx context.Context, projectID string, defaultSampleRatio float64) (*sdktrace.TracerProvider, error) {
	exporter, err := googletrace.New(googletrace.WithProjectID(projectID))
	if err != nil {
		return nil, errors.Wrap(err, "error NewExporter")
	}

	otelResource, err := resource.New(ctx,
		resource.WithDetectors(gcp.NewDetector()),
		resource.WithTelemetrySDK(),
		//resource.WithAttributes(
		//	semconv.ServiceNameKey.String(funcName),
		//),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error resource.New")
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
