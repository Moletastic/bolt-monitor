package main

import (
	"context"
	"strings"
	"testing"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/resultstatus"
)

type recordingCreateServiceStore struct{ created monitorconfig.Service }

func (s *recordingCreateServiceStore) CreateService(_ context.Context, service monitorconfig.Service) (monitorconfig.Service, error) {
	s.created = service
	return service, nil
}

func TestCreateServiceCommandNormalizesAndPersistsService(t *testing.T) {
	store := &recordingCreateServiceStore{}
	command := newServiceOperations(store, nil, nil, nil, nil, nil, nil, nil, func() time.Time { return time.Unix(0, 0) }, identifierGenerator{newServiceID: func(time.Time) string { return "SVC_TEST" }}).create

	created, err := command.Execute(context.Background(), createServiceInput{
		TenantID: " DEFAULT ",
		Request:  monitorconfig.CreateServiceRequest{Name: "  Payments  ", Description: "  Card processing  "},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if created.ServiceID != store.created.ServiceID || created.Name != store.created.Name {
		t.Fatalf("created = %+v, stored = %+v", created, store.created)
	}
	if created.TenantID != defaultTenantID || created.Name != "Payments" || created.Description != "Card processing" {
		t.Fatalf("created = %+v", created)
	}
	if !strings.HasPrefix(created.ServiceID, "SVC_") {
		t.Fatalf("service ID = %q, want SVC_ prefix", created.ServiceID)
	}
}

type recordingUpdateServiceStore struct {
	service monitorconfig.Service
	updated monitorconfig.Service
}

func (s *recordingUpdateServiceStore) GetService(_ context.Context, _, _ string) (monitorconfig.Service, bool, error) {
	return s.service, true, nil
}

func (s *recordingUpdateServiceStore) UpdateService(_ context.Context, service monitorconfig.Service) (monitorconfig.Service, error) {
	s.updated = service
	return service, nil
}

func TestUpdateServiceCommandOwnsPatchNormalization(t *testing.T) {
	store := &recordingUpdateServiceStore{service: monitorconfig.Service{
		TenantID: defaultTenantID, ServiceID: "payments", Name: "Payments", LifecycleState: monitorconfig.ServiceLifecycleDraft,
	}}
	name := "  Billing  "
	description := "  Invoice processing  "
	updated, err := (updateServiceCommand{store: store}).Execute(context.Background(), updateServiceInput{
		TenantID: defaultTenantID, ServiceID: "payments", Request: updateServiceRequest{Name: &name, Description: &description},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if updated.Name != "Billing" || updated.Description != "Invoice processing" {
		t.Fatalf("updated = %+v", updated)
	}
	if store.updated.Name != updated.Name || store.updated.Description != updated.Description {
		t.Fatalf("stored = %+v, updated = %+v", store.updated, updated)
	}
}

type recordingGetServiceStore struct {
	service  monitorconfig.Service
	monitors []monitorconfig.Monitor
	statuses map[string]resultstatus.MonitorStatus
}

func (s recordingGetServiceStore) GetService(_ context.Context, _, _ string) (monitorconfig.Service, bool, error) {
	return s.service, true, nil
}

func (s recordingGetServiceStore) ListMonitors(_ context.Context, _, _ string) ([]monitorconfig.Monitor, error) {
	return s.monitors, nil
}

func (s recordingGetServiceStore) GetMonitorStatus(_ context.Context, _, _, monitorID string) (resultstatus.MonitorStatus, bool, error) {
	status, found := s.statuses[monitorID]
	return status, found, nil
}

func (recordingGetServiceStore) GetServiceCardMetrics(context.Context, string, string) (serviceCardMetricsResponse, error) {
	return serviceCardMetricsResponse{State: serviceCardMetricStateReady, MonitorCount: 2}, nil
}

func TestGetServiceQueryEnrichesMonitorsWithoutMutationPort(t *testing.T) {
	store := recordingGetServiceStore{
		service:  monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "payments", Name: "Payments", LifecycleState: monitorconfig.ServiceLifecycleDraft},
		monitors: []monitorconfig.Monitor{{MonitorID: "api"}, {MonitorID: "worker"}},
		statuses: map[string]resultstatus.MonitorStatus{"api": {MonitorID: "api", CurrentStatus: "up"}},
	}
	detail, found, err := (getServiceQuery{store: store}).Execute(context.Background(), defaultTenantID, "payments")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !found || detail.Metrics.MonitorCount != 2 || len(detail.Monitors) != 2 {
		t.Fatalf("detail = %+v, found = %v", detail, found)
	}
	if detail.Monitors[0].Status == nil || detail.Monitors[1].Status != nil {
		t.Fatalf("monitor statuses = %+v", detail.Monitors)
	}
}

type recordingServiceAuditStore struct {
	found      bool
	auditCalls int
}

func (s *recordingServiceAuditStore) GetService(context.Context, string, string) (monitorconfig.Service, bool, error) {
	return monitorconfig.Service{}, s.found, nil
}

func (s *recordingServiceAuditStore) ListServiceAuditEventsPage(context.Context, string, string, int32, map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error) {
	s.auditCalls++
	return historyPage[auditEventView]{Items: []auditEventView{{AuditID: "evt-1"}}}, nil
}

func TestServiceAuditQueryChecksServiceBeforeReadingAudit(t *testing.T) {
	store := &recordingServiceAuditStore{}
	page, found, err := (serviceAuditQuery{store: store}).Execute(context.Background(), defaultTenantID, "payments", historyPageSize, nil)
	if err != nil || found || store.auditCalls != 0 || len(page.Items) != 0 {
		t.Fatalf("page = %+v, found = %v, calls = %d, err = %v", page, found, store.auditCalls, err)
	}

	store.found = true
	page, found, err = (serviceAuditQuery{store: store}).Execute(context.Background(), defaultTenantID, "payments", historyPageSize, nil)
	if err != nil || !found || store.auditCalls != 1 || len(page.Items) != 1 {
		t.Fatalf("page = %+v, found = %v, calls = %d, err = %v", page, found, store.auditCalls, err)
	}
}
