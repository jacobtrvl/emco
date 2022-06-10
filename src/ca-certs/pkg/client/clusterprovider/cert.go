// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

import (
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate/distribution"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate/enrollment"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
)

// CertManager
type CertManager interface {
	// Certificates
	CreateCert(cert module.Cert, clusterProvider string, failIfExists bool) (module.Cert, bool, error)
	DeleteCert(cert, clusterProvider string) error
	GetAllCert(clusterProvider string) ([]module.Cert, error)
	GetCert(cert, clusterProvider string) (module.Cert, error)
}

// CertKey
type CertKey struct {
	Cert            string `json:"caCertCp"`
	ClusterProvider string `json:"clusterProvider"`
}

// CertClient
type CertClient struct {
}

// NewCertClient
func NewCertClient() *CertClient {
	return &CertClient{}
}

// CreateCert
func (c *CertClient) CreateCert(cert module.Cert, clusterProvider string, failIfExists bool) (module.Cert, bool, error) {
	certExists := false
	ck := CertKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: clusterProvider}

	_, certExists, err := module.NewCertClient(ck).CreateCert(cert, failIfExists)
	if err != nil {
		return module.Cert{}, certExists, err
	}

	// create enrollment stateInfo
	ek := EnrollmentKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: clusterProvider,
		Enrollment:      enrollment.AppName}
	sc := module.NewStateClient(ek)
	if err := sc.Create(""); err != nil {
		return module.Cert{}, certExists, err
	}

	// create distribution stateInfo
	dk := DistributionKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: clusterProvider,
		Distribution:    distribution.AppName}
	sc = module.NewStateClient(dk)
	if err := sc.Create(""); err != nil {
		return module.Cert{}, certExists, err
	}

	return cert, certExists, nil
}

// DeleteCert
func (c *CertClient) DeleteCert(cert, clusterProvider string) error {
	// delete enrollemnt stateInfo object
	ek := EnrollmentKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Enrollment:      enrollment.AppName}
	sc := module.NewStateClient(ek)
	if err := sc.Delete(); err != nil {
		return err
	}

	// delete distribution stateInfo
	dk := DistributionKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Distribution:    distribution.AppName}
	sc = module.NewStateClient(dk)
	if err := sc.Delete(); err != nil {
		return err
	}

	ck := CertKey{
		Cert:            cert,
		ClusterProvider: clusterProvider}

	return module.NewCertClient(ck).DeleteCert()
}

// GetAllCertificate
func (c *CertClient) GetAllCert(clusterProvider string) ([]module.Cert, error) {
	ck := CertKey{
		ClusterProvider: clusterProvider}

	return module.NewCertClient(ck).GetAllCert()
}

// GetCert
func (c *CertClient) GetCert(cert, clusterProvider string) (module.Cert, error) {
	ck := CertKey{
		Cert:            cert,
		ClusterProvider: clusterProvider}

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
