// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com).
// Copyright 2021 StackState

#ifndef DATADOG_AGENT_RTLOADER_THREE_TRANSACTION_H
#define DATADOG_AGENT_RTLOADER_THREE_TRANSACTION_H

/*! \file transaction.h
    \brief RtLoader transactional state builtin header file.

    The prototypes here defined provide functions to initialize the python transaction
    builtin module, and set its relevant callbacks for the rtloader caller.
*/
/*! \def TRANSACTION_MODULE_NAME
    \brief Transaction state module name definition.
*/
/*! \fn PyMODINIT_FUNC PyInit_transaction(void)
    \brief Initializes the transaction state builtin python module.

    The python python builtin is created and registered here as per the module_def
    PyMethodDef definition in `transaction.c` with the corresponding C-implemented python
    methods. A fresh reference to the module is created here. This function is python3
    only.
*/
/*! \fn void Py2_init_transaction()
    \brief Initializes the transaction state builtin python module.

    The transactional python builtin is created and registered here as per the module_def
    PyMethodDef definition in `transaction.c` with the corresponding C-implemented python
    methods. A fresh reference to the module is created here. This function is python2
    only.
*/
/*! \fn void _set_start_transaction_cb(cb_start_transaction_t)
    \brief Sets the start transaction callback to be used by rtloader for transactional state.
    \param object A function pointer with cb_start_transaction_t function prototype to the
    callback function.

    The callback is expected to be provided by the rtloader caller - in go-context: CGO.
*/
/*! \fn void _set_stop_transaction_cb(cb_stop_transaction_t)
    \brief Sets the stop transaction callback to be used by rtloader for transactional state.
    \param object A function pointer with cb_stop_transaction_t function prototype to the
    callback function.

    The callback is expected to be provided by the rtloader caller - in go-context: CGO.
*/
/*! \fn void _set_discard_transaction_cb(cb_discard_transaction_t)
    \brief Sets the cancel transaction callback to be used by rtloader for transactional state.
    \param object A function pointer with cb_discard_transaction_t function prototype to the
    callback function.

    The callback is expected to be provided by the rtloader caller - in go-context: CGO.
*/
/*! \fn void _set_transaction_state_cb(cb_set_transaction_state_t)
    \brief Sets the transaction state callback to be used by rtloader for transactional state.
    \param object A function pointer with cb_set_transaction_state_t function prototype to the
    callback function.

    The callback is expected to be provided by the rtloader caller - in go-context: CGO.
*/

#include <Python.h>
#include <rtloader_types.h>

#define TRANSACTION_MODULE_NAME "transaction"

#ifdef __cplusplus
extern "C" {
#endif

#ifdef DATADOG_AGENT_THREE
PyMODINIT_FUNC PyInit_transaction(void);
#elif defined(DATADOG_AGENT_TWO)
void Py2_init_transaction();
#endif

void _set_start_transaction_cb(cb_start_transaction_t);
void _set_stop_transaction_cb(cb_stop_transaction_t);
void _set_discard_transaction_cb(cb_discard_transaction_t);
void _set_transaction_state_cb(cb_set_transaction_state_t);

#ifdef __cplusplus
}
#endif

#endif
