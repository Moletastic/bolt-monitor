package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bolt-monitor/shared/api/response"
	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/dynamodbschema"
	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/notifications"
	"bolt-monitor/shared/probelocationcatalog"
	"bolt-monitor/shared/resultstatus"
	"github.com/aws/aws-lambda-go/events"
)

const (
	defaultServiceIncidentsLimit = int32(5)
	maxServiceIncidentsLimit     = int32(50)
	historyPageSize              = int32(20)
)

type monitorRepository interface {
	CreateService(context.Context, monitorconfig.Service) (monitorconfig.Service, error)
	ListServices(context.Context, string) ([]monitorconfig.Service, error)
	GetService(context.Context, string, string) (monitorconfig.Service, bool, error)
	UpdateService(context.Context, monitorconfig.Service) (monitorconfig.Service, error)
	DeleteService(context.Context, string, string) (bool, error)
	CreateMonitor(context.Context, monitorconfig.Monitor) (monitorconfig.Monitor, error)
	ListMonitors(context.Context, string, string) ([]monitorconfig.Monitor, error)
	GetMonitor(context.Context, string, string, string) (monitorconfig.Monitor, bool, error)
	UpdateMonitor(context.Context, monitorconfig.Monitor) (monitorconfig.Monitor, error)
	DeleteMonitor(context.Context, string, string, string) (bool, error)
	SetMonitorEnabled(context.Context, string, string, string, bool) (monitorconfig.Monitor, bool, error)
	SetMonitorMaintenance(context.Context, string, string, string, bool) (resultstatus.MonitorStatus, bool, error)
	ArchiveService(context.Context, string, string) (monitorconfig.Service, error)
	ReactivateService(context.Context, string, string) (monitorconfig.Service, error)
	GetMonitorStatus(context.Context, string, string, string) (resultstatus.MonitorStatus, bool, error)
	ListMonitorRuns(context.Context, string, string, string, int32) ([]resultstatus.CheckRun, error)
	ListMonitorRunsPage(context.Context, string, string, string, int32, map[string]sharedaws.AttributeValue) (historyPage[resultstatus.CheckRun], error)
	GetServiceCardMetrics(context.Context, string, string) (serviceCardMetricsResponse, error)
	CreateManualRun(context.Context, monitorconfig.Monitor, time.Time) (manualRunRequestRecord, error)
	RecordExecutionResult(ctx context.Context, monitor monitorconfig.Monitor, runID string, result checkexecution.ExecutionResult) error
	ListIncidents(context.Context, string, string) ([]dynamodbrecord.IncidentRecord, error)
	GetIncident(context.Context, string, string) (dynamodbrecord.IncidentRecord, bool, error)
	ListIncidentActivities(context.Context, string, string) ([]dynamodbrecord.IncidentActivityRecord, error)
	ListMonitorIncidents(context.Context, string, string, string) ([]dynamodbrecord.IncidentRecord, error)
	ListMonitorIncidentsPage(context.Context, string, string, string, int32, map[string]sharedaws.AttributeValue) (historyPage[dynamodbrecord.IncidentRecord], error)
	ListServiceIncidents(context.Context, string, string, int32) ([]dynamodbrecord.IncidentRecord, error)
	AcknowledgeIncident(context.Context, string, string, time.Time) (dynamodbrecord.IncidentRecord, bool, error)
	ResolveIncident(context.Context, string, string, time.Time) (dynamodbrecord.IncidentRecord, bool, error)
	GetSchedulerConfig(context.Context, string) (dynamodbrecord.SchedulerConfigRecord, error)
	UpdateSchedulerConfig(context.Context, string, checkexecution.SchedulerConfig, time.Time) (dynamodbrecord.SchedulerConfigRecord, error)
	ListMonitorAuditEvents(context.Context, string, string, string) ([]auditEventView, error)
	ListMonitorAuditEventsPage(context.Context, string, string, string, int32, map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error)
	ListServiceAuditEvents(context.Context, string, string) ([]auditEventView, error)
	ListServiceAuditEventsPage(context.Context, string, string, int32, map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error)
	CreateEscalationPolicy(context.Context, escalation.EscalationPolicy) (escalation.EscalationPolicy, error)
	ListEscalationPolicies(context.Context, string) ([]escalation.EscalationPolicy, error)
	GetEscalationPolicy(context.Context, string, string) (*escalation.EscalationPolicy, error)
	UpdateEscalationPolicy(context.Context, escalation.EscalationPolicy) (escalation.EscalationPolicy, error)
	DeleteEscalationPolicy(context.Context, string, string) error
	ServiceReferencesEscalationPolicy(context.Context, string, string) (bool, error)
	CreateNotificationChannel(context.Context, escalation.NotificationChannel) (escalation.NotificationChannel, error)
	ListNotificationChannels(context.Context, string) ([]escalation.NotificationChannel, error)
	GetNotificationChannel(context.Context, string, string) (*escalation.NotificationChannel, error)
	UpdateNotificationChannel(context.Context, escalation.NotificationChannel) (escalation.NotificationChannel, error)
	DeleteNotificationChannel(context.Context, string, string) error
	ChannelsReferencedByRoutes(context.Context, string, string) ([]routeReference, error)
	RecordNotificationChannelTestAudit(context.Context, string, string, string, string, string, time.Time) error
	GetEscalationState(context.Context, string, string) (*escalation.EscalationState, error)
	SearchResources(context.Context, string, string, int, map[string]struct{}) ([]searchResult, error)
}

type monitorHandler struct {
	repo     monitorRepository
	catalog  probelocationcatalog.Catalog
	tenantID string
	now      func() time.Time
	senders  notifications.SenderRegistry
}

func newMonitorHandler(repo monitorRepository, catalog probelocationcatalog.Catalog, tenantID string) monitorHandler {
	return monitorHandler{repo: repo, catalog: catalog, tenantID: tenantID, now: time.Now, senders: defaultNotificationSenderRegistry()}
}

func defaultNotificationSenderRegistry() notifications.SenderRegistry {
	return notifications.SenderRegistry{
		"telegram":  notifications.NewTelegramSender(),
		"email":     notifications.NewEmailSender(),
		"sms":       notifications.NewSMSSender(),
		"webhook":   notifications.NewWebhookSender(),
		"pagerduty": notifications.NewPagerDutySender(),
	}
}

func (h monitorHandler) handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	path := request.RawPath
	method := request.RequestContext.HTTP.Method
	serviceID := strings.TrimSpace(request.PathParameters["serviceId"])
	monitorID := strings.TrimSpace(request.PathParameters["monitorId"])
	incidentID := strings.TrimSpace(request.PathParameters["incidentId"])
	policyID := strings.TrimSpace(request.PathParameters["policyId"])
	channelID := strings.TrimSpace(request.PathParameters["channelId"])
	if channelID == "" && strings.HasPrefix(path, "/api/v1/notification-channels/") {
		channelID = strings.TrimPrefix(path, "/api/v1/notification-channels/")
		channelID = strings.TrimSuffix(channelID, "/test")
	}
	switch {
	case method == http.MethodGet && path == "/api/v1/search":
		return h.searchResources(ctx, request)
	case method == http.MethodGet && path == "/api/v1/probe-locations":
		return h.listProbeLocations()
	case method == http.MethodPost && path == "/api/v1/notification-channels":
		return h.createNotificationChannel(ctx, request)
	case method == http.MethodGet && path == "/api/v1/notification-channels":
		return h.listNotificationChannels(ctx)
	case method == http.MethodGet && channelID != "" && path == "/api/v1/notification-channels/"+channelID:
		return h.getNotificationChannel(ctx, channelID)
	case method == http.MethodPut && channelID != "" && path == "/api/v1/notification-channels/"+channelID:
		return h.updateNotificationChannel(ctx, channelID, request)
	case method == http.MethodDelete && channelID != "" && path == "/api/v1/notification-channels/"+channelID:
		return h.deleteNotificationChannel(ctx, channelID)
	case method == http.MethodPost && channelID != "" && path == "/api/v1/notification-channels/"+channelID+"/test":
		return h.testNotificationChannel(ctx, channelID)
	case method == http.MethodPost && path == "/api/v1/escalation-policies":
		return h.createEscalationPolicy(ctx, request)
	case method == http.MethodGet && path == "/api/v1/escalation-policies":
		return h.listEscalationPolicies(ctx)
	case method == http.MethodGet && policyID != "" && path == "/api/v1/escalation-policies/"+policyID:
		return h.getEscalationPolicy(ctx, policyID)
	case method == http.MethodPut && policyID != "" && path == "/api/v1/escalation-policies/"+policyID:
		return h.updateEscalationPolicy(ctx, policyID, request)
	case method == http.MethodDelete && policyID != "" && path == "/api/v1/escalation-policies/"+policyID:
		return h.deleteEscalationPolicy(ctx, policyID)
	case method == http.MethodGet && path == "/api/v1/incidents":
		return h.listIncidents(ctx, request)
	case method == http.MethodGet && incidentID != "" && strings.HasSuffix(path, "/escalation-state"):
		return h.getEscalationState(ctx, incidentID)
	case method == http.MethodGet && strings.HasSuffix(path, "/activities") && incidentID != "":
		return h.getIncidentActivities(ctx, incidentID)
	case method == http.MethodGet && incidentID != "":
		return h.getIncident(ctx, incidentID)
	case method == http.MethodPost && strings.HasSuffix(path, "/ack") && incidentID != "":
		return h.acknowledgeIncident(ctx, incidentID)
	case method == http.MethodPost && strings.HasSuffix(path, "/resolve") && incidentID != "":
		return h.resolveIncident(ctx, incidentID)
	case method == http.MethodGet && path == "/api/v1/admin/scheduler-config":
		return h.getSchedulerConfig(ctx)
	case method == http.MethodPatch && path == "/api/v1/admin/scheduler-config":
		return h.updateSchedulerConfig(ctx, request)
	case method == http.MethodPost && path == "/api/v1/services":
		return h.createService(ctx, request)
	case method == http.MethodGet && path == "/api/v1/services":
		return h.listServices(ctx)
	case method == http.MethodGet && serviceID != "" && path == "/api/v1/services/"+serviceID:
		return h.getService(ctx, serviceID)
	case method == http.MethodGet && serviceID != "" && path == "/api/v1/services/"+serviceID+"/escalation-policy":
		return h.getServiceEscalationPolicy(ctx, serviceID)
	case method == http.MethodGet && serviceID != "" && strings.HasSuffix(path, "/audit") && !isMonitorAuditPath(path):
		return h.getServiceAudit(ctx, serviceID, request)
	case method == http.MethodPatch && serviceID != "" && path == "/api/v1/services/"+serviceID:
		return h.updateService(ctx, serviceID, request)
	case method == http.MethodDelete && serviceID != "" && path == "/api/v1/services/"+serviceID:
		return h.deleteService(ctx, serviceID)
	case method == http.MethodPost && serviceID != "" && path == "/api/v1/services/"+serviceID+"/monitors":
		return h.createMonitor(ctx, serviceID, request)
	case method == http.MethodGet && serviceID != "" && path == "/api/v1/services/"+serviceID+"/monitors":
		return h.listMonitors(ctx, serviceID)
	case method == http.MethodGet && serviceID != "" && monitorID != "" && path == "/api/v1/services/"+serviceID+"/monitors/"+monitorID:
		return h.getMonitor(ctx, serviceID, monitorID)
	case method == http.MethodPatch && serviceID != "" && monitorID != "" && path == "/api/v1/services/"+serviceID+"/monitors/"+monitorID:
		return h.updateMonitor(ctx, serviceID, monitorID, request)
	case method == http.MethodDelete && serviceID != "" && monitorID != "" && path == "/api/v1/services/"+serviceID+"/monitors/"+monitorID:
		return h.deleteMonitor(ctx, serviceID, monitorID)
	case method == http.MethodPost && strings.HasSuffix(path, "/enable") && serviceID != "" && monitorID != "":
		return h.setMonitorEnabled(ctx, serviceID, monitorID, true)
	case method == http.MethodPost && strings.HasSuffix(path, "/disable") && serviceID != "" && monitorID != "":
		return h.setMonitorEnabled(ctx, serviceID, monitorID, false)
	case method == http.MethodPost && strings.HasSuffix(path, "/maintenance/enable") && serviceID != "" && monitorID != "":
		return h.setMonitorMaintenance(ctx, serviceID, monitorID, true)
	case method == http.MethodPost && strings.HasSuffix(path, "/maintenance/disable") && serviceID != "" && monitorID != "":
		return h.setMonitorMaintenance(ctx, serviceID, monitorID, false)
	case method == http.MethodGet && strings.HasSuffix(path, "/status") && serviceID != "" && monitorID != "":
		return h.getMonitorStatus(ctx, serviceID, monitorID)
	case method == http.MethodGet && strings.HasSuffix(path, "/runs") && serviceID != "" && monitorID != "":
		return h.getMonitorRuns(ctx, serviceID, monitorID, request)
	case method == http.MethodPost && strings.HasSuffix(path, "/run") && serviceID != "" && monitorID != "":
		return h.runMonitor(ctx, serviceID, monitorID)
	case method == http.MethodGet && strings.HasSuffix(path, "/incidents") && serviceID != "" && monitorID == "":
		return h.getServiceIncidents(ctx, serviceID, request)
	case method == http.MethodGet && strings.HasSuffix(path, "/incidents") && serviceID != "" && monitorID != "":
		return h.getMonitorIncidents(ctx, serviceID, monitorID, request)
	case method == http.MethodGet && strings.HasSuffix(path, "/audit") && serviceID != "" && monitorID != "":
		return h.getMonitorAudit(ctx, serviceID, monitorID, request)
	case method == http.MethodPost && strings.HasSuffix(path, "/archive") && serviceID != "":
		return h.archiveService(ctx, serviceID)
	case method == http.MethodPost && strings.HasSuffix(path, "/reactivate") && serviceID != "":
		return h.reactivateService(ctx, serviceID)
	default:
		return respondAPIGateway(sharederrors.New(sharederrors.CodeNotFound, nil))
	}
}

func (h monitorHandler) searchResources(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	limit := defaultSearchLimit
	if raw := strings.TrimSpace(request.QueryStringParameters["limit"]); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return respondAPIGateway(sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "limit", "reason": "must be a positive integer"}))
		}
		limit = parsed
	}
	types, err := parseSearchTypes(request.QueryStringParameters["types"])
	if err != nil {
		return respondAPIGateway(err)
	}
	results, err := h.repo.SearchResources(ctx, h.tenantID, request.QueryStringParameters["q"], limit, types)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(searchResponse{Results: results}))
}

func historyStartKey(request events.APIGatewayV2HTTPRequest, resource, resourceKeyName string) (map[string]sharedaws.AttributeValue, error) {
	key, err := decodeHistoryCursor(strings.TrimSpace(request.QueryStringParameters["cursor"]), resource, resourceKeyName)
	if err != nil {
		return nil, sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{
			"field":  "cursor",
			"reason": "must be a valid cursor for this resource",
		})
	}
	return key, nil
}

func (h monitorHandler) createService(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var payload monitorconfig.CreateServiceRequest
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return respondAPIGateway(sharederrors.Wrap(sharederrors.CodeInvalidJSON, err, nil))
	}
	service, err := payload.ToService(h.tenantID)
	if err != nil {
		return respondAPIGateway(err)
	}
	service.ServiceID = newServiceID(h.now())
	created, err := h.repo.CreateService(ctx, service)
	if err != nil {
		return respondAPIGateway(err)
	}
	location := fmt.Sprintf("/api/v1/services/%s", created.ServiceID)
	return envelopeResponseWithLocation(http.StatusCreated, response.Ok(toServiceResponse(created)), location)
}

func (h monitorHandler) listServices(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	services, err := h.repo.ListServices(ctx, h.tenantID)
	if err != nil {
		return respondAPIGateway(err)
	}
	payload := listServicesResponse{Services: make([]serviceResponse, 0, len(services))}
	for _, service := range services {
		serviceResponse := toServiceResponse(service)
		metrics, err := h.repo.GetServiceCardMetrics(ctx, h.tenantID, service.ServiceID)
		if err != nil {
			return respondAPIGateway(err)
		}
		serviceResponse.CardMetrics = &metrics
		payload.Services = append(payload.Services, serviceResponse)
	}
	return envelopeResponse(http.StatusOK, response.OkPaginated(payload, 1, len(payload.Services), len(payload.Services)))
}

func (h monitorHandler) getService(ctx context.Context, serviceID string) (events.APIGatewayV2HTTPResponse, error) {
	service, found, err := h.repo.GetService(ctx, h.tenantID, serviceID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeServiceNotFound, nil))
	}
	monitors, err := h.repo.ListMonitors(ctx, h.tenantID, serviceID)
	if err != nil {
		return respondAPIGateway(err)
	}
	serviceResponse := toServiceResponse(service)
	metrics, err := h.repo.GetServiceCardMetrics(ctx, h.tenantID, serviceID)
	if err != nil {
		return respondAPIGateway(err)
	}
	serviceResponse.CardMetrics = &metrics
	serviceResponse.Monitors = make([]monitorResponse, 0, len(monitors))
	for _, monitor := range monitors {
		status, foundStatus, err := h.repo.GetMonitorStatus(ctx, h.tenantID, serviceID, monitor.MonitorID)
		if err != nil {
			return respondAPIGateway(err)
		}
		var statusResponse *monitorStatusResponse
		if foundStatus {
			statusResponse = toStatusResponse(status)
		}
		serviceResponse.Monitors = append(serviceResponse.Monitors, toMonitorResponse(monitor, statusResponse))
	}
	return envelopeResponse(http.StatusOK, response.Ok(serviceResponse))
}

func (h monitorHandler) updateService(ctx context.Context, serviceID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	current, found, err := h.repo.GetService(ctx, h.tenantID, serviceID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeServiceNotFound, nil))
	}
	var payload updateServiceRequest
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return respondAPIGateway(sharederrors.Wrap(sharederrors.CodeInvalidJSON, err, nil))
	}
	if payload.ServiceID != nil && !strings.EqualFold(*payload.ServiceID, current.ServiceID) {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeImmutableField, map[string]any{"field": "serviceId"}))
	}
	updated := current
	if payload.Name != nil {
		updated.Name = strings.TrimSpace(*payload.Name)
	}
	if payload.Description != nil {
		updated.Description = strings.TrimSpace(*payload.Description)
	}
	if payload.ServiceCategory != nil {
		updated.ServiceCategory = monitorconfig.ServiceCategory(strings.TrimSpace(*payload.ServiceCategory))
	}
	if payload.EscalationPolicyID != nil {
		updated.EscalationPolicyID = strings.TrimSpace(*payload.EscalationPolicyID)
	}
	if payload.BusinessHours != nil {
		updated.BusinessHours = dynamodbrecord.CloneBusinessHoursConfig(payload.BusinessHours)
	}
	if err := updated.Validate(); err != nil {
		return respondAPIGateway(err)
	}
	stored, err := h.repo.UpdateService(ctx, updated)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toServiceResponse(stored)))
}

func (h monitorHandler) deleteService(ctx context.Context, serviceID string) (events.APIGatewayV2HTTPResponse, error) {
	deleted, err := h.repo.DeleteService(ctx, h.tenantID, serviceID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !deleted {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeServiceNotFound, nil))
	}
	return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusNoContent}, nil
}

func (h monitorHandler) archiveService(ctx context.Context, serviceID string) (events.APIGatewayV2HTTPResponse, error) {
	service, err := h.repo.ArchiveService(ctx, h.tenantID, serviceID)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toServiceResponse(service)))
}

func (h monitorHandler) reactivateService(ctx context.Context, serviceID string) (events.APIGatewayV2HTTPResponse, error) {
	service, err := h.repo.ReactivateService(ctx, h.tenantID, serviceID)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toServiceResponse(service)))
}

func (h monitorHandler) createMonitor(ctx context.Context, serviceID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if _, found, err := h.repo.GetService(ctx, h.tenantID, serviceID); err != nil {
		return respondAPIGateway(err)
	} else if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeServiceNotFound, nil))
	}
	var payload monitorconfig.CreateMonitorRequest
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return respondAPIGateway(sharederrors.Wrap(sharederrors.CodeInvalidJSON, err, nil))
	}
	if len(payload.ProbeLocations) == 0 {
		payload.ProbeLocations = defaultSelectableProbeLocations(h.catalog)
	}
	var targetURL string
	if payload.HTTP != nil {
		targetURL = payload.HTTP.Target
	}
	monitorID := generateMonitorID(string(payload.Type), targetURL, payload.Name)
	monitor, err := payload.ToMonitor(serviceID, h.tenantID, monitorID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if err := monitor.ValidateWithCatalog(h.catalog); err != nil {
		return respondAPIGateway(err)
	}
	created, err := h.repo.CreateMonitor(ctx, monitor)
	if err != nil {
		return respondAPIGateway(err)
	}
	location := fmt.Sprintf("/api/v1/services/%s/monitors/%s", serviceID, created.MonitorID)
	return envelopeResponseWithLocation(http.StatusCreated, response.Ok(toMonitorResponse(created, toUnknownStatusResponse())), location)
}

func defaultSelectableProbeLocations(catalog probelocationcatalog.Catalog) []string {
	for _, location := range catalog.Locations {
		if location.Enabled {
			return []string{location.LocationID}
		}
	}
	return nil
}

func (h monitorHandler) listMonitors(ctx context.Context, serviceID string) (events.APIGatewayV2HTTPResponse, error) {
	if _, found, err := h.repo.GetService(ctx, h.tenantID, serviceID); err != nil {
		return respondAPIGateway(err)
	} else if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeServiceNotFound, nil))
	}
	monitors, err := h.repo.ListMonitors(ctx, h.tenantID, serviceID)
	if err != nil {
		return respondAPIGateway(err)
	}
	payload := listMonitorsResponse{Monitors: make([]monitorResponse, 0, len(monitors))}
	for _, monitor := range monitors {
		status, found, err := h.repo.GetMonitorStatus(ctx, h.tenantID, serviceID, monitor.MonitorID)
		if err != nil {
			return respondAPIGateway(err)
		}
		var statusResponse *monitorStatusResponse
		if found {
			statusResponse = toStatusResponse(status)
		}
		payload.Monitors = append(payload.Monitors, toMonitorResponse(monitor, statusResponse))
	}
	return envelopeResponse(http.StatusOK, response.OkPaginated(payload, 1, len(payload.Monitors), len(payload.Monitors)))
}

func (h monitorHandler) getMonitor(ctx context.Context, serviceID, monitorID string) (events.APIGatewayV2HTTPResponse, error) {
	monitor, found, err := h.repo.GetMonitor(ctx, h.tenantID, serviceID, monitorID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeMonitorNotFound, nil))
	}
	status, foundStatus, err := h.repo.GetMonitorStatus(ctx, h.tenantID, serviceID, monitorID)
	if err != nil {
		return respondAPIGateway(err)
	}
	var statusResponse *monitorStatusResponse
	if foundStatus {
		statusResponse = toStatusResponse(status)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toMonitorResponse(monitor, statusResponse)))
}

func (h monitorHandler) updateMonitor(ctx context.Context, serviceID, monitorID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	current, found, err := h.repo.GetMonitor(ctx, h.tenantID, serviceID, monitorID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeMonitorNotFound, nil))
	}
	var payload updateMonitorRequest
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return respondAPIGateway(sharederrors.Wrap(sharederrors.CodeInvalidJSON, err, nil))
	}
	updated := current
	if payload.Name != nil {
		updated.Name = *payload.Name
	}
	if payload.IntervalSeconds != nil {
		updated.IntervalSeconds = *payload.IntervalSeconds
	}
	if payload.ProbeLocations != nil {
		updated.ProbeLocations = append([]string(nil), payload.ProbeLocations...)
	}
	if payload.FailureThreshold != nil {
		updated.FailureThreshold = *payload.FailureThreshold
	}
	if payload.RecoveryThreshold != nil {
		updated.RecoveryThreshold = *payload.RecoveryThreshold
	}
	if updated.FailureThreshold < 1 {
		updated.FailureThreshold = 1
	}
	if updated.RecoveryThreshold < 1 {
		updated.RecoveryThreshold = 1
	}
	if payload.HTTP != nil {
		updated.HTTP = dynamodbrecord.CloneHTTPConfiguration(payload.HTTP)
	}
	if err := updated.ValidateWithCatalog(h.catalog); err != nil {
		return respondAPIGateway(err)
	}
	stored, err := h.repo.UpdateMonitor(ctx, updated)
	if err != nil {
		return respondAPIGateway(err)
	}
	status, foundStatus, err := h.repo.GetMonitorStatus(ctx, h.tenantID, serviceID, monitorID)
	if err != nil {
		return respondAPIGateway(err)
	}
	var statusResponse *monitorStatusResponse
	if foundStatus {
		statusResponse = toStatusResponse(status)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toMonitorResponse(stored, statusResponse)))
}

func (h monitorHandler) deleteMonitor(ctx context.Context, serviceID, monitorID string) (events.APIGatewayV2HTTPResponse, error) {
	deleted, err := h.repo.DeleteMonitor(ctx, h.tenantID, serviceID, monitorID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !deleted {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeMonitorNotFound, nil))
	}
	return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusNoContent}, nil
}

func (h monitorHandler) setMonitorEnabled(ctx context.Context, serviceID, monitorID string, enabled bool) (events.APIGatewayV2HTTPResponse, error) {
	monitor, found, err := h.repo.SetMonitorEnabled(ctx, h.tenantID, serviceID, monitorID, enabled)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeMonitorNotFound, nil))
	}
	status, foundStatus, err := h.repo.GetMonitorStatus(ctx, h.tenantID, serviceID, monitorID)
	if err != nil {
		return respondAPIGateway(err)
	}
	var statusResponse *monitorStatusResponse
	if foundStatus {
		statusResponse = toStatusResponse(status)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toMonitorResponse(monitor, statusResponse)))
}

func (h monitorHandler) setMonitorMaintenance(ctx context.Context, serviceID, monitorID string, enabled bool) (events.APIGatewayV2HTTPResponse, error) {
	_, found, err := h.repo.GetMonitor(ctx, h.tenantID, serviceID, monitorID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeMonitorNotFound, nil))
	}
	status, _, err := h.repo.SetMonitorMaintenance(ctx, h.tenantID, serviceID, monitorID, enabled)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(*toStatusResponse(status)))
}

func (h monitorHandler) getMonitorStatus(ctx context.Context, serviceID, monitorID string) (events.APIGatewayV2HTTPResponse, error) {
	status, found, err := h.repo.GetMonitorStatus(ctx, h.tenantID, serviceID, monitorID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeMonitorStatusNotFound, nil))
	}
	return envelopeResponse(http.StatusOK, response.Ok(*toStatusResponse(status)))
}

func (h monitorHandler) getMonitorRuns(ctx context.Context, serviceID, monitorID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	resource := dynamodbschema.MonitorPK(h.tenantID, serviceID, monitorID)
	startKey, err := historyStartKey(request, resource, "PK")
	if err != nil {
		return respondAPIGateway(err)
	}
	page, err := h.repo.ListMonitorRunsPage(ctx, h.tenantID, serviceID, monitorID, historyPageSize, startKey)
	if err != nil {
		return respondAPIGateway(err)
	}
	payload := monitorRunsResponse{Runs: make([]checkRunResponse, 0, len(page.Items))}
	for _, run := range page.Items {
		payload.Runs = append(payload.Runs, toRunResponse(run))
	}
	nextCursor, err := encodeHistoryCursor(resource, page.NextKey)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.OkCursorPaginated(payload, len(payload.Runs), nextCursor))
}

func (h monitorHandler) runMonitor(ctx context.Context, serviceID, monitorID string) (events.APIGatewayV2HTTPResponse, error) {
	monitor, found, err := h.repo.GetMonitor(ctx, h.tenantID, serviceID, monitorID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeMonitorNotFound, nil))
	}
	if !monitor.Enabled {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeMonitorDisabled, nil))
	}
	now := h.now()
	runID := newRunID(now)
	probeLocationID := monitor.ProbeLocations[0]
	var probeLocation probelocationcatalog.Location
	for _, loc := range h.catalog.Locations {
		if loc.LocationID == probeLocationID {
			probeLocation = loc
			break
		}
	}
	request := checkexecution.ExecutionRequest{
		Monitor:       monitor,
		ProbeLocation: probeLocation,
		RunID:         runID,
		Trigger:       checkexecution.TriggerTypeManual,
	}
	httpClient := &http.Client{
		Timeout: time.Duration(monitor.HTTP.TimeoutMs) * time.Millisecond,
	}
	result := checkexecution.ExecuteHTTP(ctx, httpClient, request)
	if err := h.repo.RecordExecutionResult(ctx, monitor, runID, result); err != nil {
		return respondAPIGateway(err)
	}
	runRecord := manualRunRequestRecord{
		RunID:      runID,
		ServiceID:  monitor.ServiceID,
		MonitorID:  monitor.MonitorID,
		TenantID:   monitor.TenantID,
		Trigger:    checkexecution.TriggerTypeManual,
		AcceptedAt: now.UTC().Format(time.RFC3339),
	}
	return envelopeResponse(http.StatusOK, response.Ok(toManualRunResponseWithResult(runRecord, result)))
}

func (h monitorHandler) listIncidents(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	status := strings.ToLower(strings.TrimSpace(request.QueryStringParameters["status"]))
	incidents, err := h.repo.ListIncidents(ctx, h.tenantID, status)
	if err != nil {
		return respondAPIGateway(err)
	}
	payload := listIncidentsResponse{Incidents: make([]incidentResponse, 0, len(incidents))}
	for _, incident := range incidents {
		payload.Incidents = append(payload.Incidents, toIncidentResponse(incident))
	}
	return envelopeResponse(http.StatusOK, response.Ok(payload))
}

func (h monitorHandler) getIncident(ctx context.Context, incidentID string) (events.APIGatewayV2HTTPResponse, error) {
	incident, found, err := h.repo.GetIncident(ctx, h.tenantID, incidentID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeIncidentNotFound, nil))
	}
	return envelopeResponse(http.StatusOK, response.Ok(toIncidentResponse(incident)))
}

func (h monitorHandler) getEscalationState(ctx context.Context, incidentID string) (events.APIGatewayV2HTTPResponse, error) {
	if _, found, err := h.repo.GetIncident(ctx, h.tenantID, incidentID); err != nil {
		return respondAPIGateway(err)
	} else if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeIncidentNotFound, nil))
	}
	state, err := h.repo.GetEscalationState(ctx, h.tenantID, incidentID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if state == nil {
		return envelopeResponse(http.StatusOK, response.Ok(escalationStateResponse{Exists: false}))
	}
	return envelopeResponse(http.StatusOK, response.Ok(toEscalationStateResponse(*state)))
}

func (h monitorHandler) getIncidentActivities(ctx context.Context, incidentID string) (events.APIGatewayV2HTTPResponse, error) {
	if _, found, err := h.repo.GetIncident(ctx, h.tenantID, incidentID); err != nil {
		return respondAPIGateway(err)
	} else if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeIncidentNotFound, nil))
	}
	activities, err := h.repo.ListIncidentActivities(ctx, h.tenantID, incidentID)
	if err != nil {
		return respondAPIGateway(err)
	}
	payload := incidentActivitiesResponse{Activities: make([]incidentActivityResponse, 0, len(activities))}
	for _, activity := range activities {
		payload.Activities = append(payload.Activities, toIncidentActivityResponse(activity))
	}
	return envelopeResponse(http.StatusOK, response.Ok(payload))
}

func (h monitorHandler) getMonitorIncidents(ctx context.Context, serviceID, monitorID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if _, found, err := h.repo.GetMonitor(ctx, h.tenantID, serviceID, monitorID); err != nil {
		return respondAPIGateway(err)
	} else if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeMonitorNotFound, nil))
	}
	resource := dynamodbschema.MonitorPK(h.tenantID, serviceID, monitorID)
	startKey, err := historyStartKey(request, resource, "PK")
	if err != nil {
		return respondAPIGateway(err)
	}
	page, err := h.repo.ListMonitorIncidentsPage(ctx, h.tenantID, serviceID, monitorID, historyPageSize, startKey)
	if err != nil {
		return respondAPIGateway(err)
	}
	payload := listIncidentsResponse{Incidents: make([]incidentResponse, 0, len(page.Items))}
	for _, incident := range page.Items {
		payload.Incidents = append(payload.Incidents, toIncidentResponse(incident))
	}
	nextCursor, err := encodeHistoryCursor(resource, page.NextKey)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.OkCursorPaginated(payload, len(payload.Incidents), nextCursor))
}

func (h monitorHandler) getServiceIncidents(ctx context.Context, serviceID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if _, found, err := h.repo.GetService(ctx, h.tenantID, serviceID); err != nil {
		return respondAPIGateway(err)
	} else if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeServiceNotFound, nil))
	}
	limit := defaultServiceIncidentsLimit
	if raw := strings.TrimSpace(request.QueryStringParameters["limit"]); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || parsed <= 0 || parsed > int64(maxServiceIncidentsLimit) {
			return respondAPIGateway(sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{
				"field":  "limit",
				"reason": "must be a positive integer no greater than the maximum",
			}))
		}
		limit = int32(parsed)
	}
	incidents, err := h.repo.ListServiceIncidents(ctx, h.tenantID, serviceID, limit)
	if err != nil {
		return respondAPIGateway(err)
	}
	payload := listIncidentsResponse{Incidents: make([]incidentResponse, 0, len(incidents))}
	for _, incident := range incidents {
		payload.Incidents = append(payload.Incidents, toIncidentResponse(incident))
	}
	return envelopeResponse(http.StatusOK, response.Ok(payload))
}

func (h monitorHandler) acknowledgeIncident(ctx context.Context, incidentID string) (events.APIGatewayV2HTTPResponse, error) {
	incident, found, err := h.repo.AcknowledgeIncident(ctx, h.tenantID, incidentID, h.now())
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeIncidentNotFound, nil))
	}
	return envelopeResponse(http.StatusOK, response.Ok(toIncidentResponse(incident)))
}

func (h monitorHandler) resolveIncident(ctx context.Context, incidentID string) (events.APIGatewayV2HTTPResponse, error) {
	incident, found, err := h.repo.ResolveIncident(ctx, h.tenantID, incidentID, h.now())
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeIncidentNotFound, nil))
	}
	return envelopeResponse(http.StatusOK, response.Ok(toIncidentResponse(incident)))
}

func (h monitorHandler) getSchedulerConfig(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	config, err := h.repo.GetSchedulerConfig(ctx, h.tenantID)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toSchedulerConfigResponse(config)))
}

func (h monitorHandler) updateSchedulerConfig(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var payload updateSchedulerConfigRequest
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return respondAPIGateway(sharederrors.Wrap(sharederrors.CodeInvalidJSON, err, nil))
	}
	config := checkexecution.SchedulerConfig{RecurringEnabled: payload.RecurringEnabled, StopControlMode: checkexecution.StopControlMode(payload.StopControlMode)}
	if err := config.Validate(); err != nil {
		return respondAPIGateway(err)
	}
	updated, err := h.repo.UpdateSchedulerConfig(ctx, h.tenantID, config, h.now())
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toSchedulerConfigResponse(updated)))
}

func (h monitorHandler) getMonitorAudit(ctx context.Context, serviceID, monitorID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if _, found, err := h.repo.GetMonitor(ctx, h.tenantID, serviceID, monitorID); err != nil {
		return respondAPIGateway(err)
	} else if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeMonitorNotFound, nil))
	}
	resource := dynamodbschema.AuditResourceItem(h.tenantID, serviceID, monitorID, "cursor", "cursor").GSI3PK
	startKey, err := historyStartKey(request, resource, "GSI3PK")
	if err != nil {
		return respondAPIGateway(err)
	}
	page, err := h.repo.ListMonitorAuditEventsPage(ctx, h.tenantID, serviceID, monitorID, historyPageSize, startKey)
	if err != nil {
		return respondAPIGateway(err)
	}
	payload := monitorAuditEventsResponse{Events: make([]auditEventResponse, 0, len(page.Items))}
	for _, event := range page.Items {
		payload.Events = append(payload.Events, toAuditEventResponse(event))
	}
	nextCursor, err := encodeHistoryCursor(resource, page.NextKey)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.OkCursorPaginated(payload, len(payload.Events), nextCursor))
}

func (h monitorHandler) getServiceAudit(ctx context.Context, serviceID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if _, found, err := h.repo.GetService(ctx, h.tenantID, serviceID); err != nil {
		return respondAPIGateway(err)
	} else if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeServiceNotFound, nil))
	}
	resource := dynamodbschema.AuditResourceItem(h.tenantID, serviceID, "", "cursor", "cursor").GSI3PK
	startKey, err := historyStartKey(request, resource, "GSI3PK")
	if err != nil {
		return respondAPIGateway(err)
	}
	page, err := h.repo.ListServiceAuditEventsPage(ctx, h.tenantID, serviceID, historyPageSize, startKey)
	if err != nil {
		return respondAPIGateway(err)
	}
	payload := monitorAuditEventsResponse{Events: make([]auditEventResponse, 0, len(page.Items))}
	for _, event := range page.Items {
		payload.Events = append(payload.Events, toAuditEventResponse(event))
	}
	nextCursor, err := encodeHistoryCursor(resource, page.NextKey)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.OkCursorPaginated(payload, len(payload.Events), nextCursor))
}

func isMonitorAuditPath(path string) bool {
	return strings.Contains(path, "/monitors/")
}

func (h monitorHandler) listProbeLocations() (events.APIGatewayV2HTTPResponse, error) {
	payload := listProbeLocationsResponse{ProbeLocations: make([]probeLocationResponse, 0, len(h.catalog.Locations))}
	for _, location := range h.catalog.Locations {
		if !location.Enabled {
			continue
		}
		payload.ProbeLocations = append(payload.ProbeLocations, probeLocationResponse{LocationID: location.LocationID, DisplayName: location.DisplayName, Enabled: location.Enabled})
	}
	return envelopeResponse(http.StatusOK, response.Ok(payload))
}

func (h monitorHandler) createNotificationChannel(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	channel, resp, ok := h.notificationChannelFromRequest(request, nil)
	if !ok {
		return resp, nil
	}
	channel.TenantID = h.tenantID
	channel.ChannelID = newNotificationChannelID(h.now())
	created, err := h.repo.CreateNotificationChannel(ctx, channel)
	if err != nil {
		return respondAPIGateway(err)
	}
	location := fmt.Sprintf("/api/v1/notification-channels/%s", created.ChannelID)
	return envelopeResponseWithLocation(http.StatusCreated, response.Ok(toNotificationChannelResponse(created)), location)
}

func (h monitorHandler) listNotificationChannels(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	channels, err := h.repo.ListNotificationChannels(ctx, h.tenantID)
	if err != nil {
		return respondAPIGateway(err)
	}
	payload := listNotificationChannelsResponse{Channels: make([]notificationChannelResponse, 0, len(channels))}
	for _, channel := range channels {
		payload.Channels = append(payload.Channels, toNotificationChannelResponse(channel))
	}
	return envelopeResponse(http.StatusOK, response.Ok(payload))
}

func (h monitorHandler) getNotificationChannel(ctx context.Context, channelID string) (events.APIGatewayV2HTTPResponse, error) {
	channel, err := h.repo.GetNotificationChannel(ctx, h.tenantID, channelID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if channel == nil {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeChannelNotFound, nil))
	}
	return envelopeResponse(http.StatusOK, response.Ok(toNotificationChannelResponse(*channel)))
}

func (h monitorHandler) updateNotificationChannel(ctx context.Context, channelID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	current, err := h.repo.GetNotificationChannel(ctx, h.tenantID, channelID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if current == nil {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeChannelNotFound, nil))
	}
	channel, resp, ok := h.notificationChannelFromRequest(request, current)
	if !ok {
		return resp, nil
	}
	updated, err := h.repo.UpdateNotificationChannel(ctx, channel)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toNotificationChannelResponse(updated)))
}

func (h monitorHandler) deleteNotificationChannel(ctx context.Context, channelID string) (events.APIGatewayV2HTTPResponse, error) {
	channel, err := h.repo.GetNotificationChannel(ctx, h.tenantID, channelID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if channel == nil {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeChannelNotFound, nil))
	}
	references, err := h.repo.ChannelsReferencedByRoutes(ctx, h.tenantID, channelID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if len(references) > 0 {
		return envelopeResponse(http.StatusConflict, response.Ok[channelInUseResponse](channelInUseResponse{Error: "channel in use", ReferencingRoutes: references}))
	}
	if err := h.repo.DeleteNotificationChannel(ctx, h.tenantID, channelID); err != nil {
		return respondAPIGateway(err)
	}
	return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusNoContent}, nil
}

func (h monitorHandler) testNotificationChannel(ctx context.Context, channelID string) (events.APIGatewayV2HTTPResponse, error) {
	channel, err := h.repo.GetNotificationChannel(ctx, h.tenantID, channelID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if channel == nil {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeChannelNotFound, nil))
	}
	sender, ok := h.senders.Get(string(channel.Type))
	if !ok {
		return h.recordChannelTestFailure(ctx, *channel, "sender not registered")
	}
	config, err := mergeNotificationChannelTarget(*channel)
	if err != nil {
		return h.recordChannelTestFailure(ctx, *channel, err.Error())
	}
	now := h.now().UTC()
	notification := notifications.Notification{
		EventType:   notifications.EventTypeIncidentDown,
		TenantID:    h.tenantID,
		MonitorID:   "notification-channel-test",
		ServiceID:   "notification-channel-test",
		MonitorName: "Notification channel test",
		ServiceName: "Bolt Monitor",
		Timestamp:   now,
		Message:     fmt.Sprintf("Bolt Monitor test notification\n\nChannel: %s\nType: %s\nThis is a test message from the dashboard. No incident was created.", channel.Name, channel.Type),
		IncidentID:  "notification-channel-test",
		Config:      config,
	}
	if err := sender.Send(ctx, notification); err != nil {
		return h.recordChannelTestFailure(ctx, *channel, err.Error())
	}
	if err := h.repo.RecordNotificationChannelTestAudit(ctx, h.tenantID, channel.ChannelID, string(channel.Type), "success", "", now); err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(notificationChannelTestResponse{ChannelID: channel.ChannelID, SentAt: now.Format(time.RFC3339)}, "Test notification sent."))
}

func (h monitorHandler) recordChannelTestFailure(ctx context.Context, channel escalation.NotificationChannel, reason string) (events.APIGatewayV2HTTPResponse, error) {
	now := h.now().UTC()
	sanitized := sanitizeNotificationDeliveryError(reason, channel.Config)
	if err := h.repo.RecordNotificationChannelTestAudit(ctx, h.tenantID, channel.ChannelID, string(channel.Type), "failure", sanitized, now); err != nil {
		return respondAPIGateway(err)
	}
	return respondAPIGateway(sharederrors.New(sharederrors.CodeNotificationDelivery, map[string]any{"channelId": channel.ChannelID, "type": string(channel.Type), "reason": sanitized}))
}

func mergeNotificationChannelTarget(channel escalation.NotificationChannel) (json.RawMessage, error) {
	config := map[string]any{}
	if len(channel.Config) > 0 {
		if err := json.Unmarshal(channel.Config, &config); err != nil {
			return nil, fmt.Errorf("invalid %s config", channel.Type)
		}
	}
	target := strings.TrimSpace(channel.Target)
	if target != "" {
		switch channel.Type {
		case escalation.ChannelTypeTelegram:
			config["chatId"] = target
		case escalation.ChannelTypeEmail:
			config["toEmail"] = target
		case escalation.ChannelTypeSMS:
			config["toNumber"] = target
		case escalation.ChannelTypeWebhook:
			config["url"] = target
		case escalation.ChannelTypePagerDuty:
			config["routingKey"] = target
		}
	}
	return json.Marshal(config)
}

func sanitizeNotificationDeliveryError(reason string, config json.RawMessage) string {
	trimmed := strings.TrimSpace(reason)
	if trimmed == "" {
		return "notification delivery failed"
	}
	secretValues := map[string]struct{}{}
	var cfg map[string]any
	if err := json.Unmarshal(config, &cfg); err == nil {
		for _, key := range []string{"botToken", "apiKey", "authToken", "accountSid"} {
			if value, ok := cfg[key].(string); ok && strings.TrimSpace(value) != "" {
				secretValues[value] = struct{}{}
			}
		}
	}
	for value := range secretValues {
		trimmed = strings.ReplaceAll(trimmed, value, "[redacted]")
	}
	for _, key := range []string{"botToken", "apiKey", "authToken", "accountSid", "Authorization", "Bearer"} {
		trimmed = strings.ReplaceAll(trimmed, key, "[redacted]")
	}
	if len(trimmed) > 240 {
		trimmed = trimmed[:240]
	}
	return trimmed
}

func (h monitorHandler) notificationChannelFromRequest(request events.APIGatewayV2HTTPRequest, current *escalation.NotificationChannel) (escalation.NotificationChannel, events.APIGatewayV2HTTPResponse, bool) {
	var payload notificationChannelRequest
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		resp, _ := respondAPIGateway(sharederrors.Wrap(sharederrors.CodeInvalidJSON, err, nil))
		return escalation.NotificationChannel{}, resp, false
	}
	if err := validateInput(payload); err != nil {
		resp, _ := respondAPIGateway(err)
		return escalation.NotificationChannel{}, resp, false
	}
	channel := escalation.NotificationChannel{}
	if current != nil {
		channel = *current
	}
	if payload.Name != nil {
		channel.Name = strings.TrimSpace(*payload.Name)
	}
	if payload.Type != nil {
		channel.Type = escalation.ChannelType(strings.TrimSpace(*payload.Type))
	}
	if payload.Target != nil {
		channel.Target = strings.TrimSpace(*payload.Target)
	}
	if payload.Config != nil {
		channel.Config = append(json.RawMessage(nil), payload.Config...)
	}
	if err := validateInput(notificationChannelInput{Name: channel.Name, Type: string(channel.Type), Target: channel.Target}); err != nil {
		resp, _ := respondAPIGateway(err)
		return escalation.NotificationChannel{}, resp, false
	}
	if err := validateNotificationChannel(channel); err != nil {
		resp, _ := respondAPIGateway(err)
		return escalation.NotificationChannel{}, resp, false
	}
	return channel, events.APIGatewayV2HTTPResponse{}, true
}

func validateNotificationChannel(channel escalation.NotificationChannel) error {
	cfg := map[string]any{}
	if len(channel.Config) > 0 {
		if err := json.Unmarshal(channel.Config, &cfg); err != nil {
			return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "config", "reason": "must be a json object"})
		}
	}
	require := func(key string) error {
		value, _ := cfg[key].(string)
		if strings.TrimSpace(value) == "" {
			return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "config." + key, "reason": "required"})
		}
		return nil
	}
	switch channel.Type {
	case escalation.ChannelTypeTelegram:
		return require("botToken")
	case escalation.ChannelTypeEmail:
		if err := require("apiKey"); err != nil {
			return err
		}
		return require("fromEmail")
	case escalation.ChannelTypeSMS:
		for _, key := range []string{"accountSid", "authToken", "fromNumber"} {
			if err := require(key); err != nil {
				return err
			}
		}
		return nil
	case escalation.ChannelTypeWebhook:
		return nil
	case escalation.ChannelTypePagerDuty:
		return require("routingKey")
	default:
		return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "type", "reason": "unsupported channel type"})
	}
}

func (h monitorHandler) createEscalationPolicy(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	policy, resp, ok := h.policyFromRequest(ctx, request, "")
	if !ok {
		return resp, nil
	}
	policy.PolicyID = newEscalationPolicyID(h.now())
	created, err := h.repo.CreateEscalationPolicy(ctx, policy)
	if err != nil {
		return respondAPIGateway(err)
	}
	location := fmt.Sprintf("/api/v1/escalation-policies/%s", created.PolicyID)
	return envelopeResponseWithLocation(http.StatusCreated, response.Ok(toEscalationPolicyResponse(created)), location)
}

func (h monitorHandler) listEscalationPolicies(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	policies, err := h.repo.ListEscalationPolicies(ctx, h.tenantID)
	if err != nil {
		return respondAPIGateway(err)
	}
	payload := listEscalationPoliciesResponse{Policies: make([]escalationPolicyResponse, 0, len(policies))}
	for _, policy := range policies {
		payload.Policies = append(payload.Policies, toEscalationPolicyResponse(policy))
	}
	return envelopeResponse(http.StatusOK, response.Ok(payload))
}

func (h monitorHandler) getEscalationPolicy(ctx context.Context, policyID string) (events.APIGatewayV2HTTPResponse, error) {
	policy, err := h.repo.GetEscalationPolicy(ctx, h.tenantID, policyID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if policy == nil {
		return respondAPIGateway(sharederrors.New(sharederrors.CodePolicyNotFound, nil))
	}
	return envelopeResponse(http.StatusOK, response.Ok(toEscalationPolicyResponse(*policy)))
}

func (h monitorHandler) updateEscalationPolicy(ctx context.Context, policyID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	existing, err := h.repo.GetEscalationPolicy(ctx, h.tenantID, policyID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if existing == nil {
		return respondAPIGateway(sharederrors.New(sharederrors.CodePolicyNotFound, nil))
	}
	policy, resp, ok := h.policyFromRequest(ctx, request, policyID)
	if !ok {
		return resp, nil
	}
	updated, err := h.repo.UpdateEscalationPolicy(ctx, policy)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toEscalationPolicyResponse(updated)))
}

func (h monitorHandler) deleteEscalationPolicy(ctx context.Context, policyID string) (events.APIGatewayV2HTTPResponse, error) {
	policy, err := h.repo.GetEscalationPolicy(ctx, h.tenantID, policyID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if policy == nil {
		return respondAPIGateway(sharederrors.New(sharederrors.CodePolicyNotFound, nil))
	}
	referenced, err := h.repo.ServiceReferencesEscalationPolicy(ctx, h.tenantID, policyID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if referenced {
		return respondAPIGateway(sharederrors.New(sharederrors.CodePolicyReferenced, nil))
	}
	if err := h.repo.DeleteEscalationPolicy(ctx, h.tenantID, policyID); err != nil {
		return respondAPIGateway(err)
	}
	return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusNoContent}, nil
}

func (h monitorHandler) getServiceEscalationPolicy(ctx context.Context, serviceID string) (events.APIGatewayV2HTTPResponse, error) {
	service, found, err := h.repo.GetService(ctx, h.tenantID, serviceID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeServiceNotFound, nil))
	}
	if strings.TrimSpace(service.EscalationPolicyID) == "" {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeServiceHasNoPolicy, nil))
	}
	policy, err := h.repo.GetEscalationPolicy(ctx, h.tenantID, service.EscalationPolicyID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if policy == nil {
		return respondAPIGateway(sharederrors.New(sharederrors.CodePolicyNotFound, nil))
	}
	return envelopeResponse(http.StatusOK, response.Ok(toEscalationPolicyResponse(*policy)))
}

func (h monitorHandler) policyFromRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest, policyID string) (escalation.EscalationPolicy, events.APIGatewayV2HTTPResponse, bool) {
	if requestHasInlineStepConfig(request.Body) {
		resp, _ := respondAPIGateway(sharederrors.New(sharederrors.CodeInlineChannelConfig, map[string]any{"detail": "steps must reference channels by channelId; remove target and config"}))
		return escalation.EscalationPolicy{}, resp, false
	}
	var payload escalationPolicyRequest
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		resp, _ := respondAPIGateway(sharederrors.Wrap(sharederrors.CodeInvalidJSON, err, nil))
		return escalation.EscalationPolicy{}, resp, false
	}
	if err := validateInput(payload); err != nil {
		resp, _ := respondAPIGateway(err)
		return escalation.EscalationPolicy{}, resp, false
	}
	policy := escalation.EscalationPolicy{
		TenantID:          h.tenantID,
		PolicyID:          policyID,
		Name:              strings.TrimSpace(payload.Name),
		Description:       strings.TrimSpace(payload.Description),
		BusinessHoursPath: escalationPathFromRequest(payload.BusinessHoursPath),
		OffHoursPath:      escalationPathFromRequest(payload.OffHoursPath),
	}
	if err := h.validateEscalationPolicy(ctx, policy); err != nil {
		resp, _ := respondAPIGateway(err)
		return escalation.EscalationPolicy{}, resp, false
	}
	return policy, events.APIGatewayV2HTTPResponse{}, true
}

func (h monitorHandler) validateEscalationPolicy(ctx context.Context, policy escalation.EscalationPolicy) error {
	for _, path := range []struct {
		name string
		path escalation.EscalationPath
	}{{name: "businessHoursPath", path: policy.BusinessHoursPath}, {name: "offHoursPath", path: policy.OffHoursPath}} {
		for idx, step := range path.path.Steps {
			channelID := strings.TrimSpace(step.ChannelID)
			if channelID == "" {
				return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": fmt.Sprintf("%s.steps[%d].channelId", path.name, idx), "reason": "required"})
			}
			channel, err := h.repo.GetNotificationChannel(ctx, policy.TenantID, channelID)
			if err != nil {
				return err
			}
			if channel == nil {
				return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": fmt.Sprintf("%s.steps[%d].channelId", path.name, idx), "reason": "channel not found"})
			}
		}
	}
	return nil
}

func escalationPathFromRequest(path escalationPathRequest) escalation.EscalationPath {
	steps := make([]escalation.EscalationStep, 0, len(path.Steps))
	for _, step := range path.Steps {
		steps = append(steps, escalation.EscalationStep{
			ChannelID:    strings.TrimSpace(step.ChannelID),
			DelayMinutes: step.DelayMinutes,
		})
	}
	return escalation.EscalationPath{Steps: steps}
}

func requestHasInlineStepConfig(body string) bool {
	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return false
	}
	for _, pathKey := range []string{"businessHoursPath", "offHoursPath"} {
		path, _ := payload[pathKey].(map[string]any)
		steps, _ := path["steps"].([]any)
		for _, stepValue := range steps {
			step, _ := stepValue.(map[string]any)
			if _, ok := step["target"]; ok {
				return true
			}
			if _, ok := step["config"]; ok {
				return true
			}
			if _, ok := step["channels"]; ok {
				return true
			}
		}
	}
	return false
}

func envelopeResponse[T any](statusCode int, env response.Envelope[T]) (events.APIGatewayV2HTTPResponse, error) {
	encoded, err := env.MarshalJSON()
	if err != nil {
		return respondAPIGateway(err)
	}
	return events.APIGatewayV2HTTPResponse{StatusCode: statusCode, Headers: map[string]string{"Content-Type": "application/json"}, Body: string(encoded)}, nil
}

func envelopeResponseWithLocation[T any](statusCode int, env response.Envelope[T], location string) (events.APIGatewayV2HTTPResponse, error) {
	encoded, err := env.MarshalJSON()
	if err != nil {
		return respondAPIGateway(err)
	}
	return events.APIGatewayV2HTTPResponse{StatusCode: statusCode, Headers: map[string]string{"Content-Type": "application/json", "Location": location}, Body: string(encoded)}, nil
}

func respondAPIGateway(err error) (events.APIGatewayV2HTTPResponse, error) {
	status, env := sharederrors.Respond(err)
	encoded, marshalErr := env.MarshalJSON()
	if marshalErr != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError, Headers: map[string]string{"Content-Type": "application/json"}, Body: `{"status":"error","reason":{"code":"INTERNAL","details":null}}`}, nil
	}
	return events.APIGatewayV2HTTPResponse{StatusCode: status, Headers: map[string]string{"Content-Type": "application/json"}, Body: string(encoded)}, nil
}
