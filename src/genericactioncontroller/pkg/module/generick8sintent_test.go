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
	gkiClient = module.NewGenericK8sIntentClient()
)

var _ = Describe("Create GenericK8sIntent",
	func() {
		BeforeEach(func() {
			populateGenericK8sIntentTestData()
		})
		Context("create an intent that does not exist", func() {
			It("returns the intent, no error and, the exists flag is false", func() {
				l := len(mockdb.Items)
				mgki := mockGenericK8sIntent("new-gki-1")
				gki, gkiExists, err := gkiClient.CreateGenericK8sIntent(
					mgki, v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, true)
				validateError(err, "")
				validateGenericK8sIntent(gki, mgki)
				Expect(gkiExists).To(Equal(false))
				Expect(len(mockdb.Items)).To(Equal(l + 1))
			})
		})
		Context("create an intent that already exists", func() {
			It("returns an error, no intent and, the exists flag is true", func() {
				l := len(mockdb.Items)
				mgki := mockGenericK8sIntent("test-gki-1")
				gki, gkiExists, err := gkiClient.CreateGenericK8sIntent(
					mgki, v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup, true)
				validateError(err, "GenericK8sIntent already exists")
				validateGenericK8sIntent(gki, module.GenericK8sIntent{})
				Expect(gkiExists).To(Equal(true))
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
	},
)

var _ = Describe("Delete GenericK8sIntent",
	func() {
		BeforeEach(func() {
			populateGenericK8sIntentTestData()
		})
		Context("delete an existing intent", func() {
			It("returns no error and delete the entry from the db", func() {
				l := len(mockdb.Items)
				err := gkiClient.DeleteGenericK8sIntent(
					"test-gki-1", v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup)
				validateError(err, "")
				Expect(len(mockdb.Items)).To(Equal(l - 1))
			})
		})
		Context("delete a nonexisting intent", func() {
			It("returns an error and no change in the db", func() {
				l := len(mockdb.Items)
				mockdb.Err = errors.New("db Remove resource not found")
				err := gkiClient.DeleteGenericK8sIntent(
					"non-existing-gki", v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup)
				validateError(err, "db Remove resource not found")
				Expect(len(mockdb.Items)).To(Equal(l))
			})
		})
	},
)

var _ = Describe("Get All GenericK8sIntents",
	func() {
		BeforeEach(func() {
			populateGenericK8sIntentTestData()
		})
		Context("get all the intents", func() {
			It("returns all the intents, no error", func() {
				gkis, err := gkiClient.GetAllGenericK8sIntents(
					v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup)
				validateError(err, "")
				Expect(len(gkis)).To(Equal(len(mockdb.Items)))
			})
		})
		Context("get all the intents without creating any", func() {
			It("returns an empty array, no error", func() {
				mockdb.Items = []map[string]map[string][]byte{}
				gkis, err := gkiClient.GetAllGenericK8sIntents(
					v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup)
				validateError(err, "")
				Expect(len(gkis)).To(Equal(0))
			})
		})
	},
)

var _ = Describe("Get GenericK8sIntent",
	func() {
		BeforeEach(func() {
			populateGenericK8sIntentTestData()
		})
		Context("get an existing intent", func() {
			It("returns the intent, no error", func() {
				gki, err := gkiClient.GetGenericK8sIntent(
					"test-gki-1", v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup)
				validateError(err, "")
				validateGenericK8sIntent(gki, mockGenericK8sIntent("test-gki-1"))
			})
		})
		Context("get a nonexisting intent", func() {
			It("returns an error, no intent", func() {
				gki, err := gkiClient.GetGenericK8sIntent(
					"non-existing-gki", v.Project, v.CompositeApp, v.Version, v.DeploymentIntentGroup)
				validateError(err, "GenericK8sIntent not found")
				validateGenericK8sIntent(gki, module.GenericK8sIntent{})
			})
		})
	},
)

// validateGenericK8sIntent
func validateGenericK8sIntent(in, out module.GenericK8sIntent) {
	Expect(in).To(Equal(out))
}

// mockGenericK8sIntent
func mockGenericK8sIntent(name string) module.GenericK8sIntent {
	return module.GenericK8sIntent{
		Metadata: module.Metadata{
			Name:        name,
			Description: "test intent",
			UserData1:   "some user data 1",
			UserData2:   "some user data 2",
		},
	}
}

// populateGenericK8sIntentTestData
func populateGenericK8sIntentTestData() {
	mockdb.Err = nil
	mockdb.Items = []map[string]map[string][]byte{}
	mockdb.MarshalErr = nil

	// Intent 1
	gki := mockGenericK8sIntent("test-gki-1")
	key := module.GenericK8sIntentKey{
		GenericK8sIntent:      gki.Metadata.Name,
		Project:               v.Project,
		CompositeApp:          v.CompositeApp,
		CompositeAppVersion:   v.Version,
		DeploymentIntentGroup: v.DeploymentIntentGroup,
	}
	_ = mockdb.Insert("resources", key, nil, "data", gki)

	// Intent 2
	gki = mockGenericK8sIntent("test-gki-2")
	key = module.GenericK8sIntentKey{
		GenericK8sIntent:      gki.Metadata.Name,
		Project:               v.Project,
		CompositeApp:          v.CompositeApp,
		CompositeAppVersion:   v.Version,
		DeploymentIntentGroup: v.DeploymentIntentGroup,
	}
	_ = mockdb.Insert("resources", key, nil, "data", gki)

	// Intent 3
	gki = mockGenericK8sIntent("test-gki-3")
	key = module.GenericK8sIntentKey{
		GenericK8sIntent:      gki.Metadata.Name,
		Project:               v.Project,
		CompositeApp:          v.CompositeApp,
		CompositeAppVersion:   v.Version,
		DeploymentIntentGroup: v.DeploymentIntentGroup,
	}
	_ = mockdb.Insert("resources", key, nil, "data", gki)
}
