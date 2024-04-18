package transactionmanager

import (
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"sync"
)

// TransactionManager encapsulates all the functionality of the transaction manager to keep track of transactions
type TransactionManager interface {
	Start()
	TransactionAPI
	Stop()
}

// GetTransaction returns the IntakeTransaction for a given checkID and transactionID or TransactionNotFound error.
func (txm *transactionManager) GetTransaction(transactionID string) (*IntakeTransaction, error) {
	txm.mux.RLock()
	transaction, exists := txm.transactions[transactionID]
	txm.mux.RUnlock()
	if !exists {
		return nil, TransactionNotFound{TransactionID: transactionID}
	}

	return transaction, nil
}

// GetActiveTransaction returns the IntakeTransaction for a given checkID and transactionID or TransactionNotFound error
// for all 'active' (InProgress | Stale) transactions
func (txm *transactionManager) GetActiveTransaction(transactionID string) (*IntakeTransaction, error) {
	transaction, err := txm.GetTransaction(transactionID)
	if err != nil {
		return nil, err
	}

	switch transaction.Status {
	case InProgress, Stale:
		return transaction, nil
	default:
		_ = log.Warnf("GetActiveTransaction called for transaction %s that is in %s state. Returning TransactionNotFound",
			transactionID, transaction.Status)
		return nil, TransactionCompleted{TransactionID: transactionID}
	}
}

// TransactionCount returns the amount of transactions in the Transaction Manager.
func (txm *transactionManager) TransactionCount() int {
	txm.mux.RLock()
	count := len(txm.transactions)
	txm.mux.RUnlock()
	return count
}

// StartTransaction begins a transaction for a given check
func (txm *transactionManager) StartTransaction(checkID checkid.ID, transactionID string, notifyChannel chan interface{}) {
	txm.transactionChannel <- StartTransaction{
		CheckID:       checkID,
		TransactionID: transactionID,
		NotifyChannel: notifyChannel,
	}
}

// CompleteTransaction completes a transaction for a given transactionID
func (txm *transactionManager) CompleteTransaction(transactionID string) {
	txm.transactionChannel <- CompleteTransaction{
		TransactionID: transactionID,
	}
}

// SetState adds a state to a transaction that will be committed on a successful transaction
func (txm *transactionManager) SetState(transactionID, key string, state string) {
	txm.transactionChannel <- SetTransactionState{
		TransactionID: transactionID,
		Key:           key,
		State:         state,
	}
}

// DiscardTransaction rolls back a transaction for a given transactionID and a reason for the discard
func (txm *transactionManager) DiscardTransaction(transactionID, reason string) {
	txm.transactionChannel <- DiscardTransaction{
		TransactionID: transactionID,
		Reason:        reason,
	}
}

// CommitAction commits an action for a given transaction. All actions must be acknowledged for a given transaction
func (txm *transactionManager) CommitAction(transactionID, actionID string) {
	txm.transactionChannel <- CommitAction{
		TransactionID: transactionID,
		ActionID:      actionID,
	}
}

// AcknowledgeAction acknowledges an action for a given transaction
func (txm *transactionManager) AcknowledgeAction(transactionID, actionID string) {
	txm.transactionChannel <- AckAction{
		TransactionID: transactionID,
		ActionID:      actionID,
	}
}

// RejectAction rejects an action for a given transaction. This will result in a transaction failure
func (txm *transactionManager) RejectAction(transactionID, actionID, reason string) {
	txm.transactionChannel <- RejectAction{
		TransactionID: transactionID,
		ActionID:      actionID,
		Reason:        reason,
	}
}

// Stop shuts down the transaction checkmanager and stops the transactionHandler receiver loop
func (txm *transactionManager) Stop() {
	txm.transactionChannel <- StopTransactionManager{}
	txm.transactionTicker.Stop()
	// reset the tmInit to re-init the transaction manager
	tmInit = new(sync.Once)
}
