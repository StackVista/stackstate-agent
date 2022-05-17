package transactionforwarder

func createMockForwarder() *mockForwarder {
	return &mockForwarder{PayloadChan: make(chan TransactionalPayload, 100)}
}

type mockForwarder struct {
	PayloadChan chan TransactionalPayload
}

// Start is a noop
func (mf *mockForwarder) Start() {}

// SubmitTransactionalIntake receives a TransactionalPayload and keeps it in the PayloadChan to be used in assertions
func (mf *mockForwarder) SubmitTransactionalIntake(payload TransactionalPayload) {
	mf.PayloadChan <- payload
}

// NextPayload returns the next payload in the PayloadChan
func (mf *mockForwarder) NextPayload() TransactionalPayload {
	return <-mf.PayloadChan
}

// Stop is a noop
func (mf *mockForwarder) Stop() {}
