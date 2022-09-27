//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// CronJobCollector implements the ClusterTopologyCollector interface.
type CronJobCollector struct {
	ComponentChan chan<- *topology.Component
	RelationChan  chan<- *topology.Relation
	ClusterTopologyCollector
}

// NewCronJobCollector creates a new CronJob collector
func NewCronJobCollector(componentChannel chan<- *topology.Component, relationChannel chan<- *topology.Relation,
	clusterTopologyCollector ClusterTopologyCollector) ClusterTopologyCollector {
	return &CronJobCollector{
		ComponentChan:            componentChannel,
		RelationChan:             relationChannel,
		ClusterTopologyCollector: clusterTopologyCollector,
	}
}

// GetName returns the name of the Collector
func (*CronJobCollector) GetName() string {
	return "CronJob Collector"
}

// CollectorFunction Collects and publishes CronJob components
func (cjc *CronJobCollector) CollectorFunction() error {
	var cronJobs []CronJobInterface
	cronJobs, err := cjc.getCronJobsV1B1(cronJobs)
	if err != nil {
		return err
	}
	cronJobs, err = cjc.getCronJobsV1(cronJobs)
	if err != nil {
		return err
	}

	for _, cj := range cronJobs {
		component := cjc.cronJobToStackStateComponent(cj)
		cjc.ComponentChan <- component
		cjc.RelationChan <- cjc.namespaceToCronJobStackStateRelation(cjc.buildNamespaceExternalID(cj.GetNamespace()), component.ExternalID)
	}

	return nil
}

func (cjc *CronJobCollector) getCronJobsV1B1(cronJobs []CronJobInterface) ([]CronJobInterface, error) {
	if supported, err := cjc.ClusterTopologyCollector.maximumMinorVersion(24); err != nil || !supported {
		if !supported {
			log.Debugf("CronJobs from batch/v1beta1 are not supported in this Kubernetes version")
		}
		return cronJobs, err
	}
	ingressesExt, err := cjc.GetAPIClient().GetCronJobsV1B1()
	if err != nil {
		return nil, err
	}
	for _, in := range ingressesExt {
		cronJobs = append(cronJobs, CronJobV1B1{
			o: in,
		})
	}
	return cronJobs, nil
}

func (cjc *CronJobCollector) getCronJobsV1(cronJobs []CronJobInterface) ([]CronJobInterface, error) {
	if supported, err := cjc.ClusterTopologyCollector.minimumMinorVersion(21); err != nil || !supported {
		if !supported {
			log.Debugf("CronJobs from batch/v1 are not supported in this Kubernetes version")
		}
		return cronJobs, err
	}
	ingressesExt, err := cjc.GetAPIClient().GetCronJobsV1()
	if err != nil {
		return nil, err
	}
	for _, in := range ingressesExt {
		cronJobs = append(cronJobs, CronJobV1{
			o: in,
		})
	}
	return cronJobs, nil
}

// cronJobToStackStateComponent Creates a StackState CronJob component from a Kubernetes / OpenShift Cluster
func (cjc *CronJobCollector) cronJobToStackStateComponent(cronJob CronJobInterface) *topology.Component {
	log.Tracef("Mapping CronJob to StackState component: %s", cronJob.GetString())

	tags := cjc.initTags(cronJob.GetObjectMeta())

	cronJobExternalID := cjc.buildCronJobExternalID(cronJob.GetNamespace(), cronJob.GetName())

	component := &topology.Component{
		ExternalID: cronJobExternalID,
		Type:       topology.Type{Name: "cronjob"},
		Data: map[string]interface{}{
			"name": cronJob.GetName(),
			"tags": tags,
		},
	}

	if cjc.IsSourcePropertiesFeatureEnabled() {
		component.SourceProperties = makeSourceProperties(cronJob.GetKubernetesObject())
	} else {
		component.Data.PutNonEmpty("uid", cronJob.GetUID())
		component.Data.PutNonEmpty("kind", cronJob.GetKind())
		component.Data.PutNonEmpty("creationTimestamp", cronJob.GetCreationTimestamp())
		component.Data.PutNonEmpty("generateName", cronJob.GetGenerateName())
		component.Data.PutNonEmpty("schedule", cronJob.GetSchedule())
		component.Data.PutNonEmpty("concurrencyPolicy", cronJob.GetConcurrencyPolicy())
	}

	log.Tracef("Created StackState CronJob component %s: %v", cronJobExternalID, component.JSONString())

	return component
}

// Creates a StackState relation from a Kubernetes / OpenShift Namespace to CronJob relation
func (cjc *CronJobCollector) namespaceToCronJobStackStateRelation(namespaceExternalID, cronJobExternalID string) *topology.Relation {
	log.Tracef("Mapping kubernetes namespace to cron job relation: %s -> %s", namespaceExternalID, cronJobExternalID)

	relation := cjc.CreateRelation(namespaceExternalID, cronJobExternalID, "encloses")

	log.Tracef("Created StackState namespace -> cron job relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}
