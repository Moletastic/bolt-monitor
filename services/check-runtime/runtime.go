package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/outboundhttp"
	"bolt-monitor/shared/resultstatus"
	"github.com/aws/aws-lambda-go/events"
)

type runtimeRepository interface {
	executionResultPersistence
	GetSchedulerConfig(context.Context, string) (checkexecution.SchedulerConfig, error)
	ListMonitors(context.Context, string) ([]monitorconfig.Monitor, error)
	GetLastExecution(context.Context, string, string, string) (*time.Time, error)
	RecordLastExecution(context.Context, string, string, string, time.Time) error
	EnqueueExecutionRequests(context.Context, []checkexecution.ExecutionRequest, time.Time) (int, error)
	AcknowledgeExecutionPublication(context.Context, checkexecution.ExecutionWork) error
	LoadExecutionWork(context.Context, string, string) (checkexecution.ExecutionWork, bool, error)
	ListPendingExecutionWork(context.Context, string, int32) ([]checkexecution.ExecutionWork, error)
	ListPublicationMarkers(context.Context, string, string, int32, map[string]sharedaws.AttributeValue) ([]dynamodbrecord.ExecutionMarkerRecord, map[string]sharedaws.AttributeValue, error)
	ListDispatchPending(context.Context, string, string, int32, map[string]sharedaws.AttributeValue) ([]dynamodbrecord.DispatchPendingRecord, map[string]sharedaws.AttributeValue, error)
	RemoveDispatchPending(context.Context, string, string, string, string) error
	LoadTransitionOutbox(context.Context, string, string) (dynamodbrecord.TransitionOutboxRecord, bool, error)
	ClaimExecutionWork(context.Context, checkexecution.ExecutionWork, time.Time) (checkexecution.ExecutionWork, bool, error)
	GetMonitor(context.Context, string, string, string) (monitorconfig.Monitor, bool, error)
	GetService(context.Context, string, string) (monitorconfig.Service, bool, error)
	MarkExecutionWorkSkipped(context.Context, checkexecution.ExecutionWork, time.Time, string) error
	GetMonitorStatus(context.Context, string, string, string) (resultstatus.MonitorStatus, bool, error)
}

type runtimeHandler struct {
	repo               runtimeRepository
	sqsClient          sqsClient
	queueURL           string
	escalationQueueURL string
	tenantID           string
	mode               string
	now                func() time.Time
	executor           checkexecution.HTTPExecutor
	resultCommand      executionResultCommand
	schedulerDeadline  time.Duration
}

type runtimeSummary struct {
	Mode           string `json:"mode"`
	Enqueued       int    `json:"enqueued,omitempty"`
	Processed      int    `json:"processed,omitempty"`
	Skipped        int    `json:"skipped,omitempty"`
	PendingScanned int    `json:"pendingScanned,omitempty"`
}

func (h runtimeHandler) handleSQSEventBatch(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
	response := events.SQSEventResponse{BatchItemFailures: []events.SQSBatchItemFailure{}}
	for _, record := range event.Records {
		_, err := h.handleSQSEvent(ctx, events.SQSEvent{Records: []events.SQSMessage{record}})
		if err == nil || isTerminalSQSError(err) {
			continue
		}
		response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: record.MessageId})
	}
	return response, nil
}

func isTerminalSQSError(err error) bool {
	var runtimeFailure *checkexecution.RuntimeFailure
	if !errors.As(err, &runtimeFailure) {
		return false
	}
	return !runtimeFailure.Retryable && runtimeFailure.Code != checkexecution.FailureMalformedEnvelope
}

const defaultSchedulerDeadline = 50 * time.Second

type runtimeHandlerDependencies struct {
	now               func() time.Time
	executor          checkexecution.HTTPExecutor
	resultClock       executionResultClock
	resultIDs         executionResultIDs
	schedulerDeadline time.Duration
}

func newRuntimeHandlerWithDependencies(repo runtimeRepository, sqsClient sqsClient, queueURL, escalationQueueURL, tenantID, mode string, dependencies runtimeHandlerDependencies) runtimeHandler {
	return runtimeHandler{
		repo:               repo,
		sqsClient:          sqsClient,
		queueURL:           queueURL,
		escalationQueueURL: escalationQueueURL,
		tenantID:           tenantID,
		mode:               strings.ToLower(strings.TrimSpace(mode)),
		now:                dependencies.now,
		executor:           dependencies.executor,
		resultCommand:      newExecutionResultCommand(repo, dependencies.resultClock, dependencies.resultIDs),
		schedulerDeadline:  dependencies.schedulerDeadline,
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
	if _, err := h.recoverPublicationMarkers(ctx); err != nil {
		return runtimeSummary{}, err
	}
	if _, err := h.reconcileDispatchPending(ctx); err != nil {
		return runtimeSummary{}, err
	}
	deadline := h.schedulerDeadline
	if deadline <= 0 {
		deadline = defaultSchedulerDeadline
	}
	discoveryCtx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()
	monitors, err := h.repo.ListMonitors(discoveryCtx, h.tenantID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return runtimeSummary{Mode: modeScheduler}, checkexecution.Publication("scheduler-deadline", "")
		}
		return runtimeSummary{}, err
	}

	filtered := make([]monitorconfig.Monitor, 0, len(monitors))
	for _, monitor := range monitors {
		if !monitor.Enabled {
			continue
		}
		status, found, err := h.repo.GetMonitorStatus(discoveryCtx, monitor.TenantID, monitor.ServiceID, monitor.MonitorID)
		if err != nil {
			return runtimeSummary{}, err
		}
		if found && strings.ToUpper(status.CurrentStatus) == string(resultstatus.MonitorStateMaintenance) {
			continue
		}
		filtered = append(filtered, monitor)
	}

	summary := runtimeSummary{Mode: modeScheduler}
	acceptedAt := h.now().UTC()
	for _, monitor := range filtered {
		executionMonitor := monitor
		if executionMonitor.IntervalSeconds <= 0 {
			executionMonitor.IntervalSeconds = 60
		}
		scheduledFor := checkexecution.ScheduledFor(acceptedAt, executionMonitor.IntervalSeconds)
		scheduleDefinitionVersion := checkexecution.ScheduleDefinitionVersion(executionMonitor)
		request := checkexecution.ExecutionRequest{
			Monitor:                   executionMonitor,
			RunID:                     checkexecution.RecurringRunID(executionMonitor.TenantID, executionMonitor.ServiceID, executionMonitor.MonitorID, scheduleDefinitionVersion, scheduledFor),
			Trigger:                   checkexecution.TriggerTypeRecurring,
			AcceptedAt:                acceptedAt,
			ScheduleDefinitionVersion: scheduleDefinitionVersion,
			ScheduledFor:              &scheduledFor,
		}
		// Durable work is authority. SQS only wakes a worker for this identity.
		created, err := h.repo.EnqueueExecutionRequests(ctx, []checkexecution.ExecutionRequest{request}, acceptedAt)
		if err != nil {
			runtimeOutcomeLog(outcomePublicationErr, "persist-work", request.Monitor.TenantID, request.RunID, request.Monitor.MonitorID, err.Error())
			return summary, checkexecution.Storage("persist-work", request.RunID)
		}
		if created == 0 {
			runtimeOutcomeLog(outcomeExisting, "persist-work", request.Monitor.TenantID, request.RunID, request.Monitor.MonitorID, "")
			continue
		}
		runtimeOutcomeLog(outcomeCreated, "persist-work", request.Monitor.TenantID, request.RunID, request.Monitor.MonitorID, "")
		jsonReq, err := json.Marshal(request)
		if err != nil {
			return summary, err
		}
		if err := h.sqsClient.SendMessage(ctx, h.queueURL, string(jsonReq)); err != nil {
			runtimeOutcomeLog(outcomePublicationErr, "publish-work", request.Monitor.TenantID, request.RunID, request.Monitor.MonitorID, err.Error())
			return summary, checkexecution.Publication("publish-work", request.RunID)
		}
		if err := h.repo.AcknowledgeExecutionPublication(ctx, checkexecution.ExecutionWork{
			TenantID: request.Monitor.TenantID, ServiceID: request.Monitor.ServiceID, MonitorID: request.Monitor.MonitorID,
			RunID: request.RunID, Trigger: request.Trigger, AcceptedAt: request.AcceptedAt, RequestedAt: request.AcceptedAt,
			ScheduleDefinitionVersion: request.ScheduleDefinitionVersion, ScheduledFor: request.ScheduledFor,
		}); err != nil {
			return summary, checkexecution.Publication("acknowledge-publication", request.RunID)
		}
		summary.Enqueued++
	}
	return summary, nil
}

func (h runtimeHandler) runWorker(ctx context.Context) (runtimeSummary, error) {
	works, err := h.repo.ListPendingExecutionWork(ctx, h.tenantID, 50)
	if err != nil {
		return runtimeSummary{}, err
	}
	summary := runtimeSummary{Mode: modeWorker, PendingScanned: len(works)}
	for _, work := range works {
		claimedWork, claimed, err := h.repo.ClaimExecutionWork(ctx, work, h.now())
		if err != nil {
			return summary, err
		}
		if !claimed {
			continue
		}
		work = claimedWork
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
		if skipReason, err := h.currentWorkSkipReason(ctx, monitor, work); err != nil {
			return summary, err
		} else if skipReason != "" {
			if err := h.repo.MarkExecutionWorkSkipped(ctx, work, h.now(), skipReason); err != nil {
				return summary, err
			}
			summary.Skipped++
			continue
		}
		request, skipReason, err := buildExecutionRequest(monitor, work)
		if err != nil {
			return summary, err
		}
		if skipReason != "" {
			if strings.HasPrefix(skipReason, "monitor invalid:") {
				now := h.now().UTC()
				failureKind := outboundhttp.KindInvalidURL
				if monitor.HTTP != nil {
					if _, err := outboundhttp.ValidateURL(monitor.HTTP.Target); err != nil {
						var outbound *outboundhttp.Error
						if errors.As(err, &outbound) {
							failureKind = outbound.Kind
						}
					}
				}
				result := checkexecution.ExecutionResult{
					ServiceID:   monitor.ServiceID,
					MonitorID:   monitor.MonitorID,
					TenantID:    monitor.TenantID,
					RunID:       work.RunID,
					Type:        string(monitor.Type),
					Trigger:     work.Trigger,
					StartedAt:   now,
					FinishedAt:  now,
					Outcome:     checkexecution.OutcomeError,
					FailureCode: string(failureKind),
					Error:       outboundhttp.SafeMessage(&outboundhttp.Error{Kind: failureKind}),
				}
				if _, _, err := h.resultCommand.execute(ctx, monitor, work, result); err != nil {
					return summary, err
				}
				summary.Processed++
				continue
			}
			if err := h.repo.MarkExecutionWorkSkipped(ctx, work, h.now(), skipReason); err != nil {
				return summary, err
			}
			summary.Skipped++
			continue
		}
		result := checkexecution.ExecuteHTTP(ctx, h.executor, request)
		_, _, err = h.resultCommand.execute(ctx, monitor, work, result)
		if err != nil {
			return summary, err
		}
		summary.Processed++
	}
	return summary, nil
}

func buildExecutionRequest(monitor monitorconfig.Monitor, work checkexecution.ExecutionWork) (checkexecution.ExecutionRequest, string, error) {
	if err := monitor.Validate(); err != nil {
		return checkexecution.ExecutionRequest{}, fmt.Sprintf("monitor invalid: %v", err), nil
	}
	if !monitor.Enabled {
		return checkexecution.ExecutionRequest{}, "monitor disabled", nil
	}
	return checkexecution.ExecutionRequest{Monitor: monitor, RunID: work.RunID, Trigger: work.Trigger, AcceptedAt: work.AcceptedAt, ScheduleDefinitionVersion: work.ScheduleDefinitionVersion, ScheduledFor: work.ScheduledFor}, "", nil
}

func (h runtimeHandler) handleSQSEvent(ctx context.Context, event events.SQSEvent) (runtimeSummary, error) {
	summary := runtimeSummary{Mode: modeWorker}
	for _, msg := range event.Records {
		var req checkexecution.ExecutionRequest
		if err := json.Unmarshal([]byte(msg.Body), &req); err != nil {
			return summary, checkexecution.NewRuntimeFailure(checkexecution.FailureMalformedEnvelope, true, "decode-envelope", "", nil)
		}
		if strings.TrimSpace(req.RunID) == "" || strings.TrimSpace(req.Monitor.TenantID) == "" {
			return summary, checkexecution.NewRuntimeFailure(checkexecution.FailureMalformedEnvelope, true, "validate-envelope", req.RunID, nil)
		}
		work, found, err := h.repo.LoadExecutionWork(ctx, req.Monitor.TenantID, req.RunID)
		if err != nil {
			return summary, checkexecution.Storage("load-work", req.RunID)
		}
		if !found || !sameEnvelopeIdentity(req, work) {
			return summary, checkexecution.Conflict("validate-envelope", req.RunID)
		}
		claimedWork, claimed, err := h.repo.ClaimExecutionWork(ctx, work, h.now())
		if err != nil {
			return summary, checkexecution.Storage("claim-work", work.RunID)
		}
		if !claimed {
			return summary, checkexecution.Duplicate("claim-work", work.RunID)
		}
		work = claimedWork
		monitor, found, err := h.repo.GetMonitor(ctx, work.TenantID, work.ServiceID, work.MonitorID)
		if err != nil {
			return summary, checkexecution.Storage("load-monitor", work.RunID)
		}
		if !found {
			if err := h.repo.MarkExecutionWorkSkipped(ctx, work, h.now(), "monitor not found"); err != nil {
				return summary, err
			}
			summary.Skipped++
			continue
		}
		if skipReason, err := h.currentWorkSkipReason(ctx, monitor, work); err != nil {
			return summary, err
		} else if skipReason != "" {
			if err := h.repo.MarkExecutionWorkSkipped(ctx, work, h.now(), skipReason); err != nil {
				return summary, err
			}
			return summary, nil
		}
		request, skipReason, err := buildExecutionRequest(monitor, work)
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
		result := checkexecution.ExecuteHTTP(ctx, h.executor, request)
		_, _, err = h.resultCommand.execute(ctx, monitor, work, result)
		if err != nil {
			return summary, err
		}
		summary.Processed++
	}
	return summary, nil
}

func resultIdentityMatchesWork(result checkexecution.ExecutionResult, work checkexecution.ExecutionWork) bool {
	if !strings.EqualFold(result.TenantID, work.TenantID) || !strings.EqualFold(result.ServiceID, work.ServiceID) || !strings.EqualFold(result.MonitorID, work.MonitorID) || !strings.EqualFold(result.RunID, work.RunID) || result.Trigger != work.Trigger {
		return false
	}
	if work.ScheduleDefinitionVersion != "" && result.ScheduleDefinitionVersion != work.ScheduleDefinitionVersion {
		return false
	}
	if work.ScheduledFor == nil {
		return result.ScheduledFor == nil
	}
	return result.ScheduledFor != nil && result.ScheduledFor.Equal(*work.ScheduledFor)
}

func (h runtimeHandler) currentWorkSkipReason(ctx context.Context, monitor monitorconfig.Monitor, work checkexecution.ExecutionWork) (string, error) {
	if !monitor.Enabled {
		return "monitor disabled", nil
	}
	status, found, err := h.repo.GetMonitorStatus(ctx, monitor.TenantID, monitor.ServiceID, monitor.MonitorID)
	if err != nil {
		return "", err
	}
	if found && strings.EqualFold(status.CurrentStatus, string(resultstatus.MonitorStateMaintenance)) {
		return "monitor maintenance", nil
	}
	if work.Trigger == checkexecution.TriggerTypeRecurring && work.ScheduleDefinitionVersion != checkexecution.ScheduleDefinitionVersion(monitor) {
		return "schedule definition superseded", nil
	}
	return "", nil
}

func sameEnvelopeIdentity(request checkexecution.ExecutionRequest, work checkexecution.ExecutionWork) bool {
	if !strings.EqualFold(request.Monitor.TenantID, work.TenantID) || !strings.EqualFold(request.Monitor.ServiceID, work.ServiceID) || !strings.EqualFold(request.Monitor.MonitorID, work.MonitorID) || !strings.EqualFold(request.RunID, work.RunID) || request.Trigger != work.Trigger || request.ScheduleDefinitionVersion != work.ScheduleDefinitionVersion {
		return false
	}
	if request.ScheduledFor == nil || work.ScheduledFor == nil {
		return request.ScheduledFor == nil && work.ScheduledFor == nil
	}
	return request.ScheduledFor.Equal(*work.ScheduledFor)
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

const (
	recoveryShards         = 16
	recoveryBucketsPerHour = 4
	recoveryPageLimit      = 25
	dispatchPendingShards  = 4
	dispatchPendingBuckets = 1
)

func (h runtimeHandler) reconcileDispatchPending(ctx context.Context) (int, error) {
	now := h.now().UTC()
	inspected := 0
	for bucketOffset := 0; bucketOffset < dispatchPendingBuckets; bucketOffset++ {
		bucket := now.Add(time.Duration(-bucketOffset) * time.Hour).Format(dynamodbrecord.DispatchPendingBucketFormat)
		for shard := 0; shard < dispatchPendingShards; shard++ {
			shardHex := fmt.Sprintf("%02x", shard)
			bucketShard := bucket + "|" + shardHex
			var cursor map[string]sharedaws.AttributeValue
			for {
				records, nextKey, err := h.repo.ListDispatchPending(ctx, h.tenantID, bucketShard, recoveryPageLimit, cursor)
				if err != nil {
					return inspected, err
				}
				inspected += len(records)
				if nextKey == nil {
					break
				}
				cursor = nextKey
			}
		}
	}
	return inspected, nil
}

func (h runtimeHandler) recoverPublicationMarkers(ctx context.Context) (int, error) {
	if h.queueURL == "" {
		return 0, nil
	}
	now := h.now().UTC()
	recovered := 0
	for bucketOffset := 0; bucketOffset < recoveryBucketsPerHour; bucketOffset++ {
		bucket := now.Add(time.Duration(-bucketOffset) * time.Hour).Format("2006010215")
		for shard := 0; shard < recoveryShards; shard++ {
			shardHex := fmt.Sprintf("%02x", shard)
			var cursor map[string]sharedaws.AttributeValue
			for {
				markers, nextKey, err := h.repo.ListPublicationMarkers(ctx, h.tenantID, bucket+"|"+shardHex, recoveryPageLimit, cursor)
				if err != nil {
					return recovered, err
				}
				for _, marker := range markers {
					work, found, err := h.repo.LoadExecutionWork(ctx, marker.TenantID, marker.RunID)
					if err != nil {
						return recovered, err
					}
					if !found || work.Status != checkexecution.ExecutionWorkPending {
						continue
					}
					payload, err := json.Marshal(checkexecution.ExecutionRequest{
						Monitor:                   buildMonitorFromWork(work),
						RunID:                     work.RunID,
						Trigger:                   work.Trigger,
						AcceptedAt:                work.AcceptedAt,
						ScheduleDefinitionVersion: work.ScheduleDefinitionVersion,
						ScheduledFor:              work.ScheduledFor,
					})
					if err != nil {
						return recovered, err
					}
					if err := h.sqsClient.SendMessage(ctx, h.queueURL, string(payload)); err != nil {
						return recovered, checkexecution.Publication("recover-publication", work.RunID)
					}
					if err := h.repo.AcknowledgeExecutionPublication(ctx, work); err != nil {
						return recovered, err
					}
					recovered++
					runtimeOutcomeLog(outcomeRecovered, "recover-publication", work.TenantID, work.RunID, work.MonitorID, "")
				}
				if nextKey == nil {
					break
				}
				cursor = nextKey
			}
		}
	}
	return recovered, nil
}

func buildMonitorFromWork(work checkexecution.ExecutionWork) monitorconfig.Monitor {
	return monitorconfig.Monitor{TenantID: work.TenantID, ServiceID: work.ServiceID, MonitorID: work.MonitorID, Enabled: true, IntervalSeconds: 60, HTTP: &monitorconfig.HTTPConfiguration{}}
}
