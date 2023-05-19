package gce

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGceHostnameUsesInstanceId(t *testing.T) {
	assert.Equal(t, "gke-test-default-pool-9f8f65a4-2kld.europe-west4-a.c.test-stackstate.internal", GetHostname("gce://test-stackstate/europe-west4-a/gke-test-default-pool-9f8f65a4-2kld"))
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
