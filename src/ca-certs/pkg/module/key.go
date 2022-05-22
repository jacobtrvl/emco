// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// KeyManager
type KeyManager interface {
	// Key
	Save(pk string) error
	Delete(key interface{}) error
	Get(key interface{}) (Cert, error)
}

type Key struct {
	Name string
	Val  string
}

type DBKey struct {
	Cert            string `json:"cert"`
	Cluster         string `json:"cluster"`
	ClusterProvider string `json:"clusterProvider"`
	ContextID       string `json:"contextID"`
}

// KeyClient
type KeyClient struct {
	dbInfo db.DbInfo
	dbKey  interface{}
}

// NewCertClient
func NewKeyClient(dbKey interface{}) *KeyClient {
	return &KeyClient{
		dbInfo: db.DbInfo{
			StoreName: "resources",
			TagMeta:   "key"},
		dbKey: dbKey}
}

// Save
func (c *KeyClient) Save(pk Key) error {
	return db.DBconn.Insert(c.dbInfo.StoreName, c.dbKey, nil, c.dbInfo.TagMeta, pk)
}

// Delete
func (c *KeyClient) Delete() error {
	return db.DBconn.Remove(c.dbInfo.StoreName, c.dbKey)
}

// Get
func (c *KeyClient) Get() (Key, error) {
	value, err := db.DBconn.Find(c.dbInfo.StoreName, c.dbKey, c.dbInfo.TagMeta)
	if err != nil {
		return Key{}, err
	}

	if len(value) == 0 {
		return Key{}, errors.New("Key not found")
	}

	if value != nil {
		key := Key{}
		if err = db.DBconn.Unmarshal(value[0], &key); err != nil {
			return Key{}, err
		}
		return key, nil
	}

	return Key{}, errors.New("Unknown Error")
}
