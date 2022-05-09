// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certificate

import "fmt"

// Secret holds secret data of a certain type
type ClusterIssuer struct {
	APIVersion string     `yaml:"apiVersion" json:"apiVersion"`
	Kind       string     `yaml:"kind" json:"kind"`
	MetaData   MetaData   `yaml:"metadata" json:"metadata"`
	Spec       IssuerSpec `yaml:"spec" json:"spec"`
}

type IssuerSpec struct {
	CA Issuer
}

type Issuer struct {
	SecretName string
}

// createSecret create the Secret based on the JSON  patch,
// content in the template file, and the customization file, if any
func CreateClusterIssuer() *ClusterIssuer {
	// construct the Secret base struct since there is no template associated with the Secret
	return &ClusterIssuer{
		APIVersion: "cert-manager.io/v1",
		Kind:       "ClusterIssuer"}
}

func (i *ClusterIssuer) ResourceName() string {
	return fmt.Sprintf("%s+%s", i.MetaData.Name, "clusterissuer")
}
