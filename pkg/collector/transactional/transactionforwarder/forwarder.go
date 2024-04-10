package transactionforwarder

import (
	"github.com/DataDog/datadog-agent/pkg/collector/transactional"
	"github.com/DataDog/datadog-agent/pkg/collector/transactional/transactionmanager"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/httpclient"
	"github.com/DataDog/datadog-agent/pkg/util/log"
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
	tfInit                         *sync.Once
)

func init() {
	tfInit = new(sync.Once)
}

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
	tfInit.Do(func() {
		transactionalForwarderInstance = createMockForwarder()
	})
	return transactionalForwarderInstance.(*MockTransactionalForwarder)
}

// NewPrintingTransactionalForwarder initializes the global PrintingTransactionalForwarder used for the agent check command
func NewPrintingTransactionalForwarder() *PrintingTransactionalForwarder {
	tfInit.Do(func() {
		transactionalForwarderInstance = createPrintingForwarder()
	})
	return transactionalForwarderInstance.(*PrintingTransactionalForwarder)
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
			log.Debugf("Attempting to send transactional payload,\ntransactions: %v,content: %v",
				payload.TransactionActionMap, apiKeyRegExp.ReplaceAllString(string(payload.Body), apiKeyReplacement))

			response := f.stsClient.Post(payload.Path, payload.Body)
			if response.Err != nil {
				// Payload failed, reject action
				for transactionID, payloadTransaction := range payload.TransactionActionMap {
					log.Debugf("Sending transactional payload failed, rejecting action %s for transaction %s",
						payloadTransaction.ActionID, transactionID)
					transactionmanager.GetTransactionManager().RejectAction(transactionID, payloadTransaction.ActionID, response.Err.Error())
				}
				_ = log.Errorf("Sending transactional payload failed, content: %v. %s",
					apiKeyRegExp.ReplaceAllString(string(payload.Body), apiKeyReplacement), response.Err.Error())
			} else {
				f.ProgressTransactions(payload.TransactionActionMap)

				log.Infof("Sent transactional payload, size: %d bytes.", len(payload.Body))
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("Sent transactional payload, response status: %s (%d).", response.Response.Status,
						response.Response.StatusCode)
					log.Debugf("Sent transactional payload, content: %v", apiKeyRegExp.ReplaceAllString(string(payload.Body), apiKeyReplacement))
				}
			}
		case sf := <-f.ShutdownChannel:
			log.Infof("Shutting down forwarder %v", sf)
			break forwardHandler
		}
	}
}

// ProgressTransactions is called on a successful payload post or when OnlyMarkTransactions is set to true. It acknowledges
// the actions within a transaction and completes a completed transaction.
func (f *Forwarder) ProgressTransactions(transactionMap map[string]transactional.PayloadTransaction) {
	// Payload succeeded, acknowledge action
	for transactionID, payloadTransaction := range transactionMap {
		log.Debugf("Sent transactional payload successfully, acknowledging action %s for transaction %s",
			payloadTransaction.ActionID, transactionID)
		transactionmanager.GetTransactionManager().AcknowledgeAction(transactionID, payloadTransaction.ActionID)

		// if the transaction of the payload is completed, submit a transaction complete
		if payloadTransaction.CompletedTransaction {
			log.Debugf("Sent transactional payload successfully, completing transaction %s", transactionID)
			transactionmanager.GetTransactionManager().CompleteTransaction(transactionID)
		}
	}
}

// Stop stops running the transactional forwarder.
func (f *Forwarder) Stop() {
	// Shut down the forwardHandler
	f.ShutdownChannel <- ShutdownForwarder{}

	// reset the forwarder to re-init it later
	tfInit = new(sync.Once)
}

// SubmitTransactionalIntake publishes the Payload to the PayloadChannel
func (f *Forwarder) SubmitTransactionalIntake(payload TransactionalPayload) {
	f.PayloadChannel <- payload
}
