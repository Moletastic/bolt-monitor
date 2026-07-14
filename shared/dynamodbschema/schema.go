package dynamodbschema

import (
	"fmt"
	"strings"
)

const (
	PrimaryTableName = "bolt-monitor-app"

	EntityWorkspace           = "Workspace"
	EntityService             = "Service"
	EntityServiceRef          = "ServiceRef"
	EntityServiceStatus       = "ServiceStatus"
	EntityServiceMonitorRef   = "ServiceMonitorRef"
	EntityMonitor             = "Monitor"
	EntityMonitorStatus       = "MonitorStatus"
	EntityCheckRun            = "CheckRun"
	EntityAlertState          = "AlertState"
	EntityIncident            = "Incident"
	EntityAuditEvent          = "AuditEvent"
	EntityAuditChange         = "AuditChange"
	EntityIncidentActivity    = "IncidentActivity"
	EntityExecutionWork       = "ExecutionWork"
	EntityEscalationPolicy    = "EscalationPolicy"
	EntityEscalationState     = "EscalationState"
	EntityNotificationChannel = "NotificationChannel"
	EntitySearchIndex         = "SearchIndex"
)

const (
	GSIOpenIncidents   = "gsi1"
	GSIServiceRollups  = "gsi2"
	GSIAuditByResource = "AuditByResourceIndex"

	DefaultCheckRunRetentionDays = 30
)

type Item struct {
	PK         string `json:"pk"`
	SK         string `json:"sk"`
	EntityType string `json:"entityType"`

	GSI1PK string `json:"gsi1pk,omitempty"`
	GSI1SK string `json:"gsi1sk,omitempty"`
	GSI2PK string `json:"gsi2pk,omitempty"`
	GSI2SK string `json:"gsi2sk,omitempty"`
	GSI3PK string `json:"gsi3pk,omitempty"`
	GSI3SK string `json:"gsi3sk,omitempty"`

	TenantID   string `json:"tenantId,omitempty"`
	ServiceID  string `json:"serviceId,omitempty"`
	MonitorID  string `json:"monitorId,omitempty"`
	RunID      string `json:"runId,omitempty"`
	IncidentID string `json:"incidentId,omitempty"`
	AuditID    string `json:"auditId,omitempty"`
	TTL        int64  `json:"ttl,omitempty"`
}

type AccessPattern struct {
	Name        string
	Description string
	PK          string
	SKPrefix    string
	Index       string
}

func WorkspaceItem(tenantID string) Item {
	return Item{PK: TenantPK(tenantID), SK: "META", EntityType: EntityWorkspace, TenantID: normalizeField(tenantID)}
}

func ServiceRefItem(tenantID, serviceID string) Item {
	return Item{PK: TenantPK(tenantID), SK: ServiceRefSK(serviceID), EntityType: EntityServiceRef, TenantID: normalizeField(tenantID), ServiceID: normalizeField(serviceID)}
}

func ServiceItem(tenantID, serviceID string) Item {
	return Item{PK: ServicePK(tenantID, serviceID), SK: "META", EntityType: EntityService, TenantID: normalizeField(tenantID), ServiceID: normalizeField(serviceID)}
}

func ServiceStatusItem(tenantID, serviceID, rollupStatus, updatedAt string) Item {
	return Item{
		PK:         ServicePK(tenantID, serviceID),
		SK:         "STATUS",
		EntityType: EntityServiceStatus,
		TenantID:   normalizeField(tenantID),
		ServiceID:  normalizeField(serviceID),
		GSI2PK:     TenantPK(tenantID),
		GSI2SK:     fmt.Sprintf("STATUS#%s#%s#%s", strings.ToUpper(strings.TrimSpace(rollupStatus)), updatedAt, normalizeToken(serviceID)),
	}
}

func ServiceMonitorRefItem(tenantID, serviceID, monitorID string) Item {
	return Item{PK: ServicePK(tenantID, serviceID), SK: ServiceMonitorRefSK(monitorID), EntityType: EntityServiceMonitorRef, TenantID: normalizeField(tenantID), ServiceID: normalizeField(serviceID), MonitorID: normalizeField(monitorID)}
}

func MonitorItem(tenantID, serviceID, monitorID string) Item {
	return Item{PK: MonitorPK(tenantID, serviceID, monitorID), SK: "META", EntityType: EntityMonitor, TenantID: normalizeField(tenantID), ServiceID: normalizeField(serviceID), MonitorID: normalizeField(monitorID)}
}

func MonitorStatusItem(tenantID, serviceID, monitorID, currentStatus, lastCheckedAt string) Item {
	return Item{
		PK:         MonitorPK(tenantID, serviceID, monitorID),
		SK:         "STATUS",
		EntityType: EntityMonitorStatus,
		TenantID:   normalizeField(tenantID),
		ServiceID:  normalizeField(serviceID),
		MonitorID:  normalizeField(monitorID),
	}
}

func CheckRunItem(tenantID, serviceID, monitorID, startedAt, runID string, ttl int64) Item {
	return Item{PK: MonitorPK(tenantID, serviceID, monitorID), SK: fmt.Sprintf("RUN#%s#%s", startedAt, normalizeToken(runID)), EntityType: EntityCheckRun, TenantID: normalizeField(tenantID), ServiceID: normalizeField(serviceID), MonitorID: normalizeField(monitorID), RunID: normalizeField(runID), TTL: ttl}
}

func AlertStateItem(tenantID, serviceID, monitorID string) Item {
	return Item{PK: MonitorPK(tenantID, serviceID, monitorID), SK: "ALERT_STATE", EntityType: EntityAlertState, TenantID: normalizeField(tenantID), ServiceID: normalizeField(serviceID), MonitorID: normalizeField(monitorID)}
}

func IncidentItem(tenantID, serviceID, monitorID, incidentID, openedAt, statusPrefix string) Item {
	return Item{
		PK:         MonitorPK(tenantID, serviceID, monitorID),
		SK:         fmt.Sprintf("INCIDENT#%s#%s", openedAt, normalizeToken(incidentID)),
		EntityType: EntityIncident,
		TenantID:   normalizeField(tenantID),
		ServiceID:  normalizeField(serviceID),
		MonitorID:  normalizeField(monitorID),
		IncidentID: normalizeField(incidentID),
		GSI1PK:     TenantPK(tenantID),
		GSI1SK:     fmt.Sprintf("%s#%s#%s", strings.ToUpper(statusPrefix), openedAt, normalizeToken(incidentID)),
	}
}

func AuditEventItem(tenantID, auditID, timestamp string) Item {
	return Item{PK: TenantPK(tenantID), SK: fmt.Sprintf("AUDIT#%s#%s", timestamp, normalizeToken(auditID)), EntityType: EntityAuditEvent, TenantID: normalizeField(tenantID), AuditID: normalizeField(auditID)}
}

func AuditResourceItem(tenantID, serviceID, monitorID, auditID, timestamp string) Item {
	return Item{
		GSI3PK: fmt.Sprintf("AUDIT_RESOURCE#%s#%s#%s", normalizeToken(tenantID), normalizeToken(serviceID), normalizeToken(monitorID)),
		GSI3SK: fmt.Sprintf("AUDIT#%s#%s", timestamp, normalizeToken(auditID)),
	}
}

func AuditChangeItem(auditID, fieldPath string) Item {
	return Item{PK: AuditPK(auditID), SK: "CHANGE#" + normalizeToken(fieldPath), EntityType: EntityAuditChange, AuditID: normalizeField(auditID)}
}

func ExecutionWorkItem(tenantID, requestedAt, runID string, ttl int64) Item {
	return Item{
		PK:         TenantPK(tenantID),
		SK:         fmt.Sprintf("RUN_REQUEST#%s#%s", requestedAt, normalizeToken(runID)),
		EntityType: EntityExecutionWork,
		TenantID:   normalizeField(tenantID),
		RunID:      normalizeField(runID),
		TTL:        ttl,
	}
}

func EscalationPolicyItem(tenantID, policyID string) Item {
	return Item{PK: TenantPK(tenantID), SK: "ESCALATION_POLICY#" + normalizeToken(policyID), EntityType: EntityEscalationPolicy, TenantID: normalizeField(tenantID)}
}

func NotificationChannelItem(tenantID, channelID string) Item {
	return Item{PK: TenantPK(tenantID), SK: "NOTIFICATION_CHANNEL#" + normalizeToken(channelID), EntityType: EntityNotificationChannel, TenantID: normalizeField(tenantID)}
}

func SearchIndexItem(tenantID, prefix, resourceType, resourceKey string) Item {
	return Item{PK: TenantPK(tenantID), SK: "SEARCH#" + normalizeSearchPrefix(prefix) + "#" + strings.ToUpper(strings.TrimSpace(resourceType)) + "#" + normalizeToken(resourceKey), EntityType: EntitySearchIndex, TenantID: normalizeField(tenantID)}
}

func EscalationStateItem(tenantID, incidentID string) Item {
	return Item{PK: IncidentPK(incidentID), SK: "ESCALATION_STATE", EntityType: EntityEscalationState, TenantID: normalizeField(tenantID), IncidentID: normalizeField(incidentID)}
}

func IncidentActivityItem(tenantID, incidentID, activityID, timestamp string) Item {
	return Item{PK: IncidentPK(incidentID), SK: fmt.Sprintf("ACTIVITY#%s#%s", timestamp, normalizeToken(activityID)), EntityType: EntityIncidentActivity, TenantID: normalizeField(tenantID), IncidentID: normalizeField(incidentID)}
}

func TenantPK(tenantID string) string { return "TENANT#" + normalizeToken(tenantID) }
func ServicePK(tenantID, serviceID string) string {
	return "SERVICE#" + normalizeToken(tenantID) + "#" + normalizeToken(serviceID)
}
func MonitorPK(tenantID, serviceID, monitorID string) string {
	return "MONITOR#" + normalizeToken(tenantID) + "#" + normalizeToken(serviceID) + "#" + normalizeToken(monitorID)
}
func ServiceRefSK(serviceID string) string        { return "SERVICE#" + normalizeToken(serviceID) }
func ServiceMonitorRefSK(monitorID string) string { return "MONITOR#" + normalizeToken(monitorID) }
func IncidentPK(incidentID string) string         { return "INCIDENT#" + normalizeToken(incidentID) }
func AuditPK(auditID string) string               { return "AUDIT#" + normalizeToken(auditID) }
func SearchIndexSKPrefix(query string) string     { return "SEARCH#" + normalizeSearchPrefix(query) }

func normalizeToken(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func normalizeField(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeSearchPrefix(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func NormalizeToken(value string) string { return normalizeToken(value) }

func NormalizeField(value string) string { return normalizeField(value) }

func AccessPatterns() []AccessPattern {
	return []AccessPattern{
		{Name: "list-services-for-tenant", Description: "List service refs for one tenant", PK: "TENANT#<tenantId>", SKPrefix: "SERVICE#"},
		{Name: "get-service-config", Description: "Get canonical service metadata", PK: "SERVICE#<tenantId>#<serviceId>", SKPrefix: "META"},
		{Name: "get-service-status", Description: "Get current service rollup snapshot", PK: "SERVICE#<tenantId>#<serviceId>", SKPrefix: "STATUS"},
		{Name: "get-service-monitors", Description: "List child monitor summaries for one service", PK: "SERVICE#<tenantId>#<serviceId>", SKPrefix: "MONITOR#"},
		{Name: "get-monitor-config", Description: "Get canonical nested monitor configuration", PK: "MONITOR#<tenantId>#<serviceId>#<monitorId>", SKPrefix: "META"},
		{Name: "get-monitor-status", Description: "Get current monitor status snapshot", PK: "MONITOR#<tenantId>#<serviceId>#<monitorId>", SKPrefix: "STATUS"},
		{Name: "get-recent-runs", Description: "Get recent run history for one monitor", PK: "MONITOR#<tenantId>#<serviceId>#<monitorId>", SKPrefix: "RUN#"},
		{Name: "get-open-incidents-for-tenant", Description: "List open incidents by tenant", PK: "TENANT#<tenantId>", SKPrefix: "INCIDENT_OPEN#", Index: GSIOpenIncidents},
		{Name: "get-service-rollup-dashboard", Description: "List service rollups by tenant", PK: "TENANT#<tenantId>", SKPrefix: "STATUS#", Index: GSIServiceRollups},
		{Name: "get-audit-history", Description: "List audit events for one tenant", PK: "TENANT#<tenantId>", SKPrefix: "AUDIT#"},
	}
}
