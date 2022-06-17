// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package knccservice

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newKnccConfig create the KnccConfig
func newKnccConfig() *Config {
	// construct the KnccConfig base struct
	return &Config{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kncc.k8sappconfig.com.gitlab.com/v1alpha1",
			Kind:       "ConfigCtrl",
		},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       ConfigSpec{},
	}
}

// KnccConfigName
func KnccConfigName(contextID, cert, clusterProvider, cluster string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", contextID, cert, clusterProvider, cluster, "kncc")
}

// CreateKnccConfig
func CreateKnccConfig(name, namespace, resourceName, resourceNamespace string,
	patch []map[string]string) *Config {
	c := newKnccConfig()
	c.ObjectMeta.Name = name
	c.ObjectMeta.Namespace = namespace
	c.Spec.Resource.Name = resourceName
	c.Spec.Resource.NameSpace = resourceNamespace

	for _, p := range patch {
		for k, v := range p {
			c.Spec.Patch = append(c.Spec.Patch, Patch{
				Key:   k,
				Value: v,
			})
		}
	}

	return c
}
