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

// Secret holds the secret data
type Secret struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	MetaData   MetaData          `yaml:"metadata"`
	Type       string            `yaml:"type"`
	Data       map[string]string `yaml:"data"`
}

// createSecret create the Secret data based on the JSON  patch,
// content in the template file, and the customization file, if any
func (o *updateOptions) createSecret() error {
	// create a new Secret object based on the template file
	secret, err := newSecret(o.resourceContent.Content, o.resource.Spec.ResourceGVK.Name)
	if err != nil {
		return err
	}

	if len(o.customizationContent.Content) > 0 {
		// apply the customization data to the Secret
		if err := handleSecretCustomization(secret, o.customizationContent.Content); err != nil {
			return err
		}
	}

	value, err := yamlV2.Marshal(secret)
	if err != nil {
		log.Error("Failed to serialize the secret object into a YAML document",
			log.Fields{
				"Secret": secret,
				"Error ": err.Error()})
		return err
	}

	if strings.ToLower(o.customization.Spec.PatchType) == "json" &&
		len(o.customization.Spec.PatchJSON) > 0 {
		// apply the JSON patch associated with the Secret customization
		modifiedPatch, err := applyPatch(o.customization.Spec.PatchJSON, value)
		if err != nil {
			return err
		}
		value = modifiedPatch
	}

	// create the Secret
	if err = o.create(value); err != nil {
		return err
	}

	return nil
}

// newSecret creates a new Secret object based on the template file
func newSecret(template, name string) (*Secret, error) {
	if len(template) > 0 {
		// set the base struct from the associated template file
		value, err := base64.StdEncoding.DecodeString(template)
		if err != nil {
			log.Error("Failed to decode the secret template content",
				log.Fields{
					"Error": err.Error()})
			return &Secret{}, err
		}

		if len(value) > 0 {
			secret := Secret{}
			// if the Secret template is available, then it should be YAML
			err = yamlV2.Unmarshal(value, &secret)
			if err != nil {
				log.Error("Failed to unmarshal the secret template content",
					log.Fields{
						"Error": err.Error()})
				return &Secret{}, err
			}

			if len(secret.Type) == 0 {
				secret.Type = "Opaque"
			}

			if err = validateSecret(secret); err != nil {
				return &secret, err
			}

			return &secret, nil
		}
	}

	// construct the Secret base struct since there is no template associated with the Secret
	return &Secret{
		APIVersion: "v1",
		Kind:       "Secret",
		Type:       "Opaque",
		MetaData: MetaData{
			Name: name,
		},
		Data: map[string]string{},
	}, nil
}

// handleSecretCustomization adds the specified customization data to the Secret
func handleSecretCustomization(s *Secret, customizations []module.Content) error {
	// the number of customization file contents and filenames should be equal and in the same order
	for _, c := range customizations {
		// checks whether the key name is valid
		err := validateSecretDataKey(s, c.KeyName)
		if err != nil {
			return err
		}

		s.Data[c.KeyName] = string([]byte(c.Content))
	}

	return nil
}

// validateSecretDataKey checks whether the data key name is valid
func validateSecretDataKey(s *Secret, key string) error {
	if errs := validation.IsConfigMapKey(key); len(errs) > 0 {
		return fmt.Errorf("%q is not a valid key name for a Secret: %s", key, strings.Join(errs, ","))
	}
	if _, exists := s.Data[key]; exists {
		return fmt.Errorf("cannot add key %q, another key by that name already exists in Data for Secret %q", key, s.MetaData.Name)
	}
	return nil
}

// validateSecret checks whether the secret has basic configurations
func validateSecret(s Secret) error {
	var err []string
	if len(s.APIVersion) == 0 {
		err = append(err, "apiVersion not set for secret")
	}
	if len(s.Kind) == 0 ||
		strings.ToLower(s.Kind) != "secret" {
		err = append(err, "kind not set for secret")
	}
	if len(s.MetaData.Name) == 0 {
		err = append(err, "secret name may not be empty")
	}

	if len(err) > 0 {
		return errors.New(strings.Join(err, "\n"))
	}

	return nil
}
