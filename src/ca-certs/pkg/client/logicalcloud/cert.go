// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import (
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate/distribution"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate/enrollment"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
)

type CertManager interface {
	// Certificates
	CreateCert(cert module.Cert, project string, failIfExists bool) (module.Cert, bool, error)
	DeleteCert(cert, project string) error
	GetAllCert(project string) ([]module.Cert, error)
	GetCert(cert, project string) (module.Cert, error)
}

// CertKey
type CertKey struct {
	Cert    string `json:"logicalCloudCert"`
	Project string `json:"project"`
}

// CertClient
type CertClient struct {
}

// NewCertClient
func NewCertClient() *CertClient {
	return &CertClient{}
}

// CreateCertificates
func (c *CertClient) CreateCert(cert module.Cert, project string, failIfExists bool) (module.Cert, bool, error) {
	certExists := false
	ck := CertKey{
		Cert:    cert.MetaData.Name,
		Project: project}

	_, certExists, err := module.NewCertClient(ck).CreateCert(cert, failIfExists)
	if err != nil {
		return module.Cert{}, certExists, err
	}

	// create the enrollment stateInfo
	ek := EnrollmentKey{
		Cert:       cert.MetaData.Name,
		Project:    project,
		Enrollment: enrollment.AppName}
	sc := module.NewStateClient(ek)
	if err := sc.Create(""); err != nil {
		return module.Cert{}, certExists, err
	}

	// create the distribution stateInfo
	dk := DistributionKey{
		Cert:         cert.MetaData.Name,
		Project:      project,
		Distribution: distribution.AppName}
	sc = module.NewStateClient(dk)
	if err := sc.Create(""); err != nil {
		return module.Cert{}, certExists, err
	}

	return cert, certExists, nil
}

// DeleteCertificates
func (c *CertClient) DeleteCert(cert, project string) error {
	// delete the enrollment stateInfo resource
	ek := EnrollmentKey{
		Cert:       cert,
		Project:    project,
		Enrollment: enrollment.AppName}
	sc := module.NewStateClient(ek)
	if err := sc.Delete(); err != nil {
		return err
	}

	// delete the distribution stateInfo resource
	dk := DistributionKey{
		Cert:         cert,
		Project:      project,
		Distribution: distribution.AppName}
	sc = module.NewStateClient(dk)
	if err := sc.Delete(); err != nil {
		return err
	}

	ck := CertKey{
		Cert:    cert,
		Project: project}

	return module.NewCertClient(ck).DeleteCert()
}

// GetAllCert
func (c *CertClient) GetAllCert(project string) ([]module.Cert, error) {
	ck := CertKey{
		Project: project}

	return module.NewCertClient(ck).GetAllCert()
}

// GetCertificates
func (c *CertClient) GetCert(cert, project string) (module.Cert, error) {
	ck := CertKey{
		Cert:    cert,
		Project: project}

	return module.NewCertClient(ck).GetCert()
}

// // Convert the key to string to preserve the underlying structure
// func (k CertKey) String() string {
// 	out, err := json.Marshal(k)
// 	if err != nil {
// 		return ""
// 	}
// 	return string(out)
// }
