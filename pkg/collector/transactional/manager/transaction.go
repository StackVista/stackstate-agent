package manager

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"time"
)

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
	NotifyChannel        chan interface{}
	LastUpdatedTimestamp time.Time
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
}

// EvictedTransaction is triggered once a stale transaction is evicted.
type EvictedTransaction struct {
	TransactionID string
}

// RollbackTransaction rolls back a transaction and marks a transaction as a failure.
type RollbackTransaction struct {
	TransactionID, Reason string
}

// Error returns a string representing the RollbackTransaction.
func (r RollbackTransaction) Error() string {
	return fmt.Sprintf("rolling back transaction %s. %s", r.TransactionID, r.Reason)
}

// StopTransactionManager triggers the shutdown of the transaction manager.
type StopTransactionManager struct{}

// TransactionManagerNotRunning is triggered when trying to create a transaction when the transaction manager has not
// been started yet.
type TransactionManagerNotRunning struct{}

// Error returns a string representation of the TransactionManagerNotRunning error and implements Error.
func (t TransactionManagerNotRunning) Error() string {
	return "transaction manager is not running, call TransactionManager.Start() to start it"
}

// TransactionNotFound is triggered when trying to look up a non-existing transaction in the transaction manager
type TransactionNotFound struct {
	TransactionID string
}

// Error returns a string representation of the TransactionNotFound error and implements Error.
func (t TransactionNotFound) Error() string {
	return fmt.Sprintf("transaction %s not found in transaction manager", t.TransactionID)
}

// ActionNotFound is triggered when trying to look up a non-existing action for a transaction in the transaction manager
type ActionNotFound struct {
	TransactionID, ActionID string
}

// Error returns a string representation of the ActionNotFound error and implements Error.
func (a ActionNotFound) Error() string {
	return fmt.Sprintf("action %s for transaction %s not found in transaction manager", a.ActionID, a.TransactionID)
}
