// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package checks

import (
	"context"
	"errors"
	"fmt"

	"github.com/DataDog/datadog-agent/pkg/compliance"
	"github.com/DataDog/datadog-agent/pkg/compliance/event"
	"github.com/DataDog/datadog-agent/pkg/util/jsonquery"
	"github.com/DataDog/datadog-agent/pkg/util/log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type kubeApiserverCheck struct {
	baseCheck
	kubeResource *compliance.KubernetesResource
}

const (
	kubeResourceNameKey      string = "kube_resource_name"
	kubeResourceGroupKey     string = "kube_resource_group"
	kubeResourceVersionKey   string = "kube_resource_version"
	kubeResourceNamespaceKey string = "kube_resource_namespace"
	kubeResourceKindKey      string = "kube_resource_kind"
)

func newKubeapiserverCheck(baseCheck baseCheck, kubeResource *compliance.KubernetesResource) (*kubeApiserverCheck, error) {
	check := &kubeApiserverCheck{
		baseCheck:    baseCheck,
		kubeResource: kubeResource,
	}

	kubeResource := res.KubeApiserver

	if len(kubeResource.Kind) == 0 {
		return nil, fmt.Errorf("cannot run Kubeapiserver check, resource kind is empty")
	}

	if len(kubeResource.APIRequest.Verb) == 0 {
		return nil, fmt.Errorf("cannot run Kubeapiserver check, action verb is empty")
	}

	return check, nil
}

func (c *kubeApiserverCheck) Run() error {
	log.Debugf("%s: kubeapiserver check: %s", c.ruleID, c.kubeResource.String())

	resourceSchema := schema.GroupVersionResource{
		Group:    kubeResource.Group,
		Resource: kubeResource.Kind,
		Version:  kubeResource.Version,
	}
	resourceDef := c.KubeClient().Resource(resourceSchema)

	var resourceAPI dynamic.ResourceInterface
	if len(kubeResource.Namespace) > 0 {
		resourceAPI = resourceDef.Namespace(kubeResource.Namespace)
	} else {
		resourceAPI = resourceDef
	}

	var resources []unstructured.Unstructured

	api := kubeResource.APIRequest
	switch api.Verb {
	case "get":
		if len(api.ResourceName) == 0 {
			return nil, fmt.Errorf("unable to use 'get' apirequest without resource name")
		}
		resource, err := resourceAPI.Get(kubeResource.APIRequest.ResourceName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("unable to get Kube resource:'%v', ns:'%s' name:'%s', err: %v", resourceSchema, kubeResource.Namespace, api.ResourceName, err)
		}
		resources = []unstructured.Unstructured{*resource}
	case "list":
		list, err := resourceAPI.List(metav1.ListOptions{
			LabelSelector: kubeResource.LabelSelector,
			FieldSelector: kubeResource.FieldSelector,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to list Kube resources:'%v', ns:'%s' name:'%s', err: %v", resourceSchema, kubeResource.Namespace, api.ResourceName, err)
		}
		resources = list.Items
	}

	log.Debugf("%s: Got %d resources", ruleID, len(resources))

	return &kubeResourceIterator{
		resources: resources,
	}, nil
}

type kubeResourceIterator struct {
	resources []unstructured.Unstructured
	index     int
}

func (c *kubeApiserverCheck) reportResource(p unstructured.Unstructured) error {
	kv := event.Data{}

	for _, field := range c.kubeResource.Report {
		switch field.Kind {
		case compliance.PropertyKindJSONQuery:
			reportValue, valueFound, err := jsonquery.RunSingleOutput(field.Property, p.Object)
			if err != nil {
				return fmt.Errorf("unable to report field: '%s' for kubernetes object '%s / %s / %s' - json query error: %v", field.Property, p.GroupVersionKind().String(), p.GetNamespace(), p.GetName(), err)
			}

			if !valueFound {
				continue
			}

			reportName := field.Property
			if len(field.As) > 0 {
				reportName = field.As
			}
			if len(field.Value) > 0 {
				reportValue = field.Value
			}

			kv[reportName] = reportValue
		default:
			return fmt.Errorf("unsupported field kind value: '%s' for kubeApiserver resource", field.Kind)
		}
		return instance, nil
	}
	return nil, errors.New("out of bounds iteration")
}

func (it *kubeResourceIterator) Done() bool {
	return it.index >= len(it.resources)
}

func kubeResourceJQ(resource unstructured.Unstructured) eval.Function {
	return func(_ *eval.Instance, args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf(`invalid number of arguments, expecting 1 got %d`, len(args))
		}
		query, ok := args[0].(string)
		if !ok {
			return nil, errors.New(`expecting string value for query argument"`)
		}

		v, _, err := jsonquery.RunSingleOutput(query, resource.Object)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}
