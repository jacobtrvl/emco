// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package knccservice

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Config holds the kncc patch config data
type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ConfigSpec `json:"spec,omitempty"`
}

// ConfigSpec
type ConfigSpec struct {
	Resource Resource `json:"resource,omitempty"`
	Patch    []Patch  `json:"patch,omitempty"`
}

// Resource
type Resource struct {
	Name      string `json:"name,omitempty"`
	NameSpace string `json:"namespace,omitempty"`
}

// Patch
type Patch struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
