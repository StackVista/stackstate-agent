//go:build python && test
// +build python,test

package python

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/stretchr/testify/assert"
	"testing"
)

// #include <datadog_agent_rtloader.h>
import "C"

func setAndGetState(checkId string, key string, state string) string {
	SetState(checkId, key, state)
	retrievedState := GetState(checkId, key)
	return retrievedState
}

func testSetState(t *testing.T) {
	testCheck := &check.STSTestCheck{Name: "check-id-set-state"}
	checkId := C.CString(string(testCheck.ID()))
	state := setAndGetState(checkId, "state-id", "state-body")
	assert.Equal(t, state, "random-text")
}

func testGetState(t *testing.T) {
	testCheck := &check.STSTestCheck{Name: "check-id-set-state"}
	checkId := C.CString(string(testCheck.ID()))
	state := setAndGetState(checkId, "state-id", "state-body")
	assert.Equal(t, state, "random-text")
}
