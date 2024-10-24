// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2017-present Datadog, Inc.

package corechecks

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/StackVista/stackstate-agent/pkg/aggregator/mocksender"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/defaults"
)

var (
	initConfig       = `foo: bar`
	defaultsInstance = `foo_init: bar_init`
	customInstance   = `
foo_init: bar_init
collection_interval: 60
empty_default_hostname: true
name: foobar
`
	// [sts] additional test for backwards compatibility
	legacyInstance   = `
foo_init: bar_init
min_collection_interval: 60
empty_default_hostname: true
name: foobar
`
	// [sts] additional test when legacy and new collection interval are both defined
	legacyInstanceClash   = `
foo_init: bar_init
collection_interval: 30
min_collection_interval: 60
empty_default_hostname: true
name: foobar
`
)

type dummyCheck struct {
	CheckBase
}

func TestCommonConfigure(t *testing.T) {
	checkName := "test"
	mycheck := &dummyCheck{
		CheckBase: NewCheckBase(checkName),
	}
	mockSender := mocksender.NewMockSender(mycheck.ID())

	err := mycheck.CommonConfigure([]byte(defaultsInstance), "test")
	assert.NoError(t, err)
	assert.Equal(t, defaults.DefaultCheckInterval, mycheck.Interval())
	mockSender.AssertNumberOfCalls(t, "DisableDefaultHostname", 0)

	mockSender.On("DisableDefaultHostname", true).Return().Once()
	err = mycheck.CommonConfigure([]byte(customInstance), "test")
	assert.NoError(t, err)
	assert.Equal(t, 60*time.Second, mycheck.Interval())
	mycheck.BuildID([]byte(customInstance), []byte(initConfig))
	assert.Equal(t, string(mycheck.ID()), "test:foobar:19e743a262c48886")
	mockSender.AssertExpectations(t)
}

func TestCommonConfigureCustomID(t *testing.T) {
	checkName := "test"
	mycheck := &dummyCheck{
		CheckBase: NewCheckBase(checkName),
	}
	mycheck.BuildID([]byte(customInstance), nil)
	assert.NotEqual(t, checkName, string(mycheck.ID()))
	mockSender := mocksender.NewMockSender(mycheck.ID())

	mockSender.On("DisableDefaultHostname", true).Return().Once()
	err := mycheck.CommonConfigure([]byte(customInstance), "test")
	assert.NoError(t, err)
	assert.Equal(t, 60*time.Second, mycheck.Interval())
	mycheck.BuildID([]byte(customInstance), []byte(initConfig))
	assert.Equal(t, string(mycheck.ID()), "test:foobar:19e743a262c48886")
	mockSender.AssertExpectations(t)
}

// [sts] Tests whether we are backwards compatible with MinCollectionInterval
func TestCommonConfigureMinCollectionInterval(t *testing.T) {
	checkName := "test"
	mycheck := &dummyCheck{
		CheckBase: NewCheckBase(checkName),
	}
	mycheck.BuildID([]byte(legacyInstance), nil)
	assert.NotEqual(t, checkName, string(mycheck.ID()))
	mockSender := mocksender.NewMockSender(mycheck.ID())

	mockSender.On("DisableDefaultHostname", true).Return().Once()
	err := mycheck.CommonConfigure([]byte(legacyInstance), "test")
	assert.NoError(t, err)
	assert.Equal(t, 60*time.Second, mycheck.Interval())
}

// [sts] Tests what happens when backwards compatibility clashes
func TestCommonConfigureClashMinCollectionInterval(t *testing.T) {
	checkName := "test"
	mycheck := &dummyCheck{
		CheckBase: NewCheckBase(checkName),
	}
	mycheck.BuildID([]byte(legacyInstanceClash), nil)
	assert.NotEqual(t, checkName, string(mycheck.ID()))
	mockSender := mocksender.NewMockSender(mycheck.ID())

	mockSender.On("DisableDefaultHostname", true).Return().Once()
	err := mycheck.CommonConfigure([]byte(legacyInstanceClash), "test")
	assert.NoError(t, err)
	assert.Equal(t, 30*time.Second, mycheck.Interval())
}
