// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2019-2020 Datadog, Inc.
#include <stdlib.h>

#include "rtloader_mem.h"
#include "rtloader_types.h"
#include "stringutils.h"


PyObject * yload = NULL;
PyObject * ydump = NULL;
PyObject * loader = NULL;
PyObject * dumper = NULL;

PyObject * stringio_module = NULL;
PyObject * ruamel_module = NULL;
/**
 * returns a C (NULL terminated UTF-8) string from a python string.
 *
 * \param object  A Python string to be converted to C-string.
 *
 * \return A standard C string (NULL terminated character pointer)
 *  The returned pointer is allocated from the heap and must be
 * deallocated (free()ed) by the caller
 */
char *as_string(PyObject *object)
{
    if (object == NULL) {
        return NULL;
    }

    char *retval = NULL;

// DATADOG_AGENT_THREE implementation is the default
#ifdef DATADOG_AGENT_TWO
    if (!PyString_Check(object) && !PyUnicode_Check(object)) {
        return NULL;
    }

    char *tmp = PyString_AsString(object);
    if (tmp == NULL) {
        // PyString_AsString might raise an error when python can't encode a
        // unicode string to byte
        PyErr_Clear();
        return NULL;
    }
    retval = strdupe(tmp);
#else
    PyObject *temp_bytes = NULL;

    if (PyBytes_Check(object)) {
        // We already have an encoded string, we suppose it has the correct encoding (UTF-8)
        temp_bytes = object;
        Py_INCREF(temp_bytes);
    } else if (PyUnicode_Check(object)) {
        // Encode the Unicode string that was given
        temp_bytes = PyUnicode_AsEncodedString(object, "UTF-8", "strict");
        if (temp_bytes == NULL) {
            // PyUnicode_AsEncodedString might raise an error if the codec raised an
            // exception
            PyErr_Clear();
            return NULL;
        }
    } else {
        return NULL;
    }

    retval = strdupe(PyBytes_AS_STRING(temp_bytes));
    Py_XDECREF(temp_bytes);
#endif

    return retval;
}

int init_stringutils(void) {
    PyObject *yaml = NULL;
    int ret = EXIT_FAILURE;

    char module_name[] = "yaml";
    yaml = PyImport_ImportModule(module_name);
    if (yaml == NULL) {
        goto done;
    }

    // get pyyaml load()
    char load_name[] = "load";
    yload = PyObject_GetAttrString(yaml, load_name);
    if (yload == NULL) {
        goto done;
    }

    // We try to use the C-extensions, if they're available, but it's a best effort
    char c_loader_name[] = "CSafeLoader";
    loader = PyObject_GetAttrString(yaml, c_loader_name);
    if (loader == NULL) {
        PyErr_Clear();
        char loader_name[] = "SafeLoader";
        loader = PyObject_GetAttrString(yaml, loader_name);
        if (loader == NULL) {
            goto done;
        }
    }

    // get pyyaml dump()
    char dump_name[] = "dump";
    ydump = PyObject_GetAttrString(yaml, dump_name);
    if (ydump == NULL) {
        goto done;
    }

    char c_dumper_name[] = "CSafeDumper";
    dumper = PyObject_GetAttrString(yaml, c_dumper_name);
    if (dumper == NULL) {
        PyErr_Clear();
        char dumper_name[] = "SafeDumper";
        dumper = PyObject_GetAttrString(yaml, dumper_name);
        if (dumper == NULL) {
            goto done;
        }
    }
    // from ruamel.yaml import YAML
    char module_name_r[] = "ruamel.yaml";
    PyObject *ruamel_m = PyImport_ImportModule(module_name_r);
    char module_name_YAML[] = "YAML";
    ruamel_module = PyObject_GetAttrString(ruamel_m, module_name_YAML);
    if (ruamel_module == NULL) {
        goto done;
    }

    // from ruamel.yaml.compat import StringIO
    char module_name_compat[] = "ruamel.yaml.compat";
    PyObject *ruamel_compat = PyImport_ImportModule(module_name_compat);
    char module_name_StringIO[] = "StringIO";
    stringio_module = PyObject_GetAttrString(ruamel_compat, module_name_StringIO);
    if (stringio_module == NULL) {
        goto done;
    }

    ret = EXIT_SUCCESS;

    return ret;

done:
    Py_XDECREF(module_name_r);
    Py_XDECREF(module_name_YAML);
    Py_XDECREF(ruamel_m);
    Py_XDECREF(module_name_compat);
    Py_XDECREF(module_name_StringIO);
    Py_XDECREF(yaml);
    return ret;
}

PyObject *from_yaml(const char *data) {
    PyObject *args = NULL;
    PyObject *kwargs = NULL;
    PyObject *retval = NULL;

    if (!data) {
        goto done;
    }
    if (yload == NULL) {
        goto done;
    }

    args = PyTuple_New(0);
    if (args == NULL) {
        goto done;
    }
    kwargs = Py_BuildValue("{s:s, s:O}", "stream", data, "Loader", loader);
    if (kwargs == NULL) {
        goto done;
    }
    retval = PyObject_Call(yload, args, kwargs);

done:
    Py_XDECREF(kwargs);
    Py_XDECREF(args);
    return retval;
}

char *as_yaml(PyObject *object) {
    char *retval = NULL;
    PyObject *dumped = NULL;

    PyObject *args = PyTuple_New(0);
    PyObject *kwargs = Py_BuildValue("{s:O, s:O}", "data", object, "Dumper", dumper);

    dumped = PyObject_Call(ydump, args, kwargs);
    if (dumped == NULL) {
        goto done;
    }
    retval = as_string(dumped);

done:
    //Py_XDECREF can accept (and ignore) NULL references
    Py_XDECREF(dumped);
    Py_XDECREF(kwargs);
    Py_XDECREF(args);
    return retval;
}

char *as_yaml_ruamel(PyObject *object) {
    char *retval = NULL;
    PyObject *dumped = NULL;

     // ruamel = YAML(typ='safe')
    PyObject *args = PyTuple_New(0);
    PyObject *kwargs = Py_BuildValue("{s:s}", "typ", "safe");
    PyObject *ruamel = PyObject_Call(ruamel_module, args, kwargs);

    // get ruamel dump()
    char r_dump_name[] = "dump";
    PyObject *rdump = PyObject_GetAttrString(ruamel, r_dump_name);
    if (rdump == NULL) {
        goto done;
    }

    // stream = StringIO()
    args = PyTuple_New(0);
    PyObject *stream = PyObject_Call(stringio_module, args, NULL);
    if (stream == NULL) {
        goto done;
    }

    // yaml.dump(data, stream) --> returns NULL
    args = Py_BuildValue("O,O", object, stream);
    PyObject_Call(rdump, args, NULL);
    if (PyErr_Occurred()) {
        retval = NULL; // Failure
        goto done;
    }

    // get stream getvalue()
    char get_value_name[] = "getvalue";
    PyObject *get_value_func = PyObject_GetAttrString(stream, get_value_name); // borrowed
    if (get_value_func == NULL) {
        PyErr_SetString(PyExc_TypeError, "error: no function 'getvalue' found for StringIO()");
        retval = NULL; // Failure
        goto done;
    }

    // stream.getvalue() --> returns string
    args = PyTuple_New(0);
    dumped = PyObject_Call(get_value_func, args, NULL);
    if (dumped == NULL) {
        PyErr_SetString(PyExc_TypeError, "error: nothing dumped into stream");
        retval = NULL; // Failure
        goto done;
    }
    retval = as_string(dumped);

done:
    //Py_XDECREF can accept (and ignore) NULL references
    Py_XDECREF(args);
    Py_XDECREF(kwargs);
    Py_XDECREF(stream);
    Py_XDECREF(dumped);
    Py_XDECREF(r_dump_name);
    Py_XDECREF(get_value_name);
    return retval;
}
