package manager

import (
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
	TransactionID, ActionID string
	Reason                  error
}

// StartTransaction ...
type StartTransaction struct {
	CheckID       check.ID
	TransactionID string
	OnComplete    func(transaction *IntakeTransaction)
}

// RollbackTransaction ...
type RollbackTransaction struct {
	TransactionID string
	Reason        error
}

// StopTransactionManager ...
type StopTransactionManager struct{}

// TransactionManager ...
type TransactionManager struct {
	TransactionChannel          chan interface{}
	TransactionTicker           *time.Ticker
	Transactions                map[string]*IntakeTransaction
	transactionTimeoutDuration  time.Duration
	transactionEvictionDuration time.Duration
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
				case StartTransaction:
					log.Debugf("Creating new transaction %s for check %s", msg.TransactionID, msg.CheckID)
					txm.startTransaction(msg.TransactionID, msg.OnComplete)
				case CommitAction:
					log.Debugf("Committing msg %s for manager %s", msg.ActionID, msg.TransactionID)
					txm.commitAction(msg.TransactionID, msg.ActionID)
				case AckAction:
					log.Debugf("Acknowledging msg %s for manager %s", msg.ActionID, msg.TransactionID)
					txm.ackAction(msg.TransactionID, msg.ActionID)
				case RejectAction:
					_ = log.Errorf("Rejecting msg %s for manager %s: %s", msg.ActionID, msg.TransactionID, msg.Reason)
					txm.rejectAction(msg.TransactionID, msg.ActionID)
				case RollbackTransaction:
					_ = log.Errorf("Rolling back manager %s: %s", msg.TransactionID, msg.Reason)
					txm.rollbackTransaction(msg.TransactionID)
				case StopTransactionManager:
					break transactionHandler
				default:
					_ = log.Errorf("Got unexpected msg %v", msg)
				}
			case <-txm.TransactionTicker.C:
				// expire stale transactions, clean up expired transactions that exceed the eviction duration
				for _, transaction := range txm.Transactions {
					if transaction.LastUpdatedTimestamp.Before(time.Now().Add(-txm.transactionTimeoutDuration)) {
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
			default:
			}
		}
	}()
}

// Stop ...
func (txm *TransactionManager) Stop() {
	txm.TransactionChannel <- StopTransactionManager{}
	txm.TransactionTicker.Stop()
}

// startTransaction ...
func (txm *TransactionManager) startTransaction(transactionID string, onComplete func(transaction *IntakeTransaction)) *IntakeTransaction {
	transaction := &IntakeTransaction{
		TransactionID:        transactionID,
		State:                InProgress,
		Actions:              map[string]*Action{},
		LastUpdatedTimestamp: time.Now(),
		OnComplete:           onComplete,
	}

	txm.Transactions[transaction.TransactionID] = transaction

	return transaction
}

// commitAction ...
func (txm *TransactionManager) commitAction(transactionID, actionID string) {
	transaction, exists := txm.Transactions[transactionID]
	if exists {
		action := &Action{
			ActionID:           actionID,
			CommittedTimestamp: time.Now(),
		}
		txm.updateAction(transaction, action, InProgress)
		log.Debugf("Transaction %s, committing action %s", transactionID, actionID)
	}
}

func (txm *TransactionManager) updateAction(transaction *IntakeTransaction, action *Action, state TransactionState) {
	transaction.Actions[action.ActionID] = action
	transaction.State = state
	transaction.LastUpdatedTimestamp = time.Now()
}

// ackAction ...
func (txm *TransactionManager) ackAction(transactionID, actionID string) {
	transaction, exists := txm.Transactions[transactionID]
	if exists {
		action, exists := transaction.Actions[actionID]
		if exists {
			action.Acknowledged = true
			action.AcknowledgedTimestamp = time.Now()
			txm.updateAction(transaction, action, InProgress)
			log.Debugf("Transaction %s, acknowledged action %s", transactionID, actionID)
		}
	}
}

// rejectAction ...
func (txm *TransactionManager) rejectAction(transactionID, actionID string) {
	transaction, exists := txm.Transactions[transactionID]
	if exists {
		action, exists := transaction.Actions[actionID]
		if exists {
			action.Acknowledged = true
			action.AcknowledgedTimestamp = time.Now()
			txm.updateAction(transaction, action, Failed)
			log.Debugf("Transaction %s, acknowledged action %s", transactionID, actionID)
			txm.rollbackTransaction(transactionID)
		}
	}
}

// completeTransaction marks the manager successful
func (txm *TransactionManager) completeTransaction(transactionID string) {
	transaction, exists := txm.Transactions[transactionID]
	if exists {
		// ensure all actions have been acknowledged
		for _, action := range transaction.Actions {
			if !action.Acknowledged {
				_ = log.Errorf("Not all actions have been acknowledged, rolling back manager: %s", transaction.TransactionID)
				txm.rollbackTransaction(transactionID)
			}
		}
		transaction.State = Succeeded
		transaction.LastUpdatedTimestamp = time.Now()
		log.Debugf("Transaction succeeded %s", transaction.TransactionID)
	} else {
		_ = log.Warnf("Transaction not found %s, no operation", transaction.TransactionID)
	}

	if transaction.OnComplete != nil {
		transaction.OnComplete(transaction)
	}
}

// evictTransaction evicts the intake manager due to failure or timeout
func (txm *TransactionManager) evictTransaction(transactionID string) {
	delete(txm.Transactions, transactionID)
}

// rollbackTransaction rolls back the intake manager
func (txm *TransactionManager) rollbackTransaction(transactionID string) {
	transaction, exists := txm.Transactions[transactionID]
	if exists {
		// manager failed, rollback
		transaction.State = Failed
		transaction.LastUpdatedTimestamp = time.Now()
		log.Debugf("Transaction failed %s", transaction.TransactionID)
	}

	if transaction.OnComplete != nil {
		transaction.OnComplete(transaction)
	}
}
