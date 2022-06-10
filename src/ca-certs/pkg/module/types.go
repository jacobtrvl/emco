// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"
)

// MetaData holds the data
type MetaData struct {
	Name        string `json:"name" yaml:"name"`
	Namespace   string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	UserData1   string `json:"userData1,omitempty" yaml:"userData1,omitempty"`
	UserData2   string `json:"userData2,omitempty" yaml:"userData2,omitempty"`
}

// ClusterGroup
type ClusterGroup struct {
	MetaData MetaData         `json:"metadata"`
	Spec     ClusterGroupSpec `json:"spec"`
}

// ClusterGroupSpec
type ClusterGroupSpec struct {
	Label    string `json:"label,omitempty"` // select all the clusters with the specific label within the cluster-provider
	Name     string `json:"name,omitempty"`  // select the specific cluster within the cluster-provider
	Scope    string `json:"scope"`           // indicates label or name should be used to select the cluster from cluster-provider
	Provider string `json:"clusterProvider"`
}

// CaCertStatus
type CaCertStatus struct {
	ClusterProvider           string `json:"clusterProvider,omitempty"`
	Project                   string `json:"project,omitempty"`
	status.CaCertStatusResult `json:",inline"`
}

type Cert struct {
	MetaData MetaData `json:"metadata"`
	Spec     CertSpec `json:"spec"`
}

// CertSpec
type CertSpec struct {
	CertificateAuthority   bool                   `json:"isCA,omitempty" yaml:"isCA,omitempty"` // specifies the cert is a CA or not
	CertificateSigningInfo CertificateSigningInfo `json:"csrInfo" yaml:"csrInfo"`               // represent the certificate signining request(CSR) csrInfo
	IssuerRef              certissuer.IssuerRef   `json:"issuerRef"`                            // the details of the issuer for signing the certificate request
	Duration               string                 `json:"duration,omitempty"`                   // duration of the certificate
	IssuingCluster         IssuingClusterInfo     `json:"issuingCluster"`                       // the details of the issuing cluster
	Request                string                 `json:"request,omitempty"`
}

type CertificateSigningInfo struct {
	KeySize        int       `json:"keySize,omitempty"`
	Version        int       `json:"version,omitempty"`
	DNSNames       []string  `json:"dnsNames,omitempty"`
	EmailAddresses []string  `json:"emailAddresses,omitempty"`
	KeyUsages      []string  `json:"keyUsages,omitempty"` // certificate usages
	Algorithm      Algorithm `json:"algorithm"`
	Subject        Subject   `json:"subject"`
}

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

type Locale struct {
	Country       []string `json:"country,omitempty"`
	Locality      []string `json:"locality,omitempty"`
	PostalCode    []string `json:"postalCode,omitempty"`
	Province      []string `json:"province,omitempty"`
	StreetAddress []string `json:"streetAddress,omitempty"`
}

type Organization struct {
	Names []string `json:"names,omitempty"`
	Units []string `json:"units,omitempty"`
}
type Algorithm struct {
	PublicKeyAlgorithm string `json:"publicKeyAlgorithm,omitempty"`
	SignatureAlgorithm string `json:"signatureAlgorithm,omitempty"`
}

type IssuingClusterInfo struct {
	Cluster         string `json:"cluster"`         // name of the cluster
	ClusterProvider string `json:"clusterProvider"` // name of the cluster provider
}

type CertificateRequestStatusKey struct {
	Cert        string `json:"cert"`
	Cluster     string `json:"cluster"`
	CertRequest string `json:"certRequest"`
}
