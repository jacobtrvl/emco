package metrics

import "github.com/prometheus/client_golang/prometheus"

var tacIntentLabel = []string{
	"name", "project", "composite_app", "composite_app_version", "dig",
	"hoot_type", "client_endpoint_name", "client_endpoint_port", "workflow_client_name",
}

var TACIntentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_tac_intent_resource",
	Help: "Count of Temporal Action Controller Intents",
}, tacIntentLabel)
