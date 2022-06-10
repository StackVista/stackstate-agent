//go:build python && test
// +build python,test

package python

import "testing"

func TestStartTransaction(t *testing.T) {
	testStartTransaction(t)
}

func TestStopTransaction(t *testing.T) {
	testStopTransaction(t)
}

func TestSetTransactionState(t *testing.T) {
	testSetTransactionState(t)
}
