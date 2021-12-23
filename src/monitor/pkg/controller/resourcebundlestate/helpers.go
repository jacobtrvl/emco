// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package resourcebundlestate

import (
	"context"
	"log"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceProvider interface {
	GetClient() client.Client
	UpdateStatus(*v1alpha1.ResourceBundleState, runtime.Object) (bool, error)
	DeleteObj(*v1alpha1.ResourceBundleState, string) bool
}

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
	labelSelector map[string]string, returnData runtime.Object) error {

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
	labelSelector map[string]string, returnData runtime.Object) error {
	return listResources(cli, "", labelSelector, returnData)
}

func UpdateCR(r ResourceProvider, item runtime.Object) error {

	//unstruct := item.(*unstructured.Unstructured)
	// convert the runtime.Object to unstructured.Unstructured
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(item)
	if err != nil {
		log.Printf("Error to unstruct")
	}
	u := unstructured.Unstructured{Object: unstructuredObj}

	// Find the CRs which track this pod via the labelselector
	crSelector := returnLabel(u.GetLabels())
	if crSelector == nil {
		log.Println("We should not be here. The predicate should have filtered this Pod")
	}
	// Get the CRs which have this label and update them all
	// Ideally, we will have only one CR, but there is nothing
	// preventing the creation of multiple.
	// TODO: Consider using an admission validating webook to prevent multiple
	rbStatusList := &v1alpha1.ResourceBundleStateList{}
	err = listResources(r.GetClient(), u.GetNamespace(), crSelector, rbStatusList)
	if err != nil || len(rbStatusList.Items) == 0 {
		return nil
	}

	for _, cr := range rbStatusList.Items {
		// Not scheduled for deletion
		if u.GetDeletionTimestamp() == nil {
			err := UpdateSingleCR(r, &cr, item)
			if err != nil {
				log.Printf("UpdateCR Error: %s", err.Error())
				return err
			}
		} else {
			// Scheduled for deletion
			return DeleteFromSingleCR(r, &cr, u.GetName())
		}
	}
	return nil
}

func UpdateSingleCR(r ResourceProvider, cr *v1alpha1.ResourceBundleState, item runtime.Object) error {

	found, err := r.UpdateStatus(cr, item)
	if err != nil {
		return err
	}
	// Exited for loop with no status found
	// Increment the number of tracked resources
	if !found {
		cr.Status.ResourceCount++
	}
	err = r.GetClient().Status().Update(context.TODO(), cr)
	if err != nil {
		return err
	}

	return nil
}

func DeleteFromSingleCR(r ResourceProvider, cr *v1alpha1.ResourceBundleState, name string) error {

	found := r.DeleteObj(cr, name)
	if found {
		err := r.GetClient().Status().Update(context.TODO(), cr)
		if err != nil {
			log.Printf("failed to update rbstate: %v\n", err)
			return err
		}
	}
	return nil
}

func DeleteFromAllCRs(r ResourceProvider, namespacedName types.NamespacedName) error {

	rbStatusList := &v1alpha1.ResourceBundleStateList{}
	err := listResources(r.GetClient(), namespacedName.Namespace, nil, rbStatusList)
	if err != nil || len(rbStatusList.Items) == 0 {
		log.Printf("Did not find any CRs tracking this resource\n")
		return pkgerrors.Errorf("Did not find any CRs tracking this resource")
	}
	for _, cr := range rbStatusList.Items {
		return DeleteFromSingleCR(r, &cr, namespacedName.Name)
	}
	return nil
}

func ClearLastApplied(annotations map[string]string) (map[string]string) {
	_, ok := annotations["kubectl.kubernetes.io/last-applied-configuration"]
	if ok {
		annotations["kubectl.kubernetes.io/last-applied-configuration"] = ""
	}
	return annotations
}