package module_test

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"
)

var (
	cClient = module.NewCustomizationClient()
)

var _ = Describe("Create Customization",
	func() {
		BeforeEach(func() {
			populateCustomizationTestData()
		})
		Context("ccreate a customization that does not exist", func() {
			It("returns the customization, no error and, the exists flag is false", func() {
				l := len(mockdb.Items)
				mc := mockCustomization("new-customization")
				customization, cExists, err := cClient.CreateCustomization(
					mc, module.CustomizationContent{}, v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent, v.Resource, true)
				validateError(err, "")
				validateCustomization(customization, mc)
				Expect(cExists).To(Equal(false))
				Expect(len(mockdb.Items)).To(Equal(l + 1))
			})
		})
		Context("create a customization that already exists", func() {
			It("returns an error, no customization and, the exists flag is true", func() {
				l := len(mockdb.Items)
				mc := mockCustomization("test-customization-1")
				customization, cExists, err := cClient.CreateCustomization(
					mc, module.CustomizationContent{}, v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent, v.Resource, true)
				validateError(err, "Customization already exists")
				validateCustomization(module.Customization{}, customization)
				Expect(cExists).To(Equal(true))
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
	},
)

var _ = Describe("Delete Customization",
	func() {
		BeforeEach(func() {
			populateCustomizationTestData()
		})
		Context("delete an existing customization", func() {
			It("returns no error and delete the entry from the db", func() {
				l := len(mockdb.Items)
				err := cClient.DeleteCustomization(
					"test-customization-1", v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent, v.Resource)
				validateError(err, "")
				Expect(len(mockdb.Items)).To(Equal(l - 1))
			})
		})
		Context("delete a nonexisting customization", func() {
			It("returns an error and no change in the db", func() {
				l := len(mockdb.Items)
				mockdb.Err = errors.New("db Remove resource not found")
				err := cClient.DeleteCustomization(
					"non-existing-customization", v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent, v.Resource)
				validateError(err, "db Remove resource not found")
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
	},
)

var _ = Describe("Get All Customization",
	func() {
		BeforeEach(func() {
			populateCustomizationTestData()
		})
		Context("get all the customizations", func() {
			It("returns all the customizations, no error", func() {
				customizations, err := cClient.GetAllCustomization(
					v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent, v.Resource)
				validateError(err, "")
				Expect(len(customizations)).To(Equal(len(mockdb.Items)))
			})
		})
		Context("get all the customizations without creating any", func() {
			It("returns an empty array, no error", func() {
				mockdb.Items = []map[string]map[string][]byte{}
				customizations, err := cClient.GetAllCustomization(
					v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent, v.Resource)
				validateError(err, "")
				Expect(len(customizations)).To(Equal(0))
			})
		})
	},
)

var _ = Describe("Get Customization",
	func() {
		BeforeEach(func() {
			populateCustomizationTestData()
		})
		Context("get an existing customization", func() {
			It("returns the customization, no error", func() {
				customization, err := cClient.GetCustomization(
					"test-customization-1", v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent, v.Resource)
				validateError(err, "")
				validateCustomization(customization, mockCustomization("test-customization-1"))
			})
		})
		Context("get a nonexisting customization", func() {
			It("returns an error, no customization", func() {
				customization, err := cClient.GetCustomization(
					"non-existing-customization", v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent, v.Resource)
				validateError(err, "Customization not found")
				validateCustomization(customization, module.Customization{})
			})
		})
	},
)

var _ = Describe("Get Customization Content",
	func() {
		BeforeEach(func() {
			populateCustomizationTestData()
		})
		Context("get the existing customization content", func() {
			It("returns the customization content, no error", func() {
				populateCustomizationContent("test-customization-1")
				content, err := cClient.GetCustomizationContent(
					"test-customization-1", v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent, v.Resource)
				validateError(err, "")
				Expect(content).To(Equal(mockCustomizationContent()))
			})
		})
		Context("get the nonexisting customization content", func() {
			It("returns an empty content, no error", func() {
				content, err := cClient.GetCustomizationContent(
					"non-existing-customization", v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent, v.Resource)
				validateError(err, "")
				Expect(content).To(Equal(module.CustomizationContent{}))
			})
		})
	},
)

// validateCustomization
func validateCustomization(in, out module.Customization) {
	Expect(in).To(Equal(out))
}

// mockCustomization
func mockCustomization(name string) module.Customization {
	return module.Customization{
		Metadata: module.Metadata{
			Name:        name,
			Description: "test customization",
			UserData1:   "some user data 1",
			UserData2:   "some user data 2",
		},
		Spec: module.CustomizationSpec{
			ClusterSpecific: "true",
			ClusterInfo: module.ClusterInfo{
				Scope:           "label",
				ClusterProvider: "provider-1",
				ClusterName:     "cluster-1",
				ClusterLabel:    "edge-cluster-1",
				Mode:            "allow",
			},
			PatchType: "json",
			PatchJSON: []map[string]interface{}{
				{
					"op":    "replace",
					"path":  "/spec/replicas",
					"value": float64(6),
				},
			},
			ConfigMapOptions: module.ConfigMapOptions{
				DataKeyOptions: []module.KeyOptions{
					{
						FileName: "info.json",
						KeyName:  "info",
					},
				},
			},
		},
	}
}

// populateCustomizationTestData
func populateCustomizationTestData() {
	mockdb.Err = nil
	mockdb.Items = []map[string]map[string][]byte{}
	mockdb.MarshalErr = nil

	// Customization 1
	c := mockCustomization("test-customization-1")
	key := module.CustomizationKey{
		Customization:         c.Metadata.Name,
		Project:               v.Project,
		CompositeApp:          v.CompositeApp,
		CompositeAppVersion:   v.Version,
		DeploymentIntentGroup: v.DeploymentIntentGroup,
		GenericK8sIntent:      v.Intent,
		Resource:              v.Resource,
	}
	_ = mockdb.Insert("resources", key, nil, "data", c)

	// Customization 2
	c = mockCustomization("test-customization-2")
	key = module.CustomizationKey{
		Customization:         c.Metadata.Name,
		Project:               v.Project,
		CompositeApp:          v.CompositeApp,
		CompositeAppVersion:   v.Version,
		DeploymentIntentGroup: v.DeploymentIntentGroup,
		GenericK8sIntent:      v.Intent,
		Resource:              v.Resource,
	}
	_ = mockdb.Insert("resources", key, nil, "data", c)

	// Customization 3
	c = mockCustomization("test-customization-3")
	key = module.CustomizationKey{
		Customization:         c.Metadata.Name,
		Project:               v.Project,
		CompositeApp:          v.CompositeApp,
		CompositeAppVersion:   v.Version,
		DeploymentIntentGroup: v.DeploymentIntentGroup,
		GenericK8sIntent:      v.Intent,
		Resource:              v.Resource,
	}
	_ = mockdb.Insert("resources", key, nil, "data", c)
}

// populateCustomizationContent
func populateCustomizationContent(customization string) {
	key := module.CustomizationKey{
		Customization:         customization,
		Project:               v.Project,
		CompositeApp:          v.CompositeApp,
		CompositeAppVersion:   v.Version,
		DeploymentIntentGroup: v.DeploymentIntentGroup,
		GenericK8sIntent:      v.Intent,
		Resource:              v.Resource,
	}
	content := mockCustomizationContent()
	_ = mockdb.Insert("resources", key, nil, "customizationcontent", content)
}

// mockCustomizationContent
func mockCustomizationContent() module.CustomizationContent {
	return module.CustomizationContent{
		[]module.Content{
			{
				FileName: "info.json",
				Content:  "ewogICAgImZydWl0IjogImJlcnJ5IiwKICAgICJhbmltYWwiOiAiZG9nIiwKICAgICJjb2xvciI6ICJyZWQiCn0K",
				KeyName:  "info",
			},
			{
				FileName: "data.json",
				Content:  "ewogICAgIjEiOiAiYmVycnkiLAogICAgIjIiOiAiZG9nIiwKICAgICIzIjogInJlZCIKfQo=",
				KeyName:  "data",
			},
		},
	}
}
