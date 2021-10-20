// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	pkgerrors "github.com/pkg/errors"
	gpic "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/gpic"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockAppIntentManager struct {
	// Items and err will be used to customize each test
	// via a localized instantiation of mockAppIntentManager
	Items           []moduleLib.AppIntent
	Err             error
	AppIntentExists bool
}

func (m *mockAppIntentManager) CreateAppIntent(a moduleLib.AppIntent, p string, ca string, v string, i string, digName string, failIfExists bool) (moduleLib.AppIntent, bool, error) {
	if m.AppIntentExists && failIfExists { // fail with Resource already exists
		return moduleLib.AppIntent{}, true, m.Err
	}

	if m.AppIntentExists && !failIfExists { // resource already exists. update the resource
		return m.Items[0], true, nil
	}

	if m.Err != nil {
		return moduleLib.AppIntent{}, false, m.Err
	}

	return m.Items[0], false, nil
}

func (m *mockAppIntentManager) GetAppIntent(ai string, p string, ca string, v string, i string, digName string) (moduleLib.AppIntent, error) {
	if m.Err != nil {
		return moduleLib.AppIntent{}, m.Err
	}

	return moduleLib.AppIntent{}, nil
}

func (m *mockAppIntentManager) GetAllIntentsByApp(aN, p, ca, v, i, digName string) (moduleLib.SpecData, error) {
	if m.Err != nil {
		return moduleLib.SpecData{}, m.Err
	}
	return moduleLib.SpecData{}, nil
}

func (m *mockAppIntentManager) DeleteAppIntent(ai string, p string, ca string, v string, i string, digName string) error {
	return m.Err
}

func (m *mockAppIntentManager) GetAllAppIntents(p, ca, v, i, digName string) ([]moduleLib.AppIntent, error) {
	return []moduleLib.AppIntent{}, nil
}

func init() {
	appIntentJSONFile = "../json-schemas/generic-placement-intent-app.json"
}

func Test_appintent_createHandler(t *testing.T) {
	testCases := []struct {
		label            string
		reader           io.Reader
		expected         moduleLib.AppIntent
		expectedCode     int
		errorString      string
		cAppIntentClient *mockAppIntentManager
	}{

		{
			label:        "Metadata name is missing",
			expectedCode: http.StatusBadRequest,
			errorString:  "Missing name for the intent",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"description": "description of placement_intent"
			 },
			 "spec": {
				"app": "app",
				"intent": {
				   "allOf": [
					  {
						 "clusterProvider": "p",
						 "clusterLabel": "c"
					  },
					  {
						"clusterProvider": "p",
						"clusterLabel": "d"
					 }
				   ]
				}
			 }
		  }`)),
			cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "app name is missing",
			expectedCode: http.StatusBadRequest,
			errorString:  "Missing app for the intent",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1",
				"description": "description of placement_intent"

			 },
			 "spec": {
				
			 }
		  }`)),
			cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "provider name is missing",
			expectedCode: http.StatusBadRequest,
			errorString:  "Missing clusterProvider in an intent",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1",
				"description": "description of placement_intent"

			 },
			 "spec": {
				"app": "app1",
				"intent": {
					"anyOf": [
					  {
						"clusterLabel": "c"
					  }
					]
				}
			  }		
			 }
		  }`)),
			cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "cluster label or name is missing",
			expectedCode: http.StatusBadRequest,
			errorString:  "Missing cluster or clusterLabel",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1",
				"description": "description of placement_intent"

			 },
			 "spec": {
				"app": "app1",
				"intent": {
					"anyOf": [
					  {
						"clusterProvider": "p"
					  }
					]
				}
			  }		
			 }
		  }`)),
			cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "duplicate input only one cluster label or name required",
			expectedCode: http.StatusBadRequest,
			errorString:  "Only one of cluster name or cluster label allowed",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1",
				"description": "description of placement_intent"

			 },
			 "spec": {
				"app": "app1",
				"intent": {
					"anyOf": [
					  {
						"clusterProvider": "p",
						"clusterLabel": "d",
						"cluster": "e"
					  }
					]
				}
			  }		
			 }
		  }`)),
			cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "allOf provider name missing",
			expectedCode: http.StatusBadRequest,
			errorString:  "Missing clusterProvider in an intent",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1"			 },
			 "spec": {
				"app": "app1",
				"intent": {
					"allOf": [{
							"name": "p",
							"clusterLabel": "c"
						},
						{
							"anyOf": [{
								"clusterProvider": "p",
								"clusterLabel": "d"
							}]
						}
					]
				}
			}}}`)),
			cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "allOf anyof provider name missing",
			expectedCode: http.StatusBadRequest,
			errorString:  "Missing clusterProvider in an intent",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1"			 },
			 "spec": {
				"app": "app1",
				"intent": {
					"allOf": [{
							"clusterProvider": "p",
							"clusterLabel": "c"
						},
						{
							"anyOf": [{
								"name": "p",
								"clusterLabel": "d"
							}]
						}
					]
				}
			}}}`)),
			cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "allOf duplicate input only one cluster label or name required",
			expectedCode: http.StatusBadRequest,
			errorString:  "Only one of cluster name or cluster label allowed",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1"			 },
			 "spec": {
				"app": "app1",
				"intent": {
					"allOf": [{
							"clusterProvider": "p",
							"clusterLabel": "c",
							"cluster": "e"
						},
						{
							"anyOf": [{
								"clusterProvider": "p",
								"clusterLabel": "d"
							}]
						}
					]
				}
			}}}`)),
			cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "allOf anyOf duplicate input only one cluster label or name required",
			expectedCode: http.StatusBadRequest,
			errorString:  "Only one of cluster name or cluster label allowed",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1"			 },
			 "spec": {
				"app": "app1",
				"intent": {
					"allOf": [{
							"clusterProvider": "p",
							"clusterLabel": "c"
						},
						{
							"anyOf": [{
								"clusterProvider": "p",
								"clusterLabel": "d",
								"cluster": "e"
							}]
						}
					]
				}
			}}}`)),
			cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "Success Case",
			expectedCode: http.StatusCreated,
			errorString:  "",
			reader: bytes.NewBuffer([]byte(`{   
				"metadata": {
				"name": "Test1"			 
				},
			 "spec": {
				"app": "app1",
				"intent": {
					"allOf": [{
							"clusterProvider": "aws",
							"cluster": "edge1"
						},
						{
							"clusterProvider": "aws",
							"clusterLabel": "west-us1"
						}
					]
				}
			}
			}`)),
			cAppIntentClient: &mockAppIntentManager{
				//Items that will be returned by the mocked Client
				Err: nil,
				Items: []moduleLib.AppIntent{
					{
						MetaData: moduleLib.MetaData{
							Name: "Test1",
						},
						Spec: moduleLib.SpecData{
							AppName: "app1",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
									},
								},
							},
						},
					},
				},
				AppIntentExists: false,
			},
			expected: moduleLib.AppIntent{
				MetaData: moduleLib.MetaData{
					Name: "Test1",
				},
				Spec: moduleLib.SpecData{
					AppName: "app1",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName:     "aws",
								ClusterLabelName: "west-us1",
							},
						},
					},
				},
			},
		},
		{
			label:        "App Intent Already Exists",
			expectedCode: http.StatusConflict,
			errorString:  "Intent already exists",
			reader: bytes.NewBuffer([]byte(`{   
				"metadata": {
				"name": "Test1"			 
				},
			 "spec": {
				"app": "app1",
				"intent": {
					"allOf": [{
							"clusterProvider": "aws",
							"cluster": "edge1"
						},
						{
							"clusterProvider": "aws",
							"clusterLabel": "west-us1"
						}
					]
				}
			}
			}`)),
			cAppIntentClient: &mockAppIntentManager{
				//Items that will be returned by the mocked Client
				Err: pkgerrors.New("AppIntent already exists"),
				Items: []moduleLib.AppIntent{
					{
						MetaData: moduleLib.MetaData{
							Name: "Test1",
						},
						Spec: moduleLib.SpecData{
							AppName: "app1",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
									},
								},
							},
						},
					},
				},
				AppIntentExists: true,
			},
			expected: moduleLib.AppIntent{
				MetaData: moduleLib.MetaData{
					Name: "",
				},
				Spec: moduleLib.SpecData{
					AppName: "",
					Intent: gpic.IntentStruc{
						AllOfArray: nil,
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents", testCase.reader)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, nil, testCase.cAppIntentClient, nil, nil, nil, nil, nil))
			b := resp.Body.String()

			//Check returned code
			if resp.Code != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.Code)
			}
			//Check returned body only if statusCreated
			if resp.Code == http.StatusCreated {
				got := moduleLib.AppIntent{}
				json.NewDecoder(resp.Body).Decode(&got)
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %+v;"+
						" expected %+v", got, testCase.expected)
				}
			} else {
				if !strings.Contains(b, testCase.errorString) {
					t.Fatal("Unexpected error found")
				}
			}
		})
	}
}

func Test_appintent_updateHandler(t *testing.T) {
	testCases := []struct {
		label           string
		reader          io.Reader
		expected        moduleLib.AppIntent
		expectedCode    int
		errorString     string
		appIntentClient *mockAppIntentManager
	}{

		{
			label:        "name is required",
			expectedCode: http.StatusBadRequest,
			errorString:  "Missing name for the intent",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"description": "description of placement_intent"
			 },
			 "spec": {
				"app": "app",
				"intent": {
				   "allOf": [
					  {
						 "clusterProvider": "p",
						 "clusterLabel": "c"
					  },
					  {
						"clusterProvider": "p",
						"clusterLabel": "d"
					 }
				   ]
				}
			 }
		  }`)),
			appIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "app is required",
			expectedCode: http.StatusBadRequest,
			errorString:  "Missing app for the intent",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1",
				"description": "description of placement_intent"
			 },
			 "spec": {
			 }
		  }`)),
			appIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "clusterProvider is required",
			expectedCode: http.StatusBadRequest,
			errorString:  "Missing clusterProvider in an intent",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1",
				"description": "description of placement_intent"
			 },
			 "spec": {
				"app": "app1",
				"intent": {
					"anyOf": [
					  {
						"clusterLabel": "c"
					  }
					]
				}
			  }		
			 }
		  }`)),
			appIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "cluster is required",
			expectedCode: http.StatusBadRequest,
			errorString:  "Missing cluster or clusterLabel",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1",
				"description": "description of placement_intent"
			 },
			 "spec": {
				"app": "app1",
				"intent": {
					"anyOf": [
					  {
						"clusterProvider": "p"
					  }
					]
				}
			  }		
			 }
		  }`)),
			appIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "Must not validate the schema (not)",
			expectedCode: http.StatusBadRequest,
			errorString:  "Only one of cluster name or cluster label allowed",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1",
				"description": "description of placement_intent"

			 },
			 "spec": {
				"app": "app1",
				"intent": {
					"anyOf": [
					  {
						"clusterProvider": "p",
						"clusterLabel": "d",
						"cluster": "e"
					  }
					]
				}
			  }		
			 }
		  }`)),
			appIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "Update Non Existing AppIntent",
			expectedCode: http.StatusCreated,
			errorString:  "",
			reader: bytes.NewBuffer([]byte(`{   
				"metadata": {
				"name": "Test1"			 
				},
			 "spec": {
				"app": "app1",
				"intent": {
					"allOf": [{
							"clusterProvider": "aws",
							"cluster": "edge1"
						},
						{
							"clusterProvider": "aws",
							"clusterLabel": "west-us1"
						}
					]
				}
			}
			}`)),
			appIntentClient: &mockAppIntentManager{
				//Items that will be returned by the mocked Client
				Err: nil,
				Items: []moduleLib.AppIntent{
					{
						MetaData: moduleLib.MetaData{
							Name: "Test1",
						},
						Spec: moduleLib.SpecData{
							AppName: "app1",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
									},
								},
							},
						},
					},
				},
				AppIntentExists: false,
			},
			expected: moduleLib.AppIntent{
				MetaData: moduleLib.MetaData{
					Name: "Test1",
				},
				Spec: moduleLib.SpecData{
					AppName: "app1",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName:     "aws",
								ClusterLabelName: "west-us1",
							},
						},
					},
				},
			},
		},
		{
			label:        "Update Existing AppIntent",
			expectedCode: http.StatusOK,
			errorString:  "",
			reader: bytes.NewBuffer([]byte(`{   
				"metadata": {
				"name": "Test1"			 
				},
			 "spec": {
				"app": "app1",
				"intent": {
					"allOf": [{
							"clusterProvider": "aws",
							"cluster": "edge1"
						},
						{
							"clusterProvider": "aws",
							"clusterLabel": "west-us1"
						}
					]
				}
			}
			}`)),
			appIntentClient: &mockAppIntentManager{
				//Items that will be returned by the mocked Client
				Err: nil,
				Items: []moduleLib.AppIntent{
					{
						MetaData: moduleLib.MetaData{
							Name: "Test1",
						},
						Spec: moduleLib.SpecData{
							AppName: "app1",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
									},
								},
							},
						},
					},
				},
				AppIntentExists: true,
			},
			expected: moduleLib.AppIntent{
				MetaData: moduleLib.MetaData{
					Name: "Test1",
				},
				Spec: moduleLib.SpecData{
					AppName: "app1",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName:     "aws",
								ClusterLabelName: "west-us1",
							},
						},
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("PUT", "/v2/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/generic-placement-intents/{genericPlacementIntent}/app-intents/{genericAppPlacementIntent}", testCase.reader)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, nil, testCase.appIntentClient, nil, nil, nil, nil, nil))
			b := resp.Body.String()

			//Check returned code
			if resp.Code != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.Code)
			}
			//Check returned body only if statusCreated
			if resp.Code == http.StatusCreated || resp.Code == http.StatusOK {
				ai := moduleLib.AppIntent{}
				json.NewDecoder(resp.Body).Decode(&ai)
				if reflect.DeepEqual(testCase.expected, ai) == false {
					t.Errorf("createHandler returned unexpected body: got %+v;"+
						" expected %+v", ai, testCase.expected)
				}
			} else {
				if !strings.Contains(b, testCase.errorString) {
					t.Fatal("Unexpected error found")
				}
			}
		})
	}
}
