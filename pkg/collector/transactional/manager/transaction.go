package manager

import "time"

// TransactionState is an integer representing the state of the transaction
type TransactionState int64

const (
	InProgress TransactionState = iota
	Failed
	Succeeded
	Stale
)

// String returns a string representation of TransactionState
func (state TransactionState) String() string {
	switch state {
	case Failed:
		return "failed"
	case Succeeded:
		return "succeeded"
	case Stale:
		return "stale"
	default:
		return "in progress"
	}
}

// Action represents a single operation in a manager, which consists of one or more actions
type Action struct {
	ActionID              string
	CommittedTimestamp    time.Time
	Acknowledged          bool
	AcknowledgedTimestamp time.Time
}

// IntakeTransaction represents an intake manager which consists of one or more actions
type IntakeTransaction struct {
	TransactionID        string
	State                TransactionState
	Actions              map[string]*Action
	OnComplete           func(transaction *IntakeTransaction)
	LastUpdatedTimestamp time.Time
}
