// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"reflect"
	"strings"
	"testing"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

func TestCreateGenericPlacementIntent(t *testing.T) {
	testCases := []struct {
		label                    string
		inputIntent              GenericPlacementIntent
		inputProject             string
		inputCompositeApp        string
		inputCompositeAppVersion string
		inputDepIntGrpName       string
		expectedError            string
		mockdb                   *db.MockDB
		expected                 GenericPlacementIntent
	}{
		{
			label: "Create GenericPlacementIntent",
			inputIntent: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "testGenericPlacement",
					Description: " A sample intent for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			inputProject:             "testProject",
			inputCompositeApp:        "testCompositeApp",
			inputCompositeAppVersion: "testCompositeAppVersion",
			inputDepIntGrpName:       "testDeploymentIntentGroup",
			expected: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "testGenericPlacement",
					Description: " A sample intent for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
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
			intentCli := NewGenericPlacementIntentClient()
			got, _, err := intentCli.CreateGenericPlacementIntent(testCase.inputIntent, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.inputDepIntGrpName, true)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("CreateGenericPlacementIntent returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("CreateGenericPlacementIntent returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("CreateGenericPlacementIntent returned unexpected body: got %v; "+" expected %v", got, testCase.expected)
				}
			}
		})

	}
}

func TestGetGenericPlacementIntent(t *testing.T) {

	testCases := []struct {
		label                     string
		expectedError             string
		expected                  GenericPlacementIntent
		mockdb                    *db.MockDB
		intentName                string
		projectName               string
		compositeAppName          string
		compositeAppVersion       string
		deploymentIntentGroupName string
	}{
		{
			label:                     "Get Intent",
			intentName:                "testIntent",
			projectName:               "testProject",
			compositeAppName:          "testCompositeApp",
			compositeAppVersion:       "testVersion",
			deploymentIntentGroupName: "testDeploymentIntentGroup",
			expected: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "testIntent",
					Description: "A sample intent for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						GenericPlacementIntentKey{
							Name:         "testIntent",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testVersion",
							DigName:      "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testIntent\"," +
									"\"Description\":\"A sample intent for testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}" +
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
			intentCli := NewGenericPlacementIntentClient()
			got, err := intentCli.GetGenericPlacementIntent(testCase.intentName, testCase.projectName, testCase.compositeAppName, testCase.compositeAppVersion, testCase.deploymentIntentGroupName)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetGenericPlacementIntent returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetGenericPlacementIntent returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetGenericPlacementIntent returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}

		})
	}

}

func TestUpdateGenericPlacementIntent(t *testing.T) {
	testCases := []struct {
		label                  string
		genericPlacementIntent GenericPlacementIntent
		project                string
		compositeApp           string
		compositeAppVersion    string
		deploymentIntentGroup  string
		expectedError          string
		mockDB                 *db.MockDB
		expected               GenericPlacementIntent
		exists                 bool
	}{
		{
			label:  "Update Non Existing GenericPlacementIntent",
			exists: false,
			genericPlacementIntent: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "testGenericPlacement",
					Description: " A sample intent for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			compositeAppVersion:   "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			expected: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "testGenericPlacement",
					Description: " A sample intent for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
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
						},
					},
				},
			},
		},
		{
			label:  "Update Existing GenericPlacementIntent",
			exists: true,
			genericPlacementIntent: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "testGenericPlacement",
					Description: " This is a new intent for testing",
					UserData1:   "This is a new User Data 1 for testing",
					UserData2:   "This is a new User Data 2 for testing",
				},
			},
			project:               "testProject",
			compositeApp:          "testCompositeApp",
			compositeAppVersion:   "testCompositeAppVersion",
			deploymentIntentGroup: "testDeploymentIntentGroup",
			expected: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "testGenericPlacement",
					Description: " This is a new intent for testing",
					UserData1:   "This is a new User Data 1 for testing",
					UserData2:   "This is a new User Data 2 for testing",
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
						},
						GenericPlacementIntentKey{
							Name:         "testGenericPlacement",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testGenericPlacement\"," +
									"\"Description\":\"A sample intent for testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}" +
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
			gpiCli := NewGenericPlacementIntentClient()
			gpi, gpiExists, err := gpiCli.CreateGenericPlacementIntent(testCase.genericPlacementIntent, testCase.project, testCase.compositeApp, testCase.compositeAppVersion, testCase.deploymentIntentGroup, false)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("CreateGenericPlacementIntent returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("CreateGenericPlacementIntent returned an unexpected error %s", err)
				}
			}

			if reflect.DeepEqual(testCase.expected, gpi) == false {
				t.Errorf("CreateGenericPlacementIntent returned unexpected body: got %v; "+" expected %v", gpi, testCase.expected)
			}

			if testCase.exists != gpiExists {
				t.Errorf("CreateGenericPlacementIntent returned unexpected status: got %v; "+" expected %v", gpiExists, testCase.exists)
			}

		})
	}
}
