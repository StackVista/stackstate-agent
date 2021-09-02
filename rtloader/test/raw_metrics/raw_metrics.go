package testrawmetrics

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

extern void submitRawMetricsData(char *, raw_metrics_stream_t *, char *);
extern void submitRawMetricsStartSnapshot(char *, raw_metrics_stream_t *);
extern void submitRawMetricsStopSnapshot(char *, raw_metrics_stream_t *);

static void initRawMetricsTests(rtloader_t *rtloader) {
	set_submit_raw_metrics_data_cb(rtloader, submitRawMetricsData);
	set_submit_raw_metrics_start_snapshot_cb(rtloader, submitRawMetricsStartSnapshot);
	set_submit_raw_metrics_stop_snapshot_cb(rtloader, submitRawMetricsStopSnapshot);
}
*/
import "C"

// TODO: Raw Metrics

var (
	rtloader               *C.rtloader_t
	checkID                string
	_data                  map[string]interface{}
	result                 map[string]interface{}
)

func resetOuputValues() {
	checkID = ""
	_data = nil
	result = nil
}

func setUp() error {
	// Initialize memory tracking
	helpers.InitMemoryTracker()

	rtloader = (*C.rtloader_t)(common.GetRtLoader())
	if rtloader == nil {
		return fmt.Errorf("make failed")
	}

	C.initRawMetricsTests(rtloader)

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
	import raw_metrics
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

//export submitRawMetricsData
func submitRawMetricsData(id *C.char, data *C.char) {
	checkID = C.GoString(id)
	_raw_data := C.GoString(data)
	rawMetricsPayload := &metrics.RawMetricsPayload{}
	json.Unmarshal([]byte(_raw_data), rawMetricsPayload)
	result = rawMetricsPayload.Data
}
