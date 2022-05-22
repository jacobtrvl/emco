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
func ProxyConfigName(contextID, cert, clusterProvider, cluster string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", contextID, cert, clusterProvider, cluster, "pc")
}

// ResourceName
func (pc *ProxyConfig) ResourceName() string {
	return fmt.Sprintf("%s+%s", pc.MetaData.Name, "proxyconfig")
}

// CreateProxyConfig
func CreateProxyConfig(name, namespace string, environmentVariables map[string]string) *ProxyConfig {
	pc := newProxyConfig()

	pc.MetaData.Name = name

	if len(namespace) > 0 {
		pc.MetaData.Namespace = namespace
	}

	for key, val := range environmentVariables {
		pc.Spec.EnvironmentVariables[key] = val
	}

	return pc
}
