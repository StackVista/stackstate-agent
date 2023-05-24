package gce

import (
	"context"
	"testing"

	"github.com/StackVista/stackstate-agent/pkg/util/hostname"
	"github.com/stretchr/testify/assert"
)

const (
	ZonedHostname  = "my-node.europe-west4-a.c.test-stackstate.internal"
	GlobalHostname = "my-node.c.test-stackstate.internal"
)

func TestZonalGceHostnameUsesProviderID(t *testing.T) {
	withStaticHostnameProvider("gce", ZonedHostname, func() {
		assert.Equal(t, "gke-test-default-pool-9f8f65a4-2kld.europe-west4-a.c.test-stackstate.internal", GetHostname("gce://test-stackstate/europe-west4-a/gke-test-default-pool-9f8f65a4-2kld"))
	})
}

func TestGlobalGceHostnameUsesProviderID(t *testing.T) {
	withStaticHostnameProvider("gce", GlobalHostname, func() {
		assert.Equal(t, "gke-test-default-pool-9f8f65a4-2kld.c.test-stackstate.internal", GetHostname("gce://test-stackstate/europe-west4-a/gke-test-default-pool-9f8f65a4-2kld"))
	})
}

func TestGceHostnameEmpty(t *testing.T) {
	assert.Equal(t, "", GetHostname(""))
}

func TestGceHostnameWrongFormat(t *testing.T) {
	assert.Equal(t, "", GetHostname("gce://test-stackstate/europe-west4-a/gke-test-default-pool-9f8f65a4-2kld/extra"))
	assert.Equal(t, "", GetHostname("gce://test-stackstate/europe-west4-a"))
}

func TestGceHostnameWrongPrefix(t *testing.T) {
	assert.Equal(t, "", GetHostname("abc://test/test"))
}

func withStaticHostnameProvider(providerID, staticHost string, f func()) {
	originalHostnameProvider := hostname.GetProvider(providerID)
	hostname.RegisterHostnameProvider(providerID, func(ctx context.Context, options map[string]interface{}) (string, error) {
		return staticHost, nil
	})
	defer func() { hostname.RegisterHostnameProvider(providerID, originalHostnameProvider) }()
	f()
}
