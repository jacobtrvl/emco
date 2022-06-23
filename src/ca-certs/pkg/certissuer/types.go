// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certissuer

// IssuerRef
type IssuerRef struct {
	Name    string `json:"name"`  // name of the issuer
	Kind    string `json:"kind"`  // kind of the issuer
	Group   string `json:"group"` // group of the issuer
	Version string `json:"version,omitempty"`
}

// ResourceBundleStateStatus
type ResourceBundleStateStatus struct {
	Ready            bool             `json:"ready"`
	ResourceCount    int32            `json:"resourceCount"`
	ResourceStatuses []ResourceStatus `json:"resourceStatuses,omitempty" protobuf:"varint,14,opt,name=resourceStatuses"`
}

// Resources
type ResourceStatus struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Res       string `json:"res"`
}
