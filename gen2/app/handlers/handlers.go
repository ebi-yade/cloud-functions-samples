package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Songmu/flextime"
	"github.com/ebi-yade/cloud-functions-samples/gen2/app"
	"github.com/ebi-yade/cloud-functions-samples/gen2/infra/pubsub"
	"github.com/pkg/errors"
)

type Handlers struct {
	topic pubsub.Topic
}

func New(topic pubsub.Topic) *Handlers {
	return &Handlers{
		topic: topic,
	}
}

func (h *Handlers) Start(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return errors.Wrap(err, "error io.ReadAll")
	}
	event := SomeEvent{
		Overview:  fmt.Sprintf("received an HTTP(%s) request", r.Method),
		Payload:   body,
		CreatedAt: flextime.Now(),
	}
	data, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(err, "error json.Marshal")
	}
	if err := h.topic.Publish(ctx, pubsub.Message{Data: data}); err != nil {
		return errors.Wrap(err, "error topic.Publish")
	}

	if err := app.RespondHTTP(ctx, w, nil, http.StatusNoContent); err != nil {
		return errors.Wrap(err, "error app.RespondHTTP")
	}

	return nil
}

type SomeEvent struct {
	Overview  string    `json:"overview"`
	Payload   []byte    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}
