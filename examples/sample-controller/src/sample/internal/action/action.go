// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

// Package action applies the specific action invoked by
// the application scheduler(orchestrator) using the gRPC call(s).
// The action is associated with the controller type.
// Placement controllers decide where the specific application should get placed,
// and the action controller modifies the current state of the resources.
// ref: https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/hpa-ac - HPA action controller
// ref: https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/hpa-plc - HPA placement controller
package action

import (
	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/examples/sample-controller/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/context"
)

// UpdateAppContext applies the supplied intent against the given AppContext ID
func UpdateAppContext(intentName, appContextId string) error {
	var ac appcontext.AppContext

	_, err := ac.LoadAppContext(appContextId)
	if err != nil {
		logutils.Error("Failed to get the appContext.",
			logutils.Fields{
				"ID": appContextId})
		return errors.Wrapf(err, "Failed to get the appContext with ID: %s.", appContextId)
	}

	_, err = ac.GetCompositeAppHandle()
	if err != nil {
		return err
	}

	appContext, err := context.ReadAppContext(appContextId)
	if err != nil {
		logutils.Error("Failed to get the compositeApp for the appContext.",
			logutils.Fields{
				"ID": appContextId})
		return errors.Wrapf(err, "Failed to get the compositeApp for the appContext with ID: %s", appContextId)
	}

	project := appContext.CompMetadata.Project
	app := appContext.CompMetadata.CompositeApp
	version := appContext.CompMetadata.Version
	group := appContext.CompMetadata.DeploymentIntentGroup

	// Look up all  Intents
	intents, err := module.NewClient().SampleIntent.GetSampleIntents("", project, app, version, group)
	if err != nil {
		logutils.Error("Failed to get the intents for the deploymentIntentGroup.",
			logutils.Fields{
				"DeploymentIntentGroup": group})
		return errors.Wrapf(err, "Failed to get the intents for the deploymentIntentGroup: %s", group)
	}

	if len(intents) == 0 {
		logutils.Warn("No intents are defined for the deploymentIntentGroup.",
			logutils.Fields{
				"DeploymentIntentGroup": group})
		return errors.Errorf("No intents are defined for the deploymentIntentGroup: %s", group)
	}

	for _, i := range intents {
		// Implement the action controller specific logic here.
		// Action controllers modifies the current state of the resources.
		// ref: https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/hpa-ac/internal/action - HPA action controller
		logutils.Info(i.Metadata.Name,
			logutils.Fields{})
	}

	return nil
}

// FilterClusters applies the supplied intent against the given AppContext ID
func FilterClusters(appContextID string) error {
	var ac appcontext.AppContext

	_, err := ac.LoadAppContext(appContextID)
	if err != nil {
		logutils.Error("Failed to get the appContext",
			logutils.Fields{
				"ID": appContextID})
		return errors.Wrapf(err, "Failed to get the appContext with ID: %s", appContextID)
	}

	ca, err := ac.GetCompositeAppMeta()
	if err != nil {
		logutils.Error("Failed to get the appContext metaData",
			logutils.Fields{
				"ID": appContextID})
		return errors.Wrapf(err, "Failed to get the appContext metaData with ID: %s", appContextID)
	}

	project := ca.Project
	app := ca.CompositeApp
	version := ca.Version
	group := ca.DeploymentIntentGroup

	// Look up all  Intents
	intents, err := module.NewClient().SampleIntent.GetSampleIntents("", project, app, version, group)
	if err != nil {
		logutils.Error("Failed to get the intents for the deploymentIntentGroup.",
			logutils.Fields{
				"DeploymentIntentGroup": group})
		return errors.Wrapf(err, "Failed to get the intents for the deploymentIntentGroup: %s", group)
	}

	if len(intents) == 0 {
		logutils.Warn("No intents defined for the deploymentIntentGroup.",
			logutils.Fields{
				"DeploymentIntentGroup": group})
		return errors.Errorf("No intents defined for the deploymentIntentGroup: %s", group)
	}

	for _, i := range intents {
		// Implement the placement controller specific logic here.
		// Placement controllers decide where the specific application should get placed.
		// ref: https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/hpa-plc - HPA placement controller
		logutils.Info(i.Metadata.Name,
			logutils.Fields{})
	}

	return nil
}
