package manager

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"time"
)

// CommitAction ...
type CommitAction struct {
	TransactionID, ActionID string
}

// AckAction ...
type AckAction struct {
	TransactionID, ActionID string
}

// RejectAction ...
type RejectAction struct {
	TransactionID, ActionID, Reason string
}

// StartTransaction ...
type StartTransaction struct {
	CheckID       check.ID
	TransactionID string
	OnComplete    func(transaction *IntakeTransaction)
}

// CompleteTransaction ...
type CompleteTransaction struct {
	TransactionID string
}

// RollbackTransaction ...
type RollbackTransaction struct {
	TransactionID, Reason string
}

func (r RollbackTransaction) Error() string {
	return fmt.Sprintf("rolling back transaction %s. %s", r.TransactionID, r.Reason)
}

// StopTransactionManager ...
type StopTransactionManager struct{}

type TransactionManagerNotRunning struct{}

func (t TransactionManagerNotRunning) Error() string {
	return "transaction manager is not running, call TransactionManager.Start() to start it"
}

type TransactionNotFound struct {
	TransactionID string
}

func (t TransactionNotFound) Error() string {
	return fmt.Sprintf("transaction %s not found in transaction manager", t.TransactionID)
}

type ActionNotFound struct {
	TransactionID, ActionID string
}

func (a ActionNotFound) Error() string {
	return fmt.Sprintf("action %s for transaction %s not found in transaction manager", a.ActionID, a.TransactionID)
}

// TransactionManager ...
type TransactionManager struct {
	TransactionChannel          chan interface{}
	TransactionTicker           *time.Ticker
	Transactions                map[string]*IntakeTransaction
	transactionTimeoutDuration  time.Duration
	transactionEvictionDuration time.Duration
	running                     bool
}

// MakeTransactionManager ...
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

// Start ...
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

// Stop ...
func (txm *TransactionManager) Stop() {
	txm.running = false
	txm.TransactionChannel <- StopTransactionManager{}
	txm.TransactionTicker.Stop()
}

// startTransaction ...
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

// commitAction ...
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
	log.Debugf("Transaction %s, committing action %s", transactionID, actionID)

	return nil
}

func (txm *TransactionManager) updateTransaction(transaction *IntakeTransaction, action *Action, state TransactionState) {
	transaction.Actions[action.ActionID] = action
	transaction.State = state
	transaction.LastUpdatedTimestamp = time.Now()
}

// ackAction ...
func (txm *TransactionManager) ackAction(transactionID, actionID string) error {
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
	txm.updateTransaction(transaction, action, InProgress)
	log.Debugf("Transaction %s, acknowledged action %s", transactionID, actionID)

	return nil
}

// rejectAction ...
func (txm *TransactionManager) rejectAction(transactionID, actionID string) error {
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
	log.Debugf("Transaction %s, acknowledged action %s", transactionID, actionID)

	return nil
}

// completeTransaction marks the manager successful
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
	log.Debugf("Transaction succeeded %s", transaction.TransactionID)

	if transaction.OnComplete != nil {
		transaction.OnComplete(transaction)
	}

	return nil
}

// evictTransaction evicts the intake manager due to failure or timeout
func (txm *TransactionManager) evictTransaction(transactionID string) {
	delete(txm.Transactions, transactionID)
}

// rollbackTransaction rolls back the intake manager
func (txm *TransactionManager) rollbackTransaction(transactionID string) error {
	transaction, exists := txm.Transactions[transactionID]
	if !exists {
		return TransactionNotFound{TransactionID: transactionID}
	}

	// manager failed, rollback
	transaction.State = Failed
	transaction.LastUpdatedTimestamp = time.Now()
	log.Debugf("Transaction failed %s", transaction.TransactionID)

	if transaction.OnComplete != nil {
		transaction.OnComplete(transaction)
	}

	return nil
}
