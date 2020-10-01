// +build cpython

package py

import (
	"errors"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/aggregator"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	chk "github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/metrics"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/mitchellh/mapstructure"
	"unsafe"

	"github.com/sbinet/go-python"

	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// #cgo pkg-config: python-2.7
// #cgo windows LDFLAGS: -Wl,--allow-multiple-definition
// #include "topology_api.h"
// #include "api.h"
// #include "stdlib.h"
// #include <Python.h>
import "C"

// SubmitContextualizedEvent is the method exposed to Python scripts to submit contextualized events
//export SubmitContextualizedEvent
func SubmitContextualizedEvent(check *C.PyObject, checkID *C.char, event *C.PyObject) *C.PyObject {

	goCheckID := C.GoString(checkID)
	var sender aggregator.Sender
	var err error

	sender, err = aggregator.GetSender(chk.ID(goCheckID))

	if err != nil || sender == nil {
		log.Errorf("Error submitting metric to the Sender: %v", err)
		return C._none()
	}

	if int(C._PyDict_Check(event)) == 0 {
		log.Errorf("Error submitting event to the Sender, the submitted event is not a python dict")
		return C._none()
	}

	_event, err := extractEventFromDict(event, goCheckID)
	if err != nil {
		log.Error(err)
		return nil
	}

	// Extract context
	pyKey := C.CString("context")
	defer C.free(unsafe.Pointer(pyKey))

	context := C.PyDict_GetItemString(event, pyKey) // borrowed ref
	if context != nil {
		if _context, err := extractEventContext(context, checkID); err != nil {
			log.Error(err)
			return nil
		} else {
			_event.EventContext = _context
		}
	}

	sender.Event(_event)

	return C._none()
}


// extractEventContext returns a pointer to a Event Context of the passed non-nil py object.
// The caller needs to check the returned `error`, any non-nil value indicates that the error flag is set
// on the python interpreter.
func extractEventContext(context *C.PyObject, checkID string) (*metrics.EventContext, error) {
	if !isNone(context) {
		goCheckID := C.GoString(checkID)

		eventContext := &metrics.EventContext{}

		//
		if si, err := getStringKey("source_identifier", context); err != nil{
			return nil, err
		} else {
			eventContext.SourceIdentifier = si
		}

		// Extract Element Identifier Slice
		if ed, err := extractStringSlice(goCheckID, "element_identifiers", context); err != nil {
			return nil, err
		} else {
			eventContext.ElementIdentifiers = ed
		}

		// Extract Source string value
		if s, err := getStringKey("source", context); err != nil{
			return nil, err
		} else {
			eventContext.Source = s
		}

		// Extract Category string value
		if c, err := getStringKey("category", context); err != nil{
			return nil, err
		} else {
			eventContext.Category = c
		}

		// Extract Data map
		if d, err := getDictKey(goCheckID, "data", context); err != nil {
			return nil, err
		} else {
			eventContext.Data = d
		}

		// Extract Source Links map
		if sl, err := getDictKey(goCheckID, "source_links", context); err != nil {
			return nil, err
		} else {
			// convert map[string]interface{} to map[string]string
			var decodedSL map[string]string
			err = mapstructure.Decode(sl, decodedSL)
			if err != nil {
				return nil, err
			}
			eventContext.SourceLinks = decodedSL
		}

		return eventContext, nil

	}

	return nil, errors.New("cant' extract event context")
}

// extractStringSlice returns a slice with the contents of the passed non-nil py object.
// The caller needs to check the returned `error`, any non-nil value indicates that the error flag is set
// on the python interpreter.
func extractStringSlice(checkID, key string, object *C.PyObject) (_strings []string, err error) {
	if !isNone(object) {
		if int(C.PySequence_Check(object)) == 0 {
			log.Errorf("Value for key `%s` is not a sequence, ignoring it", key)
			return
		}

		errMsg := C.CString("expected slice to be a sequence")
		defer C.free(unsafe.Pointer(errMsg))

		var seq *C.PyObject
		seq = C.PySequence_Fast(object, errMsg) // seq is a new reference, has to be decref'd
		if seq == nil {
			err = errors.New("can't iterate on slice")
			return
		}
		defer C.Py_DecRef(seq)

		var i C.Py_ssize_t
		for i = 0; i < C.PySequence_Fast_Get_Size(seq); i++ {
			item := C.PySequence_Fast_Get_Item(seq, i) // `item` is borrowed, no need to decref
			if int(C._PyString_Check(item)) == 0 {
				typeName := C.GoString(C._object_type(item))
				stringRepr := stringRepresentation(item)
				log.Infof("One of the submitted values for key `%s` for the check with ID %s is not a string " +
					"but a %s: %s, ignoring it", key, checkID, typeName, stringRepr)
				continue
			}
			// at this point we're sure that `item` is a string, no further error checking needed
			_strings = append(_strings, C.GoString(C.PyString_AsString(item)))
		}
	}

	return
}

func getStringKey(key string, context *C.PyObject) (string, error){
	pyKey := C.CString(key)
	defer C.free(unsafe.Pointer(pyKey))

	pyValue := C.PyDict_GetItemString(context, pyKey) // borrowed ref
	// key not in dict => nil ; value for key is None => None ; we need to check for both
	if pyValue != nil && !isNone(pyValue) {
		if int(C._PyString_Check(pyValue)) != 0 {
			// at this point we're sure that `pyValue` is a string, no further error checking needed
			return C.GoString(C.PyString_AsString(pyValue)), nil
		} else {
			return "", errors.New(
				fmt.Sprintf("Can't parse value for key '%s' in event context submitted from python check", key),
				)
		}
	}
	return "", errors.New(fmt.Sprintf("No value for key '%s' in event context submitted from python check", key))
}

func getDictKey(checkID, key string, context *C.PyObject) (map[string]interface{}, error){
	pyKey := C.CString(key)
	defer C.free(unsafe.Pointer(pyKey))

	data := C.PyDict_GetItemString(context, pyKey) // borrowed ref
	return extractStructureFromObject(data, checkID)
}
