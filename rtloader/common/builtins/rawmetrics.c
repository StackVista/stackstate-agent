// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com).
// Copyright 2021 StackState
#include "rawmetrics.h"
#include "rtloader_mem.h"
#include "stringutils.h"

// these must be set by the Agent
static cb_submit_raw_metrics_data_t cb_submit_raw_metrics_data = NULL;

// forward declarations
static PyObject *submit_raw_metrics_data(PyObject *self, PyObject *args);

static PyMethodDef methods[] = {
    {"submit_raw_metrics_data", (PyCFunction)submit_raw_metrics_data, METH_VARARGS, "Submit raw metrics data to the intake api."},
    {NULL, NULL}  // guards
};


#ifdef DATADOG_AGENT_THREE
static struct PyModuleDef module_def = { PyModuleDef_HEAD_INIT, RAW_METRICS_MODULE_NAME, NULL, -1, methods };

PyMODINIT_FUNC PyInit_raw_metrics_data(void)
{
    return PyModule_Create(&module_def);
}
#elif defined(DATADOG_AGENT_TWO)
// in Python2 keep the object alive for the program lifetime
static PyObject *module;

void Py2_init_raw_metrics_data()
{
    module = Py_InitModule(RAW_METRICS_MODULE_NAME, methods);
}
#endif


void _set_submit_raw_metrics_data_cb(cb_submit_raw_metrics_data_t cb)
{
    cb_submit_raw_metrics_data = cb;
}

static PyObject *submit_raw_metrics_data(PyObject *self, PyObject *args) {
    if (cb_submit_raw_metrics_data == NULL) {
        Py_RETURN_NONE;
    }

    PyObject *check = NULL; // borrowed
    char *check_id;
    PyObject *event_dict = NULL; // borrowed
    char *raw_metrics_data;

    PyGILState_STATE gstate = PyGILState_Ensure();

    if (!PyArg_ParseTuple(args, "OsO", &check, &check_id, &event_dict)) {
      goto error;
    }

    if (!PyDict_Check(event_dict)) {
        PyErr_SetString(PyExc_TypeError, "raw metrics data must be a dict");
        goto error;
    }

    raw_metrics_data = as_json(event_dict);
    if (raw_metrics_data == NULL) {
        // If as_json fails it sets a python exception, so we just return
        goto error;
    } else {
        cb_submit_raw_metrics_data(check_id, raw_metrics_data);
        _free(raw_metrics_data);
    }

    PyGILState_Release(gstate);
    Py_RETURN_NONE; // Success

error:
    PyGILState_Release(gstate);
    return NULL; // Failure
}
