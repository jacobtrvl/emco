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
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"sigs.k8s.io/yaml"
)

// generateModifiedYamlFileForExistingResources takes in the patchData and the existing resource's yaml file and returns the modified yaml file for the resource
func generateModifiedYamlFileForExistingResources(p []map[string]interface{}, existingResData []byte, resName string) ([]byte, error) {
	patchData, err := json.MarshalIndent(p, "", " ")
	if err != nil {
		log.Error("Failed to marshal indent the customization json patch data.",
			log.Fields{
				"Error": err.Error()})
		return nil, err
	}

	// Decode patch file
	decodedPatch, err := jsonpatch.DecodePatch([]byte(patchData))
	if err != nil {
		log.Error("Failed to decode the customization json patch data.",
			log.Fields{
				"Error": err.Error(),
			})
		return []byte{}, err
	}

	existingResDataJSON, err := yaml.YAMLToJSON(existingResData)
	if err != nil {
		log.Error("Failed to convert the existing resource data to json.",
			log.Fields{
				"Error": err.Error(),
			})
		return []byte{}, err
	}

	// Apply the patch
	modified, err := decodedPatch.Apply(existingResDataJSON)
	if err != nil {
		log.Error("Failed to apply the customization json patch data.",
			log.Fields{
				"Error": err.Error()})
		return []byte{}, err
	}

	modifiedYaml, err := yaml.JSONToYAML(modified)
	if err != nil {
		log.Error("Failed to convert the updated resource data to yaml.",
			log.Fields{
				"Error": err.Error(),
			})
		return []byte{}, err
	}

	return modifiedYaml, nil
}

// isValidClusterToApplyByLabel checks if the cluster being authenticated for(acName) falls under the given label(cLabel) and provider(cProvider)
func isValidClusterToApplyByLabel(cProvider, cName, cLabel, cMode string) (bool, error) {
	clusters, err := cluster.NewClusterClient().GetClustersWithLabel(cProvider, cLabel)
	if err != nil {
		log.Error("Failed to get clusters by provider and label.",
			log.Fields{
				"Provider":     cProvider,
				"Cluster":      cName,
				"ClusterLabel": cLabel,
				"Mode":         cMode})
		return false, err
	}

	cName = strings.Split(cName, SEPARATOR)[1]
	for _, c := range clusters {
		if c == cName && cMode == "allow" {
			return true, nil
		}
	}
	return false, nil
}

// isValidClusterToApplyByName checks if a given cluster(gcName) under a provider(cProvider) matches with the cluster which is authenticated for(acName).
func isValidClusterToApplyByName(cProvider, cName, gcName, cMode string) (bool, error) {
	clusters, err := cluster.NewClusterClient().GetClusters(cProvider)
	if err != nil {
		log.Error("Failed to get clusters by provider.",
			log.Fields{
				"Provider":                cProvider,
				"GivenCluster":            gcName,
				"AutheticatingForCluster": cName,
				"Mode":                    cMode,
				"Error":                   err.Error()})
		return false, err
	}
	cName = strings.Split(cName, SEPARATOR)[1]
	for _, c := range clusters {
		if c.Metadata.Name == cName && cName == gcName && cMode == "allow" {
			return true, nil
		}
	}
	return false, nil
}

// handleCustomization adds the specified customization data to the ConfigMap/ Secret
func handleCustomization(data map[string]string, customization module.SpecFileContent) (map[string]string, error) {
	// The number of customization file contents and filenames should be equal and in the same order.
	for i, f := range customization.FileNames {
		value, err := decodeString(customization.FileContents[i])
		if err != nil {
			return map[string]string{}, err
		}
		data[f] = string(value)
	}
	return data, nil
}

func decodeString(s string) ([]byte, error) {
	value, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Error("Failed to decode the resource template data",
			log.Fields{})
		return []byte{}, err
	}

	return value, nil
}
