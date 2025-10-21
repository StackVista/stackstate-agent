// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com).
// Copyright 2021 StackState

#ifndef DATADOG_AGENT_RTLOADER_THREE_STATE_H
#define DATADOG_AGENT_RTLOADER_THREE_STATE_H

/*! \file state.h
    \brief RtLoader state builtin header file.

    The prototypes here defined provide functions to initialize the python state
    builtin module, and set its relevant callbacks for the rtloader caller.
*/
/*! \def STATE_MODULE_NAME
    \brief State module name definition.
*/
/*! \fn PyMODINIT_FUNC PyInit_state(void)
    \brief Initializes the state builtin python module.

    The python python builtin is created and registered here as per the module_def
    PyMethodDef definition in `state.c` with the corresponding C-implemented python
    methods. A fresh reference to the module is created here. This function is python3
    only.
*/
/*! \fn void Py2_init_state()
    \brief Initializes the state builtin python module.

    The state python builtin is created and registered here as per the module_def
    PyMethodDef definition in `state.c` with the corresponding C-implemented python
    methods. A fresh reference to the module is created here. This function is python2
    only.
*/
/*! \fn void _set_state_cb(cb_set_state_t)
    \brief Sets the set state callback to be used by rtloader for state submission.
    \param object A function pointer with cb_set_state_t function prototype to the
    callback function.

    The callback is expected to be provided by the rtloader caller - in go-context: CGO.
*/
/*! \fn void _set_get_state_cb(cb_get_state_t)
    \brief Sets the get state callback to be used by rtloader for state retrieval.
    \param object A function pointer with cb_get_state_t function prototype to the
    callback function.

    The callback is expected to be provided by the rtloader caller - in go-context: CGO.
*/
#define PY_SSIZE_T_CLEAN
#include <Python.h>
#include <rtloader_types.h>

#define STATE_MODULE_NAME "state"

#ifdef __cplusplus
extern "C" {
#endif

#ifdef DATADOG_AGENT_THREE
PyMODINIT_FUNC PyInit_state(void);
#elif defined(DATADOG_AGENT_TWO)
void Py2_init_state();
#endif

void _set_state_cb(cb_set_state_t);
void _set_get_state_cb(cb_get_state_t);

#ifdef __cplusplus
}
#endif

#endif
