// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

// Package compliance defines common interfaces and types for Compliance Agent
package compliance

// Rule defines a rule in a compliance config
type Rule struct {
	ID           string        `yaml:"id"`
	Description  string        `yaml:"description,omitempty"`
	Scope        RuleScopeList `yaml:"scope,omitempty"`
	HostSelector string        `yaml:"hostSelector,omitempty"`
	Resources    []Resource    `yaml:"resources,omitempty"`
}

// RuleScope defines scope for applicability of a rule
type RuleScope string

const (
	// DockerScope const
	DockerScope RuleScope = "docker"
	// KubernetesNodeScope const
	KubernetesNodeScope RuleScope = "kubernetesNode"
	// KubernetesClusterScope const
	KubernetesClusterScope RuleScope = "kubernetesCluster"
)

// Scope defines when a rule can be run based on observed properties of the environment
type Scope struct {
	Docker            bool `yaml:"docker,omitempty"`
	KubernetesNode    bool `yaml:"kubernetesNode,omitempty"`
	KubernetesCluster bool `yaml:"kubernetesCluster,omitempty"`
}

// HostSelector allows to activate/deactivate dynamically based on host properties
type HostSelector struct {
	KubernetesNodeLabels []KubeNodeSelector `yaml:"kubernetesRole,omitempty"`
	KubernetesNodeRole   string             `yaml:"kubernetesNodeRole,omitempty"`
}

// Includes returns true if RuleScopeList includes the specified RuleScope value
func (l RuleScopeList) Includes(ruleScope RuleScope) bool {
	for _, s := range l {
		if s == ruleScope {
			return true
		}
	}
	return false
}
