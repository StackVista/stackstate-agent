// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build python && test
// +build python,test

package python

import (
	"context"
	collectorutils "github.com/StackVista/stackstate-agent/pkg/collector/util"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"

	"github.com/StackVista/stackstate-agent/pkg/metadata/externalhost"
	"github.com/StackVista/stackstate-agent/pkg/util"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/clustername"
	"github.com/StackVista/stackstate-agent/pkg/version"
)

import "C"

func testGetVersion(t *testing.T) {
	var v *C.char
	GetVersion(&v)
	require.NotNil(t, v)

	av, _ := version.Agent()
	assert.Equal(t, av.GetNumber(), C.GoString(v))
}

func testGetHostname(t *testing.T) {
	var h *C.char
	GetHostname(&h)
	require.NotNil(t, h)

	hostname, _ := util.GetHostname(context.Background())
	assert.Equal(t, hostname, C.GoString(h))
}

func testGetClusterName(t *testing.T) {
	var ch *C.char
	GetClusterName(&ch)
	require.NotNil(t, ch)

	assert.Equal(t, clustername.GetClusterName(context.Background(), ""), C.GoString(ch))
}

func testGetPid(t *testing.T) {
	var ch *C.char
	GetPid(&ch)
	require.NotNil(t, ch)

	assert.Equal(t, strconv.Itoa(os.Getpid()), C.GoString(ch))
}

func testGetCreateTime(t *testing.T) {
	var ch *C.char
	GetCreateTime(&ch)
	require.NotNil(t, ch)

	pid := os.Getpid()
	var goCreateTime int64
	if ct, err := collectorutils.GetProcessCreateTime(int32(pid)); err != nil {
		t.Fatalf("datadog_agent: could not get create time for process %d: %s", pid, err)
	} else {
		goCreateTime = ct
	}

	assert.Equal(t, strconv.FormatInt(goCreateTime, 10), C.GoString(ch))
}

func testHeaders(t *testing.T) {
	var headers *C.char
	Headers(&headers)
	require.NotNil(t, headers)

	h := util.HTTPHeaders()
	yamlPayload, _ := yaml.Marshal(h)
	assert.Equal(t, string(yamlPayload), C.GoString(headers))
}

func testGetConfig(t *testing.T) {
	var config *C.char

	GetConfig(C.CString("does not exist"), &config)
	require.Nil(t, config)

	GetConfig(C.CString("cmd_port"), &config)
	require.NotNil(t, config)
	assert.Equal(t, "5001\n", C.GoString(config))
}

func testSetExternalTags(t *testing.T) {
	ctags := []*C.char{C.CString("tag1"), C.CString("tag2"), nil}

	SetExternalTags(C.CString("test_hostname"), C.CString("test_source_type"), &ctags[0])

	payload := externalhost.GetPayload()
	require.NotNil(t, payload)

	yamlPayload, _ := yaml.Marshal(payload)
	assert.Equal(t,
		"- - test_hostname\n  - test_source_type:\n    - tag1\n    - tag2\n",
		string(yamlPayload))
}
