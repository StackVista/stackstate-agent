package transactionforwarder

import (
	"encoding/json"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/fatih/color"
)

func createPrintingForwarder() *PrintingTransactionalForwarder {
	return &PrintingTransactionalForwarder{PayloadChan: make(chan TransactionalPayload, 100)}
}

// PrintingTransactionalForwarder is a implementation of the transactional forwarder that prints the payload
type PrintingTransactionalForwarder struct {
	PayloadChan chan TransactionalPayload
}

// Start is a noop
func (mf *PrintingTransactionalForwarder) Start() {}

// SubmitTransactionalIntake receives a TransactionalPayload and keeps it in the PayloadChan to be used in assertions
func (mf *PrintingTransactionalForwarder) SubmitTransactionalIntake(payload TransactionalPayload) {

	// Acknowledge actions and succeed transactions
	for transactionID, payloadTransaction := range payload.TransactionActionMap {
		transactionmanager.GetTransactionManager().AcknowledgeAction(transactionID, payloadTransaction.ActionID)

		// if the transaction of the payload is completed, submit a transaction complete
		if payloadTransaction.CompletedTransaction {
			transactionmanager.GetTransactionManager().CompleteTransaction(transactionID)
		}
	}

	actualPayload := transactional.NewIntakePayload()
	_ = json.Unmarshal(payload.Body, &actualPayload)

	fmt.Fprintln(color.Output, fmt.Sprintf("=== %s ===", color.BlueString("Topology")))
	j, _ := json.MarshalIndent(actualPayload, "", "  ")
	fmt.Println(string(j))
}

// Stop is a noop
func (mf *PrintingTransactionalForwarder) Stop() {
	close(mf.PayloadChan)
}
