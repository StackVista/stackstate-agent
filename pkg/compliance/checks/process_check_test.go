// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.
package checks

import (
	"testing"

	"github.com/DataDog/datadog-agent/pkg/compliance"
	"github.com/DataDog/datadog-agent/pkg/compliance/event"
	"github.com/DataDog/datadog-agent/pkg/compliance/mocks"
	"github.com/DataDog/datadog-agent/pkg/util/cache"

	assert "github.com/stretchr/testify/require"
)

type processFixture struct {
	name    string
	process *compliance.Process

	processes map[int32]*process.FilledProcess
	expKV     event.Data
	expError  error
}

func (f *processFixture) run(t *testing.T) {
	t.Helper()
	assert := assert.New(t)

	cache.Cache.Delete(processCacheKey)
	processFetcher = func() (map[int32]*process.FilledProcess, error) {
		return f.processes, nil
	}

	reporter := &mocks.Reporter{}
	defer reporter.AssertExpectations(t)

	env := &mocks.Env{}
	defer env.AssertExpectations(t)

	if f.expKV != nil {
		env.On("Reporter").Return(reporter)
		reporter.On(
			"Report",
			newTestRuleEvent(
				[]string{"check_kind:process"},
				f.expKV,
			),
		).Once()
	}

	check, err := newProcessCheck(newTestBaseCheck(env, checkKindProcess), f.process)
	assert.NoError(t, err)

	err = check.Run()
	assert.Equal(t, f.expError, err)
}

func TestProcessCheck(t *testing.T) {
	tests := []processFixture{
		{
			name: "Simple case",
			process: &compliance.Process{
				Name: "proc1",
				Report: compliance.Report{
					{
						Kind:     "flag",
						Property: "--path",
						As:       "path",
					},
				},
				Condition: `process.flag("--path") == "foo"`,
			},
			processes: processes{
				42: {
					Name:    "proc1",
					Cmdline: []string{"arg1", "--path=foo"},
				},
			},
			expKV: event.Data{
				"path": "foo",
			},
		},
		{
			name: "Process not found",
			process: &compliance.Process{
				Name: "proc1",
				Report: compliance.Report{
					{
						Kind:     "flag",
						Property: "--path",
						As:       "path",
					},
				},
			},
			processes: processes{
				42: {
					Name:    "proc1",
					Cmdline: []string{"arg1"},
				},
				38: {
					Name:    "proc2",
					Cmdline: []string{"arg1", "--tlsverify"},
				},
			},
			expectReport: &compliance.Report{
				Passed: true,
				Data: event.Data{
					"process.name":    "proc2",
					"process.exe":     "",
					"process.cmdLine": []string{"arg1", "--tlsverify"},
				},
			},
		},
		{
			name: "Argument not found",
			process: &compliance.Process{
				Name: "proc1",
				Report: compliance.Report{
					{
						Kind:     "flag",
						Property: "--path",
						As:       "path",
					},
				},
				Condition: `process.flag("--path") == "foo"`,
			},
			processes: processes{
				42: {
					Name:    "proc2",
					Cmdline: []string{"arg1", "--path=foo"},
				},
				43: {
					Name:    "proc3",
					Cmdline: []string{"arg1", "--path=foo"},
				},
			},
			expectReport: &compliance.Report{
				Passed: false,
			},
		},
		{
			name: "Override returned value",
			process: &compliance.Process{
				Name: "proc1",
				Report: compliance.Report{
					{
						Kind:     "flag",
						Property: "--verbose",
						As:       "verbose",
						Value:    "true",
					},
				},
				Condition: `process.flag("--path") == "foo"`,
			},
			processes: processes{
				42: {
					Name:    "proc1",
					Cmdline: []string{"arg1", "--paths=foo"},
				},
			},
			expKV: event.Data{
				"verbose": "true",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestProcessCheckCache(t *testing.T) {
	// Run first fixture, populating cache
	firstContent := processFixture{
		name: "simple case",
		resource: compliance.Resource{
			Process: &compliance.Process{
				Name: "proc1",
			},
			Condition: `process.flag("--path") == "foo"`,
		},
		processes: processes{
			42: {
				Name:    "proc1",
				Cmdline: []string{"arg1", "--path=foo"},
			},
		},
		expectReport: &compliance.Report{
			Passed: true,
			Data: event.Data{
				"process.name":    "proc1",
				"process.exe":     "",
				"process.cmdLine": []string{"arg1", "--path=foo"},
			},
		},
	}
	firstContent.run(t)

	// Run second fixture, using cache
	secondFixture := processFixture{
		name: "simple case",
		resource: compliance.Resource{
			Process: &compliance.Process{
				Name: "proc1",
			},
			Condition: `process.flag("--path") == "foo"`,
		},
		useCache: true,
		expectReport: &compliance.Report{
			Passed: true,
			Data: event.Data{
				"process.name":    "proc1",
				"process.exe":     "",
				"process.cmdLine": []string{"arg1", "--path=foo"},
			},
		},
	}
	secondFixture.run(t)
}
