// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package enrollment

import (
	"strings"

	"github.com/pkg/errors"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	clm "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"

	"context"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/notifyclient"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"
)

const AppName string = "cert-enrollment"

// Instantiate the caCert enrollment
func (ctx *EnrollmentContext) Instantiate() error {
	for _, ctx.ClusterGroup = range ctx.ClusterGroups {
		// get all the clusters in this clusterGroup
		clusters, err := module.GetClusters(ctx.ClusterGroup, ctx.Project, ctx.LogicalCloud)
		if err != nil {
			return err
		}

		for _, ctx.Cluster = range clusters {
			// create resources for the edge clsuters based on the issuer
			switch ctx.CaCert.Spec.IssuerRef.Group {
			case "cert-manager.io":
				// check whether the cluster is sgx enabled or not
				// if it's an sgx capable cluster, create a cert-manager certificate resource
				// otherwise, create a cert-manger certificaterequest resource
				if val, err := clm.NewClusterClient().GetClusterKvPairsValue(context.Background(), ctx.ClusterGroup.Spec.Provider, ctx.Cluster, "sgx", "enabled"); err == nil {
					v, e := module.GetValue(val)
					if e != nil {
						return e
					}

					if v == "true" { // sgx enabled
						if err := ctx.createCertManagerCertificate(); err != nil {
							return err
						}

						continue // with the next cluster
					}
				}

				// cluster is not SGX enabled, create a certificaterequest
				if err := ctx.createCertManagerCertificateRequest(); err != nil {
					return err
				}

			default:
				err := errors.New("unsupported Issuer")
				logutils.Error("",
					logutils.Fields{
						"Issuer": ctx.CaCert.Spec.IssuerRef.Group,
						"Error":  err.Error()})
				return err
			}

			ctx.Cluster = ""
		}

		ctx.ClusterGroup = module.ClusterGroup{}
	}

	return nil
}

// Status provides the caCert enrollment status
func Status(stateInfo state.StateInfo, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (module.CaCertStatus, error) {
	statusResult, err := status.PrepareCaCertStatusResult(context.Background(), stateInfo, qInstance, qType, qOutput, fApps, fClusters, fResources)
	if err != nil {
		logutils.Error("Failed to get the enrollemnt status",
			logutils.Fields{
				"Error": err.Error()})
		return module.CaCertStatus{}, err
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

// Terminate the caCert enrollment
func (ctx *EnrollmentContext) Terminate() error {
	for _, ctx.ClusterGroup = range ctx.ClusterGroups {
		// get all the clusters in this clusterGoup
		clusters, err := module.GetClusters(ctx.ClusterGroup, ctx.Project, ctx.LogicalCloud)
		if err != nil {
			return err
		}
		// delete all the resources associated with enrollment
		for _, ctx.Cluster = range clusters {
			// delete the primary key, if it exists
			if ctx.privateKeyExists() {
				if err := ctx.deletePrivateKey(); err != nil {
					return err
				}
			}

			ctx.Cluster = ""
		}

		ctx.ClusterGroup = module.ClusterGroup{}
	}

	return nil
}

// Update the caCert enrollment
func (ctx *EnrollmentContext) Update(contextID string) error {
	// initialize the Instantiation
	if err := ctx.Instantiate(); err != nil {
		return err
	}

	// add instruction under the given handle and type
	if err := module.AddInstruction(ctx.AppContext, ctx.IssuerHandle, ctx.ResOrder); err != nil {
		return err
	}

	if err := state.UpdateAppContextStatusContextID(context.Background(), ctx.ContextID, contextID); err != nil {
		logutils.Error("Failed to update appContext status",
			logutils.Fields{
				"ContextID": ctx.ContextID,
				"AppName":   AppName,
				"Error":     err.Error()})
		return err
	}

	if err := notifyclient.CallRsyncUpdate(context.Background(), contextID, ctx.ContextID); err != nil {
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

// IssuingClusterHandle returns the handle of certificate issuing cluster
func (ctx *EnrollmentContext) IssuingClusterHandle() (interface{}, error) {
	var (
		handle interface{}
		err    error
	)

	// add handle for the issuing cluster
	handle, err = ctx.AppContext.AddCluster(context.Background(), ctx.AppHandle,
		strings.Join([]string{ctx.CaCert.Spec.IssuingCluster.ClusterProvider, ctx.CaCert.Spec.IssuingCluster.Cluster}, "+"))
	if err != nil {
		logutils.Error("Failed to add the issuing cluster",
			logutils.Fields{
				"Error": err.Error()})

		if er := ctx.AppContext.DeleteCompositeApp(context.Background()); er != nil {
			logutils.Error("Failed to delete the compositeApp",
				logutils.Fields{
					"ContextID": ctx.ContextID,
					"Error":     er.Error()})
			return handle, er
		}

		return handle, err

	}

	return handle, err
}

// VerifyEnrollmentState verify the caCert enrollment state
func VerifyEnrollmentState(stateInfo state.StateInfo) (enrollmentContextID string, err error) {
	// get the caCert enrollemnt instantiation state
	enrollmentContextID = state.GetLastContextIdFromStateInfo(stateInfo)
	if len(enrollmentContextID) == 0 {
		err := errors.New("enrollment is not completed")
		logutils.Error("",
			logutils.Fields{
				"Error": err.Error()})
		return "", err
	}

	status, err := state.GetAppContextStatus(context.Background(), enrollmentContextID)
	if err != nil {
		logutils.Error("Failed to get the appContext status",
			logutils.Fields{
				"ContextID": enrollmentContextID,
				"Error":     err.Error()})
		return "", err
	}

	if status.Status != appcontext.AppContextStatusEnum.Instantiated &&
		status.Status != appcontext.AppContextStatusEnum.Updated {
		err := errors.New("enrollment is not completed")
		logutils.Error("",
			logutils.Fields{
				"Status": status.Status,
				"Error":  err.Error()})
		return "", err
	}

	return enrollmentContextID, err
}

// ValidateEnrollmentStatus verify the caCert enrollment status
func ValidateEnrollmentStatus(stateInfo state.StateInfo) (readyCount int, err error) {
	//  verify the status of the enrollemnt
	certEnrollmentStatus, err := Status(stateInfo, "", "ready", "all", make([]string, 0), make([]string, 0), make([]string, 0))
	if err != nil {
		return readyCount, err
	}

	if strings.ToLower(string(certEnrollmentStatus.DeployedStatus)) != "instantiated" {
		err := errors.New("enrollment is not ready")
		logutils.Error("",
			logutils.Fields{
				"DeployedStatus": certEnrollmentStatus.DeployedStatus,
				"Error":          err.Error()})
		return readyCount, err
	}
	if strings.ToLower(certEnrollmentStatus.ReadyStatus) != "ready" {
		err := errors.New("enrollment is not ready")
		logutils.Error("",
			logutils.Fields{
				"ReadyStatus": certEnrollmentStatus.ReadyStatus,
				"Error":       err.Error()})
		return readyCount, err
	}

	return certEnrollmentStatus.ReadyCounts["Ready"], nil
}
