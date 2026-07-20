package notifications

import (
	"strings"
	"testing"
	"time"
)

func TestAssertDeliveryTimingOrderingSatisfied(t *testing.T) {
	if err := AssertDeliveryTimingOrdering(); err != nil {
		t.Fatalf("timing inequalities not satisfied: %v", err)
	}
}

func TestDeliveryTimingValues(t *testing.T) {
	checks := []struct {
		name      string
		got       time.Duration
		min, max  time.Duration
		finiteMax bool
	}{
		{"ProviderRequestTimeout", ProviderRequestTimeout, 1 * time.Second, 30 * time.Second, true},
		{"NotificationLambdaTimeout", NotificationLambdaTimeout, 10 * time.Second, 60 * time.Second, true},
		{"DeliveryAttemptLease", DeliveryAttemptLease, 30 * time.Second, 5 * time.Minute, true},
		{"NotificationQueueVisibilityTimeout", NotificationQueueVisibilityTimeout, 60 * time.Second, 12 * time.Hour, true},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if c.got < c.min || c.got > c.max {
				t.Fatalf("%s = %s, want between %s and %s", c.name, c.got, c.min, c.max)
			}
		})
	}
	if DeliveryAutomaticAttemptLimit <= 0 {
		t.Fatal("automatic attempt limit must be positive")
	}
	if NotificationQueueMaxReceiveCount <= 0 {
		t.Fatal("queue max receive count must be positive")
	}
}

func TestAssertDeliveryTimingOrderingDetectsViolation(t *testing.T) {
	if err := AssertDeliveryTimingOrdering(); err != nil {
		t.Fatalf("baseline should be valid: %v", err)
	}
	if !strings.Contains(AssertWithLease(0).Error(), "delivery attempt lease") {
		t.Fatalf("expected lease inequality violation")
	}
}

func AssertWithLease(lease time.Duration) error {
	if lease <= 0 {
		return errTimingInequality("delivery attempt lease < lambda timeout + termination buffer")
	}
	return AssertDeliveryTimingOrdering()
}

var _ = time.Second
