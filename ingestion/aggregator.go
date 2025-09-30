package ingestion

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"strings"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"go.uber.org/zap"

	"github.com/georgepsarakis/periscope/newcontext"
)

const topicNameEvents = "ingestion.raw_events"

type Aggregator struct {
	logger               *zap.Logger
	fingerprintGenerator FingerprintGenerator
	queue                map[GlobalEventKey][]AggregatedEvent
	lock                 *sync.RWMutex
	pubsub               *pubsub
}

type pubsub struct {
	topic        *gochannel.GoChannel
	subscription <-chan *message.Message
}

type FingerprintGenerator struct {
	hasher     func() hash.Hash
	normalizer func(elements []string) string
}

type GlobalEventKey struct {
	ProjectID uint
	Hash      string
}

func (g GlobalEventKey) String() string {
	return fmt.Sprintf("%d:%s", g.ProjectID, g.Hash)
}

type AggregatedEvent struct {
	AggregationKey GlobalEventKey `json:"aggregation_key"`
	ProjectEvent   ProjectEvent   `json:"project_event"`
}

const fingerprintDelimiter = "/"

func NewAggregator(logger *zap.Logger) *Aggregator {
	topic := gochannel.NewGoChannel(gochannel.Config{}, nil)
	return &Aggregator{
		logger: logger,
		pubsub: &pubsub{
			topic: topic,
		},
		lock: &sync.RWMutex{},
		fingerprintGenerator: FingerprintGenerator{
			hasher: sha1.New,
			normalizer: func(elements []string) string {
				return fingerprintDelimiter + strings.Join(elements, fingerprintDelimiter) + fingerprintDelimiter
			},
		},
		queue: newQueue(),
	}
}

func newQueue() map[GlobalEventKey][]AggregatedEvent {
	return make(map[GlobalEventKey][]AggregatedEvent, 1000)
}

func (a *Aggregator) Subscribe(ctx context.Context) error {
	s, err := a.pubsub.topic.Subscribe(ctx, topicNameEvents)
	if err != nil {
		return err
	}
	a.pubsub.subscription = s
	return nil
}

func (a *Aggregator) Consumer(ctx context.Context) func() error {
	return func() error {
		appLogger := newcontext.LoggerFromContext(ctx)
		appLogger.Info("starting aggregator consumer")
		for {
			select {
			case <-ctx.Done():
				appLogger.Info("aggregator consumer shutdown due to context cancellation", zap.Error(ctx.Err()))
				return nil
			case msg := <-a.pubsub.subscription:
				// we need to Acknowledge that we received and processed the message,
				// otherwise, it will be resent over and over again.
				appLogger.Info(
					fmt.Sprintf("received message: %s, payload: %s\n", msg.UUID, string(msg.Payload)))
				ev := ProjectEventMessage{}
				err := json.Unmarshal(msg.Payload, &ev)
				if err != nil {
					appLogger.Error("failed to unmarshal event", zap.Error(err))
				} else {
					pev, err := a.Extract(ev)
					if err != nil {
						appLogger.Error("failed to process event", zap.Error(err))
					}
					a.Enqueue(AggregatedEvent{
						AggregationKey: GlobalEventKey{ProjectID: pev.ProjectID, Hash: pev.Fingerprint},
						ProjectEvent:   pev,
					})
				}
				msg.Ack()
			}
		}
	}
}

func (a *Aggregator) fingerprint(elements []string) string {
	h := a.fingerprintGenerator.hasher()
	h.Write([]byte(a.fingerprintGenerator.normalizer(elements)))
	return hex.EncodeToString(h.Sum(nil))
}

type ProjectEvent struct {
	ProjectID      uint            `json:"project_id"`
	EventID        string          `json:"event_id"`
	Fingerprint    string          `json:"fingerprint"`
	RawFingerprint json.RawMessage `json:"raw_fingerprint"`
	Trace          json.RawMessage `json:"trace"`
	RawEvent       json.RawMessage `json:"event"`
	Title          string          `json:"title"`
}

func (a *Aggregator) Publish(projectID uint, event Event) error {
	p, err := json.Marshal(ProjectEventMessage{Event: event, ProjectID: projectID})
	if err != nil {
		return err
	}
	msg := message.NewMessage(watermill.NewULID(), p)
	return a.pubsub.topic.Publish(topicNameEvents, msg)
}

func (a *Aggregator) Extract(ev ProjectEventMessage) (ProjectEvent, error) {
	sdkEvent := ev.Event
	fp, err := json.Marshal(sdkEvent.Fingerprint)
	if err != nil {
		return ProjectEvent{}, err
	}
	st, err := json.Marshal(sdkEvent.Exception[0].Stacktrace)
	if err != nil {
		return ProjectEvent{}, err
	}
	re, err := json.Marshal(ev.Event)
	if err != nil {
		return ProjectEvent{}, err
	}
	return ProjectEvent{
		ProjectID:      ev.ProjectID,
		EventID:        ev.Event.EventId,
		RawFingerprint: fp,
		RawEvent:       re,
		Fingerprint:    a.fingerprint(sdkEvent.Fingerprint),
		Trace:          st,
		Title:          sdkEvent.Exception[0].Value,
	}, nil
}

func (a *Aggregator) Enqueue(ae AggregatedEvent) {
	a.lock.Lock()
	defer a.lock.Unlock()

	backlog, exists := a.queue[ae.AggregationKey]
	if !exists {
		a.queue[ae.AggregationKey] = []AggregatedEvent{ae}
	} else {
		a.queue[ae.AggregationKey] = append(backlog, ae)
	}
}

func (a *Aggregator) Flush() [][]AggregatedEvent {
	a.lock.Lock()
	defer a.lock.Unlock()
	batch := make([][]AggregatedEvent, 0, len(a.queue))
	for _, v := range a.queue {
		batch = append(batch, v)
	}
	// TODO: minimize allocations
	a.queue = newQueue()
	return batch
}
