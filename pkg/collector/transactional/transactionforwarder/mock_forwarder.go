package transactionforwarder

func createMockForwarder() *MockTransactionalForwarder {
	return &MockTransactionalForwarder{PayloadChan: make(chan TransactionalPayload, 100)}
}

// MockTransactionalForwarder is a mock implementation of the transactional forwarder
type MockTransactionalForwarder struct {
	PayloadChan chan TransactionalPayload
}

// Start is a noop
func (mf *MockTransactionalForwarder) Start() {}

// SubmitTransactionalIntake receives a TransactionalPayload and keeps it in the PayloadChan to be used in assertions
func (mf *MockTransactionalForwarder) SubmitTransactionalIntake(payload TransactionalPayload) {
	mf.PayloadChan <- payload
}

// NextPayload returns the next payload in the PayloadChan
func (mf *MockTransactionalForwarder) NextPayload() TransactionalPayload {
	return <-mf.PayloadChan
}

// Stop is a noop
func (mf *MockTransactionalForwarder) Stop() {
	close(mf.PayloadChan)
}
