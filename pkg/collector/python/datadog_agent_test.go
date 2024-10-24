// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build python,test

package python

import (
	"testing"
)

func TestGetVersion(t *testing.T) {
	testGetVersion(t)
}

func TestGetHostname(t *testing.T) {
	testGetHostname(t)
}

func TestGetClusterName(t *testing.T) {
	testGetClusterName(t)
}

func TestGetPid(t *testing.T) {
	testGetPid(t)
}

func TestGetCreateTime(t *testing.T) {
	testGetCreateTime(t)
}

func TestHeaders(t *testing.T) {
	testHeaders(t)
}

func TestGetConfig(t *testing.T) {
	testGetConfig(t)
}

func TestSetExternalTags(t *testing.T) {
	testSetExternalTags(t)
}
