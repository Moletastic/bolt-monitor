package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/notifications"
)

type oneTimeScheduler struct {
	client               sharedaws.SchedulerAPI
	scheduleGroup        string
	executionRoleArn     string
	queueArn             string
	dlqArn               string
	schedulerTargetRetry sharedaws.SchedulerRetryPolicy
}

func newOneTimeScheduler(client sharedaws.SchedulerAPI, scheduleGroup, executionRoleArn, queueArn, dlqArn string, retry sharedaws.SchedulerRetryPolicy) *oneTimeScheduler {
	return &oneTimeScheduler{
		client:               client,
		scheduleGroup:        scheduleGroup,
		executionRoleArn:     executionRoleArn,
		queueArn:             queueArn,
		dlqArn:               dlqArn,
		schedulerTargetRetry: retry,
	}
}

func (s *oneTimeScheduler) ScheduleNextStep(ctx context.Context, event scheduledInvocationEvent, when time.Time) error {
	if s == nil {
		return nil
	}
	if strings.TrimSpace(s.queueArn) == "" {
		return fmt.Errorf("scheduler requires notification queue ARN")
	}
	if strings.TrimSpace(s.executionRoleArn) == "" {
		return fmt.Errorf("scheduler requires execution role ARN")
	}
	name, err := oneTimeScheduleName(event.IncidentID, event.Step)
	if err != nil {
		return err
	}
	payload := scheduledStepPayloadFromEvent(event)
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	target := sharedaws.SchedulerTarget{
		Arn:     sharedaws.String(strings.TrimSpace(s.queueArn)),
		RoleArn: sharedaws.String(strings.TrimSpace(s.executionRoleArn)),
		Input:   sharedaws.String(string(body)),
	}
	if strings.TrimSpace(s.dlqArn) != "" {
		target.DeadLetterConfig = &sharedaws.SchedulerDeadLetterConfig{Arn: sharedaws.String(s.dlqArn)}
	}
	if s.schedulerTargetRetry != (sharedaws.SchedulerRetryPolicy{}) {
		target.RetryPolicy = &s.schedulerTargetRetry
	}
	input := &sharedaws.SchedulerCreateScheduleInput{
		Name:                       sharedaws.String(name),
		GroupName:                  sharedaws.String(s.scheduleGroup),
		ScheduleExpression:         sharedaws.String(oneTimeScheduleExpression(when)),
		ScheduleExpressionTimezone: sharedaws.String("UTC"),
		Target:                     &target,
		FlexibleTimeWindow:         &sharedaws.SchedulerFlexibleTimeWindow{Mode: sharedaws.FlexibleTimeWindowModeOff},
		ActionAfterCompletion:      sharedaws.SchedulerActionAfterCompletionDelete,
		State:                      sharedaws.ScheduleStateEnabled,
	}
	_, err = s.client.CreateSchedule(ctx, input)
	return err
}

type scheduledStepPayload struct {
	Version      string `json:"version"`
	Kind         string `json:"kind"`
	SourceKind   string `json:"sourceKind"`
	TenantID     string `json:"tenantId"`
	IncidentID   string `json:"incidentId"`
	TransitionID string `json:"transitionId"`
	StepNumber   int    `json:"stepNumber"`
}

func scheduledStepPayloadFromEvent(event scheduledInvocationEvent) scheduledStepPayload {
	return scheduledStepPayload{
		Version:      notifications.CanonicalEnvelopeVersion,
		Kind:         notifications.CanonicalKindScheduled,
		SourceKind:   notifications.CanonicalSourceSchedule,
		IncidentID:   event.IncidentID,
		TransitionID: event.IncidentID,
		StepNumber:   event.Step,
	}
}

func oneTimeScheduleName(incidentID string, step int) (string, error) {
	if strings.TrimSpace(incidentID) == "" {
		return "", fmt.Errorf("schedule name requires incident id")
	}
	digest := sha256.Sum256([]byte(fmt.Sprintf("%s\n%d", strings.ToLower(strings.TrimSpace(incidentID)), step)))
	return "escstep-" + hex.EncodeToString(digest[:16]), nil
}

func oneTimeScheduleExpression(when time.Time) string {
	utc := when.UTC().Format("2006-01-02T15:04:05")
	return fmt.Sprintf("at(%s)", utc)
}

// legacyScheduleAdapter re-enqueues canonical scheduled-step work for any
// legacy direct invocation delivered to the runtime. It preserves the
// versioned canonical envelope so the SQS handler treats it identically to a
// fresh schedule invocation.
type legacyScheduleAdapter struct {
	queue    sharedaws.SQSAPI
	queueURL string
}

func newLegacyScheduleAdapter(queue sharedaws.SQSAPI, queueURL string) *legacyScheduleAdapter {
	return &legacyScheduleAdapter{queue: queue, queueURL: queueURL}
}

func (a *legacyScheduleAdapter) Reenqueue(ctx context.Context, event scheduledInvocationEvent) error {
	if a == nil {
		return nil
	}
	body, err := json.Marshal(scheduledStepPayloadFromEvent(event))
	if err != nil {
		return err
	}
	if _, err := a.queue.SendMessage(ctx, &sharedaws.SQSSendMessageInput{QueueUrl: sharedaws.String(a.queueURL), MessageBody: sharedaws.String(string(body))}); err != nil {
		return err
	}
	return nil
}
