// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"reflect"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// CertManager
type CertManager interface {
	CreateCert(cert Cert, failIfExists bool) (Cert, bool, error)
	DeleteCert() error
	GetAllCert() ([]Cert, error)
	GetCert() (Cert, error)
}

// CertClient
type CertClient struct {
	dbInfo db.DbInfo
	dbKey  interface{}
}

// NewCertClients
func NewCertClient(dbKey interface{}) *CertClient {
	return &CertClient{
		dbInfo: db.DbInfo{
			StoreName: "resources",
			TagMeta:   "data",
			TagState:  "stateInfo"},
		dbKey: dbKey}
}

// CreateCert
func (c *CertClient) CreateCert(cert Cert, failIfExists bool) (Cert, bool, error) {
	certExists := false

	if cer, err := c.GetCert(); err == nil &&
		!reflect.DeepEqual(cer, Cert{}) {
		certExists = true
	}

	if certExists &&
		failIfExists {
		return Cert{}, certExists, errors.New("Certificate already exists")
	}

	if err := db.DBconn.Insert(c.dbInfo.StoreName, c.dbKey, nil, c.dbInfo.TagMeta, cert); err != nil {
		return Cert{}, certExists, err
	}

	return cert, certExists, nil
}

// DeleteCert
func (c *CertClient) DeleteCert() error {
	return db.DBconn.Remove(c.dbInfo.StoreName, c.dbKey)
}

// GetAllCertificate
func (c *CertClient) GetAllCert() ([]Cert, error) {
	values, err := db.DBconn.Find(c.dbInfo.StoreName, c.dbKey, c.dbInfo.TagMeta)
	if err != nil {
		return []Cert{}, err
	}

	var certs []Cert
	for _, value := range values {
		cert := Cert{}
		if err = db.DBconn.Unmarshal(value, &cert); err != nil {
			return []Cert{}, err
		}
		certs = append(certs, cert)
	}

	return certs, nil
}

// GetCert
func (c *CertClient) GetCert() (Cert, error) {
	value, err := db.DBconn.Find(c.dbInfo.StoreName, c.dbKey, c.dbInfo.TagMeta)
	if err != nil {
		return Cert{}, err
	}

	if len(value) == 0 {
		return Cert{}, errors.New("Certificate not found")
	}

	if value != nil {
		cert := Cert{}
		if err = db.DBconn.Unmarshal(value[0], &cert); err != nil {
			return Cert{}, err
		}
		return cert, nil
	}

	return Cert{}, errors.New("Unknown Error")
}

// // Convert the key to string to preserve the underlying structure
// func (k CertKey) String() string {
// 	out, err := json.Marshal(k)
// 	if err != nil {
// 		return ""
// 	}
// 	return string(out)
// }
