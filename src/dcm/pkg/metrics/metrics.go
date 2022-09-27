// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package metrics

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	Gauges  map[string]*prometheus.GaugeVec
	Handler http.Handler
}

func Initialize(startNow bool) *Metrics {
	metrics := &Metrics{
		Gauges: make(map[string]*prometheus.GaugeVec),
	}
	metrics.Gauges["LCGauge"] = LCGauge

	prometheus.MustRegister(LCGauge)

	prometheus.MustRegister(NewBuildInfoCollector("dcm"))

	if startNow {
		start()
	}
	metrics.Handler = promhttp.Handler()
	return metrics
}

func NewBuildInfoCollector(component string) *prometheus.GaugeVec {
	build := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "emco_build",
			Help: fmt.Sprintf(
				"A metric with a constant '1' value labeled by component, version, and revision from which %s was built.",
				component,
			),
		},
		[]string{
			"component",
			"revision",
			"version",
		})

	build.WithLabelValues(component, os.Getenv("EMCO_META_EMCO_SHA"), os.Getenv("EMCO_META_EMCO_VERSION")).Set(1)

	return build
}
