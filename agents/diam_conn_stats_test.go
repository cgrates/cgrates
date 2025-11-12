//go:build integration

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
	"io"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestDiamConnStats(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	ng := engine.TestEngine{
		ConfigJSON: `{
"general": {
	"log_level": 7
},
"admins": {
	"enabled": true
},
// "prometheus_agent": {
// 	"enabled": true,
// 	"stats_conns": ["*internal"],
// 	"stat_queue_ids": [
// 		"SQ_CONN_1",
// 		"SQ_CONN_2",
// 		"SQ_CONN_3"
// 	]
// },
"stats": {
	"enabled": true,
	"store_interval": "-1",
	"string_indexed_fields": ["*req.OriginHost"]
},
"thresholds": {
	"enabled": true,
	"store_interval": "-1"
},
"sessions": {
	"enabled": true
},
"diameter_agent": {
	"enabled": true,
	"stats_conns": ["*localhost"],
	// "thresholds_conns": ["*localhost"],
	"conn_status_stat_queue_ids": ["SQ_CONN_1", "SQ_CONN_2", "SQ_CONN_3"],
	// "conn_status_threshold_ids": [],
	"conn_health_check_interval": "100ms"
}
}`,
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
		// LogBuffer:        &bytes.Buffer{},
		GracefulShutdown: true,
	}
	// defer fmt.Println(ng.LogBuffer)
	client, cfg := ng.Run(t)

	setSQProfile := func(id, originHost, originRealm string, ttl time.Duration) {
		t.Helper()
		var reply string
		if err := client.Call(context.Background(), utils.AdminSv1SetStatQueueProfile,
			engine.StatQueueProfileWithAPIOpts{
				StatQueueProfile: &engine.StatQueueProfile{
					Tenant: "cgrates.org",
					ID:     id,
					FilterIDs: []string{
						"*string:~*opts.*eventType:ConnectionStatusReport",
						fmt.Sprintf("*string:~*req.OriginHost:%s", originHost),
						fmt.Sprintf("*string:~*req.OriginRealm:%s", originRealm),
					},
					QueueLength: -1,
					TTL:         ttl,
					Metrics: []*engine.MetricWithFilters{
						{
							MetricID: "*sum#~*req.ConnectionStatus",
						},
					},
					Stored:   true,
					MinItems: 1,
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}
	}

	_ = func(id string) {
		t.Helper()
		var reply string
		if err := client.Call(context.Background(), utils.AdminSv1SetThresholdProfile,
			engine.ThresholdProfileWithAPIOpts{
				ThresholdProfile: &engine.ThresholdProfile{
					Tenant:    "cgrates.org",
					ID:        id,
					FilterIDs: []string{"*string:~*opts.*eventType:ConnectionStatusReport"},
					MaxHits:   -1,
					MinHits:   8,
					MinSleep:  time.Second,
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}
	}

	initDiamConn := func(originHost, originRealm string) io.Closer {
		t.Helper()
		client, err := NewDiameterClient(cfg.DiameterAgentCfg().Listen,
			originHost, originRealm, cfg.DiameterAgentCfg().VendorID,
			cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
			cfg.DiameterAgentCfg().DictionariesPath,
			cfg.DiameterAgentCfg().ListenNet)
		if err != nil {
			t.Fatal(err)
		}

		// TODO: Remove after updating go-diameter dependency.
		time.Sleep(10 * time.Millisecond)

		return client
	}

	checkConnStatusMetric := func(sqID string, want float64) {
		t.Helper()
		var metrics map[string]float64
		err := client.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     sqID,
				},
			}, &metrics)
		if err != nil {
			t.Error(err)
		}
		metricID := "*sum#~*req.ConnectionStatus"
		got, ok := metrics[metricID]
		if !ok {
			t.Errorf("could not find metric %q", metricID)
		}
		if got != want {
			t.Errorf("%q metric value = %.0f, want %.0f", metricID, got, want)
		}
	}

	setSQProfile("SQ_CONN_1", "host1", "realm1", -1)
	setSQProfile("SQ_CONN_2", "host2", "realm1", -1)
	setSQProfile("SQ_CONN_3", "host3", "realm2", -1)

	// no connections have been established yet, expect -1
	checkConnStatusMetric("SQ_CONN_1", -1)
	checkConnStatusMetric("SQ_CONN_2", -1)
	checkConnStatusMetric("SQ_CONN_3", -1)
	// scrapePromURL(t)

	// connections have been established, expect 1
	connHost1 := initDiamConn("host1", "realm1")
	connHost2 := initDiamConn("host2", "realm1")
	connHost3 := initDiamConn("host3", "realm2")
	checkConnStatusMetric("SQ_CONN_1", 1)
	checkConnStatusMetric("SQ_CONN_2", 1)
	checkConnStatusMetric("SQ_CONN_3", 1)
	// scrapePromURL(t)

	// connections have been closed, expect 0
	connHost1.Close()
	connHost2.Close()
	connHost3.Close()

	// Ensure periodic health check happens.
	time.Sleep(100 * time.Millisecond)

	checkConnStatusMetric("SQ_CONN_1", 0)
	checkConnStatusMetric("SQ_CONN_2", 0)
	checkConnStatusMetric("SQ_CONN_3", 0)
	// scrapePromURL(t)

	// restart connection from host1
	connHost1 = initDiamConn("host1", "realm1")
	checkConnStatusMetric("SQ_CONN_1", 1)
	checkConnStatusMetric("SQ_CONN_2", 0)
	checkConnStatusMetric("SQ_CONN_3", 0)
	t.Cleanup(func() { connHost1.Close() })
	// scrapePromURL(t)
}
