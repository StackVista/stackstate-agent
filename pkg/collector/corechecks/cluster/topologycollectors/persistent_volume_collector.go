//go:build kubeapiserver

package topologycollectors

import (
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PersistentVolumeCollector implements the ClusterTopologyCollector interface.
type PersistentVolumeCollector struct {
	pvsMappers []PersistentVolumeSourceMapper
	ClusterTopologyCollector
}

// NewPersistentVolumeCollector creates a new PersistentVolumeCollector
func NewPersistentVolumeCollector(
	clusterTopologyCollector ClusterTopologyCollector,
	csiPVMapperEnabled bool) ClusterTopologyCollector {
	pvMappers := allPersistentVolumeSourceMappers
	if csiPVMapperEnabled {
		pvMappers = append(pvMappers, mapCSIPersistentVolume)
	}
	pvc := &PersistentVolumeCollector{
		ClusterTopologyCollector: clusterTopologyCollector,
		pvsMappers:               pvMappers,
	}
	return pvc
}

// GetName returns the name of the Collector
func (*PersistentVolumeCollector) GetName() string {
	return "Persistent Volume Collector"
}

// CollectorFunction Collects and Published the Persistent Volume Components
func (pvc *PersistentVolumeCollector) CollectorFunction() error {
	persistentVolumes, err := pvc.GetAPIClient().GetPersistentVolumes()
	if err != nil {
		return err
	}

	// Read the volume attachments upfront as we will set a tag on the persistent volume containing the name of the node
	volumeAttachments, err := pvc.GetAPIClient().GetVolumeAttachments()
	if err != nil {
		log.Warnc(err.Error())
	}

	// Produce a map t
	nodeByPersistentVolume := make(map[string]string)
	for _, va := range volumeAttachments {
		nodeByPersistentVolume[*va.Spec.Source.PersistentVolumeName] = va.Spec.NodeName
	}

	for _, pv := range persistentVolumes {
		component := pvc.persistentVolumeToStackStateComponent(pv, nodeByPersistentVolume)
		pvc.SubmitComponent(component)

		volumeSource, err := pvc.persistentVolumeSourceToStackStateComponent(pv)
		if err != nil {
			continue
		}

		if volumeSource != nil {
			pvc.SubmitComponent(volumeSource)
			pvc.SubmitRelation(pvc.persistentVolumeToSourceStackStateRelation(component.ExternalID, volumeSource.ExternalID))
		}
	}

	persistentVolumesClaims, err := pvc.GetAPIClient().GetPersistentVolumeClaims()
	if err != nil {
		return err
	}

	for _, pvClaim := range persistentVolumesClaims {
		persistentVolumeClaimComponent := pvc.persistentVolumeClaimToStackStateComponent(pvClaim)
		pvc.SubmitComponent(persistentVolumeClaimComponent)

		persistentVolumeComponentExternalID := pvc.buildPersistentVolumeExternalID(pvClaim.Spec.VolumeName)
		pvc.SubmitRelation(pvc.persistentVolumeClaimToPersistentVolumeStackStateRelation(persistentVolumeClaimComponent.ExternalID, persistentVolumeComponentExternalID))
	}

	for _, va := range volumeAttachments {
		persistentVolumeExternalID := pvc.buildPersistentVolumeExternalID(*va.Spec.Source.PersistentVolumeName)
		nodeExternalID := pvc.buildNodeExternalID(va.Spec.NodeName)
		pvc.SubmitRelation(pvc.nodeToPersistentVolumeStackStateRelation(nodeExternalID, persistentVolumeExternalID))
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
func (pvc *PersistentVolumeCollector) persistentVolumeToStackStateComponent(persistentVolume v1.PersistentVolume, nodeByPersistentVolume map[string]string) *topology.Component {
	log.Tracef("Mapping PersistentVolume to StackState component: %s", persistentVolume.String())

	identifiers := make([]string, 0)

	persistentVolumeExternalID := pvc.buildPersistentVolumeExternalID(persistentVolume.Name)

	// k8s object TypeMeta seem to be archived, it's always empty.
	tags := pvc.initTags(persistentVolume.ObjectMeta, metav1.TypeMeta{Kind: "PersistentVolume"})
	if nodeName, found := nodeByPersistentVolume[persistentVolume.Name]; found {
		tags["persistent-volume-node"] = nodeName
	}

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
		var sourceProperties map[string]interface{}
		if pvc.IsExposeKubernetesStatusEnabled() {
			sourceProperties = makeSourcePropertiesFullDetails(&persistentVolume)
		} else {
			sourceProperties = makeSourceProperties(&persistentVolume)
		}
		component.SourceProperties = sourceProperties
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

// Creates a Persistent Volume Claim StackState component from a Kubernetes / OpenShift Cluster
func (pvc *PersistentVolumeCollector) persistentVolumeClaimToStackStateComponent(persistentVolumeClaim v1.PersistentVolumeClaim) *topology.Component {
	log.Tracef("Mapping PersistentClaimVolume to StackState component: %s", persistentVolumeClaim.String())

	identifiers := make([]string, 0)

	persistentVolumeClaimExternalID := pvc.buildPersistentVolumeClaimExternalID(persistentVolumeClaim.Namespace, persistentVolumeClaim.Name)

	// k8s object TypeMeta seem to be archived, it's always empty.
	tags := pvc.initTags(persistentVolumeClaim.ObjectMeta, metav1.TypeMeta{Kind: "PersistentVolumeClaim"})

	component := &topology.Component{
		ExternalID: persistentVolumeClaimExternalID,
		Type:       topology.Type{Name: "persistent-volume-claim"},
		Data: map[string]interface{}{
			"name":        persistentVolumeClaim.Name,
			"tags":        tags,
			"identifiers": identifiers,
		},
	}

	if pvc.IsSourcePropertiesFeatureEnabled() {
		var sourceProperties map[string]interface{}
		if pvc.IsExposeKubernetesStatusEnabled() {
			sourceProperties = makeSourcePropertiesFullDetails(&persistentVolumeClaim)
		} else {
			sourceProperties = makeSourceProperties(&persistentVolumeClaim)
		}
		component.SourceProperties = sourceProperties
	} else {
		component.Data.PutNonEmpty("kind", persistentVolumeClaim.Kind)
		component.Data.PutNonEmpty("uid", persistentVolumeClaim.UID)
		component.Data.PutNonEmpty("creationTimestamp", persistentVolumeClaim.CreationTimestamp)
		component.Data.PutNonEmpty("generateName", persistentVolumeClaim.GenerateName)
		component.Data.PutNonEmpty("storageClassName", persistentVolumeClaim.Spec.StorageClassName)
		component.Data.PutNonEmpty("status", persistentVolumeClaim.Status.Phase)
	}

	log.Tracef("Created StackState persistent volume claim component %s: %v", persistentVolumeClaimExternalID, component.JSONString())

	return component
}

func (pvc *PersistentVolumeCollector) createStackStateVolumeSourceComponent(pv v1.PersistentVolume, name, externalID string, identifiers []string, addTags map[string]string) (*topology.Component, error) {

	tags := pvc.initTags(pv.ObjectMeta, metav1.TypeMeta{Kind: "VolumeSource"})
	for k, v := range addTags {
		tags[k] = v
	}

	data := map[string]interface{}{
		"name": name,
		"tags": tags,
	}

	if identifiers != nil {
		data["identifiers"] = identifiers
	}

	component := &topology.Component{
		ExternalID: externalID,
		Type:       topology.Type{Name: "volume-source"},
		Data:       data,
	}

	if pvc.IsSourcePropertiesFeatureEnabled() {
		var sourceProperties map[string]interface{}

		k8sVolumeSource := K8sVolumeSource{
			TypeMeta: metav1.TypeMeta{Kind: "VolumeSource"},
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				Namespace:         pv.Namespace,
				CreationTimestamp: pv.CreationTimestamp,
			},
			PersistentVolumeSource: pv.Spec.PersistentVolumeSource,
		}

		if pvc.IsExposeKubernetesStatusEnabled() {
			sourceProperties = makeSourcePropertiesFullDetails(&k8sVolumeSource)
		} else {
			sourceProperties = makeSourceProperties(&k8sVolumeSource)
		}
		component.SourceProperties = sourceProperties
	} else {
		component.Data.PutNonEmpty("source", pv.Spec.PersistentVolumeSource)
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

func (pvc *PersistentVolumeCollector) persistentVolumeClaimToPersistentVolumeStackStateRelation(persistentVolumeClaimExternalID, persistentVolumeExternalID string) *topology.Relation {
	log.Tracef("Mapping kubernetes persistent volume claim to persistent volume: %s -> %s", persistentVolumeClaimExternalID, persistentVolumeExternalID)

	relation := pvc.CreateRelation(persistentVolumeClaimExternalID, persistentVolumeExternalID, "exposes")

	log.Tracef("Created StackState persistent volume claim -> persistent volume relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}

func (pvc *PersistentVolumeCollector) nodeToPersistentVolumeStackStateRelation(nodeExternalID, persistentVolumeExternalID string) *topology.Relation {
	log.Tracef("Mapping kubernetes node to persistent volume: %s -> %s", nodeExternalID, persistentVolumeExternalID)

	relation := pvc.CreateRelation(nodeExternalID, persistentVolumeExternalID, "exposes")

	log.Tracef("Created StackState node -> persistent volume relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}
