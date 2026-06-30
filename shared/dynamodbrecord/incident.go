package dynamodbrecord

import (
	"strings"
	"time"

	"bolt-monitor/shared/dynamodbschema"
)

const IncidentRefEntityType = "IncidentRef"

type IncidentRecord struct {
	IncidentID         string
	ServiceID          string
	MonitorID          string
	TenantID           string
	Type               string
	Summary            string
	Status             string
	OpenedAt           string
	AcknowledgedAt     string
	ResolvedAt         string
	UpdatedAt          string
	Origin             string
	OriginalIncidentID string
}

type IncidentItemRecord struct {
	PK                 string `dynamodbav:"PK"`
	SK                 string `dynamodbav:"SK"`
	EntityType         string `dynamodbav:"EntityType"`
	TenantID           string `dynamodbav:"TenantID"`
	ServiceID          string `dynamodbav:"ServiceID"`
	MonitorID          string `dynamodbav:"MonitorID"`
	IncidentID         string `dynamodbav:"IncidentID"`
	Type               string `dynamodbav:"Type,omitempty"`
	Summary            string `dynamodbav:"Summary"`
	Status             string `dynamodbav:"Status"`
	OpenedAt           string `dynamodbav:"OpenedAt"`
	AcknowledgedAt     string `dynamodbav:"AcknowledgedAt,omitempty"`
	ResolvedAt         string `dynamodbav:"ResolvedAt,omitempty"`
	UpdatedAt          string `dynamodbav:"UpdatedAt"`
	Origin             string `dynamodbav:"Origin,omitempty"`
	OriginalIncidentID string `dynamodbav:"OriginalIncidentID,omitempty"`
	GSI1PK             string `dynamodbav:"GSI1PK,omitempty"`
	GSI1SK             string `dynamodbav:"GSI1SK,omitempty"`
}

func NewIncidentMonitorItemRecord(incident IncidentRecord) IncidentItemRecord {
	item := dynamodbschema.IncidentItem(incident.TenantID, incident.ServiceID, incident.MonitorID, incident.IncidentID, incident.OpenedAt, strings.ToUpper(incident.Status))
	return IncidentItemRecord{
		PK:                 item.PK,
		SK:                 item.SK,
		EntityType:         dynamodbschema.EntityIncident,
		TenantID:           dynamodbschema.NormalizeToken(incident.TenantID),
		ServiceID:          dynamodbschema.NormalizeField(incident.ServiceID),
		MonitorID:          dynamodbschema.NormalizeField(incident.MonitorID),
		IncidentID:         dynamodbschema.NormalizeToken(incident.IncidentID),
		Type:               normalizeIncidentType(incident.Type),
		Summary:            incident.Summary,
		Status:             incident.Status,
		OpenedAt:           incident.OpenedAt,
		AcknowledgedAt:     incident.AcknowledgedAt,
		ResolvedAt:         incident.ResolvedAt,
		UpdatedAt:          incident.UpdatedAt,
		Origin:             incident.Origin,
		OriginalIncidentID: dynamodbschema.NormalizeToken(incident.OriginalIncidentID),
		GSI1PK:             item.GSI1PK,
		GSI1SK:             item.GSI1SK,
	}
}

func NewIncidentRefItemRecord(incident IncidentRecord) IncidentItemRecord {
	return IncidentItemRecord{
		PK:                 dynamodbschema.TenantPK(incident.TenantID),
		SK:                 "INCIDENT#" + incident.OpenedAt + "#" + dynamodbschema.NormalizeToken(incident.IncidentID),
		EntityType:         IncidentRefEntityType,
		TenantID:           dynamodbschema.NormalizeToken(incident.TenantID),
		ServiceID:          dynamodbschema.NormalizeField(incident.ServiceID),
		MonitorID:          dynamodbschema.NormalizeField(incident.MonitorID),
		IncidentID:         dynamodbschema.NormalizeToken(incident.IncidentID),
		Type:               normalizeIncidentType(incident.Type),
		Summary:            incident.Summary,
		Status:             incident.Status,
		OpenedAt:           incident.OpenedAt,
		AcknowledgedAt:     incident.AcknowledgedAt,
		ResolvedAt:         incident.ResolvedAt,
		UpdatedAt:          incident.UpdatedAt,
		Origin:             incident.Origin,
		OriginalIncidentID: dynamodbschema.NormalizeToken(incident.OriginalIncidentID),
	}
}

func NewIncidentMetaItemRecord(incident IncidentRecord) IncidentItemRecord {
	return IncidentItemRecord{
		PK:                 dynamodbschema.IncidentPK(incident.IncidentID),
		SK:                 "META",
		EntityType:         dynamodbschema.EntityIncident,
		TenantID:           dynamodbschema.NormalizeToken(incident.TenantID),
		ServiceID:          dynamodbschema.NormalizeField(incident.ServiceID),
		MonitorID:          dynamodbschema.NormalizeField(incident.MonitorID),
		IncidentID:         dynamodbschema.NormalizeToken(incident.IncidentID),
		Type:               normalizeIncidentType(incident.Type),
		Summary:            incident.Summary,
		Status:             incident.Status,
		OpenedAt:           incident.OpenedAt,
		AcknowledgedAt:     incident.AcknowledgedAt,
		ResolvedAt:         incident.ResolvedAt,
		UpdatedAt:          incident.UpdatedAt,
		Origin:             incident.Origin,
		OriginalIncidentID: dynamodbschema.NormalizeToken(incident.OriginalIncidentID),
	}
}

func (r IncidentItemRecord) ToIncident() IncidentRecord {
	return IncidentRecord{
		IncidentID:         r.IncidentID,
		ServiceID:          r.ServiceID,
		MonitorID:          r.MonitorID,
		TenantID:           r.TenantID,
		Type:               normalizeIncidentType(r.Type),
		Summary:            r.Summary,
		Status:             strings.ToLower(r.Status),
		OpenedAt:           r.OpenedAt,
		AcknowledgedAt:     r.AcknowledgedAt,
		ResolvedAt:         r.ResolvedAt,
		UpdatedAt:          r.UpdatedAt,
		Origin:             r.Origin,
		OriginalIncidentID: r.OriginalIncidentID,
	}
}

type IncidentActivityRecord struct {
	PK         string `dynamodbav:"PK"`
	SK         string `dynamodbav:"SK"`
	EntityType string `dynamodbav:"EntityType"`
	TenantID   string `dynamodbav:"TenantID"`
	IncidentID string `dynamodbav:"IncidentID"`
	ActivityID string `dynamodbav:"ActivityID"`
	Action     string `dynamodbav:"Action"`
	Timestamp  string `dynamodbav:"Timestamp"`
}

func NewIncidentActivityRecord(tenantID, incidentID, activityID, action string, now time.Time) IncidentActivityRecord {
	item := dynamodbschema.IncidentActivityItem(tenantID, incidentID, activityID, now.UTC().Format(time.RFC3339))
	return IncidentActivityRecord{
		PK:         item.PK,
		SK:         item.SK,
		EntityType: dynamodbschema.EntityIncidentActivity,
		TenantID:   dynamodbschema.NormalizeToken(tenantID),
		IncidentID: dynamodbschema.NormalizeToken(incidentID),
		ActivityID: activityID,
		Action:     action,
		Timestamp:  now.UTC().Format(time.RFC3339),
	}
}

func normalizeIncidentType(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "monitoring"
	}
	return trimmed
}

func NormalizeIncidentType(value string) string { return normalizeIncidentType(value) }
