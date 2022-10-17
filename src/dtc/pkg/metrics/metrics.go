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
	metrics.Gauges["TrafficGroupIntentGauge"] = TrafficGroupIntentGauge
	metrics.Gauges["InboundIntentGauge"] = InboundIntentGauge
	metrics.Gauges["InboundIntentClientGauge"] = InboundIntentClientGauge
	metrics.Gauges["InboundIntentClientAPGauge"] = InboundIntentClientAPGauge

	prometheus.MustRegister(TrafficGroupIntentGauge)
	prometheus.MustRegister(InboundIntentGauge)
	prometheus.MustRegister(InboundIntentClientGauge)
	prometheus.MustRegister(InboundIntentClientAPGauge)

	prometheus.MustRegister(infra.NewBuildInfoCollector("dtc"))

	if startNow {
		start()
	}
	metrics.Handler = promhttp.Handler()
	return metrics
}
