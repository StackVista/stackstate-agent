package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/manager"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCheckHandler(t *testing.T) {
	ch := MakeCheckHandler(&check.TestCheck{Name: "my-check-handler-test-check"}, &check.TestCheckReloader{},
		manager.MakeTransactionManager(100, 100*time.Millisecond, 500*time.Millisecond,
			500*time.Millisecond), batcher.MockBatcher{}, integration.Data{1, 2, 3}, integration.Data{0, 0, 0})

	assert.Equal(t, check.ID("my-check-handler-test-check"), ch.ID())
	assert.EqualValues(t, ch.GetBatcher(), batcher.MockBatcher{})
	actualInstanceCfg, actualInitCfg := ch.GetConfig()
	assert.EqualValues(t, integration.Data{1, 2, 3}, actualInstanceCfg)
	assert.EqualValues(t, integration.Data{0, 0, 0}, actualInitCfg)
	assert.Equal(t, "test-config-source", ch.ConfigSource())

	cr := ch.GetCheckReloader().(*check.TestCheckReloader)
	assert.Equal(t, 0, cr.Reloaded)
	err := ch.ReloadCheck(ch.ID(), actualInstanceCfg, actualInitCfg, ch.ConfigSource())
	assert.NoError(t, err)
	assert.Equal(t, 1, cr.Reloaded)
	err = ch.ReloadCheck(ch.ID(), actualInstanceCfg, actualInitCfg, ch.ConfigSource())
	assert.Equal(t, 2, cr.Reloaded)
}

func TestCheckHandler_Transactions(t *testing.T) {
	testTxManager := &TestTransactionManager{}
	ch := MakeCheckHandler(&check.TestCheck{Name: "my-check-handler-test-check"}, &check.TestCheckReloader{},
		testTxManager, batcher.MockBatcher{}, integration.Data{1, 2, 3}, integration.Data{0, 0, 0}).(*checkHandler)

	ch.Start()

	ch.StartTransaction("CheckID", "TransactionID")

	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, "TransactionID", testTxManager.CurrentTransaction)

	//testTxManager.CurrentTransactionNotifyChannel <- manager.CompleteTransaction{}

	ch.Stop()
}

type TestTransactionManager struct {
	CurrentTransaction              string
	CurrentTransactionNotifyChannel chan interface{}
}

func (ttm *TestTransactionManager) StartTransaction(_ check.ID, TransactionID string, NotifyChannel chan interface{}) {
	ttm.CurrentTransaction = TransactionID
	ttm.CurrentTransactionNotifyChannel = NotifyChannel
}
func (ttm *TestTransactionManager) CompleteTransaction(transactionID string) {

}
func (ttm *TestTransactionManager) RollbackTransaction(transactionID, reason string) {

}
func (ttm *TestTransactionManager) CommitAction(transactionID, actionID string) {

}
func (ttm *TestTransactionManager) AcknowledgeAction(transactionID, actionID string) {

}
func (ttm *TestTransactionManager) RejectAction(transactionID, actionID, reason string) {

}
