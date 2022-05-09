// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certificate

// DbInfo holds the MongoDB collection and attributes info
type DbInfo struct {
	storeName string // name of the mongodb collection to use for client documents
	tagMeta   string // attribute key name for the json data of a client document
}

type Cert struct {
	MetaData MetaData `json:"metadata"`
	Spec     CertSpec `json:"spec"`
}

// MetaData holds the data
type MetaData struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Description string `json:"description,omitempty"`
	UserData1   string `json:"userData1,omitempty"`
	UserData2   string `json:"userData2,omitempty"`
}

// CertSpec
type CertSpec struct {
	CertificateAuthority   bool                   `json:"isCA,omitempty"`     // specifies the cert is a CA or not
	CertificateSigningInfo CertificateSigningInfo `json:"template"`           // represent the certificate signining request(CSR) template
	IssuerRef              IssuerRef              `json:"issuerRef"`          // the details of the issuer for signing the certificate request
	Duration               string                 `json:"duration,omitempty"` // duration of the certificate
	IssuingCluster         IssuingClusterInfo     `json:"issuingCluster"`     // the details of the issuing cluster
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

type Names struct {
	CommonName string `json:"commonName"`
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

type IssuerRef struct {
	Name  string `json:"name"`  // name of the issuer
	Kind  string `json:"kind"`  // kind of the issuer
	Group string `json:"group"` // group of the issuer
}

type IssuingClusterInfo struct {
	Cluster         string `json:"cluster"`         // name of the cluster
	ClusterProvider string `json:"clusterProvider"` // name of the cluster provider
}

type CertificateRequest struct {
	ApiVersion string                 `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty"`
	Kind       string                 `yaml:"kind,omitempty" json:"kind,omitempty"`
	MetaData   CertificateRequestMeta `yaml:"metadata,omitempty" json:"metadata,omitempty"`
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
	CertificateAuthority bool      `yaml:"isCA,omitempty" json:"isCA,omitempty"`
	Duration             string    `yaml:"duration,omitempty" json:"duration,omitempty"`
	IssuerRef            IssuerRef `yaml:"issuerRef,omitempty" json:"issuerRef,omitempty"`
	Request              string    `yaml:"request,omitempty" json:"request,omitempty"`
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

// type CertificateRequestStatus struct {
// 	ApiVersion string `yaml:"apiVersion"`
// 	Kind       string
// 	MetaData   CertificateRequestStatusMetadata
// 	Spec       CertificateRequestStatusSpec
// 	Status     CertStatus
// }

// MetaData holds the data
// type CertificateRequestMeta struct {
// 	CreationTimestamp string `yaml:"creationTimestamp"`
// 	Generation        int64
// 	Name              string
// 	Namespace         string
// }

// CertSpec
// type CertificateRequestStatusSpec struct {
// 	CertificateAuthority bool      `yaml:"isCA"`      // specifies the cert is a CA or not
// 	Duration             string    `yaml:"duration"`  // duration of the certificate
// 	IssuerRef            IssuerRef `yaml:"issuerRef"` // the details of the issuer for signing the certificate request
// 	Request              string    `json:"request,omitempty"`
// 	Status               CertStatus
// }
