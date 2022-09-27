package metrics

import "github.com/prometheus/client_golang/prometheus"

var logicalCloudLabel = []string{"project", "name", "namespace", "status"}

var LCGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "emco_logical_cloud_resource",
	Help: "Count of Logical Clouds",
}, logicalCloudLabel)
