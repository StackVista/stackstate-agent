// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2017 Datadog, Inc.
//go:build kubeapiserver

package kubeapi

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/DataDog/datadog-agent/pkg/collector/corechecks/cluster/urn"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/DataDog/datadog-agent/pkg/util/log"

	"github.com/DataDog/datadog-agent/pkg/metrics/event"
	v1 "k8s.io/api/core/v1"
)

type EventCategory string

const (
	Alerts     EventCategory = "Alerts"
	Changes                  = "Changes"
	Activities               = "Activities"
	Others                   = "Others"
)

// ValidCategories contains categories that are only expected in StackState
var ValidCategories = []EventCategory{Alerts, Changes, Activities, Others}

// DefaultEventCategoriesMap contains a mapping of Kubernetes EventReason strings to specific Event Categories.
// If an event is 'event.Type = warning' then we map it automatically to an 'Alert'
var DefaultEventCategoriesMap = map[string]EventCategory{
	// Container events
	"BackOff":             Alerts,
	"Created":             Activities,
	"ExceededGracePeriod": Activities,
	"Killing":             Activities,
	"Preempting":          Activities,
	"Started":             Activities,

	// Image events
	"Pulling": Activities,
	"Pulled":  Activities,

	// Kubelet events
	"NodeReady":                  Activities,
	"NodeNotReady":               Activities,
	"NodeSchedulable":            Activities,
	"Starting":                   Activities,
	"VolumeResizeSuccessful":     Activities,
	"FileSystemResizeSuccessful": Activities,
	//events.SuccessfulDetachVolume:               Activities, //TODO agent3
	"SuccessfulAttachVolume": Activities,
	"SuccessfulMountVolume":  Activities,
	//events.SuccessfulUnMountVolume:              Activities, //TODO agent3
	"Rebooted":                Activities,
	"ContainerGCFailed":       Activities,
	"ImageGCFailed":           Activities,
	"NodeAllocatableEnforced": Activities,
	"SandboxChanged":          Changes,

	// Seen in the wild, not keys of our current lib
	"Completed":         Activities,
	"NoPods":            Activities,
	"NotTriggerScaleUp": Alerts,
	"SawCompletedJob":   Activities,
	"ScalingReplicaSet": Activities,
	"Scheduled":         Activities,
	"SuccessfulCreate":  Activities,
	"SuccessfulDelete":  Activities,

	// HPA events
	// https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/podautoscaler/horizontal.go
	// https://github.com/kubernetes/kubernetes/blob/v1.25.0/pkg/controller/podautoscaler/horizontal_test.go
	"SuccessfulRescale":            Activities,
	"DesiredReplicasComputed":      Others,
	"FailedComputeMetricsReplicas": Alerts,
	"FailedRescale":                Alerts,

	"EnsuringLoadBalancer":        Activities,
	"AddedInterface":              Changes,
	"BuildStarted":                Activities,
	"ScaleDown":                   Activities,
	"BuildCompleted":              Activities,
	"DeploymentCreated":           Changes,
	"OperationStarted":            Activities,
	"ResourceUpdated":             Changes,
	"OperationCompleted":          Activities,
	"WaitForFirstConsumer":        Others,
	"Running":                     Activities,
	"Pending":                     Others,
	"ProvisioningSucceeded":       Activities,
	"Succeeded":                   Activities,
	"NodeNotSchedulable":          Alerts,
	"Updated":                     Changes,
	"Deleted":                     Changes,
	"Inject":                      Others,
	"DeletingNode":                Activities,
	"RemovingNode":                Activities,
	"ReplicationControllerScaled": Activities,
	"DetectedUnhealthy":           Alerts,
	"ConnectivityRestored":        Alerts,
	"DeletingLoadBalancer":        Changes,
	"JobAlreadyActive":            Others,
	"BuildFailed":                 Alerts,
	"NeedsReinstall":              Alerts,
	"AllRequirementsMet":          Activities,
	"InstallSucceeded":            Activities,
	"InstallWaiting":              Others,
	"CreatedSCCRanges":            Changes,

	// Service events
	"UpdatedLoadBalancer": Changes,

	// Pod events
	"TriggeredScaleUp": Activities,
	"RELOAD":           Changes,
	"EvictedByVPA":     Activities,

	// Statefulset events
	"ScaleUp": Activities,

	// Configmap events
	"ScaledUpGroup":  Activities,
	"ScaleDownEmpty": Activities,
	"CREATE":         Activities,

	// PVC events
	"Provisioning":         Activities,
	"ExternalProvisioning": Activities,
	"Resizing":             Activities,

	// Node events
	"NodeHasInsufficientMemory": Alerts,
}

type KubernetesEventMapperFactory func(detector apiserver.OpenShiftDetector, clusterName string, eventCategoriesOverride map[string]EventCategory) *kubernetesEventMapper

type kubernetesEventMapper struct {
	urn                     urn.Builder
	clusterName             string
	sourceType              string
	eventCategoriesOverride map[string]EventCategory
}

func newKubernetesEventMapper(detector apiserver.OpenShiftDetector, clusterName string, eventCategoriesOverride map[string]EventCategory) *kubernetesEventMapper {
	f := kubernetesFlavour(detector)
	return &kubernetesEventMapper{
		urn:                     urn.NewURNBuilder(f, clusterName),
		clusterName:             clusterName,
		sourceType:              string(f),
		eventCategoriesOverride: eventCategoriesOverride,
	}
}

var _ KubernetesEventMapperFactory = newKubernetesEventMapper // Compile-time check

func (k *kubernetesEventMapper) mapKubernetesEvent(event *v1.Event) (event.Event, error) {
	if err := checkEvent(event); err != nil {
		return event.Event{}, err
	}

	// Map Category to event type
	//

	mEvent := metrics.Event{
		Title:          fmt.Sprintf("%s - %s %s (%dx)", event.Reason, event.InvolvedObject.Name, event.InvolvedObject.Kind, event.Count),
		Host:           getHostName(event, k.clusterName),
		SourceTypeName: k.sourceType,
		Priority:       metrics.EventPriorityNormal,
		AlertType:      getAlertType(event),
		EventType:      event.Reason,
		Ts:             getTimeStamp(event),
		Tags:           k.getTags(event),
		EventContext: &metrics.EventContext{
			Source:             k.sourceType,
			Category:           string(k.getCategory(event)),
			SourceIdentifier:   string(event.GetUID()),
			ElementIdentifiers: k.externalIdentifierForInvolvedObject(event),
			SourceLinks:        []metrics.SourceLink{},
			Data:               map[string]interface{}{},
		},
		Text: event.Message,
	}

	return mEvent, nil
}

func checkEvent(event *v1.Event) error {
	// As some fields are optional, we want to avoid evaluating empty values.
	if event == nil || event.InvolvedObject.Kind == "" {
		return errors.New("could not retrieve some parent attributes of the event")
	}

	if event.Reason == "" || event.Message == "" || event.InvolvedObject.Name == "" {
		return errors.New("could not retrieve some attributes of the event")
	}

	return nil
}

func getHostName(event *v1.Event, clusterName string) string {
	if event.InvolvedObject.Kind == "Node" || event.InvolvedObject.Kind == "Pod" {
		if clusterName != "" {
			return fmt.Sprintf("%s-%s", event.Source.Host, clusterName)
		}

		return event.Source.Host
	}

	// If hostname was not defined, the aggregator will then set the local hostname
	return ""
}

var thrownCategoryWarnings sync.Map

func (k *kubernetesEventMapper) getCategory(event *v1.Event) EventCategory {
	if category, ok := k.eventCategoriesOverride[event.Reason]; ok {
		return category
	}

	alertType := getAlertType(event)
	if alertType == metrics.EventAlertTypeWarning || alertType == metrics.EventAlertTypeError {
		return Alerts
	}

	if category, ok := DefaultEventCategoriesMap[event.Reason]; ok {
		return category
	}
	if _, exists := thrownCategoryWarnings.LoadOrStore(event.Reason, true); !exists {
		_ = log.Warnf("Kubernetes event has unknown reason '%s' found, categorising as 'Others'. Involved object: '%s/%s'", event.Reason, event.InvolvedObject.Kind, event.InvolvedObject.Name)
	}
	return Others
}

func getAlertType(event *v1.Event) event.EventAlertType {
	switch strings.ToLower(event.Type) {
	case "normal":
		return event.EventAlertTypeInfo
	case "warning":
		return event.EventAlertTypeWarning
	default:
		log.Warnf("Unhandled kubernetes event type '%s', fallback to metrics.EventAlertTypeInfo", event.Type)
		return event.EventAlertTypeInfo
	}
}

func getTimeStamp(event *v1.Event) int64 {
	return event.LastTimestamp.Unix()
}

func (k *kubernetesEventMapper) getTags(event *v1.Event) []string {
	tags := []string{}

	if event.InvolvedObject.Namespace != "" {
		tags = append(tags, fmt.Sprintf("kube_namespace:%s", event.InvolvedObject.Namespace))
	}

	if event.InvolvedObject.FieldPath != "" && getContainerNameFromEvent(event) != "" {
		tags = append(tags, fmt.Sprintf("kube_container_name:%s", getContainerNameFromEvent(event)))
	}

	tags = append(tags, fmt.Sprintf("source_component:%s", event.Source.Component))
	tags = append(tags, fmt.Sprintf("kube_object_name:%s", event.InvolvedObject.Name))
	tags = append(tags, fmt.Sprintf("kube_object_kind:%s", event.InvolvedObject.Kind))
	tags = append(tags, fmt.Sprintf("kube_cluster_name:%s", k.clusterName))
	tags = append(tags, fmt.Sprintf("kube_reason:%s", event.Reason))
	tags = append(tags, fmt.Sprintf("alert_type:%s", getAlertType(event)))

	return tags
}

func (k *kubernetesEventMapper) externalIdentifierForInvolvedObject(event *v1.Event) []string {
	identifiers := []string{}
	namespace := event.InvolvedObject.Namespace
	obj := event.InvolvedObject

	if event.InvolvedObject.FieldPath != "" && getContainerNameFromEvent(event) != "" {
		identifiers = append(identifiers, k.urn.BuildContainerExternalID(namespace, obj.Name, getContainerNameFromEvent(event)))
	}

	urn, err := k.urn.BuildExternalID(obj.Kind, namespace, obj.Name)
	identifiers = append(identifiers, urn)
	if err != nil {
		log.Warnf("Unknown InvolvedObject type '%s' for obj '%s/%s' in event '%s'", obj.Kind, namespace, obj.Name, event.Name)
		identifiers = append(identifiers, "")
	}

	return identifiers
}

func getContainerNameFromEvent(event *v1.Event) string {

	containerName := ""

	if event.InvolvedObject.FieldPath != "" {
		r := regexp.MustCompile("spec.containers{(.*?)}")

		containerNameSubmatch := r.FindStringSubmatch(event.InvolvedObject.FieldPath)

		if len(containerNameSubmatch) == 2 {
			containerName = containerNameSubmatch[1]
		}

		log.Debugf("Container name '%s' extracted from event '%s'", containerName, event.InvolvedObject)
	}

	return containerName
}

func kubernetesFlavour(detector apiserver.OpenShiftDetector) urn.ClusterType {
	switch openshiftPresence := detector.DetectOpenShiftAPILevel(); openshiftPresence {
	case apiserver.OpenShiftAPIGroup, apiserver.OpenShiftOAPI:
		return urn.OpenShift
	default:
		return urn.Kubernetes
	}

}
