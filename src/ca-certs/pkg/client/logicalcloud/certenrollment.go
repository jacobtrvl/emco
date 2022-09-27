// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import (
	"context"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate/enrollment"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	dcm "gitlab.com/project-emco/core/emco-base/src/dcm/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
)

const clientName string = "cacert"

// CaCertEnrollmentManager exposes all the caCert enrollment functionalities
type CaCertEnrollmentManager interface {
	Instantiate(cert, project string) error
	Status(cert, project, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (module.CaCertStatus, error)
	Terminate(cert, project string) error
	Update(cert, project string) error
}

// CaCertEnrollmentClient holds the client properties
type CaCertEnrollmentClient struct {
}

// NewCaCertEnrollmentClient returns an instance of the CaCertDistributionClient
// which implements the Manager
func NewCaCertEnrollmentClient() *CaCertEnrollmentClient {
	return &CaCertEnrollmentClient{}
}

// Instantiate the caCert distribution
func (c *CaCertEnrollmentClient) Instantiate(cert, project string) error {
	// check the current stateInfo of the Instantiation, if any
	ek := EnrollmentKey{
		Cert:       cert,
		Project:    project,
		Enrollment: enrollment.AppName}

	sc := module.NewStateClient(ek)
	if _, err := sc.VerifyState(common.Instantiate); err != nil {
		return err
	}

	// get the caCert
	caCert, err := getCertificate(cert, project)
	if err != nil {
		return err
	}

	// get all the logcalClouds associated with this caCert
	lcs, err := getAllLogicalClouds(cert, project)
	if err != nil {
		return err
	}

	// initialize a new appContext
	ctx := module.CaCertAppContext{
		AppName:    enrollment.AppName,
		ClientName: clientName}
	if err := ctx.InitAppContext(); err != nil {
		return err
	}

	// create a new EnrollmentContext
	eCtx := enrollment.EnrollmentContext{
		AppContext: ctx.AppContext,
		AppHandle:  ctx.AppHandle,
		CaCert:     caCert,
		ContextID:  ctx.ContextID,
		Resources: enrollment.EnrollmentResource{
			CertificateRequest: map[string]*cmv1.CertificateRequest{},
			Certificate:        map[string]*cmv1.Certificate{},
		},
		Project: project,
	}

	// set the issuing cluster handle
	eCtx.IssuerHandle, err = eCtx.IssuingClusterHandle()
	if err != nil {
		return err
	}

	// you can have multiple logcalClouds under the same caCert
	// we need to process all the logcalClouds within the same appContext
	// get all the clusters associated with these logicalClouds
	for _, lc := range lcs {
		// get the logicalCloud
		l, err := dcm.NewLogicalCloudClient().Get(context.Background(), project, lc.Spec.LogicalCloud)
		if err != nil {
			return err
		}

		eCtx.LogicalCloud = l.MetaData.Name

		if len(l.Specification.NameSpace) > 0 {
			eCtx.Namespace = l.Specification.NameSpace
		}

		// get all the clusters defined under this caCert
		eCtx.ClusterGroups, err = getAllClusterGroup(lc.MetaData.Name, cert, project)
		if err != nil {
			return err
		}

		// instantiate caCert enrollment
		if err = eCtx.Instantiate(); err != nil {
			return err
		}

		eCtx.Namespace = ""
		eCtx.LogicalCloud = ""
		eCtx.ClusterGroups = []module.ClusterGroup{}

	}

	// add instruction under the given handle and type
	if err := module.AddInstruction(eCtx.AppContext, eCtx.IssuerHandle, eCtx.ResOrder); err != nil {
		return err
	}

	// invokes the rsync service
	err = ctx.CallRsyncInstall()
	if err != nil {
		return err
	}

	// update the enrollment stateInfo
	if err := sc.Update(state.StateEnum.Instantiated, ctx.ContextID, false); err != nil {
		return err
	}

	return nil
}

// Status returns the caCert enrollment status
func (c *CaCertEnrollmentClient) Status(cert, project, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (module.CaCertStatus, error) {
	// get the enrollment stateInfo
	ek := EnrollmentKey{
		Cert:       cert,
		Project:    project,
		Enrollment: enrollment.AppName}
	stateInfo, err := module.NewStateClient(ek).Get()
	if err != nil {
		return module.CaCertStatus{}, err
	}

	sval, err := enrollment.Status(stateInfo, qInstance, qType, qOutput, fApps, fClusters, fResources)
	sval.Project = project
	return sval, err
}

// Terminate the caCert enrollment
func (c *CaCertEnrollmentClient) Terminate(cert, project string) error {
	// get enrollment stateInfo
	ek := EnrollmentKey{
		Cert:       cert,
		Project:    project,
		Enrollment: enrollment.AppName}

	sc := module.NewStateClient(ek)
	// check the current state of the Instantiation, if any
	contextID, err := sc.VerifyState(common.Terminate)
	if err != nil {
		return err
	}

	// initialize a new appContext
	ctx := module.CaCertAppContext{
		ContextID: contextID}
	// call resource synchronizer to delete the CSR from the issuing cluster
	if err := ctx.CallRsyncUninstall(); err != nil {
		return err
	}

	// get the caCert
	caCert, err := getCertificate(cert, project)
	if err != nil {
		return err
	}

	// get all the logcalcloud(s) associated with this caCert
	lcs, err := getAllLogicalClouds(cert, project)
	if err != nil {
		return err
	}

	// create a new EnrollmentContext
	eCtx := enrollment.EnrollmentContext{
		CaCert:    caCert,
		ContextID: ctx.ContextID,
		Resources: enrollment.EnrollmentResource{
			CertificateRequest: map[string]*cmv1.CertificateRequest{},
			Certificate:        map[string]*cmv1.Certificate{},
		},
		Project: project,
	}

	// you can have multiple logcalCloud(s) under the same caCert
	// we need to process all the logcalCloud(s) within the same appContext
	// get all the clusters associated with these logicalCloud(s)
	for _, lc := range lcs {
		// get the logicalCloud
		l, err := dcm.NewLogicalCloudClient().Get(context.Background(), project, lc.Spec.LogicalCloud)
		if err != nil {
			return err
		}

		eCtx.LogicalCloud = l.MetaData.Name

		if len(l.Specification.NameSpace) > 0 {
			eCtx.Namespace = l.Specification.NameSpace
		}

		// get all the clusters defined under this caCert
		eCtx.ClusterGroups, err = getAllClusterGroup(lc.MetaData.Name, cert, project)
		if err != nil {
			return err
		}

		// terminate the caCert enrollment
		if err = eCtx.Terminate(); err != nil {
			return err
		}

		eCtx.Namespace = ""
		eCtx.LogicalCloud = ""
		eCtx.ClusterGroups = []module.ClusterGroup{}

	}

	// update the enrollment stateInfo
	if err := sc.Update(state.StateEnum.Terminated, contextID, false); err != nil {
		return err
	}

	return nil
}

// Update the caCert enrollment
func (c *CaCertEnrollmentClient) Update(cert, project string) error {
	// get the caCert
	caCert, err := getCertificate(cert, project)
	if err != nil {
		return err
	}

	// get the stateInfo of the instantiation
	ek := EnrollmentKey{
		Cert:       cert,
		Project:    project,
		Enrollment: enrollment.AppName}
	sc := module.NewStateClient(ek)
	stateInfo, err := sc.Get()
	if err != nil {
		return err
	}

	contextID := state.GetLastContextIdFromStateInfo(stateInfo)
	if len(contextID) > 0 {
		// get the existing appContext
		status, err := state.GetAppContextStatus(context.Background(), contextID)
		if err != nil {
			logutils.Error("Failed to get the appContext status",
				logutils.Fields{
					"ContextID": contextID,
					"Error":     err.Error()})
			return err
		}
		if status.Status == appcontext.AppContextStatusEnum.Instantiated {
			// instantiate a new appContext
			ctx := module.CaCertAppContext{
				AppName:    enrollment.AppName,
				ClientName: clientName}
			if err := ctx.InitAppContext(); err != nil {
				return err
			}

			eCtx := enrollment.EnrollmentContext{
				AppContext: ctx.AppContext,
				AppHandle:  ctx.AppHandle,
				CaCert:     caCert,
				ContextID:  ctx.ContextID,
				ClientName: clientName,
				Resources: enrollment.EnrollmentResource{
					CertificateRequest: map[string]*cmv1.CertificateRequest{},
					Certificate:        map[string]*cmv1.Certificate{},
				},
				Project: project,
			}

			// set the issuing cluster handle
			eCtx.IssuerHandle, err = eCtx.IssuingClusterHandle()
			if err != nil {
				return err
			}

			// get all the logcalCloud(s) associated with this caCert
			lcs, err := getAllLogicalClouds(cert, project)
			if err != nil {
				return err
			}

			for _, lc := range lcs {
				// get the logicalCloud
				l, err := dcm.NewLogicalCloudClient().Get(context.Background(), project, lc.Spec.LogicalCloud)
				if err != nil {
					return err
				}

				eCtx.LogicalCloud = l.MetaData.Name

				if len(l.Specification.NameSpace) > 0 {
					eCtx.Namespace = l.Specification.NameSpace
				}

				// get all the clusters defined under this caCert
				eCtx.ClusterGroups, err = getAllClusterGroup(lc.MetaData.Name, cert, project)
				if err != nil {
					return err
				}

				// update the caCert enrollment appContext
				if err := eCtx.Update(contextID); err != nil {
					return err
				}

				eCtx.Namespace = ""
				eCtx.LogicalCloud = ""
				eCtx.ClusterGroups = []module.ClusterGroup{}

			}

			// update enrollment stateInfo
			if err := sc.Update(state.StateEnum.Updated, eCtx.ContextID, false); err != nil {
				return err
			}
		}

	}

	return nil
}
