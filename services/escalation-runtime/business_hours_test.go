package main

import (
	"testing"
	"time"

	"bolt-monitor/shared/escalation"
)

func TestIsWithinBusinessHours(t *testing.T) {
	config := &escalation.BusinessHoursConfig{Timezone: "America/New_York", StartHour: 9, EndHour: 17, DaysOfWeek: []int{1, 2, 3, 4, 5}}
	inside := time.Date(2026, 6, 15, 15, 0, 0, 0, time.UTC)
	outside := time.Date(2026, 6, 15, 1, 0, 0, 0, time.UTC)
	if !IsWithinBusinessHours(config, inside) {
		t.Fatal("expected inside time to be within business hours")
	}
	if IsWithinBusinessHours(config, outside) {
		t.Fatal("expected outside time to be outside business hours")
	}
}

func TestIsWithinBusinessHoursHandlesOvernightWindow(t *testing.T) {
	config := &escalation.BusinessHoursConfig{Timezone: "UTC", StartHour: 22, EndHour: 6, DaysOfWeek: []int{1, 2, 3, 4, 5, 6, 7}}
	inside := time.Date(2026, 6, 16, 1, 0, 0, 0, time.UTC)
	if !IsWithinBusinessHours(config, inside) {
		t.Fatal("expected overnight window to include 01:00")
	}
}
