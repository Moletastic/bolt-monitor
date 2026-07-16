package main

import (
	"log"
	"os"
	"time"

	"bolt-monitor/shared/auth"
)

const unknownSecurityEventStage = "unknown"

type securityEvent struct {
	Timestamp     string             `json:"timestamp"`
	Event         auth.SecurityEvent `json:"event"`
	Outcome       string             `json:"outcome"`
	Stage         string             `json:"stage"`
	Component     string             `json:"component"`
	Subject       string             `json:"subject,omitempty"`
	CorrelationID string             `json:"correlationId,omitempty"`
}

type securityEventEmitter func(securityEvent)

// emitMonitorSecurityEvent writes only the fixed security-event schema, never request or error data.
func emitMonitorSecurityEvent(event securityEvent) {
	log.Printf(`{"timestamp":%q,"event":%q,"outcome":%q,"stage":%q,"component":%q,"subject":%q,"correlationId":%q}`,
		event.Timestamp, event.Event, event.Outcome, event.Stage, event.Component, event.Subject, event.CorrelationID)
}

func newMonitorSecurityEvent(event auth.SecurityEvent, outcome string, subject auth.Subject, correlationID string) securityEvent {
	stage := os.Getenv("SST_STAGE")
	if stage == "" {
		stage = unknownSecurityEventStage
	}
	return securityEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339), Event: event, Outcome: outcome,
		Stage: stage, Component: "monitor-api", Subject: string(subject), CorrelationID: correlationID,
	}
}
