// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

var (
	certClient = clusterprovider.NewCaCertClient()
)

var _ = Describe("Create Cert",
	func() {
		BeforeEach(func() {
			populateCertTestData()
		})
		Context("create a caCert for a clusterProvider", func() {
			It("returns the caCert, no error and, the exists flag is false", func() {
				l := len(mockdb.Items)
				mCert := mockCert("new-cert-1")
				c, cExists, err := certClient.CreateCert(mCert, "provider1", true)
				validateError(err, "")
				Expect(c).To(Equal(mCert))
				Expect(cExists).To(Equal(false))
				Expect(len(mockdb.Items)).To(Equal(l + 3)) // cert + enrollment stateInfo + distribution stateInfo
			})
		})
		Context("create a caCert for a clusterProvider that already exists", func() {
			It("returns an error, no caCert and, the exists flag is true", func() {
				l := len(mockdb.Items)
				mCert := mockCert("test-cert-1")
				c, cExists, err := certClient.CreateCert(mCert, "provider1", true)
				Expect(c).To(Equal(module.CaCert{}))
				validateError(err, "Certificate already exists")
				Expect(cExists).To(Equal(true))
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
	},
)

var _ = Describe("Delete Cert",
	func() {
		BeforeEach(func() {
			populateCertTestData()
		})
		Context("delete an existing caCert", func() {
			It("returns no error and delete the entry from the db", func() {
				l := len(mockdb.Items)
				err := certClient.DeleteCert("test-cert-1", "provider1")
				validateError(err, "")
				Expect(len(mockdb.Items)).To(Equal(l - 1))
			})
		})
		Context("delete a nonexisting caCert", func() {
			It("returns an error and no change in the db", func() {
				l := len(mockdb.Items)
				mockdb.Err = errors.New("db Remove resource not found")
				err := certClient.DeleteCert("non-existing-cert", "provider1")
				validateError(err, "db Remove resource not found")
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
	},
)

var _ = Describe("Get All Certs",
	func() {
		BeforeEach(func() {
			populateCertTestData()
		})
		Context("get all the caCert intents", func() {
			It("returns all the caCert intents, no error", func() {
				certs, err := certClient.GetAllCert("provider1")
				validateError(err, "")
				Expect(len(certs)).To(Equal(len(mockdb.Items)))
			})
		})
		Context("get all the caCert intents without creating any", func() {
			It("returns an empty array, no error", func() {
				mockdb.Items = []map[string]map[string][]byte{}
				certs, err := certClient.GetAllCert("provider1")
				validateError(err, "")
				Expect(len(certs)).To(Equal(0))
			})
		})
	},
)

var _ = Describe("Get Cert",
	func() {
		BeforeEach(func() {
			populateCertTestData()
		})
		Context("get an existing caCert", func() {
			It("returns the caCert, no error", func() {
				cert, err := certClient.GetCert("test-cert-1", "provider1")
				validateError(err, "")
				validateCert(cert, mockCert("test-cert-1"))
			})
		})
		Context("get a nonexisting caCert", func() {
			It("returns an error, no caCert", func() {
				cert, err := certClient.GetCert("non-existing-cert", "provider1")
				validateError(err, "Certificate not found")
				validateCert(cert, module.CaCert{})
			})
		})
	},
)

// validateCert
func validateCert(in, out module.CaCert) {
	Expect(in).To(Equal(out))
}

// mockCert
func mockCert(name string) module.CaCert {
	return module.CaCert{
		MetaData: types.Metadata{
			Name:        name,
			Description: "test cert",
			UserData1:   "some user data 1",
			UserData2:   "some user data 2",
		},
	}
}

// populateCertTestData
func populateCertTestData() {
	mockdb.Err = nil
	mockdb.Items = []map[string]map[string][]byte{}
	mockdb.MarshalErr = nil

	// cert 1
	cert := mockCert("test-cert-1")
	cpKey := clusterprovider.CaCertKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: "provider1"}
	_ = mockdb.Insert("resources", cpKey, nil, "data", cert)

	// cert 2
	cert = mockCert("test-cert-2")
	cpKey = clusterprovider.CaCertKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: "provider1"}
	_ = mockdb.Insert("resources", cpKey, nil, "data", cert)

	// cert 3
	cert = mockCert("test-cert-3")
	cpKey = clusterprovider.CaCertKey{
		Cert:            cert.MetaData.Name,
		ClusterProvider: "provider1"}
	_ = mockdb.Insert("resources", cpKey, nil, "data", cert)

}
