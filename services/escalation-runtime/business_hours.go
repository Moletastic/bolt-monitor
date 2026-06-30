package main

import (
	"time"

	"bolt-monitor/shared/escalation"
)

func IsWithinBusinessHours(config *escalation.BusinessHoursConfig, now time.Time) bool {
	if config == nil || config.Timezone == "" || len(config.DaysOfWeek) == 0 {
		return false
	}
	location, err := time.LoadLocation(config.Timezone)
	if err != nil {
		return false
	}
	local := now.In(location)
	weekday := int(local.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	allowedDay := false
	for _, day := range config.DaysOfWeek {
		if day == weekday {
			allowedDay = true
			break
		}
	}
	if !allowedDay {
		return false
	}
	hour := local.Hour()
	if config.StartHour == config.EndHour {
		return false
	}
	if config.StartHour < config.EndHour {
		return hour >= config.StartHour && hour < config.EndHour
	}
	return hour >= config.StartHour || hour < config.EndHour
}
