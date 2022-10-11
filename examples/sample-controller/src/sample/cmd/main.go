// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"gitlab.com/project-emco/core/emco-base/examples/sample-controller/api"
	actioncontroller "gitlab.com/project-emco/core/emco-base/examples/sample-controller/pkg/grpc/action-controller"
	placementcontroller "gitlab.com/project-emco/core/emco-base/examples/sample-controller/pkg/grpc/placement-controller"
	register "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/contextupdate"
	orchplacementcontroller "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/placementcontroller"
	contextDb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
	"google.golang.org/grpc"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()

	// Initialize the emco database(Mongo DB)
	err := db.InitializeDatabaseConnection(ctx, "emco")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize etcd
	err = contextDb.InitializeContextDatabase()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize gRPC server, if required
	grpcServer, err := register.NewGrpcServer("sample", "SERVICE_NAME", 9025,
		RegisterServices, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the EMCO databases and initialize HTTP server, if required
	server, err := controller.NewControllerServer("sample",
		api.NewRouter(nil),
		grpcServer)
	if err != nil {
		log.Fatal(err)
	}

	// Start the gRPC and HTTP controller and handle requests on incoming connections
	connection := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		server.Shutdown(ctx)
		close(connection)
	}()

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func RegisterServices(grpcServer *grpc.Server, srv interface{}) {
	// A controller can be one of the placement types or actions.
	// Placement controllers allow the orchestrator to choose the exact locations
	// to place the application in the composite application.
	// Action controllers can modify the state of a resource(create additional resources
	// to be deployed, modify or delete the existing resources).
	// You can build your controller and define packages and functionalities based on your need.
	// In this sample controller, we have shown how to register the action and placement controllers.
	// Registering the same controller as action and placement may or may not work.
	// This is for illustration purposes only since the code structure is the same for action or placement controller.
	// In EMCO, we have separate controllers for action and placement.
	// ref: https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/hpa-ac - HPA action controller
	// ref: https://gitlab.com/project-emco/core/emco-base/-/tree/main/src/hpa-plc - HPA placement controller

	// Register the action controller
	contextupdate.RegisterContextupdateServer(grpcServer, actioncontroller.NewActionControllerServer())

	// Register the placement controller
	orchplacementcontroller.RegisterPlacementControllerServer(grpcServer, placementcontroller.NewPlacementControllerServer())
}
