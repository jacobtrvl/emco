// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"reflect"
	"strings"
	"testing"

	gpic "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/gpic"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

func TestCreateAppIntent(t *testing.T) {
	testCases := []struct {
		label                        string
		inputAppIntent               AppIntent
		inputProject                 string
		inputCompositeApp            string
		inputCompositeAppVersion     string
		inputGenericPlacementIntent  string
		inputDeploymentIntentGrpName string
		expectedError                string
		mockdb                       *db.MockDB
		expected                     AppIntent
	}{
		{
			label: "Create AppIntent",
			inputAppIntent: AppIntent{
				MetaData: MetaData{
					Name:        "testAppIntent",
					Description: "A sample AppIntent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: SpecData{
					AppName: "SampleApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
								//ClusterLabelName: "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
								//ClusterLabelName: "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{ProviderName: "aws",
										ClusterLabelName: "east-us1"},
									{ProviderName: "aws",
										ClusterLabelName: "east-us2"},
									//{ClusterName: "east-us1"},
									//{ClusterName: "east-us2"},
								},
							},
						},

						AnyOfArray: []gpic.AnyOf{},
					},
				},
			},

			inputProject:                 "testProject",
			inputCompositeApp:            "testCompositeApp",
			inputCompositeAppVersion:     "testCompositeAppVersion",
			inputGenericPlacementIntent:  "testIntent",
			inputDeploymentIntentGrpName: "testDeploymentIntentGroup",
			expected: AppIntent{
				MetaData: MetaData{
					Name:        "testAppIntent",
					Description: "A sample AppIntent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: SpecData{
					AppName: "SampleApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
								//ClusterLabelName: "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
								//ClusterLabelName: "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{ProviderName: "aws",
										ClusterLabelName: "east-us1"},
									{ProviderName: "aws",
										ClusterLabelName: "east-us2"},
									//{ClusterName: "east-us1"},
									//{ClusterName: "east-us2"},
								},
							},
						},
						AnyOfArray: []gpic.AnyOf{},
					},
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
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
						GenericPlacementIntentKey{
							Name:         "testIntent",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testIntent\"," +
									"\"Description\":\"A sample intent for testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"Logical-Cloud\": \"logicalCloud1\"}}"),
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
			db.DBconn = testCase.mockdb
			appIntentCli := NewAppIntentClient()
			got, _, err := appIntentCli.CreateAppIntent(testCase.inputAppIntent, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.inputGenericPlacementIntent, testCase.inputDeploymentIntentGrpName, true)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("CreateAppIntent returned an unexpected error %s, ", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("CreateAppIntent returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("CreateAppIntent returned unexpected body: got %v; "+" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetAppIntent(t *testing.T) {
	testCases := []struct {
		label                   string
		expectedError           string
		expected                AppIntent
		mockdb                  *db.MockDB
		appIntentName           string
		projectName             string
		compositeAppName        string
		compositeAppVersion     string
		genericPlacementIntent  string
		deploymentIntentgrpName string
	}{
		{
			label:                   "Get Intent",
			appIntentName:           "testAppIntent",
			projectName:             "testProject",
			compositeAppName:        "testCompositeApp",
			compositeAppVersion:     "testCompositeAppVersion",
			genericPlacementIntent:  "testIntent",
			deploymentIntentgrpName: "testDeploymentIntentGroup",
			expected: AppIntent{
				MetaData: MetaData{
					Name:        "testAppIntent",
					Description: "testAppIntent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: SpecData{
					AppName: "SampleApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{ProviderName: "aws",
										ClusterLabelName: "east-us1"},
									{ProviderName: "aws",
										ClusterLabelName: "east-us2"},
								},
							},
						},
					},
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						AppIntentKey{
							Name:                      "testAppIntent",
							Project:                   "testProject",
							CompositeApp:              "testCompositeApp",
							Version:                   "testCompositeAppVersion",
							Intent:                    "testIntent",
							DeploymentIntentGroupName: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testAppIntent\"," +
									"\"Description\":\"testAppIntent\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"app\": \"SampleApp\"," +
									"\"intent\": {" +
									"\"allOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge2\"}," +
									"{" +
									"\"anyOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us2\"}" +
									"]}]" +
									"}}}"),
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			appIntentCli := NewAppIntentClient()
			got, err := appIntentCli.GetAppIntent(testCase.appIntentName, testCase.projectName, testCase.compositeAppName, testCase.compositeAppVersion,
				testCase.genericPlacementIntent, testCase.deploymentIntentgrpName)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetAppIntent returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetAppIntent returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetAppIntent returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}

		})
	}
}

func TestUpdateAppIntent(t *testing.T) {
	testCases := []struct {
		label                   string
		appIntent               AppIntent
		project                 string
		compositeApp            string
		compositeAppVersion     string
		genericPlacementIntent  string
		deploymentIntentGrpName string
		expectedError           string
		mockDB                  *db.MockDB
		expected                AppIntent
		exists                  bool
	}{
		{
			label:  "Update Non Existing AppIntent",
			exists: false,
			appIntent: AppIntent{
				MetaData: MetaData{
					Name:        "testAppIntent",
					Description: "A sample AppIntent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: SpecData{
					AppName: "SampleApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us2",
									},
								},
							},
						},

						AnyOfArray: []gpic.AnyOf{},
					},
				},
			},

			project:                 "testProject",
			compositeApp:            "testCompositeApp",
			compositeAppVersion:     "testCompositeAppVersion",
			genericPlacementIntent:  "testIntent",
			deploymentIntentGrpName: "testDeploymentIntentGroup",
			expected: AppIntent{
				MetaData: MetaData{
					Name:        "testAppIntent",
					Description: "A sample AppIntent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: SpecData{
					AppName: "SampleApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us2",
									},
								},
							},
						},
						AnyOfArray: []gpic.AnyOf{},
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
						GenericPlacementIntentKey{
							Name:         "testIntent",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testIntent\"," +
									"\"Description\":\"A sample intent for testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"Logical-Cloud\": \"logicalCloud1\"}}"),
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
			label:  "Update Existing AppIntent",
			exists: true,
			appIntent: AppIntent{
				MetaData: MetaData{
					Name:        "testAppIntent",
					Description: " This is a new AppIntent for testing",
					UserData1:   "This is a new User Data 1 for testing",
					UserData2:   "This is a new User Data 2 for testing",
				},
				Spec: SpecData{
					AppName: "SampleApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us2",
									},
								},
							},
						},

						AnyOfArray: []gpic.AnyOf{},
					},
				},
			},

			project:                 "testProject",
			compositeApp:            "testCompositeApp",
			compositeAppVersion:     "testCompositeAppVersion",
			genericPlacementIntent:  "testIntent",
			deploymentIntentGrpName: "testDeploymentIntentGroup",
			expected: AppIntent{
				MetaData: MetaData{
					Name:        "testAppIntent",
					Description: " This is a new AppIntent for testing",
					UserData1:   "This is a new User Data 1 for testing",
					UserData2:   "This is a new User Data 2 for testing",
				},
				Spec: SpecData{
					AppName: "SampleApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "east-us2",
									},
								},
							},
						},
						AnyOfArray: []gpic.AnyOf{},
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
						GenericPlacementIntentKey{
							Name:         "testIntent",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
							DigName:      "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testIntent\"," +
									"\"Description\":\"A sample intent for testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"Logical-Cloud\": \"logicalCloud1\"}}"),
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
						AppIntentKey{
							Name:                      "testAppIntent",
							Project:                   "testProject",
							CompositeApp:              "testCompositeApp",
							Version:                   "testCompositeAppVersion",
							Intent:                    "testIntent",
							DeploymentIntentGroupName: "testDeploymentIntentGroup",
						}.String(): {
							"data": []byte(
								"{\"metadata\":{\"Name\":\"testAppIntent\"," +
									"\"Description\":\"testAppIntent\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}," +
									"\"spec\":{\"app\": \"SampleApp\"," +
									"\"intent\": {" +
									"\"allOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"cluster\":\"edge2\"}," +
									"{" +
									"\"anyOf\":[" +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us1\"}," +
									"{" +
									"\"clusterProvider\":\"aws\"," +
									"\"clusterLabel\":\"east-us2\"}" +
									"]}]" +
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
			aiCli := NewAppIntentClient()
			ai, aiExists, err := aiCli.CreateAppIntent(testCase.appIntent, testCase.project, testCase.compositeApp, testCase.compositeAppVersion, testCase.genericPlacementIntent, testCase.deploymentIntentGrpName, false)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("CreateAppIntent returned an unexpected error %s, ", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("CreateAppIntent returned an unexpected error %s", err)
				}
			}

			if reflect.DeepEqual(testCase.expected, ai) == false {
				t.Errorf("CreateAppIntent returned unexpected body: got %v; "+" expected %v", ai, testCase.expected)
			}

			if aiExists != testCase.exists {
				t.Errorf("CreateAppIntent returned unexpected status: got %v; "+" expected %v", aiExists, testCase.exists)
			}
		})
	}
}
