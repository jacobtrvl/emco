// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certmanagerissuer

import (
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newCertificateRequest returns an instance of the CertificateRequest
func newCertificateRequest() *cmv1.CertificateRequest {
	return &cmv1.CertificateRequest{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cert-manager.io/v1",
			Kind:       "CertificateRequest",
		},
	}
}

// ResourceName returns the CertificateRequest resource name, used by the rsync
func ResourceName(name string) string {
	return fmt.Sprintf("%s+%s", name, "certificaterequest")
}

// CertificateRequestName retun the CertificateRequest name
func CertificateRequestName(contextID, cert, clusterProvider, cluster string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", contextID, cert, clusterProvider, cluster, "cr")
}

// CreateCertificateRequest retun the cert-manager CertificateRequest object
func CreateCertificateRequest(caCert module.Cert, name string, request []byte) (*cmv1.CertificateRequest, error) {
	// validate the request content
	if err := validateCSR(request); err != nil {
		return nil, err
	}

	cr := newCertificateRequest()
	cr.ObjectMeta = metav1.ObjectMeta{
		Name: name,
	}

	cr.Spec = cmv1.CertificateRequestSpec{
		Request: request,
		IsCA:    caCert.Spec.IsCA,
		Duration: &metav1.Duration{
			Duration: caCert.Spec.Duration,
		},
		IssuerRef: cmmetav1.ObjectReference{
			Name:  caCert.Spec.IssuerRef.Name,
			Kind:  caCert.Spec.IssuerRef.Kind,
			Group: caCert.Spec.IssuerRef.Group,
		},
	}

	return cr, nil
}

// TODO: Enable this once the monitor issue is fixed
// validateCertificates
func ValidateCertificateRequest(cr cmv1.CertificateRequest) error {
	// if len(cr.Status.Certificate) == 0 {
	// 	return errors.New("generated certificate is invalid")
	// }

	// if len(cr.Status.IsCA) == 0 {
	// 	return errors.New("certificate is generated by an Invalid certifcate authority")
	// }

	// approved := false
	// for _, state := range cr.Status.States {
	// 	if strings.ToLower(state.Status) == "true" &&
	// 		strings.ToLower(state.Type) == "approved" {
	// 		approved = true
	// 		break
	// 	}
	// }

	// if !approved {
	// 	return errors.New("the certificate is not yet approved by the CA")
	// }
	return nil
}

// RetrieveCertificateRequests retrieves the certificaterequests created with the cert enrollment
func RetrieveCertificateRequests(contextID string) ([]cmv1.CertificateRequest, error) {
	// var log = func(message, app, cluster, contextID, status string, err error) {
	// 	fields := make(logutils.Fields)
	// 	fields["ContextID"] = contextID
	// 	if len(app) > 0 {
	// 		fields["App"] = app
	// 	}
	// 	if len(cluster) > 0 {
	// 		fields["Cluster"] = cluster
	// 	}
	// 	if len(status) > 0 {
	// 		fields["Status"] = status
	// 	}
	// 	if err != nil {
	// 		fields["Error"] = err.Error()
	// 	}
	// 	logutils.Error(message, fields)
	// }

	var (
		appContext                 appcontext.AppContext
		certificateRequestStatuses []cmv1.CertificateRequest
		clusters                   []string
		certificateRequestList     []cmv1.CertificateRequest
	)
	// load the appContext
	_, err := appContext.LoadAppContext(contextID)
	if err != nil {
		fmt.Println("Failed to load the appContext", "", "", contextID, "", err)
		return []cmv1.CertificateRequest{}, err
	}

	// get the app instruction for 'order'
	appsOrder, err := appContext.GetAppInstruction("order")
	if err != nil {
		fmt.Println("Failed to get the app instruction for the 'order' instruction type", "", "", contextID, "", err)
		return []cmv1.CertificateRequest{}, err
	}

	var appList map[string][]string
	json.Unmarshal([]byte(appsOrder.(string)), &appList)

	for _, app := range appList["apporder"] {
		//  get all the clusters associated with the app
		clusters, err = appContext.GetClusterNames(app)
		if err != nil {
			fmt.Println("Failed to list clusters", app, "", contextID, "", err)
			return []cmv1.CertificateRequest{}, err
		}

		for _, cluster := range clusters {
			// get the resources
			resources, err := appContext.GetResourceNames(app, cluster)
			if err != nil {
				fmt.Println("Failed to get the resources", app, cluster, contextID, "", err)
				return []cmv1.CertificateRequest{}, err
			}

			// get the cluster handle
			cHandle, err := appContext.GetClusterHandle(app, cluster)
			if err != nil {
				fmt.Println("Failed to get the cluster handle", app, cluster, contextID, "", err)
				return []cmv1.CertificateRequest{}, err
			}

			// get the the cluster status handle
			sHandle, err := appContext.GetLevelHandle(cHandle, "status")
			if err != nil {
				fmt.Println("Failed to get the handle of 'status'", app, cluster, contextID, "", err)
				return []cmv1.CertificateRequest{}, err
			}

			// get the status of the cetificaterequests resource creation
			// wait for the resources to be created and available in the monitor resource bundle state
			// retry if the resources satatus are not available
			//  TODO: Confirm max retrying
			statusReady := false
			for !statusReady {
				// get the value of 'status' handle
				val, err := appContext.GetValue(sHandle)
				if err != nil {
					fmt.Println("Failed to get the value of 'status' handle", app, cluster, contextID, "", err)
					continue
				}

				s := certissuer.ResourceBundleStateStatus{}
				if err := json.Unmarshal([]byte(val.(string)), &s); err != nil {
					fmt.Println("Failed to unmarshal cluster status", app, cluster, contextID, val.(string), err)
					return []cmv1.CertificateRequest{}, err
				}

				if len(s.ResourceStatuses) == 0 {
					continue
				}

				certificateRequestStatuses = []cmv1.CertificateRequest{}
				// for each resource make sure the certificate request is created and the status is available
				for _, resource := range resources {
					for _, rStatus := range s.ResourceStatuses {
						if module.ResourceName(rStatus.Name, rStatus.Kind) == resource {
							if len(rStatus.Res) == 0 {
								logutils.Warn(fmt.Sprintf("Cluster status does not contain the certificate details for %s", rStatus.Name),
									logutils.Fields{})
								break
							}

							data, err := base64.StdEncoding.DecodeString(rStatus.Res)
							if err != nil {
								fmt.Println("Failed to decode cluster status response", app, cluster, contextID, rStatus.Res, err)
								return []cmv1.CertificateRequest{}, err
							}

							status := cmv1.CertificateRequest{}
							if err := json.Unmarshal(data, &status); err != nil {
								fmt.Println("Failed to unmarshal cluster status", app, cluster, contextID, string(data), err)
								return []cmv1.CertificateRequest{}, err
							}

							certificateRequestStatuses = append(certificateRequestStatuses, status)
							break
						}
					}
				}

				// TODO : verify the monitor bundle state should only have the resources created for the specific app context
				// the number of CR resoreces in the statuses should be equal to the number of CR resources created
				// no need to capture the resources created and validate against the statuses

				if len(resources) == len(certificateRequestStatuses) {
					logutils.Info(fmt.Sprintf("Cluster status contains the certificate for App: %s, ClusterGroup: %s and ContextID: %s", app, cluster, contextID),
						logutils.Fields{})

					certificateRequestList = append(certificateRequestList, certificateRequestStatuses...)
					// At this point we assume we have the certificate requests created and the status is available
					// no need to retry , break the loop
					statusReady = true
				}
			}

		}
	}

	return certificateRequestList, nil
}

//  THIS IS NEEDED TODO - Verify if we need to validate against the list of clusters in each cluster group
// func processCertificateRequests(contextID string, clusters []string, certificateRequests []CertificateRequest, cert string) {
// 	client := NewCertificateRequestClient()
// 	for _, cluster := range clusters {
// 		c := strings.Split(cluster, "+")
// 		crName := CertificateRequestName(contextID, cert, c[0], c[1])
// 		certReady := false
// 		for _, cr := range certificateRequests {
// 			if cr.MetaData.Name == crName {
// 				if err := client.SaveCertificateRequest(cert, c[0], c[1], cr); err != nil {
// 					return
// 				}
// 				certReady = true
// 				break
// 			}
// 		}

// 		if !certReady {
// 			fmt.Println("certificateRequest is not ready for: ", crName)
// 		}
// 	}
// }

// // CreateCertificateSecret
// func (cr *CertificateRequest) CreateCertificateSecret(name, namespace string, data map[string]string) *Secret {
// 	return CreateSecret(name, namespace, data)
// }

// validateCSR check the request is a certificate request
func validateCSR(request []byte) error {
	block, _ := pem.Decode(request)
	if block == nil ||
		block.Type != "CERTIFICATE REQUEST" {
		return errors.New("PEM block type is not CERTIFICATE REQUEST")
	}

	return nil
}