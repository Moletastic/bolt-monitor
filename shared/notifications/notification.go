package notifications

import (
	"encoding/json"
	"strings"
	"time"
)

type EventType string

const (
	EventTypeIncidentDown        EventType = "incident.down"
	EventTypeIncidentUp          EventType = "incident.up"
	EventTypeEscalationExhausted EventType = "escalation.exhausted"
)

type Notification struct {
	EventType   EventType       `json:"eventType"`
	MonitorID   string          `json:"monitorId"`
	ServiceID   string          `json:"serviceId"`
	TenantID    string          `json:"tenantId"`
	MonitorName string          `json:"monitorName"`
	ServiceName string          `json:"serviceName"`
	Timestamp   time.Time       `json:"timestamp"`
	Message     string          `json:"message"`
	IncidentID  string          `json:"incidentId,omitempty"`
	DeliveryID  string          `json:"deliveryId,omitempty"`
	ChannelType string          `json:"channelType,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
}

type NotificationEvent struct {
	TransitionID string    `json:"transitionId,omitempty"`
	EventType    EventType `json:"eventType"`
	TenantID     string    `json:"tenantId"`
	ServiceID    string    `json:"serviceId"`
	MonitorID    string    `json:"monitorId"`
	MonitorName  string    `json:"monitorName"`
	ServiceName  string    `json:"serviceName"`
	IncidentID   string    `json:"incidentId"`
	Timestamp    time.Time `json:"timestamp"`
	Message      string    `json:"message"`
}

// DeliveryTransitionID uses the canonical transition ID when available. Legacy
// recovery events use a separate identity to avoid colliding with outage work.
func (e NotificationEvent) DeliveryTransitionID() string {
	if transitionID := strings.TrimSpace(e.TransitionID); transitionID != "" {
		return transitionID
	}
	if e.EventType == EventTypeIncidentUp {
		return e.IncidentID + "#recovery"
	}
	return e.IncidentID
}

func (e NotificationEvent) ToJSON() (string, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ParseNotificationEvent(data string) (NotificationEvent, error) {
	var event NotificationEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return NotificationEvent{}, err
	}
	return event, nil
}
