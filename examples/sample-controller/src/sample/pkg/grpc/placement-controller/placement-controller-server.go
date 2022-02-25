// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

// Package placementcontroller exposes the gRPC server-specific functionalities of a placement controller.
// These functions are getting invoked by the application scheduler(orchestrator).
// The implementation of these gRPC functionalities, and their associated packages
//  (in this case, placementcontroller), will change based on the requirement(s).
// For example, a controller can be one of the placement types or actions.
// Placement controllers allow the orchestrator to choose the exact locations
// to place the application in the composite application. On the other hand,
// action controllers can modify the state of a resource(create additional resources
// to be deployed, modify or delete the existing resources). You can define your packages
// and functionalities based on your need and expose these functionalities using the gRPC server.
// In EMCO, we have separate controllers for action and placement.
// ref: https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/hpa-plc/pkg/grpc placement controller
// ref: https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/genericactioncontroller/pkg/grpc action controller
package placementcontroller

import (
	"context"
	"errors"

	"gitlab.com/project-emco/core/emco-base/examples/sample-controller/internal/action"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/placementcontroller"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

type placementControllerServer struct {
}

func (ac *placementControllerServer) FilterClusters(ctx context.Context, req *placementcontroller.ResourceRequest) (*placementcontroller.ResourceResponse, error) {
	if (req != nil) && (len(req.AppContext) > 0) {
		err := action.FilterClusters(req.AppContext)
		if err != nil {
			return &placementcontroller.ResourceResponse{AppContext: req.AppContext, Status: false, Message: "Failed to filter clusters."}, err
		}

		return &placementcontroller.ResourceResponse{AppContext: req.AppContext, Status: true, Message: ""}, nil
	}

	logutils.Error("Invalid request.",
		logutils.Fields{
			"Request": req})
	return &placementcontroller.ResourceResponse{Status: false, Message: "Invalid request."}, errors.New("invalid request")

}

// NewPlacementControllerServer exported
func NewPlacementControllerServer() *placementControllerServer {
	s := &placementControllerServer{}
	return s
}
