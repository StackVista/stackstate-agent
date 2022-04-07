package transactional

import (
	"github.com/StackVista/stackstate-agent/cmd/agent/common"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/manager"
	"github.com/StackVista/stackstate-agent/pkg/httpclient"
)

// Payloads is a slice of pointers to byte arrays, an alias for the slices of
// payloads we pass into the forwarder
type Payloads []*[]byte

// TransactionalPayload contains the payload and transactional data
type TransactionalPayload struct {
	payload                 []byte
	transactionID, actionID string
}

type ShutdownForwarder struct{}

// Response contains the response details of a successfully posted manager
type Response struct {
	Domain     string
	Body       []byte
	StatusCode int
	Err        error
}

// Forwarder is a forwarder that works in transactional manner
type Forwarder struct {
	stsClient       *httpclient.StackStateClient
	PayloadChannel  chan TransactionalPayload
	ShutdownChannel chan ShutdownForwarder
}

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
			response := f.stsClient.Post("", tPayload.payload)
			if response.Err != nil {
				// payload failed, rollback manager
				common.TxManager.TransactionChannel <- &manager.RollbackTransaction{TransactionID: tPayload.transactionID}
			} else {
				// payload succeeded, acknowledge action
				common.TxManager.TransactionChannel <- &manager.AckAction{
					TransactionID: tPayload.transactionID,
					ActionID:      tPayload.actionID,
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

func (f *Forwarder) SubmitTransactionalIntake(payload TransactionalPayload) {
	f.PayloadChannel <- payload
}
