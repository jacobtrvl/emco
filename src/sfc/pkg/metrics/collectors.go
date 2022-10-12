package metrics

import "github.com/prometheus/client_golang/prometheus"

var sfcIntentLabel = []string{
	"name", "project", "composite_app", "composite_app_version", "dig",
	"chainType", "namespace",
}

var sfcIntentLinkLabel = []string{
	"name", "project", "composite_app", "composite_app_version", "dig", "sfc",
	"left_net",
	"right_net",
	"link_label",
	"app_name",
	"workload_resource",
	"resource_type",
}

var SFCIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_sfc_intent_resource",
	Help: "Count of Network Chain Intents",
}, sfcIntentLabel)

var SFCIntentLinkGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_sfc_intent_link_resource",
	Help: "Count of Network Chain Intent Links",
}, sfcIntentLinkLabel)
