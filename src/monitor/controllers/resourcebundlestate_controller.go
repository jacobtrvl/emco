/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	k8spluginv1alpha1 "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	slog "sigs.k8s.io/controller-runtime/pkg/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/batch/v1"
	certsapi "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceBundleStateReconciler reconciles a ResourceBundleState object
type ResourceBundleStateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=k8splugin.io,resources=resourcebundlestates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8splugin.io,resources=resourcebundlestates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8splugin.io,resources=resourcebundlestates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ResourceBundleState object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ResourceBundleStateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = slog.FromContext(ctx)
	log.Println("Reconcile::", req)
	rbstate := &k8spluginv1alpha1.ResourceBundleState{}
	err := r.Get(context.TODO(), req.NamespacedName, rbstate)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("Object not found: %+v. Ignore as it must have been deleted.\n", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		log.Printf("Failed to get object: %+v\n", req.NamespacedName)
		return ctrl.Result{}, err
	}

	rbstate.Status.Ready = true
	err = r.updatePods(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding podstatuses: %v\n", err)
		return ctrl.Result{}, err
	}

	err = r.updateServices(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding servicestatuses: %v\n", err)
		return ctrl.Result{}, err
	}

	err = r.updateConfigMaps(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding configmapstatuses: %v\n", err)
		return ctrl.Result{}, err
	}

	err = r.updateDeployments(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding deploymentstatuses: %v\n", err)
		return ctrl.Result{}, err
	}

	err = r.updateDaemonSets(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding daemonSetstatuses: %v\n", err)
		return ctrl.Result{}, err
	}

	err = r.updateJobs(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding jobstatuses: %v\n", err)
		return ctrl.Result{}, err
	}

	err = r.updateStatefulSets(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding statefulSetstatuses: %v\n", err)
		return ctrl.Result{}, err
	}

	err = r.updateCsrs(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding csrStatuses: %v\n", err)
		return ctrl.Result{}, err
	}

	if len(rbstate.Status.ResourceStatuses) == 0 {
		rbstate.Status.ResourceStatuses = []k8spluginv1alpha1.ResourceStatus{}
	}

	err = r.updateDynResources(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding dynamic resources: %v\n", err)
		return ctrl.Result{}, err
	}

	err = CommitCR(r.Client, rbstate)
	if err != nil {
		log.Printf("failed to update rbstate: %v\n", err)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ResourceBundleStateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8spluginv1alpha1.ResourceBundleState{}).
		Complete(r)
}

func (r *ResourceBundleStateReconciler) updateServices(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the Services created as well
	serviceList := &corev1.ServiceList{}

	err := listResources(r.Client, rbstate.Namespace, selectors, serviceList)
	if err != nil {
		log.Printf("Failed to list services: %v", err)
		return err
	}

	rbstate.Status.ServiceStatuses = []corev1.Service{}

	for _, svc := range serviceList.Items {
		resStatus := corev1.Service{
			TypeMeta:   svc.TypeMeta,
			ObjectMeta: svc.ObjectMeta,
			Status:     svc.Status,
			Spec:       svc.Spec,
		}

		resStatus.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		resStatus.Annotations = ClearLastApplied(resStatus.Annotations)
		rbstate.Status.ServiceStatuses = append(rbstate.Status.ServiceStatuses, svc)
	}

	return nil
}

func (r *ResourceBundleStateReconciler) updatePods(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the pods tracked
	podList := &corev1.PodList{}
	err := listResources(r.Client, rbstate.Namespace, selectors, podList)
	if err != nil {
		log.Printf("Failed to list pods: %v", err)
		return err
	}

	rbstate.Status.PodStatuses = []corev1.Pod{}

	for _, pod := range podList.Items {
		resStatus := corev1.Pod{
			TypeMeta:   pod.TypeMeta,
			ObjectMeta: pod.ObjectMeta,
			Status:     pod.Status,
			Spec:       pod.Spec,
		}
		resStatus.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		resStatus.Annotations = ClearLastApplied(resStatus.Annotations)
		rbstate.Status.PodStatuses = append(rbstate.Status.PodStatuses, resStatus)
	}
	return nil
}

func (r *ResourceBundleStateReconciler) updateConfigMaps(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the ConfigMaps created as well
	configMapList := &corev1.ConfigMapList{}
	err := listResources(r.Client, rbstate.Namespace, selectors, configMapList)
	if err != nil {
		log.Printf("Failed to list configMaps: %v", err)
		return err
	}

	rbstate.Status.ConfigMapStatuses = []corev1.ConfigMap{}

	for _, cm := range configMapList.Items {
		resStatus := corev1.ConfigMap{
			TypeMeta:   cm.TypeMeta,
			ObjectMeta: cm.ObjectMeta,
		}
		resStatus.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		resStatus.Annotations = ClearLastApplied(resStatus.Annotations)
		rbstate.Status.ConfigMapStatuses = append(rbstate.Status.ConfigMapStatuses, resStatus)
	}

	return nil
}

func (r *ResourceBundleStateReconciler) updateDeployments(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the Deployments created as well
	deploymentList := &appsv1.DeploymentList{}
	err := listResources(r.Client, rbstate.Namespace, selectors, deploymentList)
	if err != nil {
		log.Printf("Failed to list deployments: %v", err)
		return err
	}

	rbstate.Status.DeploymentStatuses = []appsv1.Deployment{}

	for _, dep := range deploymentList.Items {
		resStatus := appsv1.Deployment{
			TypeMeta:   dep.TypeMeta,
			ObjectMeta: dep.ObjectMeta,
			Status:     dep.Status,
			Spec:       dep.Spec,
		}
		resStatus.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		resStatus.Annotations = ClearLastApplied(resStatus.Annotations)
		rbstate.Status.DeploymentStatuses = append(rbstate.Status.DeploymentStatuses, resStatus)
	}

	return nil
}

func (r *ResourceBundleStateReconciler) updateDaemonSets(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the DaemonSets created as well
	daemonSetList := &appsv1.DaemonSetList{}
	err := listResources(r.Client, rbstate.Namespace, selectors, daemonSetList)
	if err != nil {
		log.Printf("Failed to list DaemonSets: %v", err)
		return err
	}

	rbstate.Status.DaemonSetStatuses = []appsv1.DaemonSet{}

	for _, ds := range daemonSetList.Items {
		resStatus := appsv1.DaemonSet{
			TypeMeta:   ds.TypeMeta,
			ObjectMeta: ds.ObjectMeta,
			Spec:       ds.Spec,
			Status:     ds.Status,
		}
		resStatus.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		resStatus.Annotations = ClearLastApplied(resStatus.Annotations)
		rbstate.Status.DaemonSetStatuses = append(rbstate.Status.DaemonSetStatuses, resStatus)
	}

	return nil
}

func (r *ResourceBundleStateReconciler) updateJobs(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the Services created as well
	jobList := &v1.JobList{}
	err := listResources(r.Client, rbstate.Namespace, selectors, jobList)
	if err != nil {
		log.Printf("Failed to list jobs: %v", err)
		return err
	}

	rbstate.Status.JobStatuses = []v1.Job{}

	for _, job := range jobList.Items {
		resStatus := v1.Job{
			TypeMeta:   job.TypeMeta,
			ObjectMeta: job.ObjectMeta,
			Status:     job.Status,
			Spec:       job.Spec,
		}
		resStatus.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		resStatus.Annotations = ClearLastApplied(resStatus.Annotations)
		rbstate.Status.JobStatuses = append(rbstate.Status.JobStatuses, resStatus)
	}

	return nil
}

func (r *ResourceBundleStateReconciler) updateStatefulSets(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the StatefulSets created as well
	statefulSetList := &appsv1.StatefulSetList{}
	err := listResources(r.Client, rbstate.Namespace, selectors, statefulSetList)
	if err != nil {
		log.Printf("Failed to list statefulSets: %v", err)
		return err
	}

	rbstate.Status.StatefulSetStatuses = []appsv1.StatefulSet{}

	for _, sfs := range statefulSetList.Items {
		resStatus := appsv1.StatefulSet{
			TypeMeta:   sfs.TypeMeta,
			ObjectMeta: sfs.ObjectMeta,
			Status:     sfs.Status,
			Spec:       sfs.Spec,
		}
		resStatus.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		resStatus.Annotations = ClearLastApplied(resStatus.Annotations)
		rbstate.Status.StatefulSetStatuses = append(rbstate.Status.StatefulSetStatuses, resStatus)
	}

	return nil
}

func (r *ResourceBundleStateReconciler) updateCsrs(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the csrs tracked
	csrList := &certsapi.CertificateSigningRequestList{}
	err := listResources(r.Client, "", selectors, csrList)
	if err != nil {
		log.Printf("Failed to list csrs: %v", err)
		return err
	}

	rbstate.Status.CsrStatuses = []certsapi.CertificateSigningRequest{}

	for _, csr := range csrList.Items {
		resStatus := certsapi.CertificateSigningRequest{
			TypeMeta:   csr.TypeMeta,
			ObjectMeta: csr.ObjectMeta,
			Status:     csr.Status,
		}
		resStatus.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		resStatus.Annotations = ClearLastApplied(resStatus.Annotations)
		rbstate.Status.CsrStatuses = append(rbstate.Status.CsrStatuses, resStatus)
	}
	return nil
}
// updateDynResources updates non default resources
func (r *ResourceBundleStateReconciler) updateDynResources(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	for gvk, item := range GvkMap {
		if item.defaultRes {
			// Already handled
			continue
		}
		resourceList := &unstructured.UnstructuredList{}
		resourceList.SetGroupVersionKind(gvk)

		err := listResources(r.Client, rbstate.Namespace, selectors, resourceList)
		if err != nil {
			log.Printf("Failed to list resources: %v", err)
			return err
		}
		for _, res := range resourceList.Items {
			err = UpdateResourceStatus(r.Client, &res, res.GetName(), res.GetNamespace())
			if err != nil {
				log.Println("Error updating status for resource", gvk, res.GetName())
			}
		}
	}
	return nil
}
