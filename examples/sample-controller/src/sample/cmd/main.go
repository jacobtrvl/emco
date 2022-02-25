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
	"gitlab.com/project-emco/core/emco-base/examples/sample-controller/api"
	"gitlab.com/project-emco/core/emco-base/examples/sample-controller/pkg/grpc"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/auth"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run initializes the dependencies and start the controller
func run() error {
	rand.Seed(time.Now().UnixNano())

	// Initialize database(s)
	if err := initDataBases(); err != nil {
		return err
	}

	// Initialize grpc server, if required
	initGrpcServer()

	// Handle requests on incoming connections
	if err := serve(); err != nil {
		return err
	}

	return nil
}

// initDataBases initializes the emco databases
func initDataBases() error {
	// Initialize the emco database(Mongo DB)
	err := db.InitializeDatabaseConnection("emco")
	if err != nil {
		logutils.Error("Failed to initialize mongo database connection.",
			logutils.Fields{
				"Error": err})
		return err
	}

	// Initialize etcd
	err = contextdb.InitializeContextDatabase()
	if err != nil {
		logutils.Error("Failed to initialize etcd database connection.",
			logutils.Fields{
				"Error": err})
		return err
	}

	return nil
}

// serve start the controller and handle requests on incoming connections
func serve() error {
	p := config.GetConfiguration().ServicePort

	logutils.Info("Starting controller",
		logutils.Fields{
			"Port": p})

	r := api.NewRouter(nil)
	h := handlers.LoggingHandler(os.Stdout, r)
	server := &http.Server{
		Handler: h,
		Addr:    ":" + p,
	}

	connection := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		server.Shutdown(context.Background())
		close(connection)
	}()

	c, err := auth.GetTLSConfig("ca.cert", "server.cert", "server.key")
	if err != nil {
		logutils.Warn("Failed to get the TLS configuration. Starting without TLS.",
			logutils.Fields{})
		return server.ListenAndServe()
	}

	server.TLSConfig = c
	return server.ListenAndServeTLS("", "") // empty string. tlsconfig already has this information
}

// initGrpcServer start the gRPC server
func initGrpcServer() {
	go func() {
		if err := grpc.StartGrpcServer(); err != nil {
			log.Fatal(err)
		}
	}()
}
