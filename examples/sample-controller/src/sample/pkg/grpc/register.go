// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

// Package grpc creates a new network listener using the
// provided port and a gRPC server. Then register the
// service and its implementation to the gRPC server.
// In EMCO, each controller communicates with the
// application scheduler(orchestrator) through the gRPC calls.
package grpc

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	actioncontroller "gitlab.com/project-emco/core/emco-base/examples/sample-controller/pkg/grpc/action-controller"
	placementcontroller "gitlab.com/project-emco/core/emco-base/examples/sample-controller/pkg/grpc/placement-controller"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/contextupdate"
	orchplacementcontroller "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/placementcontroller"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"google.golang.org/grpc"
)

const default_host = "localhost"
const default_port = 9050
const default_service_name = "sample"
const ENV_SERVICE_NAME = "SERVICE_NAME"

// StartGrpcServer start the gRPC server and register with the application scheduler(orchestrator)
func StartGrpcServer() error {
	logutils.Info("Initializing the controller gRPC server",
		logutils.Fields{})

	port := getServerPort()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logutils.Error("Could not listen to port",
			logutils.Fields{
				"Port": port})
		return err
	}

	server := grpc.NewServer([]grpc.ServerOption{}...)

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
	contextupdate.RegisterContextupdateServer(server, actioncontroller.NewActionControllerServer())

	// Register the placement controller
	orchplacementcontroller.RegisterPlacementControllerServer(server, placementcontroller.NewPlacementControllerServer())

	logutils.Info("Starting the controller gRPC server",
		logutils.Fields{
			"Port": port})

	if err = server.Serve(listener); err != nil {
		logutils.Error("Failed to start the gRPC server",
			logutils.Fields{
				"Error": err})
		return err
	}
	return nil
}

// getServerPort returns the gRPC port
func getServerPort() int {
	// Expect the service name to be available in the environment variable.
	// eg: SERVICE_NAME=sample
	name := os.Getenv(ENV_SERVICE_NAME)
	if name == "" {
		name = default_service_name
		logutils.Warn("Using the default name as the service name",
			logutils.Fields{
				"Name": name})
	}

	// Expect the host name to be available in the environment variable.
	host := os.Getenv(strings.ToUpper(name) + "_SERVICE_HOST")
	if host == "" {
		host = default_host
		logutils.Warn("Using the default host as the gRPC controller",
			logutils.Fields{
				"Host": host})
	}

	// Expect the port to be available in the environment variable.
	port, err := strconv.Atoi(os.Getenv(strings.ToUpper(name) + "_SERVICE_PORT"))
	if err != nil || port < 0 {
		port = default_port
		logutils.Warn("Using the default port for the gRPC controller",
			logutils.Fields{
				"Port": port})
	}

	return port
}
