package main

import (
	"fmt"
	"strings"
	"time"
)

type runtimeConfig struct {
	TableName          string
	ExecutionQueueURL  string
	EscalationQueueURL string
	Mode               string
	WorkLeaseDuration  time.Duration
}

func newRuntimeConfig(getenv func(string) string) (runtimeConfig, error) {
	config := runtimeConfig{
		TableName:          strings.TrimSpace(getenv("TABLE_NAME")),
		ExecutionQueueURL:  strings.TrimSpace(getenv("EXECUTION_QUEUE_URL")),
		EscalationQueueURL: strings.TrimSpace(getenv("ESCALATION_QUEUE_URL")),
		Mode:               strings.ToLower(strings.TrimSpace(getenv("RUNTIME_MODE"))),
		WorkLeaseDuration:  readDurationSeconds(getenv, "WORK_LEASE_DURATION_SECONDS", defaultExecutionWorkLeaseDuration),
	}
	if config.TableName == "" {
		return runtimeConfig{}, fmt.Errorf("TABLE_NAME is required")
	}
	if config.Mode != modeScheduler && config.Mode != modeWorker {
		return runtimeConfig{}, fmt.Errorf("unsupported runtime mode %q", config.Mode)
	}
	if config.Mode == modeScheduler && config.ExecutionQueueURL == "" {
		return runtimeConfig{}, fmt.Errorf("EXECUTION_QUEUE_URL is required for scheduler mode")
	}
	return config, nil
}
