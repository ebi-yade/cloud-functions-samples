package app

import (
	"context"
	"time"

	"github.com/Songmu/flextime"
)

type ctxKey int

const (
	ctxKeyRSV ctxKey = iota
)

type RequestScopeValues struct {
	StartTime time.Time

	// mutable values accessed via getter
	statusCode *int
}

func newPointer[T any](v T) *T {
	return &v
}

const (
	statusCodeUnknown = 0
)

func NewRequestScopeValues() RequestScopeValues {
	return RequestScopeValues{
		StartTime:  flextime.Now(),
		statusCode: newPointer(statusCodeUnknown),
	}
}

func GetValues(ctx context.Context) RequestScopeValues {
	v, ok := ctx.Value(ctxKeyRSV).(RequestScopeValues)
	if !ok {
		return NewRequestScopeValues()
	}

	return v
}

func GetStatusCode(ctx context.Context) int {
	v, ok := ctx.Value(ctxKeyRSV).(RequestScopeValues)
	if !ok || v.statusCode == nil {
		return statusCodeUnknown
	}

	return *v.statusCode
}

func setStatusCode(ctx context.Context, statusCode int) {
	v := GetValues(ctx)
	v.statusCode = &statusCode
}
