//=======================================================================
// Copyright (c) 2022 Aarna Networks, Inc.
// All rights reserved.
// ======================================================================
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//           http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ========================================================================

package etcdhelper

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.etcd.io/etcd/clientv3"
)

// EtcdConfig Configuration values needed for Etcd Client
type EtcdConfig struct {
	Endpoint string
	CertFile string
	KeyFile  string
	CAFile   string
}

// EtcdClient for Etcd
type EtcdClient struct {
	cli      *clientv3.Client
	endpoint string
}

type ContextDb interface {
	HealthCheck() error
	Put(key string, value interface{}) error
	Delete(key string) error
	DeleteAll(key string) error
	Get(key string) ([]byte, error)
	GetAllKeys(path string) ([]string, error)
	PutWithCheck(key string, value interface{}) error
}

// Etcd For Mocking purposes
type Etcd interface {
	Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error)
	Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error)
}

var getEtcd = func(e *EtcdClient) Etcd {
	return e.cli
}

// NewEtcdClient function initializes Etcd client
func NewEtcdClient(store *clientv3.Client, c EtcdConfig) (ContextDb, error) {
	var endpoint string
	if store == nil {
		endpoint = "http://" + c.Endpoint + ":2379"

		etcdClient := clientv3.Config{
			Endpoints:   []string{endpoint},
			DialTimeout: 5 * time.Second,
		}
		if len(os.Getenv("CONTEXTDB_EMCO_USERNAME")) > 0 && len(os.Getenv("CONTEXTDB_EMCO_PASSWORD")) > 0 {
			etcdClient.Username = os.Getenv("CONTEXTDB_EMCO_USERNAME")
			etcdClient.Password = os.Getenv("CONTEXTDB_EMCO_PASSWORD")
		}
		var err error
		store, err = clientv3.New(etcdClient)
		if err != nil {
			return nil, errors.Errorf("Error creating etcd client: %s", err.Error())
		}
	}

	return &EtcdClient{
		cli:      store,
		endpoint: endpoint,
	}, nil
}

// Put values in Etcd DB
func (e *EtcdClient) Put(key string, value interface{}) error {
	cli := getEtcd(e)
	if cli == nil {
		return errors.Errorf("Etcd Client not initialized")
	}
	if key == "" {
		return errors.Errorf("Key is null")
	}
	if value == nil {
		return errors.Errorf("Value is nil")
	}
	v, err := json.Marshal(value)
	if err != nil {
		return errors.Errorf("Json Marshal error: %s", err.Error())
	}
	_, err = cli.Put(context.Background(), key, string(v))
	if err != nil {
		return errors.Errorf("Error creating etcd entry: %s", err.Error())
	}
	return nil
}

// Get values from Etcd DB and decodes from json
func (e *EtcdClient) Get(key string) ([]byte, error) {
	cli := getEtcd(e)
	if cli == nil {
		return nil, errors.Errorf("Etcd Client not initialized")
	}
	if key == "" {
		return nil, errors.Errorf("Key is null")
	}
	getResp, err := cli.Get(context.Background(), key)
	if err != nil {
		return nil, errors.Errorf("Error getting etcd entry: %s", err.Error())
	}
	if getResp.Count == 0 {
		return nil, errors.Errorf("Key doesn't exist")
	}
	return getResp.Kvs[0].Value, nil
}

// GetAllKeys values from Etcd DB
func (e *EtcdClient) GetAllKeys(key string) ([]string, error) {
	cli := getEtcd(e)
	if cli == nil {
		return nil, errors.Errorf("Etcd Client not initialized")
	}
	getResp, err := cli.Get(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Errorf("Error getting etcd entry: %s", err.Error())
	}
	if getResp.Count == 0 {
		return nil, errors.Errorf("Key doesn't exist")
	}
	var keys []string
	for _, ev := range getResp.Kvs {
		keys = append(keys, string(ev.Key))
	}
	return keys, nil
}

// DeleteAll keys from Etcd DB
func (e *EtcdClient) DeleteAll(key string) error {
	cli := getEtcd(e)
	if cli == nil {
		return errors.Errorf("Etcd Client not initialized")
	}
	_, err := cli.Delete(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		return errors.Errorf("Delete failed etcd entry: %s", err.Error())
	}
	return nil
}

// Delete values from Etcd DB
func (e *EtcdClient) Delete(key string) error {
	cli := getEtcd(e)
	if cli == nil {
		return errors.Errorf("Etcd Client not initialized")
	}
	_, err := cli.Delete(context.Background(), key)
	if err != nil {
		return errors.Errorf("Delete failed etcd entry: %s", err.Error())
	}
	return nil
}

// HealthCheck for checking health of the etcd cluster
func (e *EtcdClient) HealthCheck() error {
	return nil
}

// PutWithCheck Put values in Etcd DB and check if already present
func (e *EtcdClient) PutWithCheck(key string, value interface{}) error {
	cli := getEtcd(e)
	if cli == nil {
		return errors.Errorf("Etcd Client not initialized")
	}
	if key == "" {
		return errors.Errorf("Key is null")
	}
	if value == nil {
		return errors.Errorf("Value is nil")
	}
	v, err := json.Marshal(value)
	if err != nil {
		return errors.Errorf("Json Marshal error: %s", err.Error())
	}
	var opts []clientv3.OpOption
	opts = append(opts, clientv3.WithPrevKV())
	resp, err := cli.Put(context.Background(), key, string(v), opts...)
	if err != nil {
		return errors.Errorf("Error creating etcd entry: %s", err.Error())
	}
	// Check if this key was already present
	if resp.PrevKv != nil && len(resp.PrevKv.Key) > 0 {
		return errors.Errorf("Key exists %v", string(resp.PrevKv.Key))
	}
	return nil
}
