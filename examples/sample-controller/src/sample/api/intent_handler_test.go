// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

// These test cases are to validate the route handler functionalities.
package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"gitlab.com/project-emco/core/emco-base/examples/sample-controller/api"
	"gitlab.com/project-emco/core/emco-base/examples/sample-controller/pkg/model"
)

// In this example, the SampleIntentManager exposes two functionalities (CreateSampleIntent, GetIntent).
// mockSampleIntentManager implements the mock services for the SampleIntentManager.
type mockIntentManager struct {
	Items []model.SampleIntent
	Err   error
}

type test struct {
	name       string
	input      io.Reader
	intent     model.SampleIntent
	err        error
	statusCode int
	client     *mockIntentManager
}

func init() {
	api.SampleJSONFile = "../json-schemas/intent.json"
}

func (m *mockIntentManager) CreateSampleIntent(ctx context.Context, intent model.SampleIntent, project, app, version, deploymentIntentGroup string, failIfExists bool) (model.SampleIntent, error) {
	if m.Err != nil {
		return model.SampleIntent{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockIntentManager) GetSampleIntents(ctx context.Context, name, project, app, version, deploymentIntentGroup string) ([]model.SampleIntent, error) {
	if m.Err != nil {
		return []model.SampleIntent{}, m.Err
	}

	return m.Items, nil
}

var _ = Describe("IntentHandler",
	func() {
		DescribeTable("Create SampleIntent",
			func(t test) {
				i := model.SampleIntent{}
				req := httptest.NewRequest("POST", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/sampleIntents", t.input)
				res := executeRequest(req, api.NewRouter(t.client))
				Expect(res.StatusCode).To(Equal(t.statusCode))
				json.NewDecoder(res.Body).Decode(&i)
				Expect(i).To(Equal(t.intent))
			},
			Entry("successful create",
				test{
					name:       "create",
					statusCode: http.StatusCreated,
					input: bytes.NewBuffer([]byte(`{
						"metadata": {
							"name": "testSampleIntent",
							"description": "test intent",
							"userData1": "some user data 1",
							"userData2": "some user data 2"
						},
						"spec": {
							"app": "testapp",
							"sampleIntentData": "testIntentData"
						}
					}`)),
					intent: model.SampleIntent{
						Metadata: model.Metadata{
							Name:        "testSampleIntent",
							Description: "test intent",
							UserData1:   "some user data 1",
							UserData2:   "some user data 2",
						},
						Spec: model.SampleIntentSpec{
							App:              "testApp",
							SampleIntentData: "testIntentData",
						},
					},
					err: nil,
					client: &mockIntentManager{
						Err:   nil,
						Items: populateTestData(),
					},
				},
			),
		// Add more entries to cover multiple create success/ failure scenarios.
		)
		DescribeTable("Get SampleIntent",
			func(t test) {
				i := model.SampleIntent{}
				req := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/sampleIntents/"+t.name, nil)
				res := executeRequest(req, api.NewRouter(t.client))
				Expect(res.StatusCode).To(Equal(t.statusCode))
				json.NewDecoder(res.Body).Decode(&i)
				Expect(i).To(Equal(t.intent))
			},
			Entry("successful get",
				test{
					name:       "get",
					statusCode: http.StatusOK,
					err:        nil,
					intent: model.SampleIntent{
						Metadata: model.Metadata{
							Name:        "testSampleIntent",
							Description: "test intent",
							UserData1:   "some user data 1",
							UserData2:   "some user data 2",
						},
						Spec: model.SampleIntentSpec{
							App:              "testApp",
							SampleIntentData: "testIntentData",
						},
					},
					client: &mockIntentManager{
						Err:   nil,
						Items: populateTestData(),
					},
				},
			),
		// Add more entries to cover multiple get success/ failure scenarios.
		)
		// Add more tests based on the handler functionalities.
	},
)

func populateTestData() []model.SampleIntent {
	return []model.SampleIntent{
		{
			Metadata: model.Metadata{
				Name:        "testSampleIntent",
				Description: "test intent",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
			Spec: model.SampleIntentSpec{
				App:              "testApp",
				SampleIntentData: "testIntentData",
			},
		},
		{
			Metadata: model.Metadata{
				Name:        "newSampleIntent",
				Description: "new intent",
				UserData1:   "some user data 1",
				UserData2:   "some user data 2",
			},
			Spec: model.SampleIntentSpec{
				App:              "newApp",
				SampleIntentData: "newIntentData",
			},
		},
		// Add more data based on the test scenarios.
	}
}
