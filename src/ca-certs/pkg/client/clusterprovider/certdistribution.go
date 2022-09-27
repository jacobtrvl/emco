// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

import (
	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	tcsv1 "github.com/intel/trusted-certificate-issuer/api/v1alpha1"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate/distribution"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate/enrollment"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/service/istioservice"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/service/knccservice"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	v1 "k8s.io/api/core/v1"
)

// CaCertDistributionManager exposes all the clusterProvider caCert distribution functionalities
type CaCertDistributionManager interface {
	Instantiate(cert, clusterProvider string) error
	Status(cert, clusterProvider, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (module.CaCertStatus, error)
	Terminate(cert, clusterProvider string) error
	Update(cert, clusterProvider string) error
}

// CaCertDistributionClient holds the client properties
type CaCertDistributionClient struct {
}

// NewCaCertDistributionClient returns an instance of the CaCertDistributionClient
// which implements the Manager
func NewCaCertDistributionClient() *CaCertDistributionClient {
	return &CaCertDistributionClient{}
}

// Instantiate the clusterProvider caCert distribution
func (c *CaCertDistributionClient) Instantiate(cert, clusterProvider string) error {
	// check the current stateInfo of the Instantiation, if any
	dk := DistributionKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Distribution:    distribution.AppName}

	sc := module.NewStateClient(dk)
	if _, err := sc.VerifyState(common.Instantiate); err != nil {
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

	// get the caCert
	caCert, err := getCertificate(cert, clusterProvider)
	if err != nil {
		return err
	}

	// initialize a new appContext
	ctx := module.CaCertAppContext{
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
		EnrollmentContextID: enrollmentContextID,
		Resources: distribution.DistributionResource{
			ClusterIssuer: map[string]*cmv1.ClusterIssuer{},
			ProxyConfig:   map[string]*istioservice.ProxyConfig{},
			Secret:        map[string]*v1.Secret{},
			KnccConfig:    map[string]*knccservice.Config{},
			TCSIssuer:     map[string]*tcsv1.TCSIssuer{},
		}}

	// get all the clusters defined under this caCert
	dCtx.ClusterGroups, err = getAllClusterGroup(cert, clusterProvider)
	if err != nil {
		return err
	}

	// start caCert distribution instantiation
	if err = dCtx.Instantiate(); err != nil {
		return err
	}

	// invokes the rsync service
	err = ctx.CallRsyncInstall()
	if err != nil {
		return err
	}

	// update caCert distribution state
	if err := module.NewStateClient(dk).Update(state.StateEnum.Instantiated, ctx.ContextID, false); err != nil {
		return err
	}

	return nil
}

// Status returns the caCert distribution status
func (c *CaCertDistributionClient) Status(cert, clusterProvider, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (module.CaCertStatus, error) {
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

// Terminate the caCert distribution
func (c *CaCertDistributionClient) Terminate(cert, clusterProvider string) error {
	dk := DistributionKey{
		Cert:            cert,
		ClusterProvider: clusterProvider,
		Distribution:    distribution.AppName}

	return distribution.Terminate(dk)
}

// Update the caCert distribution
func (c *CaCertDistributionClient) Update(cert, clusterProvider string) error {
	// get the caCert
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
		// get all the clusters defined under this caCert
		clusterGroups, err := getAllClusterGroup(cert, clusterProvider)
		if err != nil {
			return err
		}

		// instantiate a new appContext
		ctx := module.CaCertAppContext{
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
			ClusterGroups: clusterGroups,
			Resources: distribution.DistributionResource{
				ClusterIssuer: map[string]*cmv1.ClusterIssuer{},
				ProxyConfig:   map[string]*istioservice.ProxyConfig{},
				Secret:        map[string]*v1.Secret{},
				KnccConfig:    map[string]*knccservice.Config{},
				TCSIssuer:     map[string]*tcsv1.TCSIssuer{},
			},
		}

		// start the caCert distribution instantiation
		if err := dCtx.Instantiate(); err != nil {
			return err
		}
		// update the appContext
		if err := dCtx.Update(previd); err != nil {
			return err
		}

		// update the state object for the caCert resource
		if err := module.NewStateClient(dk).Update(state.StateEnum.Updated, dCtx.ContextID, false); err != nil {
			return err
		}

	}

	return nil
}
