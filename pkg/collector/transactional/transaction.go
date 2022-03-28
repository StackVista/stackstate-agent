package transactional

import "time"

// TransactionState is an integer representing the state of the transaction
type TransactionState int64

const (
	InProgress TransactionState = iota
	Failed
	Succeeded
	TimedOut
)

// Action represents a single operation in a transaction, which consists of one or more actions
type Action struct {
	ActionID              string
	Timestamp             time.Time
	Acknowledged          bool
	AcknowledgedTimestamp time.Time
}

// IntakeTransaction represents an intake transaction which consists of one or more actions
type IntakeTransaction struct {
	TransactionID string
	State         TransactionState
	Actions       []Action
}
