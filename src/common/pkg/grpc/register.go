// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package grpc

import (
	"os"
	"strconv"
	"strings"

	log "gitlab.com/project-emco/core/emco-base/src/common/pkg/infra/logutils"
)

const default_host = "localhost"

func GetServerHostPort(defaultName, envName string, defaultPort int) (string, int) {

	// expect name of this program to be in env the variable "{envName}_NAME" - e.g. ORCHESTRATOR_NAME="orchestrator"
	serviceName := os.Getenv(envName)
	if serviceName == "" {
		serviceName = defaultName
		log.Info("Using default name for service", log.Fields{
			"Name": serviceName,
		})
	}

	// expect service name to be in env variable - e.g. ORCHESTRATOR_SERVICE_HOST
	host := os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_HOST")
	if host == "" {
		host = default_host
		log.Info("Using default host for gRPC controller", log.Fields{
			"Host": host,
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
	return host, port
}
