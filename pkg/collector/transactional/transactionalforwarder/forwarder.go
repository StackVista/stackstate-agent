package transactionalforwarder

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/StackVista/stackstate-agent/pkg/httpclient"
)

// Payloads is a slice of pointers to byte arrays, an alias for the slices of
// payloads we pass into the forwarder
type Payloads []*[]byte

// TransactionalPayload contains the Payload and transactional data
type TransactionalPayload struct {
	Payload              []byte
	TransactionActionMap map[string]transactional.PayloadTransaction
}

// ShutdownForwarder shuts down the forwarder
type ShutdownForwarder struct{}

// Response contains the response details of a successfully posted checkmanager
type Response struct {
	Domain     string
	Body       []byte
	StatusCode int
	Err        error
}

type TransactionalForwarder interface {
	SubmitTransactionalIntake(payload TransactionalPayload)
}

// Forwarder is a forwarder that works in transactional manner
type Forwarder struct {
	stsClient       *httpclient.StackStateClient
	PayloadChannel  chan TransactionalPayload
	ShutdownChannel chan ShutdownForwarder
}

// MakeForwarder returns a instance of the forwarder
func MakeForwarder() *Forwarder {
	return &Forwarder{stsClient: httpclient.NewStackStateClient()}
}

// Start initialize and runs the transactional forwarder.
func (f *Forwarder) Start() {
forwardHandler:
	for {
		// handling high priority transactions first
		select {
		case tPayload := <-f.PayloadChannel:
			response := f.stsClient.Post("", tPayload.Payload)
			if response.Err != nil {
				// Payload failed, rollback checkmanager
				for transactionID, payloadTransaction := range tPayload.TransactionActionMap {
					transactionmanager.GetTransactionManager().RejectAction(transactionID, payloadTransaction.ActionID, response.Err.Error())
				}
			} else {
				// Payload succeeded, acknowledge action
				for transactionID, payloadTransaction := range tPayload.TransactionActionMap {
					transactionmanager.GetTransactionManager().AcknowledgeAction(transactionID, payloadTransaction.ActionID)

					// if the transaction of the payload is completed, submit a transaction complete
					if payloadTransaction.CompletedTransaction {
						transactionmanager.GetTransactionManager().CompleteTransaction(transactionID)
					}
				}
			}
		case _ = <-f.ShutdownChannel:
			break forwardHandler
		default:
		}
	}
}

// Stop stops running the transactional forwarder.
func (f *Forwarder) Stop() {
	// Shut down the forwardHandler
	f.ShutdownChannel <- ShutdownForwarder{}
	defer close(f.PayloadChannel)
	defer close(f.ShutdownChannel)
}

// SubmitTransactionalIntake publishes the Payload to the PayloadChannel
func (f *Forwarder) SubmitTransactionalIntake(payload TransactionalPayload) {
	f.PayloadChannel <- payload
}
