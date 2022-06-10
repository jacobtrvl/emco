// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package istioservice

import "gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"

// ProxyConfig holds secret data of a certain type
type ProxyConfig struct {
	APIVersion string          `yaml:"apiVersion" json:"apiVersion"`
	Kind       string          `yaml:"kind" json:"kind"`
	MetaData   module.MetaData `yaml:"metadata" json:"metadata"`
	Spec       ProxyConfigSpec `yaml:"spec" json:"spec"`
}

type ProxyConfigSpec struct {
	EnvironmentVariables map[string]string `yaml:"environmentVariables" json:"environmentVariables"`
}
