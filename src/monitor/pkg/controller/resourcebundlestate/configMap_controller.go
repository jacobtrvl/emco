// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package resourcebundlestate

import (
	"context"
	"log"

	"gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"

	pkgerrors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
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

// AddConfigMapController the new controller to the controller manager
func AddConfigMapController(mgr manager.Manager) error {
	return addConfigMapController(mgr, newConfigMapReconciler(mgr))
}

func addConfigMapController(mgr manager.Manager, r *configMapReconciler) error {
	// Create a new controller
	c, err := controller.New("ConfigMap-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secondar resource ConfigMaps
	// Predicate filters Service which don't have the k8splugin label
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForObject{}, &configMapPredicate{})
	if err != nil {
		return err
	}

	return nil
}

func newConfigMapReconciler(m manager.Manager) *configMapReconciler {
	return &configMapReconciler{client: m.GetClient()}
}

type configMapReconciler struct {
	client client.Client
}

// Reconcile implements the loop that will update the ResourceBundleState CR
// whenever we get any updates from all the ConfigMaps we watch.
func (r *configMapReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Updating ResourceBundleState for ConfigMap: %+v\n", req)

	cm := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), req.NamespacedName, cm)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("ConfigMap not found: %+v. Remove from CR if it is stored there.\n", req.NamespacedName)
			// Remove the ConfigMap's status from StatusList
			// This can happen if we get the DeletionTimeStamp event
			// after the ConfigMap has been deleted.
			DeleteFromAllCRs(r, req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get ConfigMap: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}
	err = UpdateCR(r, cm)
	if err != nil {
		// Requeue the update
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *configMapReconciler) GetClient() client.Client {
	return r.client
}

func (r *configMapReconciler) DeleteObj(cr *v1alpha1.ResourceBundleState, name string) bool {
	var found bool
	length := len(cr.Status.ConfigMapStatuses)
	for i, rstatus := range cr.Status.ConfigMapStatuses {
		if rstatus.Name == name {
			found = true
			//Delete that status from the array
			cr.Status.ConfigMapStatuses[i] = cr.Status.ConfigMapStatuses[length-1]
			cr.Status.ConfigMapStatuses = cr.Status.ConfigMapStatuses[:length-1]
			break
		}
	}
	return found
}

func (r *configMapReconciler) UpdateStatus(cr *v1alpha1.ResourceBundleState, obj runtime.Object) (bool, error) {

	var found bool
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		return found, pkgerrors.Errorf("Unknown resource %v", obj)
	}

	// Update status after searching for it in the list of resourceStatuses
	for _, rstatus := range cr.Status.ConfigMapStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == cm.Name {
			found = true
			break
		}
	}

	if !found {
		// Add it to CR
		c := corev1.ConfigMap{
			TypeMeta:   cm.TypeMeta,
			ObjectMeta: cm.ObjectMeta,
		}
		c.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		c.Annotations = ClearLastApplied(c.Annotations)
		cr.Status.ConfigMapStatuses = append(cr.Status.ConfigMapStatuses, c)
	}

	return found, nil
}
