// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package util contains helper functions for processes, IP addresses, env variables, etc.
package util

import (
	"errors"
)

// GetDockerSocketPath returns a stable error in builds without Docker support.
func GetDockerSocketPath() (string, error) {
	return "", errors.New("docker is not available")
}
