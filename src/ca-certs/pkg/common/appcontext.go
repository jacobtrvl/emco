// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package common

import (
	"encoding/json"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

type Context struct {
	AppContext appcontext.AppContext
	AppHandle  interface{}
	AppName    string
	ClientName string
	ContextID  string
	Resorder   []string
}

func (ctx *Context) InitAppContext() error {
	var log = func(message, contextID string, err error) {
		fields := make(logutils.Fields)
		fields["contextID"] = contextID
		if err != nil {
			fields["Error"] = err.Error()
		}
		logutils.Error(message, fields)
	}
	appContext := appcontext.AppContext{}
	contextID, err := appContext.InitAppContext()
	if err != nil {
		return err
	}

	handle, err := appContext.CreateCompositeApp()
	if err != nil {
		return err
	}

	appHandle, err := appContext.AddApp(handle, ctx.AppName)
	if err != nil {
		if er := appContext.DeleteCompositeApp(); er != nil {
			log("Failed to delete the compositeApp", contextID.(string), err)
		}
		return err
	}

	// Add App Order
	appOrder, err := json.Marshal(map[string][]string{"apporder": {ctx.AppName}})
	if err != nil {
		log("Failed to create apporder", contextID.(string), err)
		return err
	}

	// Add app level Order and Dependency
	_, err = appContext.AddInstruction(handle, "app", "order", string(appOrder))
	if err != nil {
		if er := appContext.DeleteCompositeApp(); er != nil {
			log("Failed to delete the compositeApp", contextID.(string), err)
		}
		return err
	}
	ctx.AppContext = appContext
	ctx.AppHandle = appHandle
	ctx.ContextID = contextID.(string)

	return nil
}
