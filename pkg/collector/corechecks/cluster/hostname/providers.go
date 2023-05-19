// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package hostname

import (
	"fmt"
	"strings"

	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/cluster/hostname/aws"
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/cluster/hostname/azure"
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/cluster/hostname/gce"
	"github.com/StackVista/stackstate-agent/pkg/config"

	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// Provider is a generic function to grab the hostname and return it
type Provider func(providerID string) string

// providerCatalog holds all the various kinds of hostname providers
var providerCatalog = make(map[string]Provider)

// RegisterHostnameProvider registers a hostname provider as part of the catalog
func RegisterHostnameProvider(name string, p Provider) {
	providerCatalog[name] = p
}

// GetProvider returns a Provider if it was register before.
func GetProvider(providerName string) Provider {
	if provider, found := providerCatalog[providerName]; found {
		return provider
	}
	return nil
}

// GetHostname returns the hostname for a providerID for a specific Provider if it was register
func GetHostname(providerID string) (string, error) {
	if providerID == "" {
		return "", nil
	}
	providerName := strings.Split(providerID, "://")[0]

	if provider, found := providerCatalog[providerName]; found {
		log.Debugf("GetHostname trying provider '%s' ...", providerName)
		name := provider(providerID)
		if config.ValidHostname(name) != nil {
			return "", fmt.Errorf("Invalid hostname '%s' from %s provider", name, providerName)
		}
		return name, nil
	}
	return "", fmt.Errorf("hostname provider %s not found", providerName)
}

func init() {
	RegisterHostnameProvider("gce", gce.GetHostname)
	RegisterHostnameProvider("aws", aws.GetHostname)
	RegisterHostnameProvider("azure", azure.GetHostname)
}
