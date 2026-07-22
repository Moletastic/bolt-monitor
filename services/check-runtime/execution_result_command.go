package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/domainvalues"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/resultstatus"
)

// executionResultState is the persisted input to the pure transition decision.
type executionResultState struct {
	status        resultstatus.MonitorStatus
	statusFound   bool
	openIncident  dynamodbrecord.IncidentRecord
	incidentFound bool
}

// executionResultPublication is durable outbox intent. Its persistence remains
// part of the result transaction so retries and fencing retain their behavior.
type executionResultPublication struct {
	transition string
	incidentID string
}

type executionResultPersistence interface {
	LoadExecutionResultState(context.Context, checkexecution.ExecutionResult) (executionResultState, error)
	CommitExecutionResult(context.Context, monitorconfig.Monitor, checkexecution.ExecutionWork, checkexecution.ExecutionResult, []any, resultstatus.MonitorStatus, bool, executionResultPublication) error
}

type executionResultClock interface {
	Now() time.Time
}

type executionResultIDs interface {
	NewIncidentID(time.Time) string
	NewAuditID(time.Time) string
}

type systemExecutionResultClock struct{}

func (systemExecutionResultClock) Now() time.Time { return time.Now() }

type generatedExecutionResultIDs struct{}

func (generatedExecutionResultIDs) NewIncidentID(at time.Time) string { return newIncidentID(at) }
func (generatedExecutionResultIDs) NewAuditID(at time.Time) string    { return newAuditID(at) }

// executionResultCommand owns identity validation and transition intent;
// persistence owns the DynamoDB transaction that commits that intent.
type executionResultCommand struct {
	persistence executionResultPersistence
	clock       executionResultClock
	ids         executionResultIDs
}

func newExecutionResultCommand(persistence executionResultPersistence, clock executionResultClock, ids executionResultIDs) executionResultCommand {
	return executionResultCommand{persistence: persistence, clock: clock, ids: ids}
}

func (c executionResultCommand) execute(ctx context.Context, monitor monitorconfig.Monitor, work checkexecution.ExecutionWork, result checkexecution.ExecutionResult) (string, string, error) {
	if !resultIdentityMatchesWork(result, work) {
		return "", "", checkexecution.Conflict("commit-result", work.RunID)
	}
	if result.FinishedAt.IsZero() {
		result.FinishedAt = c.clock.Now().UTC()
	}
	state, err := c.persistence.LoadExecutionResultState(ctx, result)
	if err != nil {
		return "", "", err
	}
	thresholds := resultstatus.ThresholdConfig{FailureThreshold: monitor.FailureThreshold, RecoveryThreshold: monitor.RecoveryThreshold}
	if !state.statusFound {
		state.status = resultstatus.NewMonitorStatus(result)
	}
	records, transition, incidentID, status, err := decideExecutionResult(monitor, result, state.status, thresholds, state.openIncident, state.incidentFound, c.ids.NewIncidentID, c.ids.NewAuditID)
	if err != nil {
		return "", "", err
	}
	applyProjection := result.Trigger == checkexecution.TriggerTypeRecurring && result.ScheduledFor != nil && (!state.statusFound || resultstatus.IsNewerRecurringObservation(state.status, *result.ScheduledFor, result.RunID))
	publication := executionResultPublication{transition: transition, incidentID: incidentID}
	if err := c.persistence.CommitExecutionResult(ctx, monitor, work, result, records, status, applyProjection, publication); err != nil {
		return "", "", err
	}
	return publication.transition, publication.incidentID, nil
}

const (
	incidentStatusOpen         = "open"
	incidentStatusAcknowledged = "acknowledged"
	incidentStatusResolved     = "resolved"
)

// decideExecutionResult contains monitor-status and incident transition policy
// only. Its ID generators are inputs, so it has no AWS or process-global state.
func decideExecutionResult(monitor monitorconfig.Monitor, result checkexecution.ExecutionResult, currentStatus resultstatus.MonitorStatus, thresholdConfig resultstatus.ThresholdConfig, current dynamodbrecord.IncidentRecord, found bool, newIncidentID func(time.Time) string, newAuditID func(time.Time) string) ([]any, string, string, resultstatus.MonitorStatus, error) {
	isManual := result.Trigger == checkexecution.TriggerTypeManual
	currentState := domainvalues.MonitorStateFromStored(currentStatus.CurrentStatus)
	if currentState == "" {
		currentState = domainvalues.MonitorStateUp
	}
	failureThreshold := thresholdConfig.FailureThreshold
	if failureThreshold < 1 {
		failureThreshold = 1
	}
	recoveryThreshold := thresholdConfig.RecoveryThreshold
	if recoveryThreshold < 1 {
		recoveryThreshold = 1
	}
	newStatus := currentStatus
	newStatus.LastCheckedAt = result.FinishedAt.UTC()
	newStatus.LastDurationMs = result.DurationMs
	newStatus.LastError = result.Error
	newStatus.LastFailureCode = result.FailureCode
	newStatus.LastOutcome = result.Outcome
	var incidentRecords []any
	var transition string
	var incidentID string
	if isManual {
		newStatus.CurrentStatus = currentState.Stored()
		return incidentRecords, "", "", newStatus, nil
	}
	if result.Outcome == checkexecution.OutcomeSuccess {
		newStatus.ConsecutiveFailures = 0
		newStatus.ConsecutiveSuccesses++
		switch currentState {
		case domainvalues.MonitorStateUp:
			newStatus.CurrentStatus = domainvalues.MonitorStateUp.Stored()
		case domainvalues.MonitorStateDegraded:
			newStatus.ConsecutiveFailures = 0
			newStatus.CurrentStatus = domainvalues.MonitorStateUp.Stored()
		case domainvalues.MonitorStateDown:
			newStatus.CurrentStatus = domainvalues.MonitorStateRecovering.Stored()
		case domainvalues.MonitorStateRecovering:
			if newStatus.ConsecutiveSuccesses >= recoveryThreshold {
				newStatus.CurrentStatus = domainvalues.MonitorStateUp.Stored()
				newStatus.ConsecutiveSuccesses = 0
				if found {
					current.Status = incidentStatusResolved
					current.ResolvedAt = result.FinishedAt.UTC().Format(time.RFC3339)
					current.UpdatedAt = current.ResolvedAt
					incidentRecords = buildIncidentRecords(current, "INCIDENT_RESOLVED", current.ResolvedAt, result.RunID, result.FinishedAt, newAuditID)
					transition = "incident.up"
					incidentID = current.IncidentID
				}
			} else {
				newStatus.CurrentStatus = domainvalues.MonitorStateRecovering.Stored()
			}
		case domainvalues.MonitorStateMaintenance:
			newStatus.CurrentStatus = domainvalues.MonitorStateMaintenance.Stored()
		default:
			newStatus.CurrentStatus = domainvalues.MonitorStateUp.Stored()
		}
	} else {
		newStatus.ConsecutiveSuccesses = 0
		newStatus.ConsecutiveFailures++
		switch currentState {
		case domainvalues.MonitorStateUp:
			if failureThreshold == 1 {
				newStatus.CurrentStatus = domainvalues.MonitorStateDown.Stored()
			} else {
				newStatus.CurrentStatus = domainvalues.MonitorStateDegraded.Stored()
			}
		case domainvalues.MonitorStateDegraded:
			if newStatus.ConsecutiveFailures >= failureThreshold {
				newStatus.CurrentStatus = domainvalues.MonitorStateDown.Stored()
			} else {
				newStatus.CurrentStatus = domainvalues.MonitorStateDegraded.Stored()
			}
		case domainvalues.MonitorStateDown:
			newStatus.CurrentStatus = domainvalues.MonitorStateDown.Stored()
			if found {
				current.Summary = incidentSummary(monitor, result)
				current.UpdatedAt = result.FinishedAt.UTC().Format(time.RFC3339)
				incidentRecords = buildIncidentRecords(current, "INCIDENT_UPDATED", current.UpdatedAt, result.RunID, result.FinishedAt, newAuditID)
			}
		case domainvalues.MonitorStateRecovering:
			newStatus.CurrentStatus = domainvalues.MonitorStateDown.Stored()
			newStatus.ConsecutiveSuccesses = 0
		case domainvalues.MonitorStateMaintenance:
			newStatus.CurrentStatus = domainvalues.MonitorStateMaintenance.Stored()
		default:
			newStatus.CurrentStatus = domainvalues.MonitorStateDegraded.Stored()
		}
	}
	if newStatus.CurrentStatus == domainvalues.MonitorStateDown.Stored() && currentState != domainvalues.MonitorStateDown {
		if !found {
			incident := dynamodbrecord.IncidentRecord{IncidentID: newIncidentID(result.FinishedAt), ServiceID: strings.ToLower(result.ServiceID), MonitorID: strings.ToLower(result.MonitorID), TenantID: strings.ToUpper(result.TenantID), Type: "monitoring", Summary: incidentSummary(monitor, result), Status: incidentStatusOpen, OpenedAt: result.FinishedAt.UTC().Format(time.RFC3339), UpdatedAt: result.FinishedAt.UTC().Format(time.RFC3339), Origin: "system"}
			incidentRecords = buildIncidentRecords(incident, "INCIDENT_OPENED", incident.IncidentID, result.RunID, result.FinishedAt, newAuditID)
			transition = "incident.down"
			incidentID = incident.IncidentID
		}
	}
	return incidentRecords, transition, incidentID, newStatus, nil
}

func buildIncidentRecords(incident dynamodbrecord.IncidentRecord, action, changeValue string, runID string, now time.Time, newAuditID func(time.Time) string) []any {
	auditID := newAuditID(now)
	activityID := checkexecution.TransitionID(runID)
	auditEvent := dynamodbrecord.NewAuditEventRecord(now, auditID, incident.TenantID, action, incident.ServiceID, incident.MonitorID)
	change := dynamodbrecord.NewAuditChangeRecord(auditEvent.AuditID, "incident", "", changeValue)
	activity := dynamodbrecord.NewIncidentActivityRecord(incident.TenantID, incident.IncidentID, activityID, action, now)
	return []any{dynamodbrecord.NewIncidentMonitorItemRecord(incident), dynamodbrecord.NewIncidentRefItemRecord(incident), dynamodbrecord.NewIncidentMetaItemRecord(incident), activity, auditEvent, change}
}

func incidentSummary(monitor monitorconfig.Monitor, result checkexecution.ExecutionResult) string {
	summary := fmt.Sprintf("%s failed", monitor.Name)
	if result.Outcome == checkexecution.OutcomeSuccess {
		return summary
	}
	if result.Error != "" {
		return fmt.Sprintf("%s: %s", summary, result.Error)
	}
	if result.StatusCode != nil {
		return fmt.Sprintf("%s: status %d", summary, *result.StatusCode)
	}
	return summary
}
