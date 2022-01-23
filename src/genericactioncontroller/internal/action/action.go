package action

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"encoding/json"
	"strings"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

// SEPARATOR used while creating resourceNames to store in etcd
const SEPARATOR = "+"

// UpdateAppContext is the method which calls the backend logic of this controller.
func UpdateAppContext(intent, appContextID string) error {
	log.Info("Begin AppContext update",
		log.Fields{
			"Intent":     intent,
			"AppContext": appContextID})

	var appContext appcontext.AppContext
	_, err := appContext.LoadAppContext(appContextID)
	if err != nil {
		log.Error("Failed to load AppContext",
			log.Fields{
				"Intent":     intent,
				"AppContext": appContextID,
				"Error":      err.Error()})
		return err
	}

	appMeta, err := appContext.GetCompositeAppMeta()
	if err != nil {
		log.Error("Failed to get CompositeAppMeta",
			log.Fields{
				"Intent":     intent,
				"AppContext": appContextID,
				"Error":      err.Error()})
		return err
	}

	prj := appMeta.Project
	ca := appMeta.CompositeApp
	ver := appMeta.Version
	diGroup := appMeta.DeploymentIntentGroup

	// get all the resources under this intent
	resources, err := module.NewResourceClient().GetAllResources(prj, ca, ver, diGroup, intent)
	if err != nil {
		log.Error("Failed to get Resources",
			log.Fields{
				"AppMeta": appMeta,
				"Intent":  intent,
				"Error":   err.Error()})
		return err
	}

	// if resource is either configMap or secret, it must have customization files.
	// go through each customization, find customization content, get the patchfiles,
	// generate the modified contentfiles and update context for each of the valid cluster
	for _, resource := range resources {
		objectKind := strings.ToLower(resource.Spec.ResourceGVK.Kind)
		newObject := strings.ToLower(resource.Spec.NewObject)
		resName := resource.Metadata.Name

		customizations, err := module.NewCustomizationClient().GetAllCustomization(prj, ca, ver, diGroup, intent, resName)
		if err != nil {
			log.Error("Failed to get Customizations",
				log.Fields{
					"AppMeta":  appMeta,
					"Intent":   intent,
					"Resource": resName,
					"Error":    err.Error()})
			return err
		}

		resourceFileContent, err := module.NewResourceClient().GetResourceContent(resName, prj, ca, ver, diGroup, intent)
		if err != nil {
			log.Error("Failed to get ResourceContent",
				log.Fields{
					"AppMeta":  appMeta,
					"Intent":   intent,
					"Resource": resName,
					"Error":    err.Error(),
				})
			return err
		}

		for _, customization := range customizations {
			patchType := strings.ToLower(customization.Spec.PatchType)

			// if object kind is neither configMap nor secret, create the resource with no customization
			if objectKind != "configmap" &&
				objectKind != "secret" &&
				newObject == "true" {
				// decode the template value
				value, err := decodeString(resourceFileContent.FileContent)
				if err != nil {
					return err
				}

				err = createResource(appMeta, resource, customization, appContext, intent, value)
				if err != nil {
					return err
				}

				continue
			}

			// Revisit this
			if newObject == "false" &&
				patchType == "json" {
				err := updateExistingResource(appMeta, intent, resource, appContext, customization)
				if err != nil {
					return err
				}
				continue
			}

			customizationContent, err := module.NewCustomizationClient().GetCustomizationContent(customization.Metadata.Name, prj, ca, ver, diGroup, intent, resName)
			if err != nil {
				log.Error("Failed to get CustomizationContent",
					log.Fields{
						"AppMeta":       appMeta,
						"Customization": customization.Metadata.Name,
						"Intent":        intent,
						"Resource":      resName,
						"Error":         err.Error()})
				return err
			}

			if len(resourceFileContent.FileContent) == 0 &&
				len(customizationContent.FileNames) == 0 &&
				len(customizationContent.FileContents) == 0 {
				log.Error("Customization files or contents is empty",
					log.Fields{
						"CustomizationFileCount":    len(customizationContent.FileNames),
						"CustomizationContentCount": len(customizationContent.FileContents),
						"ResourceFileContent":       resourceFileContent.FileContent})
				return pkgerrors.New("Customization files or contents is empty")
			}

			if objectKind == "configmap" {
				err := createConfigMap(resourceFileContent.FileContent, resource.Spec.ResourceGVK.Name, intent, customizationContent, appMeta, resource, customization, appContext)
				if err != nil {
					return err
				}
			}

			if objectKind == "secret" {
				err := createSecret(resourceFileContent.FileContent, resource.Spec.ResourceGVK.Name, intent, customizationContent, appMeta, resource, customization, appContext)
				if err != nil {
					return err
				}
			}

		}
	}
	return nil
}

// updateContextForExistingResource updates the context for an existing resource
func updateExistingResource(appMeta appcontext.CompositeAppMeta, intent string, resource module.Resource, appContext appcontext.AppContext, customization module.Customization) error {
	app := resource.Spec.AppName
	cSpecific := strings.ToLower(customization.Spec.ClusterSpecific)
	cScope := strings.ToLower(customization.Spec.ClusterInfo.Scope)
	cProvider := customization.Spec.ClusterInfo.ClusterProvider
	cName := customization.Spec.ClusterInfo.ClusterName
	cLabel := customization.Spec.ClusterInfo.ClusterLabel
	cMode := strings.ToLower(customization.Spec.ClusterInfo.Mode)

	clusters, err := appContext.GetClusterNames(app)
	if err != nil {
		log.Error("Error GetClusterNames",
			log.Fields{
				"App":      app,
				"AppMeta":  appMeta,
				"Intent":   intent,
				"Resource": resource.Metadata.Name,
				"Error":    err.Error()})
		return err
	}

	for _, cluster := range clusters {
		if cSpecific == "true" && cScope == "label" {
			isValid, err := isValidClusterToApplyByLabel(cProvider, cluster, cLabel, cMode)
			if err != nil {
				return err
			}
			if !isValid {
				continue
			}
		}

		if cSpecific == "true" && cScope == "name" {
			isValid, err := isValidClusterToApplyByName(cProvider, cluster, cName, cMode)
			if err != nil {
				return err
			}
			if !isValid {
				continue
			}
		}

		rName := strings.Join([]string{resource.Spec.ResourceGVK.Name, resource.Spec.ResourceGVK.Kind}, SEPARATOR)
		handle, err := appContext.GetResourceHandle(resource.Spec.AppName, cluster, rName)
		if err != nil {
			log.Error("Failed to get resource handle.",
				log.Fields{
					"App":      resource.Spec.AppName,
					"AppMeta":  appMeta,
					"Intent":   intent,
					"Resource": resource.Metadata.Name,
					"Kind":     resource.Spec.ResourceGVK.Kind,
					"Error":    err.Error()})
			continue // or return
		}

		val, err := appContext.GetValue(handle)
		if err != nil {
			log.Error("Failed to get value.",
				log.Fields{
					"ResourceHandle": handle,
					"Error":          err.Error(),
				})
			continue // or return
		}

		log.Info("Manifest file for the resource",
			log.Fields{
				"Resource":      resource.Spec.ResourceGVK.Name,
				"Manifest-File": val.(string)})

		dataBytes, err := generateModifiedYamlFileForExistingResources(customization.Spec.PatchJSON, []byte(val.(string)), resource.Metadata.Name)
		if err != nil {
			return err
		}

		log.Info("Data bytes",
			log.Fields{
				"DataBytes": string(dataBytes)})

		err = appContext.UpdateResourceValue(handle, string(dataBytes))
		if err != nil {
			log.Error("Error UpdateResourceValue",
				log.Fields{
					"AppName":  app,
					"AppMeta":  appMeta,
					"Intent":   intent,
					"Resource": resource.Metadata.Name,
					"Error":    err.Error()})
			return err
		}
		log.Info("Resource updated in AppContext",
			log.Fields{
				"App":          app,
				"AppMeta":      appMeta,
				"Intent":       intent,
				"resourceName": resource.Metadata.Name})
	}
	return nil
}

func createResource(appMeta appcontext.CompositeAppMeta, resource module.Resource, customization module.Customization, appContext appcontext.AppContext, intent string, data []byte) error {
	cSpecific := strings.ToLower(customization.Spec.ClusterSpecific)
	cScope := strings.ToLower(customization.Spec.ClusterInfo.Scope)
	cProvider := customization.Spec.ClusterInfo.ClusterProvider
	cName := customization.Spec.ClusterInfo.ClusterName
	cLabel := customization.Spec.ClusterInfo.ClusterLabel
	cMode := strings.ToLower(customization.Spec.ClusterInfo.Mode)
	app := resource.Spec.AppName

	// Get the clusters assocaited with the app name
	clusters, err := appContext.GetClusterNames(app)
	if err != nil {
		log.Error("Failed to get cluster names.",
			log.Fields{
				"App":      app,
				"AppMeta":  appMeta,
				"Intent":   intent,
				"Resource": resource.Metadata.Name,
				"Error":    err.Error()})
		return err
	}

	for _, cluster := range clusters {
		if cSpecific == "true" && cScope == "label" {
			isValid, err := isValidClusterToApplyByLabel(cProvider, cluster, cLabel, cMode)
			if err != nil {
				return err
			}
			if !isValid {
				continue
			}
		}

		if cSpecific == "true" && cScope == "name" {
			isValid, err := isValidClusterToApplyByName(cProvider, cluster, cName, cMode)
			if err != nil {
				return err
			}
			if !isValid {
				continue
			}
		}

		handle, err := appContext.GetClusterHandle(app, cluster)
		if err != nil {
			log.Error("Failed to get cluster handle.",
				log.Fields{
					"App":      app,
					"AppMeta":  appMeta,
					"Intent":   intent,
					"Resource": resource.Metadata.Name,
					"Error":    err.Error()})
			return err
		}

		name := resource.Spec.ResourceGVK.Name + SEPARATOR + resource.Spec.ResourceGVK.Kind
		_, err = appContext.AddResource(handle, name, string(data))
		if err != nil {
			log.Error("Failed to add resource.",
				log.Fields{
					"App":          app,
					"AppMeta":      appMeta,
					"Intent":       intent,
					"ResourceName": resource.Metadata.Name,
					"Error":        err.Error()})
			return err
		}

		// update the resource order
		resorder, err := appContext.GetResourceInstruction(app, cluster, "order")
		if err != nil {
			log.Error("Failed to get resource instruction.",
				log.Fields{
					"App":     app,
					"Cluster": cluster,
					"Error":   err.Error(),
				})
			return err
		}

		aov := make(map[string][]string)
		json.Unmarshal([]byte(resorder.(string)), &aov)
		aov["resorder"] = append(aov["resorder"], name)
		jresord, _ := json.Marshal(aov)

		_, err = appContext.AddInstruction(handle, "resource", "order", string(jresord))
		if err != nil {
			log.Error("Failed to add instruction.",
				log.Fields{
					"App":     app,
					"Cluster": cluster,
					"Error":   err.Error(),
				})
			return err
		}
	}
	return nil
}
