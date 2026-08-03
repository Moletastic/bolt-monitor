package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bolt-monitor/shared/api/response"
	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/notifications"
	"github.com/aws/aws-lambda-go/events"
)

const (
	deliveryReplayRetention = 24 * time.Hour
	defaultDeliveryPageSize = int32(50)
	maxDeliveryPageSize     = int32(200)
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

func (h monitorHandler) listIncidentDeliveries(ctx context.Context, incidentID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	limit := defaultDeliveryPageSize
	if raw := strings.TrimSpace(request.QueryStringParameters["limit"]); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || parsed < 1 || parsed > int64(maxDeliveryPageSize) {
			return respondAPIGateway(sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "limit", "reason": "must be between 1 and 200"}))
		}
		limit = int32(parsed)
	}
	resource := "incident-deliveries#" + incidentID
	startKey, err := historyStartKey(request, resource, "PK")
	if err != nil {
		return respondAPIGateway(err)
	}
	records, found, err := h.operations.incidents.deliveries.Execute(ctx, h.tenantID, incidentID, limit, startKey)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeIncidentNotFound, map[string]any{"incidentId": incidentID}))
	}
	views := make([]deliveryView, 0, len(records.Items))
	for _, record := range records.Items {
		views = append(views, toDeliveryView(record))
	}
	nextCursor, err := encodeHistoryCursor(resource, records.NextKey)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.OkCursorPaginated(map[string]any{"incidentId": incidentID, "deliveries": views}, len(views), nextCursor))
}

type deliveryReplayResponse struct {
	IncidentID   string                      `json:"incidentId"`
	DeliveryID   string                      `json:"deliveryId"`
	ReplayResult string                      `json:"replayResult"`
	State        notifications.DeliveryState `json:"state"`
}

func (h monitorHandler) replayIncidentDelivery(ctx context.Context, incidentID, deliveryID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	idempotencyKey := strings.TrimSpace(requestHeader(request.Headers, "Idempotency-Key"))
	if idempotencyKey == "" {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "idempotencyKey", "reason": "required"}))
	}
	if len(idempotencyKey) > 200 {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "idempotencyKey", "reason": "must be at most 200 characters"}))
	}
	fingerprint := notifications.ReplayKeyFingerprint(h.tenantID, incidentID, deliveryID, idempotencyKey) + ":" + fingerprintOfRequest(request)
	result, err := h.operations.incidents.replayDelivery.Execute(ctx, replayIncidentDeliveryInput{
		TenantID: h.tenantID, IncidentID: incidentID, DeliveryID: deliveryID, IdempotencyKey: idempotencyKey, RequestFingerprint: fingerprint,
	})
	if err != nil {
		return respondAPIGateway(err)
	}
	if result.Queued {
		return envelopeResponse(http.StatusOK, response.Ok(deliveryReplayResponse{IncidentID: incidentID, DeliveryID: deliveryID, ReplayResult: "queued", State: notifications.DeliveryPending}, "delivery replay queued"))
	}
	return envelopeResponse(http.StatusOK, response.Ok(deliveryReplayResponse{IncidentID: incidentID, DeliveryID: deliveryID, ReplayResult: "replayed", State: result.Delivery.State}, "delivery replay acknowledged"))
}

func fingerprintOfRequest(request events.APIGatewayV2HTTPRequest) string {
	h := sha256.New()
	h.Write([]byte(requestHeader(request.Headers, "Idempotency-Key")))
	h.Write([]byte{0})
	h.Write([]byte(strings.TrimSpace(request.Body)))
	return hex.EncodeToString(h.Sum(nil))[:16]
}
