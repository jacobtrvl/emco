// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// KeyManager exposes all the private key functionalities
type KeyManager interface {
	Save(pk string) error
	Delete(key interface{}) error
	Get(key interface{}) (CaCert, error)
}

// DBKey represents the resources associated with a private key
type DBKey struct {
	Cert            string `json:"caCert"`
	Cluster         string `json:"caCertCluster"`
	ClusterProvider string `json:"caCertClusterProvider"`
	ContextID       string `json:"caCertContextID"`
}

// KeyClient holds the client properties
type KeyClient struct {
	dbInfo db.DbInfo
	dbKey  interface{}
}

// NewKeyClient returns an instance of the KeyClient which implements the Manager
func NewKeyClient(dbKey interface{}) *KeyClient {
	return &KeyClient{
		dbInfo: db.DbInfo{
			StoreName: "resources",
			TagMeta:   "key"},
		dbKey: dbKey}
}

// Save key in the mongo
func (c *KeyClient) Save(pk Key) error {
	return db.DBconn.Insert(c.dbInfo.StoreName, c.dbKey, nil, c.dbInfo.TagMeta, pk)
}

// Delete key from mongo
func (c *KeyClient) Delete() error {
	return db.DBconn.Remove(c.dbInfo.StoreName, c.dbKey)
}

// Get key from mongo
func (c *KeyClient) Get() (Key, error) {
	value, err := db.DBconn.Find(c.dbInfo.StoreName, c.dbKey, c.dbInfo.TagMeta)
	if err != nil {
		return Key{}, err
	}

	if len(value) == 0 {
		return Key{}, emcoerror.NewEmcoError(
			KeyNotFound,
			emcoerror.NotFound,
		)
	}

	if value != nil {
		key := Key{}
		if err = db.DBconn.Unmarshal(value[0], &key); err != nil {
			return Key{}, err
		}
		return key, nil
	}

	return Key{}, emcoerror.NewEmcoError(
		emcoerror.UnknownErrorMessage,
		emcoerror.Unknown,
	)
}
