package main

// monitorRoutes is the statically supported monitor API surface. Keep it in
// sync with handleRequest and the SST monitorHandler registrations.
var monitorRoutes = []routeDefinition{
	{"GET", "/api/v1/search", "searchResources"},
	{"POST", "/api/v1/notification-channels", "createNotificationChannel"}, {"GET", "/api/v1/notification-channels", "listNotificationChannels"},
	{"GET", "/api/v1/notification-channels/{channelId}", "getNotificationChannel"}, {"PUT", "/api/v1/notification-channels/{channelId}", "updateNotificationChannel"}, {"DELETE", "/api/v1/notification-channels/{channelId}", "deleteNotificationChannel"}, {"POST", "/api/v1/notification-channels/{channelId}/test", "testNotificationChannel"},
	{"POST", "/api/v1/escalation-policies", "createEscalationPolicy"}, {"GET", "/api/v1/escalation-policies", "listEscalationPolicies"}, {"GET", "/api/v1/escalation-policies/{policyId}", "getEscalationPolicy"}, {"PUT", "/api/v1/escalation-policies/{policyId}", "updateEscalationPolicy"}, {"DELETE", "/api/v1/escalation-policies/{policyId}", "deleteEscalationPolicy"},
	{"POST", "/api/v1/services", "createService"}, {"GET", "/api/v1/services", "listServices"}, {"GET", "/api/v1/services/{serviceId}", "getService"}, {"PATCH", "/api/v1/services/{serviceId}", "updateService"}, {"DELETE", "/api/v1/services/{serviceId}", "deleteService"}, {"POST", "/api/v1/services/{serviceId}/archive", "archiveService"}, {"POST", "/api/v1/services/{serviceId}/reactivate", "reactivateService"}, {"GET", "/api/v1/services/{serviceId}/escalation-policy", "getServiceEscalationPolicy"}, {"GET", "/api/v1/services/{serviceId}/audit", "getServiceAudit"},
	{"POST", "/api/v1/services/{serviceId}/monitors", "createMonitor"}, {"GET", "/api/v1/services/{serviceId}/monitors", "listMonitors"}, {"GET", "/api/v1/services/{serviceId}/monitors/{monitorId}", "getMonitor"}, {"PATCH", "/api/v1/services/{serviceId}/monitors/{monitorId}", "updateMonitor"}, {"DELETE", "/api/v1/services/{serviceId}/monitors/{monitorId}", "deleteMonitor"}, {"GET", "/api/v1/services/{serviceId}/monitors/{monitorId}/status", "getMonitorStatus"}, {"GET", "/api/v1/services/{serviceId}/monitors/{monitorId}/runs", "getMonitorRuns"}, {"POST", "/api/v1/services/{serviceId}/monitors/{monitorId}/run", "runMonitor"}, {"GET", "/api/v1/services/{serviceId}/monitors/{monitorId}/incidents", "getMonitorIncidents"}, {"GET", "/api/v1/services/{serviceId}/incidents", "getServiceIncidents"}, {"GET", "/api/v1/services/{serviceId}/monitors/{monitorId}/audit", "getMonitorAudit"}, {"POST", "/api/v1/services/{serviceId}/monitors/{monitorId}/enable", "setMonitorEnabled"}, {"POST", "/api/v1/services/{serviceId}/monitors/{monitorId}/disable", "setMonitorEnabled"}, {"POST", "/api/v1/services/{serviceId}/monitors/{monitorId}/maintenance/enable", "setMonitorMaintenance"}, {"POST", "/api/v1/services/{serviceId}/monitors/{monitorId}/maintenance/disable", "setMonitorMaintenance"},
	{"GET", "/api/v1/incidents", "listIncidents"}, {"GET", "/api/v1/incidents/{incidentId}", "getIncident"}, {"GET", "/api/v1/incidents/{incidentId}/escalation-state", "getEscalationState"}, {"GET", "/api/v1/incidents/{incidentId}/activities", "getIncidentActivities"}, {"GET", "/api/v1/incidents/{incidentId}/deliveries", "listIncidentDeliveries"}, {"POST", "/api/v1/incidents/{incidentId}/deliveries/{deliveryId}/replay", "replayIncidentDelivery"}, {"POST", "/api/v1/incidents/{incidentId}/ack", "acknowledgeIncident"}, {"POST", "/api/v1/incidents/{incidentId}/resolve", "resolveIncident"},
	{"GET", "/api/v1/admin/scheduler-config", "getSchedulerConfig"}, {"PATCH", "/api/v1/admin/scheduler-config", "updateSchedulerConfig"},
}

type routeDefinition struct {
	Method  string
	Path    string
	Handler string
}

var _ = monitorRoutes
