package transactionforwarder

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/httpclient"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"regexp"
	"sync"
)

const apiKeyReplacement = "\"apiKey\":\"*************************$1"

var apiKeyRegExp = regexp.MustCompile("\"apiKey\":\"*\\w+(\\w{5})")

// TransactionalPayload contains the Payload and transactional data
type TransactionalPayload struct {
	Body                 []byte
	Path                 string
	TransactionActionMap map[string]transactional.PayloadTransaction
	OnlyMarkTransactions bool // this is used to bypass the actual sending of data on empty payloads and just complete the transactions
}

// ShutdownForwarder shuts down the forwarder
type ShutdownForwarder struct{}

// TransactionalForwarder encapsulates the functionality for a transactional forwarder
type TransactionalForwarder interface {
	Start()
	SubmitTransactionalIntake(payload TransactionalPayload)
	Stop()
}

// Forwarder is a forwarder that works in transactional manner
type Forwarder struct {
	stsClient       httpclient.RetryableHTTPClient
	PayloadChannel  chan TransactionalPayload
	ShutdownChannel chan ShutdownForwarder
}

var (
	transactionalForwarderInstance TransactionalForwarder
	tfInit                         sync.Once
)

// InitTransactionalForwarder initializes the global transactional forwarder Instance
func InitTransactionalForwarder() {
	tfInit.Do(func() {
		transactionalForwarderInstance = newTransactionalForwarder()
	})
}

// GetTransactionalForwarder ...
func GetTransactionalForwarder() TransactionalForwarder {
	return transactionalForwarderInstance
}

// NewMockTransactionalForwarder initializes the global TransactionalForwarder with a mock version, intended for testing
func NewMockTransactionalForwarder() *MockTransactionalForwarder {
	mf := createMockForwarder()
	transactionalForwarderInstance = mf
	return mf
}

// newTransactionalForwarder returns a instance of the forwarder
func newTransactionalForwarder() *Forwarder {
	fwd := &Forwarder{
		stsClient:       httpclient.NewStackStateClient(),
		PayloadChannel:  make(chan TransactionalPayload, 100),
		ShutdownChannel: make(chan ShutdownForwarder, 1),
	}

	go fwd.Start()

	return fwd
}

// Start initialize and runs the transactional forwarder.
func (f *Forwarder) Start() {
forwardHandler:
	for {
		select {
		case payload := <-f.PayloadChannel:

			// check to see if this is an empty payload -> OnlyMarkTransactions == true
			if payload.OnlyMarkTransactions {
				f.ProgressTransactions(payload.TransactionActionMap)
				return
			}

			response := f.stsClient.Post(payload.Path, payload.Body)
			if response.Err != nil {
				// Payload failed, reject action
				for transactionID, payloadTransaction := range payload.TransactionActionMap {
					transactionmanager.GetTransactionManager().RejectAction(transactionID, payloadTransaction.ActionID, response.Err.Error())
				}
				_ = log.Errorf("Failed to send intake payload, content: %v. %s",
					apiKeyRegExp.ReplaceAllString(string(payload.Body), apiKeyReplacement), response.Err.Error())
			} else {
				f.ProgressTransactions(payload.TransactionActionMap)

				log.Infof("Sent intake payload, size: %d bytes.", len(payload.Body))
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("Sent intake payload, response status: %s (%d).", response.Response.Status,
						response.Response.StatusCode)
					log.Debugf("Sent intake payload, content: %v", apiKeyRegExp.ReplaceAllString(string(payload.Body), apiKeyReplacement))
				}
			}
		case sf := <-f.ShutdownChannel:
			log.Infof("Shutting down forwarder %v", sf)
			break forwardHandler
		default:
		}
	}
}

// ProgressTransactions is called on a successful payload post or when OnlyMarkTransactions is set to true. It acknowledges
// the actions within a transaction and completes a completed transaction.
func (f *Forwarder) ProgressTransactions(transactionMap map[string]transactional.PayloadTransaction) {
	// Payload succeeded, acknowledge action
	for transactionID, payloadTransaction := range transactionMap {
		transactionmanager.GetTransactionManager().AcknowledgeAction(transactionID, payloadTransaction.ActionID)

		// if the transaction of the payload is completed, submit a transaction complete
		if payloadTransaction.CompletedTransaction {
			transactionmanager.GetTransactionManager().CompleteTransaction(transactionID)
		}
	}
}

// Stop stops running the transactional forwarder.
func (f *Forwarder) Stop() {
	// Shut down the forwardHandler
	f.ShutdownChannel <- ShutdownForwarder{}
}

// SubmitTransactionalIntake publishes the Payload to the PayloadChannel
func (f *Forwarder) SubmitTransactionalIntake(payload TransactionalPayload) {
	f.PayloadChannel <- payload
}
