package notifications

import (
	"context"
	"encoding/json"
)

type NotificationSender interface {
	Send(ctx context.Context, notification Notification) error
	ChannelType() string
	ValidateConfig(config json.RawMessage) error
}

type SenderRegistry map[string]NotificationSender

func (r SenderRegistry) Get(channelType string) (NotificationSender, bool) {
	sender, ok := r[channelType]
	return sender, ok
}

func (r SenderRegistry) Register(channelType string, sender NotificationSender) {
	r[channelType] = sender
}
