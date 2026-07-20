package main

import (
	"context"
	"testing"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbrecord"
	"github.com/aws/aws-lambda-go/events"
)

type reconcilingRepo struct {
	fakeTransitionDispatchRepo
	pending []dynamodbrecord.DispatchPendingRecord
}

func (f *reconcilingRepo) ListPendingDispatch(_ context.Context, tenantID, bucketShard string, _ int, _ map[string]sharedaws.AttributeValue) ([]dynamodbrecord.DispatchPendingRecord, map[string]sharedaws.AttributeValue, error) {
	return f.pending, nil, nil
}

func TestReconcileRecentBoundedRejectsZeroBounds(t *testing.T) {
	repo := &fakeTransitionDispatchRepo{outbox: &dynamodbrecord.TransitionOutboxRecord{DispatchStatus: dynamodbrecord.DispatchPending}}
	queue := &fakeDispatchSQS{}
	dispatcher := newStreamDispatcher(repo, queue, "queue-url")
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	if _, err := dispatcher.ReconcileRecent(context.Background(), &reconcilingRepo{}, "DEFAULT", now, 0, 1, 1); err == nil {
		t.Fatalf("zero bucket count must be rejected")
	}
	if _, err := dispatcher.ReconcileRecent(context.Background(), &reconcilingRepo{}, "DEFAULT", now, 1, 0, 1); err == nil {
		t.Fatalf("zero shard count must be rejected")
	}
	if _, err := dispatcher.ReconcileRecent(context.Background(), &reconcilingRepo{}, "DEFAULT", now, 1, 1, 0); err == nil {
		t.Fatalf("zero page limit must be rejected")
	}
}

func TestReconcileRecentDispatchesPendingOutbox(t *testing.T) {
	outbox := &dynamodbrecord.TransitionOutboxRecord{TenantID: "DEFAULT", EventID: "TRN_1", TransitionID: "TRN_1", DispatchStatus: dynamodbrecord.DispatchPending, IncidentID: "INC_1", CreatedAt: "2026-07-19T12:00:00Z"}
	repo := &fakeTransitionDispatchRepo{outbox: outbox}
	queue := &fakeDispatchSQS{}
	dispatcher := newStreamDispatcher(repo, queue, "queue-url")
	rec := &reconcilingRepo{fakeTransitionDispatchRepo: *repo, pending: []dynamodbrecord.DispatchPendingRecord{{TenantID: "DEFAULT", RunID: "TRN_1", Bucket: "2026071912", Shard: "00"}}}
	count, err := dispatcher.ReconcileRecent(context.Background(), rec, "DEFAULT", time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC), 1, 1, 10)
	if err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}
	if count != 1 || len(queue.bodies) != 1 {
		t.Fatalf("count=%d bodies=%d", count, len(queue.bodies))
	}
}

func TestStreamDispatcherAndReconcilerAvoidDuplicateDispatch(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	pending := &dynamodbrecord.TransitionOutboxRecord{TenantID: "DEFAULT", EventID: "TRN_RACE", TransitionID: "TRN_RACE", DispatchStatus: dynamodbrecord.DispatchPending, IncidentID: "INC_RACE", CreatedAt: now.Format(time.RFC3339)}
	streamRepo := &fakeTransitionDispatchRepo{outbox: pending}
	streamQueue := &fakeDispatchSQS{}
	streamDispatcher := newStreamDispatcher(streamRepo, streamQueue, "queue-url")
	event := events.DynamoDBEvent{Records: []events.DynamoDBEventRecord{{
		EventName: "INSERT", Change: events.DynamoDBStreamRecord{SequenceNumber: "race-seq", NewImage: map[string]events.DynamoDBAttributeValue{
			"EntityType": events.NewStringAttribute("TransitionOutbox"), "TenantID": events.NewStringAttribute("DEFAULT"), "EventID": events.NewStringAttribute("TRN_RACE"),
		}},
	}}}
	if _, err := streamDispatcher.handle(context.Background(), event); err != nil {
		t.Fatalf("stream dispatch failed: %v", err)
	}
	acked := &dynamodbrecord.TransitionOutboxRecord{TenantID: "DEFAULT", EventID: "TRN_RACE", TransitionID: "TRN_RACE", DispatchStatus: "acknowledged"}
	recRepo := &fakeTransitionDispatchRepo{outbox: acked}
	recDispatcher := newStreamDispatcher(recRepo, &fakeDispatchSQS{}, "queue-url")
	rec := &reconcilingRepo{fakeTransitionDispatchRepo: *recRepo, pending: []dynamodbrecord.DispatchPendingRecord{{TenantID: "DEFAULT", RunID: "TRN_RACE", Bucket: "2026071912", Shard: "00"}}}
	count, err := recDispatcher.ReconcileRecent(context.Background(), rec, "DEFAULT", now, 1, 1, 10)
	if err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("reconciler must skip acked outbox, got count=%d", count)
	}
}

func TestRepairEventRequiresExistingOutbox(t *testing.T) {
	repo := &fakeTransitionDispatchRepo{outbox: nil}
	dispatcher := newStreamDispatcher(repo, &fakeDispatchSQS{}, "queue-url")
	if err := dispatcher.RepairEvent(context.Background(), repo, "DEFAULT", "TRN_MISSING"); err == nil {
		t.Fatalf("missing outbox should fail repair")
	}
}
