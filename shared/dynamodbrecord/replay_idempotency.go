package dynamodbrecord

import (
	"fmt"
	"strings"

	"bolt-monitor/shared/dynamodbschema"
	"bolt-monitor/shared/notifications"
)

type ReplayIdempotencyItemRecord struct {
	PK                 string `dynamodbav:"PK"`
	SK                 string `dynamodbav:"SK"`
	EntityType         string `dynamodbav:"EntityType"`
	TenantID           string `dynamodbav:"TenantID"`
	IncidentID         string `dynamodbav:"IncidentID"`
	DeliveryID         string `dynamodbav:"DeliveryID"`
	Operation          string `dynamodbav:"Operation"`
	IdempotencyKey     string `dynamodbav:"IdempotencyKey"`
	RequestFingerprint string `dynamodbav:"RequestFingerprint"`
	ResultDeliveryID   string `dynamodbav:"ResultDeliveryID"`
	CreatedAt          string `dynamodbav:"CreatedAt"`
	ExpiresAt          int64  `dynamodbav:"ExpiresAt"`
}

func ReplayIdempotencyAddress(tenantID, incidentID, deliveryID, idempotencyKey string) string {
	digest := notifications.ReplayKeyFingerprint(tenantID, incidentID, deliveryID, idempotencyKey)
	return digest
}

func NewReplayIdempotencyItemRecord(record notifications.ReplayIdempotencyRecord) ReplayIdempotencyItemRecord {
	address := ReplayIdempotencyAddress(record.TenantID, record.IncidentID, record.DeliveryID, record.IdempotencyKey)
	item := dynamodbschema.ReplayIdempotencyItem(record.TenantID, address)
	return ReplayIdempotencyItemRecord{
		PK:                 item.PK,
		SK:                 item.SK,
		EntityType:         item.EntityType,
		TenantID:           record.TenantID,
		IncidentID:         record.IncidentID,
		DeliveryID:         record.DeliveryID,
		Operation:          record.Operation,
		IdempotencyKey:     record.IdempotencyKey,
		RequestFingerprint: record.RequestFingerprint,
		ResultDeliveryID:   record.ResultDeliveryID,
		CreatedAt:          record.CreatedAt,
		ExpiresAt:          record.ExpiresAt,
	}
}

func (r ReplayIdempotencyItemRecord) ToRecord() notifications.ReplayIdempotencyRecord {
	return notifications.ReplayIdempotencyRecord{
		TenantID:           r.TenantID,
		IncidentID:         r.IncidentID,
		DeliveryID:         r.DeliveryID,
		Operation:          r.Operation,
		IdempotencyKey:     r.IdempotencyKey,
		RequestFingerprint: r.RequestFingerprint,
		ResultDeliveryID:   r.ResultDeliveryID,
		CreatedAt:          r.CreatedAt,
		ExpiresAt:          r.ExpiresAt,
	}
}

func (r ReplayIdempotencyItemRecord) MatchesFingerprint(fingerprint string) bool {
	return strings.EqualFold(r.RequestFingerprint, fingerprint)
}

func ValidateReplayIdempotencyKey(key string) error {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return fmt.Errorf("idempotency key is required")
	}
	if len(trimmed) > 200 {
		return fmt.Errorf("idempotency key must be at most 200 characters")
	}
	return nil
}
