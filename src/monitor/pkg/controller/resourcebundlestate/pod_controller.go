// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package resourcebundlestate

import (
	"context"
	"log"

	"gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//	"k8s.io/apimachinery/pkg/types"
	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AddPodController the new controller to the controller manager
func AddPodController(mgr manager.Manager) error {
	return addPodController(mgr, newPodReconciler(mgr))
}

func addPodController(mgr manager.Manager, r *podReconciler) error {
	// Create a new controller
	c, err := controller.New("Pod-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secondar resource Pods
	// Predicate filters pods which don't have the k8splugin label
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForObject{}, &podPredicate{})
	if err != nil {
		return err
	}

	return nil
}

func newPodReconciler(m manager.Manager) *podReconciler {
	return &podReconciler{client: m.GetClient()}
}

type podReconciler struct {
	client client.Client
}

func (r *podReconciler) GetClient() client.Client {
	return r.client
}

// Reconcile implements the loop that will update the ResourceBundleState CR
// whenever we get any updates from all the pods we watch.
func (r *podReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Updating ResourceBundleState for Pod: %+v\n", req)

	pod := &corev1.Pod{}
	err := r.client.Get(context.TODO(), req.NamespacedName, pod)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("Pod not found: %+v. Remove from CR if it is stored there.\n", req.NamespacedName)
			// Remove the Pod's status from StatusList
			// This can happen if we get the DeletionTimeStamp event
			// after the POD has been deleted.
			//r.deletePodFromAllCRs(req.NamespacedName)
			DeleteFromAllCRs(r, req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get pod: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}

	err = UpdateCR(r, pod)
	if err != nil {
		// Requeue the update
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *podReconciler) UpdateStatus(cr *v1alpha1.ResourceBundleState, obj runtime.Object) (bool, error) {
	var found bool
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return found, pkgerrors.Errorf("Unknown resource %v", obj)
	}
	for i, rstatus := range cr.Status.PodStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == pod.Name {
			pod.Status.DeepCopyInto(&cr.Status.PodStatuses[i].Status)
			found = true
			break
		}
	}
	if !found {
		// Add it to CR
		ps := corev1.Pod{
			TypeMeta:   pod.TypeMeta,
			ObjectMeta: pod.ObjectMeta,
			Status:     pod.Status,
			Spec:       pod.Spec,
		}
		ps.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		ps.Annotations = ClearLastApplied(ps.Annotations)
		cr.Status.PodStatuses = append(cr.Status.PodStatuses, ps)
	}
	return found, nil
}

func (r *podReconciler) DeleteObj(cr *v1alpha1.ResourceBundleState, name string) bool {
	var found bool
	length := len(cr.Status.PodStatuses)
	for i, rstatus := range cr.Status.PodStatuses {
		if rstatus.Name == name {
			found = true
			//Delete that status from the array
			cr.Status.PodStatuses[i] = cr.Status.PodStatuses[length-1]
			cr.Status.PodStatuses[length-1] = corev1.Pod{}
			cr.Status.PodStatuses = cr.Status.PodStatuses[:length-1]
			break
		}
	}
	return found
}
