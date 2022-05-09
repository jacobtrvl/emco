// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certificate

import (
	"fmt"
)

// Secret holds secret data of a certain type
type Secret struct {
	APIVersion string            `yaml:"apiVersion" json:"apiVersion"`
	Kind       string            `yaml:"kind" json:"kind"`
	MetaData   MetaData          `yaml:"metadata" json:"metadata"`
	Type       string            `yaml:"type" json:"type"`
	Data       map[string]string `yaml:"data" json:"data"`
}

// newSecret creates a new Secret object based on the template file
func CreateSecret() *Secret {
	// construct the Secret base struct since there is no template associated with the Secret
	return &Secret{
		APIVersion: "v1",
		Kind:       "Secret",
		Type:       "Opaque",
		Data:       map[string]string{}}
}

func (s *Secret) ResourceName() string {
	return fmt.Sprintf("%s+%s", s.MetaData.Name, "secret")
}
