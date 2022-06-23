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

// CaCertManager
type CaCertManager interface {
	// Certificates
	CreateCert(cert module.CaCert, clusterProvider string, failIfExists bool) (module.CaCert, bool, error)
	DeleteCert(cert, clusterProvider string) error
	GetAllCert(clusterProvider string) ([]module.CaCert, error)
	GetCert(cert, clusterProvider string) (module.CaCert, error)
}

// CaCertKey
type CaCertKey struct {
	Cert            string `json:"caCertCp"`
	ClusterProvider string `json:"clusterProvider"`
}

// CaCertClient
type CaCertClient struct {
}

// NewCaCertClient
func NewCaCertClient() *CaCertClient {
	return &CaCertClient{}
}

// CreateCert
func (c *CaCertClient) CreateCert(cert module.CaCert, clusterProvider string, failIfExists bool) (module.CaCert, bool, error) {
	certExists := false
	ck := CaCertKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: clusterProvider}
	cc := module.NewCaCertClient(ck)

	if cer, err := cc.GetCert(); err == nil &&
		!reflect.DeepEqual(cer, module.CaCert{}) {
		certExists = true
	}

	if certExists &&
		failIfExists {
		return module.CaCert{}, certExists, errors.New("Certificate already exists")
	}

	if certExists {
		// check the enrollment state
		if err := verifyEnrollmentStateBeforeUpdate(cert.MetaData.Name, clusterProvider); err != nil {
			return module.CaCert{}, certExists, err
		}

		// check the distribution state
		if err := verifyDistributionStateBeforeUpdate(cert.MetaData.Name, clusterProvider); err != nil {
			return module.CaCert{}, certExists, err
		}

		return cert, certExists, cc.UpdateCert(cert)
	}

	_, certExists, err := cc.CreateCert(cert, failIfExists)
	if err != nil {
		return module.CaCert{}, certExists, err
	}

	// create enrollment stateInfo
	ek := EnrollmentKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: clusterProvider,
		Enrollment:      enrollment.AppName}
	sc := module.NewStateClient(ek)
	if err := sc.Create(""); err != nil {
		return module.CaCert{}, certExists, err
	}

	// create distribution stateInfo
	dk := DistributionKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: clusterProvider,
		Distribution:    distribution.AppName}
	sc = module.NewStateClient(dk)
	if err := sc.Create(""); err != nil {
		return module.CaCert{}, certExists, err
	}

	return cert, certExists, nil
}

// DeleteCert
func (c *CaCertClient) DeleteCert(cert, clusterProvider string) error {
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
	ck := CaCertKey{
		Cert:            cert,
		ClusterProvider: clusterProvider}

	return module.NewCaCertClient(ck).DeleteCert()
}

// GetAllCert
func (c *CaCertClient) GetAllCert(clusterProvider string) ([]module.CaCert, error) {
	ck := CaCertKey{
		ClusterProvider: clusterProvider}

	return module.NewCaCertClient(ck).GetAllCert()
}

// GetCert
func (c *CaCertClient) GetCert(cert, clusterProvider string) (module.CaCert, error) {
	ck := CaCertKey{
		Cert:            cert,
		ClusterProvider: clusterProvider}

	return module.NewCaCertClient(ck).GetCert()
}

// verifyEnrollmentStateBeforeDelete
func verifyEnrollmentStateBeforeDelete(cert, clusterProvider string) error {
	k := EnrollmentKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Enrollment:      enrollment.AppName}

	return module.NewCaCertClient(k).VerifyStateBeforeDelete(cert, enrollment.AppName)
}

// verifyDistributionStateBeforeDelete
func verifyDistributionStateBeforeDelete(cert, clusterProvider string) error {
	k := DistributionKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Distribution:    distribution.AppName}

	return module.NewCaCertClient(k).VerifyStateBeforeDelete(cert, distribution.AppName)

}

// verifyEnrollmentStateBeforeUpdate
func verifyEnrollmentStateBeforeUpdate(cert, clusterProvider string) error {
	k := EnrollmentKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Enrollment:      enrollment.AppName}

	return module.NewCaCertClient(k).VerifyStateBeforeUpdate(cert, enrollment.AppName)
}

// verifyDistributionStateBeforeUpdate
func verifyDistributionStateBeforeUpdate(cert, clusterProvider string) error {
	k := DistributionKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Distribution:    distribution.AppName}

	return module.NewCaCertClient(k).VerifyStateBeforeUpdate(cert, distribution.AppName)

}
