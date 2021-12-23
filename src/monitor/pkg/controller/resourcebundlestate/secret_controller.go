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

// AddSecretController the new controller to the controller manager
func AddSecretController(mgr manager.Manager) error {
	return addSecretController(mgr, newSecretReconciler(mgr))
}

func addSecretController(mgr manager.Manager, r *secretReconciler) error {
	// Create a new controller
	c, err := controller.New("Secret-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secondar resource Secret
	// Predicate filters Secret which don't have the k8splugin label
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForObject{}, &secretPredicate{})
	if err != nil {
		return err
	}

	return nil
}

func newSecretReconciler(m manager.Manager) *secretReconciler {
	return &secretReconciler{client: m.GetClient()}
}

type secretReconciler struct {
	client client.Client
}

// Reconcile implements the loop that will update the ResourceBundleState CR
// whenever we get any updates from all the Secrets we watch.
func (r *secretReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Updating ResourceBundleState for Secret: %+v\n", req)

	sec := &corev1.Secret{}
	err := r.client.Get(context.TODO(), req.NamespacedName, sec)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("Secret not found: %+v. Remove from CR if it is stored there.\n", req.NamespacedName)
			// Remove the Secret's status from StatusList
			// This can happen if we get the DeletionTimeStamp event
			// after the Secret has been deleted.
			DeleteFromAllCRs(r, req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get Secret: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}
	err = UpdateCR(r, sec)

	if err != nil {
		// Requeue the update
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *secretReconciler) GetClient() client.Client {
	return r.client
}

func (r *secretReconciler) DeleteObj(cr *v1alpha1.ResourceBundleState, name string) bool {
	var found bool
	length := len(cr.Status.SecretStatuses)
	for i, rstatus := range cr.Status.SecretStatuses {
		if rstatus.Name == name {
			found = true
			//Delete that status from the array
			cr.Status.SecretStatuses[i] = cr.Status.SecretStatuses[length-1]
			cr.Status.SecretStatuses = cr.Status.SecretStatuses[:length-1]
			break
		}
	}
	return found
}

func (r *secretReconciler) UpdateStatus(cr *v1alpha1.ResourceBundleState, obj runtime.Object) (bool, error) {

	var found bool
	sec, ok := obj.(*v1.Secret)
	if !ok {
		return found, pkgerrors.Errorf("Unknown resource %v", obj)
	}
	// Update status after searching for it in the list of resourceStatuses
	for _, rstatus := range cr.Status.SecretStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == sec.Name {
			found = true
			break
		}
	}
	if !found {
		// Add it to CR
		c := v1.Secret{
			TypeMeta:   sec.TypeMeta,
			ObjectMeta: sec.ObjectMeta,
		}
		c.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		c.Annotations = ClearLastApplied(c.Annotations)
		cr.Status.SecretStatuses = append(cr.Status.SecretStatuses, c)
	}
	return found, nil
}
