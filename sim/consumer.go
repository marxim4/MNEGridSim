package sim

// Consumer defines a consumer's identity and allocation priority.
type Consumer struct {
	ID       string
	Priority int // Lower values have precedence.
}

// ConsumerDemand reports a consumer's requested power in kW for one tick.
type ConsumerDemand struct {
	ConsumerID  string
	RequestedKW int64
}
