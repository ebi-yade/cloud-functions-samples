package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/Songmu/flextime"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/ebi-yade/cloud-functions-samples/gen2/app"
	"github.com/ebi-yade/cloud-functions-samples/gen2/infra/topic"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

type Handlers struct {
	topic topic.Topic
}

func New(topic topic.Topic) *Handlers {
	return &Handlers{
		topic: topic,
	}
}

func (h *Handlers) Start(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return errors.Wrap(err, "error io.ReadAll")
	}
	e := SomeEvent{
		Overview:  fmt.Sprintf("received an HTTP(%s) request", r.Method),
		Payload:   body,
		CreatedAt: flextime.Now(),
	}
	data, err := json.Marshal(e)
	if err != nil {
		return errors.Wrap(err, "error json.Marshal")
	}
	if err := h.topic.Publish(ctx, topic.Message{Data: data}); err != nil {
		return errors.Wrap(err, "error topic.Publish")
	}

	if err := app.RespondHTTP(ctx, w, nil, http.StatusNoContent); err != nil {
		return errors.Wrap(err, "error app.RespondHTTP")
	}

	return nil
}

type SomeEvent struct {
	Overview  string    `json:"overview" validate:"required"`
	Payload   []byte    `json:"payload" validate:"required"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
}

func (h *Handlers) Hook(ctx context.Context, eventContext event.Event) error {
	var e SomeEvent
	if err := eventContext.DataAs(&e); err != nil {
		return errors.Wrap(err, "error eventContext.DataAs")
	}

	if err := validator.New().Struct(e); err != nil {
		return errors.Wrap(err, "error validator.New().Struct")
	}

	slog.InfoContext(ctx, fmt.Sprintf("received an event"), slog.Any("event", e))

	return nil
}
