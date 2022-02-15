package action

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"sigs.k8s.io/yaml"
)

// MetaData holds the object Name, Namespace and Annotations
type MetaData struct {
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type updateError struct {
	message, cluster string
	handle           interface{}
	err              error
}

// applyPatch reconciles a modified configuration with an original configuration
func applyPatch(patch []map[string]interface{}, original []byte) ([]byte, error) {
	patchData, err := json.MarshalIndent(patch, "", " ")
	if err != nil {
		log.Error("Failed to marshal the customization json patch",
			log.Fields{
				"Patch": patch,
				"Error": err.Error()})
		return nil, err
	}

	originalData, err := yaml.YAMLToJSON(original)
	if err != nil {
		log.Error("Failed to convert the existing resource yaml to json document",
			log.Fields{
				"Error": err.Error(),
			})
		return []byte{}, err
	}

	decodedPatch, err := jsonpatch.DecodePatch([]byte(patchData))
	if err != nil {
		log.Error("Failed to decode the customization json patch data",
			log.Fields{
				"Error": err.Error(),
			})
		return []byte{}, err
	}

	modifiedData, err := decodedPatch.Apply(originalData)
	if err != nil {
		log.Error("Failed to apply the customization json patch data",
			log.Fields{
				"Error": err.Error()})
		return []byte{}, err
	}

	modifiedPatch, err := yaml.JSONToYAML(modifiedData)
	if err != nil {
		log.Error("Failed to convert the updated json document to yaml",
			log.Fields{
				"Error": err.Error(),
			})
		return []byte{}, err
	}

	return modifiedPatch, nil
}

// isValidClusterToApplyByLabel checks if a given cluster falls under the given label and provider
func isValidClusterToApplyByLabel(provider, clusterName, clusterLabel, mode string) (bool, error) {
	clusters, err := cluster.NewClusterClient().GetClustersWithLabel(provider, clusterLabel)
	if err != nil {
		log.Error("Failed to get clusters by the provider and label",
			log.Fields{
				"Provider":                provider,
				"AutheticatingForCluster": clusterName,
				"ClusterLabel":            clusterLabel,
				"Mode":                    mode})
		return false, err
	}

	clusterName = strings.Split(clusterName, SEPARATOR)[1]
	for _, c := range clusters {
		if c == clusterName && mode == "allow" {
			return true, nil
		}
	}

	return false, nil
}

// isValidClusterToApplyByName checks if a given cluster under a provider matches with the cluster which is authenticated for
func isValidClusterToApplyByName(provider, authenticatedCluster, givenCluster, mode string) (bool, error) {
	clusters, err := cluster.NewClusterClient().GetClusters(provider)
	if err != nil {
		log.Error("Failed to get clusters by the provider",
			log.Fields{
				"Provider":                provider,
				"GivenCluster":            givenCluster,
				"AutheticatingForCluster": authenticatedCluster,
				"Mode":                    mode,
				"Error":                   err.Error()})
		return false, err
	}

	authenticatedCluster = strings.Split(authenticatedCluster, SEPARATOR)[1]
	for _, c := range clusters {
		if c.Metadata.Name == authenticatedCluster && authenticatedCluster == givenCluster && mode == "allow" {
			return true, nil
		}
	}

	return false, nil
}

// decodeString returns the bytes represented by the base64 string s
func decodeString(s string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Error("Failed to decode the base64 string",
			log.Fields{
				"Error": err.Error()})
		return []byte{}, err
	}

	return data, nil
}

// getResourceContent retrieves the resource template from the db
func (o *updateOptions) getResourceContent() error {
	resourceContent, err := module.NewResourceClient().GetResourceContent(o.resource.Metadata.Name, o.appMeta.Project,
		o.appMeta.CompositeApp, o.appMeta.Version, o.appMeta.DeploymentIntentGroup, o.intent)
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to get resource content",
				err:     err})
		return err
	}

	o.resourceContent = resourceContent

	return nil
}

// getAllCustomization returns all the customizations
func (o *updateOptions) getAllCustomization() ([]module.Customization, error) {
	customizations, err := module.NewCustomizationClient().GetAllCustomization(o.appMeta.Project,
		o.appMeta.CompositeApp, o.appMeta.Version, o.appMeta.DeploymentIntentGroup, o.intent, o.resource.Metadata.Name)
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to get customizations",
				err:     err})
		return []module.Customization{}, err
	}

	if len(customizations) == 0 {
		log.Warn("No customization is available for the resource",
			log.Fields{
				"Resource": o.resource.Metadata.Name})
	}

	return customizations, nil
}

// getCustomizationContent retrieves the customization template from the db
func (o *updateOptions) getCustomizationContent() error {
	customizationContent, err := module.NewCustomizationClient().GetCustomizationContent(o.customization.Metadata.Name, o.appMeta.Project,
		o.appMeta.CompositeApp, o.appMeta.Version, o.appMeta.DeploymentIntentGroup, o.intent, o.resource.Metadata.Name)
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to get customization content",
				err:     err})
		return err
	}

	o.customizationContent = customizationContent

	return nil
}

// getClusterNames returns a list of all clusters for a given app
func (o *updateOptions) getClusterNames() ([]string, error) {
	clusters, err := o.appContext.GetClusterNames(o.resource.Spec.AppName)
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to get cluster names",
				err:     err})
		return []string{}, err
	}

	return clusters, nil
}

// getClusterHandle returns the handle for a given app and cluster
func (o *updateOptions) getClusterHandle(cluster string) (interface{}, error) {
	handle, err := o.appContext.GetClusterHandle(o.resource.Spec.AppName, cluster)
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to get cluster handle",
				cluster: cluster,
				err:     err})
		return nil, err
	}

	return handle, nil
}

// addResource add the resource under the app and cluster
func (o *updateOptions) addResource(handle interface{}, resource, value string) error {
	_, err := o.appContext.AddResource(handle, resource, value)
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to add the resource",
				handle:  handle,
				err:     err})
		return err
	}

	return nil
}

// getResourceInstruction returns the resource instruction for a given instruction type
func (o *updateOptions) getResourceInstruction(cluster string) (interface{}, error) {
	resorder, err := o.appContext.GetResourceInstruction(o.resource.Spec.AppName, cluster, "order")
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to get resource instruction",
				cluster: cluster,
				err:     err})
		return nil, err
	}

	return resorder, nil
}

// addInstruction add instruction under the given handle and instruction type
func (o *updateOptions) addInstruction(handle, resorder interface{}, cluster, resource string) error {
	v := make(map[string][]string)
	json.Unmarshal([]byte(resorder.(string)), &v)
	v["resorder"] = append(v["resorder"], resource)
	data, _ := json.Marshal(v)
	if _, err := o.appContext.AddInstruction(handle, "resource", "order", string(data)); err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to add instruction",
				cluster: cluster,
				handle:  handle,
				err:     err})
		return err
	}

	return nil
}

// getResourceHandle returns the handle for the given app, cluster, and resource
func (o *updateOptions) getResourceHandle(cluster, resource string) (interface{}, error) {
	handle, err := o.appContext.GetResourceHandle(o.resource.Spec.AppName, cluster, resource)
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to get resource handle",
				cluster: cluster,
				err:     err})
		return nil, err
	}

	return handle, nil
}

// getValue returns the value for a given handle
func (o *updateOptions) getValue(handle interface{}) (interface{}, error) {
	val, err := o.appContext.GetValue(handle)
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to get handle value",
				handle:  handle,
				err:     err})
		return nil, err
	}

	log.Info("Manifest file for the resource",
		log.Fields{
			"Resource":      o.resource.Spec.ResourceGVK.Name,
			"Manifest-File": val.(string)})

	return val, nil
}

// updateResourceValue updates the resource value using the given handle
func (o *updateOptions) updateResourceValue(handle interface{}, value string) error {
	err := o.appContext.UpdateResourceValue(handle, value)
	if err != nil {
		o.logUpdateError(
			updateError{
				message: "Failed to update resource value",
				handle:  handle,
				err:     err})
		return err
	}

	log.Info("Resource updated in appContext",
		log.Fields{
			"AppName":      o.resource.Spec.AppName,
			"AppMeta":      o.appMeta,
			"Intent":       o.intent,
			"resourceName": o.resource.Metadata.Name})

	return nil
}

// logError writes the error details to the log
func logError(message, appContext, intent string, appMeta appcontext.CompositeAppMeta, err error) {
	log.Error(message,
		log.Fields{
			"AppContext": appContext,
			"AppMeta":    appMeta,
			"Intent":     intent,
			"Error":      err.Error()})
}

// logUpdateError writes the update errors to the log
func (o *updateOptions) logUpdateError(uError updateError) {
	fields := make(log.Fields)
	fields["AppMeta"] = o.appMeta
	if len(o.resource.Spec.AppName) > 0 {
		fields["AppName"] = o.resource.Spec.AppName
	}
	if len(uError.cluster) > 0 {
		fields["Clsuter"] = uError.cluster
	}
	if len(o.customization.Metadata.Name) > 0 {
		fields["Customization"] = o.customization.Metadata.Name
	}
	if uError.err != nil {
		fields["Error"] = uError.err.Error()
	}
	if uError.handle != nil {
		fields["Handle"] = uError.handle
	}
	fields["Intent"] = o.intent
	if len(o.resource.Spec.ResourceGVK.Kind) > 0 {
		fields["Kind"] = o.resource.Spec.ResourceGVK.Kind
	}
	if len(o.resource.Metadata.Name) > 0 {
		fields["Resource"] = o.resource.Metadata.Name
	}

	log.Error(uError.message, fields)

}
