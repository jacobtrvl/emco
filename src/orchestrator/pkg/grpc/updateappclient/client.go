// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package updateappclient

import (
	"context"
	"time"

	pkgerrors "github.com/pkg/errors"
	inc "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/installappclient"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/rpc"
	updatepb "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/updateapp"
)

const rsyncName = "rsync"

func InvokeUpdateApp(ctx context.Context, FromAppContextID, ToAppContextID string) error {
	var err error
	var rpcClient updatepb.UpdateappClient
	var updateAppRes *updatepb.UpdateAppResponse

	ctx, cancel := context.WithTimeout(ctx, 600*time.Second)
	defer cancel()

	conn := rpc.GetRpcConn(ctx, rsyncName)
	if conn == nil {
		inc.InitRsyncClient()
		conn = rpc.GetRpcConn(ctx, rsyncName)
	}

	if conn != nil {
		rpcClient = updatepb.NewUpdateappClient(conn)
		updateReq := new(updatepb.UpdateAppRequest)
		updateReq.UpdateFromAppContext = FromAppContextID
		updateReq.UpdateToAppContext = ToAppContextID

		updateAppRes, err = rpcClient.UpdateApp(ctx, updateReq)
		if err == nil {
			log.Info("Response from UpdateApp GRPC call", log.Fields{
				"Succeeded": updateAppRes.AppContextUpdated,
				"Message":   updateAppRes.AppContextUpdateMessage,
			})
		}
	} else {
		return pkgerrors.Errorf("UpdateApp Failed - Could not get InstallAppClient: %v", "rsync")
	}

	if err == nil {
		if updateAppRes.AppContextUpdated {
			log.Info("UpdateApp Success", log.Fields{
				"FromAppContext": FromAppContextID,
				"ToAppContext":   ToAppContextID,
				"Message":        updateAppRes.AppContextUpdateMessage,
			})
			return nil
		} else {
			log.Info("UpdateApp Success", log.Fields{
				"FromAppContext": FromAppContextID,
				"ToAppContext":   ToAppContextID,
				"Message":        updateAppRes.AppContextUpdateMessage,
			})
			return pkgerrors.Errorf("UpdateApp Failed: %v", updateAppRes.AppContextUpdateMessage)
		}
	}
	return err
}
