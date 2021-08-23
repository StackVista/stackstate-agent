// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com).
// Copyright 2021 StackState
#include "raw_metrics.h"
#include "rtloader_mem.h"
#include "stringutils.h"

// TODO: Raw Metrics

// these must be set by the Agent
static cb_submit_raw_metrics_data_t cb_submit_raw_metrics_data = NULL;
static cb_submit_raw_metrics_start_snapshot_t cb_submit_raw_metrics_start_snapshot = NULL;
static cb_submit_raw_metrics_stop_snapshot_t cb_submit_raw_metrics_stop_snapshot = NULL;

// forward declarations
static PyObject *submit_raw_metrics_data(PyObject *self, PyObject *args);
static PyObject *submit_raw_metrics_start_snapshot(PyObject *self, PyObject *args);
static PyObject *submit_raw_metrics_stop_snapshot(PyObject *self, PyObject *args);

static PyMethodDef methods[] = {
    {"submit_raw_metrics_data", (PyCFunction)submit_raw_metrics_data, METH_VARARGS, "Submit raw metrics data to the raw metrics api."},
    {"submit_raw_metrics_start_snapshot", (PyCFunction)submit_raw_metrics_start_snapshot, METH_VARARGS, "Submit a raw metrics snapshot start to the raw metrics api."},
    {"submit_raw_metrics_stop_snapshot", (PyCFunction)submit_raw_metrics_stop_snapshot, METH_VARARGS, "Submit a raw metrics snapshot stop to the raw metrics api."},
    {NULL, NULL}  // guards
};


#ifdef DATADOG_AGENT_THREE
static struct PyModuleDef module_def = { PyModuleDef_HEAD_INIT, RAW_METRICS_MODULE_NAME, NULL, -1, methods };

PyMODINIT_FUNC PyInit_raw_metrics(void)
{
    return PyModule_Create(&module_def);
}
#elif defined(DATADOG_AGENT_TWO)
// in Python2 keep the object alive for the program lifetime
static PyObject *module;

void Py2_init_raw_metrics()
{
    module = Py_InitModule(RAW_METRICS_MODULE_NAME, methods);
}
#endif


void _set_submit_raw_metrics_data_cb(cb_submit_raw_metrics_data_t cb)
{
    cb_submit_raw_metrics_data = cb;
}

void _set_submit_raw_metrics_start_snapshot_cb(cb_submit_raw_metrics_start_snapshot_t cb)
{
    cb_submit_raw_metrics_start_snapshot = cb;
}

void _set_submit_raw_metrics_stop_snapshot_cb(cb_submit_raw_metrics_stop_snapshot_t cb)
{
    cb_submit_raw_metrics_stop_snapshot = cb;
}


/*! \fn submit_raw_metrics_check_data(PyObject *self, PyObject *args)
    \brief Raw metrics builtin class method for raw metrics data submission.
    \param self A PyObject * pointer to self - the raw metrics module.
    \param args A PyObject * pointer to the python args or kwargs.
    \return This function returns a new reference to None (already INCREF'd), or NULL in case of error.

    This function implements the `submit_raw_metrics_data` python callable in C and is used from the python code.
    More specifically, in the context of rtloader and datadog-agent, this is called from our python base check
    class to submit raw metrics data to the batcher.
*/
static PyObject *submit_raw_metrics_data(PyObject *self, PyObject *args) {
    if (cb_submit_raw_metrics_data == NULL) {
        Py_RETURN_NONE;
    }

    PyObject *check = NULL; // borrowed
    char *check_id;
    PyObject *raw_metrics_stream_dict = NULL; // borrowed
    PyObject *data_dict = NULL; // borrowed
    char *urn = NULL;
    char *sub_stream = NULL;
    raw_metrics_stream_t *raw_metrics_stream_key = NULL;
    char *json_data = NULL;
    PyObject * stream = NULL;
    PyObject * raw_metrics = NULL;
    PyObject * retval = NULL;

    PyGILState_STATE gstate = PyGILState_Ensure();

    if (!PyArg_ParseTuple(args, "OsOO", &check, &check_id, &raw_metrics_stream_dict, &data_dict)) {
        retval = NULL; // Failure
        goto done;
    }

    if (!PyDict_Check(raw_metrics_stream_dict)) {
        PyErr_SetString(PyExc_TypeError, "raw metrics stream must be a dict");
        retval = NULL; // Failure
        goto done;
    }

    if (!PyDict_Check(data_dict)) {
        PyErr_SetString(PyExc_TypeError, "raw metrics data must be a dict");
        retval = NULL; // Failure
        goto done;
    }

    if (!(raw_metrics_stream_key = (raw_metrics_stream_t *)_malloc(sizeof(raw_metrics_stream_t)))) {
        PyErr_SetString(PyExc_RuntimeError, "could not allocate memory for raw metrics stream key");
        retval = NULL; // Failure
        goto done;
    }

    // notice: PyDict_GetItemString returns a borrowed ref or NULL if key was not found
    urn = as_string(PyDict_GetItemString(raw_metrics_stream_dict, "urn"));
    sub_stream = as_string(PyDict_GetItemString(raw_metrics_stream_dict, "sub_stream"));
    raw_metrics_stream_key->urn = urn;
    raw_metrics_stream_key->sub_stream = sub_stream;

    stream = Py_BuildValue("{s:s, s:s}", "urn", urn, "sub_stream", sub_stream);
    raw_metrics = Py_BuildValue("{s:O, s:O}", "stream", stream, "data", data_dict);
    json_data = as_json(raw_metrics);
    if (json_data == NULL) {
        // If as_json fails it sets a python exception, so we just return
        retval = NULL; // Failure
        goto done;
    } else {
        cb_submit_raw_metrics_data(check_id, raw_metrics_stream_key, json_data);

        Py_INCREF(Py_None); // Increment, since we are not using the macro Py_RETURN_NONE that does it for us
        retval = Py_None; // Success
    }

done:
    if (raw_metrics_stream_key != NULL) {
        _free(raw_metrics_stream_key->urn);
        _free(raw_metrics_stream_key->sub_stream);
        _free(raw_metrics_stream_key);
    }
    if (json_data != NULL) {
        _free(json_data);
    }
    Py_XDECREF(stream);
    Py_XDECREF(raw_metrics);
    PyGILState_Release(gstate);
    return retval;
}

/*! \fn submit_raw_metrics_start_snapshot(PyObject *self, PyObject *args)
    \brief Raw metrics builtin class method to signal the start of a raw metrics snapshot submission.
    \param self A PyObject * pointer to self - the raw metrics module.
    \param args A PyObject * pointer to the python args or kwargs.
    \return This function returns a new reference to None (already INCREF'd), or NULL in case of error.

    This function implements the `submit_raw_metrics_start_snapshot` python callable in C and is used from the python code.
    More specifically, in the context of rtloader and datadog-agent, this is called from our python base check
    class to submit the start of raw metrics snapshot collection to the batcher.
*/
static PyObject *submit_raw_metrics_start_snapshot(PyObject *self, PyObject *args) {
    if (cb_submit_raw_metrics_start_snapshot == NULL) {
        Py_RETURN_NONE;
    }

    PyObject *check = NULL; // borrowed
    char *check_id;
    PyObject *raw_metrics_stream_dict = NULL; // borrowed
    raw_metrics_stream_t *raw_metrics_stream_key = NULL;

    PyGILState_STATE gstate = PyGILState_Ensure();

    if (!PyArg_ParseTuple(args, "OsOii", &check, &check_id, &raw_metrics_stream_dict)) {
      goto error;
    }

    if (!PyDict_Check(raw_metrics_stream_dict)) {
        PyErr_SetString(PyExc_TypeError, "raw metrics stream must be a dict");
        goto error;
    }

    if (!(raw_metrics_stream_key = (raw_metrics_stream_t *)_malloc(sizeof(raw_metrics_stream_t)))) {
        PyErr_SetString(PyExc_RuntimeError, "could not allocate memory for raw metrics stream key");
        goto error;
    }

    // notice: PyDict_GetItemString returns a borrowed ref or NULL if key was not found
    raw_metrics_stream_key->urn = as_string(PyDict_GetItemString(raw_metrics_stream_dict, "urn"));
    raw_metrics_stream_key->sub_stream = as_string(PyDict_GetItemString(raw_metrics_stream_dict, "sub_stream"));

    cb_submit_raw_metrics_start_snapshot(check_id, raw_metrics_stream_key);

    _free(raw_metrics_stream_key->urn);
    _free(raw_metrics_stream_key->sub_stream);
    _free(raw_metrics_stream_key);

    PyGILState_Release(gstate);
    Py_RETURN_NONE; // Success

error:
    PyGILState_Release(gstate);
    return NULL; // Failure
}

/*! \fn submit_raw_metrics_stop_snapshot(PyObject *self, PyObject *args)
    \brief Raw metrics builtin class method to signal the stop of raw metrics snapshot submission.
    \param self A PyObject * pointer to self - the raw metrics module.
    \param args A PyObject * pointer to the python args or kwargs.
    \return This function returns a new reference to None (already INCREF'd), or NULL in case of error.

    This function implements the `submit_raw_metrics_stop_snapshot` python callable in C and is used from the python code.
    More specifically, in the context of rtloader and datadog-agent, this is called from our python base check
    class to submit the stop of raw metrics collection to the batcher.
*/
static PyObject *submit_raw_metrics_stop_snapshot(PyObject *self, PyObject *args) {
    if (cb_submit_raw_metrics_stop_snapshot == NULL) {
        Py_RETURN_NONE;
    }

    PyObject *check = NULL; // borrowed
    char *check_id;
    PyObject *raw_metrics_stream_dict = NULL; // borrowed
    raw_metrics_stream_t *raw_metrics_stream_key = NULL;

    PyGILState_STATE gstate = PyGILState_Ensure();

    if (!PyArg_ParseTuple(args, "OsO", &check, &check_id, &raw_metrics_stream_dict)) {
      goto error;
    }

    if (!PyDict_Check(raw_metrics_stream_dict)) {
       PyErr_SetString(PyExc_TypeError, "raw metrics stream must be a dict");
       goto error;
   }

   if (!(raw_metrics_stream_key = (raw_metrics_stream_t *)_malloc(sizeof(raw_metrics_stream_t)))) {
       PyErr_SetString(PyExc_RuntimeError, "could not allocate memory for raw metrics stream key");
       goto error;
   }

   // notice: PyDict_GetItemString returns a borrowed ref or NULL if key was not found
   raw_metrics_stream_key->urn = as_string(PyDict_GetItemString(raw_metrics_stream_dict, "urn"));
   raw_metrics_stream_key->sub_stream = as_string(PyDict_GetItemString(raw_metrics_stream_dict, "sub_stream"));

   cb_submit_raw_metrics_stop_snapshot(check_id, raw_metrics_stream_key);

   _free(raw_metrics_stream_key->urn);
   _free(raw_metrics_stream_key->sub_stream);
   _free(raw_metrics_stream_key);

   PyGILState_Release(gstate);
   Py_RETURN_NONE; // Success

error:
    PyGILState_Release(gstate);
    return NULL; // Failure
}
