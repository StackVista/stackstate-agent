// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com/).
// Copyright 2018-present StackState

//go:build kubeapiserver

package apiserver

import (
	"k8s.io/apimachinery/pkg/version"
)

// GetVersion retrieves the version of the Kubernetes cluster
func (c *APIClient) GetVersion() (*version.Info, error) {
	version, err := c.Cl.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}

	return version, nil
}
