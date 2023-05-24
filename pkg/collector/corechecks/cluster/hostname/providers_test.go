// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package hostname

import (
	"testing"

	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/cluster/hostname/testutil"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func dummyProvider(node v1.Node) string {
	return "dummy-hostname"
}

func dummyErrorProvider(node v1.Node) string {
	return ""
}

func dummyInvalidProvider(node v1.Node) string {
	return "some invalid hostname"
}

func TestRegisterHostnameProvider(t *testing.T) {
	RegisterHostnameProvider("dummy", dummyProvider)
	assert.Contains(t, providerCatalog, "dummy")
	delete(providerCatalog, "dummy")
}

func TestGetProvider(t *testing.T) {
	RegisterHostnameProvider("dummy", dummyProvider)
	defer delete(providerCatalog, "dummy")
	assert.NotNil(t, GetProvider("dummy"))
	assert.Nil(t, GetProvider("does not exists"))
}

func TestGetHostname(t *testing.T) {
	RegisterHostnameProvider("dummy", dummyProvider)
	defer delete(providerCatalog, "dummy")

	name, err := GetHostname(testutil.TestNodeForProviderID("dummy://dummy-hostname"))
	assert.NoError(t, err)
	assert.Equal(t, "dummy-hostname", name)
}

func TestGetHostnameUnknown(t *testing.T) {
	_, err := GetHostname(testutil.TestNodeForProviderID("dummy"))
	assert.Error(t, err)
}

func TestGetHostnameNoName(t *testing.T) {
	RegisterHostnameProvider("dummy", dummyErrorProvider)
	defer delete(providerCatalog, "dummy")

	_, err := GetHostname(testutil.TestNodeForProviderID("dummy://"))
	assert.Error(t, err)
}

func TestGetHostnameInvalid(t *testing.T) {
	RegisterHostnameProvider("dummy", dummyInvalidProvider)
	defer delete(providerCatalog, "dummy")

	_, err := GetHostname(testutil.TestNodeForProviderID("dummy://dummy-hostname"))
	assert.Error(t, err)
}
