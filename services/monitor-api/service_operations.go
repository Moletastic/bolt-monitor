package main

import (
	"context"
	"strings"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbrecord"
	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/resultstatus"
)

// serviceLookup is consumed by service-adjacent routes that only need to
// establish that a service exists. Lifecycle routes use the named operations
// below instead of this shared lookup port.
type serviceLookup interface {
	GetService(context.Context, string, string) (monitorconfig.Service, bool, error)
}

type servicePolicyReferenceQuery interface {
	ServiceReferencesEscalationPolicy(context.Context, string, string) (bool, error)
}

type createServiceStore interface {
	CreateService(context.Context, monitorconfig.Service) (monitorconfig.Service, error)
}

type updateServiceStore interface {
	GetService(context.Context, string, string) (monitorconfig.Service, bool, error)
	UpdateService(context.Context, monitorconfig.Service) (monitorconfig.Service, error)
}

type deleteServiceStore interface {
	DeleteService(context.Context, string, string) (bool, error)
}

type archiveServiceStore interface {
	ArchiveService(context.Context, string, string) (monitorconfig.Service, error)
}

type reactivateServiceStore interface {
	ReactivateService(context.Context, string, string) (monitorconfig.Service, error)
}

type listServicesStore interface {
	ListServices(context.Context, string) ([]monitorconfig.Service, error)
	GetServiceCardMetrics(context.Context, string, string) (serviceCardMetricsResponse, error)
}

type getServiceStore interface {
	GetService(context.Context, string, string) (monitorconfig.Service, bool, error)
	ListMonitors(context.Context, string, string) ([]monitorconfig.Monitor, error)
	GetMonitorStatus(context.Context, string, string, string) (resultstatus.MonitorStatus, bool, error)
	GetServiceCardMetrics(context.Context, string, string) (serviceCardMetricsResponse, error)
}

type serviceAuditStore interface {
	serviceLookup
	ListServiceAuditEventsPage(context.Context, string, string, int32, map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error)
}

type createServiceCommand struct {
	store createServiceStore
	now   commandClock
	ids   identifierGenerator
}

type updateServiceCommand struct{ store updateServiceStore }
type deleteServiceCommand struct{ store deleteServiceStore }
type archiveServiceCommand struct{ store archiveServiceStore }
type reactivateServiceCommand struct{ store reactivateServiceStore }
type listServicesQuery struct{ store listServicesStore }
type getServiceQuery struct{ store getServiceStore }
type serviceAuditQuery struct{ store serviceAuditStore }

type serviceOperations struct {
	create     createServiceCommand
	update     updateServiceCommand
	delete     deleteServiceCommand
	archive    archiveServiceCommand
	reactivate reactivateServiceCommand
	list       listServicesQuery
	get        getServiceQuery
	audit      serviceAuditQuery
}

func newServiceOperations(create createServiceStore, update updateServiceStore, delete deleteServiceStore, archive archiveServiceStore, reactivate reactivateServiceStore, list listServicesStore, get getServiceStore, audit serviceAuditStore, now commandClock, ids identifierGenerator) serviceOperations {
	return serviceOperations{
		create:     createServiceCommand{store: create, now: now, ids: ids},
		update:     updateServiceCommand{store: update},
		delete:     deleteServiceCommand{store: delete},
		archive:    archiveServiceCommand{store: archive},
		reactivate: reactivateServiceCommand{store: reactivate},
		list:       listServicesQuery{store: list},
		get:        getServiceQuery{store: get},
		audit:      serviceAuditQuery{store: audit},
	}
}

type createServiceInput struct {
	TenantID string
	Request  monitorconfig.CreateServiceRequest
}

func (c createServiceCommand) Execute(ctx context.Context, input createServiceInput) (monitorconfig.Service, error) {
	service, err := input.Request.ToService(input.TenantID)
	if err != nil {
		return monitorconfig.Service{}, err
	}
	service.ServiceID = c.ids.newServiceID(c.now())
	return c.store.CreateService(ctx, service)
}

type updateServiceInput struct {
	TenantID  string
	ServiceID string
	Request   updateServiceRequest
}

func (c updateServiceCommand) Execute(ctx context.Context, input updateServiceInput) (monitorconfig.Service, error) {
	current, found, err := c.store.GetService(ctx, input.TenantID, input.ServiceID)
	if err != nil {
		return monitorconfig.Service{}, err
	}
	if !found {
		return monitorconfig.Service{}, sharederrors.New(sharederrors.CodeServiceNotFound, nil)
	}
	if input.Request.ServiceID != nil && !strings.EqualFold(*input.Request.ServiceID, current.ServiceID) {
		return monitorconfig.Service{}, sharederrors.New(sharederrors.CodeImmutableField, map[string]any{"field": "serviceId"})
	}
	updated := current
	if input.Request.Name != nil {
		updated.Name = strings.TrimSpace(*input.Request.Name)
	}
	if input.Request.Description != nil {
		updated.Description = strings.TrimSpace(*input.Request.Description)
	}
	if input.Request.ServiceCategory != nil {
		updated.ServiceCategory = monitorconfig.ServiceCategory(strings.TrimSpace(*input.Request.ServiceCategory))
	}
	if input.Request.EscalationPolicyID != nil {
		updated.EscalationPolicyID = strings.TrimSpace(*input.Request.EscalationPolicyID)
	}
	if input.Request.BusinessHours != nil {
		updated.BusinessHours = dynamodbrecord.CloneBusinessHoursConfig(input.Request.BusinessHours)
	}
	if err := updated.Validate(); err != nil {
		return monitorconfig.Service{}, err
	}
	return c.store.UpdateService(ctx, updated)
}

func (c deleteServiceCommand) Execute(ctx context.Context, tenantID, serviceID string) (bool, error) {
	return c.store.DeleteService(ctx, tenantID, serviceID)
}

func (c archiveServiceCommand) Execute(ctx context.Context, tenantID, serviceID string) (monitorconfig.Service, error) {
	return c.store.ArchiveService(ctx, tenantID, serviceID)
}

func (c reactivateServiceCommand) Execute(ctx context.Context, tenantID, serviceID string) (monitorconfig.Service, error) {
	return c.store.ReactivateService(ctx, tenantID, serviceID)
}

type serviceListItem struct {
	Service monitorconfig.Service
	Metrics serviceCardMetricsResponse
}

func (q listServicesQuery) Execute(ctx context.Context, tenantID string) ([]serviceListItem, error) {
	services, err := q.store.ListServices(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	items := make([]serviceListItem, 0, len(services))
	for _, service := range services {
		metrics, err := q.store.GetServiceCardMetrics(ctx, tenantID, service.ServiceID)
		if err != nil {
			return nil, err
		}
		items = append(items, serviceListItem{Service: service, Metrics: metrics})
	}
	return items, nil
}

type serviceMonitorDetail struct {
	Monitor monitorconfig.Monitor
	Status  *resultstatus.MonitorStatus
}

type serviceDetail struct {
	Service  monitorconfig.Service
	Metrics  serviceCardMetricsResponse
	Monitors []serviceMonitorDetail
}

func (q getServiceQuery) Execute(ctx context.Context, tenantID, serviceID string) (serviceDetail, bool, error) {
	service, found, err := q.store.GetService(ctx, tenantID, serviceID)
	if err != nil || !found {
		return serviceDetail{}, found, err
	}
	monitors, err := q.store.ListMonitors(ctx, tenantID, serviceID)
	if err != nil {
		return serviceDetail{}, false, err
	}
	metrics, err := q.store.GetServiceCardMetrics(ctx, tenantID, serviceID)
	if err != nil {
		return serviceDetail{}, false, err
	}
	detail := serviceDetail{Service: service, Metrics: metrics, Monitors: make([]serviceMonitorDetail, 0, len(monitors))}
	for _, monitor := range monitors {
		status, foundStatus, err := q.store.GetMonitorStatus(ctx, tenantID, serviceID, monitor.MonitorID)
		if err != nil {
			return serviceDetail{}, false, err
		}
		var statusValue *resultstatus.MonitorStatus
		if foundStatus {
			statusValue = &status
		}
		detail.Monitors = append(detail.Monitors, serviceMonitorDetail{Monitor: monitor, Status: statusValue})
	}
	return detail, true, nil
}

func (q serviceAuditQuery) Execute(ctx context.Context, tenantID, serviceID string, limit int32, startKey map[string]sharedaws.AttributeValue) (historyPage[auditEventView], bool, error) {
	if _, found, err := q.store.GetService(ctx, tenantID, serviceID); err != nil || !found {
		return historyPage[auditEventView]{}, found, err
	}
	page, err := q.store.ListServiceAuditEventsPage(ctx, tenantID, serviceID, limit, startKey)
	return page, true, err
}
