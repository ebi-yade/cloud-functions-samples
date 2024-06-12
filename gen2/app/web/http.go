package web

import (
	"context"
	"encoding/json"
	"net/http"
)

type HandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type Middleware func(next HandlerFunc) HandlerFunc

// BuildStdHttpFunc はいい感じにラップした標準パッケージの http.HandlerFunc を組み立てます。
func BuildStdHttpFunc(mids []Middleware, handlerFunc HandlerFunc) http.HandlerFunc {
	for i := len(mids) - 1; i >= 0; i-- {
		midFunc := mids[i] // loop backwards
		if midFunc != nil {
			handlerFunc = midFunc(handlerFunc)
		}
	}

	stdHandler := http.NewServeMux()
	stdHandler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		v := NewRequestScopeValues()
		ctx = context.WithValue(ctx, ctxKeyRSV, v)

		if err := handlerFunc(ctx, w, r); err != nil {
			status := http.StatusInternalServerError
			http.Error(w, http.StatusText(status), status)
		}
	})

	return stdHandler.ServeHTTP
}

// ===============================================================
// HTTP Middlewares
// ===============================================================

// Recover はパニックを回復し、アプリケーション全体がクラッシュするのを防ぎます。
func Recover(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) (returningError error) {
		defer func() {
			if r := recover(); r != nil {
				returningError = r.(error)
			}
		}()

		return next(ctx, w, r)
	}
}

// ===============================================================
// HTTP Utilities
// ===============================================================

// Respond は HTTP レスポンスを返します。
func Respond(ctx context.Context, w http.ResponseWriter, data any, statusCode int) error {
	setStatusCode(ctx, statusCode)

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := encoder.Encode(data); err != nil {
		return err
	}

	return nil
}
