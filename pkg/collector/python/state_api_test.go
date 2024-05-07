//go:build python && test
// +build python,test

package python

import (
	"testing"
)

func TestSetAndGetState(t *testing.T) {
	testSetAndGetState(t)
}
