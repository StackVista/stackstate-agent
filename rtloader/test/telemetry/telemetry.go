package testtelemetry

import (
	"encoding/json"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/metrics"
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

extern void submitTopologyEvent(char *, char *);
extern void submitRawMetricsData(char *, char *, float, char **, char *, long long);

static void initTelemetryTests(rtloader_t *rtloader) {
	set_submit_topology_event_cb(rtloader, submitTopologyEvent);
	set_submit_raw_metrics_data_cb(rtloader, submitRawMetricsData);
}
*/
import "C"

var (
	rtloader 	*C.rtloader_t
	checkID  	string
	_data    	map[string]interface{}
	_topoEvt 	metrics.Event
	rawName		string
	rawHostname  string
	rawValue 	float64
	rawTags      []string
	rawTimestamp int64
)

func resetOuputValues() {
	checkID = ""
	_data = nil
	_topoEvt = metrics.Event{}
	rawName = ""
	rawHostname = ""
	rawValue = 0
	rawTags = nil
	rawTimestamp = 0
}

func setUp() error {
	// Initialize memory tracking
	helpers.InitMemoryTracker()

	rtloader = (*C.rtloader_t)(common.GetRtLoader())
	if rtloader == nil {
		return fmt.Errorf("make failed")
	}

	C.initTelemetryTests(rtloader)

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
	import telemetry
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

func charArrayToSlice(array **C.char) (res []string) {
	pTags := uintptr(unsafe.Pointer(array))
	ptrSize := unsafe.Sizeof(*array)

	for i := uintptr(0); ; i++ {
		tagPtr := *(**C.char)(unsafe.Pointer(pTags + ptrSize*i))
		if tagPtr == nil {
			return
		}
		tag := C.GoString(tagPtr)
		res = append(res, tag)
	}
}

//export submitTopologyEvent
func submitTopologyEvent(id *C.char, data *C.char) {
	checkID = C.GoString(id)
	result := C.GoString(data)
	json.Unmarshal([]byte(result), &_topoEvt)
}

//export submitRawMetricsData
func submitRawMetricsData(id *C.char, name *C.char, value C.float, tags **C.char, hostname *C.char, timestamp C.longlong) {
	checkID = C.GoString(id)
	rawName = C.GoString(name)
	rawHostname = C.GoString(hostname)
	rawValue = float64(value)
	rawTimestamp = int64(timestamp)
	if tags != nil {
		rawTags = append(rawTags, charArrayToSlice(tags)...)
	}
}
