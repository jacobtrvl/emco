package metrics

import "github.com/prometheus/client_golang/prometheus"

var CLPLabel = []string{"name"}
var ClusterLabel = []string{"name", "clusterprovider"}

var CLPGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_cluster_provider_resource",
	Help: "Count of Cluster Providers",
}, CLPLabel)

var ClusterGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_cluster_resource",
	Help: "Count of Clusters",
}, ClusterLabel)
