// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com).
// Copyright 2021 StackState
#include "transaction.h"
#include "rtloader_mem.h"
#include "stringutils.h"
#include "util.h"

// these must be set by the Agent
static cb_start_transaction_t cb_start_transaction = NULL;
static cb_stop_transaction_t cb_stop_transaction = NULL;

// forward declarations
static PyObject *start_transaction(PyObject *self, PyObject *args);
static PyObject *stop_transaction(PyObject *self, PyObject *args);

static PyMethodDef methods[] = {
    {"start_transaction", (PyCFunction)start_transaction, METH_VARARGS, "Starts a transactional state for a Agent Check."},
    {"stop_transaction", (PyCFunction)stop_transaction, METH_VARARGS, "Stops a transactional state for a Agent Check."},
    {NULL, NULL}  // guards
};


#ifdef DATADOG_AGENT_THREE
static struct PyModuleDef module_def = { PyModuleDef_HEAD_INIT, TRANSACTION_MODULE_NAME, NULL, -1, methods };

PyMODINIT_FUNC PyInit_transaction(void)
{
    return PyModule_Create(&module_def);
}
#elif defined(DATADOG_AGENT_TWO)
// in Python2 keep the object alive for the program lifetime
static PyObject *module;

void Py2_init_transaction()
{
    module = Py_InitModule(TRANSACTION_MODULE_NAME, methods);
}
#endif


void _set_start_transaction_cb(cb_start_transaction_t cb)
{
    cb_start_transaction = cb;
}

static PyObject *start_transaction(PyObject *self, PyObject *args) {
    if (cb_start_transaction == NULL) {
        PyErr_SetString(PyExc_TypeError, "`start_transaction` is set as NULL");
        Py_RETURN_NONE;
    }

    PyObject *check = NULL; // borrowed
    char *check_id;

    PyGILState_STATE gstate = PyGILState_Ensure();

    if (!PyArg_ParseTuple(args, "Os", &check, &check_id)) {
      goto error;
    }

    cb_start_transaction(check_id);

    PyGILState_Release(gstate);
    Py_RETURN_NONE; // Success

error:
    PyGILState_Release(gstate);
    return NULL; // Failure
}

void _set_stop_transaction_cb(cb_stop_transaction_t cb)
{
    cb_stop_transaction = cb;
}

static PyObject *stop_transaction(PyObject *self, PyObject *args) {
    if (cb_stop_transaction == NULL) {
        PyErr_SetString(PyExc_TypeError, "`stop_transaction` is set as NULL");
        Py_RETURN_NONE;
    }

    PyObject *check = NULL; // borrowed
    char *check_id;

    PyGILState_STATE gstate = PyGILState_Ensure();

    if (!PyArg_ParseTuple(args, "Os", &check, &check_id)) {
      goto error;
    }

    cb_stop_transaction(check_id);

    PyGILState_Release(gstate);
    Py_RETURN_NONE; // Success

error:
    PyGILState_Release(gstate);
    return NULL; // Failure
}
