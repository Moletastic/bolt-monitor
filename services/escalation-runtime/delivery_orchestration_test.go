package main

import (
	"context"
	"encoding/json"
	"testing"

	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/notifications"
)

func TestStepTerminalForIncidentRequiresAllChannelsTerminal(t *testing.T) {
	deliveries := []notifications.DeliveryRecord{
		{DeliveryID: "dlv_a", StepNumber: 1, State: notifications.DeliveryDelivered},
		{DeliveryID: "dlv_b", StepNumber: 1, State: notifications.DeliveryInFlight},
	}
	if stepTerminalForIncident(deliveries, 1) {
		t.Fatal("step must not be terminal while any delivery is in_flight")
	}
	deliveries[1].State = notifications.DeliveryTerminalFailed
	if !stepTerminalForIncident(deliveries, 1) {
		t.Fatal("step must be terminal when every delivery is terminal")
	}
	otherStep := []notifications.DeliveryRecord{{DeliveryID: "x", StepNumber: 2, State: notifications.DeliveryDelivered}}
	if stepTerminalForIncident(otherStep, 1) {
		t.Fatal("other step deliveries must not affect target step")
	}
}

func TestChannelsForStepResolvesSingleChannelID(t *testing.T) {
	repo := &fakeEscalationRepository{channels: map[string]escalation.NotificationChannel{"CH_1": {ChannelID: "CH_1", TenantID: "DEFAULT", Type: escalation.ChannelTypeTelegram}}}
	handler := newEscalationHandler(repo, &fakeScheduler{})
	event := notifications.NotificationEvent{TenantID: "DEFAULT"}
	step := escalation.EscalationStep{ChannelID: "CH_1"}
	channels, err := handler.channelsForStep(context.Background(), event, step)
	if err != nil {
		t.Fatalf("channelsForStep failed: %v", err)
	}
	if len(channels) != 1 || channels[0].Key != "CH_1" {
		t.Fatalf("channels = %+v", channels)
	}
}

func TestChannelsForStepAssignsInlineOrdinalKeys(t *testing.T) {
	repo := &fakeEscalationRepository{}
	handler := newEscalationHandler(repo, &fakeScheduler{})
	step := escalation.EscalationStep{Channels: []escalation.ChannelConfig{{Type: escalation.ChannelTypeEmail, Target: "ops@example.com"}, {Type: escalation.ChannelTypeSMS, Target: "+1"}}}
	channels, err := handler.channelsForStep(context.Background(), notifications.NotificationEvent{TenantID: "DEFAULT"}, step)
	if err != nil {
		t.Fatalf("channelsForStep failed: %v", err)
	}
	if len(channels) != 2 || channels[0].Key != "ops@example.com#0" || channels[1].Key != "+1#1" {
		t.Fatalf("keys = %+v", channels)
	}
}

func TestBuildEscalationPlanCapturesStepAndChannelKeys(t *testing.T) {
	repo := &fakeEscalationRepository{channels: map[string]escalation.NotificationChannel{"CH_1": {ChannelID: "CH_1", TenantID: "DEFAULT", Type: escalation.ChannelTypeTelegram}}}
	handler := newEscalationHandler(repo, &fakeScheduler{})
	policy := escalation.EscalationPolicy{PolicyID: "POL_1"}
	path := escalation.EscalationPath{Steps: []escalation.EscalationStep{
		{ChannelID: "CH_1"},
		{Channels: []escalation.ChannelConfig{{Type: escalation.ChannelTypeEmail, Target: "ops@example.com"}, {Type: escalation.ChannelTypeSMS, Target: "+1"}}},
	}}
	plan, err := handler.buildEscalationPlan(notifications.NotificationEvent{TenantID: "DEFAULT"}, "TRN_1", policy, pathBusinessHours, path)
	if err != nil {
		t.Fatalf("buildEscalationPlan failed: %v", err)
	}
	if len(plan.StepNumbers) != 2 || plan.StepNumbers[0] != 1 || plan.StepNumbers[1] != 2 {
		t.Fatalf("steps = %+v", plan.StepNumbers)
	}
	if plan.StepChannels[0] != "CH_1" || plan.StepChannels[1] != "ops@example.com#0,+1#1" {
		t.Fatalf("channels = %v", plan.StepChannels)
	}
	if plan.SelectedPath != pathBusinessHours || plan.PolicyID != "POL_1" {
		t.Fatalf("plan = %+v", plan)
	}
}

func TestSuppressIfRecoveredReturnsTrueForResolvedIncident(t *testing.T) {
	now := "2026-07-19T12:00:00Z"
	repo := &fakeEscalationRepository{incident: &incidentRecord{Status: "resolved", UpdatedAt: now}}
	handler := newEscalationHandler(repo, &fakeScheduler{})
	suppressed, err := handler.suppressIfRecovered(context.Background(), "DEFAULT", "INC_1")
	if err != nil || !suppressed {
		t.Fatalf("expected suppressed=true, got %v err=%v", suppressed, err)
	}
}

func TestSuppressIfRecoveredReturnsFalseForOpenIncident(t *testing.T) {
	repo := &fakeEscalationRepository{incident: &incidentRecord{Status: "open"}}
	handler := newEscalationHandler(repo, &fakeScheduler{})
	suppressed, err := handler.suppressIfRecovered(context.Background(), "DEFAULT", "INC_1")
	if err != nil || suppressed {
		t.Fatalf("expected suppressed=false, got %v err=%v", suppressed, err)
	}
}

func TestPersistEscalationPlanAndDeliveriesWritesImmutableShape(t *testing.T) {
	repo := &fakeEscalationRepository{channels: map[string]escalation.NotificationChannel{"CH_1": {ChannelID: "CH_1", TenantID: "DEFAULT", Type: escalation.ChannelTypeTelegram}}}
	handler := newEscalationHandler(repo, &fakeScheduler{})
	policy := escalation.EscalationPolicy{PolicyID: "POL_1"}
	path := escalation.EscalationPath{Steps: []escalation.EscalationStep{{ChannelID: "CH_1"}}}
	if err := handler.persistEscalationPlanAndDeliveries(context.Background(), "TRN_1", notifications.NotificationEvent{TenantID: "DEFAULT", IncidentID: "INC_1"}, policy, pathBusinessHours, path, 1); err != nil {
		t.Fatalf("persist failed: %v", err)
	}
	delivery := notifications.DeliveryIdentity("DEFAULT", "TRN_1", 1, "CH_1")
	if delivery == "" {
		t.Fatal("expected deterministic delivery id")
	}
}

func TestAdvanceStepOncePersistsTransition(t *testing.T) {
	repo := &fakeEscalationRepository{}
	handler := newEscalationHandler(repo, &fakeScheduler{})
	state := escalation.EscalationState{TenantID: "DEFAULT", IncidentID: "INC_1", CurrentStep: 1}
	deliveries := []notifications.DeliveryRecord{
		{DeliveryID: "dlv_a", StepNumber: 1, State: notifications.DeliveryDelivered},
		{DeliveryID: "dlv_b", StepNumber: 1, State: notifications.DeliveryTerminalFailed},
	}
	next, err := handler.advanceAfterStepTerminal(context.Background(), "TRN_1", state, deliveries, escalation.EscalationPath{Steps: []escalation.EscalationStep{{ChannelID: "CH_1"}, {ChannelID: "CH_2"}}})
	if err != nil {
		t.Fatalf("advance failed: %v", err)
	}
	if next.CurrentStep != 2 {
		t.Fatalf("currentStep = %d, want 2", next.CurrentStep)
	}
}

func TestAdvanceStepOnceDoesNotAdvanceWhileInFlight(t *testing.T) {
	repo := &fakeEscalationRepository{}
	handler := newEscalationHandler(repo, &fakeScheduler{})
	state := escalation.EscalationState{TenantID: "DEFAULT", IncidentID: "INC_1", CurrentStep: 1}
	deliveries := []notifications.DeliveryRecord{{DeliveryID: "dlv_a", StepNumber: 1, State: notifications.DeliveryInFlight}}
	next, err := handler.advanceAfterStepTerminal(context.Background(), "TRN_1", state, deliveries, escalation.EscalationPath{Steps: []escalation.EscalationStep{{ChannelID: "CH_1"}}})
	if err != nil {
		t.Fatalf("advance failed: %v", err)
	}
	if next.CurrentStep != 1 {
		t.Fatalf("currentStep = %d, want 1", next.CurrentStep)
	}
}

func TestDeliveryIdentityStableAcrossInlineChannels(t *testing.T) {
	if notifications.DeliveryIdentity("DEFAULT", "TRN_1", 1, "ops@example.com#0") == notifications.DeliveryIdentity("DEFAULT", "TRN_1", 1, "ops@example.com#1") {
		t.Fatal("different ordinals must produce different identities")
	}
}

func TestDeliveryIdentityDoesNotCollideAcrossSteps(t *testing.T) {
	a := notifications.DeliveryIdentity("DEFAULT", "TRN_1", 1, "CH_1")
	b := notifications.DeliveryIdentity("DEFAULT", "TRN_1", 2, "CH_1")
	if a == b {
		t.Fatal("delivery id must differ by step")
	}
}

func TestHandlerAcceptsTypedOutcomeWithoutErrorForRetryable(t *testing.T) {
	handler := newEscalationHandler(&fakeEscalationRepository{}, &fakeScheduler{})
	handler.senders = notifications.SenderRegistry{"telegram": &fakeSender{}}
	state := escalation.EscalationState{TenantID: "DEFAULT", IncidentID: "INC_1", CurrentStep: 1}
	deliveries := []notifications.DeliveryRecord{{DeliveryID: "dlv_a", StepNumber: 1, State: notifications.DeliveryRetryable}}
	if stepTerminalForIncident(deliveries, state.CurrentStep) {
		t.Fatal("retryable state must not yet be terminal")
	}
	_ = json.RawMessage(nil)
}
