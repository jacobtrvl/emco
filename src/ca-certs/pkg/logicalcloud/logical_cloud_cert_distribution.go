// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

// CertDistributionClient
type CertDistributionClient struct {
	db DbInfo
}

// CertDistributionManager
type CertDistributionManager interface {
	// Delete()
	// Get()
	Instantiate()
	Status()
	Terminate()
	Update()
}

// func (c *CertDistributionClient) Delete() {

// }

// func (c *CertDistributionClient) Get() {

// }

func (c *CertDistributionClient) Instantiate() {

}

func (c *CertDistributionClient) Status() {

}

func (c *CertDistributionClient) Terminate() {

}

func (c *CertDistributionClient) Update() {

}

// NewCertDistributionClient
func NewCertDistributionClient() *CertDistributionClient {
	return &CertDistributionClient{
		db: DbInfo{
			storeName: "resources",
			tagMeta:   "data"}}
}
