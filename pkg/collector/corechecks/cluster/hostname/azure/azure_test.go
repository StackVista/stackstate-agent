package azure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: This is not properly implemented yet, it should implement and test the different cases from
// github.com/StackVista/stackstate-agent/pkg/util/cloudproviders/azure
func TestAzureHostnameUsesSuffix(t *testing.T) {
	assert.Equal(t, "test", GetHostname("azure://test/test"))
}

func TestAzureHostnameEmpty(t *testing.T) {
	assert.Equal(t, "", GetHostname(""))
}
