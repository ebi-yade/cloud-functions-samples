package app

import (
	"context"
	"net/http"
)

type HTTPHandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type MiddlewareForHTTP func(next HTTPHandlerFunc) HTTPHandlerFunc
