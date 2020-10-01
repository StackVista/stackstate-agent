// +build cpython

#include "topology_api.h"

PyObject* SubmitContextualizedEvent(PyObject*, char*, PyObject*);

static PyObject *submit_contextualized_event(PyObject *self, PyObject *args) {
    PyObject *check = NULL;
    PyObject *event = NULL;
    char *check_id;

    PyGILState_STATE gstate;
    gstate = PyGILState_Ensure();

    // aggregator.submit_event(self, check_id, event)
    if (!PyArg_ParseTuple(args, "OsO", &check, &check_id, &event)) {
      PyGILState_Release(gstate);
      return NULL;
    }

    PyGILState_Release(gstate);
    return SubmitContextualizedEvent(check, check_id, event);
}

static PyMethodDef TelemetryMethods[] = {
  {"submit_contextualized_event", (PyCFunction)submit_contextualized_event, METH_VARARGS, "Submit events to the aggregator."},
  {NULL, NULL}  // guards
};

void inittelemetry()
{
  PyGILState_STATE gstate;
  gstate = PyGILState_Ensure();

  PyObject *m = Py_InitModule("telemetry", TelemetryMethods);

  PyGILState_Release(gstate);
}
