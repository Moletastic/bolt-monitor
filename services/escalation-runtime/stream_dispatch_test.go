package main

import (
	"context"
	"testing"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbrecord"
	"github.com/aws/aws-lambda-go/events"
)

type fakeTransitionDispatchRepo struct {
	outbox  *dynamodbrecord.TransitionOutboxRecord
	acked   []string
	loadErr error
}

func (f *fakeTransitionDispatchRepo) LoadTransitionOutbox(context.Context, string, string) (*dynamodbrecord.TransitionOutboxRecord, error) {
	return f.outbox, f.loadErr
}

func (f *fakeTransitionDispatchRepo) AcknowledgeDispatch(_ context.Context, tenantID, eventID string) error {
	f.acked = append(f.acked, tenantID+"/"+eventID)
	return nil
}

type fakeDispatchSQS struct {
	bodies []string
}

func (f *fakeDispatchSQS) SendMessage(_ context.Context, input *sharedaws.SQSSendMessageInput) (*sharedaws.SQSSendMessageOutput, error) {
	f.bodies = append(f.bodies, sharedaws.ToString(input.MessageBody))
	return &sharedaws.SQSSendMessageOutput{}, nil
}
func (f *fakeDispatchSQS) ReceiveMessage(context.Context, *sharedaws.SQSReceiveMessageInput) (*sharedaws.SQSReceiveMessageOutput, error) {
	return nil, nil
}
func (f *fakeDispatchSQS) DeleteMessage(context.Context, *sharedaws.SQSDeleteMessageInput) (*sharedaws.SQSDeleteMessageOutput, error) {
	return nil, nil
}
func (f *fakeDispatchSQS) ChangeMessageVisibility(context.Context, *sharedaws.SQSChangeMessageVisibilityInput) (*sharedaws.SQSChangeMessageVisibilityOutput, error) {
	return nil, nil
}

func TestStreamDispatcherDispatchesOnlyCanonicalInserts(t *testing.T) {
	repo := &fakeTransitionDispatchRepo{outbox: &dynamodbrecord.TransitionOutboxRecord{
		TenantID: "DEFAULT", EventID: "TRN_1", TransitionID: "TRN_1", DispatchStatus: dynamodbrecord.DispatchPending,
		ServiceID: "SVC_1", MonitorID: "MON_1", IncidentID: "INC_1", RunID: "RUN_1", CreatedAt: "2026-01-01T00:00:00Z",
	}}
	queue := &fakeDispatchSQS{}
	dispatcher := newStreamDispatcher(repo, queue, "queue-url")
	event := events.DynamoDBEvent{Records: []events.DynamoDBEventRecord{
		{EventName: "MODIFY", Change: events.DynamoDBStreamRecord{SequenceNumber: "1", NewImage: map[string]events.DynamoDBAttributeValue{"EntityType": events.NewStringAttribute("TransitionOutbox")}}},
		{EventName: "INSERT", Change: events.DynamoDBStreamRecord{SequenceNumber: "2", NewImage: map[string]events.DynamoDBAttributeValue{
			"EntityType": events.NewStringAttribute("TransitionOutbox"), "TenantID": events.NewStringAttribute("DEFAULT"), "EventID": events.NewStringAttribute("TRN_1"),
		}}},
	}}
	response, err := dispatcher.handle(context.Background(), event)
	if err != nil {
		t.Fatalf("handle returned error: %v", err)
	}
	if len(response.BatchItemFailures) != 0 || len(queue.bodies) != 1 || len(repo.acked) != 1 {
		t.Fatalf("response=%+v bodies=%d acked=%d", response, len(queue.bodies), len(repo.acked))
	}
}

func TestStreamDispatcherReturnsSequenceNumberForRetry(t *testing.T) {
	repo := &fakeTransitionDispatchRepo{loadErr: sharedaws.NewConditionalCheckFailedError()}
	dispatcher := newStreamDispatcher(repo, &fakeDispatchSQS{}, "queue-url")
	response, err := dispatcher.handle(context.Background(), events.DynamoDBEvent{Records: []events.DynamoDBEventRecord{{
		EventName: "INSERT", EventID: "event-id", Change: events.DynamoDBStreamRecord{SequenceNumber: "sequence-42", NewImage: map[string]events.DynamoDBAttributeValue{
			"EntityType": events.NewStringAttribute("TransitionOutbox"), "TenantID": events.NewStringAttribute("DEFAULT"), "EventID": events.NewStringAttribute("TRN_1"),
		}},
	}}})
	if err != nil {
		t.Fatalf("handle returned error: %v", err)
	}
	if len(response.BatchItemFailures) != 1 || response.BatchItemFailures[0].ItemIdentifier != "sequence-42" {
		t.Fatalf("failures=%+v", response.BatchItemFailures)
	}
}
