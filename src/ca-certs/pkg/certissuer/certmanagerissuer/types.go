// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certmanagerissuer

import (
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
)

type CertificateRequest struct {
	ApiVersion string                 `yaml:"apiVersion" json:"apiVersion"`
	Kind       string                 `yaml:"kind" json:"kind"`
	MetaData   CertificateRequestMeta `yaml:"metadata" json:"metadata,omitempty"`
	Spec       CertificateRequestSpec `yaml:"spec,omitempty" json:"spec,omitempty"`
	Status     CertStatus             `yaml:"status,omitempty" json:"status,omitempty"`
}

// MetaData holds the data
type CertificateRequestMeta struct {
	CreationTimestamp string `yaml:"creationTimestamp,omitempty" json:"creationTimestamp,omitempty"`
	Generation        int64  `yaml:"generation,omitempty" json:"generation,omitempty"`
	Name              string `yaml:"name,omitempty" json:"name,omitempty"`
	Namespace         string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
}

type CertificateRequestSpec struct {
	CertificateAuthority bool                 `yaml:"isCA,omitempty" json:"isCA,omitempty"`
	Duration             string               `yaml:"duration,omitempty" json:"duration,omitempty"`
	IssuerRef            certissuer.IssuerRef `yaml:"issuerRef,omitempty" json:"issuerRef,omitempty"`
	Request              string               `yaml:"request,omitempty" json:"request,omitempty"`
}

type CertStatus struct {
	Certificate          string      `yaml:"certificate,omitempty" json:"certificate,omitempty"`
	CertificateAuthority string      `yaml:"ca,omitempty" json:"ca,omitempty"`
	States               []CertState `yaml:"conditions,omitempty" json:"conditions,omitempty"`
}

type CertState struct {
	LastTransitionTime string `yaml:"lastTransitionTime,omitempty" json:"lastTransitionTime,omitempty"`
	Message            string `yaml:"message,omitempty" json:"message,omitempty"`
	Reason             string `yaml:"reason,omitempty" json:"reason,omitempty"`
	Status             string `yaml:"status,omitempty" json:"status,omitempty"`
	Type               string `yaml:"type,omitempty" json:"type,omitempty"`
}

// Secret holds secret data of a certain type
type ClusterIssuer struct {
	APIVersion string          `yaml:"apiVersion" json:"apiVersion"`
	Kind       string          `yaml:"kind" json:"kind"`
	MetaData   module.MetaData `yaml:"metadata" json:"metadata"`
	Spec       IssuerSpec      `yaml:"spec" json:"spec"`
}

type IssuerSpec struct {
	CA Issuer `yaml:"ca" json:"ca"`
}

type Issuer struct {
	SecretName string `yaml:"secretName" json:"secretName"`
}

// Secret holds secret data of a certain type
type Secret struct {
	APIVersion string            `yaml:"apiVersion" json:"apiVersion"`
	Kind       string            `yaml:"kind" json:"kind"`
	MetaData   module.MetaData   `yaml:"metadata" json:"metadata"`
	Type       string            `yaml:"type" json:"type"`
	Data       map[string]string `yaml:"data" json:"data"`
}
