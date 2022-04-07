package manager

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMakeTransactionManager(t *testing.T) {
	transactionManager := MakeTransactionManager(100, 100*time.Millisecond, 10*time.Second,
		10*time.Second)

	transactionManager.Start()

	// assert that we're starting on a clean slate
	assert.Len(t, transactionManager.Transactions, 0)

	completeFunc := func(transaction *IntakeTransaction) {
		// assert that LastUpdatedTimestamp was set in the last 50 milliseconds. i.e. just updated
		assert.Equal(t, time.Now().After(time.Now().Add(-50*time.Millisecond)), transaction.LastUpdatedTimestamp)
		assert.Equal(t, InProgress, transaction.State)
	}

	// start a transaction and assert it
	txID := uuid.New().String()
	transactionManager.TransactionChannel <- StartTransaction{TransactionID: txID, OnComplete: completeFunc}
	assertTransaction(t, transactionManager, txID, InProgress, map[string]*Action{})

	time.Sleep(500 * time.Millisecond)
	defer transactionManager.Stop()

}

func assertTransaction(t *testing.T, transactionManager *TransactionManager, txID string, state TransactionState,
	actions map[string]*Action) {
	time.Sleep(250 * time.Millisecond)
	assert.Len(t, transactionManager.Transactions, 1)
	transaction, found := transactionManager.Transactions[txID]
	assert.True(t, found, "Transaction %s not found in the transaction map", txID)
	assert.Equal(t, txID, transaction.TransactionID)
	assert.Equal(t, state, transaction.State)
	assert.EqualValues(t, actions, transaction.Actions)
}
