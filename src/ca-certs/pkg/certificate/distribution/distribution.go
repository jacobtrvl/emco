// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package distribution

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	tcsv1 "github.com/intel/trusted-certificate-issuer/api/v1alpha1"
	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	clm "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer/certmanagerissuer"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/notifyclient"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"

	"context"

	v1 "k8s.io/api/core/v1"
)

const (
	AppName string = "cert-distribution"
)

// Instantiate the caCert distribution
func (ctx *DistributionContext) Instantiate() error {
	// create resources for the edge clsuters based on the issuer
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

// Update the caCert distribution appContext
func (ctx *DistributionContext) Update(prevContextID string) error {
	if err := state.UpdateAppContextStatusContextID(context.Background(), ctx.ContextID, prevContextID); err != nil {
		logutils.Error("Failed to update appContext status",
			logutils.Fields{
				"ContextID": ctx.ContextID,
				"AppName":   AppName,
				"Error":     err.Error()})
		return err
	}

	if err := notifyclient.CallRsyncUpdate(context.Background(), prevContextID, ctx.ContextID); err != nil {
		logutils.Error("Rsync update failed",
			logutils.Fields{
				"ContextID": ctx.ContextID,
				"AppName":   AppName,
				"Error":     err.Error()})
		return err
	}

	// subscribe to alerts
	stream, _, err := notifyclient.InvokeReadyNotify(context.Background(), ctx.ContextID, ctx.ClientName)
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

// Terminate the caCert distribution
func Terminate(dbKey interface{}) error {
	sc := module.NewStateClient(dbKey)
	// check the current state of the Instantiation, if any
	contextID, err := sc.VerifyState(common.Terminate)
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
// in this case, secret, clusterIssuer
func (ctx *DistributionContext) createCertManagerIssuerResources() error {
	// retrieve enrolled cert-manager resource
	resources, err := certmanagerissuer.RetrieveCertManagerResources(ctx.EnrollmentContextID)
	if err != nil {
		return err
	}

	ctx.CertificateRequests = resources.CertificateRequests
	ctx.Certificates = resources.Certificates
	ctx.Secrets = resources.Secrets

	for _, ctx.ClusterGroup = range ctx.ClusterGroups {
		// get all the clusters in this clusterGroup
		clusters, err := module.GetClusters(ctx.ClusterGroup, ctx.Project, ctx.LogicalCloud)
		if err != nil {
			return err
		}

		for _, ctx.Cluster = range clusters {
			ctx.ResOrder = []string{}
			ctx.ClusterHandle, err = ctx.AppContext.AddCluster(context.Background(), ctx.AppHandle,
				strings.Join([]string{ctx.ClusterGroup.Spec.Provider, ctx.Cluster}, "+"))
			if err != nil {
				logutils.Error("Failed to add the cluster",
					logutils.Fields{
						"Error": err.Error()})

				if er := ctx.AppContext.DeleteCompositeApp(context.Background()); er != nil {
					logutils.Error("Failed to delete the compositeApp",
						logutils.Fields{
							"ContextID": ctx.ContextID,
							"Error":     er.Error()})
					return er
				}

				return err
			}
			var sgxEnabled bool
			// check whether the cluster is SGX enabled or not
			if val, err := clm.NewClusterClient().GetClusterKvPairsValue(context.Background(), ctx.ClusterGroup.Spec.Provider, ctx.Cluster, "sgx", "enabled"); err == nil {
				v, e := module.GetValue(val)
				if e != nil {
					return e
				}

				if v == "true" {
					sgxEnabled = true
				}
			}

			ready := false
			if sgxEnabled {
				certName := certmanagerissuer.CertificateName(ctx.EnrollmentContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)
				for _, cert := range ctx.Certificates {
					if cert.ObjectMeta.Name == certName { // to make sure we are creating the resource(s) in the same cluster
						if err := certmanagerissuer.ValidateCertificate(cert); err != nil {
							return err
						}

						// create a TCSIssuer to use the secret and the certificate
						if err := ctx.createTCSIssuer(cert.Spec.SecretName, false); err != nil {
							return err
						}

						ready = true
						break
					}
				}
			} else {
				crName := certmanagerissuer.CertificateRequestName(ctx.EnrollmentContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)
				for _, cr := range ctx.CertificateRequests {
					if cr.ObjectMeta.Name == crName { // to make sure we are creating the resource(s) in the same cluster
						if err := certmanagerissuer.ValidateCertificateRequest(cr); err != nil {
							return err
						}

						// create a Secret to store the certificate
						sName := certmanagerissuer.SecretName(ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)
						if err := ctx.createSecret(cr, sName, "cert-manager"); err != nil {
							return err
						}

						// create a ClusterIssuer to use the secret and the certificate
						if err := ctx.createClusterIssuer(sName); err != nil {
							return err
						}

						ready = true
						break
					}
				}
			}

			if !ready {
				err := errors.New("cert-manager resource is not ready. Update the enrollment")
				logutils.Error("",
					logutils.Fields{
						"Cluster": ctx.Cluster,
						"Error":   err.Error()})
				return err
			}

			// create service specific resources for this issuer
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

// createServiceResources creates the resources based on the service mesh type
func (ctx *DistributionContext) createServiceResources() error {
	var serviceType string
	val, err := clm.NewClusterClient().GetClusterKvPairsValue(context.Background(), ctx.ClusterGroup.Spec.Provider, ctx.Cluster, "serviceMeshInfo", "serviceType")
	if err == nil {
		v, e := module.GetValue(val)
		if e != nil {
			return e
		}

		serviceType = v
	}

	if err != nil &&
		err.Error() != "Cluster key value pair not found" &&
		err.Error() != "Cluster KV pair key value not found" {
		return err
	}

	switch serviceType {
	case "istio":
		return ctx.createIstioServiceResources()

	default:
		return ctx.createIstioServiceResources()
	}

}

// createIstioServiceResources creates the istio resources on the edge cluster
// in this case, proxyConfig, kncc resource to update the istio configMap
func (ctx *DistributionContext) createIstioServiceResources() error {
	switch ctx.CaCert.Spec.IssuerRef.Group {
	case "cert-manager.io":
		var sgxEnabled bool
		var issuerName string
		var secret v1.Secret
		// check whether the cluster is SGX enabled or not
		if val, err := clm.NewClusterClient().GetClusterKvPairsValue(context.Background(), ctx.ClusterGroup.Spec.Provider, ctx.Cluster, "sgx", "enabled"); err == nil {
			v, e := module.GetValue(val)
			if e != nil {
				return e
			}

			if v == "true" {
				sgxEnabled = true
			}
		}

		if sgxEnabled {
			// retrieve the tcsissuer
			if issuer := ctx.retrieveTCSIssuer(ctx.Cluster); !reflect.DeepEqual(issuer, tcsv1.TCSIssuer{}) {
				issuerName = issuer.ObjectMeta.Name
			}
			// retrieve the tcsissuer secret
			sName := certmanagerissuer.SecretName(ctx.EnrollmentContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)
			for _, s := range ctx.Secrets {
				if s.ObjectMeta.Name == sName {
					secret = s
					break
				}
			}
		} else {
			// retrieve the clusterissuer
			if issuer := ctx.retrieveClusterIssuer(ctx.Cluster); !reflect.DeepEqual(issuer, cmv1.ClusterIssuer{}) {
				issuerName = issuer.ObjectMeta.Name
			}
			// retrieve the clusterissuer secret
			secret = *ctx.retrieveSecret(ctx.Cluster)

		}

		if len(issuerName) == 0 {
			// a clusterIssuer is not available, return error
			err := errors.New("A clusterIssuer is not available for the cluster")
			if sgxEnabled {
				err = errors.New("A tcsIssuer is not available for the cluster")
			}
			logutils.Error("",
				logutils.Fields{
					"Cluster": ctx.Cluster,
					"Error":   err.Error()})
			return err
		}

		if reflect.DeepEqual(secret, v1.Secret{}) {
			// a secret is not available, return error
			err := errors.New("A secret is not available for the cluster")
			logutils.Error("",
				logutils.Fields{
					"Cluster": ctx.Cluster,
					"Error":   err.Error()})
			return err
		}

		var (
			key,
			value string
		)

		if len(ctx.Namespace) > 0 {
			// create a proxyconfig for this namespace
			if err := ctx.createProxyConfig(issuerName); err != nil {
				return err
			}
		}

		// create a kncc patch config using the secret created fot this issuer
		key = "mesh:\n  caCertificates\n"
		if sgxEnabled {
			pem := strings.Replace(string(bytes.Join([][]byte{secret.Data[v1.TLSCertKey], secret.Data[v1.ServiceAccountRootCAKey]},
				[]byte{})), "\n", "\n    ", -1) // create the caCert chain
			value = fmt.Sprintf("- certSigners:\n  - tcsissuer.tcs.intel.com/default.%s\n  pem: |\n    %s",
				issuerName, pem)
			if len(ctx.Namespace) > 0 {
				// set the issuer name with the namespace
				value = fmt.Sprintf("- certSigners:\n  - tcsissuer.tcs.intel.com/%s.%s\n  pem: |\n    %s",
					ctx.Namespace, issuerName, pem)
			}

		} else {
			pem := strings.Replace(string(secret.Data[v1.TLSCertKey]), "\n", "\n    ", -1)
			value = fmt.Sprintf("- certSigners:\n  - clusterissuers.cert-manager.io/%s\n  pem: |\n    %s",
				issuerName, pem)
		}

		if err := ctx.createKnccConfig("istio-system", "istio", "istio-system",
			[]map[string]string{{key: value}}); err != nil {
			return err
		}

		if len(ctx.Namespace) == 0 {
			// use this issuer as the dfault issuer for the cluster
			key = "mesh:\n  defaultConfig:\n    proxyMetadata:\n      ISTIO_META_CERT_SIGNER\n"
			value = issuerName
			if sgxEnabled {
				value = fmt.Sprintf("default.%s", issuerName)
			}
			if err := ctx.createKnccConfig("istio-system", "istio", "istio-system",
				[]map[string]string{{key: value}}); err != nil {
				return err
			}
		}

		return nil
	}

	err := errors.New("Unsupported Issuer")
	logutils.Error("",
		logutils.Fields{
			"Cluster": ctx.Cluster,
			"Error":   err.Error()})
	return err
}
