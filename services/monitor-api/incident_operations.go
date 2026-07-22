package main

import (
	"context"
	"strings"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbrecord"
	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/notifications"
)

type listIncidentsStore interface {
	ListIncidents(context.Context, string, string) ([]dynamodbrecord.IncidentRecord, error)
}

type incidentLookup interface {
	GetIncident(context.Context, string, string) (dynamodbrecord.IncidentRecord, bool, error)
}

type incidentActivitiesStore interface {
	incidentLookup
	ListIncidentActivities(context.Context, string, string) ([]dynamodbrecord.IncidentActivityRecord, error)
}

type incidentEscalationStateStore interface {
	incidentLookup
	GetEscalationState(context.Context, string, string) (*escalation.EscalationState, error)
}

type monitorIncidentsStore interface {
	GetMonitor(context.Context, string, string, string) (monitorconfig.Monitor, bool, error)
	ListMonitorIncidentsPage(context.Context, string, string, string, int32, map[string]sharedaws.AttributeValue) (historyPage[dynamodbrecord.IncidentRecord], error)
}

type serviceIncidentsStore interface {
	GetService(context.Context, string, string) (monitorconfig.Service, bool, error)
	ListServiceIncidents(context.Context, string, string, int32) ([]dynamodbrecord.IncidentRecord, error)
}

type acknowledgeIncidentStore interface {
	AcknowledgeIncident(context.Context, string, string, time.Time) (dynamodbrecord.IncidentRecord, bool, error)
}

type resolveIncidentStore interface {
	ResolveIncident(context.Context, string, string, time.Time) (dynamodbrecord.IncidentRecord, bool, error)
}

type listIncidentDeliveriesStore interface {
	ListIncidentDeliveries(context.Context, string, string) ([]notifications.DeliveryRecord, error)
}

type replayIncidentDeliveryStore interface {
	listIncidentDeliveriesStore
	PrepareDeliveryReplay(context.Context, notifications.ReplayCommand, string, time.Time, time.Duration) (string, error)
	LookupReplayIdempotency(context.Context, string, string, string, string) (*notifications.ReplayIdempotencyRecord, error)
}

type listIncidentsQuery struct{ store listIncidentsStore }
type getIncidentQuery struct{ store incidentLookup }
type incidentActivitiesQuery struct{ store incidentActivitiesStore }
type incidentEscalationStateQuery struct{ store incidentEscalationStateStore }
type monitorIncidentsQuery struct{ store monitorIncidentsStore }
type serviceIncidentsQuery struct{ store serviceIncidentsStore }
type acknowledgeIncidentCommand struct {
	store acknowledgeIncidentStore
	now   commandClock
}
type resolveIncidentCommand struct {
	store resolveIncidentStore
	now   commandClock
}
type listIncidentDeliveriesQuery struct {
	incidents incidentLookup
	store     listIncidentDeliveriesStore
}
type replayIncidentDeliveryCommand struct {
	incidents incidentLookup
	store     replayIncidentDeliveryStore
	now       commandClock
}

type incidentOperations struct {
	list             listIncidentsQuery
	get              getIncidentQuery
	activities       incidentActivitiesQuery
	escalationState  incidentEscalationStateQuery
	monitorIncidents monitorIncidentsQuery
	serviceIncidents serviceIncidentsQuery
	acknowledge      acknowledgeIncidentCommand
	resolve          resolveIncidentCommand
	deliveries       listIncidentDeliveriesQuery
	replayDelivery   replayIncidentDeliveryCommand
}

func newIncidentOperations(list listIncidentsStore, get incidentLookup, activities incidentActivitiesStore, escalationState incidentEscalationStateStore, monitorIncidents monitorIncidentsStore, serviceIncidents serviceIncidentsStore, acknowledge acknowledgeIncidentStore, resolve resolveIncidentStore, deliveries listIncidentDeliveriesStore, replay replayIncidentDeliveryStore, now commandClock) incidentOperations {
	return incidentOperations{
		list: listIncidentsQuery{store: list}, get: getIncidentQuery{store: get},
		activities: incidentActivitiesQuery{store: activities}, escalationState: incidentEscalationStateQuery{store: escalationState},
		monitorIncidents: monitorIncidentsQuery{store: monitorIncidents}, serviceIncidents: serviceIncidentsQuery{store: serviceIncidents},
		acknowledge: acknowledgeIncidentCommand{store: acknowledge, now: now}, resolve: resolveIncidentCommand{store: resolve, now: now},
		deliveries: listIncidentDeliveriesQuery{incidents: get, store: deliveries}, replayDelivery: replayIncidentDeliveryCommand{incidents: get, store: replay, now: now},
	}
}

func (q listIncidentsQuery) Execute(ctx context.Context, tenantID, status string) ([]dynamodbrecord.IncidentRecord, error) {
	return q.store.ListIncidents(ctx, tenantID, status)
}

func (q getIncidentQuery) Execute(ctx context.Context, tenantID, incidentID string) (dynamodbrecord.IncidentRecord, bool, error) {
	return q.store.GetIncident(ctx, tenantID, incidentID)
}

func (q incidentActivitiesQuery) Execute(ctx context.Context, tenantID, incidentID string) ([]dynamodbrecord.IncidentActivityRecord, bool, error) {
	if _, found, err := q.store.GetIncident(ctx, tenantID, incidentID); err != nil || !found {
		return nil, found, err
	}
	activities, err := q.store.ListIncidentActivities(ctx, tenantID, incidentID)
	return activities, true, err
}

func (q incidentEscalationStateQuery) Execute(ctx context.Context, tenantID, incidentID string) (*escalation.EscalationState, bool, error) {
	if _, found, err := q.store.GetIncident(ctx, tenantID, incidentID); err != nil || !found {
		return nil, found, err
	}
	state, err := q.store.GetEscalationState(ctx, tenantID, incidentID)
	return state, true, err
}

func (q monitorIncidentsQuery) Execute(ctx context.Context, tenantID, serviceID, monitorID string, limit int32, startKey map[string]sharedaws.AttributeValue) (historyPage[dynamodbrecord.IncidentRecord], bool, error) {
	if _, found, err := q.store.GetMonitor(ctx, tenantID, serviceID, monitorID); err != nil || !found {
		return historyPage[dynamodbrecord.IncidentRecord]{}, found, err
	}
	page, err := q.store.ListMonitorIncidentsPage(ctx, tenantID, serviceID, monitorID, limit, startKey)
	return page, true, err
}

func (q serviceIncidentsQuery) Execute(ctx context.Context, tenantID, serviceID string, limit int32) ([]dynamodbrecord.IncidentRecord, bool, error) {
	if _, found, err := q.store.GetService(ctx, tenantID, serviceID); err != nil || !found {
		return nil, found, err
	}
	incidents, err := q.store.ListServiceIncidents(ctx, tenantID, serviceID, limit)
	return incidents, true, err
}

func (c acknowledgeIncidentCommand) Execute(ctx context.Context, tenantID, incidentID string) (dynamodbrecord.IncidentRecord, bool, error) {
	return c.store.AcknowledgeIncident(ctx, tenantID, incidentID, c.now())
}

func (c resolveIncidentCommand) Execute(ctx context.Context, tenantID, incidentID string) (dynamodbrecord.IncidentRecord, bool, error) {
	return c.store.ResolveIncident(ctx, tenantID, incidentID, c.now())
}

func (q listIncidentDeliveriesQuery) Execute(ctx context.Context, tenantID, incidentID string) ([]notifications.DeliveryRecord, bool, error) {
	if _, found, err := q.incidents.GetIncident(ctx, tenantID, incidentID); err != nil || !found {
		return nil, found, err
	}
	deliveries, err := q.store.ListIncidentDeliveries(ctx, tenantID, incidentID)
	return deliveries, true, err
}

type replayIncidentDeliveryInput struct {
	TenantID           string
	IncidentID         string
	DeliveryID         string
	IdempotencyKey     string
	RequestFingerprint string
}

type replayIncidentDeliveryResult struct {
	Delivery notifications.DeliveryRecord
	Queued   bool
}

func (c replayIncidentDeliveryCommand) Execute(ctx context.Context, input replayIncidentDeliveryInput) (replayIncidentDeliveryResult, error) {
	if _, found, err := c.incidents.GetIncident(ctx, input.TenantID, input.IncidentID); err != nil {
		return replayIncidentDeliveryResult{}, err
	} else if !found {
		return replayIncidentDeliveryResult{}, sharederrors.New(sharederrors.CodeIncidentNotFound, map[string]any{"incidentId": input.IncidentID})
	}
	deliveries, err := c.store.ListIncidentDeliveries(ctx, input.TenantID, input.IncidentID)
	if err != nil {
		return replayIncidentDeliveryResult{}, err
	}
	var delivery *notifications.DeliveryRecord
	for i := range deliveries {
		if deliveries[i].DeliveryID == input.DeliveryID {
			delivery = &deliveries[i]
			break
		}
	}
	if delivery == nil {
		return replayIncidentDeliveryResult{}, sharederrors.New(sharederrors.CodeDeliveryNotFound, map[string]any{"incidentId": input.IncidentID, "deliveryId": input.DeliveryID})
	}
	if existing, err := c.store.LookupReplayIdempotency(ctx, input.TenantID, input.IncidentID, input.DeliveryID, input.IdempotencyKey); err != nil {
		return replayIncidentDeliveryResult{}, err
	} else if existing != nil {
		if !strings.EqualFold(existing.RequestFingerprint, input.RequestFingerprint) {
			return replayIncidentDeliveryResult{}, sharederrors.New(sharederrors.CodeIdempotencyConflict, map[string]any{"incidentId": input.IncidentID, "deliveryId": input.DeliveryID})
		}
		return replayIncidentDeliveryResult{Delivery: *delivery}, nil
	}
	if delivery.State != notifications.DeliveryTerminalFailed {
		return replayIncidentDeliveryResult{}, sharederrors.New(sharederrors.CodeDeliveryNotReplayable, map[string]any{"incidentId": input.IncidentID, "deliveryId": input.DeliveryID, "state": string(delivery.State)})
	}
	now := c.now().UTC()
	command := notifications.ReplayCommand{TenantID: input.TenantID, IncidentID: input.IncidentID, TransitionID: delivery.TransitionID, DeliveryID: input.DeliveryID, IdempotencyKey: input.IdempotencyKey, RequestedAt: now.Format(time.RFC3339)}
	if _, err := c.store.PrepareDeliveryReplay(ctx, command, input.RequestFingerprint, now, deliveryReplayRetention); err != nil {
		return replayIncidentDeliveryResult{}, err
	}
	return replayIncidentDeliveryResult{Delivery: *delivery, Queued: true}, nil
}
