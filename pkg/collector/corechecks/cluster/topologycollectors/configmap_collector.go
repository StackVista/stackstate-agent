// +build kubeapiserver

package topologycollectors

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	v1 "k8s.io/api/core/v1"
)

// ConfigMapCollector implements the ClusterTopologyCollector interface.
type ConfigMapCollector struct {
	ComponentChan chan<- *topology.Component
	ClusterTopologyCollector
	maxDataSize int
}

// NewConfigMapCollector
func NewConfigMapCollector(componentChannel chan<- *topology.Component, clusterTopologyCollector ClusterTopologyCollector, maxDataSize int) ClusterTopologyCollector {
	log.Infof("Initialized ConfigMap collector with %d size limit for configmap data", maxDataSize)

	return &ConfigMapCollector{
		ComponentChan:            componentChannel,
		ClusterTopologyCollector: clusterTopologyCollector,
		maxDataSize:              maxDataSize,
	}
}

// GetName returns the name of the Collector
func (*ConfigMapCollector) GetName() string {
	return "ConfigMap Collector"
}

// Collects and Published the ConfigMap Components
func (cmc *ConfigMapCollector) CollectorFunction() error {
	configMaps, err := cmc.GetAPIClient().GetConfigMaps()
	if err != nil {
		return err
	}

	for _, cm := range configMaps {
		cmc.ComponentChan <- cmc.configMapToStackStateComponent(cm)
	}

	return nil
}

// Creates a StackState ConfigMap component from a Kubernetes / OpenShift Cluster
func (cmc *ConfigMapCollector) configMapToStackStateComponent(configMap v1.ConfigMap) *topology.Component {
	log.Tracef("Mapping ConfigMap to StackState component: %s", configMap.String())

	tags := cmc.initTags(configMap.ObjectMeta)
	configMapExternalID := cmc.buildConfigMapExternalID(configMap.Namespace, configMap.Name)

	component := &topology.Component{
		ExternalID: configMapExternalID,
		Type:       topology.Type{Name: "configmap"},
		Data: map[string]interface{}{
			"name":        configMap.Name,
			"tags":        tags,
			"identifiers": []string{configMapExternalID},
		},
	}

	if cmc.IsSourcePropertiesFeatureEnabled() {
		configMapCopy := configMap
		configMapCopy.Data = cutData(configMap.Data, cmc.maxDataSize)
		for k, data := range configMapCopy.BinaryData {
			configMapCopy.BinaryData[k] = []byte(cutReplacement(data))
		}
		sourceProperties := makeSourceProperties(&configMapCopy)
		component.SourceProperties = sourceProperties
	} else {
		component.Data.PutNonEmpty("kind", configMap.Kind)
		component.Data.PutNonEmpty("creationTimestamp", configMap.CreationTimestamp)
		component.Data.PutNonEmpty("generateName", configMap.GenerateName)
		component.Data.PutNonEmpty("uid", configMap.UID)
		component.Data.PutNonEmpty("data", cutData(configMap.Data, cmc.maxDataSize))
	}

	log.Tracef("Created StackState ConfigMap component %s: %v", configMapExternalID, component.JSONString())

	return component
}

func cutReplacement(dropped []byte) string {
	hashing := sha256.New()
	_, err := hashing.Write(dropped)
	var hash string
	if err != nil {
		// doubt what error could happen, but just to satisfy linter, and in case...
		hash = fmt.Sprintf("hash error: %v", err)
	} else {
		hash = hex.EncodeToString(hashing.Sum([]byte{}))[0:16]
	}
	return fmt.Sprintf("[dropped %d chars, hashsum: %s]", len(dropped), hash)
}

// cutData tries to reduce `data` size
// it replaces values within the map completely or partially with `[dropped N, hash...]` string
// cut limit is defined as maxSize divided by entries count in data. So every entry is limited to maxSize/len(data) bytes
func cutData(data map[string]string, maxSize int) map[string]string {
	keyCount := len(data)
	if maxSize == 0 || keyCount == 0 {
		return data
	}
	maxPerKey := maxSize / keyCount
	newData := make(map[string]string, len(data))
	for k, v := range data {
		valueSize := len(v)
		if valueSize > maxPerKey {
			newData[k] = v[0:maxPerKey] + cutReplacement([]byte(v[maxPerKey:]))
		} else {
			newData[k] = v
		}
	}
	return newData
}
