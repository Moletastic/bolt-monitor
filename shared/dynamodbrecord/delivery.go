package dynamodbrecord

import (
	"strings"

	"bolt-monitor/shared/dynamodbschema"
	"bolt-monitor/shared/notifications"
)

type DeliveryItemRecord struct {
	PK                  string                             `dynamodbav:"PK"`
	SK                  string                             `dynamodbav:"SK"`
	EntityType          string                             `dynamodbav:"EntityType"`
	TenantID            string                             `dynamodbav:"TenantID"`
	IncidentID          string                             `dynamodbav:"IncidentID"`
	TransitionID        string                             `dynamodbav:"TransitionID"`
	DeliveryID          string                             `dynamodbav:"DeliveryID"`
	ChannelID           string                             `dynamodbav:"ChannelID"`
	ChannelType         string                             `dynamodbav:"ChannelType"`
	StepNumber          int                                `dynamodbav:"StepNumber"`
	State               notifications.DeliveryState        `dynamodbav:"State"`
	AttemptCount        int                                `dynamodbav:"AttemptCount"`
	LastAttemptAt       string                             `dynamodbav:"LastAttemptAt,omitempty"`
	LeaseUntil          string                             `dynamodbav:"LeaseUntil,omitempty"`
	NextAttemptAt       string                             `dynamodbav:"NextAttemptAt,omitempty"`
	FencingToken        string                             `dynamodbav:"FencingToken,omitempty"`
	LastOutcomeClass    notifications.DeliveryOutcomeClass `dynamodbav:"LastOutcomeClass,omitempty"`
	ProviderStatusClass string                             `dynamodbav:"ProviderStatusClass,omitempty"`
	ProviderRequestID   string                             `dynamodbav:"ProviderRequestID,omitempty"`
	RetryAfterSeconds   int                                `dynamodbav:"RetryAfterSeconds,omitempty"`
	CreatedAt           string                             `dynamodbav:"CreatedAt"`
	UpdatedAt           string                             `dynamodbav:"UpdatedAt"`
}

func NewDeliveryItemRecord(delivery notifications.DeliveryRecord) DeliveryItemRecord {
	item := dynamodbschema.DeliveryItem(delivery.TenantID, delivery.IncidentID, delivery.CreatedAt, delivery.DeliveryID)
	return DeliveryItemRecord{
		PK:                  item.PK,
		SK:                  item.SK,
		EntityType:          item.EntityType,
		TenantID:            delivery.TenantID,
		IncidentID:          delivery.IncidentID,
		TransitionID:        delivery.TransitionID,
		DeliveryID:          delivery.DeliveryID,
		ChannelID:           delivery.ChannelID,
		ChannelType:         delivery.ChannelType,
		StepNumber:          delivery.StepNumber,
		State:               delivery.State,
		AttemptCount:        delivery.AttemptCount,
		LastAttemptAt:       delivery.LastAttemptAt,
		LeaseUntil:          delivery.LeaseUntil,
		NextAttemptAt:       delivery.NextAttemptAt,
		FencingToken:        delivery.FencingToken,
		LastOutcomeClass:    delivery.LastOutcomeClass,
		ProviderStatusClass: delivery.ProviderMetadata.ProviderStatusClass,
		ProviderRequestID:   delivery.ProviderMetadata.ProviderRequestID,
		RetryAfterSeconds:   delivery.ProviderMetadata.RetryAfterSeconds,
		CreatedAt:           delivery.CreatedAt,
		UpdatedAt:           delivery.UpdatedAt,
	}
}

func (r DeliveryItemRecord) ToDelivery() notifications.DeliveryRecord {
	return notifications.DeliveryRecord{
		TenantID:         r.TenantID,
		IncidentID:       r.IncidentID,
		TransitionID:     r.TransitionID,
		DeliveryID:       r.DeliveryID,
		ChannelID:        r.ChannelID,
		ChannelType:      r.ChannelType,
		StepNumber:       r.StepNumber,
		State:            r.State,
		AttemptCount:     r.AttemptCount,
		LastAttemptAt:    r.LastAttemptAt,
		LeaseUntil:       r.LeaseUntil,
		NextAttemptAt:    r.NextAttemptAt,
		FencingToken:     r.FencingToken,
		LastOutcomeClass: r.LastOutcomeClass,
		ProviderMetadata: notifications.ProviderMetadata{
			ProviderStatusClass: r.ProviderStatusClass,
			ProviderRequestID:   r.ProviderRequestID,
			RetryAfterSeconds:   r.RetryAfterSeconds,
		},
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

func (r DeliveryItemRecord) BelongsToTenant(tenantID, incidentID string) bool {
	return strings.EqualFold(r.TenantID, tenantID) && strings.EqualFold(r.IncidentID, incidentID)
}
