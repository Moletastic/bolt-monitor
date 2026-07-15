package auth

import (
	"testing"
	"time"
)

func TestValidateMembershipAcceptsOnlyV1Authority(t *testing.T) {
	now := time.Date(2026, time.July, 15, 0, 0, 0, 0, time.UTC)
	membership := Membership{
		MembershipID:   "member-1",
		Subject:        "subject-1",
		TenantID:       DefaultTenantID,
		Status:         MembershipStatusActive,
		Role:           RoleAdmin,
		AuthValidAfter: 10,
		Version:        1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := ValidateMembership(membership); err != nil {
		t.Fatalf("ValidateMembership() error = %v", err)
	}
	if got, want := MembershipPK(membership.Subject), "MEMBER#subject-1"; got != want {
		t.Fatalf("MembershipPK() = %q, want %q", got, want)
	}

	for _, mutate := range []func(*Membership){
		func(m *Membership) { m.TenantID = "OTHER" },
		func(m *Membership) { m.Status = "DISABLED" },
		func(m *Membership) { m.Role = "VIEWER" },
		func(m *Membership) { m.AuthValidAfter = -1 },
		func(m *Membership) { m.Version = 0 },
	} {
		invalid := membership
		mutate(&invalid)
		if err := ValidateMembership(invalid); err == nil {
			t.Fatal("ValidateMembership() accepted invalid authority")
		}
	}
}

func TestAuthTimeMustBeAfterAuthorityBoundary(t *testing.T) {
	if IsAuthorizedAuthTime(10, 10) {
		t.Fatal("auth_time equal to AuthValidAfter authorized")
	}
	if IsAuthorizedAuthTime(9, 10) {
		t.Fatal("auth_time before AuthValidAfter authorized")
	}
	if !IsAuthorizedAuthTime(11, 10) {
		t.Fatal("auth_time after AuthValidAfter denied")
	}
}

func TestSecurityEventsUseKnownSecretSafeNames(t *testing.T) {
	if !IsSecurityEvent(EventAuthorizationDenied) {
		t.Fatal("known event rejected")
	}
	if IsSecurityEvent("auth.token.eyJhbGciOiJIUzI1NiJ9") {
		t.Fatal("unknown event accepted")
	}
}
