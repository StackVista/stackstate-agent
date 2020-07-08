// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package checks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/StackVista/stackstate-agent/pkg/compliance"
	"github.com/StackVista/stackstate-agent/pkg/compliance/checks/env"
	"github.com/StackVista/stackstate-agent/pkg/compliance/eval"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

var (
	// ErrPropertyKindNotSupported is returned for property kinds not supported by the check
	ErrPropertyKindNotSupported = errors.New("property kind '%s' not supported")

	// ErrPropertyNotSupported is returned for properties not supported by the check
	ErrPropertyNotSupported = errors.New("property '%s' not supported")
)

type pathMapper func(string) string

type fileCheck struct {
	baseCheck
	file *compliance.File
}

func newFileCheck(baseCheck baseCheck, file *compliance.File) (*fileCheck, error) {
	// TODO: validate config for the file here
	return &fileCheck{
		baseCheck: baseCheck,
		file:      file,
	}, nil
}

func (c *fileCheck) Run() error {
	// TODO: here we will introduce various cached results lookups

	var err error
	path := c.file.Path
	if path == "" {
		path, err = c.ResolveValueFrom(c.file.PathFrom)
		if err != nil {
			return err
		}
	}

	log.Debugf("%s: file check: %s", c.ruleID, path)
	if path != "" {
		return c.reportFile(c.NormalizePath(path))
	}

	return log.Error("no path for file check")
}

	path, err := resolvePath(e, file.Path)
	if err != nil {
		return nil, err
	}

	paths, err := filepath.Glob(e.NormalizeToHostRoot(path))
	if err != nil {
		return nil, err
	}

	var instances []*eval.Instance

		switch field.Kind {
		case compliance.PropertyKindAttribute:
			v, err = c.getAttribute(filePath, fi, field.Property)
		case compliance.PropertyKindJSONQuery:
			v, err = queryValueFromFile(filePath, field.Property, jsonGetter)
		case compliance.PropertyKindYAMLQuery:
			v, err = queryValueFromFile(filePath, field.Property, yamlGetter)
		default:
			return invalidInputErr(ErrPropertyKindNotSupported, field.Kind)
		}
		if err != nil {
			// This is not a failure unless we don't have any paths to act on
			log.Debugf("%s: file check failed to stat %s [%s]", ruleID, path, relPath)
			continue
		}

		instance := &eval.Instance{
			Vars: eval.VarMap{
				compliance.FileFieldPath:        relPath,
				compliance.FileFieldPermissions: uint64(fi.Mode() & os.ModePerm),
			},
			Functions: eval.FunctionMap{
				compliance.FileFuncJQ:     fileJQ(path),
				compliance.FileFuncYAML:   fileYAML(path),
				compliance.FileFuncRegexp: fileRegexp(path),
			},
		}

		user, err := getFileUser(fi)
		if err == nil {
			instance.Vars[compliance.FileFieldUser] = user
		}

		group, err := getFileGroup(fi)
		if err == nil {
			instance.Vars[compliance.FileFieldGroup] = group
		}

		instances = append(instances, instance)
	}
	return "", invalidInputErr(ErrPropertyNotSupported, property)
}

// getter applies jq query to get string value from json or yaml raw data
type getter func([]byte, string) (string, error)

	if len(instances) == 0 {
		return nil, fmt.Errorf("no files found for file check %q", file.Path)
	}

	return &instanceIterator{
		instances: instances,
	}, nil
}

func fileQuery(path string, get getter) eval.Function {
	return func(_ *eval.Instance, args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf(`invalid number of arguments, expecting 1 got %d`, len(args))
		}
		query, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf(`expecting string value for query argument`)
		}
		return queryValueFromFile(path, query, get)
	}
}

func queryValueFromFile(filePath string, query string, get getter) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

func fileYAML(path string) eval.Function {
	return fileQuery(path, yamlGetter)
}

func fileRegexp(path string) eval.Function {
	return fileQuery(path, regexpGetter)
}
