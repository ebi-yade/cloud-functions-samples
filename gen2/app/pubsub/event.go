package pubsub

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"

	"cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/pkg/errors"
)

type MessageEvent pubsub.Message

type HandlerFunc func(ctx context.Context, e MessageEvent) error

type EventDrivenFunc func(ctx context.Context, e event.Event) error

// BuildEventDrivenFunc は、Functions Framework に登録可能なイベントドリブン関数をいい感じに組み立てます。
func BuildEventDrivenFunc(handlerFunc HandlerFunc) EventDrivenFunc {

	fn := func(ctx context.Context, e event.Event) (returningError error) {
		defer func() {
			if r := recover(); r != nil {
				returningError = fmt.Errorf("panic: %v\n\n%s", r, string(debug.Stack()))
			}
		}()

		msg := pubsub.Message{}
		if err := e.DataAs(&msg); err != nil {
			return errors.Wrap(err, "error e.DataAs")
		}

		return handlerFunc(ctx, MessageEvent(msg))
	}

	return wrapLoggingErr(fn)
}

func wrapLoggingErr(next EventDrivenFunc) EventDrivenFunc {
	return func(ctx context.Context, e event.Event) error {
		if err := next(ctx, e); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("%+v", err))
			return err
		}
		return nil
	}
}
