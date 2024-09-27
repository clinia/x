package pubsubx

type SubscriberError struct {
	Message   string
	Retryable bool
}

var _ error = (*SubscriberError)(nil)

func (e *SubscriberError) Error() string {
	return e.Message
}

// This error should be returned when the handler wishes to abort the subscription.
var abortSubscribeError = &SubscriberError{Message: "abort subscription", Retryable: false}

func AbortSubscribeError() *SubscriberError {
	return abortSubscribeError
}
