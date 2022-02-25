// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"

	updatepb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/contextupdate"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	contextDb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	register "gitlab.com/project-emco/core/emco-base/src/its/pkg/grpc"
	"gitlab.com/project-emco/core/emco-base/src/its/pkg/grpc/contextupdateserver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
)

func startGrpcServer() error {
	var tls bool

	if strings.Contains(config.GetConfiguration().GrpcEnableTLS, "enable") {
		tls = true
	} else {
		tls = false
	}
	certFile := config.GetConfiguration().GrpcServerCert
	keyFile := config.GetConfiguration().GrpcServerKey

	_, port := register.GetServerHostPort()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Could not listen to port: %v", err)
	}
	var opts []grpc.ServerOption
	if tls {
		if certFile == "" {
			certFile = testdata.Path("server.pem")
		}
		if keyFile == "" {
			keyFile = testdata.Path("server.key")
		}
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			log.Fatalf("Could not generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)
	updatepb.RegisterContextupdateServer(grpcServer, contextupdateserver.NewContextupdateServer())

	log.Println("Starting Istio Traffic Sub Controller gRPC Server")
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("its grpc server is not serving %v", err)
	}
	return err
}

func main() {
	rand.Seed(time.Now().UnixNano())

	err := db.InitializeDatabaseConnection("emco")
	if err != nil {
		log.Println("Unable to initialize database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}
	err = contextDb.InitializeContextDatabase()
	if err != nil {
		log.Println("Unable to initialize database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}
	err = startGrpcServer()
	if err != nil {
		log.Fatalf("GRPC server failed to start")
	}

}
