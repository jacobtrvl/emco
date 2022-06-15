// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package enrollment

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer/certmanagerissuer"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	clm "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/notifyclient"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/readynotify"
)

const AppName string = "cert-enrollment"

// Instantiate
func (ctx *EnrollmentContext) Instantiate() error {
	for _, ctx.ClusterGroup = range ctx.ClusterGroups {
		// get all the clusters in this cluster group
		clusters, err := module.GetClusters(ctx.ClusterGroup)
		if err != nil {
			return err
		}
		// Enrollment involve creating the intermdiate certificate for each cluster in the cluster group
		// Create a certificate request for each cluster in the clustergroup
		// The resources required to generate the certificate may vary based on the issuer type
		for _, ctx.Cluster = range clusters {
			// create resources for each clsuters based on the issuer
			switch ctx.CaCert.Spec.IssuerRef.Group {
			case "cert-manager.io":
				if err := ctx.createCertManagerResources(); err != nil {
					return err
				}

			default:
				fmt.Println("Unsupported Issuer")
			}
		}
	}

	return nil
}

// Status
func Status(stateInfo state.StateInfo, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (module.CaCertStatus, error) {
	//	status, err := status.PrepareCertEnrollmentStatusResult(stateInfo, "ready")
	statusResult, err := status.PrepareCaCertStatusResult(stateInfo, qInstance, qType, qOutput, fApps, fClusters, fResources)
	if err != nil {
		fmt.Println(err.Error())
	}

	caCertStatus := module.CaCertStatus{}
	caCertStatus.Name = statusResult.Name
	caCertStatus.State = statusResult.State
	caCertStatus.DeployedStatus = statusResult.DeployedStatus
	caCertStatus.ReadyStatus = statusResult.ReadyStatus
	caCertStatus.DeployedCounts = statusResult.DeployedCounts
	caCertStatus.ReadyCounts = statusResult.ReadyCounts
	caCertStatus.Clusters = statusResult.Clusters

	return caCertStatus, nil
}

// Terminate
func (ctx *EnrollmentContext) Terminate() error {
	for _, ctx.ClusterGroup = range ctx.ClusterGroups {
		// get all the clusters in this cluster group
		clusters, err := module.GetClusters(ctx.ClusterGroup)
		if err != nil {
			return err
		}
		// delete all the resources associated with enrollment instantiation
		for _, ctx.Cluster = range clusters {
			// delete the primary key
			// TODO: Verify the return on errors
			if err := ctx.deletePrivateKey(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Update
func (context *EnrollmentContext) Update(contextID string) error {

	// initialize the Instantiation
	if err := context.Instantiate(); err != nil {
		return err
	}

	if err := state.UpdateAppContextStatusContextID(context.ContextID, contextID); err != nil {
		return err
	}

	if err := notifyclient.CallRsyncUpdate(contextID, context.ContextID); err != nil {
		return err
	}

	// subscribe to alerts
	stream, _, err := notifyclient.InvokeReadyNotify(context.ContextID, context.ClientName)
	if err != nil {
		fmt.Println("Failed to subscribe to alerts from the rsync gRPC server", context.ContextID, err)
		return err
	}

	if err := stream.CloseSend(); err != nil {
		fmt.Println("Failed to close the send stream", context.ContextID, err)
		return err
	}

	return nil
}

// IssuingClusterHandle
func (ctx *EnrollmentContext) IssuingClusterHandle() (handle interface{}, err error) {
	// add handle for the issuing cluster
	handle, err = ctx.AppContext.AddCluster(ctx.AppHandle,
		strings.Join([]string{ctx.CaCert.Spec.IssuingCluster.ClusterProvider, ctx.CaCert.Spec.IssuingCluster.Cluster}, "+"))
	if err != nil {
		ctx.AppContext.DeleteCompositeApp()
		fmt.Println(err)

	}
	return handle, err
}

// ValidateEnrollment
func (ctx *EnrollmentContext) ValidateEnrollment(stream readynotify.ReadyNotify_AlertClient, client readynotify.ReadyNotifyClient) {
	contextID := module.RetrieveAppContext(stream, client)

	switch ctx.CaCert.Spec.IssuerRef.Group {
	case "cert-manager.io":
		certmanagerissuer.RetrieveCertificateRequests(contextID)

	default:
		fmt.Println("Unsupported Issuer")

	}

	if _, err := client.Unsubscribe(context.Background(), &readynotify.Topic{ClientName: ctx.ClientName, AppContext: contextID}); err != nil {
		logutils.Error("[ReadyNotify gRPC] Failed to unsubscribe to alerts",
			logutils.Fields{"ContextID": contextID,
				"Error": err.Error()})
	}
}

// VerifyEnrollmentState
func VerifyEnrollmentState(stateInfo state.StateInfo) (enrollmentContextID string, err error) {
	// get the cert enrollemnt instantiation state
	enrollmentContextID = state.GetLastContextIdFromStateInfo(stateInfo)
	if len(enrollmentContextID) == 0 {
		return "", errors.New("enrollment is not completed")
	}

	status, err := state.GetAppContextStatus(enrollmentContextID)
	if err != nil {
		return "", err
	}

	if status.Status != appcontext.AppContextStatusEnum.Instantiated &&
		status.Status != appcontext.AppContextStatusEnum.Updated {
		return "", errors.New("enrollment is not completed")
	}

	return enrollmentContextID, err
}

// ValidateEnrollmentStatus
func ValidateEnrollmentStatus(stateInfo state.StateInfo) (readyCount int, err error) {
	//  verify the status of the enrollemnt
	certEnrollmentStatus, err := Status(stateInfo, "", "ready", "all", make([]string, 0), make([]string, 0), make([]string, 0))
	if err != nil {
		return readyCount, err
	}

	if strings.ToLower(string(certEnrollmentStatus.DeployedStatus)) != "instantiated" {
		return readyCount, errors.New("Enrollment is not ready")
	}
	if strings.ToLower(certEnrollmentStatus.ReadyStatus) != "ready" {
		return readyCount, errors.New("Enrollment is not ready")
	}

	return certEnrollmentStatus.ReadyCounts["Ready"], nil
}

func (ctx *EnrollmentContext) createCertManagerResources() error {
	// This needs to be a unique name for each cluster
	ctx.CaCert.Spec.CertificateSigningInfo.Subject.Names.CommonName = strings.Join([]string{ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster, "ca"}, "-")
	// check if a cluster specific commonName is available
	// TODO: kvpair naming
	if val, err := clm.NewClusterClient().GetClusterKvPairsValue(ctx.ClusterGroup.Spec.Provider, ctx.Cluster, "csrkvpairs", "commonName"); err == nil {
		ctx.CaCert.Spec.CertificateSigningInfo.Subject.Names.CommonName = val.(string)
	}

	// generate the private key for the csr
	pemBlock, err := certificate.GeneratePrivateKey(ctx.CaCert.Spec.CertificateSigningInfo.KeySize)
	if err != nil {
		return err
	}

	// parse the RSA key in PKCS #1, ASN.1 DER form
	pk, err := certificate.ParsePrivateKey(pemBlock.Bytes)
	if err != nil {
		return err
	}

	// create a certificate signing request
	request, err := ctx.createCertificateSigningRequest(pk)
	if err != nil {
		return err
	}

	// create the cert-manager CertificateRequest resource
	// TODO: this needs to be a unique name, check the format
	name := certmanagerissuer.CertificateRequestName(ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)
	cr, err := certmanagerissuer.CreateCertificateRequest(ctx.CaCert, name, request)
	if err != nil {
		return err
	}

	// add the CertificateRequest resource in to the app context
	if err := module.AddResource(ctx.AppContext, cr, ctx.IssuerHandle, module.ResourceName(cr.ObjectMeta.Name, cr.TypeMeta.Kind)); err != nil {
		return err
	}

	ctx.ResOrder = append(ctx.ResOrder, module.ResourceName(cr.ObjectMeta.Name, cr.TypeMeta.Kind))

	// save the PK in mongo
	ctx.savePrivateKey(name, base64.StdEncoding.EncodeToString(pem.EncodeToMemory(pemBlock)))

	return nil
}

// createCertificateSigningRequest
func (ctx *EnrollmentContext) createCertificateSigningRequest(pk *rsa.PrivateKey) ([]byte, error) {
	return certificate.CreateCertificateSigningRequest(x509.CertificateRequest{
		Version:            ctx.CaCert.Spec.CertificateSigningInfo.Version,
		SignatureAlgorithm: certificate.SignatureAlgorithm(ctx.CaCert.Spec.CertificateSigningInfo.Algorithm.SignatureAlgorithm),
		PublicKeyAlgorithm: certificate.PublicKeyAlgorithm(ctx.CaCert.Spec.CertificateSigningInfo.Algorithm.PublicKeyAlgorithm),
		Subject: pkix.Name{
			Country:            ctx.CaCert.Spec.CertificateSigningInfo.Subject.Locale.Country,
			Locality:           ctx.CaCert.Spec.CertificateSigningInfo.Subject.Locale.Locality,
			PostalCode:         ctx.CaCert.Spec.CertificateSigningInfo.Subject.Locale.PostalCode,
			Province:           ctx.CaCert.Spec.CertificateSigningInfo.Subject.Locale.Province,
			StreetAddress:      ctx.CaCert.Spec.CertificateSigningInfo.Subject.Locale.StreetAddress,
			CommonName:         ctx.CaCert.Spec.CertificateSigningInfo.Subject.Names.CommonName,
			Organization:       ctx.CaCert.Spec.CertificateSigningInfo.Subject.Organization.Names,
			OrganizationalUnit: ctx.CaCert.Spec.CertificateSigningInfo.Subject.Organization.Units},
		DNSNames:       ctx.CaCert.Spec.CertificateSigningInfo.DNSNames,
		EmailAddresses: ctx.CaCert.Spec.CertificateSigningInfo.EmailAddresses}, pk)
}

// savePrivateKey
func (ctx *EnrollmentContext) savePrivateKey(name, val string) error {
	dbKey := module.DBKey{
		Cert:            ctx.CaCert.MetaData.Name,
		Cluster:         ctx.Cluster,
		ClusterProvider: ctx.ClusterGroup.Spec.Provider,
		ContextID:       ctx.ContextID}
	key := module.Key{
		Name: name,
		Val:  val}

	return module.NewKeyClient(dbKey).Save(key)
}

// deletePrivateKey
func (ctx *EnrollmentContext) deletePrivateKey() error {
	dbKey := module.DBKey{
		Cert:            ctx.CaCert.MetaData.Name,
		Cluster:         ctx.Cluster,
		ClusterProvider: ctx.ClusterGroup.Spec.Provider,
		ContextID:       ctx.ContextID}

	return module.NewKeyClient(dbKey).Delete()
}

// TODO: remove this
func TestValidateEnrollment(cert, contextID string) {
	certmanagerissuer.RetrieveCertificateRequests(contextID)
}
