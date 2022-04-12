package grpc

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"os"
	"strconv"
	"strings"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

const DEFAULTHOST = "localhost"
const DEFAULTPORT = 9034
const DEFAULTAPPCONFIGCONTROLLER_NAME = "appconfig"
const ENV_APP_CONFIG_CONTROLLER_NAME = "APPCONFIGCONTROLLER_NAME"

func GetServerHostPort() (string, int) {

	// expect name of this contoller program to be in env variable "APP_CONFIG_CONTROLLER_NAME" - e.g. APP_CONFIG_CONTROLLER_NAME="app_config_controller"
	serviceName := os.Getenv(ENV_APP_CONFIG_CONTROLLER_NAME)
	if serviceName == "" {
		serviceName = DEFAULTAPPCONFIGCONTROLLER_NAME
		log.Info("Using default name for app-config-controller service name", log.Fields{
			"Name": serviceName,
		})
	}

	// expect service name to be in env variable - e.g. APP_CONFIG_CONTROLLER_SERVICE_HOST
	host := os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_HOST")
	if host == "" {
		host = DEFAULTHOST
		log.Info("Using default host for app-config-controller gRPC controller", log.Fields{
			"Host": host,
		})
	}

	// expect service port to be in env variable - e.g. APP_CONFIG_CONTROLLER_SERVICE_PORT
	port, err := strconv.Atoi(os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_PORT"))
	if err != nil || port < 0 {
		port = DEFAULTPORT
		log.Info("Using default port for app-config-controller gRPC controller", log.Fields{
			"Port": port,
		})
	}
	return host, port
}
