// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubelet

package kubelet

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// podListQueryMethod represents which method to use for querying the pod list
type podListQueryMethod int

const (
	// methodUnknown means we haven't determined which method works yet
	methodUnknown podListQueryMethod = iota
	// methodWithFieldSelector uses the Kubernetes 1.34+ field selector approach
	methodWithFieldSelector
	// methodLegacy uses the pre-1.34 approach without field selector
	methodLegacy
)

var (
	// detectedMethod caches which method works for this cluster
	detectedMethod      podListQueryMethod
	detectedMethodMutex sync.RWMutex
)

// getNodeNameForFieldSelector retrieves the node name from configuration or environment.
// Returns empty string if not configured.
func getNodeNameForFieldSelector() string {
	// Try config first
	if nodeName := config.Datadog.GetString("kubernetes_kubelet_nodename"); nodeName != "" {
		return nodeName
	}

	// Try various environment variables
	envVars := []string{
		"STS_KUBERNETES_KUBELET_NODENAME",
		"DD_KUBERNETES_KUBELET_NODENAME",
		"KUBERNETES_HOSTNAME",
		"K8S_NODE_NAME",
	}

	for _, envVar := range envVars {
		if nodeName := os.Getenv(envVar); nodeName != "" {
			return nodeName
		}
	}

	return ""
}

// buildPodListPathWithFieldSelector constructs the /pods path with field selector for K8s 1.34+
func buildPodListPathWithFieldSelector(nodeName string) string {
	if nodeName == "" {
		return kubeletPodPath
	}
	fieldSelector := fmt.Sprintf("spec.nodeName=%s", url.QueryEscape(nodeName))
	return fmt.Sprintf("%s?fieldSelector=%s", kubeletPodPath, fieldSelector)
}

// queryPodListWithCompatibility queries the pod list endpoint with automatic fallback
// for backward compatibility with Kubernetes < 1.34.
//
// It first attempts to use the Kubernetes 1.34+ field selector method.
// If that fails with a 403 Forbidden error, it falls back to the legacy method.
// The working method is cached for subsequent requests.
func (ku *KubeUtil) queryPodListWithCompatibility(ctx context.Context) ([]byte, int, error) {
	// Check if we've already determined which method works
	detectedMethodMutex.RLock()
	method := detectedMethod
	detectedMethodMutex.RUnlock()

	switch method {
	case methodWithFieldSelector:
		// We know field selector works, use it directly
		return ku.queryPodListWithFieldSelectorMethod(ctx)

	case methodLegacy:
		// We know legacy method is needed, use it directly
		return ku.queryPodListLegacyMethod(ctx)

	case methodUnknown:
		// First time or not yet determined, try with field selector first
		return ku.queryPodListWithAutoDetection(ctx)
	}

	return nil, 0, fmt.Errorf("invalid pod list query method")
}

// queryPodListWithAutoDetection attempts field selector first, falls back to legacy if needed
func (ku *KubeUtil) queryPodListWithAutoDetection(ctx context.Context) ([]byte, int, error) {
	nodeName := getNodeNameForFieldSelector()

	if nodeName != "" {
		// Try the Kubernetes 1.34+ method with field selector
		log.Debugf("Attempting pod list query with field selector for node: %s", nodeName)
		data, code, err := ku.queryPodListWithFieldSelectorMethod(ctx)

		// If it worked, cache this method and return
		if err == nil && code == http.StatusOK {
			log.Infof("Pod list query with field selector succeeded - will use this method for Kubernetes 1.34+")
			detectedMethodMutex.Lock()
			detectedMethod = methodWithFieldSelector
			detectedMethodMutex.Unlock()
			return data, code, err
		}

		// If it failed with Forbidden, try legacy method
		if code == http.StatusForbidden {
			log.Infof("Pod list query with field selector returned 403 Forbidden - falling back to legacy method")
		} else if err != nil {
			log.Debugf("Pod list query with field selector failed: %v - trying legacy method", err)
		}
	} else {
		log.Debugf("Node name not configured - trying legacy pod list query method")
	}

	// Try legacy method
	data, code, err := ku.queryPodListLegacyMethod(ctx)

	// If legacy method worked, cache it
	if err == nil && code == http.StatusOK {
		log.Infof("Legacy pod list query succeeded - will use this method for Kubernetes < 1.34")
		detectedMethodMutex.Lock()
		detectedMethod = methodLegacy
		detectedMethodMutex.Unlock()
	}

	return data, code, err
}

// queryPodListWithFieldSelectorMethod queries pod list using Kubernetes 1.34+ field selector
func (ku *KubeUtil) queryPodListWithFieldSelectorMethod(ctx context.Context) ([]byte, int, error) {
	nodeName := getNodeNameForFieldSelector()
	if nodeName == "" {
		return nil, 0, fmt.Errorf("node name not configured for field selector query")
	}

	path := buildPodListPathWithFieldSelector(nodeName)
	log.Tracef("Querying pod list with field selector: %s", path)
	return ku.QueryKubelet(ctx, path)
}

// queryPodListLegacyMethod queries pod list using the pre-1.34 method (no field selector)
func (ku *KubeUtil) queryPodListLegacyMethod(ctx context.Context) ([]byte, int, error) {
	log.Tracef("Querying pod list with legacy method: %s", kubeletPodPath)
	return ku.QueryKubelet(ctx, kubeletPodPath)
}

// ResetPodListQueryMethod resets the detected query method (useful for testing)
func ResetPodListQueryMethod() {
	detectedMethodMutex.Lock()
	defer detectedMethodMutex.Unlock()
	detectedMethod = methodUnknown
}
