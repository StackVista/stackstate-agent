// Code generated by mockery v2.16.0. DO NOT EDIT.

package mocks

import (
	time "time"

	procutil "github.com/DataDog/datadog-agent/pkg/process/procutil"
	mock "github.com/stretchr/testify/mock"
)

// Probe is an autogenerated mock type for the Probe type
type Probe struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Probe) Close() {
	_m.Called()
}

// ProcessesByPID provides a mock function with given fields: now, collectStats
func (_m *Probe) ProcessesByPID(now time.Time, collectStats bool) (map[int32]*procutil.Process, error) {
	ret := _m.Called(now, collectStats)

	var r0 map[int32]*procutil.Process
	if rf, ok := ret.Get(0).(func(time.Time, bool) map[int32]*procutil.Process); ok {
		r0 = rf(now, collectStats)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[int32]*procutil.Process)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(time.Time, bool) error); ok {
		r1 = rf(now, collectStats)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StatsForPIDs provides a mock function with given fields: pids, now
func (_m *Probe) StatsForPIDs(pids []int32, now time.Time) (map[int32]*procutil.Stats, error) {
	ret := _m.Called(pids, now)

	var r0 map[int32]*procutil.Stats
	if rf, ok := ret.Get(0).(func([]int32, time.Time) map[int32]*procutil.Stats); ok {
		r0 = rf(pids, now)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[int32]*procutil.Stats)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]int32, time.Time) error); ok {
		r1 = rf(pids, now)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StatsWithPermByPID provides a mock function with given fields: pids
func (_m *Probe) StatsWithPermByPID(pids []int32) (map[int32]*procutil.StatsWithPerm, error) {
	ret := _m.Called(pids)

	var r0 map[int32]*procutil.StatsWithPerm
	if rf, ok := ret.Get(0).(func([]int32) map[int32]*procutil.StatsWithPerm); ok {
		r0 = rf(pids)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[int32]*procutil.StatsWithPerm)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]int32) error); ok {
		r1 = rf(pids)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewProbe interface {
	mock.TestingT
	Cleanup(func())
}

// NewProbe creates a new instance of Probe. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewProbe(t mockConstructorTestingTNewProbe) *Probe {
	mock := &Probe{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}