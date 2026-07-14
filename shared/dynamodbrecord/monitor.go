package dynamodbrecord

import (
	"bolt-monitor/shared/dynamodbschema"
	"bolt-monitor/shared/monitorconfig"
)

type MonitorItemRecord struct {
	PK                string                           `dynamodbav:"PK"`
	SK                string                           `dynamodbav:"SK"`
	EntityType        string                           `dynamodbav:"EntityType"`
	TenantID          string                           `dynamodbav:"TenantID"`
	ServiceID         string                           `dynamodbav:"ServiceID"`
	MonitorID         string                           `dynamodbav:"MonitorID"`
	Name              string                           `dynamodbav:"Name"`
	Type              monitorconfig.MonitorType        `dynamodbav:"Type"`
	IntervalSeconds   int                              `dynamodbav:"IntervalSeconds"`
	Enabled           bool                             `dynamodbav:"Enabled"`
	FailureThreshold  int                              `dynamodbav:"FailureThreshold"`
	RecoveryThreshold int                              `dynamodbav:"RecoveryThreshold"`
	HTTP              *monitorconfig.HTTPConfiguration `dynamodbav:"HTTP,omitempty"`
	LastExecutionAt   string                           `dynamodbav:"LastExecutionAt,omitempty"`
}

func NewMonitorItemRecord(monitor monitorconfig.Monitor) MonitorItemRecord {
	return MonitorItemRecord{
		PK:                dynamodbschema.MonitorPK(monitor.TenantID, monitor.ServiceID, monitor.MonitorID),
		SK:                "META",
		EntityType:        dynamodbschema.EntityMonitor,
		TenantID:          dynamodbschema.NormalizeToken(monitor.TenantID),
		ServiceID:         dynamodbschema.NormalizeField(monitor.ServiceID),
		MonitorID:         dynamodbschema.NormalizeField(monitor.MonitorID),
		Name:              monitor.Name,
		Type:              monitor.Type,
		IntervalSeconds:   monitor.IntervalSeconds,
		Enabled:           monitor.Enabled,
		FailureThreshold:  monitor.FailureThreshold,
		RecoveryThreshold: monitor.RecoveryThreshold,
		HTTP:              CloneHTTPConfiguration(monitor.HTTP),
	}
}

func NewServiceMonitorRefItemRecord(monitor monitorconfig.Monitor) MonitorItemRecord {
	record := NewMonitorItemRecord(monitor)
	record.PK = dynamodbschema.ServicePK(monitor.TenantID, monitor.ServiceID)
	record.SK = dynamodbschema.ServiceMonitorRefSK(monitor.MonitorID)
	record.EntityType = dynamodbschema.EntityServiceMonitorRef
	return record
}

func (r MonitorItemRecord) ToMonitor() monitorconfig.Monitor {
	failureThreshold := r.FailureThreshold
	if failureThreshold < 1 {
		failureThreshold = 1
	}
	recoveryThreshold := r.RecoveryThreshold
	if recoveryThreshold < 1 {
		recoveryThreshold = 1
	}
	return monitorconfig.Monitor{
		ServiceID:         r.ServiceID,
		MonitorID:         r.MonitorID,
		TenantID:          r.TenantID,
		Name:              r.Name,
		Type:              r.Type,
		IntervalSeconds:   r.IntervalSeconds,
		Enabled:           r.Enabled,
		FailureThreshold:  failureThreshold,
		RecoveryThreshold: recoveryThreshold,
		HTTP:              CloneHTTPConfiguration(r.HTTP),
	}
}

func CloneHTTPConfiguration(input *monitorconfig.HTTPConfiguration) *monitorconfig.HTTPConfiguration {
	if input == nil {
		return nil
	}
	headers := make(map[string]string, len(input.Headers))
	for key, value := range input.Headers {
		headers[key] = value
	}
	statusCodes := append([]int(nil), input.ExpectedStatusCodes...)
	var expectedBodyContains *string
	if input.ExpectedBodyContains != nil {
		value := *input.ExpectedBodyContains
		expectedBodyContains = &value
	}
	return &monitorconfig.HTTPConfiguration{
		Target:               input.Target,
		Method:               input.Method,
		Headers:              headers,
		TimeoutMs:            input.TimeoutMs,
		ExpectedStatusCodes:  statusCodes,
		ExpectedBodyContains: expectedBodyContains,
	}
}
