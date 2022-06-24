// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package distribution

import (
	"fmt"
	"reflect"
	"strings"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	clm "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer/certmanagerissuer"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/notifyclient"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"

	v1 "k8s.io/api/core/v1"
)

const (
	AppName string = "cert-distribution"
)

// Instantiate the certificate distribution
func (ctx *DistributionContext) Instantiate() error {
	// check the certificate issuer
	switch ctx.CaCert.Spec.IssuerRef.Group {
	case "cert-manager.io":
		return ctx.createCertManagerIssuerResources()

	default:
		err := errors.New("unsupported Issuer")
		logutils.Error("",
			logutils.Fields{
				"Issuer": ctx.CaCert.Spec.IssuerRef.Group,
				"Error":  err.Error()})
		return err
	}

}

// Update the certificate distribution appContext
func (ctx *DistributionContext) Update(prevContextID string) error {
	if err := state.UpdateAppContextStatusContextID(ctx.ContextID, prevContextID); err != nil {
		logutils.Error("Failed to update appContext status",
			logutils.Fields{
				"ContextID": ctx.ContextID,
				"AppName":   AppName,
				"Error":     err.Error()})
		return err
	}

	if err := notifyclient.CallRsyncUpdate(prevContextID, ctx.ContextID); err != nil {
		logutils.Error("Rsync update failed",
			logutils.Fields{
				"ContextID": ctx.ContextID,
				"AppName":   AppName,
				"Error":     err.Error()})
		return err
	}

	// subscribe to alerts
	stream, _, err := notifyclient.InvokeReadyNotify(ctx.ContextID, ctx.ClientName)
	if err != nil {
		logutils.Error("Failed to subscribe to alerts",
			logutils.Fields{
				"ContextID":  ctx.ContextID,
				"ClientName": ctx.ClientName,
				"AppName":    AppName,
				"Error":      err.Error()})
		return err
	}

	if err := stream.CloseSend(); err != nil {
		logutils.Error("Failed to close the send stream",
			logutils.Fields{
				"ContextID":  ctx.ContextID,
				"ClientName": ctx.ClientName,
				"AppName":    AppName,
				"Error":      err.Error()})
		return err
	}

	return nil
}

// Terminate the certificate distribution
func Terminate(dbKey interface{}) error {
	sc := module.NewStateClient(dbKey)
	// check the current state of the Instantiation, if any
	contextID, err := sc.VerifyState(module.TerminateEvent)
	if err != nil {
		return err
	}

	// call resource synchronizer to delete the resources under this appContext
	ctx := module.CaCertAppContext{
		ContextID: contextID}
	if err := ctx.CallRsyncUninstall(); err != nil {
		return err
	}

	// update the state object for the caCert distribution resource
	if err := sc.Update(state.StateEnum.Terminated, contextID, false); err != nil {
		return err
	}

	return nil
}

// createCertManagerIssuerResources creates cert-manager specific resources
func (ctx *DistributionContext) createCertManagerIssuerResources() error {
	// retrieve enrolled CertificateRequests
	crs, err := certmanagerissuer.RetrieveCertificateRequests(ctx.EnrollmentContextID)
	if err != nil {
		return err
	}

	ctx.CertificateRequests = crs

	for _, ctx.ClusterGroup = range ctx.ClusterGroups {
		// get all the clusters in this clusterGroup
		clusters, err := module.GetClusters(ctx.ClusterGroup, ctx.Project, ctx.LogicalCloud)
		if err != nil {
			return err
		}

		for _, ctx.Cluster = range clusters {
			ctx.ResOrder = []string{}
			ctx.ClusterHandle, err = ctx.AppContext.AddCluster(ctx.AppHandle,
				strings.Join([]string{ctx.ClusterGroup.Spec.Provider, ctx.Cluster}, "+"))
			if err != nil {
				logutils.Error("Failed to add the cluster",
					logutils.Fields{
						"Error": err.Error()})

				if er := ctx.AppContext.DeleteCompositeApp(); er != nil {
					logutils.Error("Failed to delete the compositeApp",
						logutils.Fields{
							"ContextID": ctx.ContextID,
							"Error":     er.Error()})
					return er
				}

				return err
			}

			available := false

			crName := certmanagerissuer.CertificateRequestName(ctx.EnrollmentContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)
			for _, cr := range ctx.CertificateRequests {
				if cr.ObjectMeta.Name == crName { // to make sure we are creating the resource(s) in the same cluster
					if err := certmanagerissuer.ValidateCertificateRequest(cr); err != nil {
						return err
					}

					// Create a Secret
					sName := certmanagerissuer.SecretName(ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)
					if err := ctx.createSecret(cr, sName, "cert-manager"); err != nil {
						return err
					}

					// Create the ClusterIssuer uisng the same secret
					if err := ctx.createClusterIssuer(sName); err != nil {
						return err
					}

					available = true
					break
				}
			}

			if !available {
				err := errors.New("certificate request is not ready for cluster. Update the enrollment")
				logutils.Error("",
					logutils.Fields{
						"Cluster": ctx.Cluster,
						"Error":   err.Error()})
				return err
			}

			// Create service specific resources for this issuer
			if err := ctx.createServiceResources(); err != nil {
				return err
			}

			if err := module.AddInstruction(ctx.AppContext, ctx.ClusterHandle, ctx.ResOrder); err != nil {
				return err
			}

			ctx.Cluster = ""
			ctx.ResOrder = []string{}
			ctx.ClusterHandle = nil
		}

		ctx.ClusterGroup = module.ClusterGroup{}
	}

	return nil
}

// createServiceResources
func (ctx *DistributionContext) createServiceResources() error {
	var serviceType string
	val, err := clm.NewClusterClient().GetClusterKvPairsValue(ctx.ClusterGroup.Spec.Provider, ctx.Cluster, "serviceMeshInfo", "serviceType")
	if err == nil {
		serviceType = val.(string)
	}

	if err != nil &&
		err.Error() != "Cluster key value pair not found" &&
		err.Error() != "Cluster KV pair key value not found" {
		return err
	}

	switch serviceType {
	case "istio":
		return ctx.createIstioServiceResourcess()

	default:
		return ctx.createIstioServiceResourcess()
	}

}

// createIstioServiceResourcess
func (ctx *DistributionContext) createIstioServiceResourcess() error {
	switch ctx.CaCert.Spec.IssuerRef.Group {
	case "cert-manager.io":
		if issuer := ctx.retrieveClusterIssuer(ctx.Cluster); !reflect.DeepEqual(issuer, cmv1.ClusterIssuer{}) {
			if len(ctx.Namespace) > 0 {
				// create a proxyconfig for this namespace
				if err := ctx.createProxyConfig(issuer); err != nil {
					return err
				}
			}

			// create a kncc patch config using the secret created fot this issuer
			if s := ctx.retrieveSecret(ctx.Cluster); !reflect.DeepEqual(s, v1.Secret{}) {
				pem := strings.Replace(string(s.Data[v1.TLSCertKey]), "\n", "\n    ", -1)
				key := "mesh:\n  caCertificates\n"
				value := fmt.Sprintf("- certSigners:\n  - clusterissuers.cert-manager.io/%s\n  pem: |\n    %s",
					issuer.ObjectMeta.Name, pem)
				if err := ctx.createKnccConfig("istio-system", "istio", "istio-system",
					[]map[string]string{{key: value}}); err != nil {
					return err
				}

				if len(ctx.Namespace) == 0 {
					// use this issuer as the dfault issuer for the cluster
					// TODO: This needs kncc support

				}

				return nil
			}

			// A secret is not available, return error
			err := errors.New("A secret is not available for the cluster")
			logutils.Error("",
				logutils.Fields{
					"Cluster": ctx.Cluster,
					"Error":   err.Error()})
			return err
		}

		// A clusterIssuer is not available, return error
		err := errors.New("A clusterIssuer is not available for cluster")
		logutils.Error("",
			logutils.Fields{
				"Cluster": ctx.Cluster,
				"Error":   err.Error()})
		return err

	}

	err := errors.New("Unsupported Issuer")
	logutils.Error("",
		logutils.Fields{
			"Cluster": ctx.Cluster,
			"Error":   err.Error()})
	return err
}
