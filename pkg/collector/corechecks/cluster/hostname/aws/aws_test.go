package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAwsHostnameUsesInstanceId(t *testing.T) {
	assert.Equal(t, "i-09e7ef36c4efd7cbe", GetHostname("aws:///eu-west-1b/i-09e7ef36c4efd7cbe"))
}

func TestAwsHostnameEmpty(t *testing.T) {
	assert.Equal(t, "", GetHostname(""))
}

func TestAwsHostnameWrongPrefix(t *testing.T) {
	assert.Equal(t, "", GetHostname("abc://test/test"))
}
