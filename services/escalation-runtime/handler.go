package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/notifications"
	"github.com/aws/aws-lambda-go/events"
)

const (
	pathBusinessHours = "business-hours"
	pathOffHours      = "off-hours"
)

type escalationRepository interface {
	GetService(context.Context, string, string) (*serviceRecord, error)
	GetEscalationPolicy(context.Context, string, string) (*escalation.EscalationPolicy, error)
	PutEscalationState(context.Context, escalation.EscalationState) error
	GetEscalationState(context.Context, string, string) (*escalation.EscalationState, error)
	GetIncident(context.Context, string) (*incidentRecord, error)
	CreateIncident(context.Context, incidentRecord) error
	GetChannel(context.Context, string, string) (*escalation.NotificationChannel, error)
}

type escalationHandler struct {
	repo      escalationRepository
	scheduler scheduleClient
	senders   notifications.SenderRegistry
	now       func() time.Time
}

func newEscalationHandler(repo escalationRepository, scheduler scheduleClient) *escalationHandler {
	return &escalationHandler{
		repo:      repo,
		scheduler: scheduler,
		senders:   notifications.NewSenderRegistry(),
		now:       time.Now,
	}
}

func (h *escalationHandler) handleScheduledInvocation(ctx context.Context, event scheduledInvocationEvent) error {
	state, err := h.repo.GetEscalationState(ctx, "DEFAULT", event.IncidentID)
	if err != nil {
		return err
	}
	if state == nil || state.Status != escalation.EscalationStatusActive {
		return nil
	}
	policy, err := h.repo.GetEscalationPolicy(ctx, state.TenantID, state.PolicyID)
	if err != nil {
		return err
	}
	if policy == nil {
		return nil
	}
	path := selectedPolicyPath(*policy, state.SelectedPath)
	stepIndex := state.CurrentStep - 1
	if stepIndex < 0 || stepIndex >= len(path.Steps) {
		return nil
	}
	step := path.Steps[stepIndex]
	notifEvent := notifications.NotificationEvent{EventType: notifications.EventTypeIncidentDown, TenantID: state.TenantID, ServiceID: state.ServiceID, MonitorID: state.MonitorID, IncidentID: state.IncidentID, Timestamp: h.now().UTC(), Message: "Escalation step fired"}
	if err := h.fireStep(ctx, notifEvent, step); err != nil {
		return err
	}
	state.StepsFired = append(state.StepsFired, state.CurrentStep)
	state.CurrentStep++
	state.ScheduledFor = ""
	state.UpdatedAt = h.now().UTC().Format(time.RFC3339)
	if err := h.exhaustIfNeeded(ctx, state, path); err != nil {
		return err
	}
	if err := h.scheduleNextIfNeeded(ctx, state, path); err != nil {
		return err
	}
	return h.repo.PutEscalationState(ctx, *state)
}

func (h *escalationHandler) exhaustIfNeeded(ctx context.Context, state *escalation.EscalationState, path escalation.EscalationPath) error {
	if state.CurrentStep <= len(path.Steps) {
		return nil
	}
	original, err := h.repo.GetIncident(ctx, state.IncidentID)
	if err != nil {
		return err
	}
	state.Status = escalation.EscalationStatusExhausted
	state.ScheduledFor = ""
	if original == nil {
		return nil
	}
	if original.Status != incidentStatusOpen && original.Status != incidentStatusAcknowledged {
		return nil
	}
	return h.repo.CreateIncident(ctx, newEscalationExhaustedIncident(*original, h.now()))
}

func (h *escalationHandler) handleSQSEvent(ctx context.Context, event events.SQSEvent) error {
	for _, msg := range event.Records {
		eventData, err := notifications.ParseNotificationEvent(msg.Body)
		if err != nil {
			log.Printf("failed to parse escalation event: %v", err)
			continue
		}
		switch eventData.EventType {
		case notifications.EventTypeIncidentDown:
			if err := h.handleIncidentDown(ctx, eventData); err != nil {
				return err
			}
		case notifications.EventTypeIncidentUp:
			if err := h.handleIncidentUp(ctx, eventData); err != nil {
				return err
			}
		default:
			log.Printf("ignoring unsupported escalation event type %s", eventData.EventType)
		}
	}
	return nil
}

func (h *escalationHandler) handleIncidentUp(ctx context.Context, event notifications.NotificationEvent) error {
	state, err := h.repo.GetEscalationState(ctx, event.TenantID, event.IncidentID)
	if err != nil {
		return err
	}
	if state == nil {
		log.Printf("no escalation state found for incident %s", event.IncidentID)
		return nil
	}
	state.Status = escalation.EscalationStatusSuppressed
	state.UpdatedAt = event.Timestamp.UTC().Format(time.RFC3339)
	return h.repo.PutEscalationState(ctx, *state)
}

func (h *escalationHandler) handleIncidentDown(ctx context.Context, event notifications.NotificationEvent) error {
	service, err := h.repo.GetService(ctx, event.TenantID, event.ServiceID)
	if err != nil {
		return err
	}
	if service == nil || service.EscalationPolicyID == "" {
		log.Printf("service %s has no escalation policy; skipping incident %s", event.ServiceID, event.IncidentID)
		return nil
	}
	policy, err := h.repo.GetEscalationPolicy(ctx, event.TenantID, service.EscalationPolicyID)
	if err != nil {
		return err
	}
	if policy == nil {
		log.Printf("policy %s not found for service %s", service.EscalationPolicyID, event.ServiceID)
		return nil
	}
	selectedPath := pathOffHours
	path := policy.OffHoursPath
	if IsWithinBusinessHours(service.BusinessHours, event.Timestamp) {
		selectedPath = pathBusinessHours
		path = policy.BusinessHoursPath
	}
	if len(path.Steps) == 0 {
		log.Printf("policy %s has no steps for selected path %s", policy.PolicyID, selectedPath)
		return nil
	}
	if err := h.fireStep(ctx, event, path.Steps[0]); err != nil {
		return err
	}
	now := event.Timestamp.UTC().Format(time.RFC3339)
	state := escalation.EscalationState{
		TenantID:     event.TenantID,
		IncidentID:   event.IncidentID,
		PolicyID:     policy.PolicyID,
		ServiceID:    event.ServiceID,
		MonitorID:    event.MonitorID,
		CurrentStep:  2,
		StepsFired:   []int{1},
		SelectedPath: selectedPath,
		Status:       escalation.EscalationStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := h.scheduleNextIfNeeded(ctx, &state, path); err != nil {
		return err
	}
	return h.repo.PutEscalationState(ctx, state)
}

func (h *escalationHandler) scheduleNextIfNeeded(ctx context.Context, state *escalation.EscalationState, path escalation.EscalationPath) error {
	if h.scheduler == nil {
		return nil
	}
	stepIndex := state.CurrentStep - 1
	if stepIndex < 0 || stepIndex >= len(path.Steps) {
		return nil
	}
	step := path.Steps[stepIndex]
	if step.DelayMinutes <= 0 {
		return nil
	}
	when := h.now().UTC().Add(time.Duration(step.DelayMinutes) * time.Minute)
	state.ScheduledFor = when.Format(time.RFC3339)
	return h.scheduler.ScheduleNextStep(ctx, scheduledInvocationEvent{IncidentID: state.IncidentID, Step: state.CurrentStep}, when)
}

func selectedPolicyPath(policy escalation.EscalationPolicy, selectedPath string) escalation.EscalationPath {
	if selectedPath == pathBusinessHours {
		return policy.BusinessHoursPath
	}
	return policy.OffHoursPath
}

func (h *escalationHandler) fireStep(ctx context.Context, event notifications.NotificationEvent, step escalation.EscalationStep) error {
	if strings.TrimSpace(step.ChannelID) != "" {
		channel, err := h.repo.GetChannel(ctx, event.TenantID, step.ChannelID)
		if err != nil {
			return err
		}
		if channel == nil {
			log.Printf("route step skipped: channel %s was deleted", step.ChannelID)
			return nil
		}
		return h.sendToChannel(ctx, event, escalation.ChannelConfig{Type: channel.Type, Target: channel.Target, Config: channel.Config})
	}
	for _, channel := range step.Channels {
		if err := h.sendToChannel(ctx, event, channel); err != nil {
			return err
		}
	}
	return nil
}

func (h *escalationHandler) sendToChannel(ctx context.Context, event notifications.NotificationEvent, channel escalation.ChannelConfig) error {
	sender, ok := h.senders.Get(string(channel.Type))
	if !ok {
		return fmt.Errorf("no sender registered for channel type %s", channel.Type)
	}
	config, err := mergeChannelTarget(channel)
	if err != nil {
		return err
	}
	notification := notifications.Notification{
		EventType:   event.EventType,
		MonitorID:   event.MonitorID,
		ServiceID:   event.ServiceID,
		TenantID:    event.TenantID,
		MonitorName: event.MonitorName,
		ServiceName: event.ServiceName,
		Timestamp:   event.Timestamp,
		Message:     event.Message,
		IncidentID:  event.IncidentID,
		Config:      config,
	}
	if err := sender.Send(ctx, notification); err != nil {
		return fmt.Errorf("send %s notification: %w", channel.Type, err)
	}
	return nil
}

func mergeChannelTarget(channel escalation.ChannelConfig) ([]byte, error) {
	config := map[string]any{}
	if len(channel.Config) > 0 {
		if err := json.Unmarshal(channel.Config, &config); err != nil {
			return nil, fmt.Errorf("invalid %s config: %w", channel.Type, err)
		}
	}
	target := strings.TrimSpace(channel.Target)
	if target != "" {
		switch channel.Type {
		case escalation.ChannelTypeTelegram:
			config["chatId"] = target
		case escalation.ChannelTypeEmail:
			config["toEmail"] = target
		case escalation.ChannelTypeSMS:
			config["toNumber"] = target
		case escalation.ChannelTypeWebhook:
			config["url"] = target
		case escalation.ChannelTypePagerDuty:
			config["routingKey"] = target
		}
	}
	return json.Marshal(config)
}
