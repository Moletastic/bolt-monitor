package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"bolt-monitor/shared/api/response"
	sharedaws "bolt-monitor/shared/aws"
	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/notifications"
	"github.com/aws/aws-lambda-go/events"
)

const (
	deliveryReplayRetention = 24 * time.Hour
	maxDeliveryResults      = 200
)

type deliveryView struct {
	DeliveryID          string                             `json:"deliveryId"`
	TransitionID        string                             `json:"transitionId"`
	ChannelID           string                             `json:"channelId"`
	ChannelType         string                             `json:"channelType"`
	StepNumber          int                                `json:"stepNumber"`
	State               notifications.DeliveryState        `json:"state"`
	AttemptCount        int                                `json:"attemptCount"`
	LastAttemptAt       string                             `json:"lastAttemptAt,omitempty"`
	NextAttemptAt       string                             `json:"nextAttemptAt,omitempty"`
	LastOutcomeClass    notifications.DeliveryOutcomeClass `json:"lastOutcomeClass,omitempty"`
	ProviderStatusClass string                             `json:"providerStatusClass,omitempty"`
	ProviderRequestID   string                             `json:"providerRequestId,omitempty"`
	RetryAfterSeconds   int                                `json:"retryAfterSeconds,omitempty"`
	CreatedAt           string                             `json:"createdAt"`
	UpdatedAt           string                             `json:"updatedAt"`
}

func toDeliveryView(d notifications.DeliveryRecord) deliveryView {
	return deliveryView{
		DeliveryID:          d.DeliveryID,
		TransitionID:        d.TransitionID,
		ChannelID:           d.ChannelID,
		ChannelType:         d.ChannelType,
		StepNumber:          d.StepNumber,
		State:               d.State,
		AttemptCount:        d.AttemptCount,
		LastAttemptAt:       d.LastAttemptAt,
		NextAttemptAt:       d.NextAttemptAt,
		LastOutcomeClass:    d.LastOutcomeClass,
		ProviderStatusClass: d.ProviderMetadata.ProviderStatusClass,
		ProviderRequestID:   d.ProviderMetadata.ProviderRequestID,
		RetryAfterSeconds:   d.ProviderMetadata.RetryAfterSeconds,
		CreatedAt:           d.CreatedAt,
		UpdatedAt:           d.UpdatedAt,
	}
}

func (h monitorHandler) listIncidentDeliveries(ctx context.Context, incidentID string, _ events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if _, _, err := h.requireIncidentOwnership(ctx, incidentID); err != nil {
		return respondAPIGateway(err)
	}
	records, err := h.operations.deliveries.ListIncidentDeliveries(ctx, h.tenantID, incidentID)
	if err != nil {
		return respondAPIGateway(err)
	}
	views := make([]deliveryView, 0, len(records))
	for _, record := range records {
		views = append(views, toDeliveryView(record))
	}
	if len(views) > maxDeliveryResults {
		views = views[:maxDeliveryResults]
	}
	return envelopeResponse(http.StatusOK, response.Ok(map[string]any{"incidentId": incidentID, "deliveries": views}, "incident deliveries listed"))
}

type deliveryReplayResponse struct {
	IncidentID   string                      `json:"incidentId"`
	DeliveryID   string                      `json:"deliveryId"`
	ReplayResult string                      `json:"replayResult"`
	State        notifications.DeliveryState `json:"state"`
}

func (h monitorHandler) replayIncidentDelivery(ctx context.Context, incidentID, deliveryID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if _, _, err := h.requireIncidentOwnership(ctx, incidentID); err != nil {
		return respondAPIGateway(err)
	}
	idempotencyKey := strings.TrimSpace(request.Headers["Idempotency-Key"])
	if idempotencyKey == "" {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "idempotencyKey", "reason": "required"}))
	}
	if len(idempotencyKey) > 200 {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "idempotencyKey", "reason": "must be at most 200 characters"}))
	}
	deliveries, err := h.operations.deliveries.ListIncidentDeliveries(ctx, h.tenantID, incidentID)
	if err != nil {
		return respondAPIGateway(err)
	}
	var delivery *notifications.DeliveryRecord
	for i := range deliveries {
		if deliveries[i].DeliveryID == deliveryID {
			delivery = &deliveries[i]
			break
		}
	}
	if delivery == nil {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeDeliveryNotFound, map[string]any{"incidentId": incidentID, "deliveryId": deliveryID}))
	}
	fingerprint := notifications.ReplayKeyFingerprint(h.tenantID, incidentID, deliveryID, idempotencyKey) + ":" + fingerprintOfRequest(request)
	if existing, err := h.operations.deliveries.LookupReplayIdempotency(ctx, h.tenantID, incidentID, deliveryID, idempotencyKey); err != nil {
		return respondAPIGateway(err)
	} else if existing != nil {
		if !strings.EqualFold(existing.RequestFingerprint, fingerprint) {
			return respondAPIGateway(sharederrors.New(sharederrors.CodeIdempotencyConflict, map[string]any{"incidentId": incidentID, "deliveryId": deliveryID}))
		}
		return envelopeResponse(http.StatusOK, response.Ok(deliveryReplayResponse{IncidentID: incidentID, DeliveryID: deliveryID, ReplayResult: "replayed", State: delivery.State}, "delivery replay acknowledged"))
	}
	if delivery.State != notifications.DeliveryTerminalFailed {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeDeliveryNotReplayable, map[string]any{"incidentId": incidentID, "deliveryId": deliveryID, "state": string(delivery.State)}))
	}
	now := h.now().UTC()
	command := notifications.ReplayCommand{TenantID: h.tenantID, IncidentID: incidentID, TransitionID: delivery.TransitionID, DeliveryID: deliveryID, IdempotencyKey: idempotencyKey, RequestedAt: now.Format(time.RFC3339)}
	if _, err := h.operations.deliveries.PrepareDeliveryReplay(ctx, command, fingerprint, now, deliveryReplayRetention); err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(deliveryReplayResponse{IncidentID: incidentID, DeliveryID: deliveryID, ReplayResult: "queued", State: notifications.DeliveryPending}, "delivery replay queued"))
}

func (h monitorHandler) requireIncidentOwnership(ctx context.Context, incidentID string) (*notifications.DeliveryRecord, bool, error) {
	_, found, err := h.operations.GetIncident(ctx, h.tenantID, incidentID)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, sharederrors.New(sharederrors.CodeIncidentNotFound, map[string]any{"incidentId": incidentID})
	}
	return nil, true, nil
}

var _ sharedaws.AttributeValue // keep import in case future expansion needs the type

func fingerprintOfRequest(request events.APIGatewayV2HTTPRequest) string {
	h := sha256.New()
	h.Write([]byte(request.Headers["Idempotency-Key"]))
	h.Write([]byte{0})
	h.Write([]byte(strings.TrimSpace(request.Body)))
	return hex.EncodeToString(h.Sum(nil))[:16]
}
