// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certificate

type CertificateRequestInfo struct {
	Version                                int
	SignatureAlgorithm, PublicKeyAlgorithm string
	DNSNames, EmailAddresses               []string
	Subject                                SubjectInfo
}

type SubjectInfo struct {
	CommonName                                             string
	Country, Locality, PostalCode, Province, StreetAddress []string
	Organization, OrganizationalUnit                       []string
}
