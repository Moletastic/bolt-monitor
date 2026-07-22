package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bolt-monitor/shared/api/response"
	"bolt-monitor/shared/auth"
	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbschema"
	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/notifications"
	"bolt-monitor/shared/outboundhttp"
	"github.com/aws/aws-lambda-go/events"
)

const (
	defaultServiceIncidentsLimit = int32(5)
	maxServiceIncidentsLimit     = int32(50)
	historyPageSize              = int32(20)
)

// monitorAPIOperations is the explicit application composition boundary. Each
// embedded port is consumer-owned; the DynamoDB adapter happens to implement
// all of them, but handlers never depend on one aggregate repository interface.
type monitorAPIOperations struct {
	services  serviceOperations
	monitors  monitorOperations
	incidents incidentOperations
	scheduler schedulerOperations
	policies  escalationPolicyOperations
	channels  notificationChannelOperations
	search    searchResourcesQuery
}

func newMonitorAPIOperations(services serviceOperations, monitors monitorOperations, incidents incidentOperations, scheduler schedulerOperations, policies escalationPolicyOperations, channels notificationChannelOperations, search searchResourcesQuery) monitorAPIOperations {
	return monitorAPIOperations{
		services: services,
		monitors: monitors, incidents: incidents,
		scheduler: scheduler, policies: policies, channels: channels, search: search,
	}
}

type monitorHandler struct {
	operations          monitorAPIOperations
	principalResolver   PrincipalResolver
	membershipResolver  MembershipResolver
	securityEvents      securityEventEmitter
	newSecurityEvent    func(auth.SecurityEvent, string, auth.Subject, string) securityEvent
	tenantID            string
	now                 func() time.Time
	senders             notifications.SenderRegistry
	validateDestination func(context.Context, string) error
	executor            checkexecution.HTTPExecutor
}

type monitorHandlerDependencies struct {
	securityEvents      securityEventEmitter
	newSecurityEvent    func(auth.SecurityEvent, string, auth.Subject, string) securityEvent
	tenantID            string
	now                 func() time.Time
	senders             notifications.SenderRegistry
	executor            checkexecution.HTTPExecutor
	validateDestination func(context.Context, string) error
}

func newAuthorizedMonitorHandlerWithDependencies(operations monitorAPIOperations, principalResolver PrincipalResolver, membershipResolver MembershipResolver, dependencies monitorHandlerDependencies) monitorHandler {
	return monitorHandler{
		operations:          operations,
		principalResolver:   principalResolver,
		membershipResolver:  membershipResolver,
		securityEvents:      dependencies.securityEvents,
		newSecurityEvent:    dependencies.newSecurityEvent,
		tenantID:            dependencies.tenantID,
		now:                 dependencies.now,
		senders:             dependencies.senders,
		executor:            dependencies.executor,
		validateDestination: dependencies.validateDestination,
	}
}

func (h monitorHandler) handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if h.principalResolver == nil || h.membershipResolver == nil {
		return respondAPIGateway(authenticationRequired())
	}
	identity, err := h.principalResolver.Resolve(ctx, request)
	if err != nil {
		// Gateway-authenticated claims that cannot be safely normalized are always
		// an application authentication failure, without resolver diagnostics.
		return respondAPIGateway(authenticationRequired())
	}
	principal, err := h.membershipResolver.Resolve(ctx, identity)
	if err != nil {
		if typed, ok := sharederrors.As(err); ok && typed.Code == sharederrors.CodeAuthorizationDenied {
			// Do not disclose why a current membership check denied this subject.
			h.emitSecurityEvent(auth.EventAuthorizationDenied, "failure", identity.Subject, monitorCorrelationID(request.RequestContext.RequestID, requestHeader(request.Headers, "X-Correlation-ID")))
			return respondAPIGateway(authorizationDenied())
		}
		return respondAPIGateway(err)
	}

	// All domain operations use the tenant from the per-request authorized principal.
	h.tenantID = string(principal.TenantID)
	path := request.RawPath
	method := request.RequestContext.HTTP.Method
	serviceID := strings.TrimSpace(request.PathParameters["serviceId"])
	monitorID := strings.TrimSpace(request.PathParameters["monitorId"])
	incidentID := strings.TrimSpace(request.PathParameters["incidentId"])
	policyID := strings.TrimSpace(request.PathParameters["policyId"])
	channelID := strings.TrimSpace(request.PathParameters["channelId"])
	deliveryID := strings.TrimSpace(request.PathParameters["deliveryId"])
	if deliveryID == "" && strings.HasPrefix(path, "/api/v1/incidents/") && strings.Contains(path, "/deliveries/") {
		rest := strings.TrimPrefix(path, "/api/v1/incidents/"+incidentID+"/deliveries/")
		if idx := strings.Index(rest, "/"); idx >= 0 {
			deliveryID = strings.TrimSpace(rest[:idx])
		} else {
			deliveryID = strings.TrimSpace(rest)
		}
	}
	if channelID == "" && strings.HasPrefix(path, "/api/v1/notification-channels/") {
		channelID = strings.TrimPrefix(path, "/api/v1/notification-channels/")
		channelID = strings.TrimSuffix(channelID, "/test")
	}
	switch {
	case method == http.MethodGet && path == "/api/v1/search":
		return h.searchResources(ctx, request)
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
	case method == http.MethodGet && strings.HasSuffix(path, "/deliveries") && incidentID != "" && deliveryID == "":
		return h.listIncidentDeliveries(ctx, incidentID, request)
	case method == http.MethodPost && strings.HasSuffix(path, "/replay") && incidentID != "" && deliveryID != "":
		return h.replayIncidentDelivery(ctx, incidentID, deliveryID, request)
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
		return h.runMonitor(ctx, serviceID, monitorID, request)
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

func requestHeader(headers map[string]string, name string) string {
	for key, value := range headers {
		if strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}

func (h monitorHandler) emitSecurityEvent(event auth.SecurityEvent, outcome string, subject auth.Subject, correlationID string) {
	if h.securityEvents != nil && h.newSecurityEvent != nil {
		h.securityEvents(h.newSecurityEvent(event, outcome, subject, correlationID))
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
	results, err := h.operations.search.Execute(ctx, h.tenantID, request.QueryStringParameters["q"], limit, types)
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
	created, err := h.operations.services.create.Execute(ctx, createServiceInput{TenantID: h.tenantID, Request: payload})
	if err != nil {
		return respondAPIGateway(err)
	}
	location := fmt.Sprintf("/api/v1/services/%s", created.ServiceID)
	return envelopeResponseWithLocation(http.StatusCreated, response.Ok(toServiceResponse(created)), location)
}

func (h monitorHandler) listServices(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	services, err := h.operations.services.list.Execute(ctx, h.tenantID)
	if err != nil {
		return respondAPIGateway(err)
	}
	payload := listServicesResponse{Services: make([]serviceResponse, 0, len(services))}
	for _, service := range services {
		serviceResponse := toServiceResponse(service.Service)
		serviceResponse.CardMetrics = &service.Metrics
		payload.Services = append(payload.Services, serviceResponse)
	}
	return envelopeResponse(http.StatusOK, response.OkPaginated(payload, 1, len(payload.Services), len(payload.Services)))
}

func (h monitorHandler) getService(ctx context.Context, serviceID string) (events.APIGatewayV2HTTPResponse, error) {
	detail, found, err := h.operations.services.get.Execute(ctx, h.tenantID, serviceID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeServiceNotFound, nil))
	}
	serviceResponse := toServiceResponse(detail.Service)
	serviceResponse.CardMetrics = &detail.Metrics
	serviceResponse.Monitors = make([]monitorResponse, 0, len(detail.Monitors))
	for _, monitor := range detail.Monitors {
		var statusResponse *monitorStatusResponse
		if monitor.Status != nil {
			statusResponse = toStatusResponse(*monitor.Status)
		}
		serviceResponse.Monitors = append(serviceResponse.Monitors, toMonitorResponse(monitor.Monitor, statusResponse))
	}
	return envelopeResponse(http.StatusOK, response.Ok(serviceResponse))
}

func (h monitorHandler) updateService(ctx context.Context, serviceID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var payload updateServiceRequest
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return respondAPIGateway(sharederrors.Wrap(sharederrors.CodeInvalidJSON, err, nil))
	}
	stored, err := h.operations.services.update.Execute(ctx, updateServiceInput{TenantID: h.tenantID, ServiceID: serviceID, Request: payload})
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toServiceResponse(stored)))
}

func (h monitorHandler) deleteService(ctx context.Context, serviceID string) (events.APIGatewayV2HTTPResponse, error) {
	deleted, err := h.operations.services.delete.Execute(ctx, h.tenantID, serviceID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !deleted {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeServiceNotFound, nil))
	}
	return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusNoContent}, nil
}

func (h monitorHandler) archiveService(ctx context.Context, serviceID string) (events.APIGatewayV2HTTPResponse, error) {
	service, err := h.operations.services.archive.Execute(ctx, h.tenantID, serviceID)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toServiceResponse(service)))
}

func (h monitorHandler) reactivateService(ctx context.Context, serviceID string) (events.APIGatewayV2HTTPResponse, error) {
	service, err := h.operations.services.reactivate.Execute(ctx, h.tenantID, serviceID)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toServiceResponse(service)))
}

func (h monitorHandler) createMonitor(ctx context.Context, serviceID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var payload monitorconfig.CreateMonitorRequest
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return respondAPIGateway(sharederrors.Wrap(sharederrors.CodeInvalidJSON, err, nil))
	}
	command := h.operations.monitors.create
	command.validateDestination = h.validateMonitorDestination
	created, err := command.Execute(ctx, createMonitorInput{TenantID: h.tenantID, ServiceID: serviceID, Request: payload})
	if err != nil {
		return respondAPIGateway(err)
	}
	location := fmt.Sprintf("/api/v1/services/%s/monitors/%s", serviceID, created.MonitorID)
	return envelopeResponseWithLocation(http.StatusCreated, response.Ok(toMonitorResponse(created, toUnknownStatusResponse())), location)
}

func (h monitorHandler) listMonitors(ctx context.Context, serviceID string) (events.APIGatewayV2HTTPResponse, error) {
	monitors, found, err := h.operations.monitors.list.Execute(ctx, h.tenantID, serviceID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeServiceNotFound, nil))
	}
	payload := listMonitorsResponse{Monitors: make([]monitorResponse, 0, len(monitors))}
	for _, monitor := range monitors {
		var statusResponse *monitorStatusResponse
		if monitor.Status != nil {
			statusResponse = toStatusResponse(*monitor.Status)
		}
		payload.Monitors = append(payload.Monitors, toMonitorResponse(monitor.Monitor, statusResponse))
	}
	return envelopeResponse(http.StatusOK, response.OkPaginated(payload, 1, len(payload.Monitors), len(payload.Monitors)))
}

func (h monitorHandler) getMonitor(ctx context.Context, serviceID, monitorID string) (events.APIGatewayV2HTTPResponse, error) {
	detail, found, err := h.operations.monitors.get.Execute(ctx, h.tenantID, serviceID, monitorID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeMonitorNotFound, nil))
	}
	var statusResponse *monitorStatusResponse
	if detail.Status != nil {
		statusResponse = toStatusResponse(*detail.Status)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toMonitorResponse(detail.Monitor, statusResponse)))
}

func (h monitorHandler) updateMonitor(ctx context.Context, serviceID, monitorID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var payload updateMonitorRequest
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return respondAPIGateway(sharederrors.Wrap(sharederrors.CodeInvalidJSON, err, nil))
	}
	command := h.operations.monitors.update
	command.validateDestination = h.validateMonitorDestination
	stored, err := command.Execute(ctx, updateMonitorInput{TenantID: h.tenantID, ServiceID: serviceID, MonitorID: monitorID, Request: payload})
	if err != nil {
		return respondAPIGateway(err)
	}
	status, foundStatus, err := h.operations.monitors.status.Execute(ctx, h.tenantID, serviceID, monitorID)
	if err != nil {
		return respondAPIGateway(err)
	}
	var statusResponse *monitorStatusResponse
	if foundStatus {
		statusResponse = toStatusResponse(status)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toMonitorResponse(stored, statusResponse)))
}

func (h monitorHandler) validateMonitorDestination(ctx context.Context, monitor monitorconfig.Monitor) error {
	if h.validateDestination == nil || monitor.HTTP == nil {
		return nil
	}
	if err := h.validateDestination(ctx, monitor.HTTP.Target); err != nil {
		return sharederrors.Wrap(sharederrors.CodeValidationFailed, nil, map[string]any{
			"field":  "http.target",
			"reason": outboundhttp.SafeMessage(err),
		})
	}
	return nil
}

func (h monitorHandler) deleteMonitor(ctx context.Context, serviceID, monitorID string) (events.APIGatewayV2HTTPResponse, error) {
	deleted, err := h.operations.monitors.delete.Execute(ctx, h.tenantID, serviceID, monitorID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !deleted {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeMonitorNotFound, nil))
	}
	return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusNoContent}, nil
}

func (h monitorHandler) setMonitorEnabled(ctx context.Context, serviceID, monitorID string, enabled bool) (events.APIGatewayV2HTTPResponse, error) {
	monitor, found, err := h.operations.monitors.enable.Execute(ctx, h.tenantID, serviceID, monitorID, enabled)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeMonitorNotFound, nil))
	}
	status, foundStatus, err := h.operations.monitors.status.Execute(ctx, h.tenantID, serviceID, monitorID)
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
	status, found, err := h.operations.monitors.maintenance.Execute(ctx, h.tenantID, serviceID, monitorID, enabled)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeMonitorNotFound, nil))
	}
	return envelopeResponse(http.StatusOK, response.Ok(*toStatusResponse(status)))
}

func (h monitorHandler) getMonitorStatus(ctx context.Context, serviceID, monitorID string) (events.APIGatewayV2HTTPResponse, error) {
	status, found, err := h.operations.monitors.status.Execute(ctx, h.tenantID, serviceID, monitorID)
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
	page, err := h.operations.monitors.runs.Execute(ctx, h.tenantID, serviceID, monitorID, historyPageSize, startKey)
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

func (h monitorHandler) runMonitor(ctx context.Context, serviceID, monitorID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	idempotencyKey, err := parseIdempotencyKey(requestHeader(request.Headers, "Idempotency-Key"))
	if err != nil {
		return respondAPIGateway(err)
	}
	command := h.operations.monitors.manualRun
	command.executor = h.executor
	result, err := command.Execute(ctx, h.tenantID, serviceID, monitorID, idempotencyKey)
	if err != nil {
		return respondAPIGateway(err)
	}
	if result.Existing != nil {
		return envelopeResponse(http.StatusOK, response.Ok(manualRunResponseFromRecord(*result.Existing)))
	}
	return envelopeResponse(http.StatusOK, response.Ok(toManualRunResponseWithResult(result.Record, *result.Execution)))
}

func (h monitorHandler) listIncidents(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	status := strings.ToLower(strings.TrimSpace(request.QueryStringParameters["status"]))
	incidents, err := h.operations.incidents.list.Execute(ctx, h.tenantID, status)
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
	incident, found, err := h.operations.incidents.get.Execute(ctx, h.tenantID, incidentID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeIncidentNotFound, nil))
	}
	return envelopeResponse(http.StatusOK, response.Ok(toIncidentResponse(incident)))
}

func (h monitorHandler) getEscalationState(ctx context.Context, incidentID string) (events.APIGatewayV2HTTPResponse, error) {
	state, found, err := h.operations.incidents.escalationState.Execute(ctx, h.tenantID, incidentID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeIncidentNotFound, nil))
	}
	if state == nil {
		return envelopeResponse(http.StatusOK, response.Ok(escalationStateResponse{Exists: false}))
	}
	return envelopeResponse(http.StatusOK, response.Ok(toEscalationStateResponse(*state)))
}

func (h monitorHandler) getIncidentActivities(ctx context.Context, incidentID string) (events.APIGatewayV2HTTPResponse, error) {
	activities, found, err := h.operations.incidents.activities.Execute(ctx, h.tenantID, incidentID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeIncidentNotFound, nil))
	}
	payload := incidentActivitiesResponse{Activities: make([]incidentActivityResponse, 0, len(activities))}
	for _, activity := range activities {
		payload.Activities = append(payload.Activities, toIncidentActivityResponse(activity))
	}
	return envelopeResponse(http.StatusOK, response.Ok(payload))
}

func (h monitorHandler) getMonitorIncidents(ctx context.Context, serviceID, monitorID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	resource := dynamodbschema.MonitorPK(h.tenantID, serviceID, monitorID)
	startKey, err := historyStartKey(request, resource, "PK")
	if err != nil {
		return respondAPIGateway(err)
	}
	page, found, err := h.operations.incidents.monitorIncidents.Execute(ctx, h.tenantID, serviceID, monitorID, historyPageSize, startKey)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeMonitorNotFound, nil))
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
	incidents, found, err := h.operations.incidents.serviceIncidents.Execute(ctx, h.tenantID, serviceID, limit)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeServiceNotFound, nil))
	}
	payload := listIncidentsResponse{Incidents: make([]incidentResponse, 0, len(incidents))}
	for _, incident := range incidents {
		payload.Incidents = append(payload.Incidents, toIncidentResponse(incident))
	}
	return envelopeResponse(http.StatusOK, response.Ok(payload))
}

func (h monitorHandler) acknowledgeIncident(ctx context.Context, incidentID string) (events.APIGatewayV2HTTPResponse, error) {
	incident, found, err := h.operations.incidents.acknowledge.Execute(ctx, h.tenantID, incidentID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeIncidentNotFound, nil))
	}
	return envelopeResponse(http.StatusOK, response.Ok(toIncidentResponse(incident)))
}

func (h monitorHandler) resolveIncident(ctx context.Context, incidentID string) (events.APIGatewayV2HTTPResponse, error) {
	incident, found, err := h.operations.incidents.resolve.Execute(ctx, h.tenantID, incidentID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeIncidentNotFound, nil))
	}
	return envelopeResponse(http.StatusOK, response.Ok(toIncidentResponse(incident)))
}

func (h monitorHandler) getSchedulerConfig(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	config, err := h.operations.scheduler.get.Execute(ctx, h.tenantID)
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
	updated, err := h.operations.scheduler.update.Execute(ctx, h.tenantID, config)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toSchedulerConfigResponse(updated)))
}

func (h monitorHandler) getMonitorAudit(ctx context.Context, serviceID, monitorID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	resource := dynamodbschema.AuditResourceItem(h.tenantID, serviceID, monitorID, "cursor", "cursor").GSI3PK
	startKey, err := historyStartKey(request, resource, "GSI3PK")
	if err != nil {
		return respondAPIGateway(err)
	}
	page, found, err := h.operations.monitors.audit.Execute(ctx, h.tenantID, serviceID, monitorID, historyPageSize, startKey)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeMonitorNotFound, nil))
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
	resource := dynamodbschema.AuditResourceItem(h.tenantID, serviceID, "", "cursor", "cursor").GSI3PK
	startKey, err := historyStartKey(request, resource, "GSI3PK")
	if err != nil {
		return respondAPIGateway(err)
	}
	page, found, err := h.operations.services.audit.Execute(ctx, h.tenantID, serviceID, historyPageSize, startKey)
	if err != nil {
		return respondAPIGateway(err)
	}
	if !found {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeServiceNotFound, nil))
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

func (h monitorHandler) createNotificationChannel(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	channel, resp, ok := h.notificationChannelFromRequest(request, nil)
	if !ok {
		return resp, nil
	}
	channel.TenantID = h.tenantID
	if err := h.validateChannelDestination(ctx, channel); err != nil {
		return respondAPIGateway(err)
	}
	created, err := h.operations.channels.create.Execute(ctx, channel)
	if err != nil {
		return respondAPIGateway(err)
	}
	location := fmt.Sprintf("/api/v1/notification-channels/%s", created.ChannelID)
	return envelopeResponseWithLocation(http.StatusCreated, response.Ok(toNotificationChannelResponse(created)), location)
}

func (h monitorHandler) listNotificationChannels(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	channels, err := h.operations.channels.list.Execute(ctx, h.tenantID)
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
	channel, err := h.operations.channels.get.Execute(ctx, h.tenantID, channelID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if channel == nil {
		return respondAPIGateway(sharederrors.New(sharederrors.CodeChannelNotFound, nil))
	}
	return envelopeResponse(http.StatusOK, response.Ok(toNotificationChannelResponse(*channel)))
}

func (h monitorHandler) updateNotificationChannel(ctx context.Context, channelID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	current, err := h.operations.channels.get.Execute(ctx, h.tenantID, channelID)
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
	if err := h.validateChannelDestination(ctx, channel); err != nil {
		return respondAPIGateway(err)
	}
	updated, err := h.operations.channels.update.Execute(ctx, channel)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toNotificationChannelResponse(updated)))
}

func (h monitorHandler) deleteNotificationChannel(ctx context.Context, channelID string) (events.APIGatewayV2HTTPResponse, error) {
	references, err := h.operations.channels.delete.Execute(ctx, h.tenantID, channelID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if len(references) > 0 {
		return envelopeResponse(http.StatusConflict, response.Ok[channelInUseResponse](channelInUseResponse{Error: "channel in use", ReferencingRoutes: references}))
	}
	if err != nil {
		return respondAPIGateway(err)
	}
	return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusNoContent}, nil
}

func (h monitorHandler) testNotificationChannel(ctx context.Context, channelID string) (events.APIGatewayV2HTTPResponse, error) {
	command := h.operations.channels.test
	command.senders = h.senders
	command.now = h.now
	result, err := command.Execute(ctx, h.tenantID, channelID)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(notificationChannelTestResponse{ChannelID: result.ChannelID, SentAt: result.SentAt.Format(time.RFC3339)}, "Test notification sent."))
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

func sanitizeNotificationDeliveryError(reason error, _ json.RawMessage) string {
	var outbound *outboundhttp.Error
	if errors.As(reason, &outbound) {
		return outboundhttp.SafeMessage(outbound)
	}
	return "notification delivery failed"
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
		if _, err := outboundhttp.ValidateURL(channel.Target); err != nil {
			return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "target", "reason": outboundhttp.SafeMessage(err)})
		}
		return nil
	case escalation.ChannelTypePagerDuty:
		return require("routingKey")
	default:
		return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "type", "reason": "unsupported channel type"})
	}
}

func (h monitorHandler) validateChannelDestination(ctx context.Context, channel escalation.NotificationChannel) error {
	if h.validateDestination == nil {
		return nil
	}
	target := ""
	field := ""
	if channel.Type == escalation.ChannelTypeWebhook {
		target, field = channel.Target, "target"
	} else if channel.Type == escalation.ChannelTypeEmail || channel.Type == escalation.ChannelTypeSMS || channel.Type == escalation.ChannelTypePagerDuty {
		var config map[string]any
		if err := json.Unmarshal(channel.Config, &config); err != nil {
			return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "config", "reason": "must be a json object"})
		}
		if value, ok := config["apiBaseUrl"].(string); ok {
			target = strings.TrimSpace(value)
			field = "config.apiBaseUrl"
		}
	}
	if target == "" {
		return nil
	}
	if _, err := outboundhttp.ValidateURL(target); err != nil {
		return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": field, "reason": outboundhttp.SafeMessage(err)})
	}
	if err := h.validateDestination(ctx, target); err != nil {
		return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": field, "reason": outboundhttp.SafeMessage(err)})
	}
	return nil
}

func (h monitorHandler) createEscalationPolicy(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	policy, resp, ok := h.policyFromRequest(ctx, request, "")
	if !ok {
		return resp, nil
	}
	created, err := h.operations.policies.create.Execute(ctx, policy)
	if err != nil {
		return respondAPIGateway(err)
	}
	location := fmt.Sprintf("/api/v1/escalation-policies/%s", created.PolicyID)
	return envelopeResponseWithLocation(http.StatusCreated, response.Ok(toEscalationPolicyResponse(created)), location)
}

func (h monitorHandler) listEscalationPolicies(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	policies, err := h.operations.policies.list.Execute(ctx, h.tenantID)
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
	policy, err := h.operations.policies.get.Execute(ctx, h.tenantID, policyID)
	if err != nil {
		return respondAPIGateway(err)
	}
	if policy == nil {
		return respondAPIGateway(sharederrors.New(sharederrors.CodePolicyNotFound, nil))
	}
	return envelopeResponse(http.StatusOK, response.Ok(toEscalationPolicyResponse(*policy)))
}

func (h monitorHandler) updateEscalationPolicy(ctx context.Context, policyID string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	policy, resp, ok := h.policyFromRequest(ctx, request, policyID)
	if !ok {
		return resp, nil
	}
	updated, err := h.operations.policies.update.Execute(ctx, policy)
	if err != nil {
		return respondAPIGateway(err)
	}
	return envelopeResponse(http.StatusOK, response.Ok(toEscalationPolicyResponse(updated)))
}

func (h monitorHandler) deleteEscalationPolicy(ctx context.Context, policyID string) (events.APIGatewayV2HTTPResponse, error) {
	if err := h.operations.policies.delete.Execute(ctx, h.tenantID, policyID); err != nil {
		return respondAPIGateway(err)
	}
	return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusNoContent}, nil
}

func (h monitorHandler) getServiceEscalationPolicy(ctx context.Context, serviceID string) (events.APIGatewayV2HTTPResponse, error) {
	policy, err := h.operations.policies.service.Execute(ctx, h.tenantID, serviceID)
	if err != nil {
		return respondAPIGateway(err)
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
	return policy, events.APIGatewayV2HTTPResponse{}, true
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
