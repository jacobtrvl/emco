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

// SecretResource consists of ApiVersion, Kind, MetaData, type and Data map
type SecretResource struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	MetaData   MetaData          `yaml:"metadata"`
	Type       string            `yaml:"type"`
	Data       map[string]string `yaml:"data"`
}

// createSecret creates the ConfigMap struct based on the input
func createSecret(file, name, intent string, customizationFiles module.SpecFileContent, appMeta appcontext.CompositeAppMeta, resource module.Resource, customization module.Customization, appContext appcontext.AppContext) error {
	cm, err := setSecretBase(file, name)
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

func setSecretBase(data, name string) (SecretResource, error) {
	// Check if the resource has a file associated with it
	if len(data) > 0 {
		value, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			log.Error("Failed to encode customization data.",
				log.Fields{
					"Error": err.Error()})
			return SecretResource{}, err
		}

		if len(value) > 0 {
			s := SecretResource{}
			fmt.Println(string(value))
			// If the configmap teplate is available then it should be YAML
			err = yamlV2.Unmarshal(value, &s)
			if err != nil {
				log.Error("Failed to unmarshal customization data.",
					log.Fields{
						"Error": err.Error()})
				return SecretResource{}, err
			}

			if s.APIVersion != "" &&
				strings.ToLower(s.Kind) == "secret" &&
				len(s.Data) != 0 &&
				(s.MetaData != MetaData{}) {
				return s, nil
			}

			return s, errors.New("Invalid secret template") // verify the logic of this.
		}
	}

	// We don't have the templae for ConfigMap. Only customization is available. Create the base config map
	s := SecretResource{
		APIVersion: "v1",
		Kind:       "ConfigMap",
		MetaData: MetaData{
			Name: name,
		},
		Data: map[string]string{},
	}
	return s, nil
}
