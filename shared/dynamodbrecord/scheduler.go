package dynamodbrecord

import (
	"time"

	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbschema"
)

const SchedulerConfigEntityType = "SchedulerConfig"

type SchedulerConfigRecord struct {
	Config    checkexecution.SchedulerConfig
	UpdatedAt string
}

type SchedulerConfigItemRecord struct {
	PK               string `dynamodbav:"PK"`
	SK               string `dynamodbav:"SK"`
	EntityType       string `dynamodbav:"EntityType"`
	TenantID         string `dynamodbav:"TenantID"`
	RecurringEnabled bool   `dynamodbav:"RecurringEnabled"`
	StopControlMode  string `dynamodbav:"StopControlMode,omitempty"`
	UpdatedAt        string `dynamodbav:"UpdatedAt"`
}

func NewSchedulerConfigItemRecord(tenantID string, config checkexecution.SchedulerConfig, now time.Time) SchedulerConfigItemRecord {
	return SchedulerConfigItemRecord{
		PK:               dynamodbschema.TenantPK(tenantID),
		SK:               "SCHEDULER_CONFIG",
		EntityType:       SchedulerConfigEntityType,
		TenantID:         dynamodbschema.NormalizeToken(tenantID),
		RecurringEnabled: config.RecurringEnabled,
		StopControlMode:  string(config.StopControlMode),
		UpdatedAt:        now.UTC().Format(time.RFC3339),
	}
}

func (r SchedulerConfigItemRecord) ToSchedulerConfig() SchedulerConfigRecord {
	return SchedulerConfigRecord{
		Config: checkexecution.SchedulerConfig{
			RecurringEnabled: r.RecurringEnabled,
			StopControlMode:  checkexecution.StopControlMode(r.StopControlMode),
		},
		UpdatedAt: r.UpdatedAt,
	}
}
