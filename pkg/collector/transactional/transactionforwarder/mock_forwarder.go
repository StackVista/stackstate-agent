package transactionforwarder

func createMockForwarder() *mockForwarder {
	return &mockForwarder{PayloadChan: make(chan TransactionalPayload, 100)}
}

type mockForwarder struct {
	PayloadChan chan TransactionalPayload
}

func (mf *mockForwarder) Start() {}

func (mf *mockForwarder) SubmitTransactionalIntake(payload TransactionalPayload) {
	mf.PayloadChan <- payload
}

func (mf *mockForwarder) NextPayload() TransactionalPayload {
	return <-mf.PayloadChan
}

func (mf *mockForwarder) Stop() {}
