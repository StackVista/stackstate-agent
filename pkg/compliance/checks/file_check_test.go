// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

//go:build !windows
// +build !windows

package checks

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/StackVista/stackstate-agent/pkg/compliance"
	"github.com/StackVista/stackstate-agent/pkg/compliance/mocks"

	"github.com/stretchr/testify/mock"
	assert "github.com/stretchr/testify/require"
)

func TestFileCheck(t *testing.T) {
	type setupFunc func(t *testing.T, env *mocks.Env) *fileCheck
	type validateFunc func(t *testing.T, kv compliance.KVMap)

	setupFile := func(file *compliance.File) setupFunc {
		return func(t *testing.T, env *mocks.Env) *fileCheck {
			if file.Path != "" {
				env.On("NormalizePath", file.Path).Return(file.Path)
			}

			return &fileCheck{
				baseCheck: newTestBaseCheck(env, checkKindFile),
				file:      file,
			}
		}

		return dir, paths
	}

	tests := []struct {
		name        string
		resource    compliance.Resource
		setup       setupFileFunc
		validate    validateFunc
		expectError error
	}{
		{
			name: "permissions",
			setup: func(t *testing.T, env *mocks.Env) *fileCheck {
				dir := os.TempDir()

				fileName := fmt.Sprintf("test-permissions-file-check-%d.dat", time.Now().Unix())
				filePath := path.Join(dir, fileName)

				f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
				defer f.Close()
				assert.NoError(t, err)

				env.On("NormalizePath", fileName).Return(filePath)

				file := &compliance.File{
					Path: fileName,
					Report: compliance.Report{
						{
							Property: "permissions",
							Kind:     compliance.PropertyKindAttribute,
						},
					},
				}
				return &fileCheck{
					baseCheck: newTestBaseCheck(env, checkKindFile),
					file:      file,
				}

				env.On("NormalizeToHostRoot", file.Path).Return(path.Join(tempDir, "/*.dat"))
			},
			validate: func(t *testing.T, file *compliance.File, report *compliance.Report) {
				assert.True(report.Passed)
				assert.Regexp("/etc/test-[0-9]-[0-9]+", report.Data["file.path"])
				assert.Equal(uint64(0644), report.Data["file.permissions"])
			},
		},
		{
			name: "file user and group",
			resource: compliance.Resource{
				File: &compliance.File{
					Path: "/tmp",
				},
				Condition: `file.user == "root" && file.group in ["root", "wheel"]`,
			},
			setup: normalizePath,
			validate: func(t *testing.T, file *compliance.File, report *compliance.Report) {
				assert.True(report.Passed)
				assert.Equal("/tmp", report.Data["file.path"])
				assert.Equal("root", report.Data["file.user"])
				assert.Contains([]string{"root", "wheel"}, report.Data["file.group"])
			},
		},
		{
			name: "jq(log-driver) - passed",
			resource: compliance.Resource{
				File: &compliance.File{
					Path: "/etc/docker/daemon.json",
				},
				Condition: `file.jq(".\"log-driver\"") == "json-file"`,
			},
			setup: func(t *testing.T, env *mocks.Env, file *compliance.File) {
				env.On("NormalizeToHostRoot", file.Path).Return("./testdata/file/daemon.json")
				env.On("RelativeToHostRoot", "./testdata/file/daemon.json").Return(file.Path)
			},
			validate: func(t *testing.T, file *compliance.File, report *compliance.Report) {
				assert.True(report.Passed)
				assert.Equal("/etc/docker/daemon.json", report.Data["file.path"])
				assert.NotEmpty(report.Data["file.user"])
				assert.NotEmpty(report.Data["file.group"])
			},
		},
		{
			name: "jq(experimental) - failed",
			resource: compliance.Resource{
				File: &compliance.File{
					Path: "/etc/docker/daemon.json",
				},
				Condition: `file.jq(".experimental") == "true"`,
			},
			setup: func(t *testing.T, env *mocks.Env, file *compliance.File) {
				env.On("NormalizeToHostRoot", file.Path).Return("./testdata/file/daemon.json")
				env.On("RelativeToHostRoot", "./testdata/file/daemon.json").Return(file.Path)
			},
			validate: func(t *testing.T, file *compliance.File, report *compliance.Report) {
				assert.False(report.Passed)
				assert.Equal("/etc/docker/daemon.json", report.Data["file.path"])
				assert.NotEmpty(report.Data["file.user"])
				assert.NotEmpty(report.Data["file.group"])
			},
		},
		{
			name: "jq(experimental) and path expression",
			resource: compliance.Resource{
				File: &compliance.File{
					Path: `process.flag("dockerd", "--config-file")`,
				},
				Condition: `file.jq(".experimental") == "false"`,
			},
			setup: func(t *testing.T, env *mocks.Env, file *compliance.File) {
				path := "/etc/docker/daemon.json"
				env.On("EvaluateFromCache", mock.Anything).Return(path, nil)
				env.On("NormalizeToHostRoot", path).Return("./testdata/file/daemon.json")
				env.On("RelativeToHostRoot", "./testdata/file/daemon.json").Return(path)
			},
			validate: func(t *testing.T, file *compliance.File, report *compliance.Report) {
				assert.True(report.Passed)
				assert.Equal("/etc/docker/daemon.json", report.Data["file.path"])
				assert.NotEmpty(report.Data["file.user"])
				assert.NotEmpty(report.Data["file.group"])
			},
		},
		{
			name: "jq(experimental) and path expression - empty path",
			resource: compliance.Resource{
				File: &compliance.File{
					Path: `process.flag("dockerd", "--config-file")`,
				},
				Condition: `file.jq(".experimental") == "false"`,
			},
			setup: func(t *testing.T, env *mocks.Env, file *compliance.File) {
				env.On("EvaluateFromCache", mock.Anything).Return("", nil)
			},
			expectError: errors.New(`failed to resolve path: empty path from process.flag("dockerd", "--config-file")`),
		},
		{
			name: "jq(experimental) and path expression - wrong type",
			resource: compliance.Resource{
				File: &compliance.File{
					Path: `process.flag("dockerd", "--config-file")`,
				},
				Condition: `file.jq(".experimental") == "false"`,
			},
			setup: func(t *testing.T, env *mocks.Env, file *compliance.File) {
				env.On("EvaluateFromCache", mock.Anything).Return(true, nil)
			},
			expectError: errors.New(`failed to resolve path: expected string from process.flag("dockerd", "--config-file") got "true"`),
		},
		{
			name: "jq(experimental) and path expression - expression failed",
			resource: compliance.Resource{
				File: &compliance.File{
					Path: `process.unknown()`,
				},
				Condition: `file.jq(".experimental") == "false"`,
			},
			setup: func(t *testing.T, env *mocks.Env, file *compliance.File) {
				env.On("EvaluateFromCache", mock.Anything).Return(nil, errors.New("1:1: unknown function process.unknown()"))
			},
			expectError: errors.New(`failed to resolve path: 1:1: unknown function process.unknown()`),
		},
		{
			name: "jq(ulimits)",
			resource: compliance.Resource{
				File: &compliance.File{
					Path: "/etc/docker/daemon.json",
				},
				Condition: `file.jq(".[\"default-ulimits\"].nofile.Hard") == "64000"`,
			},
			setup: func(t *testing.T, env *mocks.Env, file *compliance.File) {
				env.On("NormalizeToHostRoot", file.Path).Return("./testdata/file/daemon.json")
				env.On("RelativeToHostRoot", "./testdata/file/daemon.json").Return(file.Path)
			},
			validate: func(t *testing.T, file *compliance.File, report *compliance.Report) {
				assert.True(report.Passed)
				assert.Equal("/etc/docker/daemon.json", report.Data["file.path"])
				assert.NotEmpty(report.Data["file.user"])
				assert.NotEmpty(report.Data["file.group"])
			},
		},
		{
			name: "jsonquery experimental (pathFrom)",
			setup: func(t *testing.T, env *mocks.Env) *fileCheck {
				file := &compliance.File{
					PathFrom: compliance.ValueFrom{
						{
							Process: &compliance.ValueFromProcess{
								Name: "dockerd",
								Flag: "--config-file",
							},
						},
					},
					Report: compliance.Report{
						{
							Property: ".experimental",
							Kind:     "jsonquery",
							As:       "experimental",
						},
					},
				}

				path := "./testdata/file/daemon.json"
				env.On("ResolveValueFrom", file.PathFrom).Return(path, nil)
				env.On("NormalizePath", path).Return(path)

				return setupFile(file)(t, env)
			},
			validate: func(t *testing.T, kv compliance.KVMap) {
				assert.Equal(t, compliance.KVMap{
					"experimental": "false",
				}, kv)
			},
		},
		{
			name: "jsonquery ulimits",
			setup: setupFile(&compliance.File{
				Path: "./testdata/file/daemon.json",
				Report: compliance.Report{
					{
						Property: `.["default-ulimits"].nofile.Hard`,
						Kind:     "jsonquery",
						As:       "nofile_hard",
					},
				},
				Condition: `file.yaml(".apiVersion") == "v1"`,
			},
			setup: func(t *testing.T, env *mocks.Env, file *compliance.File) {
				env.On("NormalizeToHostRoot", file.Path).Return("./testdata/file/pod.yaml")
				env.On("RelativeToHostRoot", "./testdata/file/pod.yaml").Return(file.Path)
			},
			validate: func(t *testing.T, file *compliance.File, report *compliance.Report) {
				assert.True(report.Passed)
				assert.Equal("/etc/pod.yaml", report.Data["file.path"])
				assert.NotEmpty(report.Data["file.user"])
				assert.NotEmpty(report.Data["file.group"])
			},
		},
		{
			name: "regexp",
			resource: compliance.Resource{
				File: &compliance.File{
					Path: "/proc/mounts",
				},
				Condition: `file.regexp("[a-zA-Z0-9-_/]+ /boot/efi [a-zA-Z0-9-_/]+") != ""`,
			},
			setup: func(t *testing.T, env *mocks.Env, file *compliance.File) {
				env.On("NormalizeToHostRoot", file.Path).Return("./testdata/file/mounts")
				env.On("RelativeToHostRoot", "./testdata/file/mounts").Return(file.Path)
			},
			validate: func(t *testing.T, file *compliance.File, report *compliance.Report) {
				assert.True(report.Passed)
				assert.Equal("/proc/mounts", report.Data["file.path"])
				assert.NotEmpty(report.Data["file.user"])
				assert.NotEmpty(report.Data["file.group"])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reporter := &mocks.Reporter{}
			defer reporter.AssertExpectations(t)

			env := &mocks.Env{}
			defer env.AssertExpectations(t)

			env.On("Reporter").Return(reporter)

			fc := test.setup(t, env)

			reporter.On(
				"Report",
				mock.AnythingOfType("*compliance.RuleEvent"),
			).Run(func(args mock.Arguments) {
				event := args.Get(0).(*compliance.RuleEvent)
				test.validate(t, event.Data)
			})

			err := fc.Run()
			assert.NoError(t, err)
		})
	}

	for _, dir := range cleanUpDirs {
		os.RemoveAll(dir)
	}
}
