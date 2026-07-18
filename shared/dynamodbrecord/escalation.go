package dynamodbrecord

import (
	"encoding/json"
	"strings"

	"bolt-monitor/shared/dynamodbschema"
	"bolt-monitor/shared/escalation"
)

// EscalationPolicyItemRecord is the DynamoDB item representation of an
// escalation policy. The record preserves the established PK, SK, and
// attribute names so existing items remain reachable.
type EscalationPolicyItemRecord struct {
	PK                string                    `dynamodbav:"PK"`
	SK                string                    `dynamodbav:"SK"`
	EntityType        string                    `dynamodbav:"EntityType"`
	TenantID          string                    `dynamodbav:"TenantID"`
	PolicyID          string                    `dynamodbav:"PolicyID"`
	Name              string                    `dynamodbav:"Name"`
	Description       string                    `dynamodbav:"Description,omitempty"`
	BusinessHoursPath escalation.EscalationPath `dynamodbav:"BusinessHoursPath"`
	OffHoursPath      escalation.EscalationPath `dynamodbav:"OffHoursPath"`
	CreatedAt         string                    `dynamodbav:"CreatedAt,omitempty"`
	UpdatedAt         string                    `dynamodbav:"UpdatedAt,omitempty"`
}

// NewEscalationPolicyItemRecord constructs the canonical record from a domain
// escalation policy. The path fields are defensively copied.
func NewEscalationPolicyItemRecord(policy escalation.EscalationPolicy) EscalationPolicyItemRecord {
	return EscalationPolicyItemRecord{
		PK:                dynamodbschema.TenantPK(policy.TenantID),
		SK:                EscalationPolicySK(policy.PolicyID),
		EntityType:        dynamodbschema.EntityEscalationPolicy,
		TenantID:          dynamodbschema.NormalizeToken(policy.TenantID),
		PolicyID:          dynamodbschema.NormalizeToken(policy.PolicyID),
		Name:              strings.TrimSpace(policy.Name),
		Description:       strings.TrimSpace(policy.Description),
		BusinessHoursPath: escalation.CloneEscalationPath(policy.BusinessHoursPath),
		OffHoursPath:      escalation.CloneEscalationPath(policy.OffHoursPath),
		CreatedAt:         policy.CreatedAt,
		UpdatedAt:         policy.UpdatedAt,
	}
}

// ToEscalationPolicy converts the persisted record into the domain escalation
// policy, cloning the path fields so downstream mutations do not leak into the
// record.
func (r EscalationPolicyItemRecord) ToEscalationPolicy() escalation.EscalationPolicy {
	return escalation.EscalationPolicy{
		TenantID:          r.TenantID,
		PolicyID:          r.PolicyID,
		Name:              r.Name,
		Description:       r.Description,
		BusinessHoursPath: escalation.CloneEscalationPath(r.BusinessHoursPath),
		OffHoursPath:      escalation.CloneEscalationPath(r.OffHoursPath),
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}

// EscalationPolicySK returns the canonical sort key for a policy record. The
// ID is upper-cased and trimmed to match the existing schema.
func EscalationPolicySK(policyID string) string {
	return "ESCALATION_POLICY#" + dynamodbschema.NormalizeToken(policyID)
}

// NotificationChannelItemRecord is the DynamoDB item representation of a
// notification channel.
type NotificationChannelItemRecord struct {
	PK         string                 `dynamodbav:"PK"`
	SK         string                 `dynamodbav:"SK"`
	EntityType string                 `dynamodbav:"EntityType"`
	TenantID   string                 `dynamodbav:"TenantID"`
	ChannelID  string                 `dynamodbav:"ChannelID"`
	Name       string                 `dynamodbav:"Name"`
	Type       escalation.ChannelType `dynamodbav:"Type"`
	Target     string                 `dynamodbav:"Target"`
	Config     []byte                 `dynamodbav:"Config,omitempty"`
	CreatedAt  string                 `dynamodbav:"CreatedAt,omitempty"`
	UpdatedAt  string                 `dynamodbav:"UpdatedAt,omitempty"`
}

// NewNotificationChannelItemRecord constructs the canonical record from a
// domain notification channel. The Config payload is defensively copied.
func NewNotificationChannelItemRecord(channel escalation.NotificationChannel) NotificationChannelItemRecord {
	return NotificationChannelItemRecord{
		PK:         dynamodbschema.TenantPK(channel.TenantID),
		SK:         NotificationChannelSK(channel.ChannelID),
		EntityType: dynamodbschema.EntityNotificationChannel,
		TenantID:   dynamodbschema.NormalizeToken(channel.TenantID),
		ChannelID:  dynamodbschema.NormalizeToken(channel.ChannelID),
		Name:       strings.TrimSpace(channel.Name),
		Type:       channel.Type,
		Target:     strings.TrimSpace(channel.Target),
		Config:     append([]byte(nil), channel.Config...),
		CreatedAt:  channel.CreatedAt,
		UpdatedAt:  channel.UpdatedAt,
	}
}

// ToNotificationChannel converts the persisted record into the domain
// notification channel. The Config payload is cloned so downstream mutations
// do not leak into the record.
func (r NotificationChannelItemRecord) ToNotificationChannel() escalation.NotificationChannel {
	return escalation.NotificationChannel{
		TenantID:  r.TenantID,
		ChannelID: r.ChannelID,
		Name:      r.Name,
		Type:      r.Type,
		Target:    r.Target,
		Config:    append(json.RawMessage(nil), r.Config...),
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// NotificationChannelSK returns the canonical sort key for a channel record.
func NotificationChannelSK(channelID string) string {
	return "NOTIFICATION_CHANNEL#" + dynamodbschema.NormalizeToken(channelID)
}

// EscalationStateRecord is the in-memory representation of the persisted
// escalation state. It mirrors the existing schema so caller code can pass it
// through the shared facade.
type EscalationStateRecord struct {
	TenantID     string
	IncidentID   string
	PolicyID     string
	ServiceID    string
	MonitorID    string
	CurrentStep  int
	StepsFired   []int
	SelectedPath string
	ScheduledFor string
	Status       escalation.EscalationStatus
	CreatedAt    string
	UpdatedAt    string
}

// EscalationStateItemRecord is the DynamoDB item representation of the
// per-incident escalation state.
type EscalationStateItemRecord struct {
	PK           string                      `dynamodbav:"PK"`
	SK           string                      `dynamodbav:"SK"`
	EntityType   string                      `dynamodbav:"EntityType"`
	TenantID     string                      `dynamodbav:"TenantID"`
	IncidentID   string                      `dynamodbav:"IncidentID"`
	PolicyID     string                      `dynamodbav:"PolicyID"`
	ServiceID    string                      `dynamodbav:"ServiceID"`
	MonitorID    string                      `dynamodbav:"MonitorID"`
	CurrentStep  int                         `dynamodbav:"CurrentStep"`
	StepsFired   []int                       `dynamodbav:"StepsFired,omitempty"`
	SelectedPath string                      `dynamodbav:"SelectedPath,omitempty"`
	ScheduledFor string                      `dynamodbav:"ScheduledFor,omitempty"`
	Status       escalation.EscalationStatus `dynamodbav:"Status"`
	CreatedAt    string                      `dynamodbav:"CreatedAt,omitempty"`
	UpdatedAt    string                      `dynamodbav:"UpdatedAt,omitempty"`
}

// NewEscalationStateItemRecord constructs the canonical record from the
// domain escalation state, applying the established PK shape and cloning
// mutable slice fields.
func NewEscalationStateItemRecord(state escalation.EscalationState) EscalationStateItemRecord {
	item := dynamodbschema.EscalationStateItem(state.TenantID, state.IncidentID)
	return EscalationStateItemRecord{
		PK:           item.PK,
		SK:           item.SK,
		EntityType:   item.EntityType,
		TenantID:     dynamodbschema.NormalizeToken(state.TenantID),
		IncidentID:   dynamodbschema.NormalizeToken(state.IncidentID),
		PolicyID:     dynamodbschema.NormalizeToken(state.PolicyID),
		ServiceID:    dynamodbschema.NormalizeField(state.ServiceID),
		MonitorID:    dynamodbschema.NormalizeField(state.MonitorID),
		CurrentStep:  state.CurrentStep,
		StepsFired:   append([]int(nil), state.StepsFired...),
		SelectedPath: state.SelectedPath,
		ScheduledFor: state.ScheduledFor,
		Status:       state.Status,
		CreatedAt:    state.CreatedAt,
		UpdatedAt:    state.UpdatedAt,
	}
}

// ToEscalationState converts the persisted record back into the domain state.
// StepsFired is cloned so callers can append without mutating the record.
func (r EscalationStateItemRecord) ToEscalationState() escalation.EscalationState {
	return escalation.EscalationState{
		TenantID:     r.TenantID,
		IncidentID:   r.IncidentID,
		PolicyID:     r.PolicyID,
		ServiceID:    r.ServiceID,
		MonitorID:    r.MonitorID,
		CurrentStep:  r.CurrentStep,
		StepsFired:   append([]int(nil), r.StepsFired...),
		SelectedPath: r.SelectedPath,
		ScheduledFor: r.ScheduledFor,
		Status:       r.Status,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}
