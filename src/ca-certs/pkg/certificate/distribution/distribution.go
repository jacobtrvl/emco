// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package distribution

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	clm "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer/certmanagerissuer"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/notifyclient"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
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
		fmt.Println("Unsupported Issuer")
	}

	return nil
}

// Update the certificate distribution app context
func (ctx *DistributionContext) Update(prevContextID string) error {
	if err := state.UpdateAppContextStatusContextID(ctx.ContextID, prevContextID); err != nil {
		return err
	}

	if err := notifyclient.CallRsyncUpdate(prevContextID, ctx.ContextID); err != nil {
		return err
	}

	// subscribe to alerts
	stream, _, err := notifyclient.InvokeReadyNotify(ctx.ContextID, ctx.ClientName)
	if err != nil {
		fmt.Println("Failed to subscribe to alerts from the rsync gRPC server", ctx.ContextID, err)
		return err
	}

	if err := stream.CloseSend(); err != nil {
		fmt.Println("Failed to close the send stream", ctx.ContextID, err)
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

	// call resource synchronizer to delete the resources under this app context
	ctx := module.CertAppContext{
		ContextID: contextID}
	if err := ctx.CallRsyncUninstall(); err != nil {
		return err
	}

	// update the state object for the cert distribution resource
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

	// TODO: Verify the logic
	// Shoudl we check the edge cluster issuer type here, like we check for service type?
	for _, ctx.ClusterGroup = range ctx.ClusterGroups {
		// get all the clusters in this cluster group
		clusters, err := module.GetClusters(ctx.ClusterGroup)
		if err != nil {
			return err
		}

		for _, ctx.Cluster = range clusters {
			ctx.ResOrder = []string{}
			ctx.ClusterHandle, err = ctx.AppContext.AddCluster(ctx.AppHandle,
				strings.Join([]string{ctx.ClusterGroup.Spec.Provider, ctx.Cluster}, "+"))
			if err != nil {
				return err
			}

			available := false

			// TODO: this needs to be a unique name, check the format
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

			// TODO : verify the logic here
			if !available {
				return errors.New("certificate request is not ready for cluster %s. Update the enrollment")
			}

			// Create service specific resources for this issuer
			ctx.createServiceResources()

			if err := module.AddInstruction(ctx.AppContext, ctx.ClusterHandle, ctx.ResOrder); err != nil {
				return err
			}
		}
	}

	return nil
}

// createServiceResources
func (ctx *DistributionContext) createServiceResources() error {
	var serviceType string = "istio"
	// TODO: change the naming
	val, err := clm.NewClusterClient().GetClusterKvPairsValue(ctx.ClusterGroup.Spec.Provider, ctx.Cluster, "csrkvpairs", "commonName")
	if err == nil {
		serviceType = val.(string)
	}

	// TODO: Confirm should we return from here or not
	if err != nil &&
		err.Error() != "Cluster key value pair not found" &&
		err.Error() != "Cluster KV pair key value not found" {
		return err
	}

	switch serviceType {
	case "istio":
		ctx.createIstioServiceResourcess()

	default:
		ctx.createIstioServiceResourcess()
	}

	return nil
}

// createIstioServiceResourcess
func (ctx *DistributionContext) createIstioServiceResourcess() error {
	switch ctx.CaCert.Spec.IssuerRef.Group {
	case "cert-manager.io":
		if issuer := ctx.retrieveClusterIssuer(ctx.Cluster); !reflect.DeepEqual(issuer, cmv1.ClusterIssuer{}) {
			if err := ctx.createProxyConfig(issuer); err != nil {
				return err
			}
		} else {
			return errors.New("Unsupported Issuer")
		}

	default:
		return errors.New("Unsupported Issuer")
	}

	return nil
}

// TODO: Remove this
func TestValidateDistribution(contextID string) {
	var (
		appContext appcontext.AppContext
	)
	// load the appContext
	_, err := appContext.LoadAppContext(contextID)
	// TODO: Confirm return err or false? Shoudl we retry in case of appContext errors
	if err != nil {
		fmt.Println(err)
		return
	}

	// get the app instruction for 'order'
	appsOrder, err := appContext.GetAppInstruction("order")
	if err != nil {
		fmt.Println(err)
		return
	}

	var appList map[string][]string
	json.Unmarshal([]byte(appsOrder.(string)), &appList)

	for _, app := range appList["apporder"] {
		//  get all the clusters associated with the app
		clusters, err := appContext.GetClusterNames(app)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, cluster := range clusters {
			// get the resources
			resources, err := appContext.GetResourceNames(app, cluster)
			if err != nil {
				fmt.Println(err)
				return
			}

			// get the cluster handle
			cHandle, err := appContext.GetClusterHandle(app, cluster)
			if err != nil {
				fmt.Println(err)
				return
			}

			// get the the cluster status handle
			sHandle, err := appContext.GetLevelHandle(cHandle, "status")
			if err != nil {
				fmt.Println(err)
				return
			}

			val, err := appContext.GetValue(sHandle)
			if err != nil {
				continue
			}

			s := certissuer.ResourceBundleStateStatus{}
			if err := json.Unmarshal([]byte(val.(string)), &s); err != nil {
				fmt.Println(err)
				return
			}

			if len(s.ResourceStatuses) == 0 {
				continue
			}

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
							fmt.Println(err)
							return

						}

						fmt.Println(string(data))
						break
					}
				}
			}

		}
	}
}
