// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	k8spluginv1alpha1 "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// checkLabel verifies if the expected label exists and returns bool
func checkLabel(labels map[string]string) bool {

	_, ok := labels["emco/deployment-id"]
	if !ok {
		return false
	}
	return true
}

// returnLabel verifies if the expected label exists and returns a map
func returnLabel(labels map[string]string) map[string]string {

	l, ok := labels["emco/deployment-id"]
	if !ok {
		return nil
	}
	return map[string]string{
		"emco/deployment-id": l,
	}
}

// listResources lists resources based on the selectors provided
// The data is returned in the pointer to the runtime.Object
// provided as argument.
func listResources(cli client.Client, namespace string,
	labelSelector map[string]string, returnData client.ObjectList) error {

	listOptions := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(labelSelector),
	}

	err := cli.List(context.TODO(), returnData, listOptions)
	if err != nil {
		log.Printf("Failed to list CRs: %v", err)
		return err
	}

	return nil
}

// listClusterResources lists non-namespace resources based
// on the selectors provided.
// The data is returned in the pointer to the runtime.Object
// provided as argument.
func listClusterResources(cli client.Client,
	labelSelector map[string]string, returnData client.ObjectList) error {
	return listResources(cli, "", labelSelector, returnData)
}

func GetCRForResource(cli client.Client, item *unstructured.Unstructured, namespace string) (*k8spluginv1alpha1.ResourceBundleState, error) {
	// Find the CRs which track this resource via the labelselector
	crSelector := returnLabel(item.GetLabels())
	if crSelector == nil {
		log.Println("We should not be here. The predicate should have filtered this resource")
		return nil, fmt.Errorf("unexpected Error: Resource not filtered by predicate")
	}

	name := crSelector["emco/deployment-id"]
	if namespace != "" {
		// Namespace is known and the name is known
		// get a single resource
		return getCRForResource(cli, name, namespace)
	}

	// Namespace is not known and the name is known
	// assume it is a cluster scope item
	rbStatusList, err := getCRListForResource(cli, crSelector, metav1.NamespaceAll)
	if err != nil {
		return nil, err
	}

	if len(rbStatusList.Items) == 0 {
		return nil, fmt.Errorf("failed to find ResourceBundleState %s", name)
	}

	var crResource *k8spluginv1alpha1.ResourceBundleState
	found := false
	for _, rbs := range rbStatusList.Items {
		if rbs.Name == name {
			found = true
			crResource = &rbs
		}
	}

	if !found {
		return nil, fmt.Errorf("failed to find ResourceBundleState %s", name)
	}

	return crResource, nil
}

// Assume only one CR for label, multiple throws error
func getCRForResource(cli client.Client, name, namespace string) (*k8spluginv1alpha1.ResourceBundleState, error) {
	// Name of resource is same as the label value
	var namespaced types.NamespacedName
	namespaced.Name = name
	if namespace == "" {
		namespaced.Namespace = "default"
	} else {
		namespaced.Namespace = namespace
	}

	rbStatus := &k8spluginv1alpha1.ResourceBundleState{}
	err := cli.Get(context.TODO(), namespaced, rbStatus)
	if err != nil {
		return nil, err
	}

	return rbStatus, nil
}

func getCRListForResource(client client.Client, crSelector map[string]string, namespace string) (*k8spluginv1alpha1.ResourceBundleStateList, error) {
	// Get the CRs which have this label and update them all
	// Ideally, we will have only one CR, but there is nothing
	// preventing the creation of multiple.
	rbStatusList := &k8spluginv1alpha1.ResourceBundleStateList{}
	err := listResources(client, namespace, crSelector, rbStatusList)
	if err != nil {
		return nil, err
	}

	return rbStatusList, nil
}

func UpdateStatus(cr *k8spluginv1alpha1.ResourceBundleState, item *unstructured.Unstructured, name, namespace string) (bool, error) {

	switch item.GetObjectKind().GroupVersionKind() {
	case schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}:
		return ConfigMapUpdateStatus(cr, item)
	case schema.GroupVersionKind{Version: "v1", Kind: "Service"}:
		return ServiceUpdateStatus(cr, item)
	case schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"}:
		return DaemonSetUpdateStatus(cr, item)
	case schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}:
		return DeploymentUpdateStatus(cr, item)
	case schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"}:
		return JobUpdateStatus(cr, item)
	case schema.GroupVersionKind{Version: "v1", Kind: "Pod"}:
		return PodUpdateStatus(cr, item)
	case schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"}:
		return StatefulSetUpdateStatus(cr, item)
	case schema.GroupVersionKind{Group: "certificates.k8s.io", Version: "v1", Kind: "CertificateSigningRequest"}:
		return CsrUpdateStatus(cr, item)
	}
	return false, fmt.Errorf("Resource not supported explicitly")
}

func DeleteObj(cr *k8spluginv1alpha1.ResourceBundleState, name, namespace string, gvk schema.GroupVersionKind) (bool, error) {

	switch gvk {
	case schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}:
		return ConfigMapDeleteObj(cr, name)
	case schema.GroupVersionKind{Version: "v1", Kind: "Service"}:
		return ServiceDeleteObj(cr, name)
	case schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"}:
		return DaemonSetDeleteObj(cr, name)
	case schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}:
		return DeploymentDeleteObj(cr, name)
	case schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"}:
		return JobDeleteObj(cr, name)
	case schema.GroupVersionKind{Version: "v1", Kind: "Pod"}:
		return PodDeleteObj(cr, name)
	case schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"}:
		return StatefulSetDeleteObj(cr, name)
	case schema.GroupVersionKind{Group: "certificates.k8s.io", Version: "v1", Kind: "CertificateSigningRequest"}:
		return CsrDeleteObj(cr, name)
	}
	return false, fmt.Errorf("Resource not supported explicitly")
}

func ClearLastApplied(annotations map[string]string) map[string]string {
	_, ok := annotations["kubectl.kubernetes.io/last-applied-configuration"]
	if ok {
		annotations["kubectl.kubernetes.io/last-applied-configuration"] = ""
	}
	return annotations
}

func DeleteResourceStatusCR(cr *k8spluginv1alpha1.ResourceBundleState, name, namespace string, gvk schema.GroupVersionKind) (bool, error) {
	var found bool
	length := len(cr.Status.ResourceStatuses)
	for i, rstatus := range cr.Status.ResourceStatuses {
		if (rstatus.Group == gvk.Group) && (rstatus.Version == gvk.Version) && (rstatus.Kind == gvk.Kind) && (rstatus.Name == name) && (rstatus.Namespace == namespace) {
			found = true
			//Delete that status from the array
			cr.Status.ResourceStatuses[i] = cr.Status.ResourceStatuses[length-1]
			cr.Status.ResourceStatuses[length-1] = k8spluginv1alpha1.ResourceStatus{}
			cr.Status.ResourceStatuses = cr.Status.ResourceStatuses[:length-1]
			break
		}
	}
	return found, nil
}

func UpdateResourceStatusCR(cr *k8spluginv1alpha1.ResourceBundleState, item *unstructured.Unstructured, name, namespace string) (bool, error) {
	var found bool
	var res k8spluginv1alpha1.ResourceStatus

	// Clear up some fields to reduce size
	item.SetManagedFields([]metav1.ManagedFieldsEntry{})
	item.SetAnnotations(ClearLastApplied(item.GetAnnotations()))

	unstruct := item.UnstructuredContent()
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstruct, &res)
	if err != nil {
		log.Println("DefaultUnstructuredConverter error", name, namespace, err)
		return found, fmt.Errorf("Unknown resource")
	}

	group := item.GetObjectKind().GroupVersionKind().Group
	version := item.GetObjectKind().GroupVersionKind().Version
	kind := item.GetObjectKind().GroupVersionKind().Kind
	var index int
	for i, rstatus := range cr.Status.ResourceStatuses {
		if (rstatus.Group == group) && (rstatus.Version == version) && (rstatus.Kind == kind) && (rstatus.Name == name) && (rstatus.Namespace == namespace) {
			found = true
			index = i
			break
		}
	}
	if found {
		// Replace
		resBytes, err := json.Marshal(item)
		if err != nil {
			log.Println("json Marshal error for resource::", item, err, index)
			return found, err
		}
		p := &cr.Status.ResourceStatuses[index]
		p.Res = make([]byte, len(resBytes))
		copy(p.Res, resBytes)
	} else {
		resBytes, err := json.Marshal(item)
		if err != nil {
			log.Println("json Marshal error for resource::", item, err)
			return found, err
		}
		// Add resource to ResourceMap
		res := k8spluginv1alpha1.ResourceStatus{
			Group:     group,
			Version:   version,
			Kind:      kind,
			Name:      name,
			Namespace: namespace,
		}
		res.Res = make([]byte, len(resBytes))
		copy(res.Res, resBytes)
		cr.Status.ResourceStatuses = append(cr.Status.ResourceStatuses, res)
	}
	return found, nil
}

// GetServerResources returns a GVK list of all cluster defined resources
func GetServerResources(client *discovery.DiscoveryClient) ([]*schema.GroupVersionKind, error) {
	resourcesLists, err := client.ServerPreferredResources()
	if err != nil {
		return nil, fmt.Errorf("failed to get server resources :%q", err)
	}

	resourcesGVK := []*schema.GroupVersionKind{}
	for _, resourcesList := range resourcesLists {
		gv, err := schema.ParseGroupVersion(resourcesList.GroupVersion)
		if err != nil {
			continue
		}

		for _, resource := range resourcesList.APIResources {
			// Skip resources without "watch" verb
			found := false
			for _, verb := range resource.Verbs {
				if verb == "watch" {
					found = true
					break
				}
			}

			if !found {
				continue
			}

			resourcesGVK = append(resourcesGVK, &schema.GroupVersionKind{
				Group:   gv.Group,
				Version: gv.Version,
				Kind:    resource.Kind,
			})
		}
	}

	return resourcesGVK, nil
}
