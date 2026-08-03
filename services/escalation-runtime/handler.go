package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"bolt-monitor/shared/dynamodbrecord"
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
	LoadTransitionOutbox(context.Context, string, string) (*dynamodbrecord.TransitionOutboxRecord, error)
	AcknowledgeDispatch(context.Context, string, string) error
	CreateEscalationPlan(context.Context, notifications.EscalationPlan) error
	GetEscalationPlan(context.Context, string, string, string) (*notifications.EscalationPlan, error)
	CreateDelivery(context.Context, notifications.DeliveryRecord) error
	ListIncidentDeliveries(context.Context, string, string) ([]notifications.DeliveryRecord, error)
	ClaimDelivery(context.Context, string, string, string, time.Time, time.Duration) (*notifications.DeliveryRecord, string, error)
	CompleteDelivery(context.Context, notifications.DeliveryRecord, notifications.SendOutcome, string) error
	MarkDeliveryAmbiguous(context.Context, notifications.DeliveryRecord, notifications.SendOutcome) error
	AdvanceStepOnce(context.Context, string, string, int, int, string) error
	SuppressEscalation(context.Context, string, string, string) error
}

type escalationHandler struct {
	repo      escalationRepository
	scheduler scheduleClient
	senders   notifications.SenderRegistry
	now       func() time.Time
}

type escalationHandlerDependencies struct {
	senders notifications.SenderRegistry
	now     func() time.Time
}

func newEscalationHandlerWithDependencies(repo escalationRepository, scheduler scheduleClient, dependencies escalationHandlerDependencies) *escalationHandler {
	return &escalationHandler{
		repo:      repo,
		scheduler: scheduler,
		senders:   dependencies.senders,
		now:       dependencies.now,
	}
}

func (h *escalationHandler) handleScheduledInvocation(ctx context.Context, event scheduledInvocationEvent) error {
	tenantID := event.TenantID
	if tenantID == "" {
		tenantID = "DEFAULT"
	}
	state, err := h.repo.GetEscalationState(ctx, tenantID, event.IncidentID)
	if err != nil {
		return err
	}
	if state == nil || state.Status != escalation.EscalationStatusActive {
		return nil
	}
	if event.Step != state.CurrentStep {
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
	deliveries, err := h.repo.ListIncidentDeliveries(ctx, state.TenantID, state.IncidentID)
	if err != nil {
		return err
	}
	if len(deliveries) == 0 {
		if err := h.persistEscalationPlanAndDeliveries(ctx, state.IncidentID, notifEvent, *policy, state.SelectedPath, path); err != nil {
			return err
		}
	}
	if err := h.deliverStep(ctx, notifEvent, step, state.CurrentStep); err != nil {
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
	response, err := h.handleSQSEventResponse(ctx, event)
	if err != nil {
		return err
	}
	if len(response.BatchItemFailures) > 0 {
		return fmt.Errorf("notification message %s failed", response.BatchItemFailures[0].ItemIdentifier)
	}
	return nil
}

func (h *escalationHandler) handleSQSEventResponse(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
	response := events.SQSEventResponse{}
	for _, msg := range event.Records {
		if handled, err := h.handleScheduledStepEnvelope(ctx, msg.Body); handled {
			if err != nil {
				response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			}
			continue
		}
		if handled, err := h.handleTransitionEnvelope(ctx, msg.Body); handled {
			if err != nil {
				response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			}
			continue
		}
		eventData, err := notifications.ParseNotificationEvent(msg.Body)
		if err != nil {
			response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			continue
		}
		switch eventData.EventType {
		case notifications.EventTypeIncidentDown:
			if err := h.handleIncidentDown(ctx, eventData); err != nil {
				response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			}
		case notifications.EventTypeIncidentUp:
			if err := h.handleIncidentUp(ctx, eventData); err != nil {
				response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			}
		default:
			response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
		}
	}
	return response, nil
}

func (h *escalationHandler) handleScheduledStepEnvelope(ctx context.Context, body string) (bool, error) {
	var envelope notifications.CanonicalEnvelope
	if err := json.Unmarshal([]byte(body), &envelope); err != nil || envelope.Kind != notifications.CanonicalKindScheduled {
		return false, nil
	}
	if err := envelope.Validate(); err != nil {
		return true, err
	}
	if strings.TrimSpace(envelope.IncidentID) == "" || envelope.StepNumber <= 0 {
		return true, fmt.Errorf("scheduled step envelope requires incidentId and positive stepNumber")
	}
	return true, h.handleScheduledInvocation(ctx, scheduledInvocationEvent{
		TenantID:   envelope.TenantID,
		IncidentID: envelope.IncidentID,
		Step:       envelope.StepNumber,
	})
}

func (h *escalationHandler) handleTransitionEnvelope(ctx context.Context, body string) (bool, error) {
	var envelope struct {
		TenantID     string `json:"tenantId"`
		TransitionID string `json:"transitionId"`
		Kind         string `json:"kind"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil || envelope.Kind != "transition" {
		return false, nil
	}
	canonical, err := h.repo.LoadTransitionOutbox(ctx, envelope.TenantID, envelope.TransitionID)
	if err != nil {
		return true, err
	}
	if canonical == nil || canonical.DispatchStatus != "pending" {
		return true, nil
	}
	timestamp, err := time.Parse(time.RFC3339, canonical.CreatedAt)
	if err != nil {
		return true, err
	}
	event := notifications.NotificationEvent{
		TransitionID: canonical.EventID,
		EventType:    notifications.EventType(canonical.TransitionType), TenantID: canonical.TenantID,
		ServiceID: canonical.ServiceID, MonitorID: canonical.MonitorID, IncidentID: canonical.IncidentID,
		Timestamp: timestamp,
	}
	switch event.EventType {
	case notifications.EventTypeIncidentDown:
		err = h.handleIncidentDown(ctx, event)
	case notifications.EventTypeIncidentUp:
		err = h.handleIncidentUp(ctx, event)
	default:
		return true, fmt.Errorf("unsupported transition type %q", canonical.TransitionType)
	}
	if err != nil {
		return true, err
	}
	return true, h.repo.AcknowledgeDispatch(ctx, canonical.TenantID, canonical.EventID)
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
	policy, err := h.repo.GetEscalationPolicy(ctx, state.TenantID, state.PolicyID)
	if err != nil {
		return err
	}
	if policy != nil {
		if err := h.deliverRecovery(ctx, event, selectedPolicyPath(*policy, state.SelectedPath), *state); err != nil {
			return err
		}
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
	transitionID := event.DeliveryTransitionID()
	if transitionID != "" {
		if err := h.persistEscalationPlanAndDeliveries(ctx, transitionID, event, *policy, selectedPath, path); err != nil {
			return err
		}
	}
	if err := h.deliverStep(ctx, event, path.Steps[0], 1); err != nil {
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

func recoveryMessage(event notifications.NotificationEvent) string {
	return fmt.Sprintf("Incident Resolved\nService: %s\nMonitor: %s\nTime: %s", event.ServiceID, event.MonitorID, event.Timestamp.UTC().Format(time.RFC3339))
}

// deliverRecovery persists all recovery deliveries before provider I/O, then
// sends each unique channel used by a fired escalation step.
func (h *escalationHandler) deliverRecovery(ctx context.Context, event notifications.NotificationEvent, path escalation.EscalationPath, state escalation.EscalationState) error {
	type recoveryDelivery struct {
		channel    resolvedChannel
		stepNumber int
	}

	selected := make([]recoveryDelivery, 0)
	seen := make(map[string]struct{})
	for _, stepNumber := range state.StepsFired {
		stepIndex := stepNumber - 1
		if stepIndex < 0 || stepIndex >= len(path.Steps) {
			continue
		}
		channels, err := h.channelsForStep(ctx, event, path.Steps[stepIndex])
		if err != nil {
			return err
		}
		for _, channel := range channels {
			if _, exists := seen[channel.Key]; exists {
				continue
			}
			seen[channel.Key] = struct{}{}
			selected = append(selected, recoveryDelivery{channel: channel, stepNumber: stepNumber})
		}
	}

	transitionID := event.DeliveryTransitionID()
	now := h.now().UTC().Format(time.RFC3339)
	for _, selectedDelivery := range selected {
		delivery := notifications.DeliveryRecord{
			TenantID: event.TenantID, IncidentID: event.IncidentID, TransitionID: transitionID,
			DeliveryID: notifications.DeliveryIdentity(event.TenantID, transitionID, selectedDelivery.stepNumber, selectedDelivery.channel.Key),
			ChannelID:  selectedDelivery.channel.Key, ChannelType: string(selectedDelivery.channel.Channel.Type), StepNumber: selectedDelivery.stepNumber,
			State: notifications.DeliveryPending, CreatedAt: now, UpdatedAt: now,
		}
		if err := h.repo.CreateDelivery(ctx, delivery); err != nil {
			return fmt.Errorf("create recovery delivery: %w", err)
		}
	}

	event.Message = recoveryMessage(event)
	for _, selectedDelivery := range selected {
		if err := h.deliverResolvedChannel(ctx, event, selectedDelivery.channel, selectedDelivery.stepNumber); err != nil {
			return err
		}
	}
	return nil
}

const deliveryClaimLease = time.Minute

// deliverStep fences every provider request with its durable channel delivery.
func (h *escalationHandler) deliverStep(ctx context.Context, event notifications.NotificationEvent, step escalation.EscalationStep, stepNumber int) error {
	channels, err := h.channelsForStep(ctx, event, step)
	if err != nil {
		return err
	}
	for _, channel := range channels {
		if err := h.deliverResolvedChannel(ctx, event, channel, stepNumber); err != nil {
			return err
		}
	}
	return nil
}

func (h *escalationHandler) deliverResolvedChannel(ctx context.Context, event notifications.NotificationEvent, channel resolvedChannel, stepNumber int) error {
	deliveryID := notifications.DeliveryIdentity(event.TenantID, event.DeliveryTransitionID(), stepNumber, channel.Key)
	deliveries, err := h.repo.ListIncidentDeliveries(ctx, event.TenantID, event.IncidentID)
	if err != nil {
		return err
	}
	for _, delivery := range deliveries {
		if delivery.DeliveryID != deliveryID {
			continue
		}
		claimed, token, err := h.repo.ClaimDelivery(ctx, event.TenantID, event.IncidentID, deliveryID, h.now(), deliveryClaimLease)
		if err != nil {
			return err
		}
		if claimed == nil || token == "" {
			return nil
		}
		return h.sendClaimedDelivery(ctx, event, channel.Channel, *claimed)
	}
	return fmt.Errorf("delivery record missing for step %d channel %s", stepNumber, channel.Key)
}

func (h *escalationHandler) sendClaimedDelivery(ctx context.Context, event notifications.NotificationEvent, channel escalation.ChannelConfig, delivery notifications.DeliveryRecord) error {
	sender, ok := h.senders.Get(string(channel.Type))
	if !ok {
		return h.completeClaimedDelivery(ctx, delivery, notifications.SendOutcome{Class: notifications.OutcomeUnsupported})
	}
	config, err := mergeChannelTarget(channel)
	if err != nil {
		return h.completeClaimedDelivery(ctx, delivery, notifications.SendOutcome{Class: notifications.OutcomeInvalidConfig})
	}
	notification := notifications.Notification{EventType: event.EventType, MonitorID: event.MonitorID, ServiceID: event.ServiceID, TenantID: event.TenantID, MonitorName: event.MonitorName, ServiceName: event.ServiceName, Timestamp: event.Timestamp, Message: event.Message, IncidentID: event.IncidentID, Config: config}
	outcome, sendErr := sender.Send(ctx, notification)
	if outcome.Class == "" {
		outcome = notifications.SendOutcome{Class: notifications.OutcomeTransport, Retryable: true}
	}
	if sendErr != nil && outcome.Class == notifications.OutcomeAccepted {
		if err := h.repo.MarkDeliveryAmbiguous(ctx, delivery, outcome); err != nil {
			return fmt.Errorf("mark accepted delivery ambiguous: %w", err)
		}
		return fmt.Errorf("send %s notification: %w", channel.Type, sendErr)
	}
	if err := h.completeClaimedDelivery(ctx, delivery, outcome); err != nil {
		return err
	}
	if sendErr != nil {
		return fmt.Errorf("send %s notification: %w", channel.Type, sendErr)
	}
	return nil
}

func (h *escalationHandler) completeClaimedDelivery(ctx context.Context, delivery notifications.DeliveryRecord, outcome notifications.SendOutcome) error {
	if err := h.repo.CompleteDelivery(ctx, delivery, outcome, ""); err != nil {
		// A timeout after provider acceptance is not proof that the state write failed.
		// The conditional update cannot overwrite a completed delivery, but records an
		// ambiguous result when the original write did not take effect.
		if ambiguousErr := h.repo.MarkDeliveryAmbiguous(ctx, delivery, outcome); ambiguousErr != nil {
			return fmt.Errorf("complete delivery: %w; mark ambiguous: %v", err, ambiguousErr)
		}
		return fmt.Errorf("complete delivery: %w", err)
	}
	return nil
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
	return h.scheduler.ScheduleNextStep(ctx, scheduledInvocationEvent{TenantID: state.TenantID, IncidentID: state.IncidentID, Step: state.CurrentStep}, when)
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
	if _, err := sender.Send(ctx, notification); err != nil {
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
