// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.

//go:build kubeapiserver

package apiserver

//// Covered by test/integration/util/kube_apiserver/events_test.go

import (
	"context"
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"strconv"
	"time"
)

// GetPods retrieves all the pods in the Kubernetes cluster across all namespaces.
func (c *APIClient) GetPods() ([]v1.Pod, error) {
	podList, err := c.Cl.CoreV1().Pods(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return []v1.Pod{}, err
	}

	return podList.Items, nil
}

// RunPodCollection Fined tuned version of GetPods to allow fetching pods based on a resourceVersion and catching up to the latest pod state
func (c *APIClient) RunPodCollection(lastSyncTime time.Time, watchPodsCatchupTimeout int64, podCardinalityLimit int64, resync int64, resourceVersion string, forceWatch bool) ([]*v1.Pod, string, time.Time, error) {
	log.Debug("Starting pod collection")

	// If we are forcing a retrieval on watch then we skip the initial fetch period and allow it to happen within the watch
	if !forceWatch {
		// Determine if the resource value is empty or the sync time has expired
		// If it is then we attempt to reset the resource version
		syncTimeout := time.Duration(resync) * time.Second
		syncDiffTime := time.Now().Sub(lastSyncTime)
		if resourceVersion == "" || syncDiffTime > syncTimeout {
			log.Debugf("Return list of pods - syncDiffTime: %d/%d", syncDiffTime, syncTimeout)
			return c.FetchListOfPods(watchPodsCatchupTimeout, podCardinalityLimit, resourceVersion)
		}
	}

	// Watch pods and trigger a channel event for any pod changes
	// From the pod events we can start extracting changes like status updates
	podsWatcher, err := c.Cl.CoreV1().Pods(metav1.NamespaceAll).Watch(context.TODO(), metav1.ListOptions{
		Watch:           true,
		ResourceVersion: resourceVersion,
		Limit:           podCardinalityLimit,
	})

	defer podsWatcher.Stop()

	// If there is an error in the watch event then return all the pods that was already captured within the last cycle
	if err != nil {
		return []*v1.Pod{}, resourceVersion, lastSyncTime, err
	}

	return c.CatchupToLatestPods(lastSyncTime, watchPodsCatchupTimeout, podCardinalityLimit, resourceVersion, podsWatcher)
}

func (c *APIClient) CatchupToLatestPods(lastSyncTime time.Time, watchPodsCatchupTimeout int64, podCardinalityLimit int64, resourceVersion string, podsWatcher watch.Interface) ([]*v1.Pod, string, time.Time, error) {
	var pods []*v1.Pod

	log.Debugf("Starting to watch pods from %s", resourceVersion)

	// The catchup timeout trigger is responsible to stop the pod watch channel events
	// The following line "c.Cl.CoreV1().Pods(metav1.NamespaceAll).Watch" will create a channel that pushes in any and all
	// changes received to pods from the kube api. This is a channel that will never close thus we need to forcefully close
	// it within x time, this is where the catchupTimeoutTrigger comes into play.
	catchupTimeoutTrigger := time.NewTimer(time.Duration(watchPodsCatchupTimeout) * time.Second)

	for {
		select {
		case podEvent, ok := <-podsWatcher.ResultChan():
			// If the channel is closed then return the pods that was already captured within the last cycle
			if !ok {
				return pods, resourceVersion, lastSyncTime, fmt.Errorf("unexpected channel close while watching pods")
			}

			switch podEvent.Type {
			// Determine if an error occurred when receiving the latest pod update from the watch event
			case watch.Error:
				// Attempt to extract the status to determine the error status
				status, ok := podEvent.Object.(*metav1.Status)
				if !ok {
					return pods, resourceVersion, lastSyncTime, fmt.Errorf("could not unmarshall the status of the pod watched event")
				}

				switch status.Reason {
				case "Expired":
					log.Debug("Resource Version is too old, listing all events and collecting only the new ones")
					return c.FetchListOfPods(watchPodsCatchupTimeout, podCardinalityLimit, resourceVersion)

				default:
					return pods, resourceVersion, lastSyncTime, fmt.Errorf("received an unexpected status while collecting the events: %s", status.Reason)
				}

			// The events informer sends the state of an object immediately before deletion.
			// We're not interested in re-processing these events because they should be processed already when they were added.
			// This happens when an event reaches the events TTL, an apiserver config (default 1 hour).
			// Ignoring this type of informer events will prevent from sending duplicated datadog events.
			case watch.Deleted:
				continue

			default:
				pod, ok := podEvent.Object.(*v1.Pod)
				// Could not cast to a pod, might as well drop this pod, and continue.
				if !ok {
					_ = log.Errorf("The event object for %v cannot be safely converted, skipping it.", podEvent.Object)
					continue
				}

				podResourceVersionInt, err := strconv.Atoi(pod.ResourceVersion)
				if err != nil {
					return pods, resourceVersion, lastSyncTime, err
				}

				pods = append(pods, pod)

				resourceVersionInt, err := strconv.Atoi(resourceVersion)
				if err != nil {
					_ = log.Errorf("Could not cast %s into an integer: %s", resourceVersion, err.Error())
					continue
				}
				if podResourceVersionInt > resourceVersionInt {
					// Events from the watch are not ordered necessarily, let's keep track of the newest RV.
					resourceVersion = pod.ResourceVersion
				}
			}

		case <-catchupTimeoutTrigger.C:
			log.Debugf("Collected %d pods, will resume watching from resource version %s", len(pods), resourceVersion)
			// No more events to read or the watch lasted more than `watchPodsCatchupTimeout`.
			// so return what was processed.
			return pods, resourceVersion, lastSyncTime, nil

		}
	}
}

func (c *APIClient) FetchListOfPods(watchPodsCatchupTimeout int64, podCardinalityLimit int64, resourceVersion string) ([]*v1.Pod, string, time.Time, error) {
	// Get a new list of pods seeing that the sync time has expired or resourceVersion is empty
	podList, lastResourceVersion, lastTime, err := c.GetListOfPods(watchPodsCatchupTimeout, podCardinalityLimit)
	if err != nil {
		return nil, "", time.Now(), err
	}

	// Convert the resource version to an integer, if it fails then we need to force an integer
	// to allow further integer operations to determine if the resourceVersion increased
	resourceVersionInt, ok := strconv.Atoi(resourceVersion)
	if ok != nil {
		resourceVersionInt = 0
	}

	return findPodsAfterResourceVersion(resourceVersionInt, podList), lastResourceVersion, lastTime, nil
}

// findPodsAfterResourceVersion Find all pods that is newer than a specific resource version
func findPodsAfterResourceVersion(resourceVersionInt int, currentPodList []*v1.Pod) []*v1.Pod {
	var pods []*v1.Pod

	// Run through the current pod list to determine if the resource version is a valid integer
	// and if it is valid, make sure that the version is newer than the current resource version
	for _, pod := range currentPodList {
		podResourceVersionInt, err := strconv.Atoi(pod.ResourceVersion)
		if err != nil {
			_ = log.Errorf("Could not parse resource version of an pod, will skip: %s", err)
			continue
		}

		if podResourceVersionInt > resourceVersionInt {
			pods = append(pods, pod)
		}
	}

	log.Debugf("Returning %d pods that we have not collected", len(pods))
	return pods
}

// GetListOfPods Get the current list of pods
func (c *APIClient) GetListOfPods(timeout int64, limit int64) (pods []*v1.Pod, resourceVersion string, lastListTime time.Time, err error) {
	podList, err := c.Cl.CoreV1().Pods(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
		TimeoutSeconds: &timeout,
		Limit:          limit,
	})

	if err != nil {
		_ = log.Errorf("Error Listing pods: %s", err.Error())
		return nil, resourceVersion, lastListTime, err
	}

	for id := range podList.Items {
		pods = append(pods, &podList.Items[id])
	}

	return pods, podList.ResourceVersion, time.Now(), nil
}
