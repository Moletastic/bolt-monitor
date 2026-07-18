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

// NewSenderRegistry keeps every notification path on the same outbound executor policy.
func NewSenderRegistry(executors ...HTTPExecutor) SenderRegistry {
	return SenderRegistry{
		"telegram":  NewTelegramSender(executors...),
		"email":     NewEmailSender(executors...),
		"sms":       NewSMSSender(executors...),
		"webhook":   NewWebhookSender(executors...),
		"pagerduty": NewPagerDutySender(executors...),
	}
}
