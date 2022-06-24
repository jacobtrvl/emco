// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

import (
	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate/enrollment"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
)

const clientName string = "cacert"

// CaCertEnrollmentManager
type CaCertEnrollmentManager interface {
	Instantiate(cert, clusterProvider string) error
	Status(cert, clusterProvider, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (module.CaCertStatus, error)
	Terminate(cert, clusterProvider string) error
	Update(cert, clusterProvider string) error
}

// CaCertEnrollmentClient
type CaCertEnrollmentClient struct {
}

// NewCaCertEnrollmentClient
func NewCaCertEnrollmentClient() *CaCertEnrollmentClient {
	return &CaCertEnrollmentClient{}
}

// Instantiate
func (c *CaCertEnrollmentClient) Instantiate(cert, clusterProvider string) error {
	// check the stateInfo of the Instantiation, if any
	ek := EnrollmentKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Enrollment:      enrollment.AppName}
	sc := module.NewStateClient(ek)
	if _, err := sc.VerifyState(module.InstantiateEvent); err != nil {
		return err
	}

	// get the caCert
	caCert, err := getCertificate(cert, clusterProvider)
	if err != nil {
		return err
	}

	// get all the clusters defined under this CA
	clusterGroups, err := getAllClusterGroup(cert, clusterProvider)
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
		AppContext:    ctx.AppContext,
		AppHandle:     ctx.AppHandle,
		CaCert:        caCert,
		ContextID:     ctx.ContextID,
		ClusterGroups: clusterGroups,
		Resources: enrollment.EnrollmentResource{
			CertificateRequest: map[string]*cmv1.CertificateRequest{},
		}}

	// set the issuing cluster handle
	eCtx.IssuerHandle, err = eCtx.IssuingClusterHandle()
	if err != nil {
		return err
	}

	// instantiate caCert enrollment
	if err = eCtx.Instantiate(); err != nil {
		return err
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

	// update the enrollment state
	if err := sc.Update(state.StateEnum.Instantiated, ctx.ContextID, false); err != nil {
		return err
	}

	return nil
}

// Status
func (c *CaCertEnrollmentClient) Status(cert, clusterProvider, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (module.CaCertStatus, error) {
	// get the enrollment stateInfo
	ek := EnrollmentKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Enrollment:      enrollment.AppName}
	stateInfo, err := module.NewStateClient(ek).Get()
	if err != nil {
		return module.CaCertStatus{}, err
	}

	sval, err := enrollment.Status(stateInfo, qInstance, qType, qOutput, fApps, fClusters, fResources)
	sval.ClusterProvider = clusterProvider
	return sval, err
}

// Terminate
func (c *CaCertEnrollmentClient) Terminate(cert, clusterProvider string) error {
	// get the enrollment stateInfo
	ek := EnrollmentKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Enrollment:      enrollment.AppName}
	sc := module.NewStateClient(ek)
	// check the current state of the Instantiation, if any
	contextID, err := sc.VerifyState(module.TerminateEvent)
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
	caCert, err := getCertificate(cert, clusterProvider)
	if err != nil {
		return err
	}

	// get all the clusters defined under this CA
	clusterGroups, err := getAllClusterGroup(cert, clusterProvider)
	if err != nil {
		return err
	}

	// create a new EnrollmentContext
	eCtx := enrollment.EnrollmentContext{
		CaCert:        caCert,
		ContextID:     ctx.ContextID,
		ClusterGroups: clusterGroups,
		Resources: enrollment.EnrollmentResource{
			CertificateRequest: map[string]*cmv1.CertificateRequest{},
		}}

	// terminate the caCert enrollment
	if err = eCtx.Terminate(); err != nil {
		return err
	}

	// update enrollment stateInfo
	if err := sc.Update(state.StateEnum.Terminated, contextID, false); err != nil {
		return err
	}

	return nil
}

// Update
func (c *CaCertEnrollmentClient) Update(cert, clusterProvider string) error {
	// get the caCert
	caCert, err := getCertificate(cert, clusterProvider)
	if err != nil {
		return err
	}

	// get the stateInfo of the instantiation
	ek := EnrollmentKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Enrollment:      enrollment.AppName}
	sc := module.NewStateClient(ek)
	stateInfo, err := sc.Get()
	if err != nil {
		return err
	}

	contextID := state.GetLastContextIdFromStateInfo(stateInfo)
	if len(contextID) > 0 {
		// get the existing appContext
		status, err := state.GetAppContextStatus(contextID)
		if err != nil {
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

			// get all the clusters defined under this CA
			clusterGroups, err := getAllClusterGroup(cert, clusterProvider)
			if err != nil {
				return err
			}

			eCtx := enrollment.EnrollmentContext{
				AppContext:    ctx.AppContext,
				AppHandle:     ctx.AppHandle,
				CaCert:        caCert,
				ContextID:     ctx.ContextID,
				ClientName:    clientName,
				ClusterGroups: clusterGroups,
				Resources: enrollment.EnrollmentResource{
					CertificateRequest: map[string]*cmv1.CertificateRequest{},
				}}
			// update the caCert enrollment app context
			if err := eCtx.Update(contextID); err != nil {
				return err
			}

			// update the state object for the caCert resource
			if err := sc.Update(state.StateEnum.Updated, eCtx.ContextID, false); err != nil {
				return err
			}
		}

	}

	return nil
}
