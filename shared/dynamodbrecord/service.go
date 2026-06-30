package dynamodbrecord

import (
	"sort"
	"strings"

	"bolt-monitor/shared/dynamodbschema"
	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/monitorconfig"
)

type ServiceItemRecord struct {
	PK                 string                          `dynamodbav:"PK"`
	SK                 string                          `dynamodbav:"SK"`
	EntityType         string                          `dynamodbav:"EntityType"`
	TenantID           string                          `dynamodbav:"TenantID"`
	ServiceID          string                          `dynamodbav:"ServiceID"`
	Name               string                          `dynamodbav:"Name"`
	Description        string                          `dynamodbav:"Description,omitempty"`
	LifecycleState     string                          `dynamodbav:"LifecycleState"`
	TechnologyKey      string                          `dynamodbav:"TechnologyKey,omitempty"`
	EscalationPolicyID string                          `dynamodbav:"EscalationPolicyID,omitempty"`
	BusinessHours      *escalation.BusinessHoursConfig `dynamodbav:"BusinessHours,omitempty"`
	CreatedAt          string                          `dynamodbav:"CreatedAt,omitempty"`
	UpdatedAt          string                          `dynamodbav:"UpdatedAt,omitempty"`
}

func NewServiceItemRecord(service monitorconfig.Service) ServiceItemRecord {
	return ServiceItemRecord{
		PK:                 dynamodbschema.ServicePK(service.TenantID, service.ServiceID),
		SK:                 "META",
		EntityType:         dynamodbschema.EntityService,
		TenantID:           dynamodbschema.NormalizeToken(service.TenantID),
		ServiceID:          dynamodbschema.NormalizeField(service.ServiceID),
		Name:               service.Name,
		Description:        service.Description,
		LifecycleState:     string(service.LifecycleState),
		TechnologyKey:      service.TechnologyKey,
		EscalationPolicyID: strings.TrimSpace(service.EscalationPolicyID),
		BusinessHours:      CloneBusinessHoursConfig(service.BusinessHours),
		CreatedAt:          service.CreatedAt,
		UpdatedAt:          service.UpdatedAt,
	}
}

func NewServiceRefItemRecord(service monitorconfig.Service) ServiceItemRecord {
	record := NewServiceItemRecord(service)
	record.PK = dynamodbschema.TenantPK(service.TenantID)
	record.SK = dynamodbschema.ServiceRefSK(service.ServiceID)
	record.EntityType = dynamodbschema.EntityServiceRef
	return record
}

func (r ServiceItemRecord) ToService() monitorconfig.Service {
	return monitorconfig.Service{
		TenantID:           r.TenantID,
		ServiceID:          r.ServiceID,
		Name:               r.Name,
		Description:        r.Description,
		LifecycleState:     monitorconfig.ServiceLifecycle(r.LifecycleState),
		TechnologyKey:      r.TechnologyKey,
		EscalationPolicyID: strings.TrimSpace(r.EscalationPolicyID),
		BusinessHours:      CloneBusinessHoursConfig(r.BusinessHours),
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
	}
}

type ServiceStatusRecord struct {
	PK                  string `dynamodbav:"PK"`
	SK                  string `dynamodbav:"SK"`
	EntityType          string `dynamodbav:"EntityType"`
	TenantID            string `dynamodbav:"TenantID"`
	ServiceID           string `dynamodbav:"ServiceID"`
	LifecycleState      string `dynamodbav:"LifecycleState"`
	RollupStatus        string `dynamodbav:"RollupStatus"`
	MonitorCount        int    `dynamodbav:"MonitorCount"`
	EnabledMonitorCount int    `dynamodbav:"EnabledMonitorCount"`
	UpdatedAt           string `dynamodbav:"UpdatedAt"`
	GSI2PK              string `dynamodbav:"GSI2PK,omitempty"`
	GSI2SK              string `dynamodbav:"GSI2SK,omitempty"`
}

func NewServiceStatusItemRecord(service monitorconfig.Service, updatedAt string) ServiceStatusRecord {
	item := dynamodbschema.ServiceStatusItem(service.TenantID, service.ServiceID, service.RollupStatus, updatedAt)
	return ServiceStatusRecord{
		PK:                  item.PK,
		SK:                  item.SK,
		EntityType:          item.EntityType,
		TenantID:            dynamodbschema.NormalizeToken(service.TenantID),
		ServiceID:           dynamodbschema.NormalizeField(service.ServiceID),
		LifecycleState:      string(service.LifecycleState),
		RollupStatus:        service.RollupStatus,
		MonitorCount:        service.MonitorCount,
		EnabledMonitorCount: service.EnabledCount,
		UpdatedAt:           updatedAt,
		GSI2PK:              item.GSI2PK,
		GSI2SK:              item.GSI2SK,
	}
}

func CloneBusinessHoursConfig(input *escalation.BusinessHoursConfig) *escalation.BusinessHoursConfig {
	if input == nil {
		return nil
	}
	daysOfWeek := append([]int(nil), input.DaysOfWeek...)
	sort.Ints(daysOfWeek)
	return &escalation.BusinessHoursConfig{
		Timezone:   strings.TrimSpace(input.Timezone),
		StartHour:  input.StartHour,
		EndHour:    input.EndHour,
		DaysOfWeek: daysOfWeek,
	}
}
