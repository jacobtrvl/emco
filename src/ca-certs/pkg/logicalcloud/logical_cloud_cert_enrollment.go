// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

// CertEnrollmentClient
type CertEnrollmentClient struct {
	db DbInfo
}

// CertEnrollmentManager
type CertEnrollmentManager interface {
	// Delete()
	// Get()
	Instantiate()
	Status()
	Terminate()
	Update()
}

// func (c *CertEnrollmentClient) Delete() {

// }

// func (c *CertEnrollmentClient) Get() {

// }

func (c *CertEnrollmentClient) Instantiate() {

}

func (c *CertEnrollmentClient) Status() {

}

func (c *CertEnrollmentClient) Terminate() {

}

func (c *CertEnrollmentClient) Update() {

}

// NewCertEnrollmentClient
func NewCertEnrollmentClient() *CertEnrollmentClient {
	return &CertEnrollmentClient{
		db: DbInfo{
			storeName: "resources",
			tagMeta:   "data"}}
}
