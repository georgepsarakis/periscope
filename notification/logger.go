package notification

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"
)

type LogNotifier struct {
	Channel
	Logger *zap.Logger
}

func (l LogNotifier) Serialize(event Event) ([]byte, error) {
	return json.Marshal(event)
}

func (l LogNotifier) Emit(_ context.Context, event Event) error {
	l.Logger.Error("notification", zap.Any("event", event))
	return nil
}

func (l LogNotifier) OnSuccess(func(ctx context.Context, event Event) error) {}
