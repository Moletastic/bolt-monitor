package auth

import (
	"strings"
	"time"

	sharederrors "bolt-monitor/shared/errors"
)

const (
	DefaultTenantID TenantID = "DEFAULT"

	membershipPrefix = "MEMBER#"
	MembershipSK     = "MEMBERSHIP"
)

type MembershipID string
type Subject string
type TenantID string
type MembershipStatus string
type Role string

const (
	MembershipStatusActive MembershipStatus = "ACTIVE"
	RoleAdmin              Role             = "ADMIN"
)

// Membership is the versioned AuthTable authority record for one Cognito subject.
type Membership struct {
	MembershipID   MembershipID
	Subject        Subject
	TenantID       TenantID
	Status         MembershipStatus
	Role           Role
	AuthValidAfter int64
	Version        int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Principal is trusted request authority after membership validation.
type Principal struct {
	Subject  Subject
	TenantID TenantID
	Role     Role
}

// AuthenticatedIdentity is the validated identity data needed before membership authorization.
// IssuedAt is diagnostic-only and must not be used as authorization authority.
type AuthenticatedIdentity struct {
	Subject  Subject
	AuthTime int64
	IssuedAt *int64
}

func MembershipPK(subject Subject) string {
	return membershipPrefix + string(subject)
}

func IsAuthorizedMembership(m Membership) bool {
	return ValidateMembership(m) == nil
}

func ValidateMembership(m Membership) error {
	if strings.TrimSpace(string(m.MembershipID)) == "" {
		return invalidMembership("membershipId")
	}
	if strings.TrimSpace(string(m.Subject)) == "" {
		return invalidMembership("subject")
	}
	if m.TenantID != DefaultTenantID {
		return invalidMembership("tenantId")
	}
	if m.Status != MembershipStatusActive {
		return invalidMembership("status")
	}
	if m.Role != RoleAdmin {
		return invalidMembership("role")
	}
	if !IsUnixSecond(m.AuthValidAfter) {
		return invalidMembership("authValidAfter")
	}
	if m.Version <= 0 {
		return invalidMembership("version")
	}
	if m.CreatedAt.IsZero() || m.UpdatedAt.IsZero() || m.UpdatedAt.Before(m.CreatedAt) {
		return invalidMembership("timestamps")
	}
	return nil
}

func IsUnixSecond(value int64) bool {
	return value >= 0
}

func IsAuthorizedAuthTime(authTime, authValidAfter int64) bool {
	return IsUnixSecond(authTime) && IsUnixSecond(authValidAfter) && authTime > authValidAfter
}

func invalidMembership(field string) error {
	return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{
		"field":  field,
		"reason": "invalid membership authority",
	})
}

type SecurityEvent string

const (
	EventBootstrapReconciled     SecurityEvent = "auth.bootstrap.reconciled"
	EventSignInSucceeded         SecurityEvent = "auth.sign_in.succeeded"
	EventSignInFailed            SecurityEvent = "auth.sign_in.failed"
	EventRecoveryRequested       SecurityEvent = "auth.recovery.requested"
	EventRecoveryCompleted       SecurityEvent = "auth.recovery.completed"
	EventTOTPEnrollmentSucceeded SecurityEvent = "auth.totp_enrollment.succeeded"
	EventTOTPEnrollmentFailed    SecurityEvent = "auth.totp_enrollment.failed"
	EventTOTPChallengeSucceeded  SecurityEvent = "auth.totp_challenge.succeeded"
	EventTOTPChallengeFailed     SecurityEvent = "auth.totp_challenge.failed"
	EventSessionCreated          SecurityEvent = "auth.session.created"
	EventSessionTerminated       SecurityEvent = "auth.session.terminated"
	EventRefreshFailed           SecurityEvent = "auth.refresh.failed"
	EventAuthorizationDenied     SecurityEvent = "auth.authorization.denied"
	EventMembershipStatusChanged SecurityEvent = "auth.membership.status_changed"
	EventAuthValidAfterAdvanced  SecurityEvent = "auth.membership.auth_valid_after_advanced"
)

func IsSecurityEvent(event SecurityEvent) bool {
	switch event {
	case EventBootstrapReconciled, EventSignInSucceeded, EventSignInFailed,
		EventRecoveryRequested, EventRecoveryCompleted, EventTOTPEnrollmentSucceeded,
		EventTOTPEnrollmentFailed, EventTOTPChallengeSucceeded, EventTOTPChallengeFailed, EventSessionCreated,
		EventSessionTerminated, EventRefreshFailed, EventAuthorizationDenied,
		EventMembershipStatusChanged, EventAuthValidAfterAdvanced:
		return true
	default:
		return false
	}
}
