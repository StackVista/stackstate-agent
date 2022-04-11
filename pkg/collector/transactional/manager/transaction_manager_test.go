package manager

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTransactionManager_HappyFlow(t *testing.T) {
	transactionManager := MakeTransactionManager(100, 100*time.Millisecond, 500*time.Millisecond,
		500*time.Millisecond)

	assert.False(t, transactionManager.running)
	transactionManager.Start()
	assert.True(t, transactionManager.running)

	// assert that we're starting on a clean slate
	assert.Len(t, transactionManager.Transactions, 0)

	completeFunc := func(transaction *IntakeTransaction) {
		// assert that LastUpdatedTimestamp was set in the last 50 milliseconds. i.e. just updated
		assert.True(t, transaction.LastUpdatedTimestamp.After(time.Now().Add(-50*time.Millisecond)))
		assert.Equal(t, Succeeded, transaction.State)
	}

	// start a transaction and assert it
	txID := uuid.New().String()
	transactionManager.TransactionChannel <- StartTransaction{TransactionID: txID, OnComplete: completeFunc}
	assertTransaction(t, transactionManager, txID, InProgress, map[string]*Action{})

	// commit 15 actions and assert them
	actions := make(map[string]*Action, 5)
	for i := 0; i < 15; i++ {
		actID := fmt.Sprintf("action-%d", i)
		transactionManager.TransactionChannel <- CommitAction{TransactionID: txID, ActionID: actID}
		actions[actID] = &Action{ActionID: actID, Acknowledged: false}
		assertTransaction(t, transactionManager, txID, InProgress, actions)
	}
	// acknowledge 15 action and assert them
	for i := 0; i < 15; i++ {
		actID := fmt.Sprintf("action-%d", i)
		transactionManager.TransactionChannel <- AckAction{TransactionID: txID, ActionID: actID}
		actions[actID] = &Action{ActionID: actID, Acknowledged: true}
		assertTransaction(t, transactionManager, txID, InProgress, actions)
	}

	// start a transaction and assert it
	transactionManager.TransactionChannel <- CompleteTransaction{TransactionID: txID}
	assertTransaction(t, transactionManager, txID, Succeeded, actions)

	defer transactionManager.Stop()

	// sleep and wait for automatic cleanup to remove the successful transaction
	time.Sleep(1 * time.Second)
	assert.Len(t, transactionManager.Transactions, 0)
}

func TestTransactionManager_TransactionRollback(t *testing.T) {
	transactionManager := MakeTransactionManager(100, 100*time.Millisecond, 500*time.Millisecond,
		500*time.Millisecond)

	transactionManager.Start()

	for _, tc := range []struct {
		testCase  string
		operation func(t *testing.T, manager *TransactionManager)
	}{
		{
			testCase: "Transaction rollback triggered by external party (check handler)",
			operation: func(t *testing.T, manager *TransactionManager) {
				txID := uuid.New().String()
				transactionManager.TransactionChannel <- StartTransaction{TransactionID: txID}
				assertTransaction(t, transactionManager, txID, InProgress, map[string]*Action{})
				transactionManager.TransactionChannel <- RollbackTransaction{TransactionID: txID, Reason: "check failed"}
				assertTransaction(t, transactionManager, txID, Failed, map[string]*Action{})
			},
		},
		{
			testCase: "Transaction rollback triggered by an un-acknowledged action",
			operation: func(t *testing.T, manager *TransactionManager) {
				txID := uuid.New().String()
				transactionManager.TransactionChannel <- StartTransaction{TransactionID: txID}
				assertTransaction(t, transactionManager, txID, InProgress, map[string]*Action{})

				actions := make(map[string]*Action, 1)
				actID := uuid.New().String()
				transactionManager.TransactionChannel <- CommitAction{TransactionID: txID, ActionID: actID}
				actions[actID] = &Action{ActionID: actID, Acknowledged: false}
				assertTransaction(t, transactionManager, txID, InProgress, actions)

				transactionManager.TransactionChannel <- CompleteTransaction{TransactionID: txID}
				assertTransaction(t, transactionManager, txID, Failed, actions)
			},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			tc.operation(t, transactionManager)
		})
	}

	defer transactionManager.Stop()

	// sleep and wait for automatic cleanup to remove the successful transaction
	time.Sleep(1 * time.Second)
	assert.Len(t, transactionManager.Transactions, 0)
}

func assertTransaction(t *testing.T, transactionManager *TransactionManager, txID string, state TransactionState,
	actions map[string]*Action) {
	// give the transaction manager a bit of time to insert the transaction before running the assertion
	time.Sleep(10 * time.Millisecond)
	transaction, found := transactionManager.Transactions[txID]
	assert.True(t, found, "Transaction %s not found in the transaction map", txID)
	assert.Equal(t, txID, transaction.TransactionID)
	assert.Equal(t, state, transaction.State)
	assert.Equal(t, len(actions), len(transaction.Actions))
	for _, action := range transaction.Actions {
		expectedAction, found := actions[action.ActionID]
		assert.True(t, found)
		assert.Equal(t, expectedAction.ActionID, action.ActionID)
		assert.Equal(t, expectedAction.Acknowledged, action.Acknowledged)
	}
}
