// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certmanagerissuer

import (
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"

	"context"

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

// ResourceName returns the CertificateRequest resource name, used by rsync
func ResourceName(name string) string {
	return fmt.Sprintf("%s+%s", name, "certificaterequest")
}

// CertificateRequestName retun the CertificateRequest name
func CertificateRequestName(contextID, cert, clusterProvider, cluster string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", contextID, cert, clusterProvider, cluster, "cr")
}

// CreateCertificateRequest retun the cert-manager CertificateRequest object
func CreateCertificateRequest(caCert module.CaCert, name string, request []byte) (*cmv1.CertificateRequest, error) {
	// validate the request content
	if err := validateCSR(request); err != nil {
		return nil, err
	}
	// parse certificate duration
	duration, err := time.ParseDuration(caCert.Spec.Duration)
	if err != nil {
		logutils.Error("Failed to parse the certificate duration",
			logutils.Fields{
				"Duration": caCert.Spec.Duration,
				"Error":    err.Error()})
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
			Duration: duration,
		},
		IssuerRef: cmmetav1.ObjectReference{
			Name:  caCert.Spec.IssuerRef.Name,
			Kind:  caCert.Spec.IssuerRef.Kind,
			Group: caCert.Spec.IssuerRef.Group,
		},
	}

	return cr, nil
}

// ValidateCertificateRequest validate the certificaterequest status
func ValidateCertificateRequest(cr cmv1.CertificateRequest) error {
	if len(cr.Status.Certificate) == 0 {
		err := errors.New("generated certificate is invalid")
		logutils.Error("",
			logutils.Fields{
				"CertificateRequest": cr.ObjectMeta.Name,
				"Error":              err.Error()})
		return err
	}

	if len(cr.Status.CA) == 0 {
		err := errors.New("certificate is generated by an Invalid certifcate authority")
		logutils.Error("",
			logutils.Fields{
				"CertificateRequest": cr.ObjectMeta.Name,
				"Error":              err.Error()})
		return err
	}

	approved := false
	for _, state := range cr.Status.Conditions {
		if state.Status == cmmetav1.ConditionTrue &&
			state.Type == cmv1.CertificateRequestConditionApproved {
			approved = true
			break
		}
	}

	if !approved {
		err := errors.New("the certificate is not yet approved by the CA")
		logutils.Error("",
			logutils.Fields{
				"CertificateRequest": cr.ObjectMeta.Name,
				"Error":              err.Error()})
		return err
	}
	return nil
}

// RetrieveCertificateRequests retrieves the certificaterequests created by the caCert enrollment
func RetrieveCertificateRequests(contextID string) ([]cmv1.CertificateRequest, error) {
	var (
		appContext                 appcontext.AppContext
		certificateRequestStatuses []cmv1.CertificateRequest
		clusters                   []string
		certificateRequestList     []cmv1.CertificateRequest
	)

	// load the appContext
	_, err := appContext.LoadAppContext(context.Background(), contextID)
	if err != nil {
		logutils.Error("Failed to load the appContext",
			logutils.Fields{
				"ContextID": contextID,
				"Error":     err.Error()})
		return []cmv1.CertificateRequest{}, err
	}

	// get the app instruction for 'order'
	appsOrder, err := appContext.GetAppInstruction(context.Background(), "order")
	if err != nil {
		logutils.Error("Failed to get the app instruction for the 'order' instruction type",
			logutils.Fields{
				"ContextID": contextID,
				"Error":     err.Error()})
		return []cmv1.CertificateRequest{}, err
	}

	var appList map[string][]string
	if err := json.Unmarshal([]byte(appsOrder.(string)), &appList); err != nil {
		logutils.Error("Failed to unmarshal app order",
			logutils.Fields{
				"ContextID": contextID,
				"Error":     err.Error()})
		return []cmv1.CertificateRequest{}, err
	}

	for _, app := range appList["apporder"] {
		//  get all the clusters associated with the app
		clusters, err = appContext.GetClusterNames(context.Background(), app)
		if err != nil {
			logutils.Error("Failed to get cluster names",
				logutils.Fields{
					"App":       app,
					"ContextID": contextID,
					"Error":     err.Error()})
			return []cmv1.CertificateRequest{}, err
		}

		for _, cluster := range clusters {
			// get the resources
			resources, err := appContext.GetResourceNames(context.Background(), app, cluster)
			if err != nil {
				logutils.Error("Failed to get the resource names",
					logutils.Fields{
						"App":       app,
						"Cluster":   cluster,
						"ContextID": contextID,
						"Error":     err.Error()})
				return []cmv1.CertificateRequest{}, err
			}

			// get the cluster handle
			cHandle, err := appContext.GetClusterHandle(context.Background(), app, cluster)
			if err != nil {
				logutils.Error("Failed to get the cluster handle",
					logutils.Fields{
						"App":       app,
						"Cluster":   cluster,
						"ContextID": contextID,
						"Error":     err.Error()})
				return []cmv1.CertificateRequest{}, err
			}

			// get the cluster status handle
			sHandle, err := appContext.GetLevelHandle(context.Background(), cHandle, "status")
			if err != nil {
				logutils.Error("Failed to get the handle of level 'status'",
					logutils.Fields{
						"App":       app,
						"Cluster":   cluster,
						"ContextID": contextID,
						"Error":     err.Error()})
				return []cmv1.CertificateRequest{}, err
			}

			// get the status of the cetificaterequests
			// wait for the resources to be created and available in the monitor resource bundle state
			// retry if the resources satatus are not available
			statusReady := false
			for !statusReady {
				// get the value of 'status' handle
				val, err := appContext.GetValue(context.Background(), sHandle)
				if err != nil {
					logutils.Error("Failed to get the value of 'status' handle",
						logutils.Fields{
							"App":       app,
							"Cluster":   cluster,
							"ContextID": contextID,
							"Error":     err.Error()})
					continue
				}

				s := certissuer.ResourceBundleStateStatus{}
				if err := json.Unmarshal([]byte(val.(string)), &s); err != nil {
					logutils.Error("Failed to unmarshal cetificaterequest status",
						logutils.Fields{
							"App":       app,
							"Cluster":   cluster,
							"ContextID": contextID,
							"Error":     err.Error()})
					return []cmv1.CertificateRequest{}, err
				}

				if len(s.ResourceStatuses) == 0 {
					continue
				}

				certificateRequestStatuses = []cmv1.CertificateRequest{}
				// for each resource make sure the certificaterequest is created and the status is available
				for _, resource := range resources {
					for _, rStatus := range s.ResourceStatuses {
						if module.ResourceName(rStatus.Name, rStatus.Kind) == resource {
							if len(rStatus.Res) == 0 {
								logutils.Warn(fmt.Sprintf("CetificateRequest status does not contain the certificate details for %s", rStatus.Name),
									logutils.Fields{})
								break
							}

							data, err := base64.StdEncoding.DecodeString(rStatus.Res)
							if err != nil {
								logutils.Error("Failed to decode cetificaterequest status response",
									logutils.Fields{
										"App":       app,
										"Cluster":   cluster,
										"ContextID": contextID,
										"Error":     err.Error()})
								return []cmv1.CertificateRequest{}, err
							}

							status := cmv1.CertificateRequest{}
							if err := json.Unmarshal(data, &status); err != nil {
								logutils.Error("Failed to unmarshal cetificaterequests",
									logutils.Fields{
										"App":       app,
										"Cluster":   cluster,
										"ContextID": contextID,
										"Error":     err.Error()})
								return []cmv1.CertificateRequest{}, err
							}

							certificateRequestStatuses = append(certificateRequestStatuses, status)
							break
						}
					}
				}

				if len(resources) == len(certificateRequestStatuses) {
					logutils.Info(fmt.Sprintf("CetificateRequest status contains the certificate for App: %s, ClusterGroup: %s and ContextID: %s", app, cluster, contextID),
						logutils.Fields{})

					certificateRequestList = append(certificateRequestList, certificateRequestStatuses...)
					// At this point we assume we have the certificaterequests created and the status is available
					statusReady = true
				}
			}

		}
	}

	return certificateRequestList, nil
}

// validateCSR check the request is a certificaterequest
func validateCSR(request []byte) error {
	block, _ := pem.Decode(request)
	if block == nil ||
		block.Type != "CERTIFICATE REQUEST" {
		err := errors.New("PEM block type is not CERTIFICATE REQUEST")
		logutils.Error("",
			logutils.Fields{
				"Error": err.Error()})
		return err
	}

	return nil
}
