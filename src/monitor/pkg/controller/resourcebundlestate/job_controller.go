// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package resourcebundlestate

import (
	"context"
	"log"

	"gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"

	v1 "k8s.io/api/batch/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	pkgerrors "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AddJobController the new controller to the controller manager
func AddJobController(mgr manager.Manager) error {
	return addJobController(mgr, newJobReconciler(mgr))
}

func addJobController(mgr manager.Manager, r *jobReconciler) error {
	// Create a new controller
	c, err := controller.New("Job-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secondar resource Jobs
	// Predicate filters Job which don't have the k8splugin label
	err = c.Watch(&source.Kind{Type: &v1.Job{}}, &handler.EnqueueRequestForObject{}, &jobPredicate{})
	if err != nil {
		return err
	}

	return nil
}

func newJobReconciler(m manager.Manager) *jobReconciler {
	return &jobReconciler{client: m.GetClient()}
}

type jobReconciler struct {
	client client.Client
}

// Reconcile implements the loop that will update the ResourceBundleState CR
// whenever we get any updates from all the jobs we watch.
func (r *jobReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Updating ResourceBundleState for Job: %+v\n", req)

	job := &v1.Job{}
	err := r.client.Get(context.TODO(), req.NamespacedName, job)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("Job not found: %+v. Remove from CR if it is stored there.\n", req.NamespacedName)
			// Remove the Job's status from StatusList
			// This can happen if we get the DeletionTimeStamp event
			// after the Job has been deleted.
			DeleteFromAllCRs(r, req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get Job: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}

	err = UpdateCR(r, job)
	if err != nil {
		// Requeue the update
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *jobReconciler) GetClient() client.Client {
	return r.client
}

func (r *jobReconciler) DeleteObj(cr *v1alpha1.ResourceBundleState, name string) bool {
	var found bool
	length := len(cr.Status.JobStatuses)
	for i, rstatus := range cr.Status.JobStatuses {
		if rstatus.Name == name {
			found = true
			//Delete that status from the array
			cr.Status.JobStatuses[i] = cr.Status.JobStatuses[length-1]
			cr.Status.JobStatuses[length-1].Status = v1.JobStatus{}
			cr.Status.JobStatuses = cr.Status.JobStatuses[:length-1]
			break
		}
	}
	return found
}

func (r *jobReconciler) UpdateStatus(cr *v1alpha1.ResourceBundleState, obj runtime.Object) (bool, error) {

	var found bool
	job, ok := obj.(*v1.Job)
	if !ok {
		return found, pkgerrors.Errorf("Unknown resource %v", obj)
	}
	// Update status after searching for it in the list of resourceStatuses
	for i, rstatus := range cr.Status.JobStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == job.Name {
			job.Status.DeepCopyInto(&cr.Status.JobStatuses[i].Status)
			found = true
			break
		}
	}
	if !found {
		// Add it to CR
		c := v1.Job{
			TypeMeta:   job.TypeMeta,
			ObjectMeta: job.ObjectMeta,
			Spec:       job.Spec,
			Status:     job.Status,
		}
		c.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		c.Annotations = ClearLastApplied(c.Annotations)
		cr.Status.JobStatuses = append(cr.Status.JobStatuses, c)
	}
	return found, nil
}
