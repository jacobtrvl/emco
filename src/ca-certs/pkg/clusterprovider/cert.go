// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

import (
	"reflect"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// CertManager
type CertManager interface {
	// Certificates
	CreateCert(cert certificate.Cert, clusterProvider string, failIfExists bool) (certificate.Cert, bool, error)
	DeleteCert(cert, clusterProvider string) error
	GetAllCert(clusterProvider string) ([]certificate.Cert, error)
	GetCert(cert, clusterProvider string) (certificate.Cert, error)
}

// CertKey
type CertKey struct {
	Cert            string `json:"clusterProviderCert"`
	ClusterProvider string `json:"clusterProvider"`
}

// CertClient
type CertClient struct {
	db DbInfo
}

// NewCertClient
func NewCertClient() *CertClient {
	return &CertClient{
		db: DbInfo{
			storeName: "resources",
			tagMeta:   "data",
			tagState:  "stateInfo"}}
}

// CreateCert
func (c *CertClient) CreateCert(cert certificate.Cert, clusterProvider string, failIfExists bool) (certificate.Cert, bool, error) {
	certExists := false
	key := CertKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: clusterProvider}

	if cer, err := c.GetCert(cert.MetaData.Name, clusterProvider); err == nil &&
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

	k := StateKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: clusterProvider,
		AppName:         enrollmentApp}

	// create a state object for the cert resource
	if err := common.NewStateClient().CreateState(k, ""); err != nil {
		return certificate.Cert{}, certExists, err
	}

	//  TODo - revisit this logic
	k = StateKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: clusterProvider,
		AppName:         distributionApp}

	// create a state object for the cert resource
	if err := common.NewStateClient().CreateState(k, ""); err != nil {
		return certificate.Cert{}, certExists, err
	}

	return cert, certExists, nil
}

// DeleteCert
func (c *CertClient) DeleteCert(cert, clusterProvider string) error {
	key := CertKey{
		Cert:            cert,
		ClusterProvider: clusterProvider}

	return db.DBconn.Remove(c.db.storeName, key)

}

// GetAllCertificate
func (c *CertClient) GetAllCert(clusterProvider string) ([]certificate.Cert, error) {
	key := CertKey{
		ClusterProvider: clusterProvider}

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

// GetCert
func (c *CertClient) GetCert(cert, clusterProvider string) (certificate.Cert, error) {
	key := CertKey{
		Cert:            cert,
		ClusterProvider: clusterProvider}

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

// // Convert the key to string to preserve the underlying structure
// func (k CertKey) String() string {
// 	out, err := json.Marshal(k)
// 	if err != nil {
// 		return ""
// 	}
// 	return string(out)
// }
