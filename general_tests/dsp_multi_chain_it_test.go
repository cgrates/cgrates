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

package general_tests

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var async = flag.Bool("async", false, "Run dispatcher sessions concurrently")

// TestDispatcherMultiChain tests concurrent sessions going through dispatchers.
// Tests that requests for the same account always get routed to the same RALs engine.
// Uses 2 HTTPAgents, 2 Dispatchers, 4 Sessions, 4 RALs, 1 shared engine.
func TestDispatcherMultiChain(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaMySQL:
	case utils.MetaInternal, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	// Shared engine: Chargers, CDRs, Apiers (port 6012/6013/6080)
	cfgShared := `{
"listen": {
	"rpc_json": "127.0.0.1:6012",
	"rpc_gob": "127.0.0.1:6013",
	"http": "127.0.0.1:6080"
},
"chargers": {
	"enabled": true
},
"cdrs": {
	"enabled": true,
	"chargers_conns": ["*internal"],
	"rals_conns": ["*internal"]
},
"rals": {
	"enabled": true
},
"apiers": {
	"enabled": true
}
}`

	// RALS1: port 5012/5013/5080
	cfgRALS1 := `{
"listen": {
	"rpc_json": "127.0.0.1:5012",
	"rpc_gob": "127.0.0.1:5013",
	"http": "127.0.0.1:5080"
},
"rals": {
	"enabled": true
}
}`

	// RALS2: port 5112/5113/5180
	cfgRALS2 := `{
"listen": {
	"rpc_json": "127.0.0.1:5112",
	"rpc_gob": "127.0.0.1:5113",
	"http": "127.0.0.1:5180"
},
"rals": {
	"enabled": true
}
}`

	// RALS3: port 5212/5213/5280
	cfgRALS3 := `{
"listen": {
	"rpc_json": "127.0.0.1:5212",
	"rpc_gob": "127.0.0.1:5213",
	"http": "127.0.0.1:5280"
},
"rals": {
	"enabled": true
}
}`

	// RALS4: port 5312/5313/5380
	cfgRALS4 := `{
"listen": {
	"rpc_json": "127.0.0.1:5312",
	"rpc_gob": "127.0.0.1:5313",
	"http": "127.0.0.1:5380"
},
"rals": {
	"enabled": true
}
}`

	// SM1: port 4012/4013/4080, routes RALs through DSP1
	cfgSM1 := `{
"listen": {
	"rpc_json": "127.0.0.1:4012",
	"rpc_gob": "127.0.0.1:4013",
	"http": "127.0.0.1:4080"
},
"rpc_conns": {
	"conn_dsp1": {
		"conns": [{"address": "127.0.0.1:3012", "transport": "*json"}]
	},
	"conn_shared": {
		"conns": [{"address": "127.0.0.1:6012", "transport": "*json"}]
	}
},
"sessions": {
	"enabled": true,
	"listen_bijson": "127.0.0.1:2014",
	"chargers_conns": ["conn_shared"],
	"rals_conns": ["conn_dsp1"],
	"cdrs_conns": ["conn_shared"]
}
}`

	// SM2: port 4112/4113/4180, routes RALs through DSP1
	cfgSM2 := `{
"listen": {
	"rpc_json": "127.0.0.1:4112",
	"rpc_gob": "127.0.0.1:4113",
	"http": "127.0.0.1:4180"
},
"rpc_conns": {
	"conn_dsp1": {
		"conns": [{"address": "127.0.0.1:3012", "transport": "*json"}]
	},
	"conn_shared": {
		"conns": [{"address": "127.0.0.1:6012", "transport": "*json"}]
	}
},
"sessions": {
	"enabled": true,
	"listen_bijson": "127.0.0.1:2114",
	"chargers_conns": ["conn_shared"],
	"rals_conns": ["conn_dsp1"],
	"cdrs_conns": ["conn_shared"]
}
}`

	// SM3: port 4212/4213/4280, routes RALs through DSP2
	cfgSM3 := `{
"listen": {
	"rpc_json": "127.0.0.1:4212",
	"rpc_gob": "127.0.0.1:4213",
	"http": "127.0.0.1:4280"
},
"rpc_conns": {
	"conn_dsp2": {
		"conns": [{"address": "127.0.0.1:3112", "transport": "*json"}]
	},
	"conn_shared": {
		"conns": [{"address": "127.0.0.1:6012", "transport": "*json"}]
	}
},
"sessions": {
	"enabled": true,
	"listen_bijson": "127.0.0.1:2214",
	"chargers_conns": ["conn_shared"],
	"rals_conns": ["conn_dsp2"],
	"cdrs_conns": ["conn_shared"]
}
}`

	// SM4: port 4312/4313/4380, routes RALs through DSP2
	cfgSM4 := `{
"listen": {
	"rpc_json": "127.0.0.1:4312",
	"rpc_gob": "127.0.0.1:4313",
	"http": "127.0.0.1:4380"
},
"rpc_conns": {
	"conn_dsp2": {
		"conns": [{"address": "127.0.0.1:3112", "transport": "*json"}]
	},
	"conn_shared": {
		"conns": [{"address": "127.0.0.1:6012", "transport": "*json"}]
	}
},
"sessions": {
	"enabled": true,
	"listen_bijson": "127.0.0.1:2314",
	"chargers_conns": ["conn_shared"],
	"rals_conns": ["conn_dsp2"],
	"cdrs_conns": ["conn_shared"]
}
}`

	// DSP1: port 3012/3013/3080, routes sessions to SM1/SM2, RALs to RALS1/RALS2
	cfgDSP1 := `{
"listen": {
	"rpc_json": "127.0.0.1:3012",
	"rpc_gob": "127.0.0.1:3013",
	"http": "127.0.0.1:3080"
},
"dispatchers": {
	"enabled": true,
	"prefix_indexed_fields": ["*req.Account"]
}
}`

	// DSP2: port 3112/3113/3180, routes sessions to SM3/SM4, RALs to RALS3/RALS4
	cfgDSP2 := `{
"listen": {
	"rpc_json": "127.0.0.1:3112",
	"rpc_gob": "127.0.0.1:3113",
	"http": "127.0.0.1:3180"
},
"dispatchers": {
	"enabled": true,
	"prefix_indexed_fields": ["*req.Account"]
}
}`

	// HA1: HTTPAgent on port 2012/2013/2080, sends to DSP1
	cfgHA1 := `{
"listen": {
	"rpc_json": "127.0.0.1:2012",
	"rpc_gob": "127.0.0.1:2013",
	"http": "127.0.0.1:2080"
},
"rpc_conns": {
	"conn_dsp1": {
		"conns": [{"address": "127.0.0.1:3012", "transport": "*json"}]
	}
},
"http_agent": [
	{
		"id": "HTTPAgent_HA1",
		"url": "/sessions",
		"sessions_conns": ["conn_dsp1"],
		"request_payload": "*url",
		"reply_payload": "*xml",
		"request_processors": [
			{
				"id": "auth",
				"filters": ["*string:~*req.request_type:auth"],
				"flags": ["*authorize", "*accounts"],
				"request_fields": [
					{"tag": "ToR", "path": "*cgreq.ToR", "type": "*constant", "value": "*voice"},
					{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.session_id"},
					{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*constant", "value": "*prepaid"},
					{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*constant", "value": "cgrates.org"},
					{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", "value": "call"},
					{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account"},
					{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", "value": "~*req.account"},
					{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.destination"},
					{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*constant", "value": "*now"},
					{"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable", "value": "~*req.usage"},
					{"tag": "RouteID", "path": "*opts.*routeID", "type": "*variable", "value": "~*req.account"}
				],
				"reply_fields": [
					{"tag": "MaxUsage", "path": "*rep.MaxUsage", "type": "*variable", "value": "~*cgrep.MaxUsage{*duration_seconds}"}
				]
			},
			{
				"id": "init",
				"filters": ["*string:~*req.request_type:init"],
				"flags": ["*initiate", "*accounts"],
				"request_fields": [
					{"tag": "ToR", "path": "*cgreq.ToR", "type": "*constant", "value": "*voice"},
					{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.session_id"},
					{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*constant", "value": "*prepaid"},
					{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*constant", "value": "cgrates.org"},
					{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", "value": "call"},
					{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account"},
					{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", "value": "~*req.account"},
					{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.destination"},
					{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*constant", "value": "*now"},
					{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*constant", "value": "*now"},
					{"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable", "value": "~*req.usage"},
					{"tag": "RouteID", "path": "*opts.*routeID", "type": "*variable", "value": "~*req.account"}
				],
				"reply_fields": [
					{"tag": "MaxUsage", "path": "*rep.MaxUsage", "type": "*variable", "value": "~*cgrep.MaxUsage{*duration_seconds}"}
				]
			},
			{
				"id": "update",
				"filters": ["*string:~*req.request_type:update"],
				"flags": ["*update", "*accounts"],
				"request_fields": [
					{"tag": "ToR", "path": "*cgreq.ToR", "type": "*constant", "value": "*voice"},
					{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.session_id"},
					{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*constant", "value": "*prepaid"},
					{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*constant", "value": "cgrates.org"},
					{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", "value": "call"},
					{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account"},
					{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", "value": "~*req.account"},
					{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.destination"},
					{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*constant", "value": "*now"},
					{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*constant", "value": "*now"},
					{"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable", "value": "~*req.usage"},
					{"tag": "RouteID", "path": "*opts.*routeID", "type": "*variable", "value": "~*req.account"}
				],
				"reply_fields": [
					{"tag": "MaxUsage", "path": "*rep.MaxUsage", "type": "*variable", "value": "~*cgrep.MaxUsage{*duration_seconds}"}
				]
			},
			{
				"id": "terminate",
				"filters": ["*string:~*req.request_type:terminate"],
				"flags": ["*terminate", "*accounts", "*cdrs"],
				"request_fields": [
					{"tag": "ToR", "path": "*cgreq.ToR", "type": "*constant", "value": "*voice"},
					{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.session_id"},
					{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*constant", "value": "*prepaid"},
					{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*constant", "value": "cgrates.org"},
					{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", "value": "call"},
					{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account"},
					{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", "value": "~*req.account"},
					{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.destination"},
					{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*constant", "value": "*now"},
					{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*constant", "value": "*now"},
					{"tag": "RouteID", "path": "*opts.*routeID", "type": "*variable", "value": "~*req.account"}
				],
				"reply_fields": [
					{"tag": "Status", "path": "*rep.Status", "type": "*constant", "value": "OK"}
				]
			}
		]
	}
]
}`

	// HA2: HTTPAgent on port 2112/2113/2180, sends to DSP2
	cfgHA2 := `{
"listen": {
	"rpc_json": "127.0.0.1:2112",
	"rpc_gob": "127.0.0.1:2113",
	"http": "127.0.0.1:2180"
},
"rpc_conns": {
	"conn_dsp2": {
		"conns": [{"address": "127.0.0.1:3112", "transport": "*json"}]
	}
},
"http_agent": [
	{
		"id": "HTTPAgent_HA2",
		"url": "/sessions",
		"sessions_conns": ["conn_dsp2"],
		"request_payload": "*url",
		"reply_payload": "*xml",
		"request_processors": [
			{
				"id": "auth",
				"filters": ["*string:~*req.request_type:auth"],
				"flags": ["*authorize", "*accounts"],
				"request_fields": [
					{"tag": "ToR", "path": "*cgreq.ToR", "type": "*constant", "value": "*voice"},
					{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.session_id"},
					{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*constant", "value": "*prepaid"},
					{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*constant", "value": "cgrates.org"},
					{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", "value": "call"},
					{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account"},
					{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", "value": "~*req.account"},
					{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.destination"},
					{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*constant", "value": "*now"},
					{"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable", "value": "~*req.usage"},
					{"tag": "RouteID", "path": "*opts.*routeID", "type": "*variable", "value": "~*req.account"}
				],
				"reply_fields": [
					{"tag": "MaxUsage", "path": "*rep.MaxUsage", "type": "*variable", "value": "~*cgrep.MaxUsage{*duration_seconds}"}
				]
			},
			{
				"id": "init",
				"filters": ["*string:~*req.request_type:init"],
				"flags": ["*initiate", "*accounts"],
				"request_fields": [
					{"tag": "ToR", "path": "*cgreq.ToR", "type": "*constant", "value": "*voice"},
					{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.session_id"},
					{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*constant", "value": "*prepaid"},
					{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*constant", "value": "cgrates.org"},
					{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", "value": "call"},
					{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account"},
					{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", "value": "~*req.account"},
					{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.destination"},
					{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*constant", "value": "*now"},
					{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*constant", "value": "*now"},
					{"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable", "value": "~*req.usage"},
					{"tag": "RouteID", "path": "*opts.*routeID", "type": "*variable", "value": "~*req.account"}
				],
				"reply_fields": [
					{"tag": "MaxUsage", "path": "*rep.MaxUsage", "type": "*variable", "value": "~*cgrep.MaxUsage{*duration_seconds}"}
				]
			},
			{
				"id": "update",
				"filters": ["*string:~*req.request_type:update"],
				"flags": ["*update", "*accounts"],
				"request_fields": [
					{"tag": "ToR", "path": "*cgreq.ToR", "type": "*constant", "value": "*voice"},
					{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.session_id"},
					{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*constant", "value": "*prepaid"},
					{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*constant", "value": "cgrates.org"},
					{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", "value": "call"},
					{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account"},
					{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", "value": "~*req.account"},
					{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.destination"},
					{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*constant", "value": "*now"},
					{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*constant", "value": "*now"},
					{"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable", "value": "~*req.usage"},
					{"tag": "RouteID", "path": "*opts.*routeID", "type": "*variable", "value": "~*req.account"}
				],
				"reply_fields": [
					{"tag": "MaxUsage", "path": "*rep.MaxUsage", "type": "*variable", "value": "~*cgrep.MaxUsage{*duration_seconds}"}
				]
			},
			{
				"id": "terminate",
				"filters": ["*string:~*req.request_type:terminate"],
				"flags": ["*terminate", "*accounts", "*cdrs"],
				"request_fields": [
					{"tag": "ToR", "path": "*cgreq.ToR", "type": "*constant", "value": "*voice"},
					{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.session_id"},
					{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*constant", "value": "*prepaid"},
					{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*constant", "value": "cgrates.org"},
					{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", "value": "call"},
					{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account"},
					{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", "value": "~*req.account"},
					{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.destination"},
					{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*constant", "value": "*now"},
					{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*constant", "value": "*now"},
					{"tag": "RouteID", "path": "*opts.*routeID", "type": "*variable", "value": "~*req.account"}
				],
				"reply_fields": [
					{"tag": "Status", "path": "*rep.Status", "type": "*constant", "value": "OK"}
				]
			}
		]
	}
]
}`

	// Tariff plan files with dispatcher profiles for both DSP1 and DSP2
	tpFiles := map[string]string{
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,0,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,
cgrates.org,call,1002,2014-01-14T00:00:00Z,RP_ANY,
cgrates.org,call,1003,2014-01-14T00:00:00Z,RP_ANY,
cgrates.org,call,1004,2014-01-14T00:00:00Z,RP_ANY,
cgrates.org,call,1005,2014-01-14T00:00:00Z,RP_ANY,
cgrates.org,call,2001,2014-01-14T00:00:00Z,RP_ANY,
cgrates.org,call,2002,2014-01-14T00:00:00Z,RP_ANY,
cgrates.org,call,2003,2014-01-14T00:00:00Z,RP_ANY,
cgrates.org,call,2004,2014-01-14T00:00:00Z,RP_ANY,
cgrates.org,call,2005,2014-01-14T00:00:00Z,RP_ANY,`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
		utils.FiltersCsv: `#Tenant,ID,Type,Path,Values,ActivationInterval
cgrates.org,FLTR_ACNT_10xx,*prefix,~*req.Account,10,
cgrates.org,FLTR_ACNT_20xx,*prefix,~*req.Account,20,`,
		utils.DispatcherHostsCsv: `#Tenant[0],ID[1],Address[2],Transport[3],ConnectAttempts[4],Reconnects[5],MaxReconnectInterval[6],ConnectTimeout[7],ReplyTimeout[8],Tls[9],ClientKey[10],ClientCertificate[11],CaCertificate[12]
cgrates.org,SM1,127.0.0.1:4012,*json,1,1,,2s,2s,,,,
cgrates.org,SM2,127.0.0.1:4112,*json,1,1,,2s,2s,,,,
cgrates.org,SM3,127.0.0.1:4212,*json,1,1,,2s,2s,,,,
cgrates.org,SM4,127.0.0.1:4312,*json,1,1,,2s,2s,,,,
cgrates.org,RALS1,127.0.0.1:5012,*json,1,1,,2s,2s,,,,
cgrates.org,RALS2,127.0.0.1:5112,*json,1,1,,2s,2s,,,,
cgrates.org,RALS3,127.0.0.1:5212,*json,1,1,,2s,2s,,,,
cgrates.org,RALS4,127.0.0.1:5312,*json,1,1,,2s,2s,,,,`,
		utils.DispatcherProfilesCsv: `#Tenant,ID,Subsystems,FilterIDs,ActivationInterval,Strategy,StrategyParameters,ConnID,ConnFilterIDs,ConnWeight,ConnBlocker,ConnParameters,Weight
cgrates.org,DSP1_SM,*sessions,FLTR_ACNT_10xx,,*round_robin,,SM1,,10,false,,10
cgrates.org,DSP1_SM,,,,,,SM2,,10,,,
cgrates.org,DSP1_RALS,*responder,FLTR_ACNT_10xx,,*round_robin,,RALS1,,10,false,,10
cgrates.org,DSP1_RALS,,,,,,RALS2,,10,,,
cgrates.org,DSP2_SM,*sessions,FLTR_ACNT_20xx,,*round_robin,,SM3,,10,false,,10
cgrates.org,DSP2_SM,,,,,,SM4,,10,,,
cgrates.org,DSP2_RALS,*responder,FLTR_ACNT_20xx,,*round_robin,,RALS3,,10,false,,10
cgrates.org,DSP2_RALS,,,,,,RALS4,,10,,,`,
	}

	// Start engines in correct order:
	// 1. Shared engine first (load tariff plan through it)
	// 2. RALS1-4
	// 3. DSP1-2 (must be up before SM engines since SM connects to DSP for rals_conns)
	// 4. SM1-4
	// 5. HA1-2

	ngShared := engine.TestEngine{
		ConfigJSON: cfgShared,
		TpFiles:    tpFiles,
	}
	clientShared, _ := ngShared.Run(t)

	ngRALS1 := engine.TestEngine{
		ConfigJSON:     cfgRALS1,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngRALS1.Run(t)

	ngRALS2 := engine.TestEngine{
		ConfigJSON:     cfgRALS2,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngRALS2.Run(t)

	ngRALS3 := engine.TestEngine{
		ConfigJSON:     cfgRALS3,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngRALS3.Run(t)

	ngRALS4 := engine.TestEngine{
		ConfigJSON:     cfgRALS4,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngRALS4.Run(t)

	ngDSP1 := engine.TestEngine{
		ConfigJSON:     cfgDSP1,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngDSP1.Run(t)

	ngDSP2 := engine.TestEngine{
		ConfigJSON:     cfgDSP2,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngDSP2.Run(t)

	ngSM1 := engine.TestEngine{
		ConfigJSON:     cfgSM1,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngSM1.Run(t)

	ngSM2 := engine.TestEngine{
		ConfigJSON:     cfgSM2,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngSM2.Run(t)

	ngSM3 := engine.TestEngine{
		ConfigJSON:     cfgSM3,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngSM3.Run(t)

	ngSM4 := engine.TestEngine{
		ConfigJSON:     cfgSM4,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngSM4.Run(t)

	ngHA1 := engine.TestEngine{
		ConfigJSON:     cfgHA1,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngHA1.Run(t)

	ngHA2 := engine.TestEngine{
		ConfigJSON:     cfgHA2,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngHA2.Run(t)

	httpC := &http.Client{}
	var sessionNo atomic.Int64

	sendRequest := func(t *testing.T, port int, reqType, sessionID, account, destination, usage string) string {
		t.Helper()
		resp, err := httpC.PostForm(fmt.Sprintf("http://127.0.0.1:%d/sessions", port),
			url.Values{
				"request_type": {reqType},
				"session_id":   {sessionID},
				"account":      {account},
				"destination":  {destination},
				"usage":        {usage},
			})
		if err != nil {
			t.Fatalf("HTTP request failed: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("HTTP request returned %d: %s", resp.StatusCode, string(body))
		}
		return string(body)
	}

	runHTTPSession := func(t *testing.T, port int, account, destination string, initUsage time.Duration, updateUsages ...time.Duration) {
		t.Helper()
		sessionID := fmt.Sprintf("session_%s_%d", account, sessionNo.Add(1))

		sendRequest(t, port, "auth", sessionID, account, destination, initUsage.String())
		sendRequest(t, port, "init", sessionID, account, destination, initUsage.String())

		for _, u := range updateUsages {
			sendRequest(t, port, "update", sessionID, account, destination, u.String())
		}

		sendRequest(t, port, "terminate", sessionID, account, destination, "")
	}

	setBalance := func(t *testing.T, acc string, value float64) {
		t.Helper()
		var reply string
		if err := clientShared.Call(context.Background(), utils.APIerSv2SetBalance,
			utils.AttrSetBalance{
				Tenant:      "cgrates.org",
				Account:     acc,
				Value:       value,
				BalanceType: utils.MetaMonetary,
				Balance:     map[string]any{utils.ID: "test"},
			}, &reply); err != nil {
			t.Fatal(err)
		}
	}

	checkBalance := func(t *testing.T, acc string, want float64) {
		t.Helper()
		var acnt engine.Account
		if err := clientShared.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{
				Tenant:  "cgrates.org",
				Account: acc,
			}, &acnt); err != nil {
			t.Fatalf("GetAccount failed: %v", err)
		}
		if bal := acnt.BalanceMap[utils.MetaMonetary][0]; bal == nil {
			t.Errorf("balance not found for account %q", acc)
		} else if bal.Value != want {
			t.Errorf("account %q balance = %v, want %v", acc, bal.Value, want)
		}
	}

	// Set balances for all 10 accounts
	setBalance(t, "1001", 100)
	setBalance(t, "1002", 100)
	setBalance(t, "1003", 100)
	setBalance(t, "1004", 100)
	setBalance(t, "1005", 100)
	setBalance(t, "2001", 100)
	setBalance(t, "2002", 100)
	setBalance(t, "2003", 100)
	setBalance(t, "2004", 100)
	setBalance(t, "2005", 100)

	type sessionParams struct {
		port        int
		account     string
		destination string
		initUsage   time.Duration
		updates     []time.Duration
	}

	// 2 sessions per account, all run concurrently when -async flag is set.
	// 18s + 10s = 28s total per account, expected balance: 100 - 28 = 72
	sessions := []sessionParams{
		{2080, "1001", "1099", 10 * time.Second, []time.Duration{5 * time.Second, 3 * time.Second}},
		{2080, "1001", "1099", 5 * time.Second, []time.Duration{3 * time.Second, 2 * time.Second}},
		{2080, "1002", "1099", 10 * time.Second, []time.Duration{5 * time.Second, 3 * time.Second}},
		{2080, "1002", "1099", 5 * time.Second, []time.Duration{3 * time.Second, 2 * time.Second}},
		{2080, "1003", "1099", 10 * time.Second, []time.Duration{5 * time.Second, 3 * time.Second}},
		{2080, "1003", "1099", 5 * time.Second, []time.Duration{3 * time.Second, 2 * time.Second}},
		{2080, "1004", "1099", 10 * time.Second, []time.Duration{5 * time.Second, 3 * time.Second}},
		{2080, "1004", "1099", 5 * time.Second, []time.Duration{3 * time.Second, 2 * time.Second}},
		{2080, "1005", "1099", 10 * time.Second, []time.Duration{5 * time.Second, 3 * time.Second}},
		{2080, "1005", "1099", 5 * time.Second, []time.Duration{3 * time.Second, 2 * time.Second}},
		{2180, "2001", "2099", 10 * time.Second, []time.Duration{5 * time.Second, 3 * time.Second}},
		{2180, "2001", "2099", 5 * time.Second, []time.Duration{3 * time.Second, 2 * time.Second}},
		{2180, "2002", "2099", 10 * time.Second, []time.Duration{5 * time.Second, 3 * time.Second}},
		{2180, "2002", "2099", 5 * time.Second, []time.Duration{3 * time.Second, 2 * time.Second}},
		{2180, "2003", "2099", 10 * time.Second, []time.Duration{5 * time.Second, 3 * time.Second}},
		{2180, "2003", "2099", 5 * time.Second, []time.Duration{3 * time.Second, 2 * time.Second}},
		{2180, "2004", "2099", 10 * time.Second, []time.Duration{5 * time.Second, 3 * time.Second}},
		{2180, "2004", "2099", 5 * time.Second, []time.Duration{3 * time.Second, 2 * time.Second}},
		{2180, "2005", "2099", 10 * time.Second, []time.Duration{5 * time.Second, 3 * time.Second}},
		{2180, "2005", "2099", 5 * time.Second, []time.Duration{3 * time.Second, 2 * time.Second}},
	}

	if *async {
		var wg sync.WaitGroup
		for _, s := range sessions {
			wg.Add(1)
			go func(sp sessionParams) {
				defer wg.Done()
				runHTTPSession(t, sp.port, sp.account, sp.destination, sp.initUsage, sp.updates...)
			}(s)
		}
		wg.Wait()
	} else {
		for _, s := range sessions {
			runHTTPSession(t, s.port, s.account, s.destination, s.initUsage, s.updates...)
		}
	}

	checkBalance(t, "1001", 72)
	checkBalance(t, "1002", 72)
	checkBalance(t, "1003", 72)
	checkBalance(t, "1004", 72)
	checkBalance(t, "1005", 72)
	checkBalance(t, "2001", 72)
	checkBalance(t, "2002", 72)
	checkBalance(t, "2003", 72)
	checkBalance(t, "2004", 72)
	checkBalance(t, "2005", 72)
}
