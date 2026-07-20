package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/notifications"
	"github.com/aws/aws-lambda-go/events"
)

type transitionDispatchRepository interface {
	LoadTransitionOutbox(context.Context, string, string) (*dynamodbrecord.TransitionOutboxRecord, error)
	AcknowledgeDispatch(context.Context, string, string) error
}

type streamDispatcher struct {
	repo     transitionDispatchRepository
	queue    sharedaws.SQSAPI
	queueURL string
}

func newStreamDispatcher(repo transitionDispatchRepository, queue sharedaws.SQSAPI, queueURL string) *streamDispatcher {
	return &streamDispatcher{repo: repo, queue: queue, queueURL: queueURL}
}

func (d *streamDispatcher) handle(ctx context.Context, event events.DynamoDBEvent) (events.DynamoDBEventResponse, error) {
	response := events.DynamoDBEventResponse{}
	for _, record := range event.Records {
		if err := d.dispatchRecord(ctx, record); err != nil {
			if strings.TrimSpace(record.Change.SequenceNumber) == "" {
				return response, err
			}
			response.BatchItemFailures = append(response.BatchItemFailures, events.DynamoDBBatchItemFailure{ItemIdentifier: record.Change.SequenceNumber})
		}
	}
	return response, nil
}

func (d *streamDispatcher) dispatchRecord(ctx context.Context, record events.DynamoDBEventRecord) error {
	if record.EventName != "INSERT" || record.Change.NewImage == nil {
		return nil
	}
	entity := record.Change.NewImage["EntityType"].String()
	if entity != dynamodbrecordEntityTransitionOutbox {
		return nil
	}
	tenantID := record.Change.NewImage["TenantID"].String()
	eventID := record.Change.NewImage["EventID"].String()
	if strings.TrimSpace(eventID) == "" {
		eventID = record.Change.NewImage["TransitionID"].String()
	}
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(eventID) == "" {
		return fmt.Errorf("canonical outbox stream record missing identity")
	}
	outbox, err := d.repo.LoadTransitionOutbox(ctx, tenantID, eventID)
	if err != nil {
		return err
	}
	return d.dispatchOutbox(ctx, outbox)
}

func (d *streamDispatcher) dispatchOutbox(ctx context.Context, outbox *dynamodbrecord.TransitionOutboxRecord) error {
	if outbox == nil || outbox.DispatchStatus != dynamodbrecord.DispatchPending {
		return nil
	}
	kind := outbox.Kind
	if kind == "" {
		kind = notifications.CanonicalKindTransition
	}
	source := outbox.SourceKind
	if source == "" {
		source = notifications.CanonicalSourceTransition
	}
	envelope := notifications.CanonicalEnvelope{
		Version:      notifications.CanonicalEnvelopeVersion,
		Kind:         kind,
		SourceKind:   source,
		TenantID:     outbox.TenantID,
		ServiceID:    outbox.ServiceID,
		MonitorID:    outbox.MonitorID,
		IncidentID:   outbox.IncidentID,
		TransitionID: outbox.EventID,
		RunID:        outbox.RunID,
		CreatedAt:    outbox.CreatedAt,
	}
	if err := envelope.Validate(); err != nil {
		return err
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	if _, err := d.queue.SendMessage(ctx, &sharedaws.SQSSendMessageInput{QueueUrl: sharedaws.String(d.queueURL), MessageBody: sharedaws.String(string(body))}); err != nil {
		return err
	}
	return d.repo.AcknowledgeDispatch(ctx, outbox.TenantID, outbox.EventID)
}

type pendingDispatchRepository interface {
	transitionDispatchRepository
	ListPendingDispatch(context.Context, string, string, int, map[string]sharedaws.AttributeValue) ([]dynamodbrecord.DispatchPendingRecord, map[string]sharedaws.AttributeValue, error)
}

func (d *streamDispatcher) ReconcileRecent(ctx context.Context, repo pendingDispatchRepository, tenantID string, now time.Time, bucketCount, shardCount, pageLimit int) (int, error) {
	if bucketCount <= 0 || shardCount <= 0 || pageLimit <= 0 {
		return 0, fmt.Errorf("reconciliation bounds must be positive")
	}
	count := 0
	for bucketOffset := 0; bucketOffset < bucketCount; bucketOffset++ {
		bucket := now.UTC().Add(-time.Duration(bucketOffset) * time.Hour).Format(dynamodbrecord.DispatchPendingBucketFormat)
		for shard := 0; shard < shardCount; shard++ {
			shardID := fmt.Sprintf("%02x", shard)
			var cursor map[string]sharedaws.AttributeValue
			for {
				records, next, err := repo.ListPendingDispatch(ctx, tenantID, bucket+"|"+shardID, pageLimit, cursor)
				if err != nil {
					return count, err
				}
				for _, pending := range records {
					outbox, err := repo.LoadTransitionOutbox(ctx, pending.TenantID, pending.RunID)
					if err != nil {
						return count, err
					}
					if outbox == nil || outbox.DispatchStatus != dynamodbrecord.DispatchPending {
						continue
					}
					if err := d.dispatchOutbox(ctx, outbox); err != nil {
						return count, err
					}
					count++
				}
				if next == nil {
					break
				}
				cursor = next
			}
		}
	}
	return count, nil
}

func (d *streamDispatcher) RepairEvent(ctx context.Context, repo transitionDispatchRepository, tenantID, eventID string) error {
	outbox, err := repo.LoadTransitionOutbox(ctx, tenantID, eventID)
	if err != nil {
		return err
	}
	if outbox == nil {
		return fmt.Errorf("canonical event %q not found", eventID)
	}
	return d.dispatchOutbox(ctx, outbox)
}

const dynamodbrecordEntityTransitionOutbox = "TransitionOutbox"
