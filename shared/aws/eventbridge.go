package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	eventbridgetypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

type EventBridgeAPI interface {
	PutEvents(ctx context.Context, params *EventBridgePutEventsInput) (*EventBridgePutEventsOutput, error)
}

type EventBridgePutEventsInput = eventbridge.PutEventsInput
type EventBridgePutEventsOutput = eventbridge.PutEventsOutput
type PutEventsRequestEntry = eventbridgetypes.PutEventsRequestEntry

type eventBridge struct {
	client *eventbridge.Client
}

func NewEventBridge(client *eventbridge.Client) EventBridgeAPI {
	return &eventBridge{client: client}
}

func (e *eventBridge) PutEvents(ctx context.Context, params *EventBridgePutEventsInput) (*EventBridgePutEventsOutput, error) {
	return e.client.PutEvents(ctx, params)
}
