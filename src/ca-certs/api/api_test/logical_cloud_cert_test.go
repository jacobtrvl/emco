// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

// These test cases are to validate the route handler functionalities
package api_test

// import (
// 	"errors"
// 	"net/http"

// 	"gitlab.com/project-emco/core/emco-base/src/ca-certs/api"
// 	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate"

// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/ginkgo/extensions/table"
// )

// type mockLogicalCloudCertManager struct {
// 	Items []module.Cert
// 	Err   error
// }

// func init() {
// 	// api.CertificateSchemaJson = "../json-schemas/certificate.json"
// 	api.CertificateSchemaJson = "C:\\Users\\subinjoh\\OneDrive - Intel Corporation\\Desktop\\Intel\\emco\\dev\\emco-base\\src\\ca-certs\\json-schemas\\certificate.json"
// }

// func (m *mockLogicalCloudCertManager) CreateCert(cert module.Cert, project string, failIfExists bool) (module.Cert, bool, error) {
// 	iExists := false
// 	index := 0

// 	if m.Err != nil {
// 		return module.Cert{}, iExists, m.Err
// 	}

// 	for i, item := range m.Items {
// 		if item.MetaData.Name == cert.MetaData.Name {
// 			iExists = true
// 			index = i
// 			break
// 		}
// 	}

// 	if iExists && failIfExists { // cert already exists
// 		return module.Cert{}, iExists, errors.New("Certificate already exists")
// 	}

// 	if iExists && !failIfExists { // cert already exists. update the cert
// 		m.Items[index] = cert
// 		return m.Items[index], iExists, nil
// 	}

// 	m.Items = append(m.Items, cert) // create the cert

// 	return m.Items[len(m.Items)-1], iExists, nil

// }
// func (m *mockLogicalCloudCertManager) DeleteCert(cert, project string) error {
// 	if m.Err != nil {
// 		return m.Err
// 	}

// 	for k, item := range m.Items {
// 		if item.MetaData.Name == cert { // cert exist
// 			m.Items[k] = m.Items[len(m.Items)-1]
// 			m.Items[len(m.Items)-1] = module.Cert{}
// 			m.Items = m.Items[:len(m.Items)-1]
// 			return nil
// 		}
// 	}

// 	return errors.New("db Remove resource not found") // cert does not exist

// }

// func (m *mockLogicalCloudCertManager) GetAllCert(project string) ([]module.Cert, error) {
// 	if m.Err != nil {
// 		return []module.Cert{}, m.Err
// 	}

// 	var certs []module.Cert
// 	certs = append(certs, m.Items...)

// 	return certs, nil

// }
// func (m *mockLogicalCloudCertManager) GetCert(cert, project string) (module.Cert, error) {
// 	if m.Err != nil {
// 		return module.Cert{}, m.Err
// 	}

// 	for _, item := range m.Items {
// 		if item.MetaData.Name == cert {
// 			return item, nil
// 		}
// 	}

// 	return module.Cert{}, errors.New("Certificate not found")
// }

// var _ = Describe("Test create cert handler",
// 	func() {
// 		DescribeTable("Create Cert",
// 			func(t test) {
// 				client := t.client.(*mockLogicalCloudCertManager)
// 				res := executeRequest(http.MethodPost, "", logicalCloudCertURL, client, t.input)
// 				validateCertResponse(res, t)
// 			},
// 			Entry("request body validation",
// 				test{
// 					entry:      "request body validation",
// 					input:      certInput(""), // create an empty clusterGroup payload
// 					result:     module.Cert{},
// 					err:        errors.New("cert name may not be empty\n"),
// 					statusCode: http.StatusBadRequest,
// 					client: &mockLogicalCloudCertManager{
// 						Err:   nil,
// 						Items: populateCertTestData(),
// 					},
// 				},
// 			),
// 			Entry("successful create",
// 				test{
// 					entry:      "successful create",
// 					input:      certInput("testCert"),
// 					result:     certResult("testCert"),
// 					err:        nil,
// 					statusCode: http.StatusCreated,
// 					client: &mockLogicalCloudCertManager{
// 						Err:   nil,
// 						Items: populateCertTestData(),
// 					},
// 				},
// 			),
// 			Entry("cert already exists",
// 				test{
// 					entry:      "cert already exists",
// 					input:      certInput("testCert1"),
// 					result:     module.Cert{},
// 					err:        errors.New("certificate already exists\n"),
// 					statusCode: http.StatusConflict,
// 					client: &mockLogicalCloudCertManager{
// 						Err:   nil,
// 						Items: populateCertTestData(),
// 					},
// 				},
// 			),
// 		)
// 	},
// )

// var _ = Describe("Test get cert handler",
// 	func() {
// 		DescribeTable("Get Cert",
// 			func(t test) {
// 				client := t.client.(*mockLogicalCloudCertManager)
// 				res := executeRequest(http.MethodGet, "/"+t.name, logicalCloudCertURL, client, nil)
// 				validateCertResponse(res, t)
// 			},
// 			Entry("successful get",
// 				test{
// 					entry:      "successful get",
// 					name:       "testCert1",
// 					statusCode: http.StatusOK,
// 					err:        nil,
// 					result:     certResult("testCert1"),
// 					client: &mockLogicalCloudCertManager{
// 						Err:   nil,
// 						Items: populateCertTestData(),
// 					},
// 				},
// 			),
// 			Entry("cert not found",
// 				test{
// 					entry:      "cert not found",
// 					name:       "nonExistingCert",
// 					statusCode: http.StatusNotFound,
// 					err:        errors.New("certificate not found\n"),
// 					result:     module.Cert{},
// 					client: &mockLogicalCloudCertManager{
// 						Err:   nil,
// 						Items: populateCertTestData(),
// 					},
// 				},
// 			),
// 		)
// 	},
// )

// var _ = Describe("Test update cert handler",
// 	func() {
// 		DescribeTable("Update Cert",
// 			func(t test) {
// 				client := t.client.(*mockLogicalCloudCertManager)
// 				res := executeRequest(http.MethodPut, "/"+t.name, logicalCloudCertURL, client, t.input)
// 				validateCertResponse(res, t)
// 			},
// 			Entry("request body validation",
// 				test{
// 					entry:      "request body validation",
// 					name:       "testCert",
// 					input:      certInput(""), // create an empty cert payload
// 					result:     module.Cert{},
// 					err:        errors.New("cert name may not be empty\n"),
// 					statusCode: http.StatusBadRequest,
// 					client: &mockLogicalCloudCertManager{
// 						Err:   nil,
// 						Items: populateCertTestData(),
// 					},
// 				},
// 			),
// 			Entry("successful update",
// 				test{
// 					entry:      "successful update",
// 					name:       "testCert",
// 					input:      certInput("testCert"),
// 					result:     certResult("testCert"),
// 					err:        nil,
// 					statusCode: http.StatusCreated,
// 					client: &mockLogicalCloudCertManager{
// 						Err:   nil,
// 						Items: populateCertTestData(),
// 					},
// 				},
// 			),
// 			Entry("cert already exists",
// 				test{
// 					entry:      "cert already exists",
// 					name:       "testCert4",
// 					input:      certInput("testCert4"),
// 					result:     certResult("testCert4"),
// 					err:        nil,
// 					statusCode: http.StatusOK,
// 					client: &mockLogicalCloudCertManager{
// 						Err:   nil,
// 						Items: populateCertTestData(),
// 					},
// 				},
// 			),
// 		)
// 	},
// )

// var _ = Describe("Test delete cert handler",
// 	func() {
// 		DescribeTable("Delete Cert",
// 			func(t test) {
// 				client := t.client.(*mockLogicalCloudCertManager)
// 				res := executeRequest(http.MethodDelete, "/"+t.name, logicalCloudCertURL, client, nil)
// 				validateCertResponse(res, t)
// 			},
// 			Entry("successful delete",
// 				test{
// 					entry:      "successful delete",
// 					name:       "testCert1",
// 					statusCode: http.StatusNoContent,
// 					err:        nil,
// 					result:     module.Cert{},
// 					client: &mockLogicalCloudCertManager{
// 						Err:   nil,
// 						Items: populateCertTestData(),
// 					},
// 				},
// 			),
// 			Entry("db remove cert not found",
// 				test{
// 					entry:      "db remove cert not found",
// 					name:       "nonExistingCert",
// 					statusCode: http.StatusNotFound,
// 					err:        errors.New("The requested resource not found\n"),
// 					result:     module.Cert{},
// 					client: &mockLogicalCloudCertManager{
// 						Err:   nil,
// 						Items: populateCertTestData(),
// 					},
// 				},
// 			),
// 		)
// 	},
// )
