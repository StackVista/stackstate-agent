//go:build python && test
// +build python,test

package python

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/handler"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

// #include <datadog_agent_rtloader.h>
import "C"

func testSetAndGetState(t *testing.T) {
	path := config.Datadog.Get("check_state_root_path").(string)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	defer os.RemoveAll(path)

	mockConfig := config.Mock()

	// Set the run path to the temp directory above, this will allow the persistent cache to have a folder to write into
	// Without doing the above persistent cache will generate a folder does not exist error
	mockConfig.Set("run_path", path)
	mockConfig.Set("check_state_root_path", path)

	SetupTransactionalComponents()
	testCheck := &check.STSTestCheck{
		Name: "check-id-set-state",
	}

	handler.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	checkId := C.CString(string(testCheck.ID()))
	stateKey := C.CString("random-state-id")

	// Calling get state on nothing to make sure a empty works
	GetState(checkId, stateKey)

	SetState(checkId, stateKey, C.CString("{}"))
	getFirstState := GetState(checkId, stateKey)
	assert.Equal(t, C.CString("{}"), getFirstState)

	time.Sleep(5 * time.Second)

	SetState(checkId, stateKey, C.CString("{\"persistent_counter\": 1}"))
	getSecondState := GetState(checkId, stateKey)
	assert.Equal(t, C.CString("{\"persistent_counter\": 1}"), getSecondState)

	time.Sleep(5 * time.Second)

	SetState(checkId, stateKey, C.CString("{\"persistent_counter\": 2}"))
	getThirdState := GetState(checkId, stateKey)
	assert.Equal(t, C.CString("{\"persistent_counter\": 2}"), getThirdState)
}
