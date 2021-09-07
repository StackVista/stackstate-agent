// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com).
// Copyright 2021 StackState
#include "telemetry.h"
#include "rtloader_mem.h"
#include "stringutils.h"

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

// TODO: Reuse from aggregator.c
/*! \fn py_tag_to_c(PyObject *py_tags)
    \brief A function to convert a list of python strings (tags) into an
    array of C-strings.
    \return a char ** pointer to the C-representation of the provided python
    tag list. In the event of failure NULL is returned.

    The returned char ** string array pointer is heap allocated here and should
    be subsequently freed by the caller. This function may set and raise python
    interpreter errors. The function is static and not in the builtin's API.
*/
static char **py_tag_to_c(PyObject *py_tags)
{
    char **tags = NULL;
    PyObject *py_tags_list = NULL; // new reference

    if (!PySequence_Check(py_tags)) {
        PyErr_SetString(PyExc_TypeError, "tags must be a sequence");
        return NULL;
    }

    int len = PySequence_Length(py_tags);
    if (len == -1) {
        PyErr_SetString(PyExc_RuntimeError, "could not compute tags length");
        return NULL;
    } else if (len == 0) {
        if (!(tags = _malloc(sizeof(*tags)))) {
            PyErr_SetString(PyExc_RuntimeError, "could not allocate memory for tags");
            return NULL;
        }
        tags[0] = NULL;
        return tags;
    }

    py_tags_list = PySequence_Fast(py_tags, "py_tags is not a sequence"); // new reference
    if (py_tags_list == NULL) {
        goto done;
    }

    if (!(tags = _malloc(sizeof(*tags) * (len + 1)))) {
        PyErr_SetString(PyExc_RuntimeError, "could not allocate memory for tags");
        goto done;
    }
    int nb_valid_tag = 0;
    int i;
    for (i = 0; i < len; i++) {
        // `item` is borrowed, no need to decref
        PyObject *item = PySequence_Fast_GET_ITEM(py_tags_list, i);

        char *ctag = as_string(item);
        if (ctag == NULL) {
            continue;
        }
        tags[nb_valid_tag] = ctag;
        nb_valid_tag++;
    }
    tags[nb_valid_tag] = NULL;

done:
    Py_XDECREF(py_tags_list);
    return tags;
}

// TODO: Reuse from aggregator.c
/*! \fn free_tags(char **tags)
    \brief A helper function to free the memory allocated by the py_tag_to_c() function.

    This function is for internal use and expects the tag array to be properly intialized,
    and have a NULL canary at the end of the array, just like py_tag_to_c() initializes and
    populates the array. Be mindful if using this function in any other context.
*/
static void free_tags(char **tags)
{
    int i;
    for (i = 0; tags[i] != NULL; i++) {
        _free(tags[i]);
    }
    _free(tags);
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
        PyErr_SetString(PyExc_TypeError, "`cb_submit_raw_metrics_data` is not defined");
        Py_RETURN_NONE;
    }

    PyObject *check = NULL; // borrowed
    PyObject *py_tags = NULL; // borrowed
    char *json_data = NULL;
    char *name = NULL;
    char *hostname = NULL;
    char *timestamp = NULL;
    char *check_id = NULL;
    char **tags = NULL;
    float value;
    PyObject * data = NULL;
    PyObject * retval = NULL;

    PyGILState_STATE gstate = PyGILState_Ensure();

    if (!PyArg_ParseTuple(args, "OssfOsi", &check, &check_id, &name, &value, &py_tags, &hostname, &timestamp)) {
        retval = NULL; // Failure
        PyErr_SetString(PyExc_TypeError, "Raw metrics, Unable to parse arguments passed to `submit_raw_metrics_data`");
        goto done;
    }

    if ((tags = py_tag_to_c(py_tags)) == NULL)
        retval = NULL; // Failure
        PyErr_SetString(PyExc_TypeError, "Unable to set raw metric tags");
        goto done;

    data = Py_BuildValue("{s:s, s:f, s:O, s:s, s:i}", "name", name, "value", value, "tags", tags, "hostname", hostname, "timestamp", timestamp);
    json_data = as_json(data);

    if (json_data == NULL) {
        // If as_json fails it sets a python exception, so we just return
        retval = NULL; // Failure
        PyErr_SetString(PyExc_TypeError, "Unable to create a raw metric JSON data");
        goto done;
    } else {
        cb_submit_raw_metrics_data(check_id, json_data);
        Py_INCREF(Py_None); // Increment, since we are not using the macro Py_RETURN_NONE that does it for us
        retval = Py_None; // Success
    }

done:
    if (json_data != NULL) {
        _free(json_data);
    }
    free_tags(tags);
    PyGILState_Release(gstate);
    return retval;
}
