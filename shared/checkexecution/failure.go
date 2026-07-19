package checkexecution

import "fmt"

type FailureCode string

const (
	FailureMalformedEnvelope FailureCode = "malformed_envelope"
	FailureIdentityConflict  FailureCode = "immutable_identity_conflict"
	FailureDuplicate         FailureCode = "duplicate"
	FailureSkip              FailureCode = "skipped"
	FailureLeaseLost         FailureCode = "lease_lost"
	FailureStaleObservation  FailureCode = "stale_observation"
	FailureStorage           FailureCode = "storage"
	FailureResultCommit      FailureCode = "result_commit"
	FailurePublication       FailureCode = "publication"
)

// RuntimeFailure distinguishes safe no-ops from retryable execution failures.
type RuntimeFailure struct {
	Code      FailureCode
	Retryable bool
	Operation string
	RunID     string
	Details   map[string]string
}

func (e *RuntimeFailure) Error() string {
	return fmt.Sprintf("execution %s failed: %s", e.Operation, e.Code)
}

func NewRuntimeFailure(code FailureCode, retryable bool, operation, runID string, details map[string]string) *RuntimeFailure {
	return &RuntimeFailure{Code: code, Retryable: retryable, Operation: operation, RunID: runID, Details: details}
}

func Duplicate(operation, runID string) *RuntimeFailure {
	return NewRuntimeFailure(FailureDuplicate, false, operation, runID, nil)
}

func Conflict(operation, runID string) *RuntimeFailure {
	return NewRuntimeFailure(FailureIdentityConflict, false, operation, runID, nil)
}

func Skip(operation, runID, reason string) *RuntimeFailure {
	return NewRuntimeFailure(FailureSkip, false, operation, runID, map[string]string{"reason": reason})
}

func LeaseLost(operation, runID string) *RuntimeFailure {
	return NewRuntimeFailure(FailureLeaseLost, false, operation, runID, nil)
}

func StaleObservation(operation, runID string) *RuntimeFailure {
	return NewRuntimeFailure(FailureStaleObservation, false, operation, runID, nil)
}

func Storage(operation, runID string) *RuntimeFailure {
	return NewRuntimeFailure(FailureStorage, true, operation, runID, nil)
}

func ResultCommit(operation, runID string) *RuntimeFailure {
	return NewRuntimeFailure(FailureResultCommit, true, operation, runID, nil)
}

func Publication(operation, runID string) *RuntimeFailure {
	return NewRuntimeFailure(FailurePublication, true, operation, runID, nil)
}
