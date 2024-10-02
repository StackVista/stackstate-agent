//go:build python && test

package python

import "testing"

func TestStartTransaction(t *testing.T) {
	testStartTransaction(t)
}

func TestStopTransaction(t *testing.T) {
	testStopTransaction(t)
}

func TestDiscardTransaction(t *testing.T) {
	testDiscardTransaction(t)
}

func TestSetTransactionState(t *testing.T) {
	testSetTransactionState(t)
}
