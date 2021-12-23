// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package resourcebundlestate

import (
	"context"
	"log"

	"gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"

	pkgerrors "github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AddDeploymentController the new controller to the controller manager
func AddDeploymentController(mgr manager.Manager) error {
	return addDeploymentController(mgr, newDeploymentReconciler(mgr))
}

func addDeploymentController(mgr manager.Manager, r *deploymentReconciler) error {
	// Create a new controller
	c, err := controller.New("Deployment-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secondar resource Deployments
	// Predicate filters Deployment which don't have the k8splugin label
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForObject{}, &deploymentPredicate{})
	if err != nil {
		return err
	}

	return nil
}

func newDeploymentReconciler(m manager.Manager) *deploymentReconciler {
	return &deploymentReconciler{client: m.GetClient()}
}

type deploymentReconciler struct {
	client client.Client
}

// Reconcile implements the loop that will update the ResourceBundleState CR
// whenever we get any updates from all the deployments we watch.
func (r *deploymentReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Updating ResourceBundleState for Deployment: %+v\n", req)

	dep := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), req.NamespacedName, dep)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("Deployment not found: %+v. Remove from CR if it is stored there.\n", req.NamespacedName)
			// Remove the Deployment's status from StatusList
			// This can happen if we get the DeletionTimeStamp event
			// after the Deployment has been deleted.
			DeleteFromAllCRs(r, req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get deployment: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}

	// Find the CRs which track this deployment via the labelselector
	crSelector := returnLabel(dep.GetLabels())
	if crSelector == nil {
		log.Println("We should not be here. The predicate should have filtered this Deployment")
	}

	// Get the CRs which have this label and update them all
	// Ideally, we will have only one CR, but there is nothing
	// preventing the creation of multiple.
	// TODO: Consider using an admission validating webook to prevent multiple
	rbStatusList := &v1alpha1.ResourceBundleStateList{}
	err = listResources(r.client, req.Namespace, crSelector, rbStatusList)
	if err != nil || len(rbStatusList.Items) == 0 {
		log.Printf("Did not find any CRs tracking this resource\n")
		return reconcile.Result{}, nil
	}

	err = UpdateCR(r, dep)
	if err != nil {
		// Requeue the update
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *deploymentReconciler) GetClient() client.Client {
	return r.client
}

func (r *deploymentReconciler) DeleteObj(cr *v1alpha1.ResourceBundleState, name string) bool {
	var found bool
	length := len(cr.Status.DeploymentStatuses)
	for i, rstatus := range cr.Status.DeploymentStatuses {
		if rstatus.Name == name {
			found = true
			//Delete that status from the array
			cr.Status.DeploymentStatuses[i] = cr.Status.DeploymentStatuses[length-1]
			cr.Status.DeploymentStatuses[length-1].Status = appsv1.DeploymentStatus{}
			cr.Status.DeploymentStatuses = cr.Status.DeploymentStatuses[:length-1]
			break
		}
	}
	return found
}

func (r *deploymentReconciler) UpdateStatus(cr *v1alpha1.ResourceBundleState, obj runtime.Object) (bool, error) {

	var found bool
	dm, ok := obj.(*v1.Deployment)
	if !ok {
		return found, pkgerrors.Errorf("Unknown resource %v", obj)
	}
	// Update status after searching for it in the list of resourceStatuses
	for i, rstatus := range cr.Status.DeploymentStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == dm.Name {
			dm.Status.DeepCopyInto(&cr.Status.DeploymentStatuses[i].Status)
			found = true
			break
		}
	}
	if !found {
		// Add it to CR
		c := v1.Deployment{
			TypeMeta:   dm.TypeMeta,
			ObjectMeta: dm.ObjectMeta,
			Spec:       dm.Spec,
			Status:     dm.Status,
		}
		c.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		c.Annotations = ClearLastApplied(c.Annotations)
		cr.Status.DeploymentStatuses = append(cr.Status.DeploymentStatuses, c)
	}
	return found, nil
}
