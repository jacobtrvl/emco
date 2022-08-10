// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package grpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/soheilhy/cmux"
	updatepb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/contextupdate"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/statusnotify"
	statusnotifypb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/statusnotify"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	Serve    func() error
	Shutdown func(context.Context) error
}

func RegisterStatusNotifyService(grpcServer *grpc.Server, srv interface{}) {
	statusnotifypb.RegisterStatusNotifyServer(grpcServer, srv.(statusnotify.StatusNotifyServer))
}

func RegisterContextUpdateService(grpcServer *grpc.Server, srv interface{}) {
	updatepb.RegisterContextupdateServer(grpcServer, srv.(updatepb.ContextupdateServer))
}

func StartGrpcServer(defaultName, envName string, defaultPort int, registerFn func(*grpc.Server, interface{}), srv interface{}) error {
	grpcServer, err := newGrpcServer(defaultName, envName, defaultPort, registerFn, srv, "")
	if err != nil {
		return err
	}
	return grpcServer.Serve()
}

func NewGrpcServerWithMetrics(defaultName, envName string, defaultPort int, registerFn func(*grpc.Server, interface{}), srv interface{}) (*GrpcServer, error) {
	return newGrpcServer(defaultName, envName, defaultPort, registerFn, srv, "/metrics")
}

func newGrpcServer(defaultName, envName string, defaultPort int, registerFn func(*grpc.Server, interface{}), srv interface{}, metricsPath string) (*GrpcServer, error) {
	port := getGrpcServerPort(defaultName, envName, defaultPort)

	grpcServer := grpc.NewServer()
	registerFn(grpcServer, srv)

	var httpServer *http.Server
	if metricsPath != "" {
		httpRouter := mux.NewRouter()
		httpRouter.Handle(metricsPath, promhttp.Handler())
		httpServer = &http.Server{
			Handler: httpRouter,
		}
	}

	return &GrpcServer{
		Serve: func() error {
			log.Info("Starting gRPC on port", log.Fields{"Port": port})
			lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err != nil {
				log.Error("Could not listen to gRPC port", log.Fields{"Error": err})
				return err
			}

			m := cmux.New(lis)
			grpcLis := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
			var httpLis net.Listener
			if httpServer != nil {
				httpLis = m.Match(cmux.HTTP1Fast())
			}

			go grpcServer.Serve(grpcLis)
			if httpLis != nil {
				go httpServer.Serve(httpLis)
			}

			err = m.Serve()
			if err != nil {
				log.Error("gRPC server is not serving", log.Fields{"Error": err})
			}
			return err
		},
		Shutdown: func(ctx context.Context) error {
			var err error
			grpcServer.Stop()
			if httpServer != nil {
				err = httpServer.Shutdown(ctx)
				if err != nil {
					log.Error("http server shutdown failed", log.Fields{"Error": err})
				}
			}
			return err
		},
	}, nil
}

func getGrpcServerPort(defaultName, envName string, defaultPort int) int {

	// expect name of this program to be in env the variable "{envName}_NAME" - e.g. ORCHESTRATOR_NAME="orchestrator"
	serviceName := os.Getenv(envName)
	if serviceName == "" {
		serviceName = defaultName
		log.Info("Using default name for service", log.Fields{
			"Name": serviceName,
		})
	}

	// expect service port to be in env variable - e.g. ORCHESTRATOR_SERVICE_PORT
	port, err := strconv.Atoi(os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_PORT"))
	if err != nil || port < 0 {
		port = defaultPort
		log.Info("Using default port for gRPC controller", log.Fields{
			"Name": serviceName,
			"Port": port,
		})
	}
	return port
}
