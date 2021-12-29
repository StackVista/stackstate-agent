// +build kubeapiserver

package topologycollectors

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	v1 "k8s.io/api/core/v1"
	"math"
)

// ConfigMapCollector implements the ClusterTopologyCollector interface.
type ConfigMapCollector struct {
	ComponentChan chan<- *topology.Component
	ClusterTopologyCollector
	maxDataSize uint
}

// NewConfigMapCollector
func NewConfigMapCollector(componentChannel chan<- *topology.Component, clusterTopologyCollector ClusterTopologyCollector, maxDataSize uint) ClusterTopologyCollector {
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
			"name":              configMap.Name,
			"creationTimestamp": configMap.CreationTimestamp,
			"tags":              tags,
			"uid":               configMap.UID,
			"identifiers":       []string{configMapExternalID},
		},
	}

	component.Data.PutNonEmpty("generateName", configMap.GenerateName)
	component.Data.PutNonEmpty("kind", configMap.Kind)
	component.Data.PutNonEmpty("data", cutData(configMap.Data, int(cmc.maxDataSize)))

	log.Tracef("Created StackState ConfigMap component %s: %v", configMapExternalID, component.JSONString())

	return component
}

func cutReplacement(dropped string) string {
	hashing := sha256.New()
	hashing.Write([]byte(dropped))
	hash := hex.EncodeToString(hashing.Sum([]byte{}))[0:16]
	return fmt.Sprintf("[dropped %d chars, hashsum: %s]", len(dropped), hash)
}

// cutData tries to reduce `data` size
// it replaces values within the map completely or partially with `[dropped N, hash...]` string
// it's intended to strip some big files out of configmap,
// and on scale of hundred of KiB it's precise enough: it doesn't take into account replacement string
// Algorithm:
// 1. remove values that are on their own are bigger than limit
// 2. proportionally cut the rest of the values
func cutData(data map[string]string, maxSize int) map[string]string {
	if maxSize == 0 {
		return data
	}
	newData := make(map[string]string, len(data))
	toEquallyReduce := make([]string, 0, len(data))
	restSize := 0
	for k, v := range data {
		vSize := len(v)
		// immediately remove values that are themselves bigger than limit
		if vSize > maxSize {
			newData[k] = cutReplacement(v)
		} else {
			toEquallyReduce = append(toEquallyReduce, k)
			restSize += vSize
		}
	}

	if restSize > maxSize {
		ratio := float64(maxSize) / float64(restSize)
		for _, k := range toEquallyReduce {
			v := data[k]
			leaveSize := int(math.Floor(ratio * float64(len(v))))
			newData[k] = v[0:leaveSize] + cutReplacement(v[leaveSize:])
		}
	} else {
		for _, k := range toEquallyReduce {
			newData[k] = data[k]
		}
	}
	return newData
}
