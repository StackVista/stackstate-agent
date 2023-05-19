package azure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: This is not properly implemented yet, it should implement and test the different cases from
// github.com/StackVista/stackstate-agent/pkg/util/cloudproviders/azure
func TestGceHostnameUsesInstanceId(t *testing.T) {
	assert.Equal(t, "", GetHostname("azure://test/test"))
}

func TestGceHostnameEmpty(t *testing.T) {
	assert.Equal(t, "", GetHostname(""))
}

func TestGceHostnameWrongPrefix(t *testing.T) {
	assert.Equal(t, "", GetHostname("abc://test/test"))
}
