package testtransaction

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"unsafe"

	common "github.com/StackVista/stackstate-agent/rtloader/test/common"
	"github.com/StackVista/stackstate-agent/rtloader/test/helpers"
)

/*
#include "rtloader_mem.h"
#include "datadog_agent_rtloader.h"

extern void startTransaction(char *);
extern void stopTransaction(char *);
extern void discardTransaction(char *, char *);
extern void setTransactionState(char *, char *, char *);

static void initTransactionTests(rtloader_t *rtloader) {
	set_start_transaction_cb(rtloader, startTransaction);
	set_stop_transaction_cb(rtloader, stopTransaction);
	set_discard_transaction_cb(rtloader, discardTransaction);
	set_transaction_state_cb(rtloader, setTransactionState);
}
*/
import "C"

type TransactionState struct {
	key   string
	value string
}

var (
	rtloader                      *C.rtloader_t
	checkID                       string
	transactionID                 string
	transactionInstanceBatchState transactionbatcher.TransactionCheckInstanceBatchState
	transactionStarted            bool
	transactionCompleted          bool
	transactionDiscardReason      string
	transactionState              TransactionState
)

func resetOuputValues() {
	checkID = ""
	transactionID = ""
	transactionState = TransactionState{}
	transactionInstanceBatchState = transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: &transactionbatcher.BatchTransaction{},
		Topology:    &topology.Topology{},
		Metrics:     &telemetry.Metrics{},
		Health:      map[string]health.Health{},
	}
	transactionStarted = false
	transactionCompleted = false
	transactionDiscardReason = ""
}

func setUp() error {
	// Initialize memory tracking
	helpers.InitMemoryTracker()

	rtloader = (*C.rtloader_t)(common.GetRtLoader())
	if rtloader == nil {
		return fmt.Errorf("make failed")
	}

	C.initTransactionTests(rtloader)

	// Updates sys.path so testing Check can be found
	C.add_python_path(rtloader, C.CString("../python"))

	if ok := C.init(rtloader); ok != 1 {
		return fmt.Errorf("`init` failed: %s", C.GoString(C.get_error(rtloader)))
	}

	return nil
}

func run(call string) (string, error) {
	resetOuputValues()
	tmpfile, err := ioutil.TempFile("", "testout")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	code := (*C.char)(helpers.TrackedCString(fmt.Sprintf(`
try:
	import transaction
	%s
except Exception as e:
	with open(r'%s', 'w') as f:
		f.write("{}: {}\n".format(type(e).__name__, e))
`, call, tmpfile.Name())))
	defer C._free(unsafe.Pointer(code))

	runtime.LockOSThread()
	state := C.ensure_gil(rtloader)

	ret := C.run_simple_string(rtloader, code) == 1

	C.release_gil(rtloader, state)
	runtime.UnlockOSThread()

	if !ret {
		return "", fmt.Errorf("`run_simple_string` errored")
	}

	var output []byte
	output, err = ioutil.ReadFile(tmpfile.Name())

	return strings.TrimSpace(string(output)), err
}

//export startTransaction
func startTransaction(id *C.char) {
	checkID = C.GoString(id)
	transactionID = checkID + "-transaction-id"
	transactionStarted = true
}

//export stopTransaction
func stopTransaction(id *C.char) {
	checkID = C.GoString(id)
	transactionID = ""
	transactionCompleted = true
}

//export discardTransaction
func discardTransaction(id *C.char, reason *C.char) {
	checkID = C.GoString(id)
	discardReason := C.GoString(reason)

	transactionID = ""
	transactionCompleted = true
	transactionDiscardReason = discardReason
}

//export setTransactionState
func setTransactionState(id *C.char, key *C.char, state *C.char) {
	checkID = C.GoString(id)
	keyValue := C.GoString(key)
	stateValue := C.GoString(state)

	transactionState = TransactionState{
		key:   keyValue,
		value: stateValue,
	}
}
