package main

import "testing"

func TestNewEscalationRuntimeConfig(t *testing.T) {
	config, err := newEscalationRuntimeConfig(func(key string) string {
		return map[string]string{"TABLE_NAME": "app-table", "NOTIFICATION_QUEUE_URL": "queue-url", "SCHEDULE_GROUP_NAME": "group"}[key]
	})
	if err != nil {
		t.Fatalf("newEscalationRuntimeConfig: %v", err)
	}
	if config.TableName != "app-table" || config.NotificationQueueURL != "queue-url" || config.ScheduleGroupName != "group" {
		t.Fatalf("config = %+v", config)
	}
}

func TestNewEscalationRuntimeConfigRequiresTable(t *testing.T) {
	if _, err := newEscalationRuntimeConfig(func(string) string { return "" }); err == nil {
		t.Fatal("newEscalationRuntimeConfig error = nil")
	}
}
