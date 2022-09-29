package metrics

import "github.com/prometheus/client_golang/prometheus"

var networkLabel = []string{"clusterprovider", "cluster", "name", "cnitype"}
var providerNetworkLabel = []string{"clusterprovider", "cluster", "name", "cnitype", "nettype", "vlanid", "providerinterfacename", "logicalinterfacename", "vlannodeselector"}

var NetworkGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_cluster_network_resource",
	Help: "Count of Cluster Networks",
}, networkLabel)

var ProviderNetworkGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_cluster_provider_network_resource",
	Help: "Count of Cluster Provider Networks",
}, providerNetworkLabel)
