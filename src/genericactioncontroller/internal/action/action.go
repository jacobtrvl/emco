package action

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
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
			return err // hwo to test this case ? is this a valid use case ?
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
		modifiedPatch, err := applyPatch(o.customization.Spec.PatchJSON, value)
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
	clusterSpecific := strings.ToLower(o.customization.Spec.ClusterSpecific)
	scope := strings.ToLower(o.customization.Spec.ClusterInfo.Scope)
	provider := o.customization.Spec.ClusterInfo.ClusterProvider
	clusterName := o.customization.Spec.ClusterInfo.ClusterName
	label := o.customization.Spec.ClusterInfo.ClusterLabel
	mode := strings.ToLower(o.customization.Spec.ClusterInfo.Mode)

	clusters, err := o.getClusterNames()
	if err != nil {
		return err
	}

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
	clusterSpecific := strings.ToLower(o.customization.Spec.ClusterSpecific)
	scope := strings.ToLower(o.customization.Spec.ClusterInfo.Scope)
	provider := o.customization.Spec.ClusterInfo.ClusterProvider
	clusterName := o.customization.Spec.ClusterInfo.ClusterName
	label := o.customization.Spec.ClusterInfo.ClusterLabel
	mode := strings.ToLower(o.customization.Spec.ClusterInfo.Mode)

	clusters, err := o.getClusterNames()
	if err != nil {
		return err
	}

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

		modifiedPatch, err := applyPatch(o.customization.Spec.PatchJSON, []byte(val.(string)))
		if err != nil {
			return err
		}

		if err = o.updateResourceValue(handle, string(modifiedPatch)); err != nil {
			return err
		}
	}

	return nil
}
