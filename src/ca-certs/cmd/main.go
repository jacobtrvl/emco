// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/handlers"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/api"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/contextupdate"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run initializes the dependencies and start the controller
func run() error {
	rand.Seed(time.Now().UnixNano())

	// initialize database(s)
	if err := initDataBases(); err != nil {
		return err
	}

	// initialize grpc server, if required
	initGrpcServer()

	// handle requests on incoming connections
	serve()

	return nil
}

// initDataBases initializes the emco databases
func initDataBases() error {
	// initialize the emco database(Mongo DB)
	err := db.InitializeDatabaseConnection(context.Background(), "emco")
	if err != nil {
		logutils.Error("Failed to initialize mongo database connection.",
			logutils.Fields{
				"Error": err.Error()})
		return err
	}

	// initialize etcd
	err = contextdb.InitializeContextDatabase()
	if err != nil {
		logutils.Error("Failed to initialize etcd database connection.",
			logutils.Fields{
				"Error": err.Error()})
		return err
	}

	return nil
}

// serve start the controller and handle requests on incoming connections
func serve() {
	logutils.Info("Starting CaCert Controller", logutils.Fields{"Port": config.GetConfiguration().ServicePort})

	r := api.NewRouter(nil)
	h := handlers.LoggingHandler(os.Stdout, r)
	server := &http.Server{
		Handler: h,
		Addr:    ":" + config.GetConfiguration().ServicePort,
	}

	connection := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		server.Shutdown(context.Background())
		close(connection)
	}()

	if err := server.ListenAndServe(); err != nil {
		logutils.Error("HTTP server failed",
			logutils.Fields{
				"Error": err.Error()})
	}
}

// initGrpcServer start the gRPC server
func initGrpcServer() {
	type contextupdateServer struct {
		contextupdate.UnimplementedContextupdateServer
	}

	go func() {
		err := grpc.StartGrpcServer("ca-certs", "CACERT_NAME", 9035,
			grpc.RegisterContextUpdateService, &contextupdateServer{})
		if err != nil {
			logutils.Error("GRPC server failed to start",
				logutils.Fields{
					"Error": err.Error()})
			os.Exit(1)
		}
	}()
}
