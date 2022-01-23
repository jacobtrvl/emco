package action

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	yamlV2 "gopkg.in/yaml.v2"
)

// ConfigMapResource consists of ApiVersion, Kind, MetaData and Data map
type ConfigMapResource struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	MetaData   MetaData          `yaml:"metadata"`
	Data       map[string]string `yaml:"data"`
}

// MetaDataStr consists of Name and Namespace. Namespace is optional
type MetaData struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace,omitempty"`
}

// createConfigMap creates the ConfigMap struct based on the input
func createConfigMap(file, name, intent string, customizationFiles module.SpecFileContent, appMeta appcontext.CompositeAppMeta, resource module.Resource, customization module.Customization, appContext appcontext.AppContext) error {
	cm, err := setConfigMapBase(file, name)
	if err != nil {
		return err
	}
	data, err := handleCustomization(cm.Data, customizationFiles)
	if err != nil {
		return err
	}

	cm.Data = data

	value, err := yamlV2.Marshal(&cm)
	if err != nil {
		log.Error("error",
			log.Fields{
				"Error ": err.Error()})
		return err
	}

	fmt.Println(string(value))

	err = createResource(appMeta, resource, customization, appContext, intent, value)
	if err != nil {
		return err
	}

	return nil
}

func setConfigMapBase(data, name string) (ConfigMapResource, error) {
	// Check if the resource has a file associated with it
	if len(data) > 0 {
		value, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			log.Error("Failed to encode customization data.",
				log.Fields{
					"Error": err.Error()})
			return ConfigMapResource{}, err
		}

		if len(value) > 0 {
			cm := ConfigMapResource{}
			fmt.Println(string(value))
			// If the configmap teplate is available then it should be YAML
			err = yamlV2.Unmarshal(value, &cm)
			if err != nil {
				log.Error("Failed to unmarshal customization data.",
					log.Fields{
						"Error": err.Error()})
				return ConfigMapResource{}, err
			}

			if cm.APIVersion != "" &&
				strings.ToLower(cm.Kind) == "configmap" &&
				len(cm.Data) != 0 &&
				(cm.MetaData != MetaData{}) {
				return cm, nil
			}

			return cm, errors.New("Invalid configmap template") // verify the logic of this.
		}
	}

	// We don't have the templae for ConfigMap. Only customization is available. Create the base config map
	cm := ConfigMapResource{
		APIVersion: "v1",
		Kind:       "ConfigMap",
		MetaData: MetaData{
			Name: name,
		},
		Data: map[string]string{},
	}
	return cm, nil
}
