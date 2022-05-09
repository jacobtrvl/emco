// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package notifyclient

import (
	"context"
	"fmt"
	"time"

	pkgerrors "github.com/pkg/errors"
	"google.golang.org/grpc"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/installappclient"
	rsyncclient "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/installappclient"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/updateappclient"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/rpc"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
	readynotifypb "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/readynotify"
)

const rsyncName = "rsync"

/*
RsyncInfo consists of rsyncName, hostName and portNumber.
*/
type RsyncInfo struct {
	RsyncName  string
	hostName   string
	portNumber int
}

// queryDBAndSetRsyncInfo queries the MCO db to find the record the sync controller and then sets the RsyncInfo global variable
func queryDBAndSetRsyncInfo() (installappclient.RsyncInfo, error) {
	client := controller.NewControllerClient("resources", "data", "orchestrator")
	vals, _ := client.GetControllers()
	for _, v := range vals {
		if v.Metadata.Name == rsyncName {
			log.Info("Initializing RPC connection to resource synchronizer", log.Fields{
				"Controller": v.Metadata.Name,
			})
			rsyncInfoInstallAppClient := installappclient.NewRsyncInfo(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
			return rsyncInfoInstallAppClient, nil
		}
	}
	return installappclient.RsyncInfo{}, pkgerrors.Errorf("queryRsyncInfoInMCODB Failed - Could not get find rsync by name : %v", rsyncName)
}

/*
CallRsyncInstall method shall take in the app context id and invokes the rsync service via grpc
*/
func CallRsyncInstall(contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo()
	log.Info("Calling the Rsync ", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	err = rsyncclient.InvokeInstallApp(appContextID)
	if err != nil {
		return err
	}
	return nil
}

/*
CallRsyncUpdate method shall take in the new and existing appContextID and invokes the rsync service via grpc
*/

func CallRsyncUpdate(from, to interface{}) error {
	if _, err := queryDBAndSetRsyncInfo(); err != nil {
		return err
	}

	fromAppContextID := fmt.Sprintf("%v", from)
	toAppContextID := fmt.Sprintf("%v", to)
	if err := updateappclient.InvokeUpdateApp(fromAppContextID, toAppContextID); err != nil {
		return err
	}

	return nil
}

/*
CallRsyncUninstall method shall take in the app context id and invokes the rsync service via grpc
*/

func CallRsyncUninstall(contextid interface{}) error {
	if _, err := queryDBAndSetRsyncInfo(); err != nil {
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	if err := rsyncclient.InvokeUninstallApp(appContextID); err != nil {
		return err
	}

	return nil
}

// InvokeReadyNotify will make a gRPC call to the resource synchronizer and will
// subscribe the clients to alerts from the rsync gRPC server ("ready-notify")
func InvokeReadyNotify(appContextID, clientName string) (readynotifypb.ReadyNotify_AlertClient, readynotifypb.ReadyNotifyClient, error) {

	var stream readynotifypb.ReadyNotify_AlertClient
	var client readynotifypb.ReadyNotifyClient
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rsyncInfo, err := queryDBAndSetRsyncInfo()
	log.Info("Calling rsync", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return stream, client, pkgerrors.Wrapf(err, "Unable to find the rsync info from EMCO db")
	}

	conn := rpc.GetRpcConn(rsyncName)
	if conn == nil {
		installappclient.InitRsyncClient()
		conn = rpc.GetRpcConn(rsyncName)
		if conn == nil {
			return stream, client, pkgerrors.Wrapf(err, "Unable to connect to rsync")
		}
	}

	client = readynotifypb.NewReadyNotifyClient(conn)

	if client != nil {
		stream, err = client.Alert(context.Background(), &readynotifypb.Topic{ClientName: clientName, AppContext: appContextID}, grpc.WaitForReady(true))
		if err != nil {
			log.Error("[ReadyNotify gRPC] Failed to subscribe to alerts", log.Fields{"err": err, "clientName": clientName, "appContextId": appContextID})
			time.Sleep(5 * time.Second)
			InvokeReadyNotify(appContextID, clientName)
		}

		log.Info("[ReadyNotify gRPC] Subscribing to alerts about appcontext ID", log.Fields{"appContextId": appContextID})

	}

	return stream, client, nil
}
