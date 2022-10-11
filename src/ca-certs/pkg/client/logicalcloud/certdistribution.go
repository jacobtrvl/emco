// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import (
	"context"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	tcsv1 "github.com/intel/trusted-certificate-issuer/api/v1alpha1"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate/distribution"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate/enrollment"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/service/istioservice"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/service/knccservice"
	dcm "gitlab.com/project-emco/core/emco-base/src/dcm/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	v1 "k8s.io/api/core/v1"
)

// CaCertDistributionManager exposes all the logicalCloud caCert distribution functionalities
type CaCertDistributionManager interface {
	Instantiate(cert, project string) error
	Status(cert, project, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (module.CaCertStatus, error)
	Terminate(cert, project string) error
	Update(cert, project string) error
}

// CaCertDistributionClient holds the client properties
type CaCertDistributionClient struct {
}

// NewCaCertDistributionClient returns an instance of the CaCertDistributionClient
// which implements the Manager
func NewCaCertDistributionClient() *CaCertDistributionClient {
	return &CaCertDistributionClient{}
}

// Instantiate the logicalCloud caCert distribution
func (c *CaCertDistributionClient) Instantiate(cert, project string) error {
	// check the current stateInfo of the Instantiation, if any
	dk := DistributionKey{
		Cert:         cert,
		Project:      project,
		Distribution: distribution.AppName}

	sc := module.NewStateClient(dk)
	if _, err := sc.VerifyState(common.Instantiate); err != nil {
		return err
	}

	// verify the enrollment state
	ek := EnrollmentKey{
		Cert:       cert,
		Project:    project,
		Enrollment: enrollment.AppName}
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
	caCert, err := getCertificate(cert, project)
	if err != nil {
		return err
	}

	// get all the logicalCloud(s) associated with this caCert
	lcs, err := getAllLogicalClouds(cert, project)
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
		},
		Project: project,
	}

	//  you can have multiple logicalCloud(s) under the same caCert
	//  we need to process all the logicalCloud(s) within the same appContext
	// get all the clusters associated with these logicalCloud(s)
	for _, lc := range lcs {
		// get the logical cloud
		l, err := dcm.NewLogicalCloudClient().Get(context.Background(), project, lc.Spec.LogicalCloud)
		if err != nil {
			return err
		}

		dCtx.LogicalCloud = l.MetaData.Name

		if len(l.Specification.NameSpace) > 0 {
			dCtx.Namespace = l.Specification.NameSpace
		}

		// get all the clusters defined under this caCert
		dCtx.ClusterGroups, err = getAllClusterGroup(lc.MetaData.Name, cert, project)
		if err != nil {
			return err
		}

		// instantiate the caCert distribution
		if err = dCtx.Instantiate(); err != nil {
			return err
		}

		dCtx.Namespace = ""
		dCtx.LogicalCloud = ""
		dCtx.ClusterGroups = []module.ClusterGroup{}
	}

	// invokes the rsync service
	err = ctx.CallRsyncInstall()
	if err != nil {
		return err
	}

	// update caCert distribution stateInfo
	if err := module.NewStateClient(dk).Update(state.StateEnum.Instantiated, ctx.ContextID, false); err != nil {
		return err
	}

	return nil
}

// Status returns the caCert distribution status
func (c *CaCertDistributionClient) Status(cert, project, qInstance, qType, qOutput string, fApps, fClusters, fResources []string) (module.CaCertStatus, error) {
	dk := DistributionKey{
		Cert:         cert,
		Project:      project,
		Distribution: distribution.AppName}

	// get the current state of the Distribution
	stateInfo, err := module.NewStateClient(dk).Get()
	if err != nil {
		return module.CaCertStatus{}, err
	}

	sval, err := enrollment.Status(stateInfo, qInstance, qType, qOutput, fApps, fClusters, fResources)
	sval.Project = project
	return sval, err
}

// Terminate the caCert distribution
func (c *CaCertDistributionClient) Terminate(cert, project string) error {
	dk := DistributionKey{
		Cert:         cert,
		Project:      project,
		Distribution: distribution.AppName}

	return distribution.Terminate(dk)
}

// Update the caCert distribution
func (c *CaCertDistributionClient) Update(cert, project string) error {
	// get the caCert
	caCert, err := getCertificate(cert, project)
	if err != nil {
		return err
	}

	dk := DistributionKey{
		Cert:         cert,
		Project:      project,
		Distribution: distribution.AppName}

	previd, status, err := module.GetAppContextStatus(dk)
	if err != nil {
		return err
	}

	if status == appcontext.AppContextStatusEnum.Instantiated {
		// instantiate a new appContext
		ctx := module.CaCertAppContext{
			AppName:    distribution.AppName,
			ClientName: clientName}
		if err := ctx.InitAppContext(); err != nil {
			return err
		}

		dCtx := distribution.DistributionContext{
			AppContext: ctx.AppContext,
			AppHandle:  ctx.AppHandle,
			CaCert:     caCert,
			ContextID:  ctx.ContextID,
			ClientName: clientName,
			Resources: distribution.DistributionResource{
				ClusterIssuer: map[string]*cmv1.ClusterIssuer{},
				ProxyConfig:   map[string]*istioservice.ProxyConfig{},
				Secret:        map[string]*v1.Secret{},
				KnccConfig:    map[string]*knccservice.Config{},
				TCSIssuer:     map[string]*tcsv1.TCSIssuer{},
			},
			Project: project,
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

			dCtx.LogicalCloud = l.MetaData.Name

			if len(l.Specification.NameSpace) > 0 {
				dCtx.Namespace = l.Specification.NameSpace
			}

			// get all the clusters defined under this caCert
			dCtx.ClusterGroups, err = getAllClusterGroup(lc.MetaData.Name, cert, project)
			if err != nil {
				return err
			}

			// start the caCert distribution instantiation
			if err := dCtx.Instantiate(); err != nil {
				return err
			}

			dCtx.Namespace = ""
			dCtx.LogicalCloud = ""
			dCtx.ClusterGroups = []module.ClusterGroup{}
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
