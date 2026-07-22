package main

import (
	"context"
	"testing"
	"time"

	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/notifications"
)

type recordingAcknowledgeIncidentStore struct {
	receivedAt time.Time
}

func (s *recordingAcknowledgeIncidentStore) AcknowledgeIncident(_ context.Context, _, _ string, now time.Time) (dynamodbrecord.IncidentRecord, bool, error) {
	s.receivedAt = now
	return dynamodbrecord.IncidentRecord{IncidentID: "inc-1", Status: incidentStatusAcknowledged}, true, nil
}

func TestAcknowledgeIncidentCommandUsesInjectedClockAndNarrowStore(t *testing.T) {
	now := time.Date(2026, time.July, 22, 10, 0, 0, 0, time.UTC)
	store := &recordingAcknowledgeIncidentStore{}
	incident, found, err := (acknowledgeIncidentCommand{store: store, now: func() time.Time { return now }}).Execute(context.Background(), defaultTenantID, "inc-1")
	if err != nil || !found || incident.Status != incidentStatusAcknowledged {
		t.Fatalf("incident = %+v, found = %v, err = %v", incident, found, err)
	}
	if !store.receivedAt.Equal(now) {
		t.Fatalf("store received time %s, want %s", store.receivedAt, now)
	}
}

type recordingIncidentActivitiesStore struct {
	activitiesListed bool
}

func (s *recordingIncidentActivitiesStore) GetIncident(context.Context, string, string) (dynamodbrecord.IncidentRecord, bool, error) {
	return dynamodbrecord.IncidentRecord{IncidentID: "inc-1"}, true, nil
}

func (s *recordingIncidentActivitiesStore) ListIncidentActivities(context.Context, string, string) ([]dynamodbrecord.IncidentActivityRecord, error) {
	s.activitiesListed = true
	return []dynamodbrecord.IncidentActivityRecord{{ActivityID: "activity-1"}}, nil
}

func TestIncidentActivitiesQueryUsesOnlyItsReadPort(t *testing.T) {
	store := &recordingIncidentActivitiesStore{}
	activities, found, err := (incidentActivitiesQuery{store: store}).Execute(context.Background(), defaultTenantID, "inc-1")
	if err != nil || !found || len(activities) != 1 || !store.activitiesListed {
		t.Fatalf("activities = %+v, found = %v, listed = %v, err = %v", activities, found, store.activitiesListed, err)
	}
}

type recordingIncidentDeliveryStore struct {
	incident   dynamodbrecord.IncidentRecord
	deliveries []notifications.DeliveryRecord
	prepared   notifications.ReplayCommand
}

func (s *recordingIncidentDeliveryStore) GetIncident(context.Context, string, string) (dynamodbrecord.IncidentRecord, bool, error) {
	return s.incident, true, nil
}

func (s *recordingIncidentDeliveryStore) ListIncidentDeliveries(context.Context, string, string) ([]notifications.DeliveryRecord, error) {
	return s.deliveries, nil
}

func (s *recordingIncidentDeliveryStore) LookupReplayIdempotency(context.Context, string, string, string, string) (*notifications.ReplayIdempotencyRecord, error) {
	return nil, nil
}

func (s *recordingIncidentDeliveryStore) PrepareDeliveryReplay(_ context.Context, command notifications.ReplayCommand, _ string, _ time.Time, _ time.Duration) (string, error) {
	s.prepared = command
	return command.DeliveryID, nil
}

func TestReplayIncidentDeliveryCommandOwnsReplayPreparation(t *testing.T) {
	now := time.Date(2026, time.July, 22, 10, 0, 0, 0, time.UTC)
	store := &recordingIncidentDeliveryStore{
		incident: dynamodbrecord.IncidentRecord{IncidentID: "inc-1"},
		deliveries: []notifications.DeliveryRecord{{
			DeliveryID: "delivery-1", TransitionID: "transition-1", State: notifications.DeliveryTerminalFailed,
		}},
	}
	result, err := (replayIncidentDeliveryCommand{incidents: store, store: store, now: func() time.Time { return now }}).Execute(context.Background(), replayIncidentDeliveryInput{
		TenantID: defaultTenantID, IncidentID: "inc-1", DeliveryID: "delivery-1", IdempotencyKey: "retry-1", RequestFingerprint: "fingerprint",
	})
	if err != nil || !result.Queued {
		t.Fatalf("result = %+v, err = %v", result, err)
	}
	if store.prepared.IncidentID != "inc-1" || store.prepared.DeliveryID != "delivery-1" || store.prepared.RequestedAt != now.Format(time.RFC3339) {
		t.Fatalf("prepared command = %+v", store.prepared)
	}
}
