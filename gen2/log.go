package gen2

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/logging"
	"github.com/Songmu/flextime" // time compatible package with utilities for testing
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

var rootDir = sync.OnceValue(func() string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Dir(currentFile)
})

var (
	logAttrReporting = slog.String(
		"@type",
		"type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent",
	)
)

const (
	logMessageKey        = "message"
	logSeverityKey       = "severity"
	logSourceLocationKey = "logging.googleapis.com/sourceLocation"
	logTraceKey          = "logging.googleapis.com/trace"
	logSpanIDKey         = "logging.googleapis.com/spanId"
	logTraceSampledKey   = "logging.googleapis.com/traceSampled"
	logInsertIDKey       = "insertId"
	logTimestampKey      = "timestamp"
)

// LogHandler は構造化ログ世界のバックエンドとしての slog.LogHandler をラップします
type LogHandler struct {
	base      slog.Handler
	projectID string
}

func NewLogHandler(w io.Writer, minLevel slog.Level) *LogHandler {
	replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey { // 絶対パスを相対パスに変換
			if source, ok := a.Value.Any().(*slog.Source); ok {
				return slog.Any(logSourceLocationKey, &slog.Source{
					Function: source.Function,
					File:     "{root}" + strings.TrimPrefix(source.File, rootDir()),
					Line:     source.Line,
				})
			}
		}
		if a.Key == slog.MessageKey {
			a.Key = logMessageKey
			return a
		}
		if a.Key == slog.LevelKey {
			return slog.String(logSeverityKey, logging.Severity(a.Value.Any().(slog.Level)).String())
		}

		return a
	}
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource:   true,
		Level:       minLevel,
		ReplaceAttr: replaceAttr,
	})
	return &LogHandler{base: handler}
}

func (h *LogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

func (h *LogHandler) Handle(ctx context.Context, record slog.Record) error {
	// 一意なエントリを保証するための情報を生成
	now := flextime.Now()
	insertId := uuid.NewString()

	// 属性を追加
	attrs := []slog.Attr{
		slog.String(logInsertIDKey, insertId),
		slog.String(logTimestampKey, now.Format(time.RFC3339Nano)),
	}

	// Open Telemetry トレース情報を追加
	if span := trace.SpanFromContext(ctx); span != nil {
		sc := span.SpanContext()
		if h.projectID != "" {
			traceID := "projects/" + h.projectID + "/traces/" + sc.TraceID().String()
			attrs = append(attrs, slog.String(logTraceKey, traceID))
		}
		attrs = append(attrs, slog.String(logSpanIDKey, sc.SpanID().String()))
		attrs = append(attrs, slog.Bool(logTraceSampledKey, sc.IsSampled()))
	}

	if record.Level == slog.LevelError {
		attrs = append(attrs, logAttrReporting)
	}
	record.AddAttrs(attrs...)

	// timestamp, insertId によって重複対策をしているので、リトライを書いても良い?
	return h.base.Handle(ctx, record)
}

const (
	AttrKeyProjectID = "project_id"
)

func (h *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var unknownAttrs []slog.Attr
	for _, a := range attrs {
		switch a.Key {
		case AttrKeyProjectID:
			h.projectID = a.Value.String()
		default:
			unknownAttrs = append(unknownAttrs, a)
		}
	}

	var base slog.Handler
	if len(unknownAttrs) > 0 {
		base = h.base.WithAttrs(unknownAttrs)
	} else {
		base = h.base
	}
	return &LogHandler{base: base, projectID: h.projectID}
}

func (h *LogHandler) WithGroup(name string) slog.Handler {
	base := h.base.WithGroup(name)
	return &LogHandler{base: base, projectID: h.projectID}
}
