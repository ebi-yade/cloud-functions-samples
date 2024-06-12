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

func InjectOtel(ctx context.Context, message *pubsub.Message) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(message.Attributes))
}

func ExtractOtel(ctx context.Context, message *pubsub.Message) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(message.Attributes))
}

type Topic interface {
	Publish(ctx context.Context, message Message) error
}

// GoogleTopic is a wrapper for Google Cloud Pub/Sub Topic.
type GoogleTopic struct {
	client  *pubsub.Client
	topicID string
}

func NewGoogleTopic(ctx context.Context, projectID, topicID string) (*GoogleTopic, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, errors.Wrap(err, "error pubsub.NewClient")
	}

	return &GoogleTopic{
		client:  client,
		topicID: topicID,
	}, nil
}

func (t *GoogleTopic) Close(_ context.Context) error {
	return t.client.Close()
}

func (t *GoogleTopic) Publish(ctx context.Context, message Message) error {
	ctx, span := tracer.Start(ctx, "topic.Publish")
	defer span.End()
	sending := message.toGoogle()
	if sending.Attributes == nil {
		sending.Attributes = map[string]string{}
	}
	InjectOtel(ctx, sending)

	res := t.client.Topic(t.topicID).Publish(ctx, sending)
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
