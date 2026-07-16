package main

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"time"

	"bolt-monitor/shared/auth"
)

const unknownSecurityEventStage = "unknown"

var correlationIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type securityEvent struct {
	Timestamp     string             `json:"timestamp"`
	Event         auth.SecurityEvent `json:"event"`
	Outcome       string             `json:"outcome"`
	Stage         string             `json:"stage"`
	Component     string             `json:"component"`
	Subject       string             `json:"subject,omitempty"`
	CorrelationID string             `json:"correlationId,omitempty"`
	Operation     string             `json:"operation,omitempty"`
	Events        int                `json:"AuthenticationEvents,omitempty"`
	EMF           *embeddedMetric    `json:"_aws,omitempty"`
}

type embeddedMetric struct {
	Timestamp         int64                 `json:"Timestamp"`
	CloudWatchMetrics []cloudWatchMetricSet `json:"CloudWatchMetrics"`
}

type cloudWatchMetricSet struct {
	Namespace  string     `json:"Namespace"`
	Dimensions [][]string `json:"Dimensions"`
	Metrics    []metric   `json:"Metrics"`
}

type metric struct {
	Name string `json:"Name"`
	Unit string `json:"Unit"`
}

type securityEventEmitter func(securityEvent)

// emitMonitorSecurityEvent writes only the fixed security-event schema, never request or error data.
func emitMonitorSecurityEvent(event securityEvent) {
	encoded, err := json.Marshal(event)
	if err != nil {
		log.Print(`{"event":"auth.audit.failed"}`)
		return
	}
	log.Print(string(encoded))
}

func newMonitorSecurityEvent(event auth.SecurityEvent, outcome string, subject auth.Subject, correlationID string) securityEvent {
	stage := os.Getenv("SST_STAGE")
	if stage == "" {
		stage = unknownSecurityEventStage
	}
	result := securityEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339), Event: event, Outcome: outcome,
		Stage: stage, Component: "monitor-api", Subject: string(subject), CorrelationID: correlationID,
	}
	if event == auth.EventAuthorizationDenied {
		result.Operation = "authorization"
		result.Events = 1
		result.EMF = authenticationEMF()
	}
	return result
}

func authenticationEMF() *embeddedMetric {
	return &embeddedMetric{
		Timestamp: time.Now().UTC().UnixMilli(),
		CloudWatchMetrics: []cloudWatchMetricSet{{
			Namespace: "BoltMonitor/Auth", Dimensions: [][]string{{"stage", "component", "operation", "outcome"}},
			Metrics: []metric{{Name: "AuthenticationEvents", Unit: "Count"}},
		}},
	}
}

func monitorCorrelationID(requestID, propagated string) string {
	if correlationIDPattern.MatchString(propagated) {
		return propagated
	}
	return requestID
}
