// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	register "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/metrics"
	installpb "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/installapp"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/installappserver"
	readynotifypb "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/readynotify"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/readynotifyserver"
	updatepb "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/updateapp"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/updateappserver"

	contextDb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	con "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/context"
	"google.golang.org/grpc"
)

func RegisterRsyncServices(grpcServer *grpc.Server, srv interface{}) {
	installpb.RegisterInstallappServer(grpcServer, installappserver.NewInstallAppServer())
	readynotifypb.RegisterReadyNotifyServer(grpcServer, readynotifyserver.NewReadyNotifyServer())
	updatepb.RegisterUpdateappServer(grpcServer, updateappserver.NewUpdateAppServer())
}

func main() {

	rand.Seed(time.Now().UnixNano())

	prometheus.MustRegister(metrics.NewBuildInfoCollector("orchestrator"))

	// Initialize the mongodb
	err := db.InitializeDatabaseConnection("emco")
	if err != nil {
		log.Error("Unable to initialize mongo database connection", log.Fields{"Error": err})
		os.Exit(1)
	}

	// Initialize contextdb
	err = contextDb.InitializeContextDatabase()
	if err != nil {
		log.Error("Unable to initialize etcd database connection", log.Fields{"Error": err})
		os.Exit(1)
	}

	grpcServer, err := register.NewGrpcServerWithMetrics("rsync", "RSYNC_NAME", 9031,
		RegisterRsyncServices, nil)
	if err != nil {
		log.Error("Unable to create grpc server", log.Fields{"Error": err})
		os.Exit(1)
	}

	err = con.RestoreActiveContext()
	if err != nil {
		log.Error("RestoreActiveContext failed", log.Fields{"Error": err})
	}

	connectionsClose := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		grpcServer.Shutdown(context.Background())
		close(connectionsClose)
	}()

	err = grpcServer.Serve()
	if err != nil {
		log.Error("gRPC server failed", log.Fields{"Error": err})
	}
}
