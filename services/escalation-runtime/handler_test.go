package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/notifications"
	"bolt-monitor/shared/outboundhttp"
	"github.com/aws/aws-lambda-go/events"
)

type fakeEscalationRepository struct {
	service         *serviceRecord
	policy          *escalation.EscalationPolicy
	channels        map[string]escalation.NotificationChannel
	state           *escalation.EscalationState
	incident        *incidentRecord
	createdIncident *incidentRecord
	outbox          map[string]dynamodbrecord.TransitionOutboxRecord
}

type fakeScheduler struct {
	event  scheduledInvocationEvent
	when   time.Time
	called bool
}

func (f *fakeScheduler) ScheduleNextStep(_ context.Context, event scheduledInvocationEvent, when time.Time) error {
	f.event = event
	f.when = when
	f.called = true
	return nil
}

type fakeSender struct {
	notifications []notifications.Notification
}

var testEscalationNow = time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)

func newTestEscalationHandler(repo escalationRepository, scheduler scheduleClient) *escalationHandler {
	return newTestEscalationHandlerWithDependencies(repo, scheduler, notifications.SenderRegistry{}, testEscalationNow)
}

func newTestEscalationHandlerWithDependencies(repo escalationRepository, scheduler scheduleClient, senders notifications.SenderRegistry, now time.Time) *escalationHandler {
	return newEscalationHandlerWithDependencies(repo, scheduler, escalationHandlerDependencies{
		senders: senders,
		now:     func() time.Time { return now },
	})
}

type capturingEscalationExecutor struct {
	request outboundhttp.Request
}

func (e *capturingEscalationExecutor) Execute(_ context.Context, request outboundhttp.Request) (outboundhttp.Response, error) {
	e.request = request
	return outboundhttp.Response{StatusCode: http.StatusOK}, nil
}

func (f *fakeSender) Send(_ context.Context, notification notifications.Notification) (notifications.SendOutcome, error) {
	f.notifications = append(f.notifications, notification)
	return notifications.SendOutcome{Class: notifications.OutcomeAccepted}, nil
}

func (f *fakeSender) ChannelType() string { return "fake" }

func (f *fakeSender) ValidateConfig(config json.RawMessage) error { return nil }

func (f *fakeEscalationRepository) GetService(context.Context, string, string) (*serviceRecord, error) {
	return f.service, nil
}

func (f *fakeEscalationRepository) GetEscalationPolicy(context.Context, string, string) (*escalation.EscalationPolicy, error) {
	return f.policy, nil
}

func (f *fakeEscalationRepository) GetChannel(_ context.Context, tenantID, channelID string) (*escalation.NotificationChannel, error) {
	channel, ok := f.channels[channelID]
	if !ok || channel.TenantID != tenantID {
		return nil, nil
	}
	return &channel, nil
}

func (f *fakeEscalationRepository) LoadTransitionOutbox(_ context.Context, tenantID, eventID string) (*dynamodbrecord.TransitionOutboxRecord, error) {
	if f.outbox == nil {
		return nil, nil
	}
	record, ok := f.outbox[tenantID+"|"+eventID]
	if !ok {
		return nil, nil
	}
	return &record, nil
}

func (f *fakeEscalationRepository) AcknowledgeDispatch(_ context.Context, tenantID, eventID string) error {
	if f.outbox == nil {
		return nil
	}
	record, ok := f.outbox[tenantID+"|"+eventID]
	if !ok || record.DispatchStatus != "pending" {
		return nil
	}
	record.DispatchStatus = "acknowledged"
	f.outbox[tenantID+"|"+eventID] = record
	return nil
}

func (f *fakeEscalationRepository) PutEscalationState(_ context.Context, state escalation.EscalationState) error {
	f.state = &state
	return nil
}

func (f *fakeEscalationRepository) GetEscalationState(context.Context, string, string) (*escalation.EscalationState, error) {
	return f.state, nil
}

func (f *fakeEscalationRepository) GetIncident(context.Context, string) (*incidentRecord, error) {
	return f.incident, nil
}

func (f *fakeEscalationRepository) CreateIncident(_ context.Context, incident incidentRecord) error {
	f.createdIncident = &incident
	return nil
}

func (f *fakeEscalationRepository) CreateEscalationPlan(context.Context, notifications.EscalationPlan) error {
	return nil
}

func (f *fakeEscalationRepository) GetEscalationPlan(context.Context, string, string, string) (*notifications.EscalationPlan, error) {
	return nil, nil
}

func (f *fakeEscalationRepository) CreateDelivery(context.Context, notifications.DeliveryRecord) error {
	return nil
}

func (f *fakeEscalationRepository) ListIncidentDeliveries(context.Context, string, string) ([]notifications.DeliveryRecord, error) {
	return nil, nil
}

func (f *fakeEscalationRepository) AdvanceStepOnce(context.Context, string, string, int, int, string) error {
	return nil
}

func (f *fakeEscalationRepository) SuppressEscalation(context.Context, string, string, string) error {
	return nil
}

func TestHandleIncidentDownCreatesEscalationState(t *testing.T) {
	tgSender := &fakeSender{}
	emailSender := &fakeSender{}
	repo := &fakeEscalationRepository{
		service:  &serviceRecord{TenantID: "DEFAULT", ServiceID: "auth", EscalationPolicyID: "POL_1", BusinessHours: &escalation.BusinessHoursConfig{Timezone: "UTC", StartHour: 9, EndHour: 17, DaysOfWeek: []int{2}}},
		policy:   &escalation.EscalationPolicy{TenantID: "DEFAULT", PolicyID: "POL_1", Name: "Primary", BusinessHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{{DelayMinutes: 0, Channels: []escalation.ChannelConfig{{Type: escalation.ChannelTypeTelegram, Target: "chat-1", Config: json.RawMessage(`{"botToken":"token"}`)}, {Type: escalation.ChannelTypeEmail, Target: "ops@example.com", Config: json.RawMessage(`{"apiKey":"key","fromEmail":"from@example.com"}`)}}}}}},
		incident: &incidentRecord{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "public-http", IncidentID: "INC_1", Status: incidentStatusOpen, OpenedAt: "2026-06-16T10:00:00Z", UpdatedAt: "2026-06-16T10:00:00Z"},
	}
	scheduler := &fakeScheduler{}
	handler := newTestEscalationHandlerWithDependencies(repo, scheduler, notifications.SenderRegistry{"telegram": tgSender, "email": emailSender}, testEscalationNow)
	jsonEvent, _ := notifications.NotificationEvent{EventType: notifications.EventTypeIncidentDown, TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "public-http", IncidentID: "INC_1", Timestamp: time.Date(2026, 6, 16, 10, 0, 0, 0, time.UTC)}.ToJSON()
	err := handler.handleSQSEvent(context.Background(), events.SQSEvent{Records: []events.SQSMessage{{Body: jsonEvent}}})
	if err != nil {
		t.Fatalf("handleSQSEvent returned error: %v", err)
	}
	if repo.state == nil {
		t.Fatal("expected escalation state to be created")
	}
	if repo.state.PolicyID != "POL_1" {
		t.Fatalf("policyID = %q, want POL_1", repo.state.PolicyID)
	}
	if repo.state.SelectedPath != pathBusinessHours {
		t.Fatalf("selectedPath = %q, want %q", repo.state.SelectedPath, pathBusinessHours)
	}
	if repo.state.CurrentStep != 2 {
		t.Fatalf("currentStep = %d, want 2", repo.state.CurrentStep)
	}
	if len(repo.state.StepsFired) != 1 || repo.state.StepsFired[0] != 1 {
		t.Fatalf("stepsFired = %v, want [1]", repo.state.StepsFired)
	}
	if len(tgSender.notifications) != 1 || len(emailSender.notifications) != 1 {
		t.Fatalf("telegram sends=%d email sends=%d, want 1 each", len(tgSender.notifications), len(emailSender.notifications))
	}
	var tgConfig map[string]any
	if err := json.Unmarshal(tgSender.notifications[0].Config, &tgConfig); err != nil {
		t.Fatalf("json.Unmarshal telegram config: %v", err)
	}
	if tgConfig["chatId"] != "chat-1" {
		t.Fatalf("telegram config chatId = %v, want chat-1", tgConfig["chatId"])
	}
	var emailConfig map[string]any
	if err := json.Unmarshal(emailSender.notifications[0].Config, &emailConfig); err != nil {
		t.Fatalf("json.Unmarshal email config: %v", err)
	}
	if emailConfig["toEmail"] != "ops@example.com" {
		t.Fatalf("email config toEmail = %v, want ops@example.com", emailConfig["toEmail"])
	}
	if scheduler.called {
		t.Fatal("did not expect scheduler to be called for single-step policy")
	}
}

func TestHandleIncidentDownResolvesChannelID(t *testing.T) {
	tgSender := &fakeSender{}
	repo := &fakeEscalationRepository{
		service:  &serviceRecord{TenantID: "DEFAULT", ServiceID: "auth", EscalationPolicyID: "POL_1"},
		policy:   &escalation.EscalationPolicy{TenantID: "DEFAULT", PolicyID: "POL_1", Name: "Primary", OffHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{{ChannelID: "CH_1", DelayMinutes: 0}}}},
		channels: map[string]escalation.NotificationChannel{"CH_1": {TenantID: "DEFAULT", ChannelID: "CH_1", Name: "Telegram", Type: escalation.ChannelTypeTelegram, Target: "chat-1", Config: json.RawMessage(`{"botToken":"token"}`)}},
	}
	handler := newTestEscalationHandlerWithDependencies(repo, &fakeScheduler{}, notifications.SenderRegistry{"telegram": tgSender}, testEscalationNow)
	jsonEvent, _ := notifications.NotificationEvent{EventType: notifications.EventTypeIncidentDown, TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "public-http", IncidentID: "INC_1", Timestamp: time.Date(2026, 6, 16, 2, 0, 0, 0, time.UTC)}.ToJSON()

	if err := handler.handleSQSEvent(context.Background(), events.SQSEvent{Records: []events.SQSMessage{{Body: jsonEvent}}}); err != nil {
		t.Fatalf("handleSQSEvent returned error: %v", err)
	}
	if len(tgSender.notifications) != 1 {
		t.Fatalf("notifications = %d, want 1", len(tgSender.notifications))
	}
	var config map[string]any
	if err := json.Unmarshal(tgSender.notifications[0].Config, &config); err != nil {
		t.Fatalf("json.Unmarshal config: %v", err)
	}
	if config["chatId"] != "chat-1" || config["botToken"] != "token" {
		t.Fatalf("config = %+v", config)
	}
}

func TestEscalationDispatchUsesProductionSenderRegistryWithInjectedExecutor(t *testing.T) {
	executor := &capturingEscalationExecutor{}
	repo := &fakeEscalationRepository{
		channels: map[string]escalation.NotificationChannel{
			"CH_1": {TenantID: "DEFAULT", ChannelID: "CH_1", Name: "Webhook", Type: escalation.ChannelTypeWebhook, Target: "https://hooks.example.com", Config: json.RawMessage(`{"headers":{"Authorization":"Bearer secret"}}`)},
		},
	}
	handler := newTestEscalationHandlerWithDependencies(repo, &fakeScheduler{}, notifications.NewSenderRegistry(executor), testEscalationNow)
	event := notifications.NotificationEvent{EventType: notifications.EventTypeIncidentDown, TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "public-http", IncidentID: "INC_1", Timestamp: time.Date(2026, 6, 16, 2, 0, 0, 0, time.UTC), Message: "disk full"}

	if err := handler.fireStep(context.Background(), event, escalation.EscalationStep{ChannelID: "CH_1"}); err != nil {
		t.Fatalf("fireStep returned error: %v", err)
	}
	if executor.request.URL != "https://hooks.example.com" || executor.request.Header.Get("Authorization") != "Bearer secret" {
		t.Fatalf("request = %#v", executor.request)
	}
	if executor.request.Timeout != outboundhttp.NotificationTimeout || executor.request.Body == nil {
		t.Fatalf("request = %#v", executor.request)
	}
}

func TestHandleIncidentDownSkipsDeletedChannel(t *testing.T) {
	tgSender := &fakeSender{}
	repo := &fakeEscalationRepository{
		service: &serviceRecord{TenantID: "DEFAULT", ServiceID: "auth", EscalationPolicyID: "POL_1"},
		policy:  &escalation.EscalationPolicy{TenantID: "DEFAULT", PolicyID: "POL_1", Name: "Primary", OffHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{{ChannelID: "CH_MISSING", DelayMinutes: 0}}}},
	}
	handler := newTestEscalationHandlerWithDependencies(repo, &fakeScheduler{}, notifications.SenderRegistry{"telegram": tgSender}, testEscalationNow)
	jsonEvent, _ := notifications.NotificationEvent{EventType: notifications.EventTypeIncidentDown, TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "public-http", IncidentID: "INC_1", Timestamp: time.Date(2026, 6, 16, 2, 0, 0, 0, time.UTC)}.ToJSON()

	if err := handler.handleSQSEvent(context.Background(), events.SQSEvent{Records: []events.SQSMessage{{Body: jsonEvent}}}); err != nil {
		t.Fatalf("handleSQSEvent returned error: %v", err)
	}
	if len(tgSender.notifications) != 0 {
		t.Fatalf("notifications = %d, want 0", len(tgSender.notifications))
	}
}

func TestHandleIncidentUpSuppressesEscalationState(t *testing.T) {
	repo := &fakeEscalationRepository{
		state: &escalation.EscalationState{TenantID: "DEFAULT", IncidentID: "INC_1", PolicyID: "POL_1", ServiceID: "auth", MonitorID: "public-http", CurrentStep: 2, Status: escalation.EscalationStatusActive, CreatedAt: "2026-06-16T00:00:00Z", UpdatedAt: "2026-06-16T00:00:00Z"},
	}
	handler := newTestEscalationHandler(repo, &fakeScheduler{})
	jsonEvent, _ := notifications.NotificationEvent{EventType: notifications.EventTypeIncidentUp, TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "public-http", IncidentID: "INC_1", Timestamp: time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)}.ToJSON()
	err := handler.handleSQSEvent(context.Background(), events.SQSEvent{Records: []events.SQSMessage{{Body: jsonEvent}}})
	if err != nil {
		t.Fatalf("handleSQSEvent returned error: %v", err)
	}
	if repo.state == nil {
		t.Fatal("expected escalation state to remain present")
	}
	if repo.state.Status != escalation.EscalationStatusSuppressed {
		t.Fatalf("status = %q, want %q", repo.state.Status, escalation.EscalationStatusSuppressed)
	}
	if repo.state.UpdatedAt != "2026-06-16T12:00:00Z" {
		t.Fatalf("updatedAt = %q, want 2026-06-16T12:00:00Z", repo.state.UpdatedAt)
	}
}

func TestHandleIncidentDownSchedulesNextStep(t *testing.T) {
	tgSender := &fakeSender{}
	scheduler := &fakeScheduler{}
	repo := &fakeEscalationRepository{
		service:  &serviceRecord{TenantID: "DEFAULT", ServiceID: "auth", EscalationPolicyID: "POL_1", BusinessHours: &escalation.BusinessHoursConfig{Timezone: "UTC", StartHour: 9, EndHour: 17, DaysOfWeek: []int{2}}},
		policy:   &escalation.EscalationPolicy{TenantID: "DEFAULT", PolicyID: "POL_1", Name: "Primary", BusinessHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{{DelayMinutes: 0, Channels: []escalation.ChannelConfig{{Type: escalation.ChannelTypeTelegram, Target: "chat-1", Config: json.RawMessage(`{"botToken":"token"}`)}}}, {DelayMinutes: 15, Channels: []escalation.ChannelConfig{{Type: escalation.ChannelTypeTelegram, Target: "chat-2", Config: json.RawMessage(`{"botToken":"token"}`)}}}}}},
		incident: &incidentRecord{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "public-http", IncidentID: "INC_1", Status: incidentStatusOpen, OpenedAt: "2026-06-16T10:00:00Z", UpdatedAt: "2026-06-16T10:00:00Z"},
	}
	now := time.Date(2026, 6, 16, 10, 0, 0, 0, time.UTC)
	handler := newTestEscalationHandlerWithDependencies(repo, scheduler, notifications.SenderRegistry{"telegram": tgSender}, now)
	jsonEvent, _ := notifications.NotificationEvent{EventType: notifications.EventTypeIncidentDown, TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "public-http", IncidentID: "INC_1", Timestamp: now}.ToJSON()
	err := handler.handleSQSEvent(context.Background(), events.SQSEvent{Records: []events.SQSMessage{{Body: jsonEvent}}})
	if err != nil {
		t.Fatalf("handleSQSEvent returned error: %v", err)
	}
	if !scheduler.called {
		t.Fatal("expected scheduler to be called")
	}
	if scheduler.event.Step != 2 {
		t.Fatalf("scheduled step = %d, want 2", scheduler.event.Step)
	}
	if !scheduler.when.Equal(now.Add(15 * time.Minute)) {
		t.Fatalf("scheduled time = %v, want %v", scheduler.when, now.Add(15*time.Minute))
	}
	if repo.state == nil || repo.state.ScheduledFor != now.Add(15*time.Minute).Format(time.RFC3339) {
		t.Fatalf("scheduledFor = %+v, want %s", repo.state, now.Add(15*time.Minute).Format(time.RFC3339))
	}
}

func TestHandleScheduledInvocationFiresNextStep(t *testing.T) {
	tgSender := &fakeSender{}
	scheduler := &fakeScheduler{}
	repo := &fakeEscalationRepository{
		state:    &escalation.EscalationState{TenantID: "DEFAULT", IncidentID: "INC_1", PolicyID: "POL_1", ServiceID: "auth", MonitorID: "public-http", CurrentStep: 2, StepsFired: []int{1}, SelectedPath: pathBusinessHours, Status: escalation.EscalationStatusActive, CreatedAt: "2026-06-16T00:00:00Z", UpdatedAt: "2026-06-16T00:00:00Z"},
		policy:   &escalation.EscalationPolicy{TenantID: "DEFAULT", PolicyID: "POL_1", Name: "Primary", BusinessHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{{DelayMinutes: 0, Channels: []escalation.ChannelConfig{{Type: escalation.ChannelTypeTelegram, Target: "chat-1", Config: json.RawMessage(`{"botToken":"token"}`)}}}, {DelayMinutes: 0, Channels: []escalation.ChannelConfig{{Type: escalation.ChannelTypeTelegram, Target: "chat-2", Config: json.RawMessage(`{"botToken":"token"}`)}}}}}},
		incident: &incidentRecord{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "public-http", IncidentID: "INC_1", Status: incidentStatusOpen, OpenedAt: "2026-06-16T10:00:00Z", UpdatedAt: "2026-06-16T10:00:00Z"},
	}
	now := time.Date(2026, 6, 16, 10, 15, 0, 0, time.UTC)
	handler := newTestEscalationHandlerWithDependencies(repo, scheduler, notifications.SenderRegistry{"telegram": tgSender}, now)
	if err := handler.handleScheduledInvocation(context.Background(), scheduledInvocationEvent{IncidentID: "INC_1", Step: 2}); err != nil {
		t.Fatalf("handleScheduledInvocation returned error: %v", err)
	}
	if len(tgSender.notifications) != 1 {
		t.Fatalf("notifications = %d, want 1", len(tgSender.notifications))
	}
	if repo.state.CurrentStep != 3 {
		t.Fatalf("currentStep = %d, want 3", repo.state.CurrentStep)
	}
	if len(repo.state.StepsFired) != 2 || repo.state.StepsFired[1] != 2 {
		t.Fatalf("stepsFired = %v, want [1 2]", repo.state.StepsFired)
	}
	if repo.state.Status != escalation.EscalationStatusExhausted {
		t.Fatalf("status = %q, want %q", repo.state.Status, escalation.EscalationStatusExhausted)
	}
	if repo.createdIncident == nil {
		t.Fatal("expected exhaustion incident to be created")
	}
	if repo.createdIncident.Type != "escalation.exhausted" {
		t.Fatalf("type = %q, want escalation.exhausted", repo.createdIncident.Type)
	}
	if repo.createdIncident.OriginalIncidentID != "INC_1" {
		t.Fatalf("originalIncidentID = %q, want INC_1", repo.createdIncident.OriginalIncidentID)
	}
}

func TestScheduledInvocationSchedulesNextStep(t *testing.T) {
	tgSender := &fakeSender{}
	scheduler := &fakeScheduler{}
	repo := &fakeEscalationRepository{
		service:  &serviceRecord{TenantID: "DEFAULT", ServiceID: "auth", EscalationPolicyID: "POL_1", BusinessHours: &escalation.BusinessHoursConfig{Timezone: "UTC", StartHour: 9, EndHour: 17, DaysOfWeek: []int{2}}},
		state:    &escalation.EscalationState{TenantID: "DEFAULT", IncidentID: "INC_1", PolicyID: "POL_1", ServiceID: "auth", MonitorID: "public-http", CurrentStep: 2, StepsFired: []int{1}, SelectedPath: pathBusinessHours, Status: escalation.EscalationStatusActive, CreatedAt: "2026-06-16T00:00:00Z", UpdatedAt: "2026-06-16T00:00:00Z"},
		policy:   &escalation.EscalationPolicy{TenantID: "DEFAULT", PolicyID: "POL_1", Name: "Primary", BusinessHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{{DelayMinutes: 0, Channels: []escalation.ChannelConfig{{Type: escalation.ChannelTypeTelegram, Target: "chat-1", Config: json.RawMessage(`{"botToken":"token"}`)}}}, {DelayMinutes: 30, Channels: []escalation.ChannelConfig{{Type: escalation.ChannelTypeTelegram, Target: "chat-2", Config: json.RawMessage(`{"botToken":"token"}`)}}}, {DelayMinutes: 45, Channels: []escalation.ChannelConfig{{Type: escalation.ChannelTypeTelegram, Target: "chat-3", Config: json.RawMessage(`{"botToken":"token"}`)}}}}}},
		incident: &incidentRecord{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "public-http", IncidentID: "INC_1", Status: incidentStatusOpen, OpenedAt: "2026-06-16T10:00:00Z", UpdatedAt: "2026-06-16T10:00:00Z"},
	}
	now := time.Date(2026, 6, 16, 10, 15, 0, 0, time.UTC)
	handler := newTestEscalationHandlerWithDependencies(repo, scheduler, notifications.SenderRegistry{"telegram": tgSender}, now)
	if err := handler.handleScheduledInvocation(context.Background(), scheduledInvocationEvent{IncidentID: "INC_1", Step: 2}); err != nil {
		t.Fatalf("handleScheduledInvocation returned error: %v", err)
	}
	if !scheduler.called {
		t.Fatal("expected scheduler to be called for delayed step 3")
	}
	if scheduler.event.Step != 3 {
		t.Fatalf("scheduled step = %d, want 3", scheduler.event.Step)
	}
	if !scheduler.when.Equal(now.Add(45 * time.Minute)) {
		t.Fatalf("scheduled when = %v, want %v", scheduler.when, now.Add(45*time.Minute))
	}
	if repo.state.ScheduledFor != now.Add(45*time.Minute).Format(time.RFC3339) {
		t.Fatalf("scheduledFor = %q, want %s", repo.state.ScheduledFor, now.Add(45*time.Minute).Format(time.RFC3339))
	}
	if repo.state.Status != escalation.EscalationStatusActive {
		t.Fatalf("status = %q, want %q", repo.state.Status, escalation.EscalationStatusActive)
	}
}

func TestSuppressionBeforeDelayedStepSkipsFiring(t *testing.T) {
	tgSender := &fakeSender{}
	scheduler := &fakeScheduler{}
	repo := &fakeEscalationRepository{
		state: &escalation.EscalationState{TenantID: "DEFAULT", IncidentID: "INC_1", PolicyID: "POL_1", ServiceID: "auth", MonitorID: "public-http", CurrentStep: 2, StepsFired: []int{1}, SelectedPath: pathBusinessHours, Status: escalation.EscalationStatusActive, CreatedAt: "2026-06-16T00:00:00Z", UpdatedAt: "2026-06-16T00:00:00Z", ScheduledFor: "2026-06-16T10:15:00Z"},
	}
	handler := newTestEscalationHandlerWithDependencies(repo, scheduler, notifications.SenderRegistry{"telegram": tgSender}, testEscalationNow)
	if err := handler.handleIncidentUp(context.Background(), notifications.NotificationEvent{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "public-http", IncidentID: "INC_1", Timestamp: time.Date(2026, 6, 16, 10, 10, 0, 0, time.UTC)}); err != nil {
		t.Fatalf("handleIncidentUp returned error: %v", err)
	}
	if err := handler.handleScheduledInvocation(context.Background(), scheduledInvocationEvent{IncidentID: "INC_1", Step: 2}); err != nil {
		t.Fatalf("handleScheduledInvocation returned error: %v", err)
	}
	if len(tgSender.notifications) != 0 {
		t.Fatalf("expected 0 notifications, got %d", len(tgSender.notifications))
	}
	if repo.state.Status != escalation.EscalationStatusSuppressed {
		t.Fatalf("status = %q, want %q", repo.state.Status, escalation.EscalationStatusSuppressed)
	}
	if scheduler.called {
		t.Fatal("scheduler should not be called when state is suppressed")
	}
}

func TestExhaustionSkippedWhenOriginalIncidentResolved(t *testing.T) {
	tgSender := &fakeSender{}
	scheduler := &fakeScheduler{}
	repo := &fakeEscalationRepository{
		state:    &escalation.EscalationState{TenantID: "DEFAULT", IncidentID: "INC_1", PolicyID: "POL_1", ServiceID: "auth", MonitorID: "public-http", CurrentStep: 2, StepsFired: []int{1}, SelectedPath: pathBusinessHours, Status: escalation.EscalationStatusActive, CreatedAt: "2026-06-16T00:00:00Z", UpdatedAt: "2026-06-16T00:00:00Z"},
		policy:   &escalation.EscalationPolicy{TenantID: "DEFAULT", PolicyID: "POL_1", Name: "Primary", BusinessHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{{DelayMinutes: 0, Channels: []escalation.ChannelConfig{{Type: escalation.ChannelTypeTelegram, Target: "chat-1", Config: json.RawMessage(`{"botToken":"token"}`)}}}, {DelayMinutes: 0, Channels: []escalation.ChannelConfig{{Type: escalation.ChannelTypeTelegram, Target: "chat-2", Config: json.RawMessage(`{"botToken":"token"}`)}}}}}},
		incident: &incidentRecord{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "public-http", IncidentID: "INC_1", Status: incidentStatusResolved, OpenedAt: "2026-06-16T10:00:00Z", UpdatedAt: "2026-06-16T10:14:00Z", ResolvedAt: "2026-06-16T10:14:00Z"},
	}
	now := time.Date(2026, 6, 16, 10, 15, 0, 0, time.UTC)
	handler := newTestEscalationHandlerWithDependencies(repo, scheduler, notifications.SenderRegistry{"telegram": tgSender}, now)
	if err := handler.handleScheduledInvocation(context.Background(), scheduledInvocationEvent{IncidentID: "INC_1", Step: 2}); err != nil {
		t.Fatalf("handleScheduledInvocation returned error: %v", err)
	}
	if repo.state.Status != escalation.EscalationStatusExhausted {
		t.Fatalf("status = %q, want %q", repo.state.Status, escalation.EscalationStatusExhausted)
	}
	if repo.createdIncident != nil {
		t.Fatalf("expected no exhausted incident, got %+v", repo.createdIncident)
	}
}

func TestScheduledInvocationDoesNotScheduleAfterFinalStep(t *testing.T) {
	tgSender := &fakeSender{}
	scheduler := &fakeScheduler{}
	repo := &fakeEscalationRepository{
		service:  &serviceRecord{TenantID: "DEFAULT", ServiceID: "auth", EscalationPolicyID: "POL_1", BusinessHours: &escalation.BusinessHoursConfig{Timezone: "UTC", StartHour: 9, EndHour: 17, DaysOfWeek: []int{2}}},
		state:    &escalation.EscalationState{TenantID: "DEFAULT", IncidentID: "INC_1", PolicyID: "POL_1", ServiceID: "auth", MonitorID: "public-http", CurrentStep: 2, StepsFired: []int{1}, SelectedPath: pathBusinessHours, Status: escalation.EscalationStatusActive, CreatedAt: "2026-06-16T00:00:00Z", UpdatedAt: "2026-06-16T00:00:00Z"},
		policy:   &escalation.EscalationPolicy{TenantID: "DEFAULT", PolicyID: "POL_1", Name: "Primary", BusinessHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{{DelayMinutes: 0, Channels: []escalation.ChannelConfig{{Type: escalation.ChannelTypeTelegram, Target: "chat-1", Config: json.RawMessage(`{"botToken":"token"}`)}}}, {DelayMinutes: 0, Channels: []escalation.ChannelConfig{{Type: escalation.ChannelTypeTelegram, Target: "chat-2", Config: json.RawMessage(`{"botToken":"token"}`)}}}}}},
		incident: &incidentRecord{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "public-http", IncidentID: "INC_1", Status: incidentStatusOpen, OpenedAt: "2026-06-16T10:00:00Z", UpdatedAt: "2026-06-16T10:00:00Z"},
	}
	now := time.Date(2026, 6, 16, 10, 15, 0, 0, time.UTC)
	handler := newTestEscalationHandlerWithDependencies(repo, scheduler, notifications.SenderRegistry{"telegram": tgSender}, now)
	if err := handler.handleScheduledInvocation(context.Background(), scheduledInvocationEvent{IncidentID: "INC_1", Step: 2}); err != nil {
		t.Fatalf("handleScheduledInvocation returned error: %v", err)
	}
	if scheduler.called {
		t.Fatal("scheduler should not be called when policy has no more steps")
	}
	if repo.state.Status != escalation.EscalationStatusExhausted {
		t.Fatalf("status = %q, want %q", repo.state.Status, escalation.EscalationStatusExhausted)
	}
	if repo.state.ScheduledFor != "" {
		t.Fatalf("scheduledFor = %q, want empty", repo.state.ScheduledFor)
	}
}
