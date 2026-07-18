package domainvalues

import (
	"strings"

	sharederrors "bolt-monitor/shared/errors"
)

// MonitorRef is the composite identity used by monitor-scoped status, key,
// run, incident, audit, and storage operations. It eliminates the three
// primitive string arguments at call sites that require all three identifiers.
type MonitorRef struct {
	Tenant  TenantID
	Service ServiceID
	Monitor MonitorID
}

// NewMonitorRef validates that all three identifiers are present and returns
// a populated reference. The constructor never mutates the inputs.
func NewMonitorRef(tenant TenantID, service ServiceID, monitor MonitorID) (MonitorRef, error) {
	if tenant == "" {
		return MonitorRef{}, sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "tenantId", "reason": "required"})
	}
	if service == "" {
		return MonitorRef{}, sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "serviceId", "reason": "required"})
	}
	if monitor == "" {
		return MonitorRef{}, sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "monitorId", "reason": "required"})
	}
	return MonitorRef{Tenant: tenant, Service: service, Monitor: monitor}, nil
}

// MustMonitorRef returns the reference without validation. Callers must have
// already validated the inputs.
func MustMonitorRef(tenant TenantID, service ServiceID, monitor MonitorID) MonitorRef {
	return MonitorRef{Tenant: tenant, Service: service, Monitor: monitor}
}

// String returns the canonical "<tenant>/<service>/<monitor>" representation.
func (r MonitorRef) String() string {
	return string(r.Tenant) + "/" + string(r.Service) + "/" + string(r.Monitor)
}

// StatusMapKey returns the canonical "<service>/<monitor>" key used by
// status-map caches. Replacing the previous primitive helper with this method
// locks the format in one place.
func (r MonitorRef) StatusMapKey() string {
	return string(r.Service) + "/" + string(r.Monitor)
}

// PartitionKey returns the DynamoDB partition key for the monitor's
// service-scoped items. The PK uses the upper-cased token form because the
// canonical schema normalizes service and monitor slugs into tokens for key
// construction.
func (r MonitorRef) PartitionKey() string {
	return "MONITOR#" + string(r.Tenant) + "#" + tokenize(string(r.Service)) + "#" + tokenize(string(r.Monitor))
}

// ServicePartitionKey returns the DynamoDB partition key for the monitor's
// parent service partition.
func (r MonitorRef) ServicePartitionKey() string {
	return "SERVICE#" + string(r.Tenant) + "#" + tokenize(string(r.Service))
}

// TenantPartitionKey returns the DynamoDB partition key for the tenant.
func (r MonitorRef) TenantPartitionKey() string {
	return "TENANT#" + string(r.Tenant)
}

// ServiceMonitorRefSK returns the DynamoDB sort key for the service-to-monitor
// reference item.
func (r MonitorRef) ServiceMonitorRefSK() string {
	return "MONITOR#" + tokenize(string(r.Monitor))
}

// tokenize mirrors the existing dynamodbschema.normalizeToken contract for key
// construction. MonitorRef.PartitionKey and friends must produce the same
// strings as the schema constructors so existing items remain reachable.
func tokenize(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

// MetaSK returns the canonical META sort key for the monitor's primary item.
func (r MonitorRef) MetaSK() string { return "META" }

// StatusSK returns the canonical STATUS sort key for the monitor status item.
func (r MonitorRef) StatusSK() string { return "STATUS" }

// RunPrefix returns the canonical SK prefix for monitor run history items.
func (r MonitorRef) RunPrefix() string { return "RUN#" }

// IncidentPrefix returns the canonical SK prefix for monitor-scoped incident
// items.
func (r MonitorRef) IncidentPrefix() string { return "INCIDENT#" }
