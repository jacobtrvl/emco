// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import (
	"reflect"

	"github.com/pkg/errors"
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
	Cert    string `json:"caCertLc"`
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
		if err := verifyEnrollmentStateBeforeUpdate(cert.MetaData.Name, project); err != nil {
			return module.Cert{}, certExists, err
		}

		// check the distribution state
		if err := verifyDistributionStateBeforeUpdate(cert.MetaData.Name, project); err != nil {
			return module.Cert{}, certExists, err
		}

		return cert, certExists, cc.UpdateCert(cert)
	}

	_, certExists, err := cc.CreateCert(cert, failIfExists)
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
	// check the enrollment state
	if err := verifyEnrollmentStateBeforeDelete(cert, project); err != nil {
		// if the StateInfo cannot be found, then a caCert record may not present
		if err.Error() != "StateInfo not found" {
			return err
		}
	}

	// check the distribution state
	if err := verifyDistributionStateBeforeDelete(cert, project); err != nil {
		// if the StateInfo cannot be found, then a caCert record may not present
		if err.Error() != "StateInfo not found" {
			return err
		}
	}

	// delete enrollment stateInfo
	ek := EnrollmentKey{
		Cert:       cert,
		Project:    project,
		Enrollment: enrollment.AppName}
	sc := module.NewStateClient(ek)
	if err := sc.Delete(); err != nil {
		return err
	}

	// delete distribution stateInfo
	dk := DistributionKey{
		Cert:         cert,
		Project:      project,
		Distribution: distribution.AppName}
	sc = module.NewStateClient(dk)
	if err := sc.Delete(); err != nil {
		return err
	}

	// delete caCert
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

// verifyEnrollmentState
func verifyEnrollmentStateBeforeDelete(cert, project string) error {
	k := EnrollmentKey{
		Cert:       cert,
		Project:    project,
		Enrollment: enrollment.AppName}

	return module.NewCertClient(k).VerifyStateBeforeDelete(cert, enrollment.AppName)
}

// verifyDistributionState
func verifyDistributionStateBeforeDelete(cert, project string) error {
	k := DistributionKey{
		Cert:         cert,
		Project:      project,
		Distribution: distribution.AppName}

	return module.NewCertClient(k).VerifyStateBeforeDelete(cert, distribution.AppName)

}

// verifyEnrollmentStateBeforeUpdate
func verifyEnrollmentStateBeforeUpdate(cert, project string) error {
	k := EnrollmentKey{
		Cert:       cert,
		Project:    project,
		Enrollment: enrollment.AppName}

	return module.NewCertClient(k).VerifyStateBeforeUpdate(cert, enrollment.AppName)
}

// verifyDistributionStateBeforeUpdate
func verifyDistributionStateBeforeUpdate(cert, project string) error {
	k := DistributionKey{
		Cert:         cert,
		Project:      project,
		Distribution: distribution.AppName}

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
