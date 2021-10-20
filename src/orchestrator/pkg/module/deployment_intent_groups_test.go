// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"reflect"
	"strings"
	"testing"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

func TestCreateDeploymentIntentGroup(t *testing.T) {
	testCases := []struct {
		label                    string
		inputDeploymentIntentGrp DeploymentIntentGroup
		inputProject             string
		inputCompositeApp        string
		inputCompositeAppVersion string
		expectedError            string
		mockdb                   *db.MockDB
		expected                 DeploymentIntentGroup
	}{
		{
			label: "Create DeploymentIntentGroup",
			inputDeploymentIntentGrp: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: "DescriptionTestDeploymentIntentGroup",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: DepSpecData{
					Profile: "Testprofile",
					Version: "version of deployment",
					OverrideValuesObj: []OverrideValues{
						{AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							}},
						{AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							}},
					},
					LogicalCloud: "cloud1",
				},
			},
			inputProject:             "testProject",
			inputCompositeApp:        "testCompositeApp",
			inputCompositeAppVersion: "testCompositeAppVersion",
			expected: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: "DescriptionTestDeploymentIntentGroup",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: DepSpecData{
					Profile: "Testprofile",
					Version: "version of deployment",
					OverrideValuesObj: []OverrideValues{
						{AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							}},
						{AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							}},
					},
					LogicalCloud: "cloud1",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"project-name\":\"testProject\"," +
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
									"\"compositeAppVersion\":\"version of the composite app\"}}"),
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			depIntentCli := NewDeploymentIntentGroupClient()
			got, _, err := depIntentCli.CreateDeploymentIntentGroup(testCase.inputDeploymentIntentGrp, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, true)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("CreateDeploymentIntentGroup returned an unexpected error %s, ", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("CreateDeploymentIntentGroup returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("CreateDeploymentIntentGroup returned unexpected body: got %v; "+" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetDeploymentIntentGroup(t *testing.T) {
	testCases := []struct {
		label                    string
		inputDeploymentIntentGrp string
		inputProject             string
		inputCompositeApp        string
		inputCompositeAppVersion string
		expected                 DeploymentIntentGroup
		expectedError            string
		mockdb                   *db.MockDB
	}{
		{
			label:                    "Get DeploymentIntentGroup",
			inputDeploymentIntentGrp: "testDeploymentIntentGroup",
			inputProject:             "testProject",
			inputCompositeApp:        "testCompositeApp",
			inputCompositeAppVersion: "testCompositeAppVersion",
			expected: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: "DescriptionTestDeploymentIntentGroup",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: DepSpecData{
					Profile: "Testprofile",
					Version: "version of deployment",
					OverrideValuesObj: []OverrideValues{
						{AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							}},
						{AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							}},
					},
					LogicalCloud: "cloud1",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
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
									"\"spec\":{\"compositeProfile\": \"Testprofile\"," +
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
			db.DBconn = testCase.mockdb
			depIntentCli := NewDeploymentIntentGroupClient()
			got, err := depIntentCli.GetDeploymentIntentGroup(testCase.inputDeploymentIntentGrp, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetDeploymentIntentGroup returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetDeploymentIntentGroup returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetDeploymentIntentGroup returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestUpdateDeploymentIntentGroup(t *testing.T) {
	testCases := []struct {
		label                 string
		deploymentIntentGroup DeploymentIntentGroup
		project               string
		compositeApp          string
		compositeAppVersion   string
		expectedError         string
		mockDB                *db.MockDB
		expected              DeploymentIntentGroup
		exists                bool
	}{
		{
			label:  "Update Non Existing DeploymentIntentGroup",
			exists: false,
			deploymentIntentGroup: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: "DescriptionTestDeploymentIntentGroup",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: DepSpecData{
					Profile: "Testprofile",
					Version: "version of deployment",
					OverrideValuesObj: []OverrideValues{
						{AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							}},
						{AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							}},
					},
					LogicalCloud: "cloud1",
				},
			},
			project:             "testProject",
			compositeApp:        "testCompositeApp",
			compositeAppVersion: "testCompositeAppVersion",
			expected: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: "DescriptionTestDeploymentIntentGroup",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: DepSpecData{
					Profile: "Testprofile",
					Version: "version of deployment",
					OverrideValuesObj: []OverrideValues{
						{AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							}},
						{AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							}},
					},
					LogicalCloud: "cloud1",
				},
			},
			expectedError: "",
			mockDB: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"project-name\":\"testProject\"," +
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
									"\"compositeAppVersion\":\"version of the composite app\"}}"),
						},
					},
				},
			},
		},
		{
			label:  "Update Existing DeploymentIntentGroup",
			exists: true,
			deploymentIntentGroup: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: " This is a new DeploymentIntentGrou for testing",
					UserData1:   "This is a new User Data 1 for testing",
					UserData2:   "This is a new User Data 2 for testing",
				},
				Spec: DepSpecData{
					Profile: "Testprofile",
					Version: "version of deployment",
					OverrideValuesObj: []OverrideValues{
						{AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							}},
						{AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							}},
					},
					LogicalCloud: "cloud1",
				},
			},
			project:             "testProject",
			compositeApp:        "testCompositeApp",
			compositeAppVersion: "testCompositeAppVersion",
			expected: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: " This is a new DeploymentIntentGrou for testing",
					UserData1:   "This is a new User Data 1 for testing",
					UserData2:   "This is a new User Data 2 for testing",
				},
				Spec: DepSpecData{
					Profile: "Testprofile",
					Version: "version of deployment",
					OverrideValuesObj: []OverrideValues{
						{AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							}},
						{AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							}},
					},
					LogicalCloud: "cloud1",
				},
			},
			expectedError: "",
			mockDB: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"name\":\"testProject\"," +
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
									"\"compositeAppVersion\":\"version of the composite app\"}}"),
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
									"\"spec\":{\"compositeProfile\": \"Testprofile\"," +
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
							"stateInfo": []byte(
								"{ \"statusctxid\": \"\"," +
									"\"actions\": [{" +
									"\"state\":\"Created\"," +
									"\"instance\":\"\"," +
									"\"time\":\"2021-10-15T19:26:06.865+00:00\", " +
									"\"revision\":0" +
									"}]" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:  "Update Existing DeploymentIntentGroup with Approved state",
			exists: true,
			deploymentIntentGroup: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "testDeploymentIntentGroup",
					Description: " This is a new DeploymentIntentGrou for testing",
					UserData1:   "This is a new User Data 1 for testing",
					UserData2:   "This is a new User Data 2 for testing",
				},
				Spec: DepSpecData{
					Profile: "Testprofile",
					Version: "version of deployment",
					OverrideValuesObj: []OverrideValues{
						{AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							}},
						{AppName: "TestAppName",
							ValuesObj: map[string]string{
								"imageRepository": "registry.hub.docker.com",
							}},
					},
					LogicalCloud: "cloud1",
				},
			},
			project:             "testProject",
			compositeApp:        "testCompositeApp",
			compositeAppVersion: "testCompositeAppVersion",
			expected: DeploymentIntentGroup{
				MetaData: DepMetaData{
					Name:        "",
					Description: "",
					UserData1:   "",
					UserData2:   "",
				},
				Spec: DepSpecData{
					Profile:           "",
					Version:           "",
					OverrideValuesObj: nil,
					LogicalCloud:      "",
				},
			},
			expectedError: "The DeploymentIntentGroup is not updated",
			mockDB: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"data": []byte(
								"{\"name\":\"testProject\"," +
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
									"\"compositeAppVersion\":\"version of the composite app\"}}"),
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
									"\"spec\":{\"compositeProfile\": \"Testprofile\"," +
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
							"stateInfo": []byte(
								"{ \"statusctxid\": \"\"," +
									"\"actions\": [{" +
									"\"state\":\"Approved\"," +
									"\"instance\":\"\"," +
									"\"time\":\"2021-10-15T19:26:06.865+00:00\", " +
									"\"revision\":0" +
									"}]" +
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
			digCli := NewDeploymentIntentGroupClient()
			dig, digExists, err := digCli.CreateDeploymentIntentGroup(testCase.deploymentIntentGroup, testCase.project, testCase.compositeApp, testCase.compositeAppVersion, false)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("CreateDeploymentIntentGroup returned an unexpected error %s, ", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("CreateDeploymentIntentGroup returned an unexpected error %s", err)
				}
			}

			if reflect.DeepEqual(testCase.expected, dig) == false {
				t.Errorf("CreateDeploymentIntentGroup returned unexpected body: got %v; "+" expected %v", dig, testCase.expected)
			}

			if digExists != testCase.exists {
				t.Errorf("CreateDeploymentIntentGroup returned unexpected status: got %v; "+" expected %v", digExists, testCase.exists)
			}

		})
	}
}
