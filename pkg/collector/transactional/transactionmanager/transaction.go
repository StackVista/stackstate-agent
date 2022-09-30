package transactionmanager

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"time"
)

// TransactionStatus is an integer representing the state of the transaction
type TransactionStatus int64

const (
	// InProgress is used to represent a InProgress transaction
	InProgress TransactionStatus = iota
	// Failed is used to represent a Failed transaction
	Failed
	// Succeeded is used to represent a Succeeded transaction
	Succeeded
	// Stale is used to represent a Stale transaction
	Stale
)

// String returns a string representation of TransactionStatus
func (state TransactionStatus) String() string {
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

// ActionStatus is an integer representing the state of the action
type ActionStatus int64

const (
	// Committed is used to represent a Committed action
	Committed ActionStatus = iota
	// Acknowledged is used to represent an Acknowledged action. ie successfully "completed" action
	Acknowledged
	// Rejected is used to represent an Rejected action. ie unsuccessfully "completed" action
	Rejected
)

// String returns a string representation of ActionStatus
func (state ActionStatus) String() string {
	switch state {
	case Acknowledged:
		return "acknowledged"
	case Rejected:
		return "rejected"
	default:
		return "committed"
	}
}

// Action represents a single operation in a checkmanager, which consists of one or more actions
type Action struct {
	ActionID               string
	CommittedTimestamp     time.Time
	Status                 ActionStatus
	StatusUpdatedTimestamp time.Time
}

// IntakeTransaction represents an intake checkmanager which consists of one or more actions
type IntakeTransaction struct {
	TransactionID        string
	CheckID              check.ID
	Status               TransactionStatus
	Actions              map[string]*Action // pointer to allow in-place mutation instead of setting the value again
	NotifyChannel        chan interface{}
	LastUpdatedTimestamp time.Time
	State                *TransactionState // the State of the TransactionState will be updated each time, no need for a pointer
}

// TransactionState keeps the state for a given key
type TransactionState struct {
	Key, State string
}

// SetTransactionState is used to set transaction state for a given transactionID and Key.
type SetTransactionState struct {
	TransactionID, Key, State string
}

// CommitAction is used to commit an action for a certain transaction.
type CommitAction struct {
	TransactionID, ActionID string
}

// AckAction acknowledges an action for a given transaction.
type AckAction struct {
	TransactionID, ActionID string
}

// RejectAction rejects an action for a given transaction. This results in a failed transaction.
type RejectAction struct {
	TransactionID, ActionID, Reason string
}

// StartTransaction starts a transaction for a given checkID, with an optional OnComplete callback function.
type StartTransaction struct {
	CheckID       check.ID
	TransactionID string
	NotifyChannel chan interface{}
}

// CompleteTransaction completes a transaction. If all actions are acknowledges, the transaction is considered a success.
type CompleteTransaction struct {
	TransactionID string
	State         *TransactionState
}

// EvictedTransaction is triggered once a stale transaction is evicted.
type EvictedTransaction struct {
	TransactionID string
}

// DiscardTransaction rolls back a transaction and marks a transaction as a failure.
type DiscardTransaction struct {
	TransactionID, Reason string
}

// Error returns a string representing the DiscardTransaction.
func (r DiscardTransaction) Error() string {
	return fmt.Sprintf("discarding transaction %s. %s", r.TransactionID, r.Reason)
}

// StopTransactionManager triggers the shutdown of the transaction checkmanager.
type StopTransactionManager struct{}

// TransactionNotFound is triggered when trying to look up a non-existing transaction in the transaction checkmanager
type TransactionNotFound struct {
	TransactionID string
}

// Error returns a string representation of the TransactionNotFound error and implements Error.
func (t TransactionNotFound) Error() string {
	return fmt.Sprintf("transaction %s not found in transaction checkmanager", t.TransactionID)
}

// TransactionCompleted is triggered when trying to look up a transaction that is already in a failed / succeeded state.
type TransactionCompleted struct {
	TransactionID string
}

// Error returns a string representation of the TransactionCompleted error and implements Error.
func (t TransactionCompleted) Error() string {
	return fmt.Sprintf("transaction %s has already been completed", t.TransactionID)
}

// ActionNotFound is triggered when trying to look up a non-existing action for a transaction in the transaction checkmanager
type ActionNotFound struct {
	TransactionID, ActionID string
}

// Error returns a string representation of the ActionNotFound error and implements Error.
func (a ActionNotFound) Error() string {
	return fmt.Sprintf("action %s for transaction %s not found in transaction checkmanager", a.ActionID, a.TransactionID)
}
