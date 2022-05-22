// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module_test

// import (
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// 	"github.com/pkg/errors"

// 	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/clusterprovider"
// 	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/logicalcloud"
// 	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
// 	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
// )

// var (
// 	mockdb *db.NewMockDB
// )

// var _ = Describe("Create ClusterGroup",
// 	func() {
// 		BeforeEach(func() {
// 			populateClusterGroupTestData()
// 		})
// 		Context("create a cluster for a cluster provider", func() {
// 			It("returns the cluster, no error and, the exists flag is false", func() {
// 				l := len(mockdb.Items)
// 				mClusterGroup := mockClusterGroup("new-cluster-1")
// 				key := clusterprovider.ClusterGroupKey{
// 					Cert:            "new-cluster-1",
// 					ClusterGroup:    "edge1-clusters",
// 					ClusterProvider: "provider1"}
// 				client := module.NewClusterClient(key)
// 				cExists, err := client.CreateClusterGroup(mClusterGroup, true)
// 				validateError(err, "")
// 				Expect(cExists).To(Equal(false))
// 				Expect(len(mockdb.Items)).To(Equal(l + 1))
// 			})
// 		})
// 		Context("create a cluster for a cluster provider that already exists", func() {
// 			It("returns an error, no cluster and, the exists flag is true", func() {
// 				l := len(mockdb.Items)
// 				mClusterGroup := mockClusterGroup("test-cluster-1")
// 				key := clusterprovider.ClusterGroupKey{
// 					ClusterGroup:    "cluster1",
// 					ClusterProvider: "provider1"}
// 				client := module.NewClusterClient(key)
// 				cExists, err := client.CreateClusterGroup(mClusterGroup, true)
// 				validateError(err, "ClusterGroup already exists")
// 				Expect(cExists).To(Equal(true))
// 				Expect(len(mockdb.Items)).To(Equal(l))
// 			})
// 		})
// 		Context("create a cluster for a logical cloud", func() {
// 			It("returns the cluster, no error and, the exists flag is false", func() {
// 				l := len(mockdb.Items)
// 				mClusterGroup := mockClusterGroup("new-cluster-1")
// 				key := logicalcloud.ClusterGroupKey{
// 					ClusterGroup: "cluster1",
// 					Project:      "proj1"}
// 				client := module.NewClusterClient(key)
// 				cExists, err := client.CreateClusterGroup(mClusterGroup, true)
// 				validateError(err, "")
// 				Expect(cExists).To(Equal(false))
// 				Expect(len(mockdb.Items)).To(Equal(l + 1))
// 			})
// 		})
// 		Context("create a cluster for a logical cloud that already exists", func() {
// 			It("returns an error, no cluster and, the exists flag is true", func() {
// 				l := len(mockdb.Items)
// 				mClusterGroup := mockClusterGroup("test-cluster-1")
// 				key := logicalcloud.ClusterGroupKey{
// 					ClusterGroup: "cluster1",
// 					Project:      "proj1"}
// 				client := module.NewClusterClient(key)
// 				cExists, err := client.CreateClusterGroup(mClusterGroup, true)
// 				validateError(err, "ClusterGroup already exists")
// 				Expect(cExists).To(Equal(true))
// 				Expect(len(mockdb.Items)).To(Equal(l))
// 			})
// 		})
// 	},
// )

// var _ = Describe("Delete ClusterGroup",
// 	func() {
// 		BeforeEach(func() {
// 			populateClusterGroupTestData()
// 		})
// 		Context("delete an existing cluster", func() {
// 			It("returns no error and delete the entry from the db", func() {
// 				l := len(mockdb.Items)
// 				key := clusterprovider.ClusterGroupKey{
// 					ClusterGroup:    "test-cluster-1",
// 					ClusterProvider: "provider1"}
// 				client := module.NewClusterClient(key)
// 				err := client.DeleteClusterGroup()
// 				validateError(err, "")
// 				Expect(len(mockdb.Items)).To(Equal(l - 1))
// 			})
// 		})
// 		Context("delete a nonexisting cluster", func() {
// 			It("returns an error and no change in the db", func() {
// 				l := len(mockdb.Items)
// 				mockdb.Err = errors.New("db Remove resource not found")
// 				key := clusterprovider.ClusterGroupKey{
// 					ClusterGroup:    "non-existing-cluster",
// 					ClusterProvider: "provider1"}
// 				client := module.NewClusterClient(key)
// 				err := client.DeleteClusterGroup()
// 				validateError(err, "db Remove resource not found")
// 				Expect(len(mockdb.Items)).To(Equal(l))
// 			})
// 		})
// 		Context("delete an existing cluster", func() {
// 			It("returns no error and delete the entry from the db", func() {
// 				l := len(mockdb.Items)
// 				key := logicalcloud.ClusterGroupKey{
// 					ClusterGroup: "test-cluster-1",
// 					Project:      "proj1"}
// 				client := module.NewClusterClient(key)
// 				err := client.DeleteClusterGroup()
// 				validateError(err, "")
// 				Expect(len(mockdb.Items)).To(Equal(l - 1))
// 			})
// 		})
// 		Context("delete a nonexisting cluster", func() {
// 			It("returns an error and no change in the db", func() {
// 				l := len(mockdb.Items)
// 				mockdb.Err = errors.New("db Remove resource not found")
// 				key := logicalcloud.ClusterGroupKey{
// 					ClusterGroup: "non-existing-cluster",
// 					Project:      "proj1"}
// 				client := module.NewClusterClient(key)
// 				err := client.DeleteClusterGroup()
// 				validateError(err, "db Remove resource not found")
// 				Expect(len(mockdb.Items)).To(Equal(l))
// 			})
// 		})
// 	},
// )

// var _ = Describe("Get All GenericK8sIntents",
// 	func() {
// 		BeforeEach(func() {
// 			populateClusterGroupTestData()
// 		})
// 		Context("get all the intents", func() {
// 			It("returns all the intents, no error", func() {
// 				key := clusterprovider.ClusterGroupKey{
// 					ClusterProvider: "provider1"}
// 				client := module.NewClusterClient(key)
// 				clusters, err := client.GetAllClusterGroups()
// 				validateError(err, "")
// 				Expect(len(clusters)).To(Equal(len(mockdb.Items)))
// 			})
// 		})
// 		Context("get all the intents without creating any", func() {
// 			It("returns an empty array, no error", func() {
// 				mockdb.Items = []map[string]map[string][]byte{}
// 				key := clusterprovider.ClusterGroupKey{
// 					ClusterProvider: "provider1"}
// 				client := module.NewClusterClient(key)
// 				clusters, err := client.GetAllClusterGroups()
// 				validateError(err, "")
// 				Expect(len(clusters)).To(Equal(0))
// 			})
// 		})

// 		Context("get all the intents", func() {
// 			It("returns all the intents, no error", func() {
// 				key := logicalcloud.ClusterGroupKey{
// 					Project: "proj1"}
// 				client := module.NewClusterClient(key)
// 				clusters, err := client.GetAllClusterGroups()
// 				validateError(err, "")
// 				Expect(len(clusters)).To(Equal(len(mockdb.Items)))
// 			})
// 		})
// 		Context("get all the intents without creating any", func() {
// 			It("returns an empty array, no error", func() {
// 				mockdb.Items = []map[string]map[string][]byte{}
// 				key := logicalcloud.ClusterGroupKey{
// 					Project: "proj1"}
// 				client := module.NewClusterClient(key)
// 				clusters, err := client.GetAllClusterGroups()
// 				validateError(err, "")
// 				Expect(len(clusters)).To(Equal(0))
// 			})
// 		})
// 	},
// )

// var _ = Describe("Get ClusterGroup",
// 	func() {
// 		BeforeEach(func() {
// 			populateClusterGroupTestData()
// 		})
// 		Context("get an existing cluster", func() {
// 			It("returns the cluster, no error", func() {
// 				key := clusterprovider.ClusterGroupKey{
// 					ClusterGroup:    "test-cluster-1",
// 					ClusterProvider: "provider1"}
// 				client := module.NewClusterClient(key)
// 				cluster, err := client.GetClusterGroup()
// 				validateError(err, "")
// 				validateClusterGroup(cluster, mockClusterGroup("test-cluster-1"))
// 			})
// 		})
// 		Context("get a nonexisting cluster", func() {
// 			It("returns an error, no cluster", func() {
// 				key := clusterprovider.ClusterGroupKey{
// 					ClusterGroup:    "non-existing-cluster",
// 					ClusterProvider: "provider1"}
// 				client := module.NewClusterClient(key)
// 				cluster, err := client.GetClusterGroup()
// 				validateError(err, "ClusterGroup not found")
// 				validateClusterGroup(cluster, module.ClusterGroup{})
// 			})
// 		})
// 	},
// )

// // validateClusterGroup
// func validateClusterGroup(in, out module.ClusterGroup) {
// 	Expect(in).To(Equal(out))
// }

// // mockClusterGroup
// func mockClusterGroup(name string) module.ClusterGroup {
// 	return module.ClusterGroup{
// 		MetaData: module.MetaData{
// 			Name:        name,
// 			Description: "test cluster",
// 			UserData1:   "some user data 1",
// 			UserData2:   "some user data 2",
// 		},
// 	}
// }

// // populateClusterGroupTestData
// func populateClusterGroupTestData() {
// 	mockdb.Err = nil
// 	mockdb.Items = []map[string]map[string][]byte{}
// 	mockdb.MarshalErr = nil

// 	// cluster 1
// 	cluster := mockClusterGroup("test-cluster-1")
// 	cpKey := clusterprovider.ClusterGroupKey{
// 		ClusterGroup:    "group1",
// 		ClusterProvider: "provider1"}
// 	_ = mockdb.Insert("resources", cpKey, nil, "data", cluster)

// 	// cluster 2
// 	cluster = mockClusterGroup("test-cluster-2")
// 	cpKey = clusterprovider.ClusterGroupKey{
// 		ClusterGroup:    "group1",
// 		ClusterProvider: "provider1"}
// 	_ = mockdb.Insert("resources", cpKey, nil, "data", cluster)

// 	// cluster 3
// 	cluster = mockClusterGroup("test-cluster-3")
// 	cpKey = clusterprovider.ClusterGroupKey{
// 		ClusterGroup:    "group1",
// 		ClusterProvider: "provider1"}
// 	_ = mockdb.Insert("resources", cpKey, nil, "data", cluster)

// 	// cluster 4
// 	cluster = mockClusterGroup("test-cluster-4")
// 	lcKey := logicalcloud.ClusterGroupKey{
// 		ClusterGroup: "group1",
// 		Project:      "proj1"}
// 	_ = mockdb.Insert("resources", lcKey, nil, "data", cluster)

// 	// cluster 5
// 	cluster = mockClusterGroup("test-cluster-5")
// 	lcKey = logicalcloud.ClusterGroupKey{
// 		ClusterGroup: "group1",
// 		Project:      "proj1"}
// 	_ = mockdb.Insert("resources", lcKey, nil, "data", cluster)

// 	// cluster 6
// 	cluster = mockClusterGroup("test-cluster-6")
// 	lcKey = logicalcloud.ClusterGroupKey{
// 		ClusterGroup: "group1",
// 		Project:      "proj1"}
// 	_ = mockdb.Insert("resources", lcKey, nil, "data", cluster)

// }

// func validateError(err error, message string) {
// 	if len(message) == 0 {
// 		Expect(err).NotTo(HaveOccurred())
// 		Expect(err).To(BeNil())
// 		return
// 	}
// 	Expect(err.Error()).To(ContainSubstring(message))
// }
