// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package main

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/handlers"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/api"

	// "gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/grpc/contextupdateserver"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/grpc/contextupdateserver"
	register "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	contextDb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	err := db.InitializeDatabaseConnection("emco")
	if err != nil {
		logutils.Error("Unable to initialize mongo database connection", logutils.Fields{"Error": err})
		os.Exit(1)
	}
	err = contextDb.InitializeContextDatabase()
	if err != nil {
		logutils.Error("Unable to initialize etcd database connection", logutils.Fields{"Error": err})
		os.Exit(1)
	}

	httpRouter := api.NewRouter(nil)
	loggedRouter := handlers.LoggingHandler(os.Stdout, httpRouter)
	logutils.Info("Starting Certificate Controller", logutils.Fields{"Port": config.GetConfiguration().ServicePort})

	httpServer := &http.Server{
		Handler: loggedRouter,
		Addr:    ":" + config.GetConfiguration().ServicePort,
	}

	go func() {
		err := register.StartGrpcServer("cacert", "CACERT_NAME", 9035,
			register.RegisterContextUpdateService, contextupdateserver.NewContextupdateServer())
		if err != nil {
			logutils.Error("GRPC server failed to start", logutils.Fields{"Error": err})
			os.Exit(1)
		}
	}()

	connectionsClose := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		httpServer.Shutdown(context.Background())
		close(connectionsClose)
	}()

	err = httpServer.ListenAndServe()
	if err != nil {
		logutils.Error("HTTP server failed", logutils.Fields{"Error": err})
	}
}
