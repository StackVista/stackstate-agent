// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com).
// Copyright 2021 StackState

#ifndef DATADOG_AGENT_RTLOADER_THREE_RAW_METRICS_H
#define DATADOG_AGENT_RTLOADER_THREE_RAW_METRICS_H

// TODO: Raw Metrics

/*! \file raw_metrics.h
    \brief RtLoader raw_metrics builtin header file.

    The prototypes here defined provide functions to initialize the python raw metrics
    builtin module, and set its relevant callbacks for the rtloader caller.
*/
/*! \def RAW_METRICS_MODULE_NAME
    \brief Raw metrics module name definition.
*/
/*! \fn PyMODINIT_FUNC PyInit_raw_metrics(void)
    \brief Initializes the raw metrics builtin python module.

    The python python builtin is created and registered here as per the module_def
    PyMethodDef definition in `raw_metrics.c` with the corresponding C-implemented python
    methods. A fresh reference to the module is created here. This function is python3
    only.
*/
/*! \fn void Py2_init_raw_metrics()
    \brief Initializes the raw metrics builtin python module.

    The raw metrics python builtin is created and registered here as per the module_def
    PyMethodDef definition in `raw_metrics.c` with the corresponding C-implemented python
    methods. A fresh reference to the module is created here. This function is python2
    only.
*/
/*! \fn void _set_submit_raw_metrics_data_t(cb_submit_raw_metrics_data_t)
    \brief Sets the submit raw metrics check data callback to be used by rtloader for check data submission.
    \param object A function pointer with cb_submit_raw_metrics_data_t function prototype to the
    callback function.

    The callback is expected to be provided by the rtloader caller - in go-context: CGO.
*/
/*! \fn void _set_submit_raw_metrics_start_snapshot_t(cb_submit_raw_metrics_start_snapshot_t)
    \brief Sets the submit start raw metrics snapshot callback to be used by rtloader to signal the raw metrics snapshot start.
    \param object A function pointer with cb_submit_raw_metrics_start_snapshot_t function prototype to the
    callback function.

    The callback is expected to be provided by the rtloader caller - in go-context: CGO.
*/
/*! \fn void _set_submit_raw_metrics_stop_snapshot_t(cb_submit_raw_metrics_stop_snapshot_t)
    \brief Sets the submit raw metrics stop snapshot callback to be used by rtloader to signal the raw metrics snapshot stop.
    \param object A function pointer with cb_submit_raw_metrics_stop_snapshot_t function prototype to the
    callback function.

    The callback is expected to be provided by the rtloader caller - in go-context: CGO.
*/

#include <Python.h>
#include <rtloader_types.h>

#define RAW_METRICS_MODULE_NAME "raw_metrics"

#ifdef __cplusplus
extern "C" {
#endif

#ifdef DATADOG_AGENT_THREE
PyMODINIT_FUNC PyInit_raw_metrics(void);
#elif defined(DATADOG_AGENT_TWO)
void Py2_init_raw_metrics();
#endif

void _set_submit_raw_metrics_data_cb(cb_submit_raw_metrics_data_t);
void _set_submit_raw_metrics_start_snapshot_cb(cb_submit_raw_metrics_start_snapshot_t);
void _set_submit_raw_metrics_stop_snapshot_cb(cb_submit_raw_metrics_stop_snapshot_t);


#ifdef __cplusplus
}
#endif

#endif
