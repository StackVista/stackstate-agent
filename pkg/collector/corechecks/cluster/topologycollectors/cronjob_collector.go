//go:build kubeapiserver

package topologycollectors

import (
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CronJobCollector implements the ClusterTopologyCollector interface.
type CronJobCollector struct {
	ClusterTopologyCollector
}

// NewCronJobCollector creates a new CronJob collector
func NewCronJobCollector(
	clusterTopologyCollector ClusterTopologyCollector) ClusterTopologyCollector {
	return &CronJobCollector{
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
	var err error
	if supported := cjc.minimumMinorVersion(21); supported {
		cronJobs, err = cjc.getCronJobsV1(cronJobs)
	} else {
		cronJobs, err = cjc.getCronJobsV1B1(cronJobs)
	}
	if err != nil {
		return err
	}

	for _, cj := range cronJobs {
		component := cjc.cronJobToStackStateComponent(cj)
		cjc.SubmitComponent(component)
		cjc.SubmitRelation(cjc.namespaceToCronJobStackStateRelation(cjc.buildNamespaceExternalID(cj.GetNamespace()), component.ExternalID))
	}

	return nil
}

func (cjc *CronJobCollector) getCronJobsV1B1(cronJobs []CronJobInterface) ([]CronJobInterface, error) {
	ingressesExt, err := cjc.GetAPIClient().GetCronJobsV1B1()
	if err != nil {
		return nil, err
	}
	for _, cj := range ingressesExt {
		log.Debugf("Got CronJob '%s' from batch/v1beta1", cj.Name)
		cronJobs = append(cronJobs, CronJobV1B1{
			o: cj,
		})
	}
	return cronJobs, nil
}

func (cjc *CronJobCollector) getCronJobsV1(cronJobs []CronJobInterface) ([]CronJobInterface, error) {
	ingressesExt, err := cjc.GetAPIClient().GetCronJobsV1()
	if err != nil {
		return nil, err
	}
	for _, cj := range ingressesExt {
		log.Debugf("Got CronJob '%s' from batch/v1", cj.Name)
		cronJobs = append(cronJobs, CronJobV1{
			o: cj,
		})
	}
	return cronJobs, nil
}

// cronJobToStackStateComponent Creates a StackState CronJob component from a Kubernetes / OpenShift Cluster
func (cjc *CronJobCollector) cronJobToStackStateComponent(cronJob CronJobInterface) *topology.Component {
	log.Tracef("Mapping CronJob to StackState component: %s", cronJob.GetString())

	// k8s object TypeMeta seem to be archived, it's always empty.
	tags := cjc.initTags(cronJob.GetObjectMeta(), metaV1.TypeMeta{Kind: "CronJob"})

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
		var sourceProperties map[string]interface{}
		if cjc.IsExposeKubernetesStatusEnabled() {
			sourceProperties = makeSourcePropertiesFullDetails(cronJob.GetKubernetesObject())
		} else {
			sourceProperties = makeSourceProperties(cronJob.GetKubernetesObject())
		}
		component.SourceProperties = sourceProperties

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
