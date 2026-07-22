package main

import (
	"context"
	"testing"
	"time"

	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/escalation"
)

type recordingSchedulerConfigStore struct {
	updated checkexecution.SchedulerConfig
	now     time.Time
}

func (s *recordingSchedulerConfigStore) GetSchedulerConfig(context.Context, string) (dynamodbrecord.SchedulerConfigRecord, error) {
	return dynamodbrecord.SchedulerConfigRecord{}, nil
}
func (s *recordingSchedulerConfigStore) UpdateSchedulerConfig(_ context.Context, _ string, config checkexecution.SchedulerConfig, now time.Time) (dynamodbrecord.SchedulerConfigRecord, error) {
	s.updated, s.now = config, now
	return dynamodbrecord.SchedulerConfigRecord{Config: config}, nil
}

func TestUpdateSchedulerConfigCommandValidatesAndUsesInjectedClock(t *testing.T) {
	store := &recordingSchedulerConfigStore{}
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	config := checkexecution.SchedulerConfig{RecurringEnabled: true, StopControlMode: checkexecution.StopControlMonitorDisable}
	got, err := newSchedulerOperations(store, store, func() time.Time { return now }).update.Execute(context.Background(), defaultTenantID, config)
	if err != nil || got.Config != config || store.updated != config || !store.now.Equal(now) {
		t.Fatalf("updated = %+v, stored = %+v at %v, err = %v", got, store.updated, store.now, err)
	}
}

type recordingPolicyStore struct {
	created escalation.EscalationPolicy
	channel *escalation.NotificationChannel
}

func (s *recordingPolicyStore) CreateEscalationPolicy(_ context.Context, policy escalation.EscalationPolicy) (escalation.EscalationPolicy, error) {
	s.created = policy
	return policy, nil
}
func (s *recordingPolicyStore) GetNotificationChannel(context.Context, string, string) (*escalation.NotificationChannel, error) {
	return s.channel, nil
}

func TestCreateEscalationPolicyCommandOwnsIDAndChannelValidation(t *testing.T) {
	store := &recordingPolicyStore{channel: &escalation.NotificationChannel{ChannelID: "channel-1"}}
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	policy := escalation.EscalationPolicy{TenantID: defaultTenantID, Name: "Primary", BusinessHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{{ChannelID: "channel-1"}}}, OffHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{{ChannelID: "channel-1"}}}}
	operations := newEscalationPolicyOperations(store, nil, nil, nil, nil, store, nil, nil, func() time.Time { return now }, identifierGenerator{newEscalationPolicyID: func(time.Time) string { return "POL_TEST" }})
	got, err := operations.create.Execute(context.Background(), policy)
	if err != nil || got.PolicyID == "" || store.created.PolicyID != got.PolicyID {
		t.Fatalf("created = %+v, got = %+v, err = %v", store.created, got, err)
	}
}

type recordingChannelDeleteStore struct {
	channel    *escalation.NotificationChannel
	references []routeReference
	deleted    bool
}

func (s *recordingChannelDeleteStore) ListNotificationChannels(context.Context, string) ([]escalation.NotificationChannel, error) {
	return nil, nil
}
func (s *recordingChannelDeleteStore) GetNotificationChannel(context.Context, string, string) (*escalation.NotificationChannel, error) {
	return s.channel, nil
}
func (s *recordingChannelDeleteStore) ChannelsReferencedByRoutes(context.Context, string, string) ([]routeReference, error) {
	return s.references, nil
}
func (s *recordingChannelDeleteStore) DeleteNotificationChannel(context.Context, string, string) error {
	s.deleted = true
	return nil
}

func TestDeleteNotificationChannelCommandDoesNotDeleteReferencedChannel(t *testing.T) {
	store := &recordingChannelDeleteStore{channel: &escalation.NotificationChannel{ChannelID: "channel-1"}, references: []routeReference{{PolicyID: "policy-1"}}}
	references, err := (deleteNotificationChannelCommand{channels: store, store: store}).Execute(context.Background(), defaultTenantID, "channel-1")
	if err != nil || len(references) != 1 || store.deleted {
		t.Fatalf("references = %+v, deleted = %v, err = %v", references, store.deleted, err)
	}
}

type recordingSearchResourcesStore struct{ called bool }

func (s *recordingSearchResourcesStore) SearchResources(_ context.Context, tenantID, query string, limit int, types map[string]struct{}) ([]searchResult, error) {
	s.called = tenantID == defaultTenantID && query == "payments" && limit == 8 && len(types) == 1
	return []searchResult{{ID: "service-1"}}, nil
}

func TestSearchResourcesQueryUsesOnlySearchPort(t *testing.T) {
	store := &recordingSearchResourcesStore{}
	results, err := (searchResourcesQuery{store: store}).Execute(context.Background(), defaultTenantID, "payments", 8, map[string]struct{}{searchResourceService: {}})
	if err != nil || !store.called || len(results) != 1 {
		t.Fatalf("results = %+v, called = %v, err = %v", results, store.called, err)
	}
}
