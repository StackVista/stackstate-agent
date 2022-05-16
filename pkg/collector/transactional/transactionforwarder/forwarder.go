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

// Payloads is a slice of pointers to byte arrays, an alias for the slices of
// payloads we pass into the forwarder
type Payloads []*[]byte

// TransactionalPayload contains the Payload and transactional data
type TransactionalPayload struct {
	Payload              []byte
	Path                 string
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
func NewMockTransactionalForwarder() *mockForwarder {
	mf := createMockForwarder()
	transactionalForwarderInstance = mf
	return mf
}

// newTransactionalForwarder returns a instance of the forwarder
func newTransactionalForwarder() *Forwarder {
	return &Forwarder{stsClient: httpclient.NewStackStateClient()}
}

// Start initialize and runs the transactional forwarder.
func (f *Forwarder) Start() {
forwardHandler:
	for {
		select {
		case tPayload := <-f.PayloadChannel:
			response := f.stsClient.Post(tPayload.Path, tPayload.Payload)
			if response.Err != nil {
				// Payload failed, rollback transaction
				for transactionID, payloadTransaction := range tPayload.TransactionActionMap {
					transactionmanager.GetTransactionManager().RejectAction(transactionID, payloadTransaction.ActionID, response.Err.Error())
				}
				_ = log.Errorf("Failed to send intake payload, content: %v. %s",
					apiKeyRegExp.ReplaceAllString(string(tPayload.Payload), apiKeyReplacement), response.Err.Error())
			} else {
				// Payload succeeded, acknowledge action
				for transactionID, payloadTransaction := range tPayload.TransactionActionMap {
					transactionmanager.GetTransactionManager().AcknowledgeAction(transactionID, payloadTransaction.ActionID)

					// if the transaction of the payload is completed, submit a transaction complete
					if payloadTransaction.CompletedTransaction {
						transactionmanager.GetTransactionManager().CompleteTransaction(transactionID)
					}
				}

				log.Infof("Sent intake payload, size: %d bytes.", len(tPayload.Payload))
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("Sent intake payload, content: %v", apiKeyRegExp.ReplaceAllString(string(tPayload.Payload), apiKeyReplacement))
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
