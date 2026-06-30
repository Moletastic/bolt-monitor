package checkexecution

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/probelocationcatalog"
)

type TriggerType string

const (
	TriggerTypeManual    TriggerType = "manual"
	TriggerTypeRecurring TriggerType = "recurring"
)

type StopControlMode string

const (
	StopControlMonitorDisable StopControlMode = "monitor-disable"
	StopControlGlobalPause    StopControlMode = "global-pause"
)

type Outcome string

const (
	OutcomeSuccess Outcome = "success"
	OutcomeFailure Outcome = "failure"
	OutcomeTimeout Outcome = "timeout"
	OutcomeError   Outcome = "error"
)

type ExecutionRequest struct {
	Monitor       monitorconfig.Monitor         `json:"monitor"`
	ProbeLocation probelocationcatalog.Location `json:"probeLocation"`
	RunID         string                        `json:"runId,omitempty"`
	Trigger       TriggerType                   `json:"trigger"`
}

type ExecutionResult struct {
	ServiceID       string      `json:"serviceId"`
	MonitorID       string      `json:"monitorId"`
	TenantID        string      `json:"tenantId"`
	RunID           string      `json:"runId,omitempty"`
	Type            string      `json:"type"`
	ProbeLocationID string      `json:"probeLocationId"`
	Trigger         TriggerType `json:"trigger"`
	StartedAt       time.Time   `json:"startedAt"`
	FinishedAt      time.Time   `json:"finishedAt"`
	DurationMs      int64       `json:"durationMs"`
	Outcome         Outcome     `json:"outcome"`
	StatusCode      *int        `json:"statusCode,omitempty"`
	Error           string      `json:"error,omitempty"`
}

type SchedulerConfig struct {
	RecurringEnabled bool            `json:"recurringEnabled"`
	StopControlMode  StopControlMode `json:"stopControlMode,omitempty"`
}

func (c SchedulerConfig) Validate() error {
	if !c.RecurringEnabled {
		return nil
	}
	if c.StopControlMode != StopControlMonitorDisable && c.StopControlMode != StopControlGlobalPause {
		return fmt.Errorf("recurring execution requires reliable stop control")
	}
	return nil
}

func BuildExecutionRequests(monitors []monitorconfig.Monitor, catalog probelocationcatalog.Catalog, trigger TriggerType) ([]ExecutionRequest, error) {
	if err := catalog.Validate(); err != nil {
		return nil, err
	}
	requests := make([]ExecutionRequest, 0)
	for _, monitor := range monitors {
		if !monitor.Enabled {
			continue
		}
		if err := monitor.ValidateWithCatalog(catalog); err != nil {
			return nil, err
		}
		for _, locationID := range monitor.ProbeLocations {
			location, ok := lookupLocation(catalog, locationID)
			if !ok {
				return nil, fmt.Errorf("probe location %q not found in catalog", locationID)
			}
			requests = append(requests, ExecutionRequest{
				Monitor:       monitor,
				ProbeLocation: location,
				Trigger:       trigger,
			})
		}
	}
	return requests, nil
}

func ExecuteHTTP(ctx context.Context, client *http.Client, request ExecutionRequest) ExecutionResult {
	startedAt := time.Now().UTC()
	result := ExecutionResult{
		ServiceID:       request.Monitor.ServiceID,
		MonitorID:       request.Monitor.MonitorID,
		TenantID:        request.Monitor.TenantID,
		RunID:           request.RunID,
		Type:            string(request.Monitor.Type),
		ProbeLocationID: request.ProbeLocation.LocationID,
		Trigger:         request.Trigger,
		StartedAt:       startedAt,
	}

	httpConfig := request.Monitor.HTTP
	req, err := http.NewRequestWithContext(ctx, httpConfig.Method, httpConfig.Target, nil)
	if err != nil {
		result.Outcome = OutcomeError
		result.Error = err.Error()
		result.FinishedAt = time.Now().UTC()
		result.DurationMs = result.FinishedAt.Sub(startedAt).Milliseconds()
		return result
	}
	for key, value := range httpConfig.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	finishedAt := time.Now().UTC()
	result.FinishedAt = finishedAt
	result.DurationMs = finishedAt.Sub(startedAt).Milliseconds()
	if err != nil {
		result.Outcome = classifyError(err)
		result.Error = err.Error()
		return result
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Outcome = OutcomeError
		result.Error = err.Error()
		return result
	}
	statusCode := resp.StatusCode
	result.StatusCode = &statusCode
	if len(httpConfig.ExpectedStatusCodes) != 0 && !containsStatus(httpConfig.ExpectedStatusCodes, statusCode) {
		result.Outcome = OutcomeFailure
		result.Error = fmt.Sprintf("unexpected status code %d", statusCode)
		return result
	}
	if httpConfig.ExpectedBodyContains != nil && !strings.Contains(string(body), *httpConfig.ExpectedBodyContains) {
		result.Outcome = OutcomeFailure
		result.Error = fmt.Sprintf("response body missing expected content %q", *httpConfig.ExpectedBodyContains)
		return result
	}
	result.Outcome = OutcomeSuccess
	return result
}

func classifyError(err error) Outcome {
	if err == nil {
		return OutcomeSuccess
	}
	if ne, ok := err.(interface{ Timeout() bool }); ok && ne.Timeout() {
		return OutcomeTimeout
	}
	return OutcomeError
}

func lookupLocation(catalog probelocationcatalog.Catalog, locationID string) (probelocationcatalog.Location, bool) {
	for _, location := range catalog.Locations {
		if location.LocationID == locationID {
			return location, true
		}
	}
	return probelocationcatalog.Location{}, false
}

func containsStatus(list []int, value int) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
