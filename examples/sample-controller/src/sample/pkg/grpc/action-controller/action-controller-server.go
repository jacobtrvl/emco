// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

// Package actioncontroller exposes the gRPC server-specific functionalities of an action controller.
// These functions are getting invoked by the application scheduler(orchestrator).
// The implementation of these gRPC functionalities, and their associated packages
//  (in this case, actioncontroller), will change based on the requirement(s).
// For example, a controller can be one of the placement types or actions.
// Placement controllers allow the orchestrator to choose the exact locations
// to place the application in the composite application. On the other hand,
// action controllers can modify the state of a resource(create additional resources
// to be deployed, modify or delete the existing resources). You can define your packages
// and functionalities based on your need and expose these functionalities using the gRPC server.
// In EMCO, we have separate controllers for action and placement.
// ref: https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/hpa-ac - HPA action controller
// ref: https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/hpa-plc - HPA placement controller
package actioncontroller

import (
	"context"

	"gitlab.com/project-emco/core/emco-base/examples/sample-controller/internal/action"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/contextupdate"
)

type actionControllerServer struct {
	contextupdate.UnimplementedContextupdateServer
}

func (ac *actionControllerServer) UpdateAppContext(ctx context.Context, req *contextupdate.ContextUpdateRequest) (*contextupdate.ContextUpdateResponse, error) {
	err := action.UpdateAppContext(ctx, req.IntentName, req.AppContext)
	if err != nil {
		return &contextupdate.ContextUpdateResponse{AppContextUpdated: false, AppContextUpdateMessage: err.Error()}, err
	}

	return &contextupdate.ContextUpdateResponse{AppContextUpdated: true, AppContextUpdateMessage: "Context updated successfully."}, nil
}

// NewActionControllerServer exported
func NewActionControllerServer() *actionControllerServer {
	s := &actionControllerServer{}
	return s
}
