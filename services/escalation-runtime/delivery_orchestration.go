package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/notifications"
)

type deliveryDispatchRepository interface {
	escalationRepository
	CreateEscalationPlan(context.Context, notifications.EscalationPlan) error
	GetEscalationPlan(context.Context, string, string, string) (*notifications.EscalationPlan, error)
	CreateDelivery(context.Context, notifications.DeliveryRecord) error
	ListIncidentDeliveries(context.Context, string, string) ([]notifications.DeliveryRecord, error)
	AdvanceStepOnce(context.Context, string, string, int, int, string) error
	SuppressEscalation(context.Context, string, string, string) error
}

var _ deliveryDispatchRepository = (*dynamoEscalationRepository)(nil)

// resolvedChannel pairs a delivery-channel key with the channel config used to
// generate it. Inline step entries use a stable "<target>#<ordinal>" key.
type resolvedChannel struct {
	Key     string
	Channel escalation.ChannelConfig
}

// persistEscalationPlanAndDeliveries records the immutable route plan and one
// pending delivery per channel in the step, before any provider I/O. It does
// not perform provider calls; the caller schedules the SQS work.
func (h *escalationHandler) persistEscalationPlanAndDeliveries(ctx context.Context, transitionID string, event notifications.NotificationEvent, policy escalation.EscalationPolicy, selectedPath string, path escalation.EscalationPath, stepNumber int) error {
	plan, err := h.buildEscalationPlan(event, transitionID, policy, selectedPath, path)
	if err != nil {
		return err
	}
	if err := h.repo.CreateEscalationPlan(ctx, plan); err != nil {
		return fmt.Errorf("persist escalation plan: %w", err)
	}
	stepIndex := stepNumber - 1
	if stepIndex < 0 || stepIndex >= len(path.Steps) {
		return fmt.Errorf("step %d out of bounds", stepNumber)
	}
	channels, err := h.channelsForStep(ctx, event, path.Steps[stepIndex])
	if err != nil {
		return err
	}
	now := h.now().UTC().Format(time.RFC3339)
	for _, channel := range channels {
		delivery := notifications.DeliveryRecord{
			TenantID:     event.TenantID,
			IncidentID:   event.IncidentID,
			TransitionID: transitionID,
			DeliveryID:   notifications.DeliveryIdentity(event.TenantID, transitionID, stepNumber, channel.Key),
			ChannelID:    channel.Key,
			ChannelType:  string(channel.Channel.Type),
			StepNumber:   stepNumber,
			State:        notifications.DeliveryPending,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := h.repo.CreateDelivery(ctx, delivery); err != nil {
			return fmt.Errorf("create pending delivery: %w", err)
		}
	}
	return nil
}

func (h *escalationHandler) buildEscalationPlan(event notifications.NotificationEvent, transitionID string, policy escalation.EscalationPolicy, selectedPath string, path escalation.EscalationPath) (notifications.EscalationPlan, error) {
	now := h.now().UTC().Format(time.RFC3339)
	steps := make([]int, 0, len(path.Steps))
	channels := make([]string, 0, len(path.Steps))
	for index, step := range path.Steps {
		resolved, err := h.channelsForStep(context.Background(), event, step)
		if err != nil {
			return notifications.EscalationPlan{}, err
		}
		steps = append(steps, index+1)
		if len(resolved) == 1 {
			channels = append(channels, resolved[0].Key)
		} else {
			keys := make([]string, len(resolved))
			for i, c := range resolved {
				keys[i] = c.Key
			}
			channels = append(channels, strings.Join(keys, ","))
		}
	}
	return notifications.EscalationPlan{
		TenantID:     event.TenantID,
		IncidentID:   event.IncidentID,
		TransitionID: transitionID,
		PolicyID:     policy.PolicyID,
		SelectedPath: selectedPath,
		StepNumbers:  steps,
		StepChannels: channels,
		CreatedAt:    now,
	}, nil
}

// channelsForStep resolves the channels referenced by a step (single ID or
// inline list) without performing provider I/O. Inline step entries use a
// stable "<target>#<ordinal>" key so deliveries remain deterministic.
func (h *escalationHandler) channelsForStep(ctx context.Context, event notifications.NotificationEvent, step escalation.EscalationStep) ([]resolvedChannel, error) {
	if strings.TrimSpace(step.ChannelID) != "" {
		channel, err := h.repo.GetChannel(ctx, event.TenantID, step.ChannelID)
		if err != nil {
			return nil, err
		}
		if channel == nil {
			log.Printf("route step skipped: channel %s was deleted", step.ChannelID)
			return nil, nil
		}
		return []resolvedChannel{{Key: channel.ChannelID, Channel: escalation.ChannelConfig{Type: channel.Type, Target: channel.Target, Config: channel.Config}}}, nil
	}
	out := make([]resolvedChannel, 0, len(step.Channels))
	for index, channel := range step.Channels {
		key := fmt.Sprintf("%s#%d", strings.TrimSpace(channel.Target), index)
		out = append(out, resolvedChannel{Key: key, Channel: channel})
	}
	return out, nil
}

func stepTerminalForIncident(deliveries []notifications.DeliveryRecord, stepNumber int) bool {
	count := 0
	for _, d := range deliveries {
		if d.StepNumber != stepNumber {
			continue
		}
		count++
		if !d.State.IsTerminal() {
			return false
		}
	}
	return count > 0
}

// advanceAfterStepTerminal persists a single-step advance when every delivery
// for the step has reached a terminal state. Duplicate workers converge on
// the first successful advance; later advances are ignored.
func (h *escalationHandler) advanceAfterStepTerminal(ctx context.Context, _ string, state escalation.EscalationState, deliveries []notifications.DeliveryRecord, _ escalation.EscalationPath) (escalation.EscalationState, error) {
	if !stepTerminalForIncident(deliveries, state.CurrentStep) {
		return state, nil
	}
	nextStep := state.CurrentStep + 1
	now := h.now().UTC().Format(time.RFC3339)
	if err := h.repo.AdvanceStepOnce(ctx, state.TenantID, state.IncidentID, state.CurrentStep, nextStep, now); err != nil {
		return state, fmt.Errorf("advance step: %w", err)
	}
	state.CurrentStep = nextStep
	state.UpdatedAt = now
	return state, nil
}

// suppressIfRecovered re-reads the incident and escalation state before any
// scheduled work. A resolved incident or already-suppressed escalation causes
// the step to be dropped without provider I/O.
func (h *escalationHandler) suppressIfRecovered(ctx context.Context, tenantID, incidentID string) (bool, error) {
	incident, err := h.repo.GetIncident(ctx, incidentID)
	if err != nil {
		return false, err
	}
	if incident != nil && incident.Status == "resolved" {
		now := h.now().UTC().Format(time.RFC3339)
		if err := h.repo.SuppressEscalation(ctx, tenantID, incidentID, now); err != nil {
			return true, err
		}
		return true, nil
	}
	state, err := h.repo.GetEscalationState(ctx, tenantID, incidentID)
	if err != nil {
		return false, err
	}
	if state == nil {
		return false, nil
	}
	if state.Status == escalation.EscalationStatusSuppressed {
		return true, nil
	}
	return false, nil
}
