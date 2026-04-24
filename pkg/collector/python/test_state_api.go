//go:build python && test

package python

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/comp/core/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/check/handler"
	"github.com/DataDog/datadog-agent/pkg/collector/check/test"
	configmock "github.com/DataDog/datadog-agent/pkg/config/mock"
	pkgconfigsetup "github.com/DataDog/datadog-agent/pkg/config/setup"
	"github.com/stretchr/testify/assert"
)

// #include <datadog_agent_rtloader.h>
import "C"

func testSetAndGetState(t *testing.T) {
	path := pkgconfigsetup.Datadog().Get("run_path").(string)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	defer os.RemoveAll(path)

	mockConfig := configmock.New(t)

	// Set the run path to the temp directory above, this will allow the persistent cache to have a folder to write into
	// Without doing the above persistent cache will generate a folder does not exist error
	mockConfig.SetWithoutSource("run_path", path)
	mockConfig.SetWithoutSource("run_path", path)

	_, _, _, checkManager := handler.SetupMockTransactionalComponents()

	release := scopeInitCheckManager(checkManager)
	defer release()

	testCheck := &test.STSTestCheck{
		Name: "check-id-set-state",
	}

	checkManager.RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

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
