// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubelet

package kubelet

import (
	"os"
	"testing"

	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestGetNodeNameForFieldSelector(t *testing.T) {
	// Save and restore environment
	originalEnvVars := map[string]string{
		"STS_KUBERNETES_KUBELET_NODENAME": os.Getenv("STS_KUBERNETES_KUBELET_NODENAME"),
		"DD_KUBERNETES_KUBELET_NODENAME":  os.Getenv("DD_KUBERNETES_KUBELET_NODENAME"),
		"KUBERNETES_HOSTNAME":             os.Getenv("KUBERNETES_HOSTNAME"),
		"K8S_NODE_NAME":                   os.Getenv("K8S_NODE_NAME"),
	}
	defer func() {
		for key, value := range originalEnvVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	tests := []struct {
		name           string
		configValue    string
		envVars        map[string]string
		expectedResult string
	}{
		{
			name:           "node name from config",
			configValue:    "config-node-1",
			envVars:        map[string]string{},
			expectedResult: "config-node-1",
		},
		{
			name:        "node name from STS env var",
			configValue: "",
			envVars: map[string]string{
				"STS_KUBERNETES_KUBELET_NODENAME": "sts-node-1",
			},
			expectedResult: "sts-node-1",
		},
		{
			name:        "node name from KUBERNETES_HOSTNAME",
			configValue: "",
			envVars: map[string]string{
				"KUBERNETES_HOSTNAME": "k8s-hostname-node-1",
			},
			expectedResult: "k8s-hostname-node-1",
		},
		{
			name:        "config takes precedence",
			configValue: "config-node-1",
			envVars: map[string]string{
				"STS_KUBERNETES_KUBELET_NODENAME": "sts-node-1",
			},
			expectedResult: "config-node-1",
		},
		{
			name:           "no node name available",
			configValue:    "",
			envVars:        map[string]string{},
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Unsetenv("STS_KUBERNETES_KUBELET_NODENAME")
			os.Unsetenv("DD_KUBERNETES_KUBELET_NODENAME")
			os.Unsetenv("KUBERNETES_HOSTNAME")
			os.Unsetenv("K8S_NODE_NAME")

			// Set test environment
			for key, value := range tt.envVars {
				if value != "" {
					os.Setenv(key, value)
				}
			}

			// Mock config
			mockConfig := config.Mock(t)
			if tt.configValue != "" {
				mockConfig.SetWithoutSource("kubernetes_kubelet_nodename", tt.configValue)
			}

			result := getNodeNameForFieldSelector()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestBuildPodListPathWithFieldSelector(t *testing.T) {
	tests := []struct {
		name         string
		nodeName     string
		expectedPath string
	}{
		{
			name:         "with node name",
			nodeName:     "worker-node-1",
			expectedPath: "/pods?fieldSelector=spec.nodeName=worker-node-1",
		},
		{
			name:         "with special characters",
			nodeName:     "node-1.example.com",
			expectedPath: "/pods?fieldSelector=spec.nodeName=node-1.example.com",
		},
		{
			name:         "empty node name",
			nodeName:     "",
			expectedPath: "/pods",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPodListPathWithFieldSelector(tt.nodeName)
			assert.Equal(t, tt.expectedPath, result)
		})
	}
}

func TestResetPodListQueryMethod(t *testing.T) {
	// Set a method
	detectedMethodMutex.Lock()
	detectedMethod = methodWithFieldSelector
	detectedMethodMutex.Unlock()

	// Verify it's set
	detectedMethodMutex.RLock()
	assert.Equal(t, methodWithFieldSelector, detectedMethod)
	detectedMethodMutex.RUnlock()

	// Reset it
	ResetPodListQueryMethod()

	// Verify it's reset
	detectedMethodMutex.RLock()
	assert.Equal(t, methodUnknown, detectedMethod)
	detectedMethodMutex.RUnlock()
}
