// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package controllers

import (
	"context"
	"fmt"
	"log"
	"sync"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	slog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	k8spluginv1alpha1 "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
)

// Resources with special handling in Monitor
var defaultResourcesMap = map[schema.GroupVersionKind]bool{
	{Version: "v1", Kind: "ConfigMap"}:                                               true,
	{Group: "apps", Version: "v1", Kind: "DaemonSet"}:                                true,
	{Group: "apps", Version: "v1", Kind: "Deployment"}:                               true,
	{Group: "batch", Version: "v1", Kind: "Job"}:                                     true,
	{Version: "v1", Kind: "Service"}:                                                 true,
	{Version: "v1", Kind: "Pod"}:                                                     true,
	{Group: "apps", Version: "v1", Kind: "StatefulSet"}:                              true,
	{Group: "certificates.k8s.io", Version: "v1", Kind: "CertificateSigningRequest"}: true,
}

// Resources to not be monitored
var ignoredResourcesMap = map[schema.GroupKind]bool{
	{Group: k8spluginv1alpha1.GroupVersion.Group, Kind: "ResourceBundleState"}: true,
}

// ResourceBundleStateReconciler reconciles a ResourceBundleState object
type ControllerListReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	gvk    *schema.GroupVersionKind
	mutex  *sync.Mutex
}

func isDefaultResource(resourceGVK schema.GroupVersionKind) bool {
	_, ok := defaultResourcesMap[resourceGVK]
	return ok
}

func isIgnoredResource(resourceGK schema.GroupKind) bool {
	_, ok := ignoredResourcesMap[resourceGK]
	return ok
}

//+kubebuilder:rbac:groups=k8splugin.io,resources=resourcebundlestates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8splugin.io,resources=resourcebundlestates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8splugin.io,resources=resourcebundlestates/finalizers,verbs=update

func (r *ControllerListReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = slog.FromContext(ctx)

	log.Println("Reconcile", req.Name, req.Namespace, r.gvk)
	// Note: Create an unstructued type for the Client to use and set the Kind
	// so that it can GET/UPDATE/DELETE etc
	resource := &unstructured.Unstructured{}
	resource.SetGroupVersionKind(*r.gvk)

	err := r.Get(ctx, req.NamespacedName, resource)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}

		dc := DeleteStatusClient{
			c:          r.Client,
			name:       req.Name,
			namespace:  req.Namespace,
			gvk:        *r.gvk,
			defaultRes: isDefaultResource(*r.gvk),
		}

		r.mutex.Lock()
		defer r.mutex.Unlock()

		err = dc.Delete()
		return ctrl.Result{}, err
	}

	// If resource not a default resource for the controller
	// Add status to ResourceStatues array
	uc := UpdateStatusClient{
		c:          r.Client,
		item:       resource,
		name:       req.NamespacedName.Name,
		namespace:  req.NamespacedName.Namespace,
		gvk:        *r.gvk,
		defaultRes: isDefaultResource(*r.gvk),
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	err = uc.Update()
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControllerListReconciler) SetupWithManager(mgr ctrl.Manager) error {

	resource := &unstructured.Unstructured{}
	resource.SetGroupVersionKind(*r.gvk)

	return ctrl.NewControllerManagedBy(mgr).
		Named(r.gvk.Kind+"-emco-monitor").
		For(resource, builder.WithPredicates(predicate.NewPredicateFuncs(func(object client.Object) bool {
			labels := object.GetLabels()
			_, ok := labels["emco/deployment-id"]
			return ok
		}))).
		Complete(r)
}

func SetupControllerForType(mgr ctrl.Manager, resourceGVK *schema.GroupVersionKind, mutex *sync.Mutex) error {
	r := &ControllerListReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		gvk:    resourceGVK,
		mutex:  mutex,
	}

	if err := r.SetupWithManager(mgr); err != nil {
		return fmt.Errorf("failed to create controller for resource %+v: %q", resourceGVK, err)
	}

	return nil
}

func SetupControllers(mgr ctrl.Manager, controllersMutex *sync.Mutex) error {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		return fmt.Errorf("failed to create discovery client :%q", err)
	}

	resourcesGVK, err := GetServerResources(discoveryClient)
	if err != nil {
		return fmt.Errorf("failed to get server resources :%q", err)
	}

	for _, resourceGVK := range resourcesGVK {
		if isIgnoredResource(resourceGVK.GroupKind()) {
			continue
		}

		log.Printf("Adding controller for %+v", resourceGVK)
		if err = SetupControllerForType(mgr, resourceGVK, controllersMutex); err != nil {
			return fmt.Errorf("failed to add controller for %+v: %q", resourceGVK, err)
		}
	}

	return nil
}
