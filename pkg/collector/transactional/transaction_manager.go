package transactional

import (
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"github.com/google/uuid"
	"time"
)

// CommitAction ...
type CommitAction struct {
	transactionID, actionID string
}

// AckAction ...
type AckAction struct {
	transactionID, actionID string
}

// RejectAction ...
type RejectAction struct {
	transactionID, actionID string
	reason                  error
}

// RollbackTransaction ...
type RollbackTransaction struct {
	transactionID string
	reason        error
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
func (txm TransactionManager) Start() {
transactionHandler:
	for {
		select {
		case input := <-txm.TransactionChannel:
			switch msg := input.(type) {
			case CommitAction:
				log.Debugf("Committing msg %s for transaction %s", msg.actionID, msg.transactionID)
				txm.commitAction(msg.transactionID, msg.actionID)
			case AckAction:
				log.Debugf("Acknowledging msg %s for transaction %s", msg.actionID, msg.transactionID)
				txm.ackAction(msg.transactionID, msg.actionID)
			case RejectAction:
				_ = log.Errorf("Rejecting msg %s for transaction %s: %s", msg.actionID, msg.transactionID, msg.reason)
				txm.rejectAction(msg.transactionID, msg.actionID)
			case RollbackTransaction:
				_ = log.Errorf("Rolling back transaction %s: %s", msg.transactionID, msg.reason)
				txm.rollbackTransaction(msg.transactionID)
			case StopTransactionManager:
				break transactionHandler
			}
		case <-txm.TransactionTicker.C:
			// expire stale transactions, clean up expired transactions that exceed the eviction duration
			for _, transaction := range txm.Transactions {
				if transaction.LastUpdatedTimestamp.After(time.Now().Add(-txm.transactionTimeoutDuration)) {
					// last updated timestamp is before current time - transaction timeout duration => Tx is stale
					transaction.State = Stale
				} else if transaction.State == Stale && transaction.LastUpdatedTimestamp.After(time.Now().Add(-txm.transactionEvictionDuration)) {
					// last updated timestamp is before current time - transaction eviction duration => Tx can be evicted
					txm.evictTransaction(transaction.TransactionID)
				}
			}
		default:
		}
	}
}

// Stop ...
func (txm TransactionManager) Stop() {
	txm.TransactionChannel <- StopTransactionManager{}
	txm.TransactionTicker.Stop()
}

// startTransaction ...
func (txm TransactionManager) startTransaction() *IntakeTransaction {
	transaction := &IntakeTransaction{
		TransactionID:        uuid.New().String(),
		State:                InProgress,
		Actions:              map[string]*Action{},
		LastUpdatedTimestamp: time.Now(),
	}
	txm.Transactions[transaction.TransactionID] = transaction

	return transaction
}

// commitAction ...
func (txm TransactionManager) commitAction(transactionID, actionID string) {
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

func (txm TransactionManager) updateAction(transaction *IntakeTransaction, action *Action, state TransactionState) {
	transaction.Actions[action.ActionID] = action
	transaction.State = state
	transaction.LastUpdatedTimestamp = time.Now()
}

// ackAction ...
func (txm TransactionManager) ackAction(transactionID, actionID string) {
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
func (txm TransactionManager) rejectAction(transactionID, actionID string) {
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

// completeTransaction marks the transaction successful
func (txm TransactionManager) completeTransaction(transactionID string) {
	transaction, exists := txm.Transactions[transactionID]
	if exists {
		// ensure all actions have been acknowledged
		for _, action := range transaction.Actions {
			if !action.Acknowledged {
				_ = log.Errorf("Not all actions have been acknowledged, rolling back transaction: %s", transaction.TransactionID)
				txm.rollbackTransaction(transactionID)
			}
		}
		transaction.State = Succeeded
		transaction.LastUpdatedTimestamp = time.Now()
		log.Debugf("Transaction succeeded %s", transaction.TransactionID)
	} else {
		_ = log.Warnf("Transaction not found %s, no operation", transaction.TransactionID)
	}

	transaction.OnComplete(transaction)

	txm.evictTransaction(transactionID)
}

// evictTransaction evicts the intake transaction due to failure or timeout
func (txm TransactionManager) evictTransaction(transactionID string) {
	delete(txm.Transactions, transactionID)
}

// rollbackTransaction rolls back the intake transaction
func (txm TransactionManager) rollbackTransaction(transactionID string) {
	transaction, exists := txm.Transactions[transactionID]
	if exists {
		// transaction failed, rollback
		transaction.State = Failed
		transaction.LastUpdatedTimestamp = time.Now()
		log.Debugf("Transaction failed %s", transaction.TransactionID)
	}

	transaction.OnComplete(transaction)

	txm.evictTransaction(transactionID)
}
