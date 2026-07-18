package notifications

import (
	"context"
	"encoding/json"

	"bolt-monitor/shared/outboundhttp"
)

type NotificationSender interface {
	Send(ctx context.Context, notification Notification) error
	ChannelType() string
	ValidateConfig(config json.RawMessage) error
}

type SenderRegistry map[string]NotificationSender

type HTTPExecutor interface {
	Execute(context.Context, outboundhttp.Request) (outboundhttp.Response, error)
}

func (r SenderRegistry) Get(channelType string) (NotificationSender, bool) {
	sender, ok := r[channelType]
	return sender, ok
}

func (r SenderRegistry) Register(channelType string, sender NotificationSender) {
	r[channelType] = sender
}
