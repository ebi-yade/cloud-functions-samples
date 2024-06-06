package app

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/cloudevents/sdk-go/v2/event"
)

type HandlerFuncEvent func(ctx context.Context, e event.Event) error

type MiddlewareEvent func(next HandlerFuncEvent) HandlerFuncEvent

// WrapEvent はイベントドリブン関数用のミドルウェアたちを受け取り、それらを後ろから順に適用する形でハンドラをラップします。
func WrapEvent(mids []MiddlewareEvent, handler HandlerFuncEvent) HandlerFuncEvent {
	for i := len(mids) - 1; i >= 0; i-- {
		mwFunc := mids[i] // loop backwards
		if mwFunc != nil {
			handler = mwFunc(handler)
		}
	}

	return handler
}

// ===============================================================
// Event Middlewares
// ===============================================================

// LogErrorEvent はエラーをログに出力します。
// Cloud Functions はエラーを返すとログが出力されますが、構造化ログで ErrorReporting に送るためにこちらでもログを出力します。
func LogErrorEvent(next HandlerFuncEvent) HandlerFuncEvent {
	return func(ctx context.Context, e event.Event) error {
		err := next(ctx, e)
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("%+v", err))
		}

		return err
	}
}

// RecoverEvent はパニックを回復し、アプリケーション全体がクラッシュするのを防ぎます。
func RecoverEvent(next HandlerFuncEvent) HandlerFuncEvent {
	return func(ctx context.Context, e event.Event) (returningError error) {
		defer func() {
			if r := recover(); r != nil {
				returningError = fmt.Errorf("panic: %v\n\n%s", r, string(debug.Stack()))
				slog.ErrorContext(ctx, returningError.Error())
			}
		}()

		return next(ctx, e)
	}
}
