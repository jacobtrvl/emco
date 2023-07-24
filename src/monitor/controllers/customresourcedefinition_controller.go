// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package controllers

import (
	"context"
	"sync"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	slog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// CustomResourceDefinitionReconciler reconciles a CustomResourceDefinition object
type CustomResourceDefinitionReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	ResourcesMap map[schema.GroupVersionKind]bool
	mgr          ctrl.Manager
	mutex        *sync.Mutex
}

//+kubebuilder:rbac:groups=io,resources=customresourcedefinitions,verbs=get;list;watch

func (r *CustomResourceDefinitionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = slog.FromContext(ctx)

	crd := &apiextensionsv1.CustomResourceDefinition{}
	if err := r.Get(context.TODO(), req.NamespacedName, crd); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	for _, version := range crd.Spec.Versions {
		gvk := schema.GroupVersionKind{
			Group:   crd.Spec.Group,
			Version: version.Name,
			Kind:    crd.Spec.Names.Kind,
		}

		if isIgnoredResource(gvk.GroupKind()) {
			continue
		}

		if _, exist := r.ResourcesMap[gvk]; exist {
			// Resource already exists
			continue
		}

		if err := SetupControllerForType(r.mgr, &gvk, r.mutex); err != nil {
			return ctrl.Result{}, err
		}

		r.ResourcesMap[gvk] = true
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomResourceDefinitionReconciler) SetupWithManager(mgr ctrl.Manager, controllersMutex *sync.Mutex) error {
	r.mgr = mgr
	r.mutex = controllersMutex
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiextensionsv1.CustomResourceDefinition{}).WithEventFilter(pred).
		Complete(r)
}
