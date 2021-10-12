// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com).
// Copyright 2021 StackState
#include "telemetry.h"
#include "rtloader_mem.h"
#include "stringutils.h"
#include "util.h"

// these must be set by the Agent
static cb_submit_topology_event_t cb_submit_topology_event = NULL;
static cb_submit_raw_metrics_data_t cb_submit_raw_metrics_data = NULL;

// forward declarations
static PyObject *submit_topology_event(PyObject *self, PyObject *args);
static PyObject *submit_raw_metrics_data(PyObject *self, PyObject *args);

static PyMethodDef methods[] = {
    {"submit_topology_event", (PyCFunction)submit_topology_event, METH_VARARGS, "Submit a topology event to the intake api."},
    {"submit_raw_metrics_data", (PyCFunction)submit_raw_metrics_data, METH_VARARGS, "Submit raw metrics data to the raw metrics api."},
    {NULL, NULL}  // guards
};


#ifdef DATADOG_AGENT_THREE
static struct PyModuleDef module_def = { PyModuleDef_HEAD_INIT, TELEMETRY_MODULE_NAME, NULL, -1, methods };

PyMODINIT_FUNC PyInit_telemetry(void)
{
    return PyModule_Create(&module_def);
}
#elif defined(DATADOG_AGENT_TWO)
// in Python2 keep the object alive for the program lifetime
static PyObject *module;

void Py2_init_telemetry()
{
    module = Py_InitModule(TELEMETRY_MODULE_NAME, methods);
}
#endif


void _set_submit_topology_event_cb(cb_submit_topology_event_t cb)
{
    cb_submit_topology_event = cb;
}

static PyObject *submit_topology_event(PyObject *self, PyObject *args) {
    if (cb_submit_topology_event == NULL) {
        PyErr_SetString(PyExc_TypeError, "`submit_topology_event` is set as NULL");
        Py_RETURN_NONE;
    }

    PyObject *check = NULL; // borrowed
    char *check_id;
    PyObject *event_dict = NULL; // borrowed
    char *topology_event;

    PyGILState_STATE gstate = PyGILState_Ensure();

    if (!PyArg_ParseTuple(args, "OsO", &check, &check_id, &event_dict)) {
      goto error;
    }

    if (!PyDict_Check(event_dict)) {
        PyErr_SetString(PyExc_TypeError, "topology event must be a dict");
        goto error;
    }

    topology_event = as_json(event_dict);
    if (topology_event == NULL) {
        // If as_json fails it sets a python exception, so we just return
        goto error;
    } else {
        cb_submit_topology_event(check_id, topology_event);
        _free(topology_event);
    }

    PyGILState_Release(gstate);
    Py_RETURN_NONE; // Success

error:
    PyGILState_Release(gstate);
    return NULL; // Failure
}

void _set_submit_raw_metrics_data_cb(cb_submit_raw_metrics_data_t cb)
{
    cb_submit_raw_metrics_data = cb;
}

static PyObject *submit_raw_metrics_data(PyObject *self, PyObject *args) {
    if (cb_submit_raw_metrics_data == NULL) {
        PyErr_SetString(PyExc_TypeError, "`cb_submit_raw_metrics_data` is not defined");
        Py_RETURN_NONE;
    }

    PyGILState_STATE gstate = PyGILState_Ensure();

    // Arguments passed to `submit_raw_metrics_data` *args
    PyObject *check = NULL; // borrowed
    PyObject *py_tags = NULL; // borrowed
    char *check_id = NULL;
    char *name = NULL;
    char **tags = NULL;
    char *hostname = NULL;
    float value;
    long long timestamp;

    // Python call: telemetry.submit_raw_metrics_data(self, check_id, name, value, tags, hostname, timestamp)
    if (!PyArg_ParseTuple(args, "OssfOsl", &check, &check_id, &name, &value, &py_tags, &hostname, &timestamp)) {
        PyErr_SetObject(PyExc_TypeError, args);
        // PyErr_SetString(PyExc_TypeError, "Raw metrics, Unable to parse arguments passed to `submit_raw_metrics_data`");
        goto error;
    }

    if ((tags = py_tag_to_c(py_tags)) == NULL)
        goto error;

    cb_submit_raw_metrics_data(check_id, name, value, tags, hostname, timestamp);

    free_tags(tags);

    PyGILState_Release(gstate);
    Py_RETURN_NONE;

error:
    PyGILState_Release(gstate);
    return NULL;
}
