package notifications

import "time"

const (
	// ProviderRequestTimeout caps a single provider HTTP attempt.
	ProviderRequestTimeout = 8 * time.Second
	// ProviderCompletionBuffer covers conditional outcome persistence after the provider response lands.
	ProviderCompletionBuffer = 1 * time.Second
	// NotificationLambdaTimeout is the Lambda budget for one delivery invocation.
	NotificationLambdaTimeout = 30 * time.Second
	// LambdaTerminationBuffer is the slack between Lambda timeout and delivery lease expiry.
	LambdaTerminationBuffer = 5 * time.Second
	// DeliveryAttemptLease is the fenced claim window for one provider attempt.
	DeliveryAttemptLease = 60 * time.Second
	// ClaimStartBudget bounds the work between receipt and the conditional claim.
	ClaimStartBudget = 5 * time.Second
	// RedeliveryBuffer is the slack between lease expiry and SQS redelivery visibility.
	RedeliveryBuffer = 5 * time.Second
	// NotificationQueueVisibilityTimeout is the SQS visibility window for one delivery message.
	NotificationQueueVisibilityTimeout = 120 * time.Second
	// DeliveryAutomaticAttemptLimit is the maximum automatic retry attempts per delivery.
	DeliveryAutomaticAttemptLimit = 5
	// NotificationQueueMaxReceiveCount is the SQS redrive threshold.
	NotificationQueueMaxReceiveCount = 8
	// NotificationRetryBackoffMax bounds provider-notified retry-after values.
	NotificationRetryBackoffMax = 30 * time.Second
	// SchedulerTargetRetryAge bounds the EventBridge Scheduler retry window.
	SchedulerTargetRetryAge = 60 * time.Second
	// SchedulerTargetRetryBackoffMax bounds the Scheduler retry backoff.
	SchedulerTargetRetryBackoffMax = 30 * time.Second
)

// AssertDeliveryTimingOrdering pins the inequalities required by the
// notification-delivery assurance spec.
func AssertDeliveryTimingOrdering() error {
	if ProviderRequestTimeout+ProviderCompletionBuffer >= NotificationLambdaTimeout {
		return errTimingInequality("provider timeout + completion buffer < lambda timeout")
	}
	if NotificationLambdaTimeout+LambdaTerminationBuffer >= DeliveryAttemptLease {
		return errTimingInequality("lambda timeout + termination buffer < delivery attempt lease")
	}
	if ClaimStartBudget+DeliveryAttemptLease+RedeliveryBuffer > NotificationQueueVisibilityTimeout {
		return errTimingInequality("claim budget + lease + redelivery buffer <= queue visibility timeout")
	}
	if DeliveryAutomaticAttemptLimit > NotificationQueueMaxReceiveCount {
		return errTimingInequality("automatic attempt limit <= queue max receive count")
	}
	if NotificationRetryBackoffMax >= DeliveryAttemptLease {
		return errTimingInequality("retry backoff max < delivery attempt lease")
	}
	if SchedulerTargetRetryBackoffMax > SchedulerTargetRetryAge {
		return errTimingInequality("scheduler target backoff <= scheduler target retry age")
	}
	return nil
}

type timingError struct{ msg string }

func (e *timingError) Error() string { return "delivery timing: " + e.msg }

func errTimingInequality(msg string) error { return &timingError{msg: msg} }
