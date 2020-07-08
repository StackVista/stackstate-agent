// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package compliance

import (
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

const testResourceFile = `
file:
  path: /etc/docker/daemon.json
condition: file.owner == "root"
`

const testResourceFileReportingPermissions = `
file:
  path: /etc/docker/daemon.json
  report:
  - property: permissions
    kind: attribute
`

const testResourceFilePathFromCommand = `
file:
  pathFrom:
  - command:
      shell:
        run: systemctl show -p FragmentPath docker.service
  report:
  - property: owner
    kind: attribute
`

const testResourceFileReportingJSONPath = `
file:
  path: /etc/docker/daemon.json
  report:
  - property: tlsverify
    kind: jsonquery
`
const testResourceProcessWithFallback = `
process:
  name: dockerd
condition: process.flag("--tlsverify") != ""
fallback:
  condition: >-
    !process.hasFlag("--tlsverify")
  resource:
    file:
      path: /etc/docker/daemon.json
    condition: file.jq(".tlsverify") == "true"
`

const testResourceCommand = `
command:
  shell:
    run: mountpoint -- "$(docker info -f '{{ .DockerRootDir }}')"
condition: command.exitCode == 0
`

const testResourceAudit = `
audit:
  path: /usr/bin/dockerd
  report:
  - property: enabled
    kind: attribute
`

const testResourceAuditPathFromCommand = `
audit:
  pathFrom:
  - command:
      shell:
        run: systemctl show -p FragmentPath docker.socket
  report:
  - property: enabled
    kind: attribute
`

const testResourceGroup = `
group:
  name: docker
condition: >-
  "root" in group.users
`

const testResourceDockerImage = `
docker:
  kind: image
condition: docker.template("{{ $.Config.Healthcheck }}") != ""
`

func TestResources(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Resource
	}{
		{
			name:  "file",
			input: testResourceFile,
			expected: Resource{
				File: &File{
					Path: `/etc/docker/daemon.json`,
				},
				Condition: `file.owner == "root"`,
			},
		},
		{
			name:  "process",
			input: testResourceProcess,
			expected: Resource{
				File: &File{
					PathFrom: ValueFrom{
						{
							Command: &ValueFromCommand{
								ShellCmd: &ShellCmd{
									Run: `systemctl show -p FragmentPath docker.service`,
								},
							},
						},
					},
					Report: Report{
						{
							Property: "owner",
							Kind:     PropertyKindAttribute,
						},
					},
				},
			},
		},
		{
			name:  "file reporting jsonquery property",
			input: testResourceFileReportingJSONPath,
			expected: Resource{
				File: &File{
					Path: `/etc/docker/daemon.json`,
					Report: Report{
						{
							Property: "tlsverify",
							Kind:     PropertyKindJSONQuery,
						},
					},
				},
				Condition: `process.flag("--tlsverify") != ""`,
			},
		},
		{
			name:  "process with fallback",
			input: testResourceProcessWithFallback,
			expected: Resource{
				Process: &Process{
					Name: "dockerd",
				},
				Condition: `process.flag("--tlsverify") != ""`,
				Fallback: &Fallback{
					Condition: `!process.hasFlag("--tlsverify")`,
					Resource: Resource{
						File: &File{
							Path: `/etc/docker/daemon.json`,
						},
						Condition: `file.jq(".tlsverify") == "true"`,
					},
				},
			},
		},
		{
			name:  "command",
			input: testResourceCommand,
			expected: Resource{
				Command: &Command{
					ShellCmd: &ShellCmd{
						Run: `mountpoint -- "$(docker info -f '{{ .DockerRootDir }}')"`,
					},
				},
				Condition: `command.exitCode == 0`,
			},
		},
		{
			name:  "audit",
			input: testResourceAudit,
			expected: Resource{
				Audit: &Audit{
					Path: "/usr/bin/dockerd",
					Report: Report{
						{
							Property: "enabled",
							Kind:     "attribute",
						},
					},
				},
			},
		},
		{
			name:  "audit with file path from command",
			input: testResourceAuditPathFromCommand,
			expected: Resource{
				Audit: &Audit{
					PathFrom: ValueFrom{
						{
							Command: &ValueFromCommand{
								ShellCmd: &ShellCmd{
									Run: `systemctl show -p FragmentPath docker.socket`,
								},
							},
						},
					},
					Report: Report{
						{
							Property: "enabled",
							Kind:     "attribute",
						},
					},
				},
				Condition: `audit.enabled`,
			},
		},
		{
			name:  "group",
			input: testResourceGroup,
			expected: Resource{
				Group: &Group{
					Name: "docker",
				},
				Condition: `"root" in group.users`,
			},
		},
		{

			name:  "docker image",
			input: testResourceDockerImage,
			expected: Resource{
				Docker: &DockerResource{
					Kind: "image",
				},
				Condition: `docker.template("{{ $.Config.Healthcheck }}") != ""`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var r Resource
			err := yaml.Unmarshal([]byte(test.input), &r)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, r)
		})
	}

}
