// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

// These test cases are to validate the module level functionalities.
package module_test

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/project-emco/core/emco-base/examples/sample-controller/pkg/model"
	"gitlab.com/project-emco/core/emco-base/examples/sample-controller/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	orchmodule "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

var _ = Describe("SampleIntent",
	func() {
		var (
			project   orchmodule.Project
			pClient   *orchmodule.ProjectClient
			app       orchmodule.CompositeApp
			aClient   *orchmodule.CompositeAppClient
			diGroup   orchmodule.DeploymentIntentGroup
			digClient *orchmodule.DeploymentIntentGroupClient
			intent    model.SampleIntent
			iClient   *module.SampleIntentClient
			mdb       *db.NewMockDB
		)

		BeforeEach(
			func() {
				pClient = orchmodule.NewProjectClient()
				project = orchmodule.Project{
					MetaData: orchmodule.ProjectMetaData{
						Name: "testProj",
					},
				}

				aClient = orchmodule.NewCompositeAppClient()
				app = orchmodule.CompositeApp{
					Metadata: orchmodule.CompositeAppMetaData{
						Name: "app",
					},
					Spec: orchmodule.CompositeAppSpec{
						Version: "v1",
					},
				}

				digClient = orchmodule.NewDeploymentIntentGroupClient()
				diGroup = orchmodule.DeploymentIntentGroup{
					MetaData: orchmodule.DepMetaData{
						Name: "diGroup",
					},
					Spec: orchmodule.DepSpecData{
						Profile:      "profilename",
						Version:      "testver",
						LogicalCloud: "logCloud",
					},
				}

				iClient = module.NewIntentClient()
				intent = model.SampleIntent{
					Metadata: model.Metadata{
						Name: "sampleIntentName",
					},
				}

				mdb = new(db.NewMockDB)
				mdb.Err = nil
				db.DBconn = mdb
			},
		)

		Describe("Create  SampleIntent",
			func() {
				It("successful creation of  SampleIntent",
					func() {
						ctx := context.Background()
						// set up prerequisites
						_, err := (*pClient).CreateProject(ctx, project, false)
						Expect(err).To(BeNil())
						_, err = (*aClient).CreateCompositeApp(ctx, app, "testProj", false)
						Expect(err).To(BeNil())
						_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, diGroup, "testProj", "app", "v1", true)
						Expect(err).To(BeNil())

						// test  intent creation
						_, err = (*iClient).CreateSampleIntent(ctx, intent, "testProj", "app", "v1", "diGroup", true)
						Expect(err).To(BeNil())
					},
				)
			},
		)

		Describe("Get  SampleIntent",
			func() {
				It("successful get of SampleIntent",
					func() {
						ctx := context.Background()
						// set up prerequisites
						_, err := (*pClient).CreateProject(ctx, project, false)
						Expect(err).To(BeNil())
						_, err = (*aClient).CreateCompositeApp(ctx, app, "testProj", false)
						Expect(err).To(BeNil())
						_, _, err = (*digClient).CreateDeploymentIntentGroup(ctx, diGroup, "testProj", "app", "v1", true)
						Expect(err).To(BeNil())

						// test  intent creation
						_, err = (*iClient).CreateSampleIntent(ctx, intent, "testProj", "app", "v1", "diGroup", false)
						Expect(err).To(BeNil())

						_, err = (*iClient).GetSampleIntents(ctx, "sampleIntentName", "testProj", "app", "v1", "diGroup")
						Expect(err).To(BeNil())
					},
				)
			},
		)
	},
)
