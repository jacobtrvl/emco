package action

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	"k8s.io/apimachinery/pkg/runtime/schema"
	strategicpatch "k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/runtime"
)

// SEPARATOR used while creating resourceNames to store in etcd
const SEPARATOR = "+"

// updateOptions
type updateOptions struct {
	appContext           appcontext.AppContext
	appMeta              appcontext.CompositeAppMeta
	customization        module.Customization
	customizationContent module.CustomizationContent
	intent               string
	objectKind           string
	resource             module.Resource
	resourceContent      module.ResourceContent
}

// UpdateAppContext creates/updates the k8s resources in the given appContext and intent
func UpdateAppContext(intent, appContextID string) error {
	log.Info("Begin app context update",
		log.Fields{
			"AppContext": appContextID,
			"Intent":     intent})

	var appContext appcontext.AppContext
	if _, err := appContext.LoadAppContext(appContextID); err != nil {
		logError("failed to load appContext", appContextID, intent, appcontext.CompositeAppMeta{}, err)
		return err
	}

	appMeta, err := appContext.GetCompositeAppMeta()
	if err != nil {
		logError("failed to get compositeApp meta", appContextID, intent, appcontext.CompositeAppMeta{}, err)
		return err
	}

	resources, err := module.NewResourceClient().GetAllResources(
		appMeta.Project, appMeta.CompositeApp, appMeta.Version, appMeta.DeploymentIntentGroup, intent)
	if err != nil {
		logError("failed to get resources", appContextID, intent, appMeta, err)
		return err
	}

	for _, resource := range resources {
		o := updateOptions{
			appContext: appContext,
			appMeta:    appMeta,
			intent:     intent,
		}
		o.objectKind = strings.ToLower(resource.Spec.ResourceGVK.Kind)
		o.resource = resource
		if err := o.createOrUpdateResource(); err != nil {
			return err
		}
	}

	return nil
}

// createOrUpdateResource creates a new k8s object or updates the existing one
func (o *updateOptions) createOrUpdateResource() error {
	if err := o.getResourceContent(); err != nil {
		return err
	}

	customizations, err := o.getAllCustomization()
	if err != nil {
		return err
	}

	for _, customization := range customizations {
		o.customization = customization
		if o.objectKind == "configmap" ||
			o.objectKind == "secret" {
			// customization using files is supported only for ConfigMap/Secret
			if err = o.getCustomizationContent(); err != nil {
				return err
			}
		}

		if strings.ToLower(o.resource.Spec.NewObject) == "true" {
			if err = o.createNewResource(); err != nil {
				return err
			}
			continue
		}

		if err = o.updateExistingResource(); err != nil {
			return err
		}
	}

	return nil
}

// createNewResource creates a new k8s object
func (o *updateOptions) createNewResource() error {
	switch o.objectKind {
	case "configmap":
		if err := o.createConfigMap(); err != nil {
			return err
		}
	case "secret":
		if err := o.createSecret(); err != nil {
			return err
		}
	default:
		if err := o.createK8sResource(); err != nil {
			return err
		}
	}

	return nil
}

// createK8sResource creates a new k8s object
func (o *updateOptions) createK8sResource() error {
	if len(o.resourceContent.Content) == 0 {
		o.logUpdateError(
			updateError{
				message: "resources content is empty"})
		return errors.New("resources content is empty")
	}

	// decode the template value
	value, err := decodeString(o.resourceContent.Content)
	if err != nil {
		return err
	}

	if strings.ToLower(o.customization.Spec.PatchType) == "json" &&
		len(o.customization.Spec.PatchJSON) > 0 {
		// validate the JSON patch value before applying
		if err := o.validateJSONPatchValue(); err != nil {
			return err
		}

		modifiedPatch, err := applyJSONPatch(o.customization.Spec.PatchJSON, value)
		if err != nil {
			return err
		}
		value = modifiedPatch // use the merge patch to create the resource
	}

	if err = o.create(value); err != nil {
		return err
	}

	return nil
}

// create adds the resource under the app and cluster
// also add instruction under the given handle and instruction type
func (o *updateOptions) create(data []byte) error {
	clusters, err := o.getClusterNames()
	if err != nil {
		return err
	}

	clusterName := o.customization.Spec.ClusterInfo.ClusterName
	clusterSpecific := strings.ToLower(o.customization.Spec.ClusterSpecific)
	label := o.customization.Spec.ClusterInfo.ClusterLabel
	mode := strings.ToLower(o.customization.Spec.ClusterInfo.Mode)
	provider := o.customization.Spec.ClusterInfo.ClusterProvider
	scope := strings.ToLower(o.customization.Spec.ClusterInfo.Scope)

	for _, cluster := range clusters {
		if clusterSpecific == "true" && scope == "label" {
			isValid, err := isValidClusterToApplyByLabel(provider, cluster, label, mode)
			if err != nil {
				return err
			}
			if !isValid {
				continue
			}
		}

		if clusterSpecific == "true" && scope == "name" {
			isValid, err := isValidClusterToApplyByName(provider, cluster, clusterName, mode)
			if err != nil {
				return err
			}
			if !isValid {
				continue
			}
		}

		handle, err := o.getClusterHandle(cluster)
		if err != nil {
			return err
		}

		resource := o.resource.Spec.ResourceGVK.Name + SEPARATOR + o.resource.Spec.ResourceGVK.Kind

		if err = o.addResource(handle, resource, string(data)); err != nil {
			return err
		}

		resorder, err := o.getResourceInstruction(cluster)
		if err != nil {
			return err
		}

		if err = o.addInstruction(handle, resorder, cluster, resource); err != nil {
			return err
		}

	}

	return nil
}

// updateExistingResource update the existing k8s object
func (o *updateOptions) updateExistingResource() error {

	if len(o.customization.Spec.PatchType) == 0 {
		return errors.New("patch type not defined") // check this message
	}

	var (
		modifiedPatch []byte
		err           error
	)

	clusters, err := o.getClusterNames()
	if err != nil {
		return err
	}

	clusterName := o.customization.Spec.ClusterInfo.ClusterName
	clusterSpecific := strings.ToLower(o.customization.Spec.ClusterSpecific)
	label := o.customization.Spec.ClusterInfo.ClusterLabel
	mode := strings.ToLower(o.customization.Spec.ClusterInfo.Mode)
	provider := o.customization.Spec.ClusterInfo.ClusterProvider
	scope := strings.ToLower(o.customization.Spec.ClusterInfo.Scope)

	for _, cluster := range clusters {
		if clusterSpecific == "true" && scope == "label" {
			isValid, err := isValidClusterToApplyByLabel(provider, cluster, label, mode)
			if err != nil {
				return err
			}
			if !isValid {
				continue
			}
		}

		if clusterSpecific == "true" && scope == "name" {
			isValid, err := isValidClusterToApplyByName(provider, cluster, clusterName, mode)
			if err != nil {
				return err
			}
			if !isValid {
				continue
			}
		}

		handle, err := o.getResourceHandle(cluster, strings.Join([]string{o.resource.Spec.ResourceGVK.Name,
			o.resource.Spec.ResourceGVK.Kind}, SEPARATOR))
		if err != nil {
			continue
		}

		val, err := o.getValue(handle)
		if err != nil {
			continue
		}

		switch strings.ToLower(o.customization.Spec.PatchType) {
		case "json":
			// make sure we have a valid JSON patch to update the resource
			if len(o.customization.Spec.PatchJSON) == 0 {
				o.logUpdateError(
					updateError{
						message: "invalid json patch"})
				return errors.New("invalid json patch")
			}

			// validate the JSON patch value before applying
			if err := o.validateJSONPatchValue(); err != nil {
				return err
			}

			modifiedPatch, err = applyJSONPatch(o.customization.Spec.PatchJSON, []byte(val.(string)))
			if err != nil {
				return err
			}

		case "merge":
			// make sure we have the cutomization files
			if len(o.customizationContent.Content) == 0 {
				return errors.New("no patch file")
			}

			original := []byte(val.(string))

			for _, c := range o.customizationContent.Content {
				data, err := decodeString(c.Content)
				if err != nil {
					return err
				}

				patch, err := yaml.YAMLToJSON(data)
				if err != nil {
					return err
				}

				ds, err := getResourceStructFromGVK(o.resource.Spec.ResourceGVK.APIVersion, o.resource.Spec.ResourceGVK.Kind)
				if err != nil {
					return err
				}

				modifiedPatch, err = strategicpatch.StrategicMergePatch(original, patch, ds)
				if err != nil {
					return err
				}

				original = modifiedPatch
			}
		}

		if err = o.updateResourceValue(handle, string(modifiedPatch)); err != nil {
			return err
		}
	}

	return nil
}

func getResourceStructFromGVK(apiVersion, kind string) (runtime.Object, error) {
	resourceGVK := schema.GroupVersionKind{Kind: kind}
	if gv, err := schema.ParseGroupVersion(apiVersion); err == nil {
		resourceGVK = schema.GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: kind}
	}

	obj, err := runtime.NewScheme().New(resourceGVK)
	if err != nil {
		log.Error("Failed to get the resource struct type using the GVK details",
			log.Fields{
				"APIVersion": apiVersion,
				"Kind":       kind,
				"Error":      err.Error()})

	}

	return obj, err
}
