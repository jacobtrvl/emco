// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package cmd

import (
	"fmt"
	"os"
	"strconv"
)

// Configurations exported
type EmcoConfigurations struct {
	Ingress      ControllerConfigurations
	Orchestrator ControllerConfigurations
	Clm          ControllerConfigurations
	Ncm          ControllerConfigurations
	Dcm          ControllerConfigurations
	Rsync        ControllerConfigurations
	OvnAction    ControllerConfigurations
	Gac          ControllerConfigurations
	Dtc          ControllerConfigurations
	HpaPlacement ControllerConfigurations
	Sfc          ControllerConfigurations
	SfcClient    ControllerConfigurations
	WorkflowMgr  ControllerConfigurations
	CaCert       ControllerConfigurations
}

// ControllerConfigurations exported
type ControllerConfigurations struct {
	Port       int
	StatusPort int
	Host       string
}

const urlVersion string = "v2"
const urlPrefix string = "http://"

var Configurations EmcoConfigurations

// SetDefaultConfiguration default configuration if t
func SetDefaultConfiguration() {
	Configurations.Orchestrator.Host = "localhost"
	Configurations.Orchestrator.Port = 9015
	Configurations.Orchestrator.StatusPort = 9016
	Configurations.Clm.Host = "localhost"
	Configurations.Clm.Port = 9061
	Configurations.Ncm.Host = "localhost"
	Configurations.Ncm.Port = 9081
	Configurations.Ncm.StatusPort = 9082
	Configurations.Dcm.Host = "localhost"
	Configurations.Dcm.Port = 9077
	Configurations.Dcm.StatusPort = 9078
	Configurations.OvnAction.Host = "localhost"
	Configurations.OvnAction.Port = 9051
	Configurations.Dtc.Host = "localhost"
	Configurations.Dtc.Port = 9018
	Configurations.Gac.Host = "localhost"
	Configurations.Gac.Port = 9020
	Configurations.HpaPlacement.Host = "localhost"
	Configurations.HpaPlacement.Port = 9091
	Configurations.Sfc.Host = "localhost"
	Configurations.Sfc.Port = 9055
	Configurations.SfcClient.Host = "localhost"
	Configurations.SfcClient.Port = 9057
	Configurations.WorkflowMgr.Host = "localhost"
	Configurations.WorkflowMgr.Port = 9095
	Configurations.CaCert.Host = "localhost"
	Configurations.CaCert.Port = 9036
}

// GetIngressURL Url for Ingress
func GetIngressURL() string {
	if Configurations.Ingress.Host == "" || Configurations.Ingress.Port == 0 {
		return ""
	}
	return urlPrefix + Configurations.Ingress.Host + ":" + strconv.Itoa(Configurations.Ingress.Port) + "/" + urlVersion
}

// GetOrchestratorURL Url for Orchestrator
func GetOrchestratorURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Orchestrator.Host == "" || Configurations.Orchestrator.Port == 0 {
		fmt.Println("Fatal: No Orchestrator Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + Configurations.Orchestrator.Host + ":" + strconv.Itoa(Configurations.Orchestrator.Port) + "/" + urlVersion
}

// GetClmURL Url for Clm
func GetClmURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Clm.Host == "" || Configurations.Clm.Port == 0 {
		fmt.Println("Fatal: No Clm Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + Configurations.Clm.Host + ":" + strconv.Itoa(Configurations.Clm.Port) + "/" + urlVersion
}

// GetNcmURL Url for Ncm
func GetNcmURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Ncm.Host == "" || Configurations.Ncm.Port == 0 {
		fmt.Println("Fatal: No Ncm Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + Configurations.Ncm.Host + ":" + strconv.Itoa(Configurations.Ncm.Port) + "/" + urlVersion
}

// GetDcmURL Url for Dcm
func GetDcmURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Dcm.Host == "" || Configurations.Dcm.Port == 0 {
		fmt.Println("Fatal: No Dcm Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + Configurations.Dcm.Host + ":" + strconv.Itoa(Configurations.Dcm.Port) + "/" + urlVersion
}

// GetOvnactionURL Url for Ovnaction
func GetOvnactionURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.OvnAction.Host == "" || Configurations.OvnAction.Port == 0 {
		fmt.Println("Fatal: No Ovn Action Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + Configurations.OvnAction.Host + ":" + strconv.Itoa(Configurations.OvnAction.Port) + "/" + urlVersion
}

// GetDtcURL Url for Dtc
func GetDtcURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Dtc.Host == "" || Configurations.Dtc.Port == 0 {
		fmt.Println("Fatal: No DTC Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + Configurations.Dtc.Host + ":" + strconv.Itoa(Configurations.Dtc.Port) + "/" + urlVersion
}

// GetGacURL Url for gac
func GetGacURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Gac.Host == "" || Configurations.Gac.Port == 0 {
		fmt.Println("Fatal: No GAC Action Information in Config File!")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + Configurations.Gac.Host + ":" + strconv.Itoa(Configurations.Gac.Port) + "/" + urlVersion
}

// GetHpaPlacementURL Url for Hpa Placement controller
func GetHpaPlacementURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.HpaPlacement.Host == "" || Configurations.HpaPlacement.Port == 0 {
		fmt.Println("Fatal: No HPA Placement Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + Configurations.HpaPlacement.Host + ":" + strconv.Itoa(Configurations.HpaPlacement.Port) + "/" + urlVersion
}

// GetSfcURL Url for sfc
func GetSfcURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Sfc.Host == "" || Configurations.Sfc.Port == 0 {
		fmt.Println("Fatal: No SFC Action Information in Config File!")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + Configurations.Sfc.Host + ":" + strconv.Itoa(Configurations.Sfc.Port) + "/" + urlVersion
}

// GetSfcClientURL Url for sfcclient
func GetSfcClientURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.SfcClient.Host == "" || Configurations.SfcClient.Port == 0 {
		fmt.Println("Fatal: No SFC Client Action Information in Config File!")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + Configurations.SfcClient.Host + ":" + strconv.Itoa(Configurations.SfcClient.Port) + "/" + urlVersion
}

func GetWorkflowMgrURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.WorkflowMgr.Host == "" || Configurations.WorkflowMgr.Port == 0 {
		fmt.Println("Fatal: No WorkflowMgr Information in Config File!")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + Configurations.WorkflowMgr.Host + ":" + strconv.Itoa(Configurations.WorkflowMgr.Port) + "/" + urlVersion
}

// GetOrchestratorGrpcEndpoint gRPC endpoint for Orchestrator
func GetOrchestratorGrpcEndpoint() string {
	if Configurations.Orchestrator.Host == "" || Configurations.Orchestrator.StatusPort == 0 {
		fmt.Println("Fatal: No Orchestrator StatusPort Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return Configurations.Orchestrator.Host + ":" + strconv.Itoa(Configurations.Orchestrator.StatusPort)
}

// GetNcmGrpcEndpoint gRPC endpoint for ncm
func GetNcmGrpcEndpoint() string {
	if Configurations.Ncm.Host == "" || Configurations.Ncm.StatusPort == 0 {
		fmt.Println("Fatal: No NCM StatusPort Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return Configurations.Ncm.Host + ":" + strconv.Itoa(Configurations.Ncm.StatusPort)
}

// GetDcmGrpcEndpoint gRPC endpoint for dcm
func GetDcmGrpcEndpoint() string {
	if Configurations.Dcm.Host == "" || Configurations.Dcm.StatusPort == 0 {
		fmt.Println("Fatal: No DCM StatusPort Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return Configurations.Dcm.Host + ":" + strconv.Itoa(Configurations.Dcm.StatusPort)
}

// GetCaCertUrl construct the baseUrl for CaCert controller
func GetCaCertUrl() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}

	if Configurations.CaCert.Host == "" ||
		Configurations.CaCert.Port == 0 {
		fmt.Println("Fatal: No CA Cert Information in Config File!")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + Configurations.CaCert.Host + ":" + strconv.Itoa(Configurations.CaCert.Port) + "/" + urlVersion
}
