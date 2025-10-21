// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com).
// Copyright 2021 StackState
#include "state.h"
#include "rtloader_mem.h"
#include "stringutils.h"
#include "util.h"

// these must be set by the Agent
static cb_set_state_t cb_set_state = NULL;
static cb_get_state_t cb_get_state = NULL;

// forward declarations
static PyObject *set_state(PyObject *self, PyObject *args);
static PyObject *get_state(PyObject *self, PyObject *args);

static PyMethodDef methods[] = {
    {"set_state", (PyCFunction)set_state, METH_VARARGS, "Sets state for a Agent Check."},
    {"get_state", (PyCFunction)get_state, METH_VARARGS, "Gets state a Agent Check."},
    {NULL, NULL}  // guards
};


#ifdef DATADOG_AGENT_THREE
static struct PyModuleDef module_def = { PyModuleDef_HEAD_INIT, STATE_MODULE_NAME, NULL, -1, methods };

PyMODINIT_FUNC PyInit_state(void)
{
    return PyModule_Create(&module_def);
}
#elif defined(DATADOG_AGENT_TWO)
// in Python2 keep the object alive for the program lifetime
static PyObject *module;

void Py2_init_state()
{
    module = Py_InitModule(STATE_MODULE_NAME, methods);
}
#endif


void _set_state_cb(cb_set_state_t cb)
{
    cb_set_state = cb;
}

void _set_get_state_cb(cb_get_state_t cb)
{
    cb_get_state = cb;
}


/*! \fn set_state(PyObject *self, PyObject *args)
    \brief Aggregator builtin class method for topology component submission.
    \param self A PyObject * pointer to self - the aggregator module.
    \param args A PyObject * pointer to the python args or kwargs.
    \return This function returns a new reference to None (already INCREF'd), or NULL in case of error.

    This function implements the `set_state` python callable in C and is used from the python code.
    More specifically, in the context of rtloader and datadog-agent, this is called from our python base check
    class to set a state for a agent check.
*/
static PyObject *set_state(PyObject *self, PyObject *args) {
    if (cb_set_state == NULL) {
        PyErr_SetString(PyExc_TypeError, "`set_state` is set as NULL");
        Py_RETURN_NONE;
    }

    PyObject *check = NULL; // borrowed
    char *check_id;
    char *key;
    char *state;

    PyGILState_STATE gstate = PyGILState_Ensure();

    if (!PyArg_ParseTuple(args, "Osss", &check, &check_id, &key, &state)) {
      goto error;
    }
    else if (check == NULL) {
        PyErr_SetString(PyExc_TypeError, "Invalid C set state check parameter");
        goto error;
    }
    else if (check_id == NULL) {
        PyErr_SetString(PyExc_TypeError, "Invalid C set state check id parameter");
        goto error;
    }
    else if (key == NULL) {
        PyErr_SetString(PyExc_TypeError, "Invalid C set state key parameter");
        goto error;
    }
    else if (state == NULL) {
        PyErr_SetString(PyExc_TypeError, "Invalid C set state-state parameter");
        goto error;
    }

    cb_set_state(check_id, key, state);

    PyGILState_Release(gstate);
    Py_RETURN_NONE; // Success

error:
    PyGILState_Release(gstate);
    return NULL; // Failure
}

/*! \fn get_state(PyObject *self, PyObject *args)
    \brief Aggregator builtin class method for topology component submission.
    \param self A PyObject * pointer to self - the aggregator module.
    \param args A PyObject * pointer to the python args or kwargs.
    \return This function returns a new reference to None (already INCREF'd), or NULL in case of error.

    This function implements the `get_state` python callable in C and is used from the python code.
    More specifically, in the context of rtloader and datadog-agent, this is called from our python base check
    class to get a state for a agent check.
*/
static PyObject *get_state(PyObject *self, PyObject *args) {
    if (cb_get_state == NULL) {
        PyErr_SetString(PyExc_TypeError, "`get_state` is set as NULL");
        Py_RETURN_NONE;
    }

    PyObject *check = NULL; // borrowed
    char *check_id;
    char *key;
    char *state_value;

    PyGILState_STATE gstate = PyGILState_Ensure();

    if (!PyArg_ParseTuple(args, "Oss", &check, &check_id, &key)) {
      goto error;
    }
    else if (check == NULL) {
        PyErr_SetString(PyExc_TypeError, "Invalid C get state check parameter");
        goto error;
    }
    else if (check_id == NULL) {
        PyErr_SetString(PyExc_TypeError, "Invalid C get state check id parameter");
        goto error;
    }
    else if (key == NULL) {
        PyErr_SetString(PyExc_TypeError, "Invalid C get state key parameter");
        goto error;
    }



    state_value = cb_get_state(check_id, key);

    if (state_value != NULL) {
        PyGILState_Release(gstate);
        return PyStringFromCString(state_value);
    }

    PyErr_SetString(PyExc_TypeError, "Error, Unable to GetState, Get state returned as null");
    goto error;

error:
    PyGILState_Release(gstate);
    return NULL; // Failure
}
