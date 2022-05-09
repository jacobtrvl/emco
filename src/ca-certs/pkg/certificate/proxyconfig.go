// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certificate

import (
	"fmt"
)

// Secret holds secret data of a certain type
type ProxyConfig struct {
	APIVersion string          `yaml:"apiVersion" json:"apiVersion"`
	Kind       string          `yaml:"kind" json:"kind"`
	MetaData   MetaData        `yaml:"metadata" json:"metadata"`
	Spec       ProxyConfigSpec `yaml:"spec" json:"spec"`
}

type ProxyConfigSpec struct {
	EnvironmentVariables map[string]string
}

// createSecret create the Secret based on the JSON  patch,
// content in the template file, and the customization file, if any
func CreateProxyConfig() *ProxyConfig {
	// construct the Secret base struct since there is no template associated with the Secret
	return &ProxyConfig{
		APIVersion: "networking.istio.io/v1beta1",
		Kind:       "ProxyConfig",
		Spec: ProxyConfigSpec{
			EnvironmentVariables: map[string]string{}}}
}

func (pc *ProxyConfig) ResourceName() string {
	return fmt.Sprintf("%s+%s", pc.MetaData.Name, "proxyconfig")
}
