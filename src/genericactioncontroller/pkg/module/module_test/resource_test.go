package module_test

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"
)

var (
	rClient = module.NewResourceClient()
)

var _ = Describe("Create Resource",
	func() {
		BeforeEach(func() {
			populateResourceTestData()
		})
		Context("create a resource that does not exist", func() {
			It("returns the resource, no error and, the exists flag is false", func() {
				l := len(mockdb.Items)
				mr := mockResource("new-resource")
				res, rExists, err := rClient.CreateResource(
					mr, module.ResourceContent{}, v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent, true)
				validateError(err, "")
				validateResource(res, mr)
				Expect(rExists).To(Equal(false))
				Expect(len(mockdb.Items)).To(Equal(l + 1))
			})
		})
		Context("create a resource that already exists", func() {
			It("returns an error, no resource and, the exists flag is true", func() {
				l := len(mockdb.Items)
				mr := mockResource("test-resource-1")
				res, rExists, err := rClient.CreateResource(
					mr, module.ResourceContent{}, v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent, true)
				validateError(err, "Resource already exists")
				validateResource(res, module.Resource{})
				Expect(rExists).To(Equal(true))
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
	},
)

var _ = Describe("Delete Resource",
	func() {
		BeforeEach(func() {
			populateResourceTestData()
		})
		Context("delete an existing resource", func() {
			It("returns no error and delete the entry from the db", func() {
				l := len(mockdb.Items)
				err := rClient.DeleteResource(
					"test-resource-1", v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent)
				validateError(err, "")
				Expect(len(mockdb.Items)).To(Equal(l - 1))
			})
		})
		Context("delete a nonexisting resource", func() {
			It("returns an error and no change in the db", func() {
				l := len(mockdb.Items)
				mockdb.Err = errors.New("db Remove resource not found")
				err := rClient.DeleteResource(
					"non-existing-resource", v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent)
				validateError(err, "db Remove resource not found")
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
	},
)

var _ = Describe("Get All Resources",
	func() {
		BeforeEach(func() {
			populateResourceTestData()
		})
		Context("get all the resources", func() {
			It("returns all the resources, no error", func() {
				l := len(mockdb.Items)
				res, err := rClient.GetAllResources(
					v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent)
				validateError(err, "")
				Expect(len(res)).To(Equal(l))
			})
		})
		Context("get all the resources without creating any", func() {
			It("returns an empty array, no error", func() {
				mockdb.Items = []map[string]map[string][]byte{}
				res, err := rClient.GetAllResources(
					v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent)
				validateError(err, "")
				Expect(len(res)).To(Equal(0))
			})
		})
	},
)

var _ = Describe("Get Resource",
	func() {
		BeforeEach(func() {
			populateResourceTestData()
		})
		Context("get an existing resource", func() {
			It("returns the resource, no error", func() {
				res, err := rClient.GetResource(
					"test-resource-1", v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent)
				validateError(err, "")
				validateResource(res, mockResource("test-resource-1"))
			})
		})
		Context("get a nonexisting resource", func() {
			It("returns an error, no resource", func() {
				res, err := rClient.GetResource(
					"non-existing-resource", v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent)
				validateError(err, "Resource not found")
				validateResource(res, module.Resource{})
			})
		})
	},
)

var _ = Describe("Get Resource Content",
	func() {
		BeforeEach(func() {
			populateResourceTestData()
		})
		Context("get the existing resource content", func() {
			It("returns the resource content, no error", func() {
				populateResourceContent("test-resource-1")
				content, err := rClient.GetResourceContent(
					"test-resource-1", v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent)
				validateError(err, "")
				Expect(content.Content).To(Equal("YXBpVmVyc2lvbjogdjEKa2luZDogQ29"))
			})
		})
		Context("get the nonexisting resource content", func() {
			It("returns no content", func() {
				content, err := rClient.GetResourceContent(
					"non-existing-resource", v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, v.Intent)
				validateError(err, "")
				Expect(content).To(Equal(module.ResourceContent{}))
			})
		})
	},
)

// validateResource
func validateResource(in, out module.Resource) {
	Expect(in).To(Equal(out))
}

// mockResource
func mockResource(name string) module.Resource {
	return module.Resource{
		Metadata: module.Metadata{
			Name:        name,
			Description: "test resource",
			UserData1:   "some user data 1",
			UserData2:   "some user data 2",
		},
		Spec: module.ResourceSpec{
			AppName:   "operator",
			NewObject: "true",
			ResourceGVK: module.ResourceGVK{
				APIVersion: "v1",
				Kind:       "ConfigMap",
				Name:       "my-cm",
			},
		},
	}
}

// populateResourceTestData
func populateResourceTestData() {
	var (
		r   module.Resource
		key module.ResourceKey
	)

	mockdb.Err = nil
	mockdb.Items = []map[string]map[string][]byte{}
	mockdb.MarshalErr = nil

	// Resource 1
	r = mockResource("test-resource-1")
	key = module.ResourceKey{
		Resource:              r.Metadata.Name,
		Project:               v.Project,
		CompositeApp:          v.CompositeApp,
		CompositeAppVersion:   v.Version,
		DeploymentIntentGroup: v.DeploymentIntentGroup,
		GenericK8sIntent:      v.Intent,
	}
	_ = mockdb.Insert("resources", key, nil, "data", r)

	// Resource 2
	r = mockResource("test-resource-2")
	key = module.ResourceKey{
		Resource:              r.Metadata.Name,
		Project:               v.Project,
		CompositeApp:          v.CompositeApp,
		CompositeAppVersion:   v.Version,
		DeploymentIntentGroup: v.DeploymentIntentGroup,
		GenericK8sIntent:      v.Intent,
	}
	_ = mockdb.Insert("resources", key, nil, "data", r)

	// Resource 3
	r = mockResource("test-resource-3")
	key = module.ResourceKey{
		Resource:              r.Metadata.Name,
		Project:               v.Project,
		CompositeApp:          v.CompositeApp,
		CompositeAppVersion:   v.Version,
		DeploymentIntentGroup: v.DeploymentIntentGroup,
		GenericK8sIntent:      v.Intent,
	}
	_ = mockdb.Insert("resources", key, nil, "data", r)
}

// populateResourceContent
func populateResourceContent(resource string) {
	key := module.ResourceKey{
		Resource:              resource,
		Project:               v.Project,
		CompositeApp:          v.CompositeApp,
		CompositeAppVersion:   v.Version,
		DeploymentIntentGroup: v.DeploymentIntentGroup,
		GenericK8sIntent:      v.Intent,
	}
	rContent := module.ResourceContent{
		Content: "YXBpVmVyc2lvbjogdjEKa2luZDogQ29",
	}
	_ = mockdb.Insert("resources", key, nil, "resourcecontent", rContent)
}
