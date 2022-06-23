// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"encoding/json"
	"fmt"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/notifyclient"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/readynotify"
)

// CaCertAppContext
type CaCertAppContext struct {
	AppContext appcontext.AppContext
	AppHandle  interface{}
	AppName    string
	ClientName string
	ContextID  string
	Resorder   []string
}

// InitAppContext
func (ctx *CaCertAppContext) InitAppContext() error {
	appContext := appcontext.AppContext{}
	contextID, err := appContext.InitAppContext()
	if err != nil {
		logutils.Error("Failed to initialize the appContext",
			logutils.Fields{
				"Error": err.Error()})
		return err
	}

	compAppHandle, err := appContext.CreateCompositeApp()
	if err != nil {
		logutils.Error("Failed to create the compositeApp",
			logutils.Fields{
				"Error": err.Error()})
		return err
	}

	appHandle, err := appContext.AddApp(compAppHandle, ctx.AppName)
	if err != nil {
		logutils.Error("Failed to add app to the compositeApp",
			logutils.Fields{
				"App":   ctx.AppName,
				"Error": err.Error()})

		if er := appContext.DeleteCompositeApp(); er != nil {
			logutils.Error("Failed to delete the compositeApp",
				logutils.Fields{
					"ContextID": contextID.(string),
					"Error":     er.Error()})
			return er
		}

		return err
	}

	// Add App Order
	appOrder, err := json.Marshal(map[string][]string{"apporder": {ctx.AppName}})
	if err != nil {
		logutils.Error("Failed to marshal apporder",
			logutils.Fields{
				"Apporder": ctx.AppName,
				"Error":    err.Error()})
		return err
	}

	// Add app level Order
	if _, err = appContext.AddInstruction(compAppHandle, "app", "order", string(appOrder)); err != nil {
		logutils.Error("Failed to add the app level instruction order",
			logutils.Fields{
				"App":   ctx.AppName,
				"Error": err.Error()})

		if er := appContext.DeleteCompositeApp(); er != nil {
			logutils.Error("Failed to delete the compositeApp",
				logutils.Fields{
					"ContextID": contextID.(string),
					"Error":     er.Error()})
			return er
		}

		return err
	}

	ctx.AppContext = appContext
	ctx.AppHandle = appHandle
	ctx.ContextID = contextID.(string)

	return nil
}

// CallRsyncInstall
func (ctx *CaCertAppContext) CallRsyncInstall() error {
	// invokes the rsync service
	if err := notifyclient.CallRsyncInstall(ctx.ContextID); err != nil {
		logutils.Error("Rsync install failed",
			logutils.Fields{
				"ContextID": ctx.ContextID,
				"Error":     err.Error()})

		if er := ctx.AppContext.DeleteCompositeApp(); er != nil {
			logutils.Error("Failed to delete the compositeApp",
				logutils.Fields{
					"ContextID": ctx.ContextID,
					"Error":     er.Error()})
			return er
		}

		return err
	}

	// subscribe to alerts
	stream, _, err := notifyclient.InvokeReadyNotify(ctx.ContextID, ctx.ClientName)
	if err != nil {
		logutils.Error("Failed to subscribe to alerts",
			logutils.Fields{
				"ContextID":  ctx.ContextID,
				"ClientName": ctx.ClientName,
				"Error":      err.Error()})
		return err
	}

	if err := stream.CloseSend(); err != nil {
		logutils.Error("Failed to close the send stream",
			logutils.Fields{
				"ContextID":  ctx.ContextID,
				"ClientName": ctx.ClientName,
				"Error":      err.Error()})
		return err
	}

	return nil
}

// RetrieveAppContext retrieve the appContext from the stream
func RetrieveAppContext(stream readynotify.ReadyNotify_AlertClient, client readynotify.ReadyNotifyClient) string {
	var (
		contextID  string
		backOff    int = config.GetConfiguration().BackOff
		maxBackOff int = config.GetConfiguration().MaxBackOff
	)

	// retrieve the appContextID from the stream, wait till we get the notification response
	for {
		resp, err := stream.Recv()
		if err != nil {
			logutils.Error(fmt.Sprintf("Failed to receive ReadyNotify notification, retry after %d seconds.", backOff),
				logutils.Fields{
					"Error": err.Error()})
			// instead of retrying immediately, waits some amount of time between tries
			time.Sleep(time.Duration(backOff) * time.Second)

			if backOff*2 < maxBackOff {
				backOff *= 2
			} else {
				backOff = maxBackOff
			}

			continue
		}

		contextID = resp.AppContext
		logutils.Info("Received ReadyNotify notification alert",
			logutils.Fields{
				"appContextID": contextID})
		// received notification response
		break
	}

	return contextID
}

// CallRsyncUninstall
func (ctx *CaCertAppContext) CallRsyncUninstall() error {
	if err := notifyclient.CallRsyncUninstall(ctx.ContextID); err != nil {
		logutils.Error("Rsync uninstall failed",
			logutils.Fields{
				"Error": err.Error()})
		return err
	}

	return nil
}

// AddResource
func AddResource(appContext appcontext.AppContext, resource, handle interface{}, name string) error {
	value, err := json.Marshal(resource)
	if err != nil {
		logutils.Error("Failed to marshal the resource",
			logutils.Fields{
				"Resource": resource,
				"Error":    err.Error()})
		return err
	}

	if _, err = appContext.AddResource(handle, name, string(value)); err != nil {
		logutils.Error("Failed to add the resource",
			logutils.Fields{
				"Resource": name,
				"Error":    err.Error()})

		if er := appContext.DeleteCompositeApp(); er != nil {
			logutils.Error("Failed to delete the compositeApp",
				logutils.Fields{
					"Error": er.Error()})
			return er
		}

		return err
	}

	return nil
}

// AddInstruction
func AddInstruction(appContext appcontext.AppContext, handle interface{}, resOrder []string) error {
	order, err := json.Marshal(map[string][]string{"resorder": resOrder})
	if err != nil {
		logutils.Error("Failed to marshal resorder",
			logutils.Fields{
				"Error": err.Error()})
		return err
	}

	if _, err = appContext.AddInstruction(handle, "resource", "order", string(order)); err != nil {
		logutils.Error("Failed to add the resource level instruction order",
			logutils.Fields{
				"Error": err.Error()})

		if er := appContext.DeleteCompositeApp(); er != nil {
			logutils.Error("Failed to delete the compositeApp",
				logutils.Fields{
					"Error": err.Error()})
		}

		return err
	}

	return nil
}

// GetAppContextStatus
func GetAppContextStatus(key interface{}) (string, appcontext.StatusValue, error) {
	// get the current state of the instantiation
	stateInfo, err := NewStateClient(key).Get()
	if err != nil {
		return "", "", err
	}

	contextID := state.GetLastContextIdFromStateInfo(stateInfo)
	if len(contextID) > 0 {
		status, err := state.GetAppContextStatus(contextID)
		if err != nil {
			logutils.Error("Failed to get the appContext status",
				logutils.Fields{
					"ContextID": contextID,
					"Error":     err.Error()})
			return contextID, "", err
		}

		return contextID, status.Status, nil
	}

	return contextID, "", nil
}
