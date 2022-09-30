package metrics

import "github.com/prometheus/client_golang/prometheus"

var networkControllerIntentLabel = []string{"name", "project", "composite_app", "composite_app_version", "dig"}
var workloadIntentLabel = []string{"name", "project", "composite_app", "composite_app_version", "dig", "network_controller_intent", "app_label", "workload_resource", "type"}
var workloadInterfaceIntentLabel = []string{
	"name",
	"project",
	"composite_app",
	"composite_app_version",
	"dig",
	"network_controller_intent",
	"workload_intent",
	"interface",
	"network_name",
	"default_gateway",
	"ip_address",
	"mac_address",
}

var NetworkControllerIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_network_controller_intent_resource",
	Help: "Count of Network Controller Intents",
}, networkControllerIntentLabel)

var WorkloadIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_workload_intent_resource",
	Help: "Count of Workload Intents",
}, workloadIntentLabel)

var WorkloadInterfaceIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_workload_interface_intent_resource",
	Help: "Count of Workload Interface Intents",
}, workloadInterfaceIntentLabel)
