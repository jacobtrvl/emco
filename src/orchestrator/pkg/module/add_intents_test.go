// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"reflect"
	"strings"
	"testing"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

func TestCreateIntent(t *testing.T) {
	testCases := []struct {
		label                 string
		intent                Intent
		project               string
		compositeApp          string
		compositeAppVersion   string
		deploymentIntentGroup string
		expectedError         string
		mockDB                *db.MockDB
		expected              Intent
		exists                bool
	}{
		{
			label:  "Create Intent",
			exists: false,
			intent: Intent{
				MetaData: IntentMetaData{
					Name:        "testIntent",
					Description: "A sample Intent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: IntentSpecData{
					Intent: map[string]string{
						"genericPlacementIntent": "testGenericPlacementIntent",
					},
				},
			},
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			compositeAppVersion:   "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			expected: Intent{
				MetaData: IntentMetaData{
					Name:        "testIntent",
					Description: "A sample Intent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: IntentSpecData{
					Intent: map[string]string{
						"genericPlacementIntent": "testGenericPlacementIntent",
					},
				},
			},
			expectedError: "",
			mockDB: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"project\":\"testProject\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						CompositeAppKey{CompositeAppName: "testCompositeApp",
							Version: "testCompositeAppVersion", Project: "testProject"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"testCompositeApp\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						DeploymentIntentGroupKey{
							Name:         "testDeploymentIntentGroup",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"testDeploymentIntentGroup\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"overrideValues\":[" +
									"{" +
									"\"app\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logicalCloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockDB
			iCli := NewIntentClient()
			i, iExists, err := iCli.AddIntent(testCase.intent, testCase.project, testCase.compositeApp, testCase.compositeAppVersion, testCase.deploymentIntentGroup, true)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("AddIntent returned an unexpected error %s, ", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("AddIntent returned an unexpected error %s", err)
				}
			}

			if reflect.DeepEqual(testCase.expected, i) == false {
				t.Errorf("AddIntent returned unexpected body: got %v; "+" expected %v", i, testCase.expected)
			}

			if iExists != testCase.exists {
				t.Errorf("AddIntent returned unexpected status: got %v; "+" expected %v", iExists, testCase.exists)
			}

		})
	}
}

func TestUpdateIntent(t *testing.T) {
	testCases := []struct {
		label                 string
		intent                Intent
		project               string
		compositeApp          string
		compositeAppVersion   string
		deploymentIntentGroup string
		expectedError         string
		mockDB                *db.MockDB
		expected              Intent
		exists                bool
	}{
		{
			label:  "Update Non Existing Intent",
			exists: false,
			intent: Intent{
				MetaData: IntentMetaData{
					Name:        "testIntent",
					Description: "A sample Intent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: IntentSpecData{
					Intent: map[string]string{
						"genericPlacementIntent": "testGenericPlacementIntent",
					},
				},
			},
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			compositeAppVersion:   "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			expected: Intent{
				MetaData: IntentMetaData{
					Name:        "testIntent",
					Description: "A sample Intent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: IntentSpecData{
					Intent: map[string]string{
						"genericPlacementIntent": "testGenericPlacementIntent",
					},
				},
			},
			expectedError: "",
			mockDB: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"project\":\"testProject\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						CompositeAppKey{CompositeAppName: "testCompositeApp",
							Version: "testCompositeAppVersion", Project: "testProject"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"testCompositeApp\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						DeploymentIntentGroupKey{
							Name:         "testDeploymentIntentGroup",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"testDeploymentIntentGroup\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"overrideValues\":[" +
									"{" +
									"\"app\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logicalCloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:  "Update Existing Intent",
			exists: true,
			intent: Intent{
				MetaData: IntentMetaData{
					Name:        "testIntent",
					Description: "A sample Intent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: IntentSpecData{
					Intent: map[string]string{
						"genericPlacementIntent": "testGenericPlacementIntent",
					},
				},
			},
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			compositeAppVersion:   "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			expected: Intent{
				MetaData: IntentMetaData{
					Name:        "testIntent",
					Description: "A sample Intent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: IntentSpecData{
					Intent: map[string]string{
						"genericPlacementIntent": "testGenericPlacementIntent",
					},
				},
			},
			expectedError: "",
			mockDB: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"project\":\"testProject\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						CompositeAppKey{CompositeAppName: "testCompositeApp",
							Version: "testCompositeAppVersion", Project: "testProject"}.String(): {
							"data": []byte(
								"{\"metadata\":{" +
									"\"name\":\"testCompositeApp\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						DeploymentIntentGroupKey{
							Name:         "testDeploymentIntentGroup",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"name\":\"testDeploymentIntentGroup\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"overrideValues\":[" +
									"{" +
									"\"app\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logicalCloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						IntentKey{
							Name:                  "testIntent",
							Project:               "testProject",
							CompositeApp:          "testCompositeApp",
							Version:               "testCompositeAppVersion",
							DeploymentIntentGroup: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\": {\"name\":\"testIntent\"," +
									"\"description\":\"A sample Intent\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\": {" +
									"\"intent\": {" +
									"\"genericPlacementIntent\": \"testGenericPlacementIntent\"" +
									"}}}"),
						},
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockDB
			intentCli := NewIntentClient()
			intent, iExists, err := intentCli.AddIntent(testCase.intent, testCase.project, testCase.compositeApp, testCase.compositeAppVersion, testCase.deploymentIntentGroup, false)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("CreateAppIntent returned an unexpected error %s, ", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("CreateAppIntent returned an unexpected error %s", err)
				}
			}

			if reflect.DeepEqual(testCase.expected, intent) == false {
				t.Errorf("CreateAppIntent returned unexpected body: got %v; "+" expected %v", intent, testCase.expected)
			}

			if iExists != testCase.exists {
				t.Errorf("CreateAppIntent returned unexpected status: got %v; "+" expected %v", iExists, testCase.exists)
			}
		})
	}
}
