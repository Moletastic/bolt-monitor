package checkexecution

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/outboundhttp"
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
	Monitor monitorconfig.Monitor `json:"monitor"`
	RunID   string                `json:"runId,omitempty"`
	Trigger TriggerType           `json:"trigger"`
}

type ExecutionResult struct {
	ServiceID   string      `json:"serviceId"`
	MonitorID   string      `json:"monitorId"`
	TenantID    string      `json:"tenantId"`
	RunID       string      `json:"runId,omitempty"`
	Type        string      `json:"type"`
	Trigger     TriggerType `json:"trigger"`
	StartedAt   time.Time   `json:"startedAt"`
	FinishedAt  time.Time   `json:"finishedAt"`
	DurationMs  int64       `json:"durationMs"`
	Outcome     Outcome     `json:"outcome"`
	StatusCode  *int        `json:"statusCode,omitempty"`
	Error       string      `json:"error,omitempty"`
	FailureCode string      `json:"failureCode,omitempty"`
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
		return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "stopControlMode", "reason": "recurring execution requires reliable stop control"})
	}
	return nil
}

func BuildExecutionRequests(monitors []monitorconfig.Monitor, trigger TriggerType) ([]ExecutionRequest, error) {
	requests := make([]ExecutionRequest, 0)
	for _, monitor := range monitors {
		if !monitor.Enabled {
			continue
		}
		if err := monitor.Validate(); err != nil {
			return nil, err
		}
		requests = append(requests, ExecutionRequest{Monitor: monitor, Trigger: trigger})
	}
	return requests, nil
}

type HTTPExecutor interface {
	Execute(context.Context, outboundhttp.Request) (outboundhttp.Response, error)
}

func ExecuteHTTP(ctx context.Context, executor HTTPExecutor, request ExecutionRequest) ExecutionResult {
	startedAt := time.Now().UTC()
	result := ExecutionResult{
		ServiceID: request.Monitor.ServiceID,
		MonitorID: request.Monitor.MonitorID,
		TenantID:  request.Monitor.TenantID,
		RunID:     request.RunID,
		Type:      string(request.Monitor.Type),
		Trigger:   request.Trigger,
		StartedAt: startedAt,
	}

	httpConfig := request.Monitor.HTTP
	headers := make(http.Header, len(httpConfig.Headers))
	for key, value := range httpConfig.Headers {
		headers.Set(key, value)
	}
	response, err := executor.Execute(ctx, outboundhttp.Request{
		Method:  httpConfig.Method,
		URL:     httpConfig.Target,
		Header:  headers,
		Timeout: time.Duration(httpConfig.TimeoutMs) * time.Millisecond,
	})
	finishedAt := time.Now().UTC()
	result.FinishedAt = finishedAt
	result.DurationMs = finishedAt.Sub(startedAt).Milliseconds()
	if err != nil {
		result.Outcome = classifyError(err)
		result.FailureCode = outboundFailureCode(err)
		result.Error = outboundhttp.SafeMessage(err)
		return result
	}
	statusCode := response.StatusCode
	result.StatusCode = &statusCode
	if len(httpConfig.ExpectedStatusCodes) != 0 && !containsStatus(httpConfig.ExpectedStatusCodes, statusCode) {
		result.Outcome = OutcomeFailure
		result.Error = fmt.Sprintf("unexpected status code %d", statusCode)
		return result
	}
	if httpConfig.ExpectedBodyContains != nil && !strings.Contains(string(response.Body), *httpConfig.ExpectedBodyContains) {
		result.Outcome = OutcomeFailure
		result.Error = fmt.Sprintf("response body missing expected content %q", *httpConfig.ExpectedBodyContains)
		return result
	}
	result.Outcome = OutcomeSuccess
	return result
}

func classifyError(err error) Outcome {
	if outboundhttp.IsKind(err, outboundhttp.KindTimeout) {
		return OutcomeTimeout
	}
	return OutcomeError
}

func outboundFailureCode(err error) string {
	var outbound *outboundhttp.Error
	if errors.As(err, &outbound) {
		return string(outbound.Kind)
	}
	return string(outboundhttp.KindTransport)
}

func containsStatus(list []int, value int) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
