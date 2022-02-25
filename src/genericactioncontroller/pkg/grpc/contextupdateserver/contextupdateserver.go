package contextupdateserver

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"context"
	"fmt"

	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/internal/action"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/contextupdate"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

type contextupdateServer struct {
	contextupdate.UnimplementedContextupdateServer
}

func (s *contextupdateServer) UpdateAppContext(ctx context.Context,
	req *contextupdate.ContextUpdateRequest) (*contextupdate.ContextUpdateResponse, error) {
	logutils.Info("Received appContext update request",
		logutils.Fields{
			"AppContext": req.AppContext,
			"Intent":     req.IntentName})

	if err := action.UpdateAppContext(req.IntentName, req.AppContext); err != nil {
		return &contextupdate.ContextUpdateResponse{
				AppContextUpdated:       false,
				AppContextUpdateMessage: err.Error()},
			nil
	}

	return &contextupdate.ContextUpdateResponse{
			AppContextUpdated:       true,
			AppContextUpdateMessage: fmt.Sprintf("Successful application of intent %v to %v", req.IntentName, req.AppContext)},
		nil
}

// NewContextUpdateServer exported
func NewContextupdateServer() *contextupdateServer {
	s := &contextupdateServer{}
	return s
}
