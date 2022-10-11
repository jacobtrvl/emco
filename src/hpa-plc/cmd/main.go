// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package main

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/handlers"
	clmcontrollerpb "gitlab.com/project-emco/core/emco-base/src/clm/pkg/grpc/controller-eventchannel"
	"gitlab.com/project-emco/core/emco-base/src/hpa-plc/api"
	clmControllerserver "gitlab.com/project-emco/core/emco-base/src/hpa-plc/pkg/grpc/clmcontrollereventchannelserver"
	placementcontrollerserver "gitlab.com/project-emco/core/emco-base/src/hpa-plc/pkg/grpc/hpaplacementcontrollerserver"
	register "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc"
	plsctrlclientpb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/placementcontroller"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	contextDb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"google.golang.org/grpc"
)

func RegisterHpaPlacementServices(grpcServer *grpc.Server, srv interface{}) {
	plsctrlclientpb.RegisterPlacementControllerServer(grpcServer, placementcontrollerserver.NewHpaPlacementControllerServer())
	clmcontrollerpb.RegisterClmControllerEventChannelServer(grpcServer, clmControllerserver.NewControllerEventchannelServer())
}

func main() {
	rand.Seed(time.Now().UnixNano())

	err := db.InitializeDatabaseConnection(context.Background(), "emco")
	if err != nil {
		log.Error("Unable to initialize mongo database connection", log.Fields{"Error": err})
		os.Exit(1)
	}
	err = contextDb.InitializeContextDatabase()
	if err != nil {
		log.Error("Unable to initialize etcd database connection", log.Fields{"Error": err})
		os.Exit(1)
	}

	httpRouter := api.NewRouter(nil)
	loggedRouter := handlers.LoggingHandler(os.Stdout, httpRouter)
	httpServer := &http.Server{
		Handler: loggedRouter,
		Addr:    ":" + config.GetConfiguration().ServicePort,
	}
	log.Info("Starting HPA PlacementController Http Server", log.Fields{"Port": config.GetConfiguration().ServicePort})

	go func() {
		err := register.StartGrpcServer("hpaplacement", "HPAPLACEMENT_NAME", 9099,
			RegisterHpaPlacementServices, nil)
		if err != nil {
			log.Error("GRPC server failed to start", log.Fields{"Error": err})
			os.Exit(1)
		}
	}()

	connectionsClose := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		err := httpServer.Shutdown(context.Background())
		if err != nil {
			log.Error("HTTP server failed to shutdown", log.Fields{"Error": err})
			os.Exit(1)
		}
		close(connectionsClose)
	}()

	err = httpServer.ListenAndServe()
	if err != nil {
		log.Error("HTTP server failed", log.Fields{"Error": err})
	}
}
