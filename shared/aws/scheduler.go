package aws

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	schedulertypes "github.com/aws/aws-sdk-go-v2/service/scheduler/types"
)

type SchedulerAPI interface {
	CreateSchedule(ctx context.Context, params *SchedulerCreateScheduleInput) (*SchedulerCreateScheduleOutput, error)
	UpdateSchedule(ctx context.Context, params *SchedulerUpdateScheduleInput) (*SchedulerUpdateScheduleOutput, error)
	DeleteSchedule(ctx context.Context, params *SchedulerDeleteScheduleInput) (*SchedulerDeleteScheduleOutput, error)
	GetSchedule(ctx context.Context, params *SchedulerGetScheduleInput) (*SchedulerGetScheduleOutput, error)
}

// NewConflictException builds a generic conflict error so callers can
// deterministically simulate a duplicate schedule name without invoking AWS.
func NewConflictException(message string) error {
	return errors.New(message)
}

type SchedulerCreateScheduleInput = scheduler.CreateScheduleInput
type SchedulerCreateScheduleOutput = scheduler.CreateScheduleOutput
type SchedulerUpdateScheduleInput = scheduler.UpdateScheduleInput
type SchedulerUpdateScheduleOutput = scheduler.UpdateScheduleOutput
type SchedulerDeleteScheduleInput = scheduler.DeleteScheduleInput
type SchedulerDeleteScheduleOutput = scheduler.DeleteScheduleOutput
type SchedulerGetScheduleInput = scheduler.GetScheduleInput
type SchedulerGetScheduleOutput = scheduler.GetScheduleOutput
type SchedulerTarget = schedulertypes.Target
type SchedulerFlexibleTimeWindow = schedulertypes.FlexibleTimeWindow
type SchedulerActionAfterCompletion = schedulertypes.ActionAfterCompletion
type SchedulerRetryPolicy = schedulertypes.RetryPolicy
type SchedulerDeadLetterConfig = schedulertypes.DeadLetterConfig
type SchedulerScheduleState = schedulertypes.ScheduleState
type FlexibleTimeWindowMode = schedulertypes.FlexibleTimeWindowMode

const (
	SchedulerActionAfterCompletionDelete = schedulertypes.ActionAfterCompletionDelete
	FlexibleTimeWindowModeOff            = schedulertypes.FlexibleTimeWindowModeOff
	ScheduleStateEnabled                 = schedulertypes.ScheduleStateEnabled
)

type schedulerClient struct {
	client *scheduler.Client
}

func NewScheduler(client *scheduler.Client) SchedulerAPI {
	return &schedulerClient{client: client}
}

func (s *schedulerClient) CreateSchedule(ctx context.Context, params *SchedulerCreateScheduleInput) (*SchedulerCreateScheduleOutput, error) {
	return s.client.CreateSchedule(ctx, params)
}

func (s *schedulerClient) UpdateSchedule(ctx context.Context, params *SchedulerUpdateScheduleInput) (*SchedulerUpdateScheduleOutput, error) {
	return s.client.UpdateSchedule(ctx, params)
}

func (s *schedulerClient) DeleteSchedule(ctx context.Context, params *SchedulerDeleteScheduleInput) (*SchedulerDeleteScheduleOutput, error) {
	return s.client.DeleteSchedule(ctx, params)
}

func (s *schedulerClient) GetSchedule(ctx context.Context, params *SchedulerGetScheduleInput) (*SchedulerGetScheduleOutput, error) {
	return s.client.GetSchedule(ctx, params)
}
