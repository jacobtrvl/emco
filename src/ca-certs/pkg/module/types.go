// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"time"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"
)

// ClusterGroup
type ClusterGroup struct {
	MetaData types.Metadata   `json:"metadata"`
	Spec     ClusterGroupSpec `json:"spec"`
}

// ClusterGroupSpec
type ClusterGroupSpec struct {
	Label    string `json:"label,omitempty"`   // select all the clusters with the specific label within the cluster-provider
	Cluster  string `json:"cluster,omitempty"` // select the specific cluster within the cluster-provider
	Provider string `json:"clusterProvider"`
	Scope    string `json:"scope"` // indicates label or name should be used to select the cluster from cluster-provider
}

// CaCertStatus
type CaCertStatus struct {
	ClusterProvider           string `json:"clusterProvider,omitempty"`
	Project                   string `json:"project,omitempty"`
	status.CaCertStatusResult `json:",inline"`
}

// Cert
type Cert struct {
	MetaData types.Metadata `json:"metadata"`
	Spec     CertSpec       `json:"spec"`
}

// CertSpec
type CertSpec struct {
	CertificateSigningInfo CertificateSigningInfo `json:"csrInfo" yaml:"csrInfo"`               // represent the certificate signining request(CSR) info
	Duration               time.Duration          `json:"duration,omitempty"`                   // duration of the certificate
	IsCA                   bool                   `json:"isCA,omitempty" yaml:"isCA,omitempty"` // specifies the cert is a CA or not
	IssuerRef              certissuer.IssuerRef   `json:"issuerRef"`                            // the details of the issuer for signing the certificate request
	IssuingCluster         IssuingClusterInfo     `json:"issuingCluster"`                       // the details of the issuing cluster
}

// CertificateSigningInfo
type CertificateSigningInfo struct {
	KeySize        int       `json:"keySize,omitempty"`
	Version        int       `json:"version,omitempty"`
	DNSNames       []string  `json:"dnsNames,omitempty"`
	EmailAddresses []string  `json:"emailAddresses,omitempty"`
	KeyUsages      []string  `json:"keyUsages,omitempty"` // certificate usages
	Algorithm      Algorithm `json:"algorithm"`
	Subject        Subject   `json:"subject"`
}

// Subject
type Subject struct {
	Locale       Locale       `json:"locale"`
	Names        Names        `json:"names"`
	Organization Organization `json:"organization"`
}

// TODO: Confirm if this is required. Common name should be unique for each csr
type Names struct {
	CommonNamePrefix string `json:"commonNamePrefix"`
	CommonName       string
}

// Locale
type Locale struct {
	Country       []string `json:"country,omitempty"`
	Locality      []string `json:"locality,omitempty"`
	PostalCode    []string `json:"postalCode,omitempty"`
	Province      []string `json:"province,omitempty"`
	StreetAddress []string `json:"streetAddress,omitempty"`
}

// Organization
type Organization struct {
	Names []string `json:"names,omitempty"`
	Units []string `json:"units,omitempty"`
}

// Algorithm
type Algorithm struct {
	PublicKeyAlgorithm string `json:"publicKeyAlgorithm,omitempty"`
	SignatureAlgorithm string `json:"signatureAlgorithm,omitempty"`
}

// IssuingClusterInfo
type IssuingClusterInfo struct {
	Cluster         string `json:"cluster"`         // name of the cluster
	ClusterProvider string `json:"clusterProvider"` // name of the cluster provider
}

// Key
type Key struct {
	Name string
	Val  string
}