package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	sharedaws "bolt-monitor/shared/aws"
)

type recordingScheduler struct {
	calls []*sharedaws.SchedulerCreateScheduleInput
	err   error
}

func (f *recordingScheduler) CreateSchedule(_ context.Context, input *sharedaws.SchedulerCreateScheduleInput) (*sharedaws.SchedulerCreateScheduleOutput, error) {
	f.calls = append(f.calls, input)
	return &sharedaws.SchedulerCreateScheduleOutput{}, f.err
}

func (f *recordingScheduler) UpdateSchedule(context.Context, *sharedaws.SchedulerUpdateScheduleInput) (*sharedaws.SchedulerUpdateScheduleOutput, error) {
	return nil, nil
}

func (f *recordingScheduler) DeleteSchedule(context.Context, *sharedaws.SchedulerDeleteScheduleInput) (*sharedaws.SchedulerDeleteScheduleOutput, error) {
	return nil, nil
}

func (f *recordingScheduler) GetSchedule(context.Context, *sharedaws.SchedulerGetScheduleInput) (*sharedaws.SchedulerGetScheduleOutput, error) {
	return nil, nil
}

func TestOneTimeSchedulerBuildsDeterministicNameAndDeleteConfig(t *testing.T) {
	client := &recordingScheduler{}
	scheduler := newOneTimeScheduler(client, "group", "arn:aws:iam::1:role/exec", "arn:aws:sqs:us-east-1:1:notif", "arn:aws:sqs:us-east-1:1:notif-dlq", sharedaws.SchedulerRetryPolicy{})
	when := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	if err := scheduler.ScheduleNextStep(context.Background(), scheduledInvocationEvent{IncidentID: "INC_1", Step: 2}, when); err != nil {
		t.Fatalf("schedule failed: %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("expected one call, got %d", len(client.calls))
	}
	call := client.calls[0]
	if !strings.HasPrefix(*call.Name, "escstep-") {
		t.Fatalf("schedule name prefix wrong: %q", *call.Name)
	}
	if *call.GroupName != "group" {
		t.Fatalf("group name = %q", *call.GroupName)
	}
	if call.ActionAfterCompletion != sharedaws.SchedulerActionAfterCompletionDelete {
		t.Fatalf("action after completion = %v, want delete", call.ActionAfterCompletion)
	}
	if call.FlexibleTimeWindow == nil || call.FlexibleTimeWindow.Mode != sharedaws.FlexibleTimeWindowModeOff {
		t.Fatalf("flexible window must be off, got %+v", call.FlexibleTimeWindow)
	}
	if call.Target == nil || call.Target.Arn == nil || *call.Target.Arn != "arn:aws:sqs:us-east-1:1:notif" {
		t.Fatalf("target ARN wrong: %+v", call.Target)
	}
	if call.Target.DeadLetterConfig == nil || *call.Target.DeadLetterConfig.Arn != "arn:aws:sqs:us-east-1:1:notif-dlq" {
		t.Fatalf("dlq not configured")
	}
	if !strings.HasPrefix(*call.ScheduleExpression, "at(") {
		t.Fatalf("schedule expression = %q", *call.ScheduleExpression)
	}
}

func TestOneTimeSchedulerNameIsDeterministic(t *testing.T) {
	a, _ := oneTimeScheduleName("INC_1", 2)
	b, _ := oneTimeScheduleName("inc_1", 2)
	if a != b {
		t.Fatalf("expected deterministic name, got %q vs %q", a, b)
	}
	c, _ := oneTimeScheduleName("INC_1", 3)
	if a == c {
		t.Fatalf("different step must produce different name")
	}
	if _, err := oneTimeScheduleName("", 2); err == nil {
		t.Fatal("missing incident id must fail")
	}
}

func TestOneTimeSchedulerPayloadEncodesCanonicalShape(t *testing.T) {
	client := &recordingScheduler{}
	scheduler := newOneTimeScheduler(client, "group", "arn:aws:iam::1:role/exec", "arn:aws:sqs:us-east-1:1:notif", "", sharedaws.SchedulerRetryPolicy{})
	when := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	if err := scheduler.ScheduleNextStep(context.Background(), scheduledInvocationEvent{IncidentID: "INC_42", Step: 3}, when); err != nil {
		t.Fatalf("schedule failed: %v", err)
	}
	body := *client.calls[0].Target.Input
	var payload scheduledStepPayload
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.Version != "1" || payload.Kind != "scheduled_step" || payload.SourceKind != "scheduler_target" {
		t.Fatalf("payload = %+v", payload)
	}
	if payload.StepNumber != 3 || payload.IncidentID != "INC_42" {
		t.Fatalf("payload mismatch: %+v", payload)
	}
}

func TestOneTimeSchedulerDetectsConflictOnDuplicate(t *testing.T) {
	client := &recordingScheduler{err: sharedaws.NewConflictException("ScheduleAlreadyExists")}
	scheduler := newOneTimeScheduler(client, "group", "arn:aws:iam::1:role/exec", "arn:aws:sqs:us-east-1:1:notif", "", sharedaws.SchedulerRetryPolicy{})
	err := scheduler.ScheduleNextStep(context.Background(), scheduledInvocationEvent{IncidentID: "INC_DUP", Step: 1}, time.Now())
	if err == nil || !strings.Contains(err.Error(), "ScheduleAlreadyExists") {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

type recordingAdapterQueue struct {
	bodies []string
}

func (q *recordingAdapterQueue) SendMessage(_ context.Context, input *sharedaws.SQSSendMessageInput) (*sharedaws.SQSSendMessageOutput, error) {
	q.bodies = append(q.bodies, sharedaws.ToString(input.MessageBody))
	return &sharedaws.SQSSendMessageOutput{}, nil
}

func (q *recordingAdapterQueue) ReceiveMessage(context.Context, *sharedaws.SQSReceiveMessageInput) (*sharedaws.SQSReceiveMessageOutput, error) {
	return nil, nil
}

func (q *recordingAdapterQueue) DeleteMessage(context.Context, *sharedaws.SQSDeleteMessageInput) (*sharedaws.SQSDeleteMessageOutput, error) {
	return nil, nil
}

func (q *recordingAdapterQueue) ChangeMessageVisibility(context.Context, *sharedaws.SQSChangeMessageVisibilityInput) (*sharedaws.SQSChangeMessageVisibilityOutput, error) {
	return nil, nil
}

func TestLegacyScheduleAdapterReenqueuesCanonicalWork(t *testing.T) {
	queue := &recordingAdapterQueue{}
	adapter := newLegacyScheduleAdapter(queue, "queue-url")
	if err := adapter.Reenqueue(context.Background(), scheduledInvocationEvent{IncidentID: "INC_LEGACY", Step: 4}); err != nil {
		t.Fatalf("reenqueue failed: %v", err)
	}
	if len(queue.bodies) != 1 {
		t.Fatalf("expected one body, got %d", len(queue.bodies))
	}
	var payload scheduledStepPayload
	if err := json.Unmarshal([]byte(queue.bodies[0]), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.Kind != "scheduled_step" || payload.SourceKind != "scheduler_target" || payload.StepNumber != 4 {
		t.Fatalf("payload = %+v", payload)
	}
}

func TestScheduleInvocationPayloadHasRequiredFields(t *testing.T) {
	payload := scheduledStepPayloadFromEvent(scheduledInvocationEvent{IncidentID: "INC_REQ", Step: 5})
	if payload.Version != "1" || payload.Kind != "scheduled_step" {
		t.Fatalf("payload missing canonical fields: %+v", payload)
	}
	if payload.SourceKind != "scheduler_target" {
		t.Fatalf("source kind = %q", payload.SourceKind)
	}
}
