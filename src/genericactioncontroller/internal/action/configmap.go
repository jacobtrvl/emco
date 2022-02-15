package action

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	yamlV2 "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/validation"
)

// ConfigMap holds the configuration data
type ConfigMap struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	MetaData   MetaData          `yaml:"metadata"`
	Data       map[string]string `yaml:"data,omitempty"`
}

// createConfigMap create the ConfigMap configurations based on the JSON  patch,
// content in the template file, and the customization file, if any
func (o *updateOptions) createConfigMap() error {
	// create a new ConfigMap object based on the template file
	configMap, err := newConfigMap(o.resourceContent.Content, o.resource.Spec.ResourceGVK.Name)
	if err != nil {
		return err
	}

	if len(o.customizationContent.Content) > 0 {
		// apply the customization data to the ConfigMap
		if err := handleConfigMapCustomization(configMap, o.customizationContent.Content); err != nil {
			return err
		}
	}

	value, err := yamlV2.Marshal(configMap)
	if err != nil {
		log.Error("Failed to serialize the configMap object into a yaml document",
			log.Fields{
				"ConfigMap": configMap,
				"Error ":    err.Error()})
		return err
	}

	if strings.ToLower(o.customization.Spec.PatchType) == "json" &&
		len(o.customization.Spec.PatchJSON) > 0 {
		// apply the JSON patch associated with the ConfigMap customization
		modifiedPatch, err := applyPatch(o.customization.Spec.PatchJSON, value)
		if err != nil {
			return err
		}
		value = modifiedPatch
	}

	// create the ConfigMap
	if err = o.create(value); err != nil {
		return err
	}

	return nil
}

// newConfigMap creates a new ConfigMap object based on the template file
func newConfigMap(template, name string) (*ConfigMap, error) {
	if len(template) > 0 {
		// set the base struct from the associated template file
		value, err := base64.StdEncoding.DecodeString(template)
		if err != nil {
			log.Error("Failed to decode the configMap template content",
				log.Fields{
					"Error": err.Error()})
			return &ConfigMap{}, err
		}

		if len(value) > 0 {
			configMap := ConfigMap{}
			err = yamlV2.Unmarshal(value, &configMap)
			if err != nil {
				log.Error("Failed to unmarshal the configMap template content",
					log.Fields{
						"Error": err.Error()})
				return &ConfigMap{}, err
			}

			if err = validateConfigMap(configMap); err != nil {
				return &configMap, err
			}

			return &configMap, nil
		}
	}

	// construct the ConfigMap base struct since there is no template associated with the ConfigMap
	return &ConfigMap{
		APIVersion: "v1",
		Kind:       "ConfigMap",
		MetaData: MetaData{
			Name: name,
		},
		Data: map[string]string{},
	}, nil
}

// handleConfigMapCustomization adds the specified customization data to the ConfigMap
func handleConfigMapCustomization(cm *ConfigMap, customizations []module.Content) error {
	// the number of customization file contents and filenames should be equal and in the same order
	for _, c := range customizations {
		// checks whether the key name is valid
		err := validateConfigMapDataKey(cm, c.KeyName)
		if err != nil {
			return err
		}

		content, err := decodeString(c.Content)
		if err != nil {
			return err
		}

		cm.Data[c.KeyName] = string(content)
	}

	return nil
}

// validateConfigMapDataKey checks whether the data key name is valid
func validateConfigMapDataKey(cm *ConfigMap, key string) error {
	if errs := validation.IsConfigMapKey(key); len(errs) > 0 {
		return fmt.Errorf("%q is not a valid key name for a ConfigMap: %s", key, strings.Join(errs, ","))
	}
	if _, exists := cm.Data[key]; exists {
		return fmt.Errorf("cannot add key %q, another key by that name already exists in Data for ConfigMap %q", key, cm.MetaData.Name)
	}
	return nil
}

// validateConfigMap checks whether the configmap has basic configurations
func validateConfigMap(cm ConfigMap) error {
	var err []string
	if len(cm.APIVersion) == 0 {
		err = append(err, "apiVersion not set for configmap")
	}
	if len(cm.Kind) == 0 ||
		strings.ToLower(cm.Kind) != "configmap" {
		err = append(err, "kind not set for configmap")
	}
	if len(cm.MetaData.Name) == 0 {
		err = append(err, "configmap name may not be empty")
	}

	if len(err) > 0 {
		return errors.New(strings.Join(err, "\n"))
	}

	return nil
}
