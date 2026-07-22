package main

import (
	"fmt"
	"strings"
)

type escalationRuntimeConfig struct {
	TableName             string
	NotificationQueueURL  string
	ScheduleGroupName     string
	ScheduleExecutionRole string
	NotificationQueueARN  string
	NotificationDLQARN    string
}

func newEscalationRuntimeConfig(getenv func(string) string) (escalationRuntimeConfig, error) {
	config := escalationRuntimeConfig{
		TableName:             strings.TrimSpace(getenv("TABLE_NAME")),
		NotificationQueueURL:  strings.TrimSpace(getenv("NOTIFICATION_QUEUE_URL")),
		ScheduleGroupName:     strings.TrimSpace(getenv("SCHEDULE_GROUP_NAME")),
		ScheduleExecutionRole: strings.TrimSpace(getenv("SCHEDULE_EXECUTION_ROLE_ARN")),
		NotificationQueueARN:  strings.TrimSpace(getenv("NOTIFICATION_QUEUE_ARN")),
		NotificationDLQARN:    strings.TrimSpace(getenv("NOTIFICATION_DLQ_ARN")),
	}
	if config.TableName == "" {
		return escalationRuntimeConfig{}, fmt.Errorf("TABLE_NAME is required")
	}
	return config, nil
}
