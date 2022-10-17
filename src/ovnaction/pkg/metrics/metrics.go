// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	infra "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/metrics"
)

type Metrics struct {
	Gauges  map[string]*prometheus.GaugeVec
	Handler http.Handler
}

func Initialize(startNow bool) *Metrics {
	metrics := &Metrics{
		Gauges: make(map[string]*prometheus.GaugeVec),
	}
	metrics.Gauges["NetworkControllerIntentGauge"] = NetworkControllerIntentGauge
	metrics.Gauges["WorkloadIntentGauge"] = WorkloadIntentGauge
	metrics.Gauges["WorkloadInterfaceIntentGauge"] = WorkloadInterfaceIntentGauge

	prometheus.MustRegister(NetworkControllerIntentGauge)
	prometheus.MustRegister(WorkloadIntentGauge)
	prometheus.MustRegister(WorkloadInterfaceIntentGauge)

	prometheus.MustRegister(infra.NewBuildInfoCollector("ovnaction"))

	if startNow {
		start()
	}
	metrics.Handler = promhttp.Handler()
	return metrics
}
