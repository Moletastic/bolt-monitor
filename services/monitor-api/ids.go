package main

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

// commandClock and identifierGenerator are explicit command collaborators.
// Composition roots choose production implementations; tests use fixed values.
type commandClock func() time.Time

type identifierGenerator struct {
	newServiceID             func(time.Time) string
	newMonitorID             func(string, string, string) string
	newRunID                 func(time.Time) string
	newEscalationPolicyID    func(time.Time) string
	newNotificationChannelID func(time.Time) string
}

func productionIdentifierGenerator() identifierGenerator {
	return identifierGenerator{
		newServiceID:             newServiceID,
		newMonitorID:             generateMonitorID,
		newRunID:                 newRunID,
		newEscalationPolicyID:    newEscalationPolicyID,
		newNotificationChannelID: newNotificationChannelID,
	}
}

func newAuditID(now time.Time) string {
	return fmt.Sprintf("AUD_%s", ulid.Make().String())
}

func newRunID(now time.Time) string {
	return fmt.Sprintf("RUN_%s", ulid.Make().String())
}

func newActivityID(now time.Time) string {
	return fmt.Sprintf("ACT_%s", ulid.Make().String())
}

func newServiceID(now time.Time) string {
	return fmt.Sprintf("SVC_%s", ulid.Make().String())
}

func newEscalationPolicyID(now time.Time) string {
	return fmt.Sprintf("POL_%s", ulid.Make().String())
}

func newNotificationChannelID(now time.Time) string {
	return fmt.Sprintf("CH_%s", ulid.Make().String())
}

func newIncidentID(now time.Time) string {
	return fmt.Sprintf("INC_%s", ulid.Make().String())
}

func generateMonitorID(monitorType, targetURL, name string) string {
	base := monitorType
	if targetURL != "" {
		parsed, err := url.Parse(targetURL)
		if err == nil && parsed.Host != "" {
			host := strings.ToLower(strings.ReplaceAll(parsed.Host, ".", "-"))
			if parsed.Path != "" && parsed.Path != "/" {
				pathSeg := strings.TrimPrefix(parsed.Path, "/")
				pathSeg = strings.ReplaceAll(pathSeg, "/", "-")
				if len(pathSeg) > 20 {
					pathSeg = pathSeg[:20]
				}
				base = host + "-" + pathSeg
			} else {
				base = host
			}
		}
	} else if name != "" {
		base = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	}
	hash := ulid.Make().String()[:6]
	return base + "-" + hash
}
