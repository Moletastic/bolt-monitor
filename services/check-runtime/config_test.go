package main

import "testing"

func TestNewRuntimeConfig(t *testing.T) {
	config, err := newRuntimeConfig(func(key string) string {
		return map[string]string{"TABLE_NAME": "app-table", "EXECUTION_QUEUE_URL": "queue-url", "RUNTIME_MODE": "scheduler"}[key]
	})
	if err != nil {
		t.Fatalf("newRuntimeConfig: %v", err)
	}
	if config.Mode != modeScheduler || config.TableName != "app-table" || config.ExecutionQueueURL != "queue-url" {
		t.Fatalf("config = %+v", config)
	}
}

func TestNewRuntimeConfigRejectsInvalidRuntime(t *testing.T) {
	for _, runtimeMode := range []string{"", "unsupported"} {
		t.Run(runtimeMode, func(t *testing.T) {
			if _, err := newRuntimeConfig(func(key string) string {
				return map[string]string{"TABLE_NAME": "app-table", "RUNTIME_MODE": runtimeMode}[key]
			}); err == nil {
				t.Fatal("newRuntimeConfig error = nil")
			}
		})
	}
}

func TestNewRuntimeConfigRequiresSchedulerQueue(t *testing.T) {
	if _, err := newRuntimeConfig(func(key string) string {
		return map[string]string{"TABLE_NAME": "app-table", "RUNTIME_MODE": modeScheduler}[key]
	}); err == nil {
		t.Fatal("newRuntimeConfig error = nil")
	}
}
