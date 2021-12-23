// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package resourcebundlestate

import (
	"context"
	"log"

	"gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	//"k8s.io/apimachinery/pkg/types"
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

// AddServiceController the new controller to the controller manager
func AddServiceController(mgr manager.Manager) error {
	return addServiceController(mgr, newServiceReconciler(mgr))
}

func addServiceController(mgr manager.Manager, r *serviceReconciler) error {
	// Create a new controller
	c, err := controller.New("Service-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secondar resource Services
	// Predicate filters Service which don't have the k8splugin label
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForObject{}, &servicePredicate{})
	if err != nil {
		return err
	}

	return nil
}

func newServiceReconciler(m manager.Manager) *serviceReconciler {
	return &serviceReconciler{client: m.GetClient()}
}

type serviceReconciler struct {
	client client.Client
}

func (r *serviceReconciler) GetClient() client.Client {
	return r.client
}

// Reconcile implements the loop that will update the ResourceBundleState CR
// whenever we get any updates from all the services we watch.
func (r *serviceReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Updating ResourceBundleState for Service: %+v\n", req)

	svc := &corev1.Service{}
	err := r.client.Get(context.TODO(), req.NamespacedName, svc)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("Service not found: %+v. Remove from CR if it is stored there.\n", req.NamespacedName)
			// Remove the Service's status from StatusList
			// This can happen if we get the DeletionTimeStamp event
			// after the Service has been deleted.
			DeleteFromAllCRs(r, req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get service: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}
	//err = r.updateCRs(rbStatusList, svc)
	err = UpdateCR(r, svc)
	if err != nil {
		// Requeue the update
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *serviceReconciler) UpdateStatus(cr *v1alpha1.ResourceBundleState, obj runtime.Object) (bool, error) {
	var found bool
	service, ok := obj.(*corev1.Service)
	if !ok {
		return found, pkgerrors.Errorf("Unknown resource %v", obj)
	}
	for i, rstatus := range cr.Status.ServiceStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == service.Name {
			service.Status.DeepCopyInto(&cr.Status.ServiceStatuses[i].Status)
			found = true
			break
		}
	}
	if !found {
		// Add it to CR
		svc := corev1.Service{
			TypeMeta:   service.TypeMeta,
			ObjectMeta: service.ObjectMeta,
			Status:     service.Status,
			Spec:       service.Spec,
		}
		svc.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		svc.Annotations = ClearLastApplied(svc.Annotations)
		cr.Status.ServiceStatuses = append(cr.Status.ServiceStatuses, svc)
	}
	return found, nil
}

func (r *serviceReconciler) DeleteObj(cr *v1alpha1.ResourceBundleState, name string) bool {
	var found bool
	length := len(cr.Status.ServiceStatuses)
	for i, rstatus := range cr.Status.ServiceStatuses {
		if rstatus.Name == name {
			found = true
			//Delete that status from the array
			cr.Status.ServiceStatuses[i] = cr.Status.ServiceStatuses[length-1]
			cr.Status.ServiceStatuses[length-1] = corev1.Service{}
			cr.Status.ServiceStatuses = cr.Status.ServiceStatuses[:length-1]
			break
		}
	}
	return found
}
