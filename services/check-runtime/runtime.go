package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/probelocationcatalog"
	"bolt-monitor/shared/resultstatus"
	"github.com/aws/aws-lambda-go/events"
)

type runtimeRepository interface {
	GetSchedulerConfig(context.Context, string) (checkexecution.SchedulerConfig, error)
	ListMonitors(context.Context, string) ([]monitorconfig.Monitor, error)
	GetLastExecution(context.Context, string, string, string) (*time.Time, error)
	RecordLastExecution(context.Context, string, string, string, time.Time) error
	EnqueueExecutionRequests(context.Context, []checkexecution.ExecutionRequest, time.Time) error
	ListPendingExecutionWork(context.Context, string, int32) ([]checkexecution.ExecutionWork, error)
	ClaimExecutionWork(context.Context, checkexecution.ExecutionWork, time.Time) (bool, error)
	GetMonitor(context.Context, string, string, string) (monitorconfig.Monitor, bool, error)
	GetService(context.Context, string, string) (monitorconfig.Service, bool, error)
	MarkExecutionWorkSkipped(context.Context, checkexecution.ExecutionWork, time.Time, string) error
	RecordExecutionResult(context.Context, monitorconfig.Monitor, checkexecution.ExecutionWork, checkexecution.ExecutionResult) (string, string, error)
	GetMonitorStatus(context.Context, string, string, string) (resultstatus.MonitorStatus, bool, error)
}

type runtimeHandler struct {
	repo               runtimeRepository
	sqsClient          sqsClient
	queueURL           string
	escalationQueueURL string
	catalog            probelocationcatalog.Catalog
	tenantID           string
	mode               string
	now                func() time.Time
	newHTTP            func(time.Duration) *http.Client
}

type runtimeSummary struct {
	Mode           string `json:"mode"`
	Enqueued       int    `json:"enqueued,omitempty"`
	Processed      int    `json:"processed,omitempty"`
	Skipped        int    `json:"skipped,omitempty"`
	PendingScanned int    `json:"pendingScanned,omitempty"`
}

func newRuntimeHandler(repo runtimeRepository, sqsClient sqsClient, queueURL, escalationQueueURL string, catalog probelocationcatalog.Catalog, tenantID, mode string) runtimeHandler {
	return runtimeHandler{
		repo:               repo,
		sqsClient:          sqsClient,
		queueURL:           queueURL,
		escalationQueueURL: escalationQueueURL,
		catalog:            catalog,
		tenantID:           tenantID,
		mode:               strings.ToLower(strings.TrimSpace(mode)),
		now:                time.Now,
		newHTTP: func(timeout time.Duration) *http.Client {
			return &http.Client{Timeout: timeout}
		},
	}
}

func (h runtimeHandler) handle(ctx context.Context, _ events.CloudWatchEvent) (runtimeSummary, error) {
	switch h.mode {
	case modeScheduler:
		return h.runScheduler(ctx)
	case modeWorker:
		return h.runWorker(ctx)
	default:
		return runtimeSummary{}, fmt.Errorf("unsupported runtime mode %q", h.mode)
	}
}

func (h runtimeHandler) runScheduler(ctx context.Context) (runtimeSummary, error) {
	config, err := h.repo.GetSchedulerConfig(ctx, h.tenantID)
	if err != nil {
		return runtimeSummary{}, err
	}
	if !config.RecurringEnabled {
		return runtimeSummary{Mode: modeScheduler}, nil
	}
	monitors, err := h.repo.ListMonitors(ctx, h.tenantID)
	if err != nil {
		return runtimeSummary{}, err
	}

	filtered := make([]monitorconfig.Monitor, 0, len(monitors))
	for _, monitor := range monitors {
		if !monitor.Enabled {
			continue
		}
		status, found, err := h.repo.GetMonitorStatus(ctx, monitor.TenantID, monitor.ServiceID, monitor.MonitorID)
		if err != nil {
			return runtimeSummary{}, err
		}
		if found && strings.ToUpper(status.CurrentStatus) == string(resultstatus.MonitorStateMaintenance) {
			continue
		}
		filtered = append(filtered, monitor)
	}

	summary := runtimeSummary{Mode: modeScheduler}
	for _, monitor := range filtered {
		due, err := h.isMonitorDue(ctx, monitor)
		if err != nil {
			return summary, err
		}
		if !due {
			summary.Skipped++
			continue
		}
		executionMonitor := monitor
		if executionMonitor.IntervalSeconds <= 0 {
			executionMonitor.IntervalSeconds = 60
		}
		requests, err := checkexecution.BuildExecutionRequests([]monitorconfig.Monitor{executionMonitor}, h.catalog, checkexecution.TriggerTypeRecurring)
		if err != nil {
			return summary, err
		}
		for _, req := range requests {
			jsonReq, err := json.Marshal(req)
			if err != nil {
				return summary, err
			}
			if err := h.sqsClient.SendMessage(ctx, h.queueURL, string(jsonReq)); err != nil {
				return summary, err
			}
		}
		if err := h.repo.EnqueueExecutionRequests(ctx, requests, h.now()); err != nil {
			return runtimeSummary{}, err
		}
		if err := h.repo.RecordLastExecution(ctx, monitor.TenantID, monitor.ServiceID, monitor.MonitorID, h.now()); err != nil {
			return summary, err
		}
		summary.Enqueued += len(requests)
	}
	return summary, nil
}

func (h runtimeHandler) isMonitorDue(ctx context.Context, monitor monitorconfig.Monitor) (bool, error) {
	if monitor.IntervalSeconds <= 0 {
		return true, nil
	}
	lastExecution, err := h.repo.GetLastExecution(ctx, monitor.TenantID, monitor.ServiceID, monitor.MonitorID)
	if err != nil {
		return false, err
	}
	if lastExecution == nil {
		return true, nil
	}
	return h.now().Sub(*lastExecution) >= time.Duration(monitor.IntervalSeconds)*time.Second, nil
}

func (h runtimeHandler) runWorker(ctx context.Context) (runtimeSummary, error) {
	works, err := h.repo.ListPendingExecutionWork(ctx, h.tenantID, 50)
	if err != nil {
		return runtimeSummary{}, err
	}
	summary := runtimeSummary{Mode: modeWorker, PendingScanned: len(works)}
	for _, work := range works {
		claimed, err := h.repo.ClaimExecutionWork(ctx, work, h.now())
		if err != nil {
			return summary, err
		}
		if !claimed {
			continue
		}
		monitor, found, err := h.repo.GetMonitor(ctx, h.tenantID, work.ServiceID, work.MonitorID)
		if err != nil {
			return summary, err
		}
		if !found {
			if err := h.repo.MarkExecutionWorkSkipped(ctx, work, h.now(), "monitor not found"); err != nil {
				return summary, err
			}
			summary.Skipped++
			continue
		}
		request, skipReason, err := buildExecutionRequest(monitor, h.catalog, work)
		if err != nil {
			return summary, err
		}
		if skipReason != "" {
			if err := h.repo.MarkExecutionWorkSkipped(ctx, work, h.now(), skipReason); err != nil {
				return summary, err
			}
			summary.Skipped++
			continue
		}
		result := checkexecution.ExecuteHTTP(ctx, h.newHTTP(time.Duration(monitor.HTTP.TimeoutMs)*time.Millisecond), request)
		transition, incidentID, err := h.repo.RecordExecutionResult(ctx, monitor, work, result)
		if err != nil {
			return summary, err
		}
		if transition != "" && h.escalationQueueURL != "" {
			service, found, err := h.repo.GetService(ctx, h.tenantID, monitor.ServiceID)
			if err == nil && found {
				escalationMsg := buildEscalationMessage(transition, monitor, service, incidentID, result)
				if escalationMsg != "" {
					_ = h.sqsClient.SendMessage(ctx, h.escalationQueueURL, escalationMsg)
				}
			}
		}
		summary.Processed++
	}
	return summary, nil
}

func buildExecutionRequest(monitor monitorconfig.Monitor, catalog probelocationcatalog.Catalog, work checkexecution.ExecutionWork) (checkexecution.ExecutionRequest, string, error) {
	if err := monitor.Validate(); err != nil {
		return checkexecution.ExecutionRequest{}, fmt.Sprintf("monitor invalid: %v", err), nil
	}
	if !monitor.Enabled {
		return checkexecution.ExecutionRequest{}, "monitor disabled", nil
	}
	selected := false
	for _, locationID := range monitor.ProbeLocations {
		if strings.EqualFold(locationID, work.ProbeLocationID) {
			selected = true
			break
		}
	}
	if !selected {
		return checkexecution.ExecutionRequest{}, "probe location no longer selected", nil
	}
	for _, location := range catalog.Locations {
		if strings.EqualFold(location.LocationID, work.ProbeLocationID) {
			if !location.Enabled {
				return checkexecution.ExecutionRequest{}, "probe location disabled", nil
			}
			return checkexecution.ExecutionRequest{
				Monitor:       monitor,
				ProbeLocation: location,
				RunID:         work.RunID,
				Trigger:       work.Trigger,
			}, "", nil
		}
	}
	return checkexecution.ExecutionRequest{}, "probe location not found", nil
}

func (h runtimeHandler) handleSQSEvent(ctx context.Context, event events.SQSEvent) (runtimeSummary, error) {
	summary := runtimeSummary{Mode: modeWorker}
	for _, msg := range event.Records {
		var req checkexecution.ExecutionRequest
		if err := json.Unmarshal([]byte(msg.Body), &req); err != nil {
			return summary, err
		}
		result := checkexecution.ExecuteHTTP(ctx, h.newHTTP(time.Duration(req.Monitor.HTTP.TimeoutMs)*time.Millisecond), req)
		work := checkexecution.ExecutionWork{
			TenantID:        req.Monitor.TenantID,
			ServiceID:       req.Monitor.ServiceID,
			MonitorID:       req.Monitor.MonitorID,
			RunID:           req.RunID,
			ProbeLocationID: req.ProbeLocation.LocationID,
			Trigger:         req.Trigger,
			RequestedAt:     h.now(),
		}
		transition, incidentID, err := h.repo.RecordExecutionResult(ctx, req.Monitor, work, result)
		if err != nil {
			return summary, err
		}
		if transition != "" && h.escalationQueueURL != "" {
			service, found, err := h.repo.GetService(ctx, h.tenantID, req.Monitor.ServiceID)
			if err == nil && found {
				escalationMsg := buildEscalationMessage(transition, req.Monitor, service, incidentID, result)
				if escalationMsg != "" {
					_ = h.sqsClient.SendMessage(ctx, h.escalationQueueURL, escalationMsg)
				}
			}
		}
		summary.Processed++
	}
	return summary, nil
}

func sortWorksByRequestedAt(works []checkexecution.ExecutionWork) {
	sort.Slice(works, func(i, j int) bool { return works[i].RequestedAt.Before(works[j].RequestedAt) })
}

func buildEscalationMessage(transition string, monitor monitorconfig.Monitor, service monitorconfig.Service, incidentID string, result checkexecution.ExecutionResult) string {
	var message string
	timestamp := result.FinishedAt.UTC()
	if transition == "incident.down" {
		message = fmt.Sprintf("🚨 Incident Opened: %s is DOWN\nService: %s\nURL: %s\nError: %s\nTime: %s",
			monitor.Name, service.Name, monitor.HTTP.Target, result.Error, timestamp.Format(time.RFC3339))
	} else if transition == "incident.up" {
		message = fmt.Sprintf("✅ Incident Resolved: %s is UP\nService: %s\nURL: %s\nTime: %s",
			monitor.Name, service.Name, monitor.HTTP.Target, timestamp.Format(time.RFC3339))
	} else {
		return ""
	}
	notifEvent := map[string]interface{}{
		"eventType":   transition,
		"tenantId":    monitor.TenantID,
		"serviceId":   monitor.ServiceID,
		"monitorId":   monitor.MonitorID,
		"monitorName": monitor.Name,
		"serviceName": service.Name,
		"incidentId":  incidentID,
		"timestamp":   timestamp.Format(time.RFC3339),
		"message":     message,
	}
	data, err := json.Marshal(notifEvent)
	if err != nil {
		return ""
	}
	return string(data)
}
