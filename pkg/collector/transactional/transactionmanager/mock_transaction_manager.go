package transactionmanager

import "github.com/StackVista/stackstate-agent/pkg/collector/check"

func newTestTransactionManager() *MockTransactionManager {
	return &MockTransactionManager{}
}

type MockTransactionManager struct {
	CurrentTransaction              string
	CurrentTransactionNotifyChannel chan interface{}
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

}
func (ttm *MockTransactionManager) RollbackTransaction(transactionID, reason string) {

}
func (ttm *MockTransactionManager) CommitAction(transactionID, actionID string) {

}
func (ttm *MockTransactionManager) AcknowledgeAction(transactionID, actionID string) {

}
func (ttm *MockTransactionManager) RejectAction(transactionID, actionID, reason string) {

}
