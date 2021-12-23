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

// AddStatefulSetController the new controller to the controller manager
func AddStatefulSetController(mgr manager.Manager) error {
	return addStatefulSetController(mgr, newStatefulSetReconciler(mgr))
}

func addStatefulSetController(mgr manager.Manager, r *statefulSetReconciler) error {
	// Create a new controller
	c, err := controller.New("Statefulset-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secondar resource StatefulSets
	// Predicate filters StatefulSet which don't have the k8splugin label
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForObject{}, &statefulSetPredicate{})
	if err != nil {
		return err
	}

	return nil
}

func newStatefulSetReconciler(m manager.Manager) *statefulSetReconciler {
	return &statefulSetReconciler{client: m.GetClient()}
}

type statefulSetReconciler struct {
	client client.Client
}

// Reconcile implements the loop that will update the ResourceBundleState CR
// whenever we get any updates from all the StatefulSets we watch.
func (r *statefulSetReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Updating ResourceBundleState for StatefulSet: %+v\n", req)

	sfs := &appsv1.StatefulSet{}
	err := r.client.Get(context.TODO(), req.NamespacedName, sfs)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("StatefulSet not found: %+v. Remove from CR if it is stored there.\n", req.NamespacedName)
			// Remove the StatefulSet's status from StatusList
			// This can happen if we get the DeletionTimeStamp event
			// after the StatefulSet has been deleted.
			DeleteFromAllCRs(r, req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get statefulSet: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}

	err = UpdateCR(r, sfs)
	if err != nil {
		// Requeue the update
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *statefulSetReconciler) GetClient() client.Client {
	return r.client
}

func (r *statefulSetReconciler) DeleteObj(cr *v1alpha1.ResourceBundleState, name string) bool {
	var found bool
	length := len(cr.Status.StatefulSetStatuses)
	for i, rstatus := range cr.Status.StatefulSetStatuses {
		if rstatus.Name == name {
			found = true
			//Delete that status from the array
			cr.Status.StatefulSetStatuses[i] = cr.Status.StatefulSetStatuses[length-1]
			cr.Status.StatefulSetStatuses[length-1].Status = appsv1.StatefulSetStatus{}
			cr.Status.StatefulSetStatuses = cr.Status.StatefulSetStatuses[:length-1]
			break
		}
	}
	return found
}

func (r *statefulSetReconciler) UpdateStatus(cr *v1alpha1.ResourceBundleState, obj runtime.Object) (bool, error) {

	var found bool
	ss, ok := obj.(*v1.StatefulSet)
	if !ok {
		return found, pkgerrors.Errorf("Unknown resource %v", obj)
	}
	// Update status after searching for it in the list of resourceStatuses
	for i, rstatus := range cr.Status.StatefulSetStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == ss.Name {
			ss.Status.DeepCopyInto(&cr.Status.StatefulSetStatuses[i].Status)
			found = true
			break
		}
	}
	if !found {
		// Add it to CR
		c := v1.StatefulSet{
			TypeMeta:   ss.TypeMeta,
			ObjectMeta: ss.ObjectMeta,
			Spec:       ss.Spec,
			Status:     ss.Status,
		}
		c.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		c.Annotations = ClearLastApplied(c.Annotations)
		cr.Status.StatefulSetStatuses = append(cr.Status.StatefulSetStatuses, c)
	}
	return found, nil
}
