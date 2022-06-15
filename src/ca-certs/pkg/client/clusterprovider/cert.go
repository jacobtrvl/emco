// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

import (
	"reflect"

	"github.com/pkg/errors"
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
	cc := module.NewCertClient(ck)

	if cer, err := cc.GetCert(); err == nil &&
		!reflect.DeepEqual(cer, module.Cert{}) {
		certExists = true
	}

	if certExists &&
		failIfExists {
		return module.Cert{}, certExists, errors.New("Certificate already exists")
	}

	if certExists {
		// check the enrollment state
		if err := verifyEnrollmentStateBeforeUpdate(cert.MetaData.Name, clusterProvider); err != nil {
			return module.Cert{}, certExists, err
		}

		// check the distribution state
		if err := verifyDistributionStateBeforeUpdate(cert.MetaData.Name, clusterProvider); err != nil {
			return module.Cert{}, certExists, err
		}

		return cert, certExists, cc.UpdateCert(cert)
	}

	_, certExists, err := cc.CreateCert(cert, failIfExists)
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
	// check the enrollment state
	if err := verifyEnrollmentStateBeforeDelete(cert, clusterProvider); err != nil {
		// if the StateInfo cannot be found, then a caCert record may not present
		if err.Error() != "StateInfo not found" {
			return err
		}
	}

	// check the distribution state
	if err := verifyDistributionStateBeforeDelete(cert, clusterProvider); err != nil {
		// if the StateInfo cannot be found, then a caCert record may not present
		if err.Error() != "StateInfo not found" {
			return err
		}
	}

	// delete enrollment stateInfo
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

	// delete caCert
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

// verifyEnrollmentStateBeforeDelete
func verifyEnrollmentStateBeforeDelete(cert, clusterProvider string) error {
	k := EnrollmentKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Enrollment:      enrollment.AppName}

	return module.NewCertClient(k).VerifyStateBeforeDelete(cert, enrollment.AppName)
}

// verifyDistributionStateBeforeDelete
func verifyDistributionStateBeforeDelete(cert, clusterProvider string) error {
	k := DistributionKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Distribution:    distribution.AppName}

	return module.NewCertClient(k).VerifyStateBeforeDelete(cert, distribution.AppName)

}

// verifyEnrollmentStateBeforeUpdate
func verifyEnrollmentStateBeforeUpdate(cert, clusterProvider string) error {
	k := EnrollmentKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Enrollment:      enrollment.AppName}

	return module.NewCertClient(k).VerifyStateBeforeUpdate(cert, enrollment.AppName)
}

// verifyDistributionStateBeforeUpdate
func verifyDistributionStateBeforeUpdate(cert, clusterProvider string) error {
	k := DistributionKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Distribution:    distribution.AppName}

	return module.NewCertClient(k).VerifyStateBeforeUpdate(cert, distribution.AppName)

}

// // Convert the key to string to preserve the underlying structure
// func (k CertKey) String() string {
// 	out, err := json.Marshal(k)
// 	if err != nil {
// 		return ""
// 	}
// 	return string(out)
// }
