// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module_test

// import (
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// 	"github.com/pkg/errors"

// 	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/logicalcloud"
// 	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"

// 	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/clusterprovider"
// 	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
// )

// var (
// 	mockdb *db.NewMockDB
// )

// var _ = Describe("Create Cert",
// 	func() {
// 		BeforeEach(func() {
// 			populateCertTestData()
// 		})
// 		Context("create a cert for a cluster provider", func() {
// 			It("returns the cert, no error and, the exists flag is false", func() {
// 				l := len(mockdb.Items)
// 				mCert := mockCert("new-cert-1")
// 				key := clusterprovider.CertKey{
// 					Cert:            "cert1",
// 					ClusterProvider: "provider1"}
// 				client := module.NewCertClient(key)
// 				cExists, err := client.CreateCert(mCert, true)
// 				validateError(err, "")
// 				Expect(cExists).To(Equal(false))
// 				Expect(len(mockdb.Items)).To(Equal(l + 1))
// 			})
// 		})
// 		Context("create a cert for a cluster provider that already exists", func() {
// 			It("returns an error, no cert and, the exists flag is true", func() {
// 				l := len(mockdb.Items)
// 				mCert := mockCert("test-cert-1")
// 				key := clusterprovider.CertKey{
// 					Cert:            "cert1",
// 					ClusterProvider: "provider1"}
// 				client := module.NewCertClient(key)
// 				cExists, err := client.CreateCert(mCert, true)
// 				validateError(err, "Cert already exists")
// 				Expect(cExists).To(Equal(true))
// 				Expect(len(mockdb.Items)).To(Equal(l))
// 			})
// 		})
// 		Context("create a cert for a logical cloud", func() {
// 			It("returns the cert, no error and, the exists flag is false", func() {
// 				l := len(mockdb.Items)
// 				mCert := mockCert("new-cert-1")
// 				key := logicalcloud.CertKey{
// 					Cert:    "cert1",
// 					Project: "proj1"}
// 				client := module.NewCertClient(key)
// 				cExists, err := client.CreateCert(mCert, true)
// 				validateError(err, "")
// 				Expect(cExists).To(Equal(false))
// 				Expect(len(mockdb.Items)).To(Equal(l + 1))
// 			})
// 		})
// 		Context("create a cert for a logical cloud that already exists", func() {
// 			It("returns an error, no cert and, the exists flag is true", func() {
// 				l := len(mockdb.Items)
// 				mCert := mockCert("test-cert-1")
// 				key := logicalcloud.CertKey{
// 					Cert:    "cert1",
// 					Project: "proj1"}
// 				client := module.NewCertClient(key)
// 				cExists, err := client.CreateCert(mCert, true)
// 				validateError(err, "Cert already exists")
// 				Expect(cExists).To(Equal(true))
// 				Expect(len(mockdb.Items)).To(Equal(l))
// 			})
// 		})
// 	},
// )

// var _ = Describe("Delete Cert",
// 	func() {
// 		BeforeEach(func() {
// 			populateCertTestData()
// 		})
// 		Context("delete an existing cert", func() {
// 			It("returns no error and delete the entry from the db", func() {
// 				l := len(mockdb.Items)
// 				key := clusterprovider.CertKey{
// 					Cert:            "test-cert-1",
// 					ClusterProvider: "provider1"}
// 				client := module.NewCertClient(key)
// 				err := client.DeleteCert()
// 				validateError(err, "")
// 				Expect(len(mockdb.Items)).To(Equal(l - 1))
// 			})
// 		})
// 		Context("delete a nonexisting cert", func() {
// 			It("returns an error and no change in the db", func() {
// 				l := len(mockdb.Items)
// 				mockdb.Err = errors.New("db Remove resource not found")
// 				key := clusterprovider.CertKey{
// 					Cert:            "non-existing-cert",
// 					ClusterProvider: "provider1"}
// 				client := module.NewCertClient(key)
// 				err := client.DeleteCert()
// 				validateError(err, "db Remove resource not found")
// 				Expect(len(mockdb.Items)).To(Equal(l))
// 			})
// 		})
// 		Context("delete an existing cert", func() {
// 			It("returns no error and delete the entry from the db", func() {
// 				l := len(mockdb.Items)
// 				key := logicalcloud.CertKey{
// 					Cert:    "test-cert-1",
// 					Project: "proj1"}
// 				client := module.NewCertClient(key)
// 				err := client.DeleteCert()
// 				validateError(err, "")
// 				Expect(len(mockdb.Items)).To(Equal(l - 1))
// 			})
// 		})
// 		Context("delete a nonexisting cert", func() {
// 			It("returns an error and no change in the db", func() {
// 				l := len(mockdb.Items)
// 				mockdb.Err = errors.New("db Remove resource not found")
// 				key := logicalcloud.CertKey{
// 					Cert:    "non-existing-cert",
// 					Project: "proj1"}
// 				client := module.NewCertClient(key)
// 				err := client.DeleteCert()
// 				validateError(err, "db Remove resource not found")
// 				Expect(len(mockdb.Items)).To(Equal(l))
// 			})
// 		})
// 	},
// )

// var _ = Describe("Get All GenericK8sIntents",
// 	func() {
// 		BeforeEach(func() {
// 			populateCertTestData()
// 		})
// 		Context("get all the intents", func() {
// 			It("returns all the intents, no error", func() {
// 				key := clusterprovider.CertKey{
// 					ClusterProvider: "provider1"}
// 				client := module.NewCertClient(key)
// 				certs, err := client.GetAllCert()
// 				validateError(err, "")
// 				Expect(len(certs)).To(Equal(len(mockdb.Items)))
// 			})
// 		})
// 		Context("get all the intents without creating any", func() {
// 			It("returns an empty array, no error", func() {
// 				mockdb.Items = []map[string]map[string][]byte{}
// 				key := clusterprovider.CertKey{
// 					ClusterProvider: "provider1"}
// 				client := module.NewCertClient(key)
// 				certs, err := client.GetAllCert()
// 				validateError(err, "")
// 				Expect(len(certs)).To(Equal(0))
// 			})
// 		})

// 		Context("get all the intents", func() {
// 			It("returns all the intents, no error", func() {
// 				key := logicalcloud.CertKey{
// 					Project: "proj1"}
// 				client := module.NewCertClient(key)
// 				certs, err := client.GetAllCert()
// 				validateError(err, "")
// 				Expect(len(certs)).To(Equal(len(mockdb.Items)))
// 			})
// 		})
// 		Context("get all the intents without creating any", func() {
// 			It("returns an empty array, no error", func() {
// 				mockdb.Items = []map[string]map[string][]byte{}
// 				key := logicalcloud.CertKey{
// 					Project: "proj1"}
// 				client := module.NewCertClient(key)
// 				certs, err := client.GetAllCert()
// 				validateError(err, "")
// 				Expect(len(certs)).To(Equal(0))
// 			})
// 		})
// 	},
// )

// var _ = Describe("Get Cert",
// 	func() {
// 		BeforeEach(func() {
// 			populateCertTestData()
// 		})
// 		Context("get an existing cert", func() {
// 			It("returns the cert, no error", func() {
// 				key := clusterprovider.CertKey{
// 					Cert:            "test-cert-1",
// 					ClusterProvider: "provider1"}
// 				client := module.NewCertClient(key)
// 				cert, err := client.GetCert()
// 				validateError(err, "")
// 				validateCert(cert, mockCert("test-cert-1"))
// 			})
// 		})
// 		Context("get a nonexisting cert", func() {
// 			It("returns an error, no cert", func() {
// 				key := clusterprovider.CertKey{
// 					Cert:            "non-existing-cert",
// 					ClusterProvider: "provider1"}
// 				client := module.NewCertClient(key)
// 				cert, err := client.GetCert()
// 				validateError(err, "Cert not found")
// 				validateCert(cert, module.Cert{})
// 			})
// 		})
// 	},
// )

// // validateCert
// func validateCert(in, out module.Cert) {
// 	Expect(in).To(Equal(out))
// }

// // mockCert
// func mockCert(name string) module.Cert {
// 	return module.Cert{
// 		MetaData: module.MetaData{
// 			Name:        name,
// 			Description: "test cert",
// 			UserData1:   "some user data 1",
// 			UserData2:   "some user data 2",
// 		},
// 	}
// }

// // populateCertTestData
// func populateCertTestData() {
// 	mockdb.Err = nil
// 	mockdb.Items = []map[string]map[string][]byte{}
// 	mockdb.MarshalErr = nil

// 	// cert 1
// 	cert := mockCert("test-cert-1")
// 	cpKey := clusterprovider.CertKey{
// 		Cert:            cert.MetaData.Name,
// 		ClusterProvider: "provider1"}
// 	_ = mockdb.Insert("resources", cpKey, nil, "data", cert)

// 	// cert 2
// 	cert = mockCert("test-cert-2")
// 	cpKey = clusterprovider.CertKey{
// 		Cert:            cert.MetaData.Name,
// 		ClusterProvider: "provider1"}
// 	_ = mockdb.Insert("resources", cpKey, nil, "data", cert)

// 	// cert 3
// 	cert = mockCert("test-cert-3")
// 	cpKey = clusterprovider.CertKey{
// 		Cert:            cert.MetaData.Name,
// 		ClusterProvider: "provider1"}
// 	_ = mockdb.Insert("resources", cpKey, nil, "data", cert)

// 	// cert 4
// 	cert = mockCert("test-cert-4")
// 	lcKey := logicalcloud.CertKey{
// 		Cert:    cert.MetaData.Name,
// 		Project: "proj1"}
// 	_ = mockdb.Insert("resources", lcKey, nil, "data", cert)

// 	// cert 5
// 	cert = mockCert("test-cert-5")
// 	lcKey = logicalcloud.CertKey{
// 		Cert:    cert.MetaData.Name,
// 		Project: "proj1"}
// 	_ = mockdb.Insert("resources", lcKey, nil, "data", cert)

// 	// cert 6
// 	cert = mockCert("test-cert-6")
// 	lcKey = logicalcloud.CertKey{
// 		Cert:    cert.MetaData.Name,
// 		Project: "proj1"}
// 	_ = mockdb.Insert("resources", lcKey, nil, "data", cert)

// }

// func validateError(err error, message string) {
// 	if len(message) == 0 {
// 		Expect(err).NotTo(HaveOccurred())
// 		Expect(err).To(BeNil())
// 		return
// 	}
// 	Expect(err.Error()).To(ContainSubstring(message))
// }
