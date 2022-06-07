// +build python,test

package python

import "testing"

func TestSetState(t *testing.T) {
	testSetState(t)
}

func TestGetState(t *testing.T) {
	testGetState(t)
}
