/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package agents

import (
	"fmt"
	"net/http"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusAgent handles metrics collection for Prometheus.
// It collects stats from StatQueues and exposes them alongside
// optional Go runtime and process metrics.
type PrometheusAgent struct {
	cfg      *config.CGRConfig
	filters  *engine.FilterS
	cm       *engine.ConnManager
	shutdown *utils.SyncedChan

	handler     http.Handler
	reg         *prometheus.Registry
	statMetrics *prometheus.GaugeVec
}

// NewPrometheusAgent creates and initializes a PrometheusAgent with
// pre-registered metrics based on the provided configuration.
func NewPrometheusAgent(cfg *config.CGRConfig, filters *engine.FilterS, cm *engine.ConnManager,
	shutdown *utils.SyncedChan) *PrometheusAgent {
	reg := prometheus.NewRegistry()
	statMetrics := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "cgrates",
			Subsystem: "stats",
			Name:      "metrics",
			Help:      "Current values for StatQueue metrics",
		}, []string{"tenant", "queue", "metric"})
	reg.MustRegister(statMetrics)
	if cfg.PrometheusAgentCfg().CollectGoMetrics {
		reg.MustRegister(collectors.NewGoCollector())
	}
	if cfg.PrometheusAgentCfg().CollectProcessMetrics {
		reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}

	handler := promhttp.InstrumentMetricHandler(
		reg,
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
	)

	return &PrometheusAgent{
		cfg:         cfg,
		filters:     filters,
		cm:          cm,
		shutdown:    shutdown,
		handler:     handler,
		reg:         reg,
		statMetrics: statMetrics,
	}
}

// updateStatsMetrics fetches and updates all StatQueue metrics by calling each
// configured StatS connection.
func (pa *PrometheusAgent) updateStatsMetrics() {
	if len(pa.cfg.PrometheusAgentCfg().StatQueueIDs) == 0 {
		return
	}
	for _, connID := range pa.cfg.PrometheusAgentCfg().StatSConns {
		for _, sqID := range pa.cfg.PrometheusAgentCfg().StatQueueIDs {

			tenantID := utils.NewTenantID(sqID)
			if tenantID.Tenant == "" {
				tenantID.Tenant = pa.cfg.GeneralCfg().DefaultTenant
			}

			var metrics map[string]float64
			err := pa.cm.Call(context.Background(), []string{connID},
				utils.StatSv1GetQueueFloatMetrics,
				&utils.TenantIDWithAPIOpts{
					TenantID: tenantID,
				}, &metrics)
			if err != nil && err.Error() != utils.ErrNotFound.Error() {
				utils.Logger.Err(fmt.Sprintf(
					"<%s> failed to retrieve metrics for StatQueue %q (connID=%q): %v",
					utils.PrometheusAgent, sqID, connID, err))
				continue
			}

			for metricID, val := range metrics {
				pa.statMetrics.WithLabelValues(tenantID.Tenant, tenantID.ID, metricID).Set(val)
			}
		}
	}
}

// ServeHTTP implements http.Handler interface. It updates all metrics on each
// scrape request before exposing them via the Prometheus HTTP handler.
func (pa *PrometheusAgent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pa.updateStatsMetrics()
	pa.handler.ServeHTTP(w, r)
}
