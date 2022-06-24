// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package istioservice

import (
	"fmt"
)

// newProxyConfig create the ProxyConfig
func newProxyConfig() *ProxyConfig {
	// construct the ProxyConfig base struct
	return &ProxyConfig{
		APIVersion: "networking.istio.io/v1beta1",
		Kind:       "ProxyConfig",
		Spec: ProxyConfigSpec{
			EnvironmentVariables: map[string]string{}}}
}

// ProxyConfigName
func ProxyConfigName(contextID, cert, clusterProvider, cluster, namespace string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s-%s", contextID, cert, clusterProvider, cluster, namespace, "pc")
}

// CreateProxyConfig
func CreateProxyConfig(name, namespace string, environmentVariables map[string]string) *ProxyConfig {
	pc := newProxyConfig()
	pc.MetaData.Name = name
	pc.MetaData.Namespace = namespace

	for key, val := range environmentVariables {
		pc.Spec.EnvironmentVariables[key] = val
	}

	return pc
}
