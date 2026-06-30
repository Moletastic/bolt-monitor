package notifications

import (
	"encoding/json"
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
	Config      json.RawMessage `json:"config,omitempty"`
}

type NotificationEvent struct {
	EventType   EventType `json:"eventType"`
	TenantID    string    `json:"tenantId"`
	ServiceID   string    `json:"serviceId"`
	MonitorID   string    `json:"monitorId"`
	MonitorName string    `json:"monitorName"`
	ServiceName string    `json:"serviceName"`
	IncidentID  string    `json:"incidentId"`
	Timestamp   time.Time `json:"timestamp"`
	Message     string    `json:"message"`
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
