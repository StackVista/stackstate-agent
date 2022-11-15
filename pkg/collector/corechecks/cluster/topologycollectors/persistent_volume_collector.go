//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	v1 "k8s.io/api/core/v1"
)

// PersistentVolumeCollector implements the ClusterTopologyCollector interface.
type PersistentVolumeCollector struct {
	pvsMappers   []PersistentVolumeSourceMapper
	RelationChan chan<- *topology.Relation
	ClusterTopologyCollector
}

// NewPersistentVolumeCollector
func NewPersistentVolumeCollector(
	relationChannel chan<- *topology.Relation,
	clusterTopologyCollector ClusterTopologyCollector,
	csiPVMapperEnabled bool) ClusterTopologyCollector {
	pvMappers := allPersistentVolumeSourceMappers
	if csiPVMapperEnabled {
		pvMappers = append(pvMappers, mapCSIPersistentVolume)
	}
	pvc := &PersistentVolumeCollector{
		RelationChan:             relationChannel,
		ClusterTopologyCollector: clusterTopologyCollector,
		pvsMappers:               pvMappers,
	}
	return pvc
}

// GetName returns the name of the Collector
func (*PersistentVolumeCollector) GetName() string {
	return "Persistent Volume Collector"
}

// Collects and Published the Persistent Volume Components
func (pvc *PersistentVolumeCollector) CollectorFunction() error {
	persistentVolumes, err := pvc.GetAPIClient().GetPersistentVolumes()
	if err != nil {
		return err
	}

	for _, pv := range persistentVolumes {
		component := pvc.persistentVolumeToStackStateComponent(pv)
		pvc.SubmitComponent(component)

		volumeSource, err := pvc.persistentVolumeSourceToStackStateComponent(pv)
		if err != nil {
			return err
		}

		if volumeSource != nil {
			pvc.SubmitComponent(volumeSource)

			pvc.RelationChan <- pvc.persistentVolumeToSourceStackStateRelation(component.ExternalID, volumeSource.ExternalID)
		}
	}

	return nil
}

func (pvc *PersistentVolumeCollector) persistentVolumeSourceToStackStateComponent(pv v1.PersistentVolume) (*topology.Component, error) {
	for _, mapper := range pvc.pvsMappers {
		c, err := mapper(pvc, pv)
		if err != nil {
			return nil, err
		}

		if c != nil {
			return c, nil
		}
	}

	log.Debugf("Unknown PersistentVolumeSource for PersistentVolume '%s', skipping it", pv.Name)

	return nil, nil
}

// Creates a Persistent Volume StackState component from a Kubernetes / OpenShift Cluster
func (pvc *PersistentVolumeCollector) persistentVolumeToStackStateComponent(persistentVolume v1.PersistentVolume) *topology.Component {
	log.Tracef("Mapping PersistentVolume to StackState component: %s", persistentVolume.String())

	identifiers := make([]string, 0)

	persistentVolumeExternalID := pvc.buildPersistentVolumeExternalID(persistentVolume.Name)

	tags := pvc.initTags(persistentVolume.ObjectMeta)

	component := &topology.Component{
		ExternalID: persistentVolumeExternalID,
		Type:       topology.Type{Name: "persistent-volume"},
		Data: map[string]interface{}{
			"name":        persistentVolume.Name,
			"tags":        tags,
			"identifiers": identifiers,
		},
	}

	if pvc.IsSourcePropertiesFeatureEnabled() {
		component.SourceProperties = makeSourceProperties(&persistentVolume)
	} else {
		component.Data.PutNonEmpty("kind", persistentVolume.Kind)
		component.Data.PutNonEmpty("uid", persistentVolume.UID)
		component.Data.PutNonEmpty("creationTimestamp", persistentVolume.CreationTimestamp)
		component.Data.PutNonEmpty("generateName", persistentVolume.GenerateName)
		component.Data.PutNonEmpty("storageClassName", persistentVolume.Spec.StorageClassName)
		component.Data.PutNonEmpty("status", persistentVolume.Status.Phase)
		component.Data.PutNonEmpty("statusMessage", persistentVolume.Status.Message)
	}

	log.Tracef("Created StackState persistent volume component %s: %v", persistentVolumeExternalID, component.JSONString())

	return component
}

func (pvc *PersistentVolumeCollector) createStackStateVolumeSourceComponent(pv v1.PersistentVolume, name, externalID string, identifiers []string, addTags map[string]string) (*topology.Component, error) {

	tags := pvc.initTags(pv.ObjectMeta)
	for k, v := range addTags {
		tags[k] = v
	}

	data := map[string]interface{}{
		"name":   name,
		"source": pv.Spec.PersistentVolumeSource,
		"tags":   tags,
	}

	if identifiers != nil {
		data["identifiers"] = identifiers
	}

	component := &topology.Component{
		ExternalID: externalID,
		Type:       topology.Type{Name: "volume-source"},
		Data:       data,
	}

	log.Tracef("Created StackState volume component %s: %v", externalID, component.JSONString())
	return component, nil
}

func (pvc *PersistentVolumeCollector) persistentVolumeToSourceStackStateRelation(persistentVolumeExternalID, persistentVolumeSourceExternalID string) *topology.Relation {
	log.Tracef("Mapping kubernetes persistent volume to persistent volume source: %s -> %s", persistentVolumeExternalID, persistentVolumeSourceExternalID)

	relation := pvc.CreateRelation(persistentVolumeExternalID, persistentVolumeSourceExternalID, "exposes")

	log.Tracef("Created StackState persistent volume -> persistent volume source relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}
