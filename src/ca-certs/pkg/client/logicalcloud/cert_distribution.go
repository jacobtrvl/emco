// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import (
	"strings"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate/distribution"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate/enrollment"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	dcm "gitlab.com/project-emco/core/emco-base/src/dcm/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
)

// CertDistributionManager
type CertDistributionManager interface {
	Instantiate(cert, project string) error
	Status(cert, project string) (module.ResourceStatus, error)
	Terminate(cert, project string) error
	Update(cert, project string) error
}

// CertDistributionClient
type CertDistributionClient struct {
}

// NewCertDistributionClient
func NewCertDistributionClient() *CertDistributionClient {
	return &CertDistributionClient{}
}

func (c *CertDistributionClient) Instantiate(cert, project string) error {
	dk := DistributionKey{
		Cert:         cert,
		Project:      project,
		Distribution: distribution.AppName}

	// get the cert enrollemnt instantiation state
	sc := module.NewStateClient(dk)
	if _, err := sc.VerifyState(module.InstantiateEvent); err != nil {
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

	// get the ca cert
	caCert, err := getCertificate(cert, project)
	if err != nil {
		return err
	}

	// get all the logcal-clouds associated with this cert
	lcs, err := getAllLogicalClouds(cert, project)
	if err != nil {
		return err
	}

	// initialize a new app context
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

	//  you can have multiple lcs under the same cert
	//  we need to process all the lcs within the same app context
	// get all the clusters associated with these logical-clouds
	for _, lc := range lcs {
		// get the logical cloud
		l, err := dcm.NewLogicalCloudClient().Get(project, lc.MetaData.Name)
		if err != nil {
			return err
		}

		if len(l.Specification.NameSpace) > 0 &&
			strings.ToLower(l.Specification.NameSpace) != "default" {
			dCtx.Namespace = l.Specification.NameSpace
		}

		// get all the clusters defined under this CA
		dCtx.ClusterGroups, err = getAllClusterGroup(lc.MetaData.Name, cert, project)
		if err != nil {
			return err
		}

		// instantiate the cert distribution
		if err = dCtx.Instantiate(); err != nil {
			return err
		}
	}

	// invokes the rsync service
	err = ctx.CallRsyncInstall()
	if err != nil {
		return err
	}

	// update distribution stateInfo
	if err := module.NewStateClient(dk).Update(state.StateEnum.Instantiated, ctx.ContextID, false); err != nil {
		return err
	}

	return nil
}

// Status
func (c *CertDistributionClient) Status(cert, project string) (module.ResourceStatus, error) {
	dk := DistributionKey{
		Cert:         cert,
		Project:      project,
		Distribution: distribution.AppName}

	// get the current state of the
	stateInfo, err := module.NewStateClient(dk).Get()
	if err != nil {
		return module.ResourceStatus{}, err
	}

	return enrollment.Status(stateInfo)
}

// Terminate
func (c *CertDistributionClient) Terminate(cert, project string) error {
	dk := DistributionKey{
		Cert:         cert,
		Project:      project,
		Distribution: distribution.AppName}

	return distribution.Terminate(dk)
}

// Update
func (c *CertDistributionClient) Update(cert, project string) error {
	// get the ca cert
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
		// instantiate a new app context
		ctx := module.CertAppContext{
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
			ClientName: clientName}

		// get all the logcal-clouds associated with this cert
		lcs, err := getAllLogicalClouds(cert, project)
		if err != nil {
			return err
		}

		for _, lc := range lcs {
			// get all the clusters defined under this CA
			clusterGroups, err := getAllClusterGroup(lc.MetaData.Name, cert, project)
			if err != nil {
				return err
			}

			dCtx.ClusterGroups = clusterGroups

			// start the cert distribution instantiation
			if err := dCtx.Instantiate(); err != nil {
				return err
			}
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
