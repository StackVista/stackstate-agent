package transactionmanager

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"sync"
)

func newTestTransactionManager() *MockTransactionManager {
	return &MockTransactionManager{TransactionActions: make(chan interface{}, 100)}
}

// MockTransactionManager is a mock implementation of the transaction manager for tests
type MockTransactionManager struct {
	mux                             sync.Mutex
	currentTransaction              string
	currentTransactionNotifyChannel chan interface{}
	TransactionActions              chan interface{}
}

// GetTransaction returns nil, nil
func (ttm *MockTransactionManager) GetTransaction(string) (*IntakeTransaction, error) {
	return nil, nil
}

// TransactionCount return 0
func (ttm *MockTransactionManager) TransactionCount() int {
	return 0
}

// Start is a noop
func (ttm *MockTransactionManager) Start() {
}

// Stop is a noop
func (ttm *MockTransactionManager) Stop() {
}

// StartTransaction sets the current transaction value and updates the notify channel
func (ttm *MockTransactionManager) StartTransaction(_ check.ID, TransactionID string, NotifyChannel chan interface{}) {
	ttm.mux.Lock()
	ttm.currentTransaction = TransactionID
	ttm.currentTransactionNotifyChannel = NotifyChannel
	ttm.mux.Unlock()
}

// GetCurrentTransactionNotifyChannel returns the currentTransactionNotifyChannel
func (ttm *MockTransactionManager) GetCurrentTransactionNotifyChannel() chan interface{} {
	ttm.mux.Lock()
	defer ttm.mux.Unlock()
	return ttm.currentTransactionNotifyChannel
}

// CompleteTransaction sends a CompleteTransaction to the TransactionActions channel to be used in assertions
func (ttm *MockTransactionManager) CompleteTransaction(transactionID string) {
	ttm.TransactionActions <- CompleteTransaction{TransactionID: transactionID}
}

// RollbackTransaction sends a RollbackTransaction to the TransactionActions channel to be used in assertions
func (ttm *MockTransactionManager) RollbackTransaction(transactionID, reason string) {
	ttm.TransactionActions <- RollbackTransaction{TransactionID: transactionID, Reason: reason}
}

// CommitAction sends a CommitAction to the TransactionActions channel to be used in assertions
func (ttm *MockTransactionManager) CommitAction(transactionID, actionID string) {
	ttm.TransactionActions <- CommitAction{TransactionID: transactionID, ActionID: actionID}
}

// AcknowledgeAction sends a AckAction to the TransactionActions channel to be used in assertions
func (ttm *MockTransactionManager) AcknowledgeAction(transactionID, actionID string) {
	ttm.TransactionActions <- AckAction{TransactionID: transactionID, ActionID: actionID}
}

// RejectAction sends a RejectAction to the TransactionActions channel to be used in assertions
func (ttm *MockTransactionManager) RejectAction(transactionID, actionID, reason string) {
	ttm.TransactionActions <- RejectAction{TransactionID: transactionID, ActionID: actionID, Reason: reason}
}

// NextAction returns the next action from the TransactionActions channel to be used in assertions
func (ttm *MockTransactionManager) NextAction() interface{} {
	return <-ttm.TransactionActions
}
