package manager

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"time"
)

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
	OnComplete    func(transaction *IntakeTransaction)
}

// CompleteTransaction completes a transaction. If all actions are acknowledges, the transaction is considered a success.
type CompleteTransaction struct {
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

// TransactionManager keeps track of all transactions for agent checks
type TransactionManager struct {
	TransactionChannel          chan interface{}
	TransactionTicker           *time.Ticker
	Transactions                map[string]*IntakeTransaction
	transactionTimeoutDuration  time.Duration
	transactionEvictionDuration time.Duration
	running                     bool
}

// MakeTransactionManager returns an instance of a TransactionManager
func MakeTransactionManager(transactionChannelBufferSize int, tickerInterval, transactionTimeoutDuration,
	transactionEvictionDuration time.Duration) *TransactionManager {
	return &TransactionManager{
		TransactionChannel:          make(chan interface{}, transactionChannelBufferSize),
		TransactionTicker:           time.NewTicker(tickerInterval),
		Transactions:                make(map[string]*IntakeTransaction),
		transactionTimeoutDuration:  transactionTimeoutDuration,
		transactionEvictionDuration: transactionEvictionDuration,
	}
}

// Start sets up the transaction manager to consume messages on the txm.TransactionChannel. It consumes one message at
// a time using the `select` statement and populates / evicts transactions in the transaction manager.
func (txm *TransactionManager) Start() {
	go func() {
	transactionHandler:
		for {
			select {
			case input := <-txm.TransactionChannel:
				switch msg := input.(type) {
				// transaction operations
				case StartTransaction:
					log.Debugf("Creating new transaction %s for check %s", msg.TransactionID, msg.CheckID)
					if _, err := txm.startTransaction(msg.TransactionID, msg.OnComplete); err != nil {
						txm.TransactionChannel <- err
					}
				case CommitAction:
					log.Debugf("Committing action %s for transaction %s", msg.ActionID, msg.TransactionID)
					if err := txm.commitAction(msg.TransactionID, msg.ActionID); err != nil {
						txm.TransactionChannel <- err
					}
				case AckAction:
					log.Debugf("Acknowledging action %s for transaction %s", msg.ActionID, msg.TransactionID)
					if err := txm.ackAction(msg.TransactionID, msg.ActionID); err != nil {
						txm.TransactionChannel <- err
					}
				case RejectAction:
					_ = log.Errorf("Rejecting action %s for transaction %s: %s", msg.ActionID, msg.TransactionID, msg.Reason)
					if err := txm.rejectAction(msg.TransactionID, msg.ActionID); err != nil {
						txm.TransactionChannel <- err
					} else {
						// rollback the transaction
						reason := fmt.Sprintf("rejected action %s for transaction %s: %s", msg.ActionID, msg.TransactionID, msg.Reason)
						txm.TransactionChannel <- RollbackTransaction{TransactionID: msg.TransactionID, Reason: reason}
					}
				case CompleteTransaction:
					log.Debugf("Completing transaction %s", msg.TransactionID)
					if err := txm.completeTransaction(msg.TransactionID); err != nil {
						txm.TransactionChannel <- err
					}
				// error cases
				case RollbackTransaction:
					_ = log.Errorf(msg.Error())
					if err := txm.rollbackTransaction(msg.TransactionID); err != nil {
						txm.TransactionChannel <- err
					}
				case TransactionNotFound:
					_ = log.Errorf(msg.Error())
				case ActionNotFound:
					_ = log.Errorf(msg.Error())
				// shutdown transaction manager
				case StopTransactionManager:
					break transactionHandler
				default:
					_ = log.Errorf("Got unexpected msg %v", msg)
				}
			case <-txm.TransactionTicker.C:
				// expire stale transactions, clean up expired transactions that exceed the eviction duration
				for _, transaction := range txm.Transactions {
					if transaction.State != Stale && transaction.LastUpdatedTimestamp.Before(time.Now().Add(-txm.transactionTimeoutDuration)) {
						// last updated timestamp is before current time - manager timeout duration => Tx is stale
						transaction.State = Stale
					} else if transaction.State == Stale && transaction.LastUpdatedTimestamp.Before(time.Now().Add(-txm.transactionEvictionDuration)) {
						// last updated timestamp is before current time - manager eviction duration => Tx can be evicted
						txm.evictTransaction(transaction.TransactionID)
					} else if transaction.State == Failed || transaction.State == Succeeded {
						log.Debugf("Cleaning up %s transaction: %s", transaction.State.String(), transaction.TransactionID)
						txm.evictTransaction(transaction.TransactionID)
					}
				}

				// TODO: produce some transaction manager metrics
			default:
			}
		}
	}()

	txm.running = true
}

// Stop shuts down the transaction manager and stops the transactionHandler receiver loop
func (txm *TransactionManager) Stop() {
	txm.running = false
	txm.TransactionChannel <- StopTransactionManager{}
	txm.TransactionTicker.Stop()
}

// startTransaction creates a transaction and puts it into the transactions map
func (txm *TransactionManager) startTransaction(transactionID string, onComplete func(transaction *IntakeTransaction)) (*IntakeTransaction, error) {
	if !txm.running {
		return nil, TransactionManagerNotRunning{}
	}

	transaction := &IntakeTransaction{
		TransactionID:        transactionID,
		State:                InProgress,
		Actions:              map[string]*Action{},
		LastUpdatedTimestamp: time.Now(),
		OnComplete:           onComplete,
	}

	txm.Transactions[transaction.TransactionID] = transaction

	return transaction, nil
}

// commitAction commits / promises an action for a certain transaction. A commit is only a promise that something needs
// to be fulfilled. An unacknowledged action results in a transaction failure.
func (txm *TransactionManager) commitAction(transactionID, actionID string) error {
	transaction, exists := txm.Transactions[transactionID]
	if !exists {
		return TransactionNotFound{TransactionID: transactionID}
	}

	action := &Action{
		ActionID:           actionID,
		CommittedTimestamp: time.Now(),
	}
	txm.updateTransaction(transaction, action, InProgress)

	return nil
}

// updateTransaction is a helper function to set the state of a transaction as well as update it's LastUpdatedTimestamp.
func (txm *TransactionManager) updateTransaction(transaction *IntakeTransaction, action *Action, state TransactionState) {
	transaction.Actions[action.ActionID] = action
	transaction.State = state
	transaction.LastUpdatedTimestamp = time.Now()
}

// ackAction acknowledges an action for a given transaction. This marks the action as acknowledged.
func (txm *TransactionManager) ackAction(transactionID, actionID string) error {
	return txm.findAndUpdateAction(transactionID, actionID, true)
}

// rejectAction acknowledges an action for a given transaction. This marks the action as acknowledged and results in a
// failed transaction and rollback.
func (txm *TransactionManager) rejectAction(transactionID, actionID string) error {
	return txm.findAndUpdateAction(transactionID, actionID, false)
}

// findAndUpdateAction is a helper function to find a transaction and action for the given ID's, marks the action as
// acknowledged and updates the transaction is updateTransaction is set to true.
func (txm *TransactionManager) findAndUpdateAction(transactionID, actionID string, updateTransaction bool) error {
	transaction, exists := txm.Transactions[transactionID]
	if !exists {
		return TransactionNotFound{TransactionID: transactionID}
	}

	action, exists := transaction.Actions[actionID]
	if !exists {
		return ActionNotFound{ActionID: actionID, TransactionID: transactionID}
	}
	action.Acknowledged = true
	action.AcknowledgedTimestamp = time.Now()

	if updateTransaction {
		txm.updateTransaction(transaction, action, InProgress)
	}

	return nil
}

// completeTransaction marks a transaction for a given transactionID as Succeeded, if all the committed actions
// of a transaction has been acknowledged
func (txm *TransactionManager) completeTransaction(transactionID string) error {
	transaction, exists := txm.Transactions[transactionID]
	if !exists {
		return TransactionNotFound{TransactionID: transactionID}
	}

	// ensure all actions have been acknowledged
	for _, action := range transaction.Actions {
		if !action.Acknowledged {
			reason := fmt.Sprintf("Not all actions have been acknowledged, rolling back manager: %s", transaction.TransactionID)
			return RollbackTransaction{TransactionID: transactionID, Reason: reason}
		}
	}
	transaction.State = Succeeded
	transaction.LastUpdatedTimestamp = time.Now()

	if transaction.OnComplete != nil {
		transaction.OnComplete(transaction)
	}

	return nil
}

// evictTransaction delete a given transactionID from the transactions map
func (txm *TransactionManager) evictTransaction(transactionID string) {
	delete(txm.Transactions, transactionID)
}

// rollbackTransaction rolls back the transaction in the event of a failure
func (txm *TransactionManager) rollbackTransaction(transactionID string) error {
	transaction, exists := txm.Transactions[transactionID]
	if !exists {
		return TransactionNotFound{TransactionID: transactionID}
	}

	transaction.State = Failed
	transaction.LastUpdatedTimestamp = time.Now()

	if transaction.OnComplete != nil {
		transaction.OnComplete(transaction)
	}

	return nil
}
