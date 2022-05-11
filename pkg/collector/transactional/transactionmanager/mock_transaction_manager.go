package transactionmanager

import "github.com/StackVista/stackstate-agent/pkg/collector/check"

func newTestTransactionManager() *MockTransactionManager {
	return &MockTransactionManager{TransactionActions: make(chan interface{}, 100)}
}

type MockTransactionManager struct {
	CurrentTransaction              string
	CurrentTransactionNotifyChannel chan interface{}
	TransactionActions              chan interface{}
}

func (ttm *MockTransactionManager) Start() {
}

func (ttm *MockTransactionManager) Stop() {
}

func (ttm *MockTransactionManager) StartTransaction(_ check.ID, TransactionID string, NotifyChannel chan interface{}) {
	ttm.CurrentTransaction = TransactionID
	ttm.CurrentTransactionNotifyChannel = NotifyChannel
}
func (ttm *MockTransactionManager) CompleteTransaction(transactionID string) {
	ttm.TransactionActions <- CompleteTransaction{TransactionID: transactionID}
}
func (ttm *MockTransactionManager) RollbackTransaction(transactionID, reason string) {
	ttm.TransactionActions <- RollbackTransaction{TransactionID: transactionID, Reason: reason}
}
func (ttm *MockTransactionManager) CommitAction(transactionID, actionID string) {
	ttm.TransactionActions <- CommitAction{TransactionID: transactionID, ActionID: actionID}
}
func (ttm *MockTransactionManager) AcknowledgeAction(transactionID, actionID string) {
	ttm.TransactionActions <- AckAction{TransactionID: transactionID, ActionID: actionID}
}
func (ttm *MockTransactionManager) RejectAction(transactionID, actionID, reason string) {
	ttm.TransactionActions <- RejectAction{TransactionID: transactionID, ActionID: actionID, Reason: reason}
}

func (ttm *MockTransactionManager) NextAction() interface{} {
	return <-ttm.TransactionActions
}
