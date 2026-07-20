package dynamodbrecord

import (
	"strings"

	"bolt-monitor/shared/dynamodbschema"
	"bolt-monitor/shared/notifications"
)

type EscalationPlanItemRecord struct {
	PK           string   `dynamodbav:"PK"`
	SK           string   `dynamodbav:"SK"`
	EntityType   string   `dynamodbav:"EntityType"`
	TenantID     string   `dynamodbav:"TenantID"`
	IncidentID   string   `dynamodbav:"IncidentID"`
	TransitionID string   `dynamodbav:"TransitionID"`
	PolicyID     string   `dynamodbav:"PolicyID"`
	SelectedPath string   `dynamodbav:"SelectedPath"`
	StepNumbers  []int    `dynamodbav:"StepNumbers"`
	StepChannels []string `dynamodbav:"StepChannels"`
	CreatedAt    string   `dynamodbav:"CreatedAt"`
}

func NewEscalationPlanItemRecord(plan notifications.EscalationPlan) EscalationPlanItemRecord {
	item := dynamodbschema.EscalationPlanItem(plan.TenantID, plan.IncidentID, plan.TransitionID)
	return EscalationPlanItemRecord{
		PK:           item.PK,
		SK:           item.SK,
		EntityType:   item.EntityType,
		TenantID:     plan.TenantID,
		IncidentID:   plan.IncidentID,
		TransitionID: plan.TransitionID,
		PolicyID:     plan.PolicyID,
		SelectedPath: plan.SelectedPath,
		StepNumbers:  plan.StepNumbers,
		StepChannels: plan.StepChannels,
		CreatedAt:    plan.CreatedAt,
	}
}

func (r EscalationPlanItemRecord) ToEscalationPlan() notifications.EscalationPlan {
	return notifications.EscalationPlan{
		TenantID:     r.TenantID,
		IncidentID:   r.IncidentID,
		TransitionID: r.TransitionID,
		PolicyID:     r.PolicyID,
		SelectedPath: r.SelectedPath,
		StepNumbers:  r.StepNumbers,
		StepChannels: r.StepChannels,
		CreatedAt:    r.CreatedAt,
	}
}

func (r EscalationPlanItemRecord) BelongsToTenant(tenantID, incidentID string) bool {
	return strings.EqualFold(r.TenantID, tenantID) && strings.EqualFold(r.IncidentID, incidentID)
}
