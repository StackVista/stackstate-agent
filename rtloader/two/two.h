// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog
// (https://www.datadoghq.com/).
// Copyright 2019-present Datadog, Inc.
#ifndef DATADOG_AGENT_RTLOADER_TWO_H
#define DATADOG_AGENT_RTLOADER_TWO_H

// Some preprocessor sanity for builds (2+3 common sources)
#ifndef DATADOG_AGENT_TWO
#    error Build requires defining DATADOG_AGENT_TWO
#elif defined(DATADOG_AGENT_TWO) && defined(DATADOG_AGENT_THREE)
#    error "DATADOG_AGENT_TWO and DATADOG_AGENT_THREE are mutually exclusive - define only one of the two."
#endif

#include <map>
#include <string>
#include <vector>

#include <Python.h>
#include <rtloader.h>

class Two : public RtLoader
{
public:
    //! Constructor.
    /*!
      \param python_home A C-string with the path to the python home for the
      python interpreter.
      \param python_exe A C-string with the path to the python interpreter.

      Basic constructor, initializes the _error string to an empty string and
      errorFlag to false and set the supplied PYTHONHOME.
    */
    Two(const char *python_home, const char *python_exe, cb_memory_tracker_t memtrack_cb);

    //! Destructor.
    /*!
      Destroys the Two instance, including relevant python teardown calls.

      We do not call Py_Finalize() since we won't be calling it from the same
      thread where we called Py_Initialize(), this is a product of the go runtime
      switch threads constantly. It's not really an issue here as we destroy this
      class instance just before exiting the agent.
      Calling Py_Finalize from a different thread cause the "threading"
      package to raise an exception: "Exception KeyError: KeyError(<current
      thread id>,) in <module 'threading'".
      Even if Python ignores it, the exception ends up in the log files for
      upstart/syslog/...
      Since we don't call Py_Finalize, we don't free _pythonHome here either.

      More info here:
      https://stackoverflow.com/questions/8774958/keyerror-in-module-threading-after-a-successful-py-test-run/12639040#12639040

    */
    ~Two();

    bool init();
    bool addPythonPath(const char *path);
    rtloader_gilstate_t GILEnsure();
    void GILRelease(rtloader_gilstate_t);

    bool getClass(const char *module, RtLoaderPyObject *&pyModule, RtLoaderPyObject *&pyClass);
    bool getAttrString(RtLoaderPyObject *obj, const char *attributeName, char *&value) const;
    bool getCheck(RtLoaderPyObject *py_class, const char *init_config_str, const char *instance_str,
                  const char *check_id_str, const char *check_name, const char *agent_config_str,
                  RtLoaderPyObject *&check);

    char *runCheck(RtLoaderPyObject *check);
    void cancelCheck(RtLoaderPyObject *check);
    char **getCheckWarnings(RtLoaderPyObject *check);
    void decref(RtLoaderPyObject *obj);
    void incref(RtLoaderPyObject *obj);
    void setModuleAttrString(char *module, char *attr, char *value);

    // const API
    py_info_t *getPyInfo();
    void freePyInfo(py_info_t *);
    bool runSimpleString(const char *code) const;
    RtLoaderPyObject *getNone() const
    {
        return reinterpret_cast<RtLoaderPyObject *>(Py_None);
    }

    // Python Helpers
    char *getIntegrationList();
    char *getInterpreterMemoryUsage();

    // aggregator API
    void setSubmitMetricCb(cb_submit_metric_t);
    void setSubmitServiceCheckCb(cb_submit_service_check_t);
    void setSubmitEventCb(cb_submit_event_t);
    void setSubmitHistogramBucketCb(cb_submit_histogram_bucket_t);
    void setSubmitEventPlatformEventCb(cb_submit_event_platform_event_t);

    // datadog_agent API
    void setGetVersionCb(cb_get_version_t);
    void setGetConfigCb(cb_get_config_t);
    void setHeadersCb(cb_headers_t);
    void setGetHostnameCb(cb_get_hostname_t);
    void setGetClusternameCb(cb_get_clustername_t);
    void setGetPidCb(cb_get_pid_t);
    void setGetCreateTimeCb(cb_get_create_time_t);
    void setGetTracemallocEnabledCb(cb_tracemalloc_enabled_t);
    void setLogCb(cb_log_t);
    void setSetCheckMetadataCb(cb_set_check_metadata_t);
    void setSetExternalTagsCb(cb_set_external_tags_t);
    void setWritePersistentCacheCb(cb_write_persistent_cache_t);
    void setReadPersistentCacheCb(cb_read_persistent_cache_t);
    void setObfuscateSqlCb(cb_obfuscate_sql_t);
    void setObfuscateSqlExecPlanCb(cb_obfuscate_sql_exec_plan_t);
    void setGetProcessStartTimeCb(cb_get_process_start_time_t);

    // _util API
    virtual void setSubprocessOutputCb(cb_get_subprocess_output_t);

    // CGO API
    void setCGOFreeCb(cb_cgo_free_t);

    // tagger
    void setTagsCb(cb_tags_t);

    // kubeutil
    void setGetConnectionInfoCb(cb_get_connection_info_t);

    // containers
    void setIsExcludedCb(cb_is_excluded_t);

    // topology
    void setSubmitComponentCb(cb_submit_component_t);
    void setSubmitRelationCb(cb_submit_relation_t);
    void setSubmitStartSnapshotCb(cb_submit_start_snapshot_t);
    void setSubmitStopSnapshotCb(cb_submit_stop_snapshot_t);
    void setSubmitDeleteCb(cb_submit_delete_t);

    // telemetry
    void setSubmitTopologyEventCb(cb_submit_topology_event_t);

    // raw metrics
    void setSubmitRawMetricsDataCb(cb_submit_raw_metrics_data_t);

    // health
    void setSubmitHealthCheckDataCb(cb_submit_health_check_data_t);
    void setSubmitHealthStartSnapshotCb(cb_submit_health_start_snapshot_t);
    void setSubmitHealthStopSnapshotCb(cb_submit_health_stop_snapshot_t);

    // transaction state
    void setStartTransactionCb(cb_start_transaction_t);
    void setStopTransactionCb(cb_stop_transaction_t);
    void setDiscardTransactionCb(cb_discard_transaction_t);
    void setTransactionStateCb(cb_set_transaction_state_t);

    // state
    void setStateCb(cb_set_state_t);
    char *setGetStateCb(cb_get_state_t);

private:
    //! initPythonHome member.
    /*!
      \brief This member function sets the Python home for the underlying python2.7 interpreter.
      \param pythonHome A C-string to the target python home for the python runtime.
    */
    void initPythonHome(const char *pythonHome = NULL);

    //! initPythonExe member.
    /*!
      \brief This member function sets the path to the underlying python2 interpreter.
      \param python_exe A C-string to the target python executable.
    */
    void initPythonExe(const char *python_exe = NULL);

    //! _importFrom member.
    /*!
      \brief This member function imports a Python object by name from the specified
      module.
      \param module A C-string representation of the Python module we wish to import from.
      \param name A C-string representation of the target Python object we wish to import.
      \return A PyObject * pointer to the imported Python object, or NULL in case of error.

      This function returns a new reference to the underlying PyObject. In case of error,
      NULL is returned with clean interpreter error flag.
    */
    PyObject *_importFrom(const char *module, const char *name);

    //! _findSubclassOf member.
    /*!
      \brief This member function attemts to find a subclass of the provided base class in
      the specified Python module.
      \param base A PyObject * pointer to the Python base class we wish to search for.
      \param moduleName A PyObject * pointer to the module we wish to find a derived class
      in.
      \return A PyObject * pointer to the found subclass Python object, or NULL in case of error.

      This function returns a new reference to the underlying PyObject. In case of error,
      NULL is returned with clean interpreter error flag.
    */
    PyObject *_findSubclassOf(PyObject *base, PyObject *moduleName);

    //! _fetchPythonError member.
    /*!
      \brief This member function retrieves the error set on the python interpreter.
      \return A string describing the python error/exception set on the underlying python
      interpreter.
    */
    std::string _fetchPythonError();

    /*! PyPaths type prototype
      \typedef PyPaths defines a vector of strings.
    */
    typedef std::vector<std::string> PyPaths;

    char *_pythonHome; /*!< string with the PYTHONHOME for the underlying interpreter */
    char *_pythonExe; /*!< string with the path to the executable of the underlying interpreter */
    PyObject *_baseClass; /*!< PyObject * pointer to the base Agent check class */
    PyPaths _pythonPaths; /*!< string vector containing paths in the PYTHONPATH */
    PyThreadState *_threadState; /*!< PyThreadState * pointer to the saved Python interpreter thread state */
};

#endif
