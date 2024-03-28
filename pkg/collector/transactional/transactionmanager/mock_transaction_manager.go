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
	transactionState                *TransactionState
	TransactionActions              chan interface{}
}

// SetState sets the mock transactionState to the given key + value
func (ttm *MockTransactionManager) SetState(_, key string, value string) {
	ttm.mux.Lock()
	ttm.transactionState = &TransactionState{
		Key:   key,
		State: value,
	}
	ttm.mux.Unlock()
}

// GetActiveTransaction returns nil, nil
func (ttm *MockTransactionManager) GetActiveTransaction(string) (*IntakeTransaction, error) {
	return nil, nil
}

// TransactionCount return 0
func (ttm *MockTransactionManager) TransactionCount() int {
	return 0
}

// Start is a noop
func (ttm *MockTransactionManager) Start() {
}

// Stop resets the singleton init
func (ttm *MockTransactionManager) Stop() {
	// reset the transaction manager init
	tmInit = new(sync.Once)
}

// StartTransaction sets the current transaction value and updates the notify channel
func (ttm *MockTransactionManager) StartTransaction(_ check.ID, TransactionID string, NotifyChannel chan interface{}) {
	ttm.mux.Lock()
	ttm.currentTransaction = TransactionID
	ttm.currentTransactionNotifyChannel = NotifyChannel
	ttm.mux.Unlock()
}

// GetCurrentTransaction returns the current transaction
func (ttm *MockTransactionManager) GetCurrentTransaction() string {
	ttm.mux.Lock()
	defer ttm.mux.Unlock()
	return ttm.currentTransaction
}

// GetCurrentTransactionNotifyChannel returns the currentTransactionNotifyChannel
func (ttm *MockTransactionManager) GetCurrentTransactionNotifyChannel() chan interface{} {
	ttm.mux.Lock()
	defer ttm.mux.Unlock()
	return ttm.currentTransactionNotifyChannel
}

// GetCurrentTransactionState returns the transactionState
func (ttm *MockTransactionManager) GetCurrentTransactionState() *TransactionState {
	ttm.mux.Lock()
	defer ttm.mux.Unlock()
	return ttm.transactionState
}

// CompleteTransaction sends a CompleteTransaction to the TransactionActions channel to be used in assertions
func (ttm *MockTransactionManager) CompleteTransaction(transactionID string) {
	ttm.TransactionActions <- CompleteTransaction{TransactionID: transactionID}
}

// DiscardTransaction sends a DiscardTransaction to the TransactionActions channel to be used in assertions
func (ttm *MockTransactionManager) DiscardTransaction(transactionID, reason string) {
	ttm.TransactionActions <- DiscardTransaction{TransactionID: transactionID, Reason: reason}
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
