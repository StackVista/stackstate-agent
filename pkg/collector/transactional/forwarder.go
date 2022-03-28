package transactional

import (
	"errors"
	"github.com/StackVista/stackstate-agent/pkg/httpclient"
	"net/http"
)

// Payloads is a slice of pointers to byte arrays, an alias for the slices of
// payloads we pass into the forwarder
type Payloads []*[]byte

// Response contains the response details of a successfully posted transaction
type Response struct {
	Domain     string
	Body       []byte
	StatusCode int
	Err        error
}

// TransactionalForwarder is a forwarder that works in transactional manner
type TransactionalForwarder struct {
	stsClient *httpclient.StackStateClient
}

func MakeTransactionalForwarder() *TransactionalForwarder {
	return &TransactionalForwarder{stsClient: httpclient.NewStackStateClient()}
}

// Start initialize and runs the transactional forwarder.
func (f *TransactionalForwarder) Start() error {

	return nil
}

// Stop stops running the transactional forwarder.
func (f *TransactionalForwarder) Stop() error {

	return nil
}

func (f *TransactionalForwarder) SubmitV1Intake(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}

func (f *TransactionalForwarder) SubmitV1Series(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}

func (f *TransactionalForwarder) SubmitV1CheckRuns(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}
func (f *TransactionalForwarder) SubmitSeries(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}
func (f *TransactionalForwarder) SubmitEvents(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}
func (f *TransactionalForwarder) SubmitServiceChecks(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}
func (f *TransactionalForwarder) SubmitSketchSeries(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}
func (f *TransactionalForwarder) SubmitHostMetadata(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}
func (f *TransactionalForwarder) SubmitMetadata(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}
func (f *TransactionalForwarder) SubmitProcessChecks(payload Payloads, extra http.Header) (chan Response, error) {
	return nil, errors.New("NotImplemented")
}
func (f *TransactionalForwarder) SubmitRTProcessChecks(payload Payloads, extra http.Header) (chan Response, error) {
	return nil, errors.New("NotImplemented")
}
func (f *TransactionalForwarder) SubmitContainerChecks(payload Payloads, extra http.Header) (chan Response, error) {
	return nil, errors.New("NotImplemented")
}
func (f *TransactionalForwarder) SubmitRTContainerChecks(payload Payloads, extra http.Header) (chan Response, error) {
	return nil, errors.New("NotImplemented")
}
func (f *TransactionalForwarder) SubmitConnectionChecks(payload Payloads, extra http.Header) (chan Response, error) {
	return nil, errors.New("NotImplemented")
}
func (f *TransactionalForwarder) SubmitPodChecks(payload Payloads, extra http.Header) (chan Response, error) {
	return nil, errors.New("NotImplemented")
}
