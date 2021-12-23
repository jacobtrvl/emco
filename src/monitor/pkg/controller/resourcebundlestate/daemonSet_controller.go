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

// AddDaemonSetController the new controller to the controller manager
func AddDaemonSetController(mgr manager.Manager) error {
	return addDaemonSetController(mgr, newDaemonSetReconciler(mgr))
}

func addDaemonSetController(mgr manager.Manager, r *daemonSetReconciler) error {
	// Create a new controller
	c, err := controller.New("Daemonset-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secondar resource DaemonSets
	// Predicate filters DaemonSets which don't have the k8splugin label
	err = c.Watch(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForObject{}, &daemonSetPredicate{})
	if err != nil {
		return err
	}

	return nil
}

func newDaemonSetReconciler(m manager.Manager) *daemonSetReconciler {
	return &daemonSetReconciler{client: m.GetClient()}
}

type daemonSetReconciler struct {
	client client.Client
}

// Reconcile implements the loop that will update the ResourceBundleState CR
// whenever we get any updates from all the daemonSets we watch.
func (r *daemonSetReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Updating ResourceBundleState for DaemonSet: %+v\n", req)

	ds := &appsv1.DaemonSet{}
	err := r.client.Get(context.TODO(), req.NamespacedName, ds)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("DaemonSet not found: %+v. Remove from CR if it is stored there.\n", req.NamespacedName)
			// Remove the DaemonSet's status from StatusList
			// This can happen if we get the DeletionTimeStamp event
			// after the DaemonSet has been deleted.
			DeleteFromAllCRs(r, req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get daemonSet: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}
	err = UpdateCR(r, ds)
	if err != nil {
		// Requeue the update
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *daemonSetReconciler) GetClient() client.Client {
	return r.client
}

func (r *daemonSetReconciler) DeleteObj(cr *v1alpha1.ResourceBundleState, name string) bool {
	var found bool
	length := len(cr.Status.DaemonSetStatuses)
	for i, rstatus := range cr.Status.DaemonSetStatuses {
		if rstatus.Name == name {
			found = true
			//Delete that status from the array
			cr.Status.DaemonSetStatuses[i] = cr.Status.DaemonSetStatuses[length-1]
			cr.Status.DaemonSetStatuses[length-1].Status = appsv1.DaemonSetStatus{}
			cr.Status.DaemonSetStatuses = cr.Status.DaemonSetStatuses[:length-1]
			break
		}
	}
	return found
}

func (r *daemonSetReconciler) UpdateStatus(cr *v1alpha1.ResourceBundleState, obj runtime.Object) (bool, error) {

	var found bool
	ds, ok := obj.(*v1.DaemonSet)
	if !ok {
		return found, pkgerrors.Errorf("Unknown resource %v", obj)
	}
	// Update status after searching for it in the list of resourceStatuses
	for i, rstatus := range cr.Status.DaemonSetStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == ds.Name {
			ds.Status.DeepCopyInto(&cr.Status.DaemonSetStatuses[i].Status)
			found = true
			break
		}
	}
	if !found {
		// Add it to CR
		c := v1.DaemonSet{
			TypeMeta:   ds.TypeMeta,
			ObjectMeta: ds.ObjectMeta,
			Spec:       ds.Spec,
			Status:     ds.Status,
		}
		c.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		c.Annotations = ClearLastApplied(c.Annotations)
		cr.Status.DaemonSetStatuses = append(cr.Status.DaemonSetStatuses, c)
	}
	return found, nil
}
