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

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/auth"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/workflowmgr/api"
)

/*
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
	//updatepb.RegisterContextupdateServer(grpcServer, contextupdateserver.NewContextupdateServer())

	log.Println("Starting Workflow Manager gRPC Server")
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("workflowmgr grpc server is not serving %v", err)
	}
	return err
}
*/

func main() {
	rand.Seed(time.Now().UnixNano())

	err := db.InitializeDatabaseConnection(context.Background(), "emco")
	if err != nil {
		log.Println("Unable to initialize mongo database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}

	/*
		    // workflowmgr does not update appcontext
			err = contextDb.InitializeContextDatabase()
			if err != nil {
				log.Println("Unable to initialize etcd database connection...")
				log.Println(err)
				log.Fatalln("Exiting...")
			}
	*/

	httpRouter := api.NewRouter(nil)
	loggedRouter := handlers.LoggingHandler(os.Stdout, httpRouter)
	log.Println("Starting Workflow Manager")

	httpServer := &http.Server{
		Handler: loggedRouter,
		Addr:    ":" + config.GetConfiguration().ServicePort,
	}
	log.Printf("workflowmgr HTTP server will listen at endpoint: %s", httpServer.Addr)

	/*
		go func() {
			err := startGrpcServer()
			if err != nil {
				log.Fatalf("GRPC server failed to start")
			}
		}()
	*/

	connectionsClose := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		httpServer.Shutdown(context.Background())
		close(connectionsClose)
	}()

	tlsConfig, err := auth.GetTLSConfig("ca.cert", "server.cert", "server.key")
	if err != nil {
		log.Println("Error Getting TLS Configuration. Starting without TLS...")
		log.Fatal(httpServer.ListenAndServe())
	}

	httpServer.TLSConfig = tlsConfig
	// empty strings because tlsconfig already has this information
	err = httpServer.ListenAndServe()
	log.Printf("HTTP server returned error: %s", err)
}
