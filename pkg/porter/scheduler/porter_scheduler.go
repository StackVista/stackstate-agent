// +build kubeapiserver

package scheduler

import (
	"errors"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	batchV1 "k8s.io/api/batch/v1"
	batchV1B "k8s.io/api/batch/v1beta1"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Porter defines the structure of a porter
type Porter struct {
	Name string
	PorterDefinition map[string]string
}

// PorterScheduler provides an interface to schedule porters
type PorterScheduler interface {
	SchedulePorter(porter Porter) error
}

// KubernetesPorterScheduler wraps the api client
type KubernetesPorterScheduler struct {
	APIClient *apiserver.APIClient
}

// MakeKubernetesPorterScheduler creates a instance of KubernetesPorterScheduler
func MakeKubernetesPorterScheduler(client *apiserver.APIClient) *KubernetesPorterScheduler {
	return &KubernetesPorterScheduler{APIClient: client}
}

// KubernetesPorterCronJobDefinition describe a porter in k8s
type KubernetesPorterCronJobDefinition struct {
	Image string
	Namespace string
}

// SchedulePorter schedules a porter as a cron job
func (k *KubernetesPorterScheduler) SchedulePorter(porter Porter) error {
	porterCJDefinition, err := k.interpretPorterDefinition(porter.PorterDefinition)
	if err != nil {
		return errors.New("could not create porter instance")
	}

	cronjob, err := k.APIClient.CreateCronJob(porterCJDefinition.Namespace, &batchV1B.CronJob{
		TypeMeta:   metav1.TypeMeta{
			Kind:       "CronJob",
			APIVersion: "batch/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:                       fmt.Sprintf("%s Porter CronJob", porter.Name),
			Namespace:                  porterCJDefinition.Namespace,
		},
		Spec:       batchV1B.CronJobSpec{
			Schedule:                   "*/10 * * * *",
			JobTemplate:                batchV1B.JobTemplateSpec{
				Spec:       batchV1.JobSpec{
					Template:                v1.PodTemplateSpec{
						Spec:       v1.PodSpec{
							Containers:                    []coreV1.Container{
								{
									Name:                     porter.Name,
									Image:                   porterCJDefinition.Image,
									Args:                     []string{"./agent-porter", "-h", "localhost", "-p",
										"50051", "-instance-type", "my-instance", "-instance-url", "my-url"},
									ImagePullPolicy:          coreV1.PullAlways,
								},
							},
							RestartPolicy: coreV1.RestartPolicyOnFailure,
						},
					},
				},
			},
			SuccessfulJobsHistoryLimit: nil,
			FailedJobsHistoryLimit:     nil,
		},
	})

	if err != nil {
		return err
	}


	log.Infof("Created Porter as cron job: %v", cronjob)

	return nil
}

func (k *KubernetesPorterScheduler) interpretPorterDefinition(porterDefinition map[string]string) (*KubernetesPorterCronJobDefinition, error) {
	imageName, ok := porterDefinition["ImageName"]; if !ok {
		return nil, errors.New("no image defined for this porter")
	}

	namespace, ok := porterDefinition["Namespace"]; if !ok {
		return nil, errors.New("no namespace defined for this porter")
	}

	return &KubernetesPorterCronJobDefinition{
		Image: imageName,
		Namespace: namespace,
	}, nil
}
