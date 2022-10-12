package metrics

import "github.com/prometheus/client_golang/prometheus"

var trafficGroupIntentLabel = []string{"name", "project", "composite_app", "composite_app_version", "dig"}
var inboundIntentLabel = []string{
	"name",
	"project",
	"composite_app",
	"composite_app_version",
	"dig",
	"traffic_group_intent",
	"spec_app",
	"app_label",
	"serviceName",
	"externalName",
	"port",
	"protocol",
	"externalSupport",
	"serviceMesh",
	"sidecarProxy",
	"tlsType",
}
var inboundIntentClientLabel = []string{
	"name",
	"project",
	"composite_app",
	"composite_app_version",
	"dig",
	"traffic_group_intent",
	"inbound_intent",
	"spec_app",
	"app_label",
	"serviceName",
}

var inboundIntentClientAPLabel = []string{
	"name",
	"project",
	"composite_app",
	"composite_app_version",
	"dig",
	"traffic_group_intent",
	"inbound_intent",
	"client_name",
	"action",
}

var TrafficGroupIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_dig_traffic_group_intent_resource",
	Help: "Count of Traffic Group Intents",
}, trafficGroupIntentLabel)

var InboundIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_dig_inbound_intent_resource",
	Help: "Count of Inbound Intents",
}, inboundIntentLabel)

var InboundIntentClientGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_dig_inbound_intent_client_resource",
	Help: "Count of Inbound Intent Clients",
}, inboundIntentClientLabel)

var InboundIntentClientAPGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_dig_inbound_intent_client_access_point_resource",
	Help: "Count of Inbound Intent Client Access Points",
}, inboundIntentClientAPLabel)
