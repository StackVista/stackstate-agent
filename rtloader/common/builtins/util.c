// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2019-2020 Datadog, Inc.
#include "util.h"
#include "datadog_agent.h"
#include "rtloader_mem.h"

#include <stringutils.h>

static PyObject *headers(PyObject *self, PyObject *args, PyObject *kwargs);
static PyObject *get_hostname(PyObject *self, PyObject *args);
static PyObject *get_clustername(PyObject *self, PyObject *args);
static PyObject *log_message(PyObject *self, PyObject *args);
static PyObject *set_external_tags(PyObject *self, PyObject *args);

static PyMethodDef methods[] = {
    { "headers", (PyCFunction)headers, METH_VARARGS | METH_KEYWORDS, "Get standard set of HTTP headers." },
    { NULL, NULL } // guards
};

#ifdef DATADOG_AGENT_THREE
static struct PyModuleDef module_def = { PyModuleDef_HEAD_INIT, UTIL_MODULE_NAME, NULL, -1, methods };

PyMODINIT_FUNC PyInit_util(void)
{
    return PyModule_Create(&module_def);
}
#elif defined(DATADOG_AGENT_TWO)
// in Python2 keep the object alive for the program lifetime
static PyObject *module;

void Py2_init_util()
{
    module = Py_InitModule(UTIL_MODULE_NAME, methods);
}
#endif

/*! \fn PyObject *headers(PyObject *self, PyObject *args, PyObject *kwargs)
    \brief This function provides a standard set of HTTP headers the caller might want to
    use for HTTP requests.
    \param self A PyObject* pointer to the util module.
    \param args A PyObject* pointer to the `agentConfig`, but not expected to be used.
    \param kwargs A PyObject* pointer to a dictonary. If the `http_host` key is present
    it will be added to the headers.
    \return a PyObject * pointer to a python dictionary with the expected headers.

    This function is callable as the `util.headers` python method, the entry point:
    `_public_headers()` is provided in the `datadog_agent` module, the method is duplicated.
*/
PyObject *headers(PyObject *self, PyObject *args, PyObject *kwargs)
{
    return _public_headers(self, args, kwargs);
}

/*! \fn py_tag_to_c(PyObject *py_tags)
    \brief A function to convert a list of python strings (tags) into an
    array of C-strings.
    \return a char ** pointer to the C-representation of the provided python
    tag list. In the event of failure NULL is returned.

    The returned char ** string array pointer is heap allocated here and should
    be subsequently freed by the caller. This function may set and raise python
    interpreter errors. The function is static and not in the builtin's API.
*/
char **py_tag_to_c(PyObject *py_tags)
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

/*! \fn free_tags(char **tags)
    \brief A helper function to free the memory allocated by the py_tag_to_c() function.

    This function is for internal use and expects the tag array to be properly intialized,
    and have a NULL canary at the end of the array, just like py_tag_to_c() initializes and
    populates the array. Be mindful if using this function in any other context.
*/
void free_tags(char **tags)
{
    int i;
    for (i = 0; tags[i] != NULL; i++) {
        _free(tags[i]);
    }
    _free(tags);
}
