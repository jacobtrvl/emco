// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

import (
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate/distribution"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate/enrollment"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
)

// CertDistributionManager exposes all the functionalities related to CA cert distribution
type CertDistributionManager interface {
	Instantiate(cert, clusterProvider string) error
	Status(cert, clusterProvider, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (module.CaCertStatus, error)
	Terminate(cert, clusterProvider string) error
	Update(cert, clusterProvider string) error
}

// CertDistributionClient holds the client properties
type CertDistributionClient struct {
}

// NewCertDistributionClient returns an instance of the CertDistributionClient
// which implements the Manager
func NewCertDistributionClient() *CertDistributionClient {
	return &CertDistributionClient{}
}

// Instantiate the cert distribution
func (c *CertDistributionClient) Instantiate(cert, clusterProvider string) error {
	// check the current stateInfo of the Instantiation, if any
	dk := DistributionKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Distribution:    distribution.AppName}

	sc := module.NewStateClient(dk)
	if _, err := sc.VerifyState(module.InstantiateEvent); err != nil {
		return err
	}

	// verify the enrollment state
	ek := EnrollmentKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Enrollment:      enrollment.AppName}
	stateInfo, err := module.NewStateClient(ek).Get()
	if err != nil {
		return err
	}

	enrollmentContextID, err := enrollment.VerifyEnrollmentState(stateInfo)
	if err != nil {
		return err
	}

	// validate the enrollment status
	_, err = enrollment.ValidateEnrollmentStatus(stateInfo)
	if err != nil {
		return err
	}

	// get the ca cert
	caCert, err := getCertificate(cert, clusterProvider)
	if err != nil {
		return err
	}

	// initialize a new dCtx
	ctx := module.CertAppContext{
		AppName:    distribution.AppName,
		ClientName: clientName}
	if err := ctx.InitAppContext(); err != nil {
		return err
	}

	// create a new Distribution Context
	dCtx := distribution.DistributionContext{
		AppContext:          ctx.AppContext,
		AppHandle:           ctx.AppHandle,
		CaCert:              caCert,
		ContextID:           ctx.ContextID,
		EnrollmentContextID: enrollmentContextID}

	// get all the clusters defined under this CA
	dCtx.ClusterGroups, err = getAllClusterGroup(cert, clusterProvider)
	if err != nil {
		return err
	}

	// start cert distribution instantiation
	if err = dCtx.Instantiate(); err != nil {
		return err
	}

	// invokes the rsync service
	err = ctx.CallRsyncInstall()
	if err != nil {
		return err
	}

	// update cert distribution state
	if err := module.NewStateClient(dk).Update(state.StateEnum.Instantiated, ctx.ContextID, false); err != nil {
		return err
	}

	return nil
}

// Status
func (c *CertDistributionClient) Status(cert, clusterProvider, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (module.CaCertStatus, error) {
	// get the current state of the
	dk := DistributionKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Distribution:    distribution.AppName}
	stateInfo, err := module.NewStateClient(dk).Get()
	if err != nil {
		return module.CaCertStatus{}, err
	}

	sval, err := enrollment.Status(stateInfo, qInstance, qType, qOutput, fApps, fClusters, fResources)
	sval.ClusterProvider = clusterProvider
	return sval, err
}

// Terminate
func (c *CertDistributionClient) Terminate(cert, clusterProvider string) error {
	dk := DistributionKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Distribution:    distribution.AppName}

	return distribution.Terminate(dk)
}

// Update
func (c *CertDistributionClient) Update(cert, clusterProvider string) error {
	// get the ca cert
	caCert, err := getCertificate(cert, clusterProvider)
	if err != nil {
		return err
	}

	dk := DistributionKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Distribution:    distribution.AppName}

	previd, status, err := module.GetAppContextStatus(dk)
	if err != nil {
		return err
	}

	if status == appcontext.AppContextStatusEnum.Instantiated {
		// get all the clusters defined under this CA
		clusterGroups, err := getAllClusterGroup(cert, clusterProvider)
		if err != nil {
			return err
		}

		// instantiate a new app context
		ctx := module.CertAppContext{
			AppName:    distribution.AppName,
			ClientName: clientName}
		if err := ctx.InitAppContext(); err != nil {
			return err
		}

		dCtx := distribution.DistributionContext{
			AppContext:    ctx.AppContext,
			AppHandle:     ctx.AppHandle,
			CaCert:        caCert,
			ContextID:     ctx.ContextID,
			ClientName:    clientName,
			ClusterGroups: clusterGroups}

		// start the cert distribution instantiation
		if err := dCtx.Instantiate(); err != nil {
			return err
		}
		// update the app context
		if err := dCtx.Update(previd); err != nil {
			return err
		}

		// update the state object for the cert resource
		if err := module.NewStateClient(dk).Update(state.StateEnum.Updated, dCtx.ContextID, false); err != nil {
			return err
		}

	}

	return nil
}
