// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import (
	"reflect"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// CertClient
type CertClient struct {
	db DbInfo
}

// CertKey
type CertKey struct {
	Cert    string `json:"logicalCloudCert"`
	Project string `json:"project"`
}

type CertManager interface {
	// Certificates
	CreateCert(cert certificate.Cert, project string, failIfExists bool) (certificate.Cert, bool, error)
	DeleteCert(cert, project string) error
	GetAllCert(project string) ([]certificate.Cert, error)
	GetCert(cert, project string) (certificate.Cert, error)
}

// CreateCertificates
func (c *CertClient) CreateCert(cert certificate.Cert, project string, failIfExists bool) (certificate.Cert, bool, error) {
	certExists := false
	key := CertKey{
		Cert:    cert.MetaData.Name,
		Project: project}

	if cer, err := c.GetCert(cert.MetaData.Name, project); err == nil &&
		!reflect.DeepEqual(cer, certificate.Cert{}) {
		certExists = true
	}

	if certExists &&
		failIfExists {
		return certificate.Cert{}, certExists, errors.New("Certificate already exists")
	}

	if err := db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagMeta, cert); err != nil {
		return certificate.Cert{}, certExists, err
	}

	return cert, certExists, nil
}

// DeleteCertificates
func (c *CertClient) DeleteCert(cert, project string) error {
	key := CertKey{
		Cert:    cert,
		Project: project}

	return db.DBconn.Remove(c.db.storeName, key)
}

// GetAllCert
func (c *CertClient) GetAllCert(project string) ([]certificate.Cert, error) {
	key := CertKey{
		Project: project}

	values, err := db.DBconn.Find(c.db.storeName, key, c.db.tagMeta)
	if err != nil {
		return []certificate.Cert{}, err
	}

	var certs []certificate.Cert
	for _, value := range values {
		cert := certificate.Cert{}
		if err = db.DBconn.Unmarshal(value, &cert); err != nil {
			return []certificate.Cert{}, err
		}
		certs = append(certs, cert)
	}

	return certs, nil
}

// GetCertificates
func (c *CertClient) GetCert(cert, project string) (certificate.Cert, error) {
	key := CertKey{
		Cert:    cert,
		Project: project}

	value, err := db.DBconn.Find(c.db.storeName, key, c.db.tagMeta)
	if err != nil {
		return certificate.Cert{}, err
	}

	if len(value) == 0 {
		return certificate.Cert{}, errors.New("Certificate not found")
	}

	if value != nil {
		cert := certificate.Cert{}
		if err = db.DBconn.Unmarshal(value[0], &cert); err != nil {
			return certificate.Cert{}, err
		}
		return cert, nil
	}

	return certificate.Cert{}, errors.New("Unknown Error")
}

// NewCertClient
func NewCertClient() *CertClient {
	return &CertClient{
		db: DbInfo{
			storeName: "resources",
			tagMeta:   "data"}}
}

// // Convert the key to string to preserve the underlying structure
// func (k CertKey) String() string {
// 	out, err := json.Marshal(k)
// 	if err != nil {
// 		return ""
// 	}
// 	return string(out)
// }
