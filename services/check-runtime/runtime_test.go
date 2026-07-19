package main

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/outboundhttp"
	"bolt-monitor/shared/resultstatus"
	"github.com/aws/aws-lambda-go/events"
)

type fakeSQSClient struct {
	sentMessages []string
}

type fakeExecutor struct{ response outboundhttp.Response }

func (e fakeExecutor) Execute(_ context.Context, _ outboundhttp.Request) (outboundhttp.Response, error) {
	return e.response, nil
}

type countingDialer struct{ calls int }

func (d *countingDialer) DialContext(context.Context, string, string) (net.Conn, error) {
	d.calls++
	return nil, errors.New("unexpected dial")
}

func (f *fakeSQSClient) SendMessage(_ context.Context, _ string, body string) error {
	f.sentMessages = append(f.sentMessages, body)
	return nil
}

type fakeRuntimeRepository struct {
	config        checkexecution.SchedulerConfig
	monitors      map[string]monitorconfig.Monitor
	services      map[string]monitorconfig.Service
	works         []checkexecution.ExecutionWork
	claims        map[string]bool
	skipped       []checkexecution.ExecutionWork
	results       []checkexecution.ExecutionResult
	recordedWorks []checkexecution.ExecutionWork
	lastExec      map[string]time.Time
	statuses      map[string]resultstatus.MonitorStatus
}

func newFakeRuntimeRepository() *fakeRuntimeRepository {
	return &fakeRuntimeRepository{monitors: map[string]monitorconfig.Monitor{}, claims: map[string]bool{}, lastExec: map[string]time.Time{}, statuses: map[string]resultstatus.MonitorStatus{}}
}

func (r *fakeRuntimeRepository) GetSchedulerConfig(context.Context, string) (checkexecution.SchedulerConfig, error) {
	return r.config, nil
}

func (r *fakeRuntimeRepository) ListMonitors(context.Context, string) ([]monitorconfig.Monitor, error) {
	out := make([]monitorconfig.Monitor, 0, len(r.monitors))
	for _, monitor := range r.monitors {
		out = append(out, monitor)
	}
	return out, nil
}

func (r *fakeRuntimeRepository) GetLastExecution(_ context.Context, _ string, serviceID, monitorID string) (*time.Time, error) {
	lastExec, ok := r.lastExec[serviceID+"/"+monitorID]
	if !ok {
		return nil, nil
	}
	return &lastExec, nil
}

func (r *fakeRuntimeRepository) RecordLastExecution(_ context.Context, _ string, serviceID, monitorID string, lastExec time.Time) error {
	r.lastExec[serviceID+"/"+monitorID] = lastExec
	return nil
}

func (r *fakeRuntimeRepository) EnqueueExecutionRequests(_ context.Context, requests []checkexecution.ExecutionRequest, now time.Time) error {
	for _, request := range requests {
		runID := request.RunID
		if runID == "" {
			runID = newRunID(now)
		}
		r.works = append(r.works, checkexecution.ExecutionWork{TenantID: request.Monitor.TenantID, ServiceID: request.Monitor.ServiceID, MonitorID: request.Monitor.MonitorID, RunID: runID, Trigger: request.Trigger, RequestedAt: now.UTC(), Status: checkexecution.ExecutionWorkPending})
	}
	return nil
}

func (r *fakeRuntimeRepository) ListPendingExecutionWork(context.Context, string, int32) ([]checkexecution.ExecutionWork, error) {
	return append([]checkexecution.ExecutionWork(nil), r.works...), nil
}

func (r *fakeRuntimeRepository) ClaimExecutionWork(_ context.Context, work checkexecution.ExecutionWork, _ time.Time) (bool, error) {
	if r.claims[work.RunID] {
		return false, nil
	}
	r.claims[work.RunID] = true
	return true, nil
}

func (r *fakeRuntimeRepository) GetMonitor(_ context.Context, _ string, serviceID, monitorID string) (monitorconfig.Monitor, bool, error) {
	monitor, ok := r.monitors[serviceID+"/"+monitorID]
	return monitor, ok, nil
}

func (r *fakeRuntimeRepository) GetService(_ context.Context, _ string, serviceID string) (monitorconfig.Service, bool, error) {
	svc, ok := r.services[serviceID]
	return svc, ok, nil
}

func (r *fakeRuntimeRepository) MarkExecutionWorkSkipped(_ context.Context, work checkexecution.ExecutionWork, _ time.Time, reason string) error {
	work.Status = checkexecution.ExecutionWorkSkipped
	work.LastError = reason
	r.skipped = append(r.skipped, work)
	return nil
}

func (r *fakeRuntimeRepository) RecordExecutionResult(_ context.Context, _ monitorconfig.Monitor, work checkexecution.ExecutionWork, result checkexecution.ExecutionResult) (string, string, error) {
	r.recordedWorks = append(r.recordedWorks, work)
	r.results = append(r.results, result)
	return "", "", nil
}

func (r *fakeRuntimeRepository) GetMonitorStatus(_ context.Context, _, serviceID, monitorID string) (resultstatus.MonitorStatus, bool, error) {
	status, ok := r.statuses[serviceID+"/"+monitorID]
	return status, ok, nil
}

func testMonitor(target string, enabled bool) monitorconfig.Monitor {
	return monitorconfig.Monitor{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Name: "Homepage", Type: monitorconfig.MonitorTypeHTTP, IntervalSeconds: 60, Enabled: enabled, FailureThreshold: 1, RecoveryThreshold: 1, HTTP: &monitorconfig.HTTPConfiguration{Target: target, Method: "GET", TimeoutMs: 5000, ExpectedStatusCodes: []int{200}}}
}

func TestRunSchedulerRespectsRecurringGate(t *testing.T) {
	repo := newFakeRuntimeRepository()
	repo.config = checkexecution.SchedulerConfig{RecurringEnabled: false}
	repo.monitors["auth/public-http"] = testMonitor("https://example.com", true)
	sqs := &fakeSQSClient{}
	handler := newRuntimeHandler(repo, sqs, "", "", defaultTenantID, modeScheduler)
	handler.now = func() time.Time { return time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC) }

	summary, err := handler.runScheduler(context.Background())
	if err != nil {
		t.Fatalf("runScheduler returned error: %v", err)
	}
	if summary.Enqueued != 0 || len(repo.works) != 0 {
		t.Fatalf("summary = %+v, works = %d, want no enqueued work", summary, len(repo.works))
	}
}

func TestRunSchedulerEnqueuesStableRecurringIdentity(t *testing.T) {
	repo := newFakeRuntimeRepository()
	repo.config = checkexecution.SchedulerConfig{RecurringEnabled: true}
	repo.monitors["auth/public-http"] = testMonitor("https://example.com", true)
	sqs := &fakeSQSClient{}
	handler := newRuntimeHandler(repo, sqs, "queue-url", "", defaultTenantID, modeScheduler)
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	handler.now = func() time.Time { return now }

	summary, err := handler.runScheduler(context.Background())
	if err != nil {
		t.Fatalf("runScheduler returned error: %v", err)
	}
	if summary.Enqueued != 1 || len(sqs.sentMessages) != 1 || len(repo.works) != 1 {
		t.Fatalf("summary = %+v, messages = %d, works = %d, want one enqueued", summary, len(sqs.sentMessages), len(repo.works))
	}
	var request checkexecution.ExecutionRequest
	if err := json.Unmarshal([]byte(sqs.sentMessages[0]), &request); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if request.RunID == "" || request.AcceptedAt != now || request.ScheduledFor == nil || !request.ScheduledFor.Equal(now) {
		t.Fatalf("request = %#v", request)
	}
}

func TestRunSchedulerEnqueuesEnabledMonitorUnderDraftService(t *testing.T) {
	repo := newFakeRuntimeRepository()
	repo.config = checkexecution.SchedulerConfig{RecurringEnabled: true}
	repo.monitors["auth/public-http"] = testMonitor("https://example.com", true)
	repo.services = map[string]monitorconfig.Service{
		"auth": {TenantID: defaultTenantID, ServiceID: "auth", LifecycleState: monitorconfig.ServiceLifecycleDraft},
	}
	sqs := &fakeSQSClient{}
	handler := newRuntimeHandler(repo, sqs, "queue-url", "", defaultTenantID, modeScheduler)
	handler.now = func() time.Time { return time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC) }

	summary, err := handler.runScheduler(context.Background())
	if err != nil {
		t.Fatalf("runScheduler returned error: %v", err)
	}
	if summary.Enqueued != 1 || len(sqs.sentMessages) != 1 || len(repo.works) != 1 {
		t.Fatalf("summary = %+v, messages = %d, works = %d, want one enqueued", summary, len(sqs.sentMessages), len(repo.works))
	}
}

func TestRunSchedulerDoesNotUseLastExecutionForIdentity(t *testing.T) {
	repo := newFakeRuntimeRepository()
	repo.config = checkexecution.SchedulerConfig{RecurringEnabled: true}
	repo.monitors["auth/public-http"] = testMonitor("https://example.com", true)
	repo.lastExec["auth/public-http"] = time.Date(2026, 5, 22, 11, 59, 30, 0, time.UTC)
	sqs := &fakeSQSClient{}
	handler := newRuntimeHandler(repo, sqs, "queue-url", "", defaultTenantID, modeScheduler)
	handler.now = func() time.Time { return time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC) }

	summary, err := handler.runScheduler(context.Background())
	if err != nil {
		t.Fatalf("runScheduler returned error: %v", err)
	}
	if summary.Enqueued != 1 || len(sqs.sentMessages) != 1 || len(repo.works) != 1 {
		t.Fatalf("summary = %+v, messages = %d, works = %d, want one", summary, len(sqs.sentMessages), len(repo.works))
	}
}

func TestRunSchedulerEnqueuesMonitorAfterIntervalElapsed(t *testing.T) {
	repo := newFakeRuntimeRepository()
	repo.config = checkexecution.SchedulerConfig{RecurringEnabled: true}
	repo.monitors["auth/public-http"] = testMonitor("https://example.com", true)
	repo.lastExec["auth/public-http"] = time.Date(2026, 5, 22, 11, 59, 0, 0, time.UTC)
	sqs := &fakeSQSClient{}
	handler := newRuntimeHandler(repo, sqs, "queue-url", "", defaultTenantID, modeScheduler)
	handler.now = func() time.Time { return time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC) }

	summary, err := handler.runScheduler(context.Background())
	if err != nil {
		t.Fatalf("runScheduler returned error: %v", err)
	}
	if summary.Enqueued != 1 || len(sqs.sentMessages) != 1 || len(repo.works) != 1 {
		t.Fatalf("summary = %+v, messages = %d, works = %d, want one enqueued", summary, len(sqs.sentMessages), len(repo.works))
	}
}

func TestRunSchedulerTreatsZeroIntervalAsAlwaysDue(t *testing.T) {
	repo := newFakeRuntimeRepository()
	repo.config = checkexecution.SchedulerConfig{RecurringEnabled: true}
	monitor := testMonitor("https://example.com", true)
	monitor.IntervalSeconds = 0
	repo.monitors["auth/public-http"] = monitor
	repo.lastExec["auth/public-http"] = time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	sqs := &fakeSQSClient{}
	handler := newRuntimeHandler(repo, sqs, "queue-url", "", defaultTenantID, modeScheduler)
	handler.now = func() time.Time { return time.Date(2026, 5, 22, 12, 0, 1, 0, time.UTC) }

	summary, err := handler.runScheduler(context.Background())
	if err != nil {
		t.Fatalf("runScheduler returned error: %v", err)
	}
	if summary.Enqueued != 1 || len(sqs.sentMessages) != 1 || len(repo.works) != 1 {
		t.Fatalf("summary = %+v, messages = %d, works = %d, want one enqueued", summary, len(sqs.sentMessages), len(repo.works))
	}
}

func TestRunWorkerSkipsDisabledMonitor(t *testing.T) {
	repo := newFakeRuntimeRepository()
	repo.monitors["auth/public-http"] = testMonitor("https://example.com", false)
	repo.works = []checkexecution.ExecutionWork{{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_1", Trigger: checkexecution.TriggerTypeManual, RequestedAt: time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC), Status: checkexecution.ExecutionWorkPending}}
	sqs := &fakeSQSClient{}
	handler := newRuntimeHandler(repo, sqs, "", "", defaultTenantID, modeWorker)
	handler.now = func() time.Time { return time.Date(2026, 5, 22, 12, 0, 1, 0, time.UTC) }

	summary, err := handler.runWorker(context.Background())
	if err != nil {
		t.Fatalf("runWorker returned error: %v", err)
	}
	if summary.Skipped != 1 || len(repo.skipped) != 1 {
		t.Fatalf("summary = %+v, skipped = %d, want one skipped work", summary, len(repo.skipped))
	}
	if len(repo.results) != 0 {
		t.Fatalf("results = %d, want 0", len(repo.results))
	}
}

func TestRunWorkerProcessesManualRunIntoRecordedResult(t *testing.T) {
	repo := newFakeRuntimeRepository()
	repo.monitors["auth/public-http"] = testMonitor("https://status.example.com", true)
	repo.works = []checkexecution.ExecutionWork{{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_MANUAL_1", Trigger: checkexecution.TriggerTypeManual, RequestedAt: time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC), Status: checkexecution.ExecutionWorkPending}}
	sqs := &fakeSQSClient{}
	handler := newRuntimeHandler(repo, sqs, "", "", defaultTenantID, modeWorker)
	handler.executor = fakeExecutor{response: outboundhttp.Response{StatusCode: 200, Body: []byte("ok")}}

	summary, err := handler.runWorker(context.Background())
	if err != nil {
		t.Fatalf("runWorker returned error: %v", err)
	}
	if summary.Processed != 1 || len(repo.results) != 1 {
		t.Fatalf("summary = %+v, results = %d, want one processed result", summary, len(repo.results))
	}
	if repo.results[0].RunID != "RUN_MANUAL_1" {
		t.Fatalf("RunID = %q, want RUN_MANUAL_1", repo.results[0].RunID)
	}
	if repo.results[0].Trigger != checkexecution.TriggerTypeManual {
		t.Fatalf("Trigger = %q, want manual", repo.results[0].Trigger)
	}
}

func TestRunWorkerRecordsUnsafePersistedTargetWithoutDialing(t *testing.T) {
	repo := newFakeRuntimeRepository()
	repo.monitors["auth/public-http"] = testMonitor("http://127.0.0.1", true)
	repo.works = []checkexecution.ExecutionWork{{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_UNSAFE_1", Trigger: checkexecution.TriggerTypeManual, RequestedAt: time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC), Status: checkexecution.ExecutionWorkPending}}
	handler := newRuntimeHandler(repo, &fakeSQSClient{}, "", "", defaultTenantID, modeWorker)
	dialer := &countingDialer{}
	handler.executor = &outboundhttp.Executor{Dialer: dialer}

	summary, err := handler.runWorker(context.Background())
	if err != nil {
		t.Fatalf("runWorker returned error: %v", err)
	}
	if summary.Processed != 1 || len(repo.results) != 1 {
		t.Fatalf("summary = %+v, results = %d", summary, len(repo.results))
	}
	result := repo.results[0]
	if result.Outcome != checkexecution.OutcomeError || result.FailureCode != "address_blocked" {
		t.Fatalf("result = %#v", result)
	}
	if dialer.calls != 0 {
		t.Fatalf("unsafe stored target dialed %d times", dialer.calls)
	}
}

func TestHandleSQSEventRecordsUnsafeQueuedTargetWithoutDialing(t *testing.T) {
	repo := newFakeRuntimeRepository()
	dialer := &countingDialer{}
	handler := newRuntimeHandler(repo, &fakeSQSClient{}, "", "", defaultTenantID, modeWorker)
	handler.executor = &outboundhttp.Executor{Dialer: dialer}
	request := checkexecution.ExecutionRequest{Monitor: testMonitor("http://127.0.0.1", true), RunID: "RUN_QUEUE_UNSAFE", Trigger: checkexecution.TriggerTypeManual}
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	summary, err := handler.handleSQSEvent(context.Background(), events.SQSEvent{Records: []events.SQSMessage{{Body: string(body)}}})
	if err != nil {
		t.Fatalf("handleSQSEvent returned error: %v", err)
	}
	if summary.Processed != 1 || len(repo.results) != 1 || repo.results[0].FailureCode != string(outboundhttp.KindAddressBlocked) {
		t.Fatalf("summary = %+v, results = %#v", summary, repo.results)
	}
	if dialer.calls != 0 {
		t.Fatalf("unsafe queued target dialed %d times", dialer.calls)
	}
}

func TestIncidentSummaryUsesErrorDetails(t *testing.T) {
	statusCode := 503
	summary := incidentSummary(testMonitor("https://example.com", true), checkexecution.ExecutionResult{ServiceID: "auth", MonitorID: "public-http", Outcome: checkexecution.OutcomeFailure, StatusCode: &statusCode, Error: "unexpected status code 503"})
	if summary == "" || summary == "Homepage failed" {
		t.Fatalf("summary = %q, want detailed summary", summary)
	}
}

func TestBuildEscalationMessageUsesIncidentDownEvent(t *testing.T) {
	monitor := testMonitor("https://example.com", true)
	service := monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth"}
	statusCode := 503
	body := buildEscalationMessage("incident.down", monitor, service, "INC_1", checkexecution.ExecutionResult{FinishedAt: time.Date(2026, 5, 22, 12, 0, 2, 0, time.UTC), Error: "boom", StatusCode: &statusCode})
	if body == "" {
		t.Fatal("body is empty")
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if payload["eventType"] != "incident.down" {
		t.Fatalf("eventType = %v, want incident.down", payload["eventType"])
	}
	if !strings.Contains(payload["message"].(string), "DOWN") {
		t.Fatalf("message = %q, want DOWN", payload["message"])
	}
}

func TestBuildEscalationMessageUsesIncidentUpEvent(t *testing.T) {
	monitor := testMonitor("https://example.com", true)
	service := monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth"}
	body := buildEscalationMessage("incident.up", monitor, service, "INC_1", checkexecution.ExecutionResult{FinishedAt: time.Date(2026, 5, 22, 12, 0, 2, 0, time.UTC)})
	if body == "" {
		t.Fatal("body is empty")
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if payload["eventType"] != "incident.up" {
		t.Fatalf("eventType = %v, want incident.up", payload["eventType"])
	}
	if !strings.Contains(payload["message"].(string), "UP") {
		t.Fatalf("message = %q, want UP", payload["message"])
	}
}
