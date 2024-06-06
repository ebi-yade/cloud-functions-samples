package topic

import (
	"context"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

var tracer = otel.Tracer("github.com/ebi-yade/cloud-functions-samples/gen2/infra/topic")

type Message struct {
	Data        []byte
	Attributes  map[string]string
	OrderingKey string
}

func (m Message) toGoogle() *pubsub.Message {
	return &pubsub.Message{
		Data:        m.Data,
		Attributes:  m.Attributes,
		OrderingKey: m.OrderingKey,
	}
}

type Topic interface {
	Publish(ctx context.Context, message Message) error
}

// GoogleTopic is a wrapper for Google Cloud Pub/Sub Topic.
type GoogleTopic struct {
	topic *pubsub.Topic
}

func NewGoogleTopic(topic *pubsub.Topic) *GoogleTopic {
	return &GoogleTopic{
		topic: topic,
	}
}

func (t *GoogleTopic) Publish(ctx context.Context, message Message) error {
	ctx, span := tracer.Start(ctx, "topic.Publish")
	defer span.End()
	sending := message.toGoogle()
	if sending.Attributes == nil {
		sending.Attributes = map[string]string{}
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(sending.Attributes))

	res := t.topic.Publish(ctx, sending)
	messageID, err := res.Get(ctx)
	if err != nil {
		return errors.Wrap(err, "error topic.Publish")
	}

	span.SetAttributes(
		attribute.String("message_id", messageID),
	)

	return nil
}

type SpyTopic struct {
	messages []Message
	mu       sync.Mutex
}

func NewSpyTopic() *SpyTopic {
	return &SpyTopic{}
}

func (t *SpyTopic) Publish(ctx context.Context, message Message) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.messages = append(t.messages, message)

	return nil
}

// ReceivedData is a test helper to get the list of message data sent by SUT.
func (t *SpyTopic) ReceivedData() []Message {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.messages
}
