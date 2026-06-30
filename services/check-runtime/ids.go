package main

import (
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
)

func newRunID(time.Time) string {
	return fmt.Sprintf("RUN_%s", ulid.Make().String())
}

func newIncidentID(time.Time) string {
	return fmt.Sprintf("INC_%s", ulid.Make().String())
}

func newActivityID(time.Time) string {
	return fmt.Sprintf("ACT_%s", ulid.Make().String())
}

func newAuditID(time.Time) string {
	return fmt.Sprintf("AUD_%s", ulid.Make().String())
}
