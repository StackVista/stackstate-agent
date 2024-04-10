//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// JobCollector implements the ClusterTopologyCollector interface.
type JobCollector struct {
	ClusterTopologyCollector
}

// NewJobCollector creates a new Job collector
func NewJobCollector(clusterTopologyCollector ClusterTopologyCollector) ClusterTopologyCollector {
	return &JobCollector{
		ClusterTopologyCollector: clusterTopologyCollector,
	}
}

// GetName returns the name of the Collector
func (*JobCollector) GetName() string {
	return "Job Collector"
}

// Collects and Published the Job Components
func (jc *JobCollector) CollectorFunction() error {
	jobs, err := jc.GetAPIClient().GetJobs()
	if err != nil {
		return err
	}

	for _, job := range jobs {
		component := jc.jobToStackStateComponent(job)
		jc.SubmitComponent(component)

		ownedByCron := false
		// Create relation to the cron job
		for _, ref := range job.OwnerReferences {
			switch kind := ref.Kind; kind {
			case CronJob:
				cronJobExternalID := jc.buildCronJobExternalID(job.Namespace, ref.Name)
				jc.SubmitRelation(jc.cronJobToJobStackStateRelation(cronJobExternalID, component.ExternalID))
				ownedByCron = true
			}
		}

		// If not owned by Cron Job, make a direct relation from the Namespace
		if !ownedByCron {
			jc.SubmitRelation(jc.namespaceToJobStackStateRelation(jc.buildNamespaceExternalID(job.Namespace), component.ExternalID))
		}
	}

	return nil
}

// Creates a StackState Job component from a Kubernetes / OpenShift Cluster
func (jc *JobCollector) jobToStackStateComponent(job v1.Job) *topology.Component {
	log.Tracef("Mapping Job to StackState component: %s", job.String())

	// k8s object TypeMeta seem to be archived, it's always empty.
	tags := jc.initTags(job.ObjectMeta, metav1.TypeMeta{Kind: "Job"})

	jobExternalID := jc.buildJobExternalID(job.Namespace, job.Name)
	component := &topology.Component{
		ExternalID: jobExternalID,
		Type:       topology.Type{Name: "job"},
		Data: map[string]interface{}{
			"name": job.Name,
			"tags": tags,
		},
	}

	if jc.IsSourcePropertiesFeatureEnabled() {
		var sourceProperties map[string]interface{}
		if jc.IsExposeKubernetesStatusEnabled() {
			sourceProperties = makeSourcePropertiesFullDetails(&job)
		} else {
			sourceProperties = makeSourceProperties(&job)
		}
		component.SourceProperties = sourceProperties
	} else {
		component.Data.PutNonEmpty("kind", job.Kind)
		component.Data.PutNonEmpty("creationTimestamp", job.CreationTimestamp)
		component.Data.PutNonEmpty("uid", job.UID)
		component.Data.PutNonEmpty("generateName", job.GenerateName)
		component.Data.PutNonEmpty("backoffLimit", job.Spec.BackoffLimit)
		component.Data.PutNonEmpty("parallelism", job.Spec.Parallelism)
	}

	log.Tracef("Created StackState Job component %s: %v", jobExternalID, component.JSONString())

	return component
}

// Creates a StackState relation from a Kubernetes / OpenShift Job to CronJob relation
func (jc *JobCollector) cronJobToJobStackStateRelation(cronJobExternalID, jobExternalID string) *topology.Relation {
	log.Tracef("Mapping kubernetes cron job to job relation: %s -> %s", cronJobExternalID, jobExternalID)

	relation := jc.CreateRelation(cronJobExternalID, jobExternalID, "creates")

	log.Tracef("Created StackState cron job -> job relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}

// Creates a StackState relation from a Kubernetes / OpenShift Namespace to Job relation
func (jc *JobCollector) namespaceToJobStackStateRelation(namespaceExternalID, jobExternalID string) *topology.Relation {
	log.Tracef("Mapping kubernetes namespace to job relation: %s -> %s", namespaceExternalID, jobExternalID)

	relation := jc.CreateRelation(namespaceExternalID, jobExternalID, "encloses")

	log.Tracef("Created StackState namespace -> job relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}
