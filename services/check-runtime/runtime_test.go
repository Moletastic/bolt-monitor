package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbrecord"
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
	config             checkexecution.SchedulerConfig
	monitors           map[string]monitorconfig.Monitor
	services           map[string]monitorconfig.Service
	works              []checkexecution.ExecutionWork
	claims             map[string]bool
	skipped            []checkexecution.ExecutionWork
	results            []checkexecution.ExecutionResult
	recordedWorks      []checkexecution.ExecutionWork
	published          []string
	lastExec           map[string]time.Time
	statuses           map[string]resultstatus.MonitorStatus
	dispatchBuckets    []string
	publicationBuckets []string
	recordErr          map[string]error
	listErr            error
}

func newFakeRuntimeRepository() *fakeRuntimeRepository {
	return &fakeRuntimeRepository{monitors: map[string]monitorconfig.Monitor{}, claims: map[string]bool{}, lastExec: map[string]time.Time{}, statuses: map[string]resultstatus.MonitorStatus{}, recordErr: map[string]error{}}
}

func (r *fakeRuntimeRepository) GetSchedulerConfig(context.Context, string) (checkexecution.SchedulerConfig, error) {
	return r.config, nil
}

func (r *fakeRuntimeRepository) ListMonitors(ctx context.Context, _ string) ([]monitorconfig.Monitor, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
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

func (r *fakeRuntimeRepository) EnqueueExecutionRequests(_ context.Context, requests []checkexecution.ExecutionRequest, now time.Time) (int, error) {
	created := 0
	for _, request := range requests {
		runID := request.RunID
		if runID == "" {
			runID = newRunID(now)
		}
		acceptedAt := request.AcceptedAt
		if acceptedAt.IsZero() {
			acceptedAt = now.UTC()
		}
		for _, existing := range r.works {
			if existing.RunID == runID {
				return created, nil
			}
		}
		r.works = append(r.works, checkexecution.ExecutionWork{TenantID: request.Monitor.TenantID, ServiceID: request.Monitor.ServiceID, MonitorID: request.Monitor.MonitorID, RunID: runID, Trigger: request.Trigger, AcceptedAt: acceptedAt, RequestedAt: acceptedAt, ScheduleDefinitionVersion: request.ScheduleDefinitionVersion, ScheduledFor: request.ScheduledFor, Status: checkexecution.ExecutionWorkPending})
		created++
	}
	return created, nil
}

func (r *fakeRuntimeRepository) AcknowledgeExecutionPublication(_ context.Context, work checkexecution.ExecutionWork) error {
	r.published = append(r.published, work.RunID)
	return nil
}

func (r *fakeRuntimeRepository) LoadExecutionWork(_ context.Context, tenantID, runID string) (checkexecution.ExecutionWork, bool, error) {
	for _, work := range r.works {
		if strings.EqualFold(work.TenantID, tenantID) && strings.EqualFold(work.RunID, runID) {
			return work, true, nil
		}
	}
	return checkexecution.ExecutionWork{}, false, nil
}

func (r *fakeRuntimeRepository) ListPendingExecutionWork(context.Context, string, int32) ([]checkexecution.ExecutionWork, error) {
	return append([]checkexecution.ExecutionWork(nil), r.works...), nil
}

func (r *fakeRuntimeRepository) ListPublicationMarkers(_ context.Context, _ string, bucketShard string, _ int32, _ map[string]sharedaws.AttributeValue) ([]dynamodbrecord.ExecutionMarkerRecord, map[string]sharedaws.AttributeValue, error) {
	r.publicationBuckets = append(r.publicationBuckets, bucketShard)
	return nil, nil, nil
}

func (r *fakeRuntimeRepository) ListDispatchPending(_ context.Context, _ string, bucketShard string, _ int32, _ map[string]sharedaws.AttributeValue) ([]dynamodbrecord.DispatchPendingRecord, map[string]sharedaws.AttributeValue, error) {
	r.dispatchBuckets = append(r.dispatchBuckets, bucketShard)
	return nil, nil, nil
}

func (r *fakeRuntimeRepository) RemoveDispatchPending(context.Context, string, string, string, string) error {
	return nil
}

func (r *fakeRuntimeRepository) LoadTransitionOutbox(context.Context, string, string) (dynamodbrecord.TransitionOutboxRecord, bool, error) {
	return dynamodbrecord.TransitionOutboxRecord{}, false, nil
}

func (r *fakeRuntimeRepository) ClaimExecutionWork(_ context.Context, work checkexecution.ExecutionWork, now time.Time) (checkexecution.ExecutionWork, bool, error) {
	if r.claims[work.RunID] {
		return work, false, nil
	}
	r.claims[work.RunID] = true
	work.FencingToken = "LEASE_TEST"
	leaseUntil := now.UTC().Add(defaultExecutionWorkLeaseDuration)
	work.LeaseUntil = &leaseUntil
	work.Status = checkexecution.ExecutionWorkInProgress
	return work, true, nil
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
	if err := r.recordErr[work.RunID]; err != nil {
		return "", "", err
	}
	r.recordedWorks = append(r.recordedWorks, work)
	r.results = append(r.results, result)
	return "", "", nil
}

func (r *fakeRuntimeRepository) LoadExecutionResultState(_ context.Context, result checkexecution.ExecutionResult) (executionResultState, error) {
	status, found := r.statuses[result.ServiceID+"/"+result.MonitorID]
	return executionResultState{status: status, statusFound: found}, nil
}

func (r *fakeRuntimeRepository) CommitExecutionResult(_ context.Context, _ monitorconfig.Monitor, work checkexecution.ExecutionWork, result checkexecution.ExecutionResult, _ []any, _ resultstatus.MonitorStatus, _ bool, _ executionResultPublication) error {
	if err := r.recordErr[work.RunID]; err != nil {
		return err
	}
	r.recordedWorks = append(r.recordedWorks, work)
	r.results = append(r.results, result)
	return nil
}

func (r *fakeRuntimeRepository) GetMonitorStatus(_ context.Context, _, serviceID, monitorID string) (resultstatus.MonitorStatus, bool, error) {
	status, ok := r.statuses[serviceID+"/"+monitorID]
	return status, ok, nil
}

func testMonitor(target string, enabled bool) monitorconfig.Monitor {
	return monitorconfig.Monitor{ServiceID: "auth", MonitorID: "public-http", TenantID: defaultTenantID, Name: "Homepage", Type: monitorconfig.MonitorTypeHTTP, IntervalSeconds: 60, Enabled: enabled, FailureThreshold: 1, RecoveryThreshold: 1, HTTP: &monitorconfig.HTTPConfiguration{Target: target, Method: "GET", TimeoutMs: 5000, ExpectedStatusCodes: []int{200}}}
}

func TestRunSchedulerStopsAtDiscoveryDeadline(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	repo := newFakeRuntimeRepository()
	repo.config = checkexecution.SchedulerConfig{RecurringEnabled: true}
	repo.listErr = context.DeadlineExceeded
	handler := newRuntimeHandler(repo, &fakeSQSClient{}, "execution-queue", "", defaultTenantID, modeScheduler)
	handler.now = func() time.Time { return now }
	handler.schedulerDeadline = 5 * time.Second

	_, err := handler.runScheduler(context.Background())
	if err == nil {
		t.Fatal("expected typed retryable deadline error")
	}
	var failure *checkexecution.RuntimeFailure
	if !errors.As(err, &failure) || !failure.Retryable || failure.Code != checkexecution.FailurePublication {
		t.Fatalf("failure = %#v", failure)
	}
}

func TestExecutionResultCommandRejectsIdentityMismatch(t *testing.T) {
	repo := newFakeRuntimeRepository()
	monitor := testMonitor("https://example.com", true)
	work := checkexecution.ExecutionWork{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_1", Trigger: checkexecution.TriggerTypeManual, AcceptedAt: time.Now()}
	mismatched := checkexecution.ExecutionResult{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_OTHER", Outcome: checkexecution.OutcomeSuccess, FinishedAt: time.Now()}

	if _, _, err := newExecutionResultCommand(repo, systemExecutionResultClock{}, generatedExecutionResultIDs{}).execute(context.Background(), monitor, work, mismatched); err == nil || !strings.Contains(err.Error(), "immutable_identity_conflict") {
		t.Fatalf("expected identity conflict, got %v", err)
	}
}

func TestRunSchedulerIgnoresExpiredFakeClockForDeadline(t *testing.T) {
	repo := newFakeRuntimeRepository()
	repo.config = checkexecution.SchedulerConfig{RecurringEnabled: true}
	repo.monitors["auth/public-http"] = testMonitor("https://example.com", true)
	handler := newRuntimeHandler(repo, &fakeSQSClient{}, "queue-url", "", defaultTenantID, modeScheduler)
	handler.schedulerDeadline = 5 * time.Second
	now := time.Now().UTC()
	handler.now = func() time.Time { return now }

	summary, err := handler.runScheduler(context.Background())
	if err != nil {
		t.Fatalf("runScheduler returned error: %v", err)
	}
	if summary.Enqueued != 1 {
		t.Fatalf("summary = %+v", summary)
	}
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

func TestHandleSQSEventSkipsDeletedMonitorDurably(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	work := checkexecution.ExecutionWork{TenantID: defaultTenantID, ServiceID: "service", MonitorID: "monitor", RunID: "run", Trigger: checkexecution.TriggerTypeRecurring, AcceptedAt: now, Status: checkexecution.ExecutionWorkPending}
	repo := newFakeRuntimeRepository()
	repo.works = []checkexecution.ExecutionWork{work}
	handler := runtimeHandler{repo: repo, tenantID: defaultTenantID, now: func() time.Time { return now }}
	body, err := json.Marshal(checkexecution.ExecutionRequest{Monitor: monitorconfig.Monitor{TenantID: defaultTenantID, ServiceID: "service", MonitorID: "monitor"}, RunID: "run", Trigger: checkexecution.TriggerTypeRecurring, AcceptedAt: now})
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	summary, err := handler.handleSQSEvent(context.Background(), events.SQSEvent{Records: []events.SQSMessage{{Body: string(body)}}})
	if err != nil {
		t.Fatalf("handleSQSEvent returned error: %v", err)
	}
	if summary.Skipped != 1 || len(repo.skipped) != 1 || repo.skipped[0].LastError != "monitor not found" {
		t.Fatalf("summary = %+v, skipped = %#v", summary, repo.skipped)
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
	repo.monitors["auth/public-http"] = request.Monitor
	repo.works = []checkexecution.ExecutionWork{{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: request.RunID, Trigger: request.Trigger, RequestedAt: time.Now().UTC(), Status: checkexecution.ExecutionWorkPending}}
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	summary, err := handler.handleSQSEvent(context.Background(), events.SQSEvent{Records: []events.SQSMessage{{Body: string(body)}}})
	if err != nil {
		t.Fatalf("handleSQSEvent returned error: %v", err)
	}
	if summary.Skipped != 1 || len(repo.results) != 0 {
		t.Fatalf("summary = %+v, results = %#v", summary, repo.results)
	}
	if dialer.calls != 0 {
		t.Fatalf("unsafe queued target dialed %d times", dialer.calls)
	}
}

func TestHandleSQSEventBatchFailsOnlyMalformedRecord(t *testing.T) {
	repo := newFakeRuntimeRepository()
	handler := newRuntimeHandler(repo, &fakeSQSClient{}, "", "", defaultTenantID, modeWorker)
	response, err := handler.handleSQSEventBatch(context.Background(), events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "bad", Body: "{"},
		{MessageId: "duplicate", Body: `{"monitor":{"tenantId":"DEFAULT"},"runId":"RUN_MISSING","trigger":"manual"}`},
	}})
	if err != nil {
		t.Fatalf("handleSQSEventBatch returned error: %v", err)
	}
	if len(response.BatchItemFailures) != 1 || response.BatchItemFailures[0].ItemIdentifier != "bad" {
		t.Fatalf("batch failures = %#v", response.BatchItemFailures)
	}
}

func TestHandleSQSEventBatchKeepsSuccessfulRecordsAcknowledged(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	repo := newFakeRuntimeRepository()
	monitor := testMonitor("https://example.com", true)
	repo.monitors["auth/public-http"] = monitor
	repo.works = []checkexecution.ExecutionWork{
		{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_OK", Trigger: checkexecution.TriggerTypeManual, AcceptedAt: now, Status: checkexecution.ExecutionWorkPending},
		{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_RETRY", Trigger: checkexecution.TriggerTypeManual, AcceptedAt: now, Status: checkexecution.ExecutionWorkPending},
	}
	repo.recordErr["RUN_RETRY"] = checkexecution.Storage("commit-result", "RUN_RETRY")
	handler := newRuntimeHandler(repo, &fakeSQSClient{}, "", "", defaultTenantID, modeWorker)
	handler.now = func() time.Time { return now }
	request := func(runID string) string {
		body, err := json.Marshal(checkexecution.ExecutionRequest{Monitor: monitor, RunID: runID, Trigger: checkexecution.TriggerTypeManual, AcceptedAt: now})
		if err != nil {
			t.Fatalf("Marshal returned error: %v", err)
		}
		return string(body)
	}

	response, err := handler.handleSQSEventBatch(context.Background(), events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "ok", Body: request("RUN_OK")},
		{MessageId: "retry", Body: request("RUN_RETRY")},
	}})
	if err != nil {
		t.Fatalf("handleSQSEventBatch returned error: %v", err)
	}
	if len(response.BatchItemFailures) != 1 || response.BatchItemFailures[0].ItemIdentifier != "retry" {
		t.Fatalf("failures = %#v", response.BatchItemFailures)
	}
	if len(repo.results) != 1 || repo.results[0].RunID != "RUN_OK" {
		t.Fatalf("results = %#v", repo.results)
	}
}

func TestReconcileDispatchPendingQueriesCurrentHourlyBucket(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 34, 0, 0, time.UTC)
	repo := newFakeRuntimeRepository()
	handler := newRuntimeHandler(repo, &fakeSQSClient{}, "", "escalation-queue", defaultTenantID, modeScheduler)
	handler.now = func() time.Time { return now }

	if _, err := handler.reconcileDispatchPending(context.Background()); err != nil {
		t.Fatalf("reconcileDispatchPending returned error: %v", err)
	}
	if len(repo.dispatchBuckets) != dispatchPendingShards {
		t.Fatalf("queried buckets = %#v", repo.dispatchBuckets)
	}
	for _, bucket := range repo.dispatchBuckets {
		if !strings.HasPrefix(bucket, "2026071912|") {
			t.Fatalf("bucket = %q, want current hourly bucket", bucket)
		}
	}
}

func TestRunSchedulerDuplicateInvocationDerivesSameIdentity(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Minute)
	repo := newFakeRuntimeRepository()
	repo.config = checkexecution.SchedulerConfig{RecurringEnabled: true}
	repo.monitors["auth/public-http"] = testMonitor("https://example.com", true)
	sqs := &fakeSQSClient{}
	handler := newRuntimeHandler(repo, sqs, "queue-url", "", defaultTenantID, modeScheduler)
	handler.now = func() time.Time { return now }
	handler.schedulerDeadline = 5 * time.Second

	first, err := handler.runScheduler(context.Background())
	if err != nil {
		t.Fatalf("first runScheduler returned error: %v", err)
	}
	second, err := handler.runScheduler(context.Background())
	if err != nil {
		t.Fatalf("second runScheduler returned error: %v", err)
	}
	if first.Enqueued != 1 || second.Enqueued != 0 {
		t.Fatalf("expected one enqueue across invocations, got %+v / %+v", first, second)
	}
}

func TestHandleSQSEventRejectsDuplicateConcurrentDelivery(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	monitor := testMonitor("https://example.com", true)
	work := checkexecution.ExecutionWork{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", RunID: "RUN_DUP", Trigger: checkexecution.TriggerTypeManual, AcceptedAt: now, Status: checkexecution.ExecutionWorkPending}
	repo := newFakeRuntimeRepository()
	repo.monitors["auth/public-http"] = monitor
	repo.works = []checkexecution.ExecutionWork{work}
	repo.claims["RUN_DUP"] = true
	handler := newRuntimeHandler(repo, &fakeSQSClient{}, "", "", defaultTenantID, modeWorker)
	handler.now = func() time.Time { return now }
	body, _ := json.Marshal(checkexecution.ExecutionRequest{Monitor: monitor, RunID: "RUN_DUP", Trigger: checkexecution.TriggerTypeManual, AcceptedAt: now})
	response, err := handler.handleSQSEventBatch(context.Background(), events.SQSEvent{Records: []events.SQSMessage{{MessageId: "dup", Body: string(body)}}})
	if err != nil {
		t.Fatalf("handleSQSEventBatch returned error: %v", err)
	}
	if len(response.BatchItemFailures) != 0 {
		t.Fatalf("expected successful ack, got failures = %#v", response.BatchItemFailures)
	}
	if len(repo.results) != 0 {
		t.Fatalf("expected no execution, got results = %#v", repo.results)
	}
}

func TestRuntimeOutcomeLogRedactsSecrets(t *testing.T) {
	buf := &bytes.Buffer{}
	previous := runtimeLogWriter
	runtimeLogWriter = buf
	defer func() { runtimeLogWriter = previous }()

	runtimeOutcomeLog(outcomeCompleted, "commit-result", "DEFAULT", "RUN_1", "monitor", "secret-with-newline\nvalue")

	entry := map[string]string{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("log entry did not decode: %v", err)
	}
	if entry["reason"] != "" {
		t.Fatalf("expected redacted reason, got %q", entry["reason"])
	}
	if entry["outcome"] != string(outcomeCompleted) {
		t.Fatalf("outcome = %q, want %q", entry["outcome"], outcomeCompleted)
	}
}

func TestRecoverPublicationMarkersQueriesBoundedHourlyBuckets(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 34, 0, 0, time.UTC)
	repo := newFakeRuntimeRepository()
	handler := newRuntimeHandler(repo, &fakeSQSClient{}, "execution-queue", "", defaultTenantID, modeScheduler)
	handler.now = func() time.Time { return now }

	if _, err := handler.recoverPublicationMarkers(context.Background()); err != nil {
		t.Fatalf("recoverPublicationMarkers returned error: %v", err)
	}
	if len(repo.publicationBuckets) != recoveryBucketsPerHour*recoveryShards {
		t.Fatalf("queried buckets = %#v", repo.publicationBuckets)
	}
	if !strings.HasPrefix(repo.publicationBuckets[0], "2026071912|") {
		t.Fatalf("first bucket = %q, want current hourly bucket", repo.publicationBuckets[0])
	}
	if !strings.HasPrefix(repo.publicationBuckets[recoveryShards], "2026071911|") {
		t.Fatalf("overlap bucket = %q, want prior hourly bucket", repo.publicationBuckets[recoveryShards])
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
