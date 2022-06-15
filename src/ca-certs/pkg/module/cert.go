// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"reflect"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
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

func (c *CertClient) UpdateCert(cert Cert) error {
	return db.DBconn.Insert(c.dbInfo.StoreName, c.dbKey, nil, c.dbInfo.TagMeta, cert)
}

// VerifyStateBeforeDelete
func (c *CertClient) VerifyStateBeforeDelete(cert, lifecycle string) error {
	sc := NewStateClient(c.dbKey)
	stateInfo, err := sc.Get()
	if err != nil {
		return err
	}

	cState, err := state.GetCurrentStateFromStateInfo(stateInfo)
	if err != nil {
		return err
	}

	if cState == state.StateEnum.Instantiated ||
		cState == state.StateEnum.InstantiateStopped {
		return errors.Errorf(
			"%s must be terminated for CaCert %s before it can be deleted", lifecycle, cert)
	}

	if cState == state.StateEnum.Terminated ||
		cState == state.StateEnum.TerminateStopped {
		// verify that the appcontext has completed terminating
		ctxID := state.GetLastContextIdFromStateInfo(stateInfo)

		acStatus, err := state.GetAppContextStatus(ctxID)
		if err == nil &&
			!(acStatus.Status == appcontext.AppContextStatusEnum.Terminated ||
				acStatus.Status == appcontext.AppContextStatusEnum.TerminateFailed) {
			return errors.Errorf("%s termination has not completed for CaCert %s", lifecycle, cert)
		}

		for _, id := range state.GetContextIdsFromStateInfo(stateInfo) {
			context, err := state.GetAppContextFromId(id)
			if err != nil {
				return err
			}
			err = context.DeleteCompositeApp()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// VerifyStateBeforeUpdate
func (c *CertClient) VerifyStateBeforeUpdate(cert, lifecycle string) error {
	sc := NewStateClient(c.dbKey)
	stateInfo, err := sc.Get()
	if err != nil {
		return err
	}

	cState, err := state.GetCurrentStateFromStateInfo(stateInfo)
	if err != nil {
		return err
	}

	// TODO: What if, the state is Terminated?
	if cState != state.StateEnum.Created {
		return errors.Errorf(
			"failed to update the CaCert. %s for the CaCert %s is in %s state", lifecycle, cert, cState)
	}

	return nil
}

// // Convert the key to string to preserve the underlying structure
// func (k CertKey) String() string {
// 	out, err := json.Marshal(k)
// 	if err != nil {
// 		return ""
// 	}
// 	return string(out)
// }
