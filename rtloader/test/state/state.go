package teststate

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"unsafe"

	common "github.com/DataDog/datadog-agent/rtloader/test/common"
	"github.com/DataDog/datadog-agent/rtloader/test/helpers"
)

/*
#include "rtloader_mem.h"
#include "datadog_agent_rtloader.h"

extern void setState(char *, char *, char *);
extern char* getState(char *, char *);

static void initStateTests(rtloader_t *rtloader) {
	set_state_cb(rtloader, setState);
	set_get_state_cb(rtloader, getState);
}
*/
import "C"

var (
	rtloader           *C.rtloader_t
	checkID            string
	lastRetrievedState string
	stateStorage       map[string]string
)

func resetOutputValues() {
	checkID = ""
	lastRetrievedState = ""
	stateStorage = map[string]string{}
}

func setUp() error {
	// Initialize memory tracking
	helpers.InitMemoryTracker()

	rtloader = (*C.rtloader_t)(common.GetRtLoader())
	if rtloader == nil {
		return fmt.Errorf("make failed")
	}

	C.initStateTests(rtloader)

	// Updates sys.path so testing Check can be found
	C.add_python_path(rtloader, C.CString("../python"))

	if ok := C.init(rtloader); ok != 1 {
		return fmt.Errorf("`init` failed: %s", C.GoString(C.get_error(rtloader)))
	}

	return nil
}

func run(call string) (string, error) {
	resetOutputValues()
	tmpfile, err := ioutil.TempFile("", "testout")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	code := (*C.char)(helpers.TrackedCString(fmt.Sprintf(`
try:
	import state
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

//export setState
func setState(id *C.char, key *C.char, state *C.char) {
	checkID = C.GoString(id)
	stateKey := C.GoString(key)
	stateValue := C.GoString(state)

	stateStorage[stateKey] = stateValue
}

//export getState
func getState(id *C.char, key *C.char) *C.char {
	checkID = C.GoString(id)
	stateKey := C.GoString(key)

	var retrievedState = "{}"

	if val, ok := stateStorage[stateKey]; ok {
		retrievedState = val
	}

	lastRetrievedState = retrievedState

	return C.CString(retrievedState)
}
