package app

import (
	"context"
	"encoding/json"
	"net/http"
)

type HandlerFuncHTTP func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type MiddlewareHTTP func(next HandlerFuncHTTP) HandlerFuncHTTP

// WrapHTTP は標準パッケージの http.HandlerFunc をメチャクチャいい感じにラップします。
func WrapHTTP(mids []MiddlewareHTTP, handlerFunc HandlerFuncHTTP) http.HandlerFunc {
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

// RecoverHTTP はパニックを回復し、アプリケーション全体がクラッシュするのを防ぎます。
func RecoverHTTP(next HandlerFuncHTTP) HandlerFuncHTTP {
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

// RespondHTTP は HTTP レスポンスを返します。
func RespondHTTP(ctx context.Context, w http.ResponseWriter, data any, statusCode int) error {
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
