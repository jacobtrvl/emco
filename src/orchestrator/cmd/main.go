// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package main

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/handlers"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/api"
	register "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	contextDb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/rpc"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/tracing"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/statusnotify"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()

	err := tracing.InitializeTracer()
	if err != nil {
		log.Error("Unable to initialize tracing", log.Fields{"Error": err})
		os.Exit(1)
	}

	err = db.InitializeDatabaseConnection(ctx, "emco")
	if err != nil {
		log.Error("Unable to initialize mongo database connection", log.Fields{"Error": err})
		os.Exit(1)
	}
	err = contextDb.InitializeContextDatabase()
	if err != nil {
		log.Error("Unable to initialize etcd database connection", log.Fields{"Error": err})
		os.Exit(1)
	}

	httpRouter := api.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	httpRouter.Use(tracing.Middleware)
	loggedRouter := handlers.LoggingHandler(os.Stdout, httpRouter)
	log.Info("Starting Kubernetes Multicloud API", log.Fields{"Port": config.GetConfiguration().ServicePort})

	httpServer := &http.Server{
		Handler: loggedRouter,
		Addr:    ":" + config.GetConfiguration().ServicePort,
	}

	controller.NewControllerClient("resources", "data", "orchestrator").InitControllers(ctx)

	go func() {
		err := register.StartGrpcServer("orchestrator", "ORCHESTRATOR_NAME", 9016,
			register.RegisterStatusNotifyService, statusnotify.StartStatusNotifyServer())
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
		httpServer.Shutdown(ctx)
		rpc.CloseAllRpcConn()
		close(connectionsClose)
	}()

	err = httpServer.ListenAndServe()
	if err != nil {
		log.Error("HTTP server failed", log.Fields{"Error": err})
	}
}
