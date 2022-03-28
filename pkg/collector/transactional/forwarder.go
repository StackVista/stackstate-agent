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

// Forwarder is a forwarder that works in transactional manner
type Forwarder struct {
	stsClient *httpclient.StackStateClient
}

func MakeForwarder() *Forwarder {
	return &Forwarder{stsClient: httpclient.NewStackStateClient()}
}

// Start initialize and runs the transactional forwarder.
func (f *Forwarder) Start() error {
	for {
		// handling high priority transactions first
		select {
		default:
		}
	}
}

// Stop stops running the transactional forwarder.
func (f *Forwarder) Stop() error {

	return nil
}

func (f *Forwarder) SubmitV1Intake(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}

func (f *Forwarder) SubmitV1Series(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}

func (f *Forwarder) SubmitV1CheckRuns(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}
func (f *Forwarder) SubmitSeries(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}
func (f *Forwarder) SubmitEvents(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}
func (f *Forwarder) SubmitServiceChecks(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}
func (f *Forwarder) SubmitSketchSeries(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}
func (f *Forwarder) SubmitHostMetadata(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}
func (f *Forwarder) SubmitMetadata(payload Payloads, extra http.Header) error {
	return errors.New("NotImplemented")
}
func (f *Forwarder) SubmitProcessChecks(payload Payloads, extra http.Header) (chan Response, error) {
	return nil, errors.New("NotImplemented")
}
func (f *Forwarder) SubmitRTProcessChecks(payload Payloads, extra http.Header) (chan Response, error) {
	return nil, errors.New("NotImplemented")
}
func (f *Forwarder) SubmitContainerChecks(payload Payloads, extra http.Header) (chan Response, error) {
	return nil, errors.New("NotImplemented")
}
func (f *Forwarder) SubmitRTContainerChecks(payload Payloads, extra http.Header) (chan Response, error) {
	return nil, errors.New("NotImplemented")
}
func (f *Forwarder) SubmitConnectionChecks(payload Payloads, extra http.Header) (chan Response, error) {
	return nil, errors.New("NotImplemented")
}
func (f *Forwarder) SubmitPodChecks(payload Payloads, extra http.Header) (chan Response, error) {
	return nil, errors.New("NotImplemented")
}
