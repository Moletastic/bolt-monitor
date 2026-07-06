package main

import (
	"encoding/json"
	"strings"
	"time"

	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/resultstatus"
)

const defaultTenantID = "DEFAULT"

type serviceResponse struct {
	TenantID           string                          `json:"tenantId"`
	ServiceID          string                          `json:"serviceId"`
	Name               string                          `json:"name"`
	Description        string                          `json:"description,omitempty"`
	LifecycleState     string                          `json:"lifecycleState"`
	TechnologyKey      string                          `json:"technologyKey,omitempty"`
	EscalationPolicyID string                          `json:"escalationPolicyId,omitempty"`
	BusinessHours      *escalation.BusinessHoursConfig `json:"businessHours,omitempty"`
	MonitorCount       int                             `json:"monitorCount"`
	EnabledCount       int                             `json:"enabledMonitorCount"`
	RollupStatus       string                          `json:"rollupStatus"`
	CreatedAt          string                          `json:"createdAt,omitempty"`
	UpdatedAt          string                          `json:"updatedAt,omitempty"`
	Monitors           []monitorResponse               `json:"monitors,omitempty"`
}

type listServicesResponse struct {
	Services []serviceResponse `json:"services"`
}

type updateServiceRequest struct {
	ServiceID          *string                         `json:"serviceId,omitempty"`
	Name               *string                         `json:"name,omitempty"`
	Description        *string                         `json:"description,omitempty"`
	TechnologyKey      *string                         `json:"technologyKey,omitempty"`
	EscalationPolicyID *string                         `json:"escalationPolicyId,omitempty"`
	BusinessHours      *escalation.BusinessHoursConfig `json:"businessHours,omitempty"`
}

type escalationPolicyResponse struct {
	TenantID          string                    `json:"tenantId"`
	PolicyID          string                    `json:"policyId"`
	Name              string                    `json:"name"`
	Description       string                    `json:"description,omitempty"`
	BusinessHoursPath escalation.EscalationPath `json:"businessHoursPath"`
	OffHoursPath      escalation.EscalationPath `json:"offHoursPath"`
	CreatedAt         string                    `json:"createdAt,omitempty"`
	UpdatedAt         string                    `json:"updatedAt,omitempty"`
}

type listEscalationPoliciesResponse struct {
	Policies []escalationPolicyResponse `json:"policies"`
}

type notificationChannelResponse struct {
	ChannelID string                 `json:"channelId"`
	TenantID  string                 `json:"tenantId"`
	Name      string                 `json:"name"`
	Type      escalation.ChannelType `json:"type"`
	Target    string                 `json:"target"`
	Config    json.RawMessage        `json:"config,omitempty"`
	CreatedAt string                 `json:"createdAt,omitempty"`
	UpdatedAt string                 `json:"updatedAt,omitempty"`
}

type listNotificationChannelsResponse struct {
	Channels []notificationChannelResponse `json:"channels"`
}

type notificationChannelRequest struct {
	Name   *string         `json:"name,omitempty" validate:"omitempty,notblank,max=80"`
	Type   *string         `json:"type,omitempty" validate:"omitempty,notblank"`
	Target *string         `json:"target,omitempty" validate:"omitempty,notblank"`
	Config json.RawMessage `json:"config,omitempty"`
}

type notificationChannelInput struct {
	Name   string `json:"name" validate:"notblank,max=80"`
	Type   string `json:"type" validate:"notblank"`
	Target string `json:"target" validate:"notblank"`
}

type routeReference struct {
	PolicyID string `json:"policyId"`
	Name     string `json:"name"`
}

type channelInUseResponse struct {
	Error             string           `json:"error"`
	ReferencingRoutes []routeReference `json:"referencingRoutes"`
}

type notificationChannelTestResponse struct {
	ChannelID string `json:"channelId"`
	SentAt    string `json:"sentAt"`
}

type escalationPolicyRequest struct {
	Name              string                `json:"name" validate:"notblank"`
	Description       string                `json:"description,omitempty"`
	BusinessHoursPath escalationPathRequest `json:"businessHoursPath" validate:"required"`
	OffHoursPath      escalationPathRequest `json:"offHoursPath" validate:"required"`
}

type escalationPathRequest struct {
	Steps []escalationStepRequest `json:"steps" validate:"min=1,dive"`
}

type escalationStepRequest struct {
	ChannelID    string `json:"channelId" validate:"notblank"`
	DelayMinutes int    `json:"delayMinutes"`
}

type escalationStateResponse struct {
	Exists       bool                        `json:"exists"`
	TenantID     string                      `json:"tenantId,omitempty"`
	IncidentID   string                      `json:"incidentId,omitempty"`
	PolicyID     string                      `json:"policyId,omitempty"`
	ServiceID    string                      `json:"serviceId,omitempty"`
	MonitorID    string                      `json:"monitorId,omitempty"`
	CurrentStep  int                         `json:"currentStep,omitempty"`
	StepsFired   []int                       `json:"stepsFired,omitempty"`
	SelectedPath string                      `json:"selectedPath,omitempty"`
	ScheduledFor string                      `json:"scheduledFor,omitempty"`
	Status       escalation.EscalationStatus `json:"status,omitempty"`
	CreatedAt    string                      `json:"createdAt,omitempty"`
	UpdatedAt    string                      `json:"updatedAt,omitempty"`
}

type monitorResponse struct {
	TenantID          string                           `json:"tenantId"`
	ServiceID         string                           `json:"serviceId"`
	MonitorID         string                           `json:"monitorId"`
	Name              string                           `json:"name"`
	Type              monitorconfig.MonitorType        `json:"type"`
	IntervalSeconds   int                              `json:"intervalSeconds"`
	ProbeLocations    []string                         `json:"probeLocations"`
	Enabled           bool                             `json:"enabled"`
	FailureThreshold  int                              `json:"failureThreshold"`
	RecoveryThreshold int                              `json:"recoveryThreshold"`
	HTTP              *monitorconfig.HTTPConfiguration `json:"http,omitempty"`
	Status            *monitorStatusResponse           `json:"status,omitempty"`
}

type listMonitorsResponse struct {
	Monitors []monitorResponse `json:"monitors"`
}

type probeLocationResponse struct {
	LocationID  string `json:"locationId"`
	DisplayName string `json:"displayName"`
	Enabled     bool   `json:"enabled"`
}

type listProbeLocationsResponse struct {
	ProbeLocations []probeLocationResponse `json:"probeLocations"`
}

type updateMonitorRequest struct {
	Name              *string                          `json:"name,omitempty"`
	IntervalSeconds   *int                             `json:"intervalSeconds,omitempty"`
	ProbeLocations    []string                         `json:"probeLocations,omitempty"`
	FailureThreshold  *int                             `json:"failureThreshold,omitempty"`
	RecoveryThreshold *int                             `json:"recoveryThreshold,omitempty"`
	HTTP              *monitorconfig.HTTPConfiguration `json:"http,omitempty"`
}

// Note: services/monitor-api retired the flat `errorResponse` shape.
// All handlers now return `response.Envelope[...]` from shared/api/response.
// See AGENTS.md → "Response envelope" for the wire format.

type monitorStatusResponse struct {
	CurrentStatus       string `json:"currentStatus"`
	LastCheckedAt       string `json:"lastCheckedAt"`
	LastDurationMs      int64  `json:"lastDurationMs"`
	LastProbeLocationID string `json:"lastProbeLocationId"`
	LastError           string `json:"lastError,omitempty"`
	LastOutcome         string `json:"lastOutcome"`
}

type checkRunResponse struct {
	RunID           string `json:"runId"`
	ServiceID       string `json:"serviceId"`
	MonitorID       string `json:"monitorId"`
	Type            string `json:"type"`
	ProbeLocationID string `json:"probeLocationId"`
	Trigger         string `json:"trigger"`
	StartedAt       string `json:"startedAt"`
	FinishedAt      string `json:"finishedAt"`
	DurationMs      int64  `json:"durationMs"`
	Outcome         string `json:"outcome"`
	StatusCode      *int   `json:"statusCode,omitempty"`
	Error           string `json:"error,omitempty"`
}

type monitorRunsResponse struct {
	Runs []checkRunResponse `json:"runs"`
}

type manualRunResponse struct {
	RunID           string `json:"runId"`
	ServiceID       string `json:"serviceId"`
	MonitorID       string `json:"monitorId"`
	Trigger         string `json:"trigger"`
	Outcome         string `json:"outcome,omitempty"`
	DurationMs      int64  `json:"durationMs,omitempty"`
	StatusCode      *int   `json:"statusCode,omitempty"`
	Error           string `json:"error,omitempty"`
	ProbeLocationID string `json:"probeLocationId,omitempty"`
	StartedAt       string `json:"startedAt,omitempty"`
	FinishedAt      string `json:"finishedAt,omitempty"`
}

type incidentResponse struct {
	IncidentID         string `json:"incidentId"`
	ServiceID          string `json:"serviceId"`
	MonitorID          string `json:"monitorId"`
	Type               string `json:"type,omitempty"`
	Summary            string `json:"summary"`
	Status             string `json:"status"`
	OpenedAt           string `json:"openedAt"`
	AcknowledgedAt     string `json:"acknowledgedAt,omitempty"`
	ResolvedAt         string `json:"resolvedAt,omitempty"`
	UpdatedAt          string `json:"updatedAt"`
	Origin             string `json:"origin,omitempty"`
	OriginalIncidentID string `json:"originalIncidentId,omitempty"`
}

type incidentActivityResponse struct {
	ActivityID string `json:"activityId"`
	IncidentID string `json:"incidentId"`
	Action     string `json:"action"`
	Timestamp  string `json:"timestamp"`
}

type listIncidentsResponse struct {
	Incidents []incidentResponse `json:"incidents"`
}

type incidentActivitiesResponse struct {
	Activities []incidentActivityResponse `json:"activities"`
}

type schedulerConfigResponse struct {
	RecurringEnabled bool   `json:"recurringEnabled"`
	StopControlMode  string `json:"stopControlMode,omitempty"`
	UpdatedAt        string `json:"updatedAt"`
}

type updateSchedulerConfigRequest struct {
	RecurringEnabled bool   `json:"recurringEnabled"`
	StopControlMode  string `json:"stopControlMode,omitempty"`
}

type auditEventResponse struct {
	AuditID    string `json:"auditId"`
	ServiceID  string `json:"serviceId,omitempty"`
	MonitorID  string `json:"monitorId,omitempty"`
	EventType  string `json:"eventType"`
	OccurredAt string `json:"occurredAt"`
	Actor      string `json:"actor,omitempty"`
	Origin     string `json:"origin,omitempty"`
}

type monitorAuditEventsResponse struct {
	Events []auditEventResponse `json:"events"`
}

type manualRunRequestRecord struct {
	RunID      string
	ServiceID  string
	MonitorID  string
	TenantID   string
	Trigger    checkexecution.TriggerType
	AcceptedAt string
}

type auditEventView struct {
	AuditID    string
	ServiceID  string
	MonitorID  string
	EventType  string
	OccurredAt string
	Actor      string
	Origin     string
}

func toServiceResponse(service monitorconfig.Service) serviceResponse {
	return serviceResponse{
		TenantID:           service.TenantID,
		ServiceID:          service.ServiceID,
		Name:               service.Name,
		Description:        service.Description,
		LifecycleState:     string(service.LifecycleState),
		TechnologyKey:      service.TechnologyKey,
		EscalationPolicyID: service.EscalationPolicyID,
		BusinessHours:      dynamodbrecord.CloneBusinessHoursConfig(service.BusinessHours),
		MonitorCount:       service.MonitorCount,
		EnabledCount:       service.EnabledCount,
		RollupStatus:       service.RollupStatus,
		CreatedAt:          service.CreatedAt,
		UpdatedAt:          service.UpdatedAt,
	}
}

func toEscalationPolicyResponse(policy escalation.EscalationPolicy) escalationPolicyResponse {
	return escalationPolicyResponse{
		TenantID:          policy.TenantID,
		PolicyID:          policy.PolicyID,
		Name:              policy.Name,
		Description:       policy.Description,
		BusinessHoursPath: cloneEscalationPath(policy.BusinessHoursPath),
		OffHoursPath:      cloneEscalationPath(policy.OffHoursPath),
		CreatedAt:         policy.CreatedAt,
		UpdatedAt:         policy.UpdatedAt,
	}
}

func toNotificationChannelResponse(channel escalation.NotificationChannel) notificationChannelResponse {
	redacted := channel
	redacted.Config = redactChannelConfig(redacted.Config)
	return notificationChannelResponse{
		ChannelID: redacted.ChannelID,
		TenantID:  redacted.TenantID,
		Name:      redacted.Name,
		Type:      redacted.Type,
		Target:    redacted.Target,
		Config:    redacted.Config,
		CreatedAt: redacted.CreatedAt,
		UpdatedAt: redacted.UpdatedAt,
	}
}

func toEscalationStateResponse(state escalation.EscalationState) escalationStateResponse {
	steps := append([]int(nil), state.StepsFired...)
	return escalationStateResponse{
		Exists:       true,
		TenantID:     state.TenantID,
		IncidentID:   state.IncidentID,
		PolicyID:     state.PolicyID,
		ServiceID:    state.ServiceID,
		MonitorID:    state.MonitorID,
		CurrentStep:  state.CurrentStep,
		StepsFired:   steps,
		SelectedPath: state.SelectedPath,
		ScheduledFor: state.ScheduledFor,
		Status:       state.Status,
		CreatedAt:    state.CreatedAt,
		UpdatedAt:    state.UpdatedAt,
	}
}

func toMonitorResponse(monitor monitorconfig.Monitor, status *monitorStatusResponse) monitorResponse {
	return monitorResponse{
		TenantID:          monitor.TenantID,
		ServiceID:         monitor.ServiceID,
		MonitorID:         monitor.MonitorID,
		Name:              monitor.Name,
		Type:              monitor.Type,
		IntervalSeconds:   monitor.IntervalSeconds,
		ProbeLocations:    append([]string(nil), monitor.ProbeLocations...),
		Enabled:           monitor.Enabled,
		FailureThreshold:  monitor.FailureThreshold,
		RecoveryThreshold: monitor.RecoveryThreshold,
		HTTP:              dynamodbrecord.CloneHTTPConfiguration(monitor.HTTP),
		Status:            status,
	}
}

func toStatusResponse(status resultstatus.MonitorStatus) *monitorStatusResponse {
	return &monitorStatusResponse{
		CurrentStatus:       strings.ToLower(status.CurrentStatus),
		LastCheckedAt:       status.LastCheckedAt.UTC().Format(time.RFC3339),
		LastDurationMs:      status.LastDurationMs,
		LastProbeLocationID: status.LastProbeLocationID,
		LastError:           status.LastError,
		LastOutcome:         string(status.LastOutcome),
	}
}

func toUnknownStatusResponse() *monitorStatusResponse {
	return &monitorStatusResponse{CurrentStatus: rollupUnknown, LastCheckedAt: time.Time{}.Format(time.RFC3339), LastOutcome: rollupUnknown}
}

func toRunResponse(run resultstatus.CheckRun) checkRunResponse {
	return checkRunResponse{RunID: run.RunID, ServiceID: run.ServiceID, MonitorID: run.MonitorID, Type: run.Type, ProbeLocationID: run.ProbeLocationID, Trigger: string(run.Trigger), StartedAt: run.StartedAt.UTC().Format(time.RFC3339), FinishedAt: run.FinishedAt.UTC().Format(time.RFC3339), DurationMs: run.DurationMs, Outcome: string(run.Outcome), StatusCode: run.StatusCode, Error: run.Error}
}

func toManualRunResponseWithResult(run manualRunRequestRecord, result checkexecution.ExecutionResult) manualRunResponse {
	return manualRunResponse{
		RunID:           run.RunID,
		ServiceID:       run.ServiceID,
		MonitorID:       run.MonitorID,
		Trigger:         string(run.Trigger),
		Outcome:         string(result.Outcome),
		DurationMs:      result.DurationMs,
		StatusCode:      result.StatusCode,
		Error:           result.Error,
		ProbeLocationID: result.ProbeLocationID,
		StartedAt:       result.StartedAt.UTC().Format(time.RFC3339),
		FinishedAt:      result.FinishedAt.UTC().Format(time.RFC3339),
	}
}

func toIncidentResponse(incident dynamodbrecord.IncidentRecord) incidentResponse {
	return incidentResponse{IncidentID: incident.IncidentID, ServiceID: incident.ServiceID, MonitorID: incident.MonitorID, Type: incident.Type, Summary: incident.Summary, Status: incident.Status, OpenedAt: incident.OpenedAt, AcknowledgedAt: incident.AcknowledgedAt, ResolvedAt: incident.ResolvedAt, UpdatedAt: incident.UpdatedAt, Origin: incident.Origin, OriginalIncidentID: incident.OriginalIncidentID}
}

func toIncidentActivityResponse(activity dynamodbrecord.IncidentActivityRecord) incidentActivityResponse {
	return incidentActivityResponse{ActivityID: activity.ActivityID, IncidentID: activity.IncidentID, Action: activity.Action, Timestamp: activity.Timestamp}
}

func toSchedulerConfigResponse(config dynamodbrecord.SchedulerConfigRecord) schedulerConfigResponse {
	return schedulerConfigResponse{RecurringEnabled: config.Config.RecurringEnabled, StopControlMode: string(config.Config.StopControlMode), UpdatedAt: config.UpdatedAt}
}

func toAuditEventResponse(event auditEventView) auditEventResponse {
	return auditEventResponse(event)
}

func cloneEscalationPath(input escalation.EscalationPath) escalation.EscalationPath {
	steps := make([]escalation.EscalationStep, 0, len(input.Steps))
	for _, step := range input.Steps {
		channels := make([]escalation.ChannelConfig, 0, len(step.Channels))
		for _, channel := range step.Channels {
			var cfg json.RawMessage
			if channel.Config != nil {
				cfg = append(json.RawMessage(nil), channel.Config...)
			}
			channels = append(channels, escalation.ChannelConfig{Type: channel.Type, Target: strings.TrimSpace(channel.Target), Config: cfg})
		}
		steps = append(steps, escalation.EscalationStep{ChannelID: strings.TrimSpace(step.ChannelID), DelayMinutes: step.DelayMinutes, Channels: channels})
	}
	return escalation.EscalationPath{Steps: steps}
}

func redactChannelConfig(input json.RawMessage) json.RawMessage {
	if len(input) == 0 {
		return nil
	}
	var cfg map[string]any
	if err := json.Unmarshal(input, &cfg); err != nil {
		return append(json.RawMessage(nil), input...)
	}
	for _, key := range []string{"botToken", "apiKey", "authToken", "accountSid"} {
		if _, ok := cfg[key]; ok {
			cfg[key] = "***REDACTED***"
		}
	}
	encoded, err := json.Marshal(cfg)
	if err != nil {
		return append(json.RawMessage(nil), input...)
	}
	return encoded
}
