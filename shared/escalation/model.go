package escalation

import "encoding/json"

type ChannelType string

const (
	ChannelTypeTelegram  ChannelType = "telegram"
	ChannelTypeEmail     ChannelType = "email"
	ChannelTypeSMS       ChannelType = "sms"
	ChannelTypeWebhook   ChannelType = "webhook"
	ChannelTypePagerDuty ChannelType = "pagerduty"
)

type EscalationStatus string

const (
	EscalationStatusActive     EscalationStatus = "ACTIVE"
	EscalationStatusSuppressed EscalationStatus = "SUPPRESSED"
	EscalationStatusExhausted  EscalationStatus = "EXHAUSTED"
)

type ChannelConfig struct {
	Type   ChannelType     `json:"type"`
	Target string          `json:"target"`
	Config json.RawMessage `json:"config,omitempty"`
}

type NotificationChannel struct {
	ChannelID string          `json:"channelId"`
	TenantID  string          `json:"tenantId"`
	Name      string          `json:"name"`
	Type      ChannelType     `json:"type"`
	Target    string          `json:"target"`
	Config    json.RawMessage `json:"config,omitempty"`
	CreatedAt string          `json:"createdAt,omitempty"`
	UpdatedAt string          `json:"updatedAt,omitempty"`
}

type EscalationStep struct {
	ChannelID    string          `json:"channelId"`
	DelayMinutes int             `json:"delayMinutes"`
	Channels     []ChannelConfig `json:"-" dynamodbav:"Channels,omitempty"`
}

func (s EscalationStep) ResolvedChannelID() string {
	return s.ChannelID
}

type EscalationPath struct {
	Steps []EscalationStep `json:"steps"`
}

type EscalationPolicy struct {
	TenantID          string         `json:"tenantId"`
	PolicyID          string         `json:"policyId"`
	Name              string         `json:"name"`
	Description       string         `json:"description,omitempty"`
	BusinessHoursPath EscalationPath `json:"businessHoursPath"`
	OffHoursPath      EscalationPath `json:"offHoursPath"`
	CreatedAt         string         `json:"createdAt,omitempty"`
	UpdatedAt         string         `json:"updatedAt,omitempty"`
}

type EscalationState struct {
	TenantID     string           `json:"tenantId"`
	IncidentID   string           `json:"incidentId"`
	PolicyID     string           `json:"policyId"`
	ServiceID    string           `json:"serviceId"`
	MonitorID    string           `json:"monitorId"`
	CurrentStep  int              `json:"currentStep"`
	StepsFired   []int            `json:"stepsFired,omitempty"`
	SelectedPath string           `json:"selectedPath,omitempty"`
	ScheduledFor string           `json:"scheduledFor,omitempty"`
	Status       EscalationStatus `json:"status"`
	CreatedAt    string           `json:"createdAt,omitempty"`
	UpdatedAt    string           `json:"updatedAt,omitempty"`
}

type BusinessHoursConfig struct {
	Timezone   string `json:"timezone"`
	StartHour  int    `json:"startHour"`
	EndHour    int    `json:"endHour"`
	DaysOfWeek []int  `json:"daysOfWeek"`
}
