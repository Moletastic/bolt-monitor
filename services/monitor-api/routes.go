package main

// monitorRoutes is the statically supported monitor API surface. Keep it in
// sync with handleRequest and the SST monitorHandler registrations.
var monitorRoutes = []routeDefinition{
	{"GET", "/api/v1/search"},
	{"POST", "/api/v1/notification-channels"}, {"GET", "/api/v1/notification-channels"},
	{"GET", "/api/v1/notification-channels/{channelId}"}, {"PUT", "/api/v1/notification-channels/{channelId}"}, {"DELETE", "/api/v1/notification-channels/{channelId}"}, {"POST", "/api/v1/notification-channels/{channelId}/test"},
	{"POST", "/api/v1/escalation-policies"}, {"GET", "/api/v1/escalation-policies"}, {"GET", "/api/v1/escalation-policies/{policyId}"}, {"PUT", "/api/v1/escalation-policies/{policyId}"}, {"DELETE", "/api/v1/escalation-policies/{policyId}"},
	{"POST", "/api/v1/services"}, {"GET", "/api/v1/services"}, {"GET", "/api/v1/services/{serviceId}"}, {"PATCH", "/api/v1/services/{serviceId}"}, {"DELETE", "/api/v1/services/{serviceId}"}, {"POST", "/api/v1/services/{serviceId}/archive"}, {"POST", "/api/v1/services/{serviceId}/reactivate"}, {"GET", "/api/v1/services/{serviceId}/escalation-policy"}, {"GET", "/api/v1/services/{serviceId}/audit"},
	{"POST", "/api/v1/services/{serviceId}/monitors"}, {"GET", "/api/v1/services/{serviceId}/monitors"}, {"GET", "/api/v1/services/{serviceId}/monitors/{monitorId}"}, {"PATCH", "/api/v1/services/{serviceId}/monitors/{monitorId}"}, {"DELETE", "/api/v1/services/{serviceId}/monitors/{monitorId}"}, {"GET", "/api/v1/services/{serviceId}/monitors/{monitorId}/status"}, {"GET", "/api/v1/services/{serviceId}/monitors/{monitorId}/runs"}, {"POST", "/api/v1/services/{serviceId}/monitors/{monitorId}/run"}, {"GET", "/api/v1/services/{serviceId}/monitors/{monitorId}/incidents"}, {"GET", "/api/v1/services/{serviceId}/incidents"}, {"GET", "/api/v1/services/{serviceId}/monitors/{monitorId}/audit"}, {"POST", "/api/v1/services/{serviceId}/monitors/{monitorId}/enable"}, {"POST", "/api/v1/services/{serviceId}/monitors/{monitorId}/disable"}, {"POST", "/api/v1/services/{serviceId}/monitors/{monitorId}/maintenance/enable"}, {"POST", "/api/v1/services/{serviceId}/monitors/{monitorId}/maintenance/disable"},
	{"GET", "/api/v1/incidents"}, {"GET", "/api/v1/incidents/{incidentId}"}, {"GET", "/api/v1/incidents/{incidentId}/escalation-state"}, {"GET", "/api/v1/incidents/{incidentId}/activities"}, {"POST", "/api/v1/incidents/{incidentId}/ack"}, {"POST", "/api/v1/incidents/{incidentId}/resolve"},
	{"GET", "/api/v1/admin/scheduler-config"}, {"PATCH", "/api/v1/admin/scheduler-config"},
}

type routeDefinition struct {
	Method string
	Path   string
}

var _ = monitorRoutes
