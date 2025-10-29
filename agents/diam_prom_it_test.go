//go:build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package agents

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestDiamPrometheus(t *testing.T) {
	t.Skip("test by looking at the log output")
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	ng := engine.TestEngine{
		ConfigJSON: `{
"apiers": {
	"enabled": true
},
"sessions": {
	"enabled": true,
	"chargers_conns": ["*internal"],
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"]
},
"chargers": {
	"enabled": true,
	"string_indexed_fields": ["*req.Account"]
},
"cdrs": {
	"enabled": true,
	"rals_conns": ["*internal"],
	"store_cdrs": false
},
"rals": {
	"enabled": true
},
"prometheus_agent": {
	"enabled": true,
	"stats_conns": ["*localhost"],
	"stat_queue_ids": ["SQ_1","SQ_2"]
},
"stats": {
	"enabled": true,
	"store_interval": "-1"
},
"thresholds": {
	"enabled": true,
	"store_interval": "-1"
},
"diameter_agent": {
	"enabled": true,
	"sessions_conns": ["*birpc_internal"],
	"stats_conns": ["*internal"],
	"thresholds_conns": ["*internal"],
	"request_processors": [{
		"id": "message",
		"filters": [
			"*string:~*vars.*cmd:CCR",
			"*prefix:~*req.Service-Context-Id:message",
			"*string:~*req.CC-Request-Type:4"
		],
		"flags": [
			"*message",
			"*accounts",
			"*cdrs",
			"*daStats:SQ_1&SQ_2",
			"*daThresholds:TH_1&TH_2",
		],
		"request_fields": [{
				"tag": "ToR",
				"path": "*cgreq.ToR",
				"type": "*constant",
				"value": "*sms"
			},
			{
				"tag": "OriginID",
				"path": "*cgreq.OriginID",
				"type": "*variable",
				"value": "~*req.Session-Id",
				"mandatory": true
			},
			{
				"tag": "Category",
				"path": "*cgreq.Category",
				"type": "*constant",
				"value": "sms"
			},
			{
				"tag": "RequestType",
				"path": "*cgreq.RequestType",
				"type": "*constant",
				"value": "*prepaid"
			},
			{
				"tag": "Account",
				"path": "*cgreq.Account",
				"type": "*variable",
				"mandatory": true,
				"value": "~*req.Subscription-Id.Subscription-Id-Data[~Subscription-Id-Type(0)]"
			},
			{
				"tag": "Destination",
				"path": "*cgreq.Destination",
				"type": "*variable",
				"mandatory": true,
				"value": "~*req.Service-Information.SMS-Information.Recipient-Address.Address-Data"
			},
			{
				"tag": "SetupTime",
				"path": "*cgreq.SetupTime",
				"type": "*variable",
				"value": "~*req.Event-Timestamp",
				"mandatory": true
			},
			{
				"tag": "AnswerTime",
				"path": "*cgreq.AnswerTime",
				"type": "*variable",
				"value": "~*req.Event-Timestamp",
				"mandatory": true
			},
			{
				"tag": "Usage",
				"path": "*cgreq.Usage",
				"type": "*variable",
				"value": "~*req.Requested-Service-Unit.CC-Time",
				"mandatory": true
			}
		],
		"reply_fields": [{
				"tag": "CCATemplate",
				"type": "*template",
				"value": "*cca"
			},
			{
				"tag": "ResultCode",
				"path": "*rep.Result-Code",
				"filters": ["*notempty:~*cgrep.Error:"],
				"type": "*constant",
				"value": "5030",
				"blocker": true
			}
		]
	}]
}
}`,
		TpFiles: map[string]string{
			// import Chargers via CSV to avoid cyclic imports (agents->v1->agents)
			utils.ChargersCsv: `
#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,*string:~*req.Account:1001,,*default,*none,10`,
		},
		DBCfg:            engine.InternalDBCfg,
		LogBuffer:        &bytes.Buffer{},
		GracefulShutdown: true,
	}
	t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, cfg := ng.Run(t)

	var reply string
	if err := client.Call(context.Background(), utils.APIerSv2SetBalance,
		utils.AttrSetBalance{
			Tenant:      "cgrates.org",
			Account:     "1001",
			Value:       100,
			BalanceType: utils.MetaSMS,
			Balance: map[string]any{
				utils.ID: "balance_sms",
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}

	if err := client.Call(context.Background(), utils.APIerSv2SetActions,
		utils.AttrSetActions{
			ActionsId: "ACT_LOG_WARNING",
			Actions: []*utils.TPAction{
				{
					Identifier: utils.MetaLog,
				},
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}

	if err := client.Call(context.Background(), utils.APIerSv1SetStatQueueProfile,
		engine.StatQueueProfileWithAPIOpts{
			StatQueueProfile: &engine.StatQueueProfile{
				Tenant:      "cgrates.org",
				ID:          "SQ_1",
				FilterIDs:   []string{"*string:~*opts.*eventType:ProcessTime"},
				QueueLength: -1,
				TTL:         5 * time.Second,
				Metrics: []*engine.MetricWithFilters{
					{
						MetricID: "*average#~*req.ProcessingTime",
					},
					{
						MetricID: "*sum#~*req.ProcessingTime",
					},
					{
						MetricID: "*highest#~*req.ProcessingTime",
					},
					{
						MetricID: "*lowest#~*req.ProcessingTime",
					},
					{
						MetricID: "*distinct#~*req.ProcessingTime",
					},
					{
						MetricID: utils.MetaREPSC,
					},
					{
						MetricID: utils.MetaREPFC,
					},
					{
						MetricID: utils.MetaREPFC + "#ERR_MESSAGE",
					},
				},
				Stored:   true,
				MinItems: 1,
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}

	if err := client.Call(context.Background(), utils.APIerSv1SetStatQueueProfile,
		engine.StatQueueProfileWithAPIOpts{
			StatQueueProfile: &engine.StatQueueProfile{
				Tenant:      "cgrates.org",
				ID:          "SQ_2",
				FilterIDs:   []string{"*string:~*opts.*eventType:ProcessTime"},
				QueueLength: -1,
				TTL:         10 * time.Second,
				Metrics: []*engine.MetricWithFilters{
					{
						MetricID: "*average#~*req.ProcessingTime",
					},
					{
						MetricID: "*sum#~*req.ProcessingTime",
					},
					{
						MetricID: "*highest#~*req.ProcessingTime",
					},
					{
						MetricID: "*lowest#~*req.ProcessingTime",
					},
					{
						MetricID: "*distinct#~*req.ProcessingTime",
					},
					{
						MetricID: utils.MetaREPSC,
					},
					{
						MetricID: utils.MetaREPFC,
					},
					{
						MetricID: utils.MetaREPFC + "#ERR_MESSAGE",
					},
				},
				Stored:   true,
				MinItems: 1,
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}

	if err := client.Call(context.Background(), utils.APIerSv1SetThresholdProfile,
		engine.ThresholdProfileWithAPIOpts{
			ThresholdProfile: &engine.ThresholdProfile{
				Tenant:    "cgrates.org",
				ID:        "TH_1",
				FilterIDs: []string{"*string:~*opts.*eventType:ProcessTime"},
				MaxHits:   -1,
				MinHits:   8,
				MinSleep:  time.Second,
				ActionIDs: []string{"ACT_LOG_WARNING"},
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}

	if err := client.Call(context.Background(), utils.APIerSv1SetThresholdProfile,
		engine.ThresholdProfileWithAPIOpts{
			ThresholdProfile: &engine.ThresholdProfile{
				Tenant:    "cgrates.org",
				ID:        "TH_2",
				FilterIDs: []string{"*string:~*opts.*eventType:ProcessTime"},
				MaxHits:   -1,
				MinHits:   10,
				MinSleep:  time.Second,
				ActionIDs: []string{"ACT_LOG_WARNING"},
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)
	diamClient, err := NewDiameterClient(cfg.DiameterAgentCfg().Listen, "localhost",
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().ListenNet)
	if err != nil {
		t.Fatal(err)
	}

	for range 10 {
		time.Sleep(time.Second)
		sendDiamCCR(t, diamClient, 5*time.Second, "2001")
		scrapePromURL(t)
	}
}

func scrapePromURL(t *testing.T) {
	t.Helper()
	url := "http://localhost:2080/prometheus"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	bodyString := string(body)
	fmt.Println(bodyString)
}
