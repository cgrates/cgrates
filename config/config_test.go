/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PU:3474RPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package config

import (
	"encoding/json"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/rpcclient"
	"github.com/google/go-cmp/cmp"

	"github.com/cgrates/cgrates/utils"
)

func TestNewDefaultConfigError(t *testing.T) {
	if _, err := newCGRConfig([]byte(CGRATES_CFG_JSON)); err != nil {
		t.Error(err)
	}
}

func TestNewCgrConfigFromBytesError(t *testing.T) {
	cfg := []byte(`{
"cores": {
	"caps": "0",
}
}`)
	expected := "json: cannot unmarshal string into Go struct field CoreSJsonCfg.Caps of type int"
	if _, err := newCGRConfig(cfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v,\n received %+v", expected, err)
	}
}

func TestNewCgrConfigFromBytesDecodeError(t *testing.T) {
	cfg := []byte(`invalidSection`)
	expected := "invalid character 'i' looking for beginning of value around line 1 and position 1\n line: \"invalidSection\""
	if _, err := newCGRConfig(cfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v,\n received %+q", expected, err)
	}
}

func TestCgrCfgConfigSharing(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	SetCgrConfig(cfg)
	cfgReturn := CgrConfig()
	if !reflect.DeepEqual(cfgReturn, cfg) {
		t.Errorf("Retrieved %v, Expected %v", cfgReturn, cfg)
	}
}

func TestCgrCfgLoadWithDefaults(t *testing.T) {
	jsnCfg := `
{
"freeswitch_agent": {
	"enabled": true,				// starts SessionManager service: <true|false>
	"event_socket_conns":[					// instantiate connections to multiple FreeSWITCH servers
		{"address": "1.2.3.4:8021", "password": "ClueCon", "reconnects": 3, "alias":"123"},
	{"address": "1.2.3.5:8021", "password": "ClueCon", "reconnects": 5, "reply_timeout": "30s", "alias":"124"}
	],
},

}`
	eCgrCfg := NewDefaultCGRConfig()
	eCgrCfg.fsAgentCfg.Enabled = true
	eCgrCfg.fsAgentCfg.EventSocketConns = []*FsConnCfg{
		{Address: "1.2.3.4:8021", Password: "ClueCon", Reconnects: 3, ReplyTimeout: time.Minute, Alias: "123"},
		{Address: "1.2.3.5:8021", Password: "ClueCon", Reconnects: 5, ReplyTimeout: 30 * time.Second, Alias: "124"},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrCfg.fsAgentCfg, cgrCfg.fsAgentCfg) {
		t.Errorf("Expected: %+v, received: %+v", eCgrCfg.fsAgentCfg, cgrCfg.fsAgentCfg)
	}
}

func TestCgrCfgDataDBPortWithoutDynamic(t *testing.T) {
	jsnCfg := `
{
"db": {
	"db_conns": {
		"*default": {	
			"db_type": "mongo",
			},
		},
	}
}`

	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if cgrCfg.DbCfg().DBConns[utils.MetaDefault].Type != utils.MetaMongo {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DbCfg().DBConns[utils.MetaDefault].Type, utils.MetaMongo)
	} else if cgrCfg.DbCfg().DBConns[utils.MetaDefault].Port != "0" {
		t.Errorf("Expected: %+v, received: %+v", "0", cgrCfg.DbCfg().DBConns[utils.MetaDefault].Port)
	}
	jsnCfg = `
{
"db": {
	"db_conns": {
		"*default": {	
			"db_type": "internal",
			},
		},
	}
}`

	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if cgrCfg.DbCfg().DBConns[utils.MetaDefault].Type != utils.MetaInternal {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DbCfg().DBConns[utils.MetaDefault].Type, utils.MetaInternal)
	} else if cgrCfg.DbCfg().DBConns[utils.MetaDefault].Port != "0" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DbCfg().DBConns[utils.MetaDefault].Port, "0")
	}
}

func TestCgrCfgDataDBPortWithDymanic(t *testing.T) {
	jsnCfg := `
{
"db": {
	"db_conns": {
		"*default": {	
			"db_type": "mongo",
			"db_port": -1,
			},
		},
	}
}`

	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if cgrCfg.DbCfg().DBConns[utils.MetaDefault].Type != utils.MetaMongo {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DbCfg().DBConns[utils.MetaDefault].Type, utils.MetaMongo)
	} else if cgrCfg.DbCfg().DBConns[utils.MetaDefault].Port != "27017" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DbCfg().DBConns[utils.MetaDefault].Port, "27017")
	}
	jsnCfg = `
{
"db": {
	"db_conns": {
		"*default": {	
			"db_type": "internal",
			"db_port": -1,
			},
		},
	}
}`

	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if cgrCfg.DbCfg().DBConns[utils.MetaDefault].Type != utils.MetaInternal {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DbCfg().DBConns[utils.MetaDefault].Type, utils.MetaInternal)
	} else if cgrCfg.DbCfg().DBConns[utils.MetaDefault].Port != "internal" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DbCfg().DBConns[utils.MetaDefault].Port, "internal")
	}
}

func TestCgrCfgListener(t *testing.T) {
	jsnCfg := `
{
"listen": {
	"rpc_json": ":2012",
	"rpc_gob": ":2013",
	"http": ":2080",
	}
}`

	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if cgrCfg.ListenCfg().RPCGOBTLSListen != "127.0.0.1:2023" {
		t.Errorf("Expected: 127.0.0.1:2023 , received: %+v", cgrCfg.ListenCfg().RPCGOBTLSListen)
	} else if cgrCfg.ListenCfg().RPCJSONTLSListen != "127.0.0.1:2022" {
		t.Errorf("Expected: 127.0.0.1:2022 , received: %+v", cgrCfg.ListenCfg().RPCJSONTLSListen)
	}
}

func TestHttpAgentCfg(t *testing.T) {
	jsnCfg := `
{
"http_agent": [
	{
		"id": "conecto1",
		"url": "/conecto",					// relative URL for requests coming in
		"sessions_conns": ["*internal"],
		"request_payload":	"*url",			// source of input data <*url>
		"reply_payload":	"*xml",			// type of output data <*xml>
		"request_processors": [],
	}
],
}
	`
	eCgrCfg := NewDefaultCGRConfig()
	eCgrCfg.httpAgentCfg = []*HTTPAgentCfg{
		{
			ID:                "conecto1",
			URL:               "/conecto",
			RequestPayload:    utils.MetaUrl,
			ReplyPayload:      utils.MetaXml,
			SessionSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
			RequestProcessors: nil,
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrCfg.HTTPAgentCfg(), cgrCfg.HTTPAgentCfg()) {
		t.Errorf("Expected: %s, received: %s",
			utils.ToJSON(eCgrCfg.httpAgentCfg), utils.ToJSON(cgrCfg.httpAgentCfg))
	}
}

func TestCgrCfgJSONDefaultsGeneral(t *testing.T) {
	if cgrCfg.GeneralCfg().RoundingDecimals != 5 {
		t.Errorf("Expected: 5, received: %+v", cgrCfg.GeneralCfg().RoundingDecimals)
	}
	if cgrCfg.GeneralCfg().DBDataEncoding != "msgpack" {
		t.Errorf("Expected: msgpack, received: %+v", cgrCfg.GeneralCfg().DBDataEncoding)
	}
	if expected := "/var/spool/cgrates/tpe"; cgrCfg.GeneralCfg().TpExportPath != expected {
		t.Errorf("Expected: %+v, received: %+v", expected, cgrCfg.GeneralCfg().TpExportPath)
	}
	if cgrCfg.GeneralCfg().DefaultReqType != "*rated" {
		t.Errorf("Expected: *rated, received: %+v", cgrCfg.GeneralCfg().DefaultReqType)
	}
	if cgrCfg.GeneralCfg().DefaultCategory != "call" {
		t.Errorf("Expected: call, received: %+v", cgrCfg.GeneralCfg().DefaultCategory)
	}
	if cgrCfg.GeneralCfg().DefaultTenant != "cgrates.org" {
		t.Errorf("Expected: cgrates.org, received: %+v", cgrCfg.GeneralCfg().DefaultTenant)
	}
	if cgrCfg.GeneralCfg().DefaultCaching != utils.MetaReload {
		t.Errorf("Expected: *reload, received: %+v", cgrCfg.GeneralCfg().DefaultCaching)
	}
	if cgrCfg.GeneralCfg().DefaultTimezone != "Local" {
		t.Errorf("Expected: Local, received: %+v", cgrCfg.GeneralCfg().DefaultTimezone)
	}
	if cgrCfg.GeneralCfg().ConnectAttempts != 5 {
		t.Errorf("Expected: 3, received: %+v", cgrCfg.GeneralCfg().ConnectAttempts)
	}
	if cgrCfg.GeneralCfg().Reconnects != -1 {
		t.Errorf("Expected: -1, received: %+v", cgrCfg.GeneralCfg().Reconnects)
	}
	if cgrCfg.GeneralCfg().ConnectTimeout != time.Second {
		t.Errorf("Expected: 1s, received: %+v", cgrCfg.GeneralCfg().ConnectTimeout)
	}
	if cgrCfg.GeneralCfg().ReplyTimeout != 2*time.Second {
		t.Errorf("Expected: 2s, received: %+v", cgrCfg.GeneralCfg().ReplyTimeout)
	}
	if cgrCfg.GeneralCfg().LockingTimeout != 0 {
		t.Errorf("Expected: 0, received: %+v", cgrCfg.GeneralCfg().LockingTimeout)
	}
	if cgrCfg.GeneralCfg().DigestSeparator != "," {
		t.Errorf("Expected: utils.CSVSep , received: %+v", cgrCfg.GeneralCfg().DigestSeparator)
	}
	if cgrCfg.GeneralCfg().DigestEqual != ":" {
		t.Errorf("Expected: ':' , received: %+v", cgrCfg.GeneralCfg().DigestEqual)
	}
}

func TestCgrCfgJSONDefaultsListen(t *testing.T) {
	if cgrCfg.ListenCfg().RPCJSONListen != "127.0.0.1:2012" {
		t.Errorf("Expected: 127.0.0.1:2012 , received: %+v", cgrCfg.ListenCfg().RPCJSONListen)
	}
	if cgrCfg.ListenCfg().RPCGOBListen != "127.0.0.1:2013" {
		t.Errorf("Expected: 127.0.0.1:2013 , received: %+v", cgrCfg.ListenCfg().RPCGOBListen)
	}
	if cgrCfg.ListenCfg().HTTPListen != "127.0.0.1:2080" {
		t.Errorf("Expected: 127.0.0.1:2080 , received: %+v", cgrCfg.ListenCfg().HTTPListen)
	}
	if cgrCfg.ListenCfg().RPCJSONTLSListen != "127.0.0.1:2022" {
		t.Errorf("Expected: 127.0.0.1:2022 , received: %+v", cgrCfg.ListenCfg().RPCJSONListen)
	}
	if cgrCfg.ListenCfg().RPCGOBTLSListen != "127.0.0.1:2023" {
		t.Errorf("Expected: 127.0.0.1:2023 , received: %+v", cgrCfg.ListenCfg().RPCGOBListen)
	}
	if cgrCfg.ListenCfg().HTTPTLSListen != "127.0.0.1:2280" {
		t.Errorf("Expected: 127.0.0.1:2280 , received: %+v", cgrCfg.ListenCfg().HTTPListen)
	}
}

func TestCgrCfgJSONDefaultsjsnDataDb(t *testing.T) {
	if cgrCfg.DbCfg().DBConns[utils.MetaDefault].Type != utils.MetaInternal {
		t.Errorf("Expecting: internal , received: %+v", cgrCfg.DbCfg().DBConns[utils.MetaDefault].Type)
	}
	if cgrCfg.DbCfg().DBConns[utils.MetaDefault].Host != "" {
		t.Errorf("Expecting: \"\" , received: %+v", cgrCfg.DbCfg().DBConns[utils.MetaDefault].Host)
	}
	if cgrCfg.DbCfg().DBConns[utils.MetaDefault].Port != "0" {
		t.Errorf("Expecting: 0 , received: %+v", cgrCfg.DbCfg().DBConns[utils.MetaDefault].Port)
	}
	if cgrCfg.DbCfg().DBConns[utils.MetaDefault].Name != "" {
		t.Errorf("Expecting: \"\"  , received: %+v", cgrCfg.DbCfg().DBConns[utils.MetaDefault].Name)
	}
	if cgrCfg.DbCfg().DBConns[utils.MetaDefault].User != "" {
		t.Errorf("Expecting: \"\" , received: %+v", cgrCfg.DbCfg().DBConns[utils.MetaDefault].User)
	}
	if cgrCfg.DbCfg().DBConns[utils.MetaDefault].Password != "" {
		t.Errorf("Expecting: \"\" , received: %+v", cgrCfg.DbCfg().DBConns[utils.MetaDefault].Password)
	}
	if len(cgrCfg.DbCfg().DBConns[utils.MetaDefault].RmtConns) != 0 {
		t.Errorf("Expecting:  0, received: %+v", len(cgrCfg.DbCfg().DBConns[utils.MetaDefault].RmtConns))
	}
	if len(cgrCfg.DbCfg().DBConns[utils.MetaDefault].RplConns) != 0 {
		t.Errorf("Expecting:  0, received: %+v", len(cgrCfg.DbCfg().DBConns[utils.MetaDefault].RplConns))
	}
}

func TestCgrCfgJSONDefaultsCDRS(t *testing.T) {
	eCdrsCfg := &CdrsCfg{
		Enabled:         false,
		SMCostRetries:   5,
		ChargerSConns:   []string{},
		AttributeSConns: []string{},
		ThresholdSConns: []string{},
		StatSConns:      []string{},
		ActionSConns:    []string{},
		EEsConns:        []string{},
		RateSConns:      []string{},
		AccountSConns:   []string{},
		ExtraFields:     utils.RSRParsers{},
		Opts: &CdrsOpts{
			Accounts:   []*DynamicBoolOpt{{}},
			Attributes: []*DynamicBoolOpt{{}},
			Chargers:   []*DynamicBoolOpt{{}},
			Export:     []*DynamicBoolOpt{{}},
			Rates:      []*DynamicBoolOpt{{}},
			Stats:      []*DynamicBoolOpt{{}},
			Thresholds: []*DynamicBoolOpt{{}},
			Refund:     []*DynamicBoolOpt{{}},
			Rerate:     []*DynamicBoolOpt{{}},
			Store:      []*DynamicBoolOpt{{value: true}},
		},
	}
	if !reflect.DeepEqual(eCdrsCfg, cgrCfg.cdrsCfg) {
		t.Errorf("Expecting: %+v , received: %+v", utils.ToJSON(eCdrsCfg), utils.ToJSON(cgrCfg.cdrsCfg))
	}
}

func TestCgrCfgJSONLoadCDRS(t *testing.T) {
	jsnCfg := `
{
"cdrs": {
	"enabled": true,
	"chargers_conns": ["*internal"],
},
}
	`
	cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg)
	if err != nil {
		t.Error(err)
	}
	if !cgrCfg.CdrsCfg().Enabled {
		t.Errorf("Expecting: true , received: %+v", cgrCfg.CdrsCfg().Enabled)
	}
	expected := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	if !reflect.DeepEqual(cgrCfg.CdrsCfg().ChargerSConns, expected) {
		t.Errorf("Expecting: %+v , received: %+v", expected, cgrCfg.CdrsCfg().ChargerSConns)
	}
}

func TestCgrCfgJSONDefaultsSMGenericCfg(t *testing.T) {
	eSessionSCfg := &SessionSCfg{
		Enabled:             false,
		ListenBiJSON:        "127.0.0.1:2014",
		ChargerSConns:       []string{},
		CDRsConns:           []string{},
		ResourceSConns:      []string{},
		IPsConns:            []string{},
		ThresholdSConns:     []string{},
		StatSConns:          []string{},
		RouteSConns:         []string{},
		AttributeSConns:     []string{},
		ActionSConns:        []string{},
		RateSConns:          []string{},
		AccountSConns:       []string{},
		ReplicationConns:    []string{},
		StoreSCosts:         false,
		SessionIndexes:      utils.StringSet{},
		ClientProtocol:      1.0,
		ChannelSyncInterval: 0,
		TerminateAttempts:   5,
		AlterableFields:     utils.NewStringSet([]string{}),
		STIRCfg: &STIRcfg{
			AllowedAttest:      utils.NewStringSet([]string{utils.MetaAny}),
			PayloadMaxduration: -1,
			DefaultAttest:      "A",
		},
		DefaultUsage: map[string]time.Duration{
			utils.MetaAny:   3 * time.Hour,
			utils.MetaVoice: 3 * time.Hour,
			utils.MetaData:  1048576,
			utils.MetaSMS:   1,
		},
		Opts: &SessionsOpts{
			Accounts:               []*DynamicBoolOpt{{}},
			Attributes:             []*DynamicBoolOpt{{}},
			CDRs:                   []*DynamicBoolOpt{{}},
			Chargers:               []*DynamicBoolOpt{{}},
			Resources:              []*DynamicBoolOpt{{}},
			IPs:                    []*DynamicBoolOpt{{}},
			Routes:                 []*DynamicBoolOpt{{}},
			Stats:                  []*DynamicBoolOpt{{}},
			Thresholds:             []*DynamicBoolOpt{{}},
			Initiate:               []*DynamicBoolOpt{{}},
			Update:                 []*DynamicBoolOpt{{}},
			Terminate:              []*DynamicBoolOpt{{}},
			Message:                []*DynamicBoolOpt{{}},
			AttributesDerivedReply: []*DynamicBoolOpt{{}},
			BlockerError:           []*DynamicBoolOpt{{}},
			CDRsDerivedReply:       []*DynamicBoolOpt{{}},
			ResourcesAuthorize:     []*DynamicBoolOpt{{}},
			ResourcesAllocate:      []*DynamicBoolOpt{{}},
			ResourcesRelease:       []*DynamicBoolOpt{{}},
			ResourcesDerivedReply:  []*DynamicBoolOpt{{}},
			IPsAuthorize:           []*DynamicBoolOpt{{}},
			IPsAllocate:            []*DynamicBoolOpt{{}},
			IPsRelease:             []*DynamicBoolOpt{{}},
			RoutesDerivedReply:     []*DynamicBoolOpt{{}},
			StatsDerivedReply:      []*DynamicBoolOpt{{}},
			ThresholdsDerivedReply: []*DynamicBoolOpt{{}},
			MaxUsage:               []*DynamicBoolOpt{{}},
			TTL:                    []*DynamicDurationOpt{{value: SessionsTTLDftOpt}},
			Chargeable:             []*DynamicBoolOpt{{value: SessionsChargeableDftOpt}},
			TTLLastUsage:           []*DynamicDurationPointerOpt{},
			TTLLastUsed:            []*DynamicDurationPointerOpt{},
			DebitInterval:          []*DynamicDurationOpt{{value: SessionsDebitIntervalDftOpt}},
			TTLMaxDelay:            []*DynamicDurationOpt{{value: SessionsTTLMaxDelayDftOpt}},
			TTLUsage:               []*DynamicDurationPointerOpt{},
			ForceUsage:             []*DynamicBoolOpt{},
			OriginID:               []*DynamicStringOpt{},
			AccountsForceUsage:     []*DynamicBoolOpt{},
		},
	}
	if !reflect.DeepEqual(eSessionSCfg, cgrCfg.sessionSCfg) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eSessionSCfg), utils.ToJSON(cgrCfg.sessionSCfg))
	}
}

func TestCgrCfgJSONDefaultsCacheCFG(t *testing.T) {
	eCacheCfg := &CacheCfg{
		Partitions: map[string]*CacheParamCfg{
			utils.CacheResourceProfiles:            {Limit: -1},
			utils.CacheResources:                   {Limit: -1},
			utils.CacheEventResources:              {Limit: -1, TTL: 0},
			utils.CacheIPProfiles:                  {Limit: -1},
			utils.CacheIPAllocations:               {Limit: -1},
			utils.CacheEventIPs:                    {Limit: -1, TTL: 0},
			utils.CacheStatQueueProfiles:           {Limit: -1},
			utils.CacheStatQueues:                  {Limit: -1},
			utils.CacheThresholdProfiles:           {Limit: -1},
			utils.CacheThresholds:                  {Limit: -1},
			utils.CacheTrendProfiles:               {Limit: -1},
			utils.CacheTrends:                      {Limit: -1},
			utils.CacheFilters:                     {Limit: -1},
			utils.CacheRouteProfiles:               {Limit: -1},
			utils.CacheAttributeProfiles:           {Limit: -1},
			utils.CacheChargerProfiles:             {Limit: -1},
			utils.CacheRateProfiles:                {Limit: -1},
			utils.CacheActionProfiles:              {Limit: -1},
			utils.CacheAccounts:                    {Limit: -1},
			utils.CacheResourceFilterIndexes:       {Limit: -1},
			utils.CacheIPFilterIndexes:             {Limit: -1},
			utils.CacheStatFilterIndexes:           {Limit: -1},
			utils.CacheThresholdFilterIndexes:      {Limit: -1},
			utils.CacheRouteFilterIndexes:          {Limit: -1},
			utils.CacheAttributeFilterIndexes:      {Limit: -1},
			utils.CacheChargerFilterIndexes:        {Limit: -1},
			utils.CacheRateProfilesFilterIndexes:   {Limit: -1},
			utils.CacheRankingProfiles:             {Limit: -1},
			utils.CacheRankings:                    {Limit: -1},
			utils.CacheRateFilterIndexes:           {Limit: -1},
			utils.CacheActionProfilesFilterIndexes: {Limit: -1},
			utils.CacheAccountsFilterIndexes:       {Limit: -1},
			utils.CacheReverseFilterIndexes:        {Limit: -1},
			utils.CacheDiameterMessages: {
				Limit: -1,
				TTL:   3 * time.Hour,
			},
			utils.CacheRadiusPackets: {
				Limit:     -1,
				TTL:       3 * time.Hour,
				Remote:    false,
				StaticTTL: false,
			},
			utils.CacheRPCResponses: {Limit: 0,
				TTL: 2 * time.Second},
			utils.CacheClosedSessions: {Limit: -1,
				TTL: 10 * time.Second},
			utils.CacheEventCharges: {Limit: 0,
				TTL: 10 * time.Second},
			utils.CacheCDRIDs: {Limit: -1,
				TTL: 10 * time.Minute},
			utils.CacheLoadIDs: {Limit: -1},
			utils.CacheRPCConnections: {Limit: -1,
				TTL: 0},
			utils.CacheUCH: {Limit: -1,
				TTL: 3 * time.Hour},
			utils.CacheSTIR: {Limit: -1,
				TTL: 3 * time.Hour},
			utils.CacheCapsEvents: {Limit: -1},

			utils.MetaAPIBan: {Limit: -1,
				TTL: 2 * time.Minute},
			utils.MetaSentryPeer: {Limit: -1,
				TTL: 24 * time.Hour, StaticTTL: true},
			utils.CacheReplicationHosts: {},
		},
		ReplicationConns: []string{},
		RemoteConns:      []string{},
	}

	if !reflect.DeepEqual(eCacheCfg, cgrCfg.CacheCfg()) {
		t.Errorf("received: %s, \nexpecting: %s",
			utils.ToJSON(cgrCfg.CacheCfg()), utils.ToJSON(eCacheCfg))
	}
}

func TestCgrCfgJSONDefaultsFsAgentConfig(t *testing.T) {
	eFsAgentCfg := &FsAgentCfg{
		Enabled:                false,
		SessionSConns:          []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		SubscribePark:          true,
		CreateCDR:              false,
		ExtraFields:            utils.RSRParsers{},
		EmptyBalanceContext:    "",
		EmptyBalanceAnnFile:    "",
		MaxWaitConnection:      2 * time.Second,
		ActiveSessionDelimiter: ",",
		EventSocketConns: []*FsConnCfg{{
			Address:      "127.0.0.1:8021",
			Password:     "ClueCon",
			Reconnects:   5,
			Alias:        "127.0.0.1:8021",
			ReplyTimeout: time.Minute,
		}},
	}

	if !reflect.DeepEqual(cgrCfg.fsAgentCfg, eFsAgentCfg) {
		t.Errorf("received: %+v, expecting: %+v", utils.ToJSON(cgrCfg.fsAgentCfg), utils.ToJSON(eFsAgentCfg))
	}
}

func TestCgrCfgJSONDefaultsKamAgentConfig(t *testing.T) {
	eKamAgentCfg := &KamAgentCfg{
		Enabled:       false,
		SessionSConns: []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		CreateCdr:     false,
		EvapiConns: []*KamConnCfg{{
			Address:    "127.0.0.1:8448",
			Reconnects: 5,
		}},
	}
	if !reflect.DeepEqual(cgrCfg.kamAgentCfg, eKamAgentCfg) {
		t.Errorf("received: %+v, expecting: %+v",
			utils.ToJSON(cgrCfg.kamAgentCfg), utils.ToJSON(eKamAgentCfg))
	}
}

func TestCgrCfgJSONDefaultssteriskAgentCfg(t *testing.T) {
	eAstAgentCfg := &AsteriskAgentCfg{
		Enabled:       false,
		SessionSConns: []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		CreateCDR:     false,
		AsteriskConns: []*AsteriskConnCfg{
			{Address: "127.0.0.1:8088",
				User: "cgrates", Password: "CGRateS.org",
				ConnectAttempts: 3, Reconnects: 5}},
	}

	if !reflect.DeepEqual(cgrCfg.asteriskAgentCfg, eAstAgentCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.asteriskAgentCfg, eAstAgentCfg)
	}
}

func TestCgrCfgJSONDefaultFiltersCfg(t *testing.T) {
	eFiltersCfg := &FilterSCfg{
		StatSConns:     []string{},
		ResourceSConns: []string{},
		AccountSConns:  []string{},
		TrendSConns:    []string{},
		RankingSConns:  []string{},
	}
	if !reflect.DeepEqual(cgrCfg.filterSCfg, eFiltersCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.filterSCfg, eFiltersCfg)
	}
}

func TestCgrCfgJSONDefaultSChargerSCfg(t *testing.T) {
	eChargerSCfg := &ChargerSCfg{
		Enabled:                false,
		IndexedSelects:         true,
		AttributeSConns:        []string{},
		StringIndexedFields:    nil,
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
	}
	if !reflect.DeepEqual(eChargerSCfg, cgrCfg.chargerSCfg) {
		t.Errorf("received: %+v, expecting: %+v", eChargerSCfg, cgrCfg.chargerSCfg)
	}
}

func TestCgrCfgJSONDefaultsResLimCfg(t *testing.T) {
	eResLiCfg := &ResourceSConfig{
		Enabled:                false,
		IndexedSelects:         true,
		ThresholdSConns:        []string{},
		StoreInterval:          0,
		StringIndexedFields:    nil,
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
		Opts: &ResourcesOpts{
			UsageID:  []*DynamicStringOpt{{value: ResourcesUsageIDDftOpt}},
			UsageTTL: []*DynamicDurationOpt{{value: ResourcesUsageTTLDftOpt}},
			Units:    []*DynamicFloat64Opt{{value: ResourcesUnitsDftOpt}},
		},
	}
	if !reflect.DeepEqual(cgrCfg.resourceSCfg, eResLiCfg) {
		t.Errorf("expecting: %s, received: %s", utils.ToJSON(eResLiCfg), utils.ToJSON(cgrCfg.resourceSCfg))
	}

}

func TestCgrCfgJSONDefaultStatsCfg(t *testing.T) {
	eStatsCfg := &StatSCfg{
		Enabled:                false,
		IndexedSelects:         true,
		StoreInterval:          0,
		ThresholdSConns:        []string{},
		StringIndexedFields:    nil,
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
		Opts: &StatsOpts{
			ProfileIDs:           []*DynamicStringSliceOpt{},
			ProfileIgnoreFilters: []*DynamicBoolOpt{{value: StatsProfileIgnoreFilters}},
			RoundingDecimals:     []*DynamicIntOpt{},
		},
		EEsConns: []string{},
	}
	if !reflect.DeepEqual(cgrCfg.statsCfg, eStatsCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.statsCfg, eStatsCfg)
	}
}

func TestCgrCfgJSONDefaultThresholdSCfg(t *testing.T) {
	eThresholdSCfg := &ThresholdSCfg{
		Enabled:                false,
		IndexedSelects:         true,
		StoreInterval:          0,
		StringIndexedFields:    nil,
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
		ActionSConns:           []string{},
		Opts: &ThresholdsOpts{
			ProfileIDs:           []*DynamicStringSliceOpt{},
			ProfileIgnoreFilters: []*DynamicBoolOpt{{value: ThresholdsProfileIgnoreFiltersDftOpt}},
		},
		EEsConns:       []string{},
		EEsExporterIDs: []string{},
	}
	if !reflect.DeepEqual(eThresholdSCfg, cgrCfg.thresholdSCfg) {
		t.Errorf("received: %s, expecting: %s", utils.ToJSON(eThresholdSCfg), utils.ToJSON(cgrCfg.thresholdSCfg))
	}
}

func TestCgrCfgJSONDefaultRouteSCfg(t *testing.T) {
	eSupplSCfg := &RouteSCfg{
		Enabled:                false,
		IndexedSelects:         true,
		StringIndexedFields:    nil,
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
		AttributeSConns:        []string{},
		ResourceSConns:         []string{},
		StatSConns:             []string{},
		RateSConns:             []string{},
		AccountSConns:          []string{},
		DefaultRatio:           1,
		Opts: &RoutesOpts{
			Context:      []*DynamicStringOpt{{value: RoutesContextDftOpt}},
			ProfileCount: []*DynamicIntPointerOpt{{value: RoutesProfileCountDftOpt}},
			IgnoreErrors: []*DynamicBoolOpt{{value: RoutesIgnoreErrorsDftOpt}},
			MaxCost:      []*DynamicInterfaceOpt{{Value: RoutesMaxCostDftOpt}},
			Limit:        []*DynamicIntPointerOpt{},
			Offset:       []*DynamicIntPointerOpt{},
			MaxItems:     []*DynamicIntPointerOpt{},
			Usage:        []*DynamicDecimalOpt{{value: RatesUsageDftOpt}},
		},
	}
	if !reflect.DeepEqual(eSupplSCfg, cgrCfg.routeSCfg) {
		t.Errorf("expected: %+v, received: %+v", utils.ToJSON(eSupplSCfg), utils.ToJSON(cgrCfg.routeSCfg))
	}
}

func TestCgrCfgJSONDefaultsDiameterAgentCfg(t *testing.T) {
	testDA := &DiameterAgentCfg{
		Enabled:           false,
		Listen:            "127.0.0.1:3868",
		ListenNet:         utils.TCP,
		DictionariesPath:  "/usr/share/cgrates/diameter/dict/",
		SessionSConns:     []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		StatSConns:        []string{},
		ThresholdSConns:   []string{},
		OriginHost:        "CGR-DA",
		OriginRealm:       "cgrates.org",
		VendorID:          0,
		ProductName:       "CGRateS",
		RequestProcessors: nil,
	}

	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.Enabled, testDA.Enabled) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.Enabled, testDA.Enabled)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.Listen, testDA.Listen) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.Listen, testDA.Listen)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.ListenNet, testDA.ListenNet) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.ListenNet, testDA.ListenNet)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.DictionariesPath, testDA.DictionariesPath) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.DictionariesPath, testDA.DictionariesPath)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.SessionSConns, testDA.SessionSConns) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.SessionSConns, testDA.SessionSConns)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.StatSConns, testDA.StatSConns) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.StatSConns, testDA.StatSConns)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.ThresholdSConns, testDA.ThresholdSConns) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.ThresholdSConns, testDA.ThresholdSConns)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.OriginHost, testDA.OriginHost) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.diameterAgentCfg.OriginHost, testDA.OriginHost)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.OriginRealm, testDA.OriginRealm) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.diameterAgentCfg.OriginRealm, testDA.OriginRealm)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.VendorID, testDA.VendorID) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.diameterAgentCfg.VendorID, testDA.VendorID)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.ProductName, testDA.ProductName) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.diameterAgentCfg.ProductName, testDA.ProductName)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.RequestProcessors, testDA.RequestProcessors) {
		t.Errorf("expecting: %+v, received: %+v", testDA.RequestProcessors, cgrCfg.diameterAgentCfg.RequestProcessors)
	}
}

func TestCgrCfgJSONDefaultsSureTax(t *testing.T) {
	localt, err := time.LoadLocation("Local")
	if err != nil {
		t.Error("time parsing error", err)
	}
	eSureTaxCfg := &SureTaxCfg{
		URL:                  "",
		ClientNumber:         "",
		ValidationKey:        "",
		BusinessUnit:         "",
		Timezone:             localt,
		IncludeLocalCost:     false,
		ReturnFileCode:       "0",
		ResponseGroup:        "03",
		ResponseType:         "D4",
		RegulatoryCode:       "03",
		ClientTracking:       utils.NewRSRParsersMustCompile("~*opts.*originID", utils.InfieldSep),
		CustomerNumber:       utils.NewRSRParsersMustCompile("~*req.Subject", utils.InfieldSep),
		OrigNumber:           utils.NewRSRParsersMustCompile("~*req.Subject", utils.InfieldSep),
		TermNumber:           utils.NewRSRParsersMustCompile("~*req.Destination", utils.InfieldSep),
		BillToNumber:         utils.NewRSRParsersMustCompile("", utils.InfieldSep),
		Zipcode:              utils.NewRSRParsersMustCompile("", utils.InfieldSep),
		P2PZipcode:           utils.NewRSRParsersMustCompile("", utils.InfieldSep),
		P2PPlus4:             utils.NewRSRParsersMustCompile("", utils.InfieldSep),
		Units:                utils.NewRSRParsersMustCompile("1", utils.InfieldSep),
		UnitType:             utils.NewRSRParsersMustCompile("00", utils.InfieldSep),
		TaxIncluded:          utils.NewRSRParsersMustCompile("0", utils.InfieldSep),
		TaxSitusRule:         utils.NewRSRParsersMustCompile("04", utils.InfieldSep),
		TransTypeCode:        utils.NewRSRParsersMustCompile("010101", utils.InfieldSep),
		SalesTypeCode:        utils.NewRSRParsersMustCompile("R", utils.InfieldSep),
		TaxExemptionCodeList: utils.NewRSRParsersMustCompile("", utils.InfieldSep),
	}

	if !reflect.DeepEqual(cgrCfg.sureTaxCfg, eSureTaxCfg) {
		t.Errorf("received: %+v, expecting: %+v", utils.ToJSON(cgrCfg.sureTaxCfg), utils.ToJSON(eSureTaxCfg))
	}
}

func TestCgrCfgJSONDefaultsHTTP(t *testing.T) {
	if cgrCfg.HTTPCfg().JsonRPCURL != "/jsonrpc" {
		t.Errorf("expecting: /jsonrpc , received: %+v", cgrCfg.HTTPCfg().JsonRPCURL)
	}
	if cgrCfg.HTTPCfg().WSURL != "/ws" {
		t.Errorf("expecting: /ws , received: %+v", cgrCfg.HTTPCfg().WSURL)
	}
	if cgrCfg.HTTPCfg().FreeswitchCDRsURL != "/freeswitch_json" {
		t.Errorf("expecting: /freeswitch_json , received: %+v", cgrCfg.HTTPCfg().FreeswitchCDRsURL)
	}
	if cgrCfg.HTTPCfg().CDRsURL != "/cdr_http" {
		t.Errorf("expecting: /cdr_http , received: %+v", cgrCfg.HTTPCfg().CDRsURL)
	}
	if cgrCfg.HTTPCfg().UseBasicAuth != false {
		t.Errorf("expecting: false , received: %+v", cgrCfg.HTTPCfg().UseBasicAuth)
	}
	if !reflect.DeepEqual(cgrCfg.HTTPCfg().AuthUsers, map[string]string{}) {
		t.Errorf("expecting: %+v , received: %+v", map[string]string{}, cgrCfg.HTTPCfg().AuthUsers)
	}
}

func TestRadiusAgentCfg(t *testing.T) {
	testRA := &RadiusAgentCfg{
		Enabled: false,
		Listeners: []RadiusListener{
			{
				Network:  "udp",
				AuthAddr: "127.0.0.1:1812",
				AcctAddr: "127.0.0.1:1813",
			},
		},
		ClientSecrets:      map[string]string{utils.MetaDefault: "CGRateS.org"},
		ClientDictionaries: map[string][]string{utils.MetaDefault: {"/usr/share/cgrates/radius/dict/"}},
		SessionSConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		StatSConns:         []string{},
		ThresholdSConns:    []string{},
		DMRTemplate:        utils.MetaDMR,
		CoATemplate:        utils.MetaCoA,
		RequestProcessors:  nil,
	}
	if !reflect.DeepEqual(cgrCfg.radiusAgentCfg, testRA) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.radiusAgentCfg, testRA)
	}
}

func TestDbDefaults(t *testing.T) {
	dbs := []string{utils.MetaMongo, utils.MetaRedis, utils.MetaMySQL, utils.MetaInternal, utils.MetaPostgres}
	for _, dbtype := range dbs {
		port := defaultDBPort(dbtype, "1234")
		if port != "1234" {
			t.Errorf("Expected %+v, received %+v", "1234", port)
		}
	}
}

func TestNewCGRConfigFromJSONStringWithDefaultsError(t *testing.T) {
	cfgJSONStr := "invalidJSON"
	expectedErr := "invalid character 'i' looking for beginning of value around line 1 and position 1\n line: \"invalidJSON\""
	if _, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	cfgJSONStr = `{
	"suretax": {
        "tax_exemption_code_list": "a{*"
    },
}`
	expectedErr = "invalid converter terminator in rule: <a{*>"
	if _, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestIsHidden(t *testing.T) {
	file := ".newFile"
	if !isHidden(file) {
		t.Errorf("File is not hidden")
	}

	file = "."
	if isHidden(file) {
		t.Errorf("Invalid input")
	}
}

func TestLoadRPCConnsError(t *testing.T) {
	cfgJSONStr := `{
     "rpc_conns": {
	     "*localhost": {
		     "conns": [
                  {"address": "127.0.0.1:2018", "TLS": true, "synchronous": true, "transport": "*json"},
             ],
             "poolSize": "two",
	      },
     },		
}`
	expected := "json: cannot unmarshal string into Go struct field RPCConnJson.PoolSize of type int"
	cgrCfg := NewDefaultCGRConfig()
	if cgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrCfg.rpcConns.Load(context.Background(), cgrJSONCfg, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadGeneralCfgError(t *testing.T) {
	cfgJSONStr := `{
      "general": {
            "node_id": [],
            "locking_timeout": "0",
            "failed_posts_ttl": "0s",
            "connect_timeout": "0s",
            "reply_timeout": "0s",
        }
    }
}`
	expected := "json: cannot unmarshal array into Go struct field GeneralJsonCfg.Node_id of type string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.generalCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadCacheCfgError(t *testing.T) {
	cfgJSONStr := `{
"caches":{
    "replication_conns": 2,
	},
}`
	expected := "json: cannot unmarshal number into Go struct field CacheJsonCfg.Replication_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.cacheCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadListenCfgError(t *testing.T) {
	cfgJSONStr := `{
	"listen": {	
        "http_tls": 1206,			
	}
}`
	expected := "json: cannot unmarshal number into Go struct field ListenJsonCfg.Http_tls of type string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.listenCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadHTTPCfgError(t *testing.T) {
	cfgJSONStr := `{
	"http": {
	   "auth_users": "user1",
     },
}`
	expected := "json: cannot unmarshal string into Go struct field HTTPJsonCfg.auth_users of type map[string]string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.httpCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadDataDBCfgErrorCase1(t *testing.T) {
	cfgJSONStr := `{
"db": {
	"db_conns": {
		"*default": {	
			"db_host": 127.0,
			},
		},
	}
}`
	expected := "json: cannot unmarshal number into Go struct field DbConnJson.Db_conns.Db_host of type string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.dbCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadDataDBCfgErrorCase2(t *testing.T) {
	cfgJSONStr := `{
"db": {
	"db_conns": {
		"*default": {	
			"remote_conns":["*internal"],
			},
		},
	}
}`
	expected := "remote connection ID needs to be different than <*internal> "
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else {
		cgrConfig.dbCfg.DBConns[utils.MetaDefault].RmtConns = []string{utils.MetaInternal}
		if err := cgrConfig.dbCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
			t.Errorf("Expected %+v, received %+v", expected, err)
		}
	}
}

func TestLoadFilterSCfgError(t *testing.T) {
	cfgJSONStr := `{
		"filters": {								
			"stats_conns": "*internal",
	},
}`
	expected := "json: cannot unmarshal string into Go struct field FilterSJsonCfg.Stats_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.filterSCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadCdrsCfgError(t *testing.T) {
	cfgJSONStr := `{
	"cdrs": {	
        "ees_conns": "*internal",
	},
}`
	expected := "json: cannot unmarshal string into Go struct field CdrsJsonCfg.Ees_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.cdrsCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadFreeswitchAgentCfgError(t *testing.T) {
	cfgJSONStr := `{
	"freeswitch_agent": {
          "sessions_conns": "*conn1",
	},
}`
	expected := "json: cannot unmarshal string into Go struct field FreeswitchAgentJsonCfg.sessions_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.fsAgentCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadKamAgentCfgError(t *testing.T) {
	cfgJSONStr := `{
		"kamailio_agent": {
			"timezone": 1234,
		},
	}`
	expected := "json: cannot unmarshal number into Go struct field KamAgentJsonCfg.Timezone of type string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.kamAgentCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadAsteriskAgentCfgError(t *testing.T) {
	cfgJSONStr := `{
	"asterisk_agent": {
		"sessions_conns": "*conn1",
	},
}`
	expected := "json: cannot unmarshal string into Go struct field AsteriskAgentJsonCfg.Sessions_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.asteriskAgentCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadDiameterAgentCfgError(t *testing.T) {
	cfgJSONStr := `{ 
      "diameter_agent": {
        "request_processors": [
	        {
		       "id": 1,
            },
         ]
      }
}`
	expected := "json: cannot unmarshal number into Go struct field ReqProcessorJsnCfg.request_processors.ID of type string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.diameterAgentCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadRadiusAgentCfgError(t *testing.T) {
	cfgJSONStr := `{
"radius_agent": {
	"listeners":[
		{
			"auth_address": 1
		}
	]
}
}`
	expected := "json: cannot unmarshal number into Go struct field RadiusListenerJsonCfg.listeners.auth_address of type string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.radiusAgentCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadDNSAgentCfgError(t *testing.T) {
	cfgJSONStr := `{
		"dns_agent": {
			"listeners":[
				{
					"address": 1278,
				}
			],
		},
	}`
	expected := "json: cannot unmarshal number into Go struct field ListenerJsnCfg.listeners.Address of type string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.dnsAgentCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadHttpAgentCfgError(t *testing.T) {
	cfgJSONStr := `{
"http_agent": [
	{
		"id": ["randomID"],
		},
	],	
}`
	expected := "json: cannot unmarshal array into Go struct field HttpAgentJsonCfg.id of type string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.httpAgentCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadChargerSCfgError(t *testing.T) {
	cfgJSONStr := `{
	"chargers": {
		"prefix_indexed_fields": "prefix",	
	},	
}`
	expected := "json: cannot unmarshal string into Go struct field ChargerSJsonCfg.Prefix_indexed_fields of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.chargerSCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadResourceSCfgError(t *testing.T) {
	cfgJSONStr := `{
		"resources": {
            "string_indexed_fields": "*req.index1",
		},	
	}`
	expected := "json: cannot unmarshal string into Go struct field ResourceSJsonCfg.String_indexed_fields of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.resourceSCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadStatSCfgError(t *testing.T) {
	cfgJSONStr := `{
		"stats": {								
            "string_indexed_fields": "*req.string",
		},	
}`
	expected := "json: cannot unmarshal string into Go struct field StatServJsonCfg.String_indexed_fields of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.statsCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadThresholdSCfgError(t *testing.T) {
	cfgJSONStr := `{
		"thresholds": {
			"store_interval": 96,						
		},		
}`
	expected := "json: cannot unmarshal number into Go struct field ThresholdSJsonCfg.Store_interval of type string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.thresholdSCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadLoaderSCfgError(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [												
	{
		"run_delay": 0,
		},
	],	
}`
	expected := "json: cannot unmarshal number into Go struct field LoaderJsonCfg.Run_delay of type string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loaderCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadRouteSCfgError(t *testing.T) {
	cfgJSONStr := `{
		"routes": {
            "string_indexed_fields": "*req.string",
		},
	}`
	expected := "json: cannot unmarshal string into Go struct field RouteSJsonCfg.String_indexed_fields of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.routeSCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadSureTaxCfgError(t *testing.T) {
	cfgJSONStr := `{
	"suretax": {
		"sales_type_code": 123,
    },
}`
	expected := "json: cannot unmarshal number into Go struct field SureTaxJsonCfg.Sales_type_code of type string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.sureTaxCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadLoaderCgrCfgError(t *testing.T) {
	cfgJSONStr := `{
	"loader": {
		"caches_conns":"*localhost",
	},
}`
	expected := "json: cannot unmarshal string into Go struct field LoaderCfgJson.Caches_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loaderCgrCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadMigratorCgrCfgError(t *testing.T) {
	cfgJSONStr := `{
	"migrator": {
        "users_filters": "users",
	},
}`
	expected := "json: cannot unmarshal string into Go struct field MigratorCfgJson.Users_filters of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.migratorCgrCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadTlsCgrCfgError(t *testing.T) {
	cfgJSONStr := `	{
	"tls":{
		"server_policy": "3",					
	},
}`
	expected := "json: cannot unmarshal string into Go struct field TlsJsonCfg.Server_policy of type int"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.tlsCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadAnalyzerCgrCfgError(t *testing.T) {
	cfgJSONStr := `{
		"analyzers":{
            "enabled": 10,  
        },
    }
}`
	expected := "json: cannot unmarshal number into Go struct field AnalyzerSJsonCfg.Enabled of type bool"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.analyzerSCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadAPIBanCgrCfgError(t *testing.T) {
	cfgJSONStr := `{
		"apiban":{
			"enabled": "no",
		},
}`
	expected := "json: cannot unmarshal string into Go struct field APIBanJsonCfg.Enabled of type bool"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.apiBanCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadApierCfgError(t *testing.T) {
	myJSONStr := `{
    "admins": {
       "actions_conns": "*internal",
    },
}`
	expected := "json: cannot unmarshal string into Go struct field AdminSJsonCfg.Actions_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(myJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.admS.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadErsCfgError(t *testing.T) {
	cfgJSONStr := `{
"ers": {										
	"sessions_conns": "*internal",
},
}`
	expected := "json: cannot unmarshal string into Go struct field ERsJsonCfg.sessions_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.ersCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadEesCfgError(t *testing.T) {
	cfgJSONStr := `{
      "ees": { 
            "attributes_conns": "*conn1",
	  }
    }`
	expected := "json: cannot unmarshal string into Go struct field EEsJsonCfg.Attributes_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.eesCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadCoreSCfgError(t *testing.T) {
	cfgJSONStr := `{
      "cores": { 
            "caps": "1",
	  }
    }`
	expected := "json: cannot unmarshal string into Go struct field CoreSJsonCfg.Caps of type int"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.coreSCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadRateSCfgError(t *testing.T) {
	cfgJSONStr := `{
     "rates": {
	         "string_indexed_fields": "*req.index",
     },
}`
	expected := "json: cannot unmarshal string into Go struct field RateSJsonCfg.String_indexed_fields of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.rateSCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadSIPAgentCfgError(t *testing.T) {
	cfgJSONStr := `{
	"sip_agent": {
		"request_processors": [
             {
               "id": 1234,
             },
		],
	},
}`
	expected := "json: cannot unmarshal number into Go struct field ReqProcessorJsnCfg.request_processors.ID of type string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.sipAgentCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadTemplateSCfgError(t *testing.T) {
	cfgJSONStr := `{
     "templates": {
           "custom_template": [
              {
                "tag": 1234,
            },
           ],
     }
}`
	expected := "json: cannot unmarshal number into Go struct field FcTemplateJsonCfg.Tag of type string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.templates.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	cfgJSONStr = `{
     "templates": {
           "custom_template": [
              {
                "value": "a{*",
            },
           ],
     }
}`
	expected = "invalid converter terminator in rule: <a{*>"
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.templates.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadConfigsCfgError(t *testing.T) {
	cfgJSONStr := `{
      "configs": {
          "url": 123,
      },
}`
	expected := "json: cannot unmarshal number into Go struct field ConfigSCfgJson.Url of type string"
	cgrConfig := NewDefaultCGRConfig()
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.configSCfg.Load(context.Background(), cgrCfgJSON, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSuretaxConfig(t *testing.T) {
	tLocal, err := time.LoadLocation("Local")
	if err != nil {
		t.Error(err)
	}
	expected := &SureTaxCfg{
		URL:                  "",
		ClientNumber:         "",
		ValidationKey:        "",
		BusinessUnit:         "",
		Timezone:             tLocal,
		IncludeLocalCost:     false,
		ReturnFileCode:       "0",
		ResponseGroup:        "03",
		ResponseType:         "D4",
		RegulatoryCode:       "03",
		ClientTracking:       utils.NewRSRParsersMustCompile("~*opts.*originID", utils.InfieldSep),
		CustomerNumber:       utils.NewRSRParsersMustCompile("~*req.Subject", utils.InfieldSep),
		OrigNumber:           utils.NewRSRParsersMustCompile("~*req.Subject", utils.InfieldSep),
		TermNumber:           utils.NewRSRParsersMustCompile("~*req.Destination", utils.InfieldSep),
		BillToNumber:         utils.NewRSRParsersMustCompile(utils.EmptyString, utils.InfieldSep),
		Zipcode:              utils.NewRSRParsersMustCompile(utils.EmptyString, utils.InfieldSep),
		Plus4:                utils.NewRSRParsersMustCompile(utils.EmptyString, utils.InfieldSep),
		P2PZipcode:           utils.NewRSRParsersMustCompile(utils.EmptyString, utils.InfieldSep),
		P2PPlus4:             utils.NewRSRParsersMustCompile(utils.EmptyString, utils.InfieldSep),
		Units:                utils.NewRSRParsersMustCompile("1", utils.InfieldSep),
		UnitType:             utils.NewRSRParsersMustCompile("00", utils.InfieldSep),
		TaxIncluded:          utils.NewRSRParsersMustCompile("0", utils.InfieldSep),
		TaxSitusRule:         utils.NewRSRParsersMustCompile("04", utils.InfieldSep),
		TransTypeCode:        utils.NewRSRParsersMustCompile("010101", utils.InfieldSep),
		SalesTypeCode:        utils.NewRSRParsersMustCompile("R", utils.InfieldSep),
		TaxExemptionCodeList: utils.NewRSRParsersMustCompile(utils.EmptyString, utils.InfieldSep),
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.SureTaxCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestDiameterAgentConfig(t *testing.T) {
	expected := &DiameterAgentCfg{
		Enabled:           false,
		ListenNet:         "tcp",
		Listen:            "127.0.0.1:3868",
		DictionariesPath:  "/usr/share/cgrates/diameter/dict/",
		SessionSConns:     []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		StatSConns:        []string{},
		ThresholdSConns:   []string{},
		OriginHost:        "CGR-DA",
		OriginRealm:       "cgrates.org",
		VendorID:          0,
		ProductName:       "CGRateS",
		SyncedConnReqs:    false,
		ASRTemplate:       "",
		RARTemplate:       "",
		ForcedDisconnect:  "*none",
		RequestProcessors: nil,
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.DiameterAgentCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestRadiusAgentConfig(t *testing.T) {
	expected := &RadiusAgentCfg{
		Enabled: false,
		Listeners: []RadiusListener{
			{
				Network:  "udp",
				AuthAddr: "127.0.0.1:1812",
				AcctAddr: "127.0.0.1:1813",
			},
		},
		ClientSecrets:      map[string]string{utils.MetaDefault: "CGRateS.org"},
		ClientDictionaries: map[string][]string{utils.MetaDefault: {"/usr/share/cgrates/radius/dict/"}},
		SessionSConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		StatSConns:         []string{},
		ThresholdSConns:    []string{},
		DMRTemplate:        utils.MetaDMR,
		CoATemplate:        utils.MetaCoA,
		RequestProcessors:  nil,
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.RadiusAgentCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestDNSAgentConfig(t *testing.T) {
	expected := &DNSAgentCfg{
		Enabled: false,
		Listeners: []DNSListener{
			{
				Address: "127.0.0.1:53",
				Network: "udp",
			},
		},
		SessionSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		StatSConns:        []string{},
		ThresholdSConns:   []string{},
		Timezone:          "",
		RequestProcessors: nil,
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.DNSAgentCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestAttributeSConfig(t *testing.T) {
	expected := &AttributeSCfg{
		Enabled:                false,
		AccountSConns:          []string{},
		StatSConns:             []string{},
		ResourceSConns:         []string{},
		IndexedSelects:         true,
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
		NestedFields:           false,
		Opts: &AttributesOpts{
			ProfileIDs:           []*DynamicStringSliceOpt{},
			ProcessRuns:          []*DynamicIntOpt{{value: AttributesProcessRunsDftOpt}},
			ProfileRuns:          []*DynamicIntOpt{{value: AttributesProfileRunsDftOpt}},
			ProfileIgnoreFilters: []*DynamicBoolOpt{{value: AttributesProfileIgnoreFiltersDftOpt}},
		},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.AttributeSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestChargersConfig(t *testing.T) {
	expected := &ChargerSCfg{
		Enabled:                false,
		IndexedSelects:         true,
		AttributeSConns:        []string{},
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
		NestedFields:           false,
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.ChargerSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestResourceSConfig(t *testing.T) {
	expected := &ResourceSConfig{
		Enabled:                false,
		IndexedSelects:         true,
		StoreInterval:          0,
		ThresholdSConns:        []string{},
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
		NestedFields:           false,
		Opts: &ResourcesOpts{
			UsageID:  []*DynamicStringOpt{{value: ResourcesUsageIDDftOpt}},
			UsageTTL: []*DynamicDurationOpt{{value: ResourcesUsageTTLDftOpt}},
			Units:    []*DynamicFloat64Opt{{value: ResourcesUnitsDftOpt}},
		},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.ResourceSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestStatSConfig(t *testing.T) {
	expected := &StatSCfg{
		Enabled:                false,
		IndexedSelects:         true,
		StoreInterval:          0,
		StoreUncompressedLimit: 0,
		ThresholdSConns:        []string{},
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
		NestedFields:           false,
		Opts: &StatsOpts{
			ProfileIDs:           []*DynamicStringSliceOpt{},
			ProfileIgnoreFilters: []*DynamicBoolOpt{{value: StatsProfileIgnoreFilters}},
			RoundingDecimals:     []*DynamicIntOpt{},
		},
		EEsConns: []string{},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.StatSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestThresholdSConfig(t *testing.T) {
	expected := &ThresholdSCfg{
		Enabled:                false,
		IndexedSelects:         true,
		StoreInterval:          0,
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
		NestedFields:           false,
		ActionSConns:           []string{},
		Opts: &ThresholdsOpts{
			ProfileIDs:           []*DynamicStringSliceOpt{},
			ProfileIgnoreFilters: []*DynamicBoolOpt{{value: ThresholdsProfileIgnoreFiltersDftOpt}},
		},
		EEsConns:       []string{},
		EEsExporterIDs: []string{},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.ThresholdSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestRouteSConfig(t *testing.T) {
	expected := &RouteSCfg{
		Enabled:                false,
		IndexedSelects:         true,
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		ExistsIndexedFields:    &[]string{},
		NotExistsIndexedFields: &[]string{},
		AttributeSConns:        []string{},
		ResourceSConns:         []string{},
		StatSConns:             []string{},
		RateSConns:             []string{},
		AccountSConns:          []string{},
		DefaultRatio:           1,
		NestedFields:           false,
		Opts: &RoutesOpts{
			Context:      []*DynamicStringOpt{{value: RoutesContextDftOpt}},
			ProfileCount: []*DynamicIntPointerOpt{{value: RoutesProfileCountDftOpt}},
			IgnoreErrors: []*DynamicBoolOpt{{value: RoutesIgnoreErrorsDftOpt}},
			MaxCost:      []*DynamicInterfaceOpt{{Value: RoutesMaxCostDftOpt}},
			Limit:        []*DynamicIntPointerOpt{},
			Offset:       []*DynamicIntPointerOpt{},
			MaxItems:     []*DynamicIntPointerOpt{},
			Usage:        []*DynamicDecimalOpt{{value: RatesUsageDftOpt}},
		},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.RouteSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestSessionSConfig(t *testing.T) {
	expected := &SessionSCfg{
		Enabled:             false,
		ListenBiJSON:        "127.0.0.1:2014",
		ChargerSConns:       []string{},
		ResourceSConns:      []string{},
		IPsConns:            []string{},
		ThresholdSConns:     []string{},
		StatSConns:          []string{},
		RouteSConns:         []string{},
		AttributeSConns:     []string{},
		CDRsConns:           []string{},
		ActionSConns:        []string{},
		RateSConns:          []string{},
		AccountSConns:       []string{},
		ReplicationConns:    []string{},
		StoreSCosts:         false,
		SessionIndexes:      utils.StringSet{},
		ClientProtocol:      1.0,
		ChannelSyncInterval: 0,
		TerminateAttempts:   5,
		AlterableFields:     utils.StringSet{},
		STIRCfg: &STIRcfg{
			AllowedAttest:      utils.StringSet{utils.MetaAny: {}},
			PayloadMaxduration: -1,
			DefaultAttest:      "A",
			PrivateKeyPath:     "",
			PublicKeyPath:      "",
		},
		DefaultUsage: map[string]time.Duration{
			utils.MetaAny:   3 * time.Hour,
			utils.MetaVoice: 3 * time.Hour,
			utils.MetaData:  1048576,
			utils.MetaSMS:   1,
		},
		Opts: &SessionsOpts{
			Accounts:               []*DynamicBoolOpt{{}},
			Attributes:             []*DynamicBoolOpt{{}},
			CDRs:                   []*DynamicBoolOpt{{}},
			Chargers:               []*DynamicBoolOpt{{}},
			Resources:              []*DynamicBoolOpt{{}},
			IPs:                    []*DynamicBoolOpt{{}},
			Routes:                 []*DynamicBoolOpt{{}},
			Stats:                  []*DynamicBoolOpt{{}},
			Thresholds:             []*DynamicBoolOpt{{}},
			Initiate:               []*DynamicBoolOpt{{}},
			Update:                 []*DynamicBoolOpt{{}},
			Terminate:              []*DynamicBoolOpt{{}},
			Message:                []*DynamicBoolOpt{{}},
			AttributesDerivedReply: []*DynamicBoolOpt{{}},
			BlockerError:           []*DynamicBoolOpt{{}},
			CDRsDerivedReply:       []*DynamicBoolOpt{{}},
			ResourcesAuthorize:     []*DynamicBoolOpt{{}},
			ResourcesAllocate:      []*DynamicBoolOpt{{}},
			ResourcesRelease:       []*DynamicBoolOpt{{}},
			ResourcesDerivedReply:  []*DynamicBoolOpt{{}},
			IPsAuthorize:           []*DynamicBoolOpt{{}},
			IPsAllocate:            []*DynamicBoolOpt{{}},
			IPsRelease:             []*DynamicBoolOpt{{}},
			RoutesDerivedReply:     []*DynamicBoolOpt{{}},
			StatsDerivedReply:      []*DynamicBoolOpt{{}},
			ThresholdsDerivedReply: []*DynamicBoolOpt{{}},
			MaxUsage:               []*DynamicBoolOpt{{}},
			TTL:                    []*DynamicDurationOpt{{value: SessionsTTLDftOpt}},
			Chargeable:             []*DynamicBoolOpt{{value: SessionsChargeableDftOpt}},
			TTLLastUsage:           []*DynamicDurationPointerOpt{},
			TTLLastUsed:            []*DynamicDurationPointerOpt{},
			DebitInterval:          []*DynamicDurationOpt{{value: SessionsDebitIntervalDftOpt}},
			TTLMaxDelay:            []*DynamicDurationOpt{{value: SessionsTTLMaxDelayDftOpt}},
			TTLUsage:               []*DynamicDurationPointerOpt{},
			ForceUsage:             []*DynamicBoolOpt{},
			OriginID:               []*DynamicStringOpt{},
			AccountsForceUsage:     []*DynamicBoolOpt{},
		},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.SessionSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestFsAgentConfig(t *testing.T) {
	expected := &FsAgentCfg{
		Enabled:                false,
		SessionSConns:          []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		SubscribePark:          true,
		CreateCDR:              false,
		LowBalanceAnnFile:      "",
		EmptyBalanceAnnFile:    "",
		EmptyBalanceContext:    "",
		MaxWaitConnection:      2000000000,
		ActiveSessionDelimiter: ",",
		ExtraFields:            utils.RSRParsers{},
		EventSocketConns: []*FsConnCfg{
			{
				Address:      "127.0.0.1:8021",
				Password:     "ClueCon",
				Reconnects:   5,
				Alias:        "127.0.0.1:8021",
				ReplyTimeout: time.Minute,
			},
		},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.FsAgentCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestKamAgentConfig(t *testing.T) {
	expected := &KamAgentCfg{
		Enabled:       false,
		SessionSConns: []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		CreateCdr:     false,
		EvapiConns:    []*KamConnCfg{{Address: "127.0.0.1:8448", Reconnects: 5, Alias: ""}},
		Timezone:      "",
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.KamAgentCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestAsteriskAgentConfig(t *testing.T) {
	expected := &AsteriskAgentCfg{
		Enabled:       false,
		SessionSConns: []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		CreateCDR:     false,
		AsteriskConns: []*AsteriskConnCfg{{
			Alias:           "",
			Address:         "127.0.0.1:8088",
			User:            "cgrates",
			Password:        "CGRateS.org",
			ConnectAttempts: 3,
			Reconnects:      5,
		}},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.AsteriskAgentCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestFilterSConfig(t *testing.T) {
	expected := &FilterSCfg{
		StatSConns:     []string{},
		ResourceSConns: []string{},
		AccountSConns:  []string{},
		TrendSConns:    []string{},
		RankingSConns:  []string{},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.FilterSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestLoaderConfig(t *testing.T) {
	ten := ""
	expected := LoaderSCfgs{
		{
			Enabled:        false,
			ID:             utils.MetaDefault,
			Tenant:         ten,
			LockFilePath:   ".cgr.lck",
			CacheSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
			FieldSeparator: ",",
			TpInDir:        "/var/spool/cgrates/loader/in",
			TpOutDir:       "/var/spool/cgrates/loader/out",
			Data:           nil,
			Action:         utils.MetaStore,
			Opts: &LoaderSOptsCfg{
				WithIndex: true,
			},
			Cache: map[string]*CacheParamCfg{
				utils.MetaFilters:        {Limit: -1, TTL: 5 * time.Second},
				utils.MetaAttributes:     {Limit: -1, TTL: 5 * time.Second},
				utils.MetaResources:      {Limit: -1, TTL: 5 * time.Second},
				utils.MetaIPs:            {Limit: -1, TTL: 5 * time.Second},
				utils.MetaStats:          {Limit: -1, TTL: 5 * time.Second},
				utils.MetaThresholds:     {Limit: -1, TTL: 5 * time.Second},
				utils.MetaRoutes:         {Limit: -1, TTL: 5 * time.Second},
				utils.MetaChargers:       {Limit: -1, TTL: 5 * time.Second},
				utils.MetaRateProfiles:   {Limit: -1, TTL: 5 * time.Second},
				utils.MetaActionProfiles: {Limit: -1, TTL: 5 * time.Second},
				utils.MetaAccounts:       {Limit: -1, TTL: 5 * time.Second},
				utils.MetaRankings:       {Limit: -1, TTL: 5 * time.Second},
				utils.MetaTrends:         {Limit: -1, TTL: 5 * time.Second},
			},
		},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.LoaderCfg()
	newConfig[0].Data = nil
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestAnalyzerConfig(t *testing.T) {
	expected := &AnalyzerSCfg{
		Enabled:         false,
		CleanupInterval: time.Hour,
		DBPath:          "/var/spool/cgrates/analyzers",
		IndexType:       utils.MetaScorch,
		EEsConns:        []string{},
		TTL:             24 * time.Hour,
		Opts: &AnalyzerSOpts{
			ExporterIDs: []*DynamicStringSliceOpt{},
		},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.AnalyzerSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestApierConfig(t *testing.T) {
	expected := &AdminSCfg{
		Enabled:         false,
		CachesConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
		ActionSConns:    []string{},
		AttributeSConns: []string{},
		EEsConns:        []string{},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.AdminSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestERSConfig(t *testing.T) {
	expected := &ERsCfg{
		Enabled:         false,
		SessionSConns:   []string{"*internal:*sessions"},
		EEsConns:        []string{},
		StatSConns:      []string{},
		ThresholdSConns: []string{},
		Readers: []*EventReaderCfg{
			{
				ID:                   utils.MetaDefault,
				Type:                 utils.MetaNone,
				RunDelay:             0,
				ConcurrentReqs:       1024,
				SourcePath:           "/var/spool/cgrates/ers/in",
				ProcessedPath:        "/var/spool/cgrates/ers/out",
				Tenant:               nil,
				Timezone:             utils.EmptyString,
				EEsIDs:               []string{},
				EEsSuccessIDs:        []string{},
				EEsFailedIDs:         []string{},
				Filters:              []string{},
				Flags:                utils.FlagsWithParams{},
				Fields:               nil,
				Reconnects:           -1,
				MaxReconnectInterval: 5 * time.Minute,
				CacheDumpFields:      make([]*FCTemplate, 0),
				PartialCommitFields:  make([]*FCTemplate, 0),
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(","),
					CSVHeaderDefineChar: utils.StringPointer(":"),
					CSVRowLength:        utils.IntPointer(0),
					PartialOrderField:   utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction:  utils.StringPointer(utils.MetaNone),
					NATSSubject:         utils.StringPointer("cgrates_cdrs"),
				},
			},
		},
		PartialCacheTTL: time.Second,
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.ERsCfg()
	newConfig.Readers[0].Fields = nil
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestEEsNoLksConfig(t *testing.T) {
	expected := &EEsCfg{
		Enabled:         false,
		AttributeSConns: []string{},
		Cache: map[string]*CacheParamCfg{
			utils.MetaFileCSV: {
				Limit: -1,
				TTL:   5 * time.Second,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.MetaNone,
				ExportPath:     "/var/spool/cgrates/ees",
				Attempts:       1,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				AttributeSIDs:  []string{},
				Flags:          utils.FlagsWithParams{},
				contentFields:  []*FCTemplate{},
				Fields:         []*FCTemplate{},
				headerFields:   []*FCTemplate{},
				trailerFields:  []*FCTemplate{},
				Opts:           &EventExporterOpts{},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
				EFsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEFs)},
			},
		},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.EEsNoLksCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestRateSConfig(t *testing.T) {
	expected := &RateSCfg{
		Enabled:                    false,
		IndexedSelects:             true,
		PrefixIndexedFields:        &[]string{},
		SuffixIndexedFields:        &[]string{},
		ExistsIndexedFields:        &[]string{},
		NotExistsIndexedFields:     &[]string{},
		NestedFields:               false,
		RateIndexedSelects:         true,
		RatePrefixIndexedFields:    &[]string{},
		RateSuffixIndexedFields:    &[]string{},
		RateExistsIndexedFields:    &[]string{},
		RateNotExistsIndexedFields: &[]string{},
		RateNestedFields:           false,
		Verbosity:                  1000,
		Opts: &RatesOpts{
			ProfileIDs:           []*DynamicStringSliceOpt{},
			StartTime:            []*DynamicStringOpt{{value: RatesStartTimeDftOpt}},
			Usage:                []*DynamicDecimalOpt{{value: RatesUsageDftOpt}},
			IntervalStart:        []*DynamicDecimalOpt{{value: RatesIntervalStartDftOpt}},
			ProfileIgnoreFilters: []*DynamicBoolOpt{{value: RatesProfileIgnoreFiltersDftOpt}},
		},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.RateSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestSIPAgentConfig(t *testing.T) {
	expected := &SIPAgentCfg{
		Enabled:             false,
		Listen:              "127.0.0.1:5060",
		ListenNet:           "udp",
		SessionSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		StatSConns:          []string{},
		ThresholdSConns:     []string{},
		Timezone:            "",
		RetransmissionTimer: 1000000000,
		RequestProcessors:   nil,
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.SIPAgentCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestRPCConnsConfig(t *testing.T) {
	expected := RPCConns{
		utils.MetaBiJSONLocalHost: {
			Strategy: rpcclient.PoolFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{{
				Address:   "127.0.0.1:2014",
				Transport: rpcclient.BiRPCJSON,
			}},
		},
		utils.MetaInternal: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					Address:   utils.MetaInternal,
					Transport: utils.EmptyString,
					TLS:       false,
				},
			},
		},
		rpcclient.BiRPCInternal: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					Address:   rpcclient.BiRPCInternal,
					Transport: utils.EmptyString,
					TLS:       false,
				},
			},
		},
		utils.MetaLocalHost: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					Address:   "127.0.0.1:2012",
					Transport: "*json",
					TLS:       false,
				},
			},
		},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.RPCConns()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestTemplatesConfig(t *testing.T) {
	expected := FCTemplates{
		"*err": {
			{
				Tag:       "SessionId",
				Type:      "*variable",
				Path:      "*rep.Session-Id",
				Layout:    "2006-01-02T15:04:05Z07:00",
				Value:     utils.NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:       "OriginHost",
				Type:      "*variable",
				Path:      "*rep.Origin-Host",
				Layout:    "2006-01-02T15:04:05Z07:00",
				Value:     utils.NewRSRParsersMustCompile("~*vars.OriginHost", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:       "OriginRealm",
				Type:      "*variable",
				Path:      "*rep.Origin-Realm",
				Layout:    "2006-01-02T15:04:05Z07:00",
				Value:     utils.NewRSRParsersMustCompile("~*vars.OriginRealm", utils.InfieldSep),
				Mandatory: true,
			},
		},
		"*errSip": {
			{
				Tag:       "Request",
				Type:      "*constant",
				Path:      "*rep.Request",
				Layout:    "2006-01-02T15:04:05Z07:00",
				Value:     utils.NewRSRParsersMustCompile("SIP/2.0 500 Internal Server Error", utils.InfieldSep),
				Mandatory: true,
			},
		},
		"*cca":           nil,
		"*asr":           nil,
		"*rar":           nil,
		"*fsa":           nil,
		utils.MetaCdrLog: nil,
		utils.MetaDMR:    nil,
		utils.MetaCoA:    nil,
	}
	for _, value := range expected {
		for _, elem := range value {
			elem.ComputePath()
		}
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.TemplatesCfg()
	newConfig["*cca"] = nil
	newConfig["*asr"] = nil
	newConfig["*rar"] = nil
	newConfig["*fsa"] = nil
	newConfig[utils.MetaCdrLog] = nil
	newConfig[utils.MetaDMR] = nil
	newConfig[utils.MetaCoA] = nil
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestConfigsConfig(t *testing.T) {
	expected := &ConfigSCfg{
		Enabled: false,
		URL:     "/configs/",
		RootDir: "/var/spool/cgrates/configs",
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.ConfigSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestAPIBanConfig(t *testing.T) {
	expected := &APIBanCfg{
		Enabled: false,
		Keys:    []string{},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.APIBanCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestRLockSections(t *testing.T) {
	cgrCfg := NewDefaultCGRConfig()
	cgrCfg.rLockSections()
	cgrCfg.rUnlockSections()
}

func TestLockSections(t *testing.T) {
	cgrCfg := NewDefaultCGRConfig()
	cgrCfg.lockSections()
	cgrCfg.unlockSections()
}

func TestRLockAndRUnlock(t *testing.T) {
	cgrCfg := NewDefaultCGRConfig()
	cgrCfg.RLocks("attributes", "ees", "general")
	cgrCfg.RUnlocks("attributes", "ees", "general")
}

func TestCgrLoaderCfgITDefaults(t *testing.T) {
	eCfg := LoaderSCfgs{
		{
			ID:             utils.MetaDefault,
			Enabled:        false,
			RunDelay:       0,
			LockFilePath:   ".cgr.lck",
			CacheSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
			FieldSeparator: ",",
			TpInDir:        "/var/spool/cgrates/loader/in",
			TpOutDir:       "/var/spool/cgrates/loader/out",
			Action:         utils.MetaStore,
			Opts: &LoaderSOptsCfg{
				WithIndex: true,
			},
			Cache: map[string]*CacheParamCfg{
				utils.MetaFilters:        {Limit: -1, TTL: 5 * time.Second},
				utils.MetaAttributes:     {Limit: -1, TTL: 5 * time.Second},
				utils.MetaResources:      {Limit: -1, TTL: 5 * time.Second},
				utils.MetaIPs:            {Limit: -1, TTL: 5 * time.Second},
				utils.MetaStats:          {Limit: -1, TTL: 5 * time.Second},
				utils.MetaThresholds:     {Limit: -1, TTL: 5 * time.Second},
				utils.MetaRoutes:         {Limit: -1, TTL: 5 * time.Second},
				utils.MetaChargers:       {Limit: -1, TTL: 5 * time.Second},
				utils.MetaRateProfiles:   {Limit: -1, TTL: 5 * time.Second},
				utils.MetaActionProfiles: {Limit: -1, TTL: 5 * time.Second},
				utils.MetaAccounts:       {Limit: -1, TTL: 5 * time.Second},
				utils.MetaTrends:         {Limit: -1, TTL: 5 * time.Second},
				utils.MetaRankings:       {Limit: -1, TTL: 5 * time.Second},
			},
			Data: []*LoaderDataType{
				{
					Type:     utils.MetaFilters,
					Filename: utils.FiltersCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "Type",
							Path:      "Rules.Type",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							NewBranch: true,
							Layout:    time.RFC3339},
						{Tag: "Element",
							Path:   "Rules.Element",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Values",
							Path:   "Rules.Values",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaAttributes,
					Filename: utils.AttributesCsv,
					Fields: []*FCTemplate{
						{Tag: "TenantID",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339,
						},
						{
							Tag:       "ProfileID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339,
						},
						{
							Tag:    "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "Blockers",
							Path:   "Blockers",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{
							Tag:       "AttributeFilterIDs",
							Path:      "Attributes.FilterIDs",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							NewBranch: true,
							Layout:    time.RFC3339},
						{
							Tag:    "AttributeBlockers",
							Path:   "Attributes.Blockers",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{
							Tag:    "Path",
							Path:   "Attributes.Path",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "Type",
							Path:   "Attributes.Type",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "Value",
							Path:   "Attributes.Value",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339,
						},
					},
				},
				{
					Type:     utils.MetaResources,
					Filename: utils.ResourcesCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "TTL",
							Path:   "UsageTTL",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Limit",
							Path:   "Limit",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "AllocationMessage",
							Path:   "AllocationMessage",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Blocker",
							Path:   "Blocker",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Stored",
							Path:   "Stored",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ThresholdIDs",
							Path:   "ThresholdIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaIPs,
					Filename: utils.IPsCsv,
					Fields: []*FCTemplate{
						{
							Tag:       "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339,
						},
						{
							Tag:       "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339,
						},
						{
							Tag:    "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{
							Tag:    "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{
							Tag:    "TTL",
							Path:   "TTL",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{
							Tag:    "Stored",
							Path:   "Stored",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{
							Tag:       "PoolID",
							Path:      "Pools.ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout:    time.RFC3339,
							NewBranch: true,
						},
						{
							Tag:    "PoolFilterIDs",
							Path:   "Pools.FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{
							Tag:    "PoolType",
							Path:   "Pools.Type",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{
							Tag:    "PoolRange",
							Path:   "Pools.Range",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{
							Tag:    "PoolStrategy",
							Path:   "Pools.Strategy",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{
							Tag:    "PoolMessage",
							Path:   "Pools.Message",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{
							Tag:    "PoolWeights",
							Path:   "Pools.Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{
							Tag:    "PoolBlockers",
							Path:   "Pools.Blockers",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.13", utils.InfieldSep),
							Layout: time.RFC3339,
						},
					},
				},
				{
					Type:     utils.MetaStats,
					Filename: utils.StatsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Blockers",
							Path:   "Blockers",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "QueueLength",
							Path:   "QueueLength",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "TTL",
							Path:   "TTL",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MinItems",
							Path:   "MinItems",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Stored",
							Path:   "Stored",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},

						{Tag: "ThresholdIDs",
							Path:   "ThresholdIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MetricIDs",
							Path:      "Metrics.MetricID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							NewBranch: true,
							Layout:    time.RFC3339},
						{Tag: "MetricFilterIDs",
							Path:   "Metrics.FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MetricBlockers",
							Path:   "Metrics.Blockers",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaThresholds,
					Filename: utils.ThresholdsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MaxHits",
							Path:   "MaxHits",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MinHits",
							Path:   "MinHits",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MinSleep",
							Path:   "MinSleep",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Blocker",
							Path:   "Blocker",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ActionProfileIDs",
							Path:   "ActionProfileIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Async",
							Path:   "Async",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "EeIDs",
							Path:   "EeIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaTrends,
					Filename: utils.TrendsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "Schedule",
							Path:   "Schedule",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "StatID",
							Path:   "StatID",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Metrics",
							Path:   "Metrics",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "TTL",
							Path:   "TTL",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "QueueLength",
							Path:   "QueueLength",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MinItems",
							Path:   "MinItems",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "CorrelationType",
							Path:   "CorrelationType",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Tolerance",
							Path:   "Tolerance",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Stored",
							Path:   "Stored",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ThresholdIDs",
							Path:   "ThresholdIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaRankings,
					Filename: utils.RankingsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "Schedule",
							Path:   "Schedule",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "StatIDs",
							Path:   "StatIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MetricIDs",
							Path:   "MetricIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Sorting",
							Path:   "Sorting",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "SortingParameters",
							Path:   "SortingParameters",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Stored",
							Path:   "Stored",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ThresholdIDs",
							Path:   "ThresholdIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaRoutes,
					Filename: utils.RoutesCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Blockers",
							Path:   "Blockers",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Sorting",
							Path:   "Sorting",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "SortingParameters",
							Path:   "SortingParameters",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteID",
							Path:      "Routes.ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							NewBranch: true,
							Layout:    time.RFC3339},
						{Tag: "RouteFilterIDs",
							Path:   "Routes.FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteAccountIDs",
							Path:   "Routes.AccountIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteRateProfileIDs",
							Path:   "Routes.RateProfileIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteResourceIDs",
							Path:   "Routes.ResourceIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteStatIDs",
							Path:   "Routes.StatIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteWeights",
							Path:   "Routes.Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.13", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteBlockers",
							Path:   "Routes.Blockers",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.14", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteParameters",
							Path:   "Routes.RouteParameters",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.15", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaChargers,
					Filename: utils.ChargersCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Blockers",
							Path:   "Blockers",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RunID",
							Path:   "RunID",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "AttributeIDs",
							Path:   "AttributeIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaRateProfiles,
					Filename: utils.RatesCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "MinCost",
							Path:   "MinCost",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "MaxCost",
							Path:   "MaxCost",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "MaxCostStrategy",
							Path:   "MaxCostStrategy",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "RateFilterIDs",
							Path:    "Rates[<~*req.7>].FilterIDs",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.7:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout:  time.RFC3339,
						},
						{Tag: "RateActivationTimes",
							Path:    "Rates[<~*req.7>].ActivationTimes",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.7:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout:  time.RFC3339,
						},
						{Tag: "RateWeights",
							Path:    "Rates[<~*req.7>].Weights",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.7:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout:  time.RFC3339,
						},
						{Tag: "RateBlocker",
							Path:    "Rates[<~*req.7>].Blocker",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.7:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout:  time.RFC3339,
						},
						{Tag: "RateIntervalStart",
							Path:      "Rates[<~*req.7>].IntervalRates.IntervalStart",
							Type:      utils.MetaVariable,
							Filters:   []string{"*notempty:~*req.7:"},
							Value:     utils.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
							Layout:    time.RFC3339,
							NewBranch: true,
						},
						{Tag: "RateFixedFee",
							Path:    "Rates[<~*req.7>].IntervalRates.FixedFee",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.7:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.13", utils.InfieldSep),
							Layout:  time.RFC3339,
						},
						{Tag: "RateRecurrentFee",
							Path:    "Rates[<~*req.7>].IntervalRates.RecurrentFee",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.7:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.14", utils.InfieldSep),
							Layout:  time.RFC3339,
						},
						{Tag: "RateUnit",
							Path:    "Rates[<~*req.7>].IntervalRates.Unit",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.7:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.15", utils.InfieldSep),
							Layout:  time.RFC3339,
						},
						{Tag: "RateIncrement",
							Path:    "Rates[<~*req.7>].IntervalRates.Increment",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.7:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.16", utils.InfieldSep),
							Layout:  time.RFC3339,
						},
					},
				},
				{
					Type:     utils.MetaActionProfiles,
					Filename: utils.ActionsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Blockers",
							Path:   "Blockers",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Schedule",
							Path:   "Schedule",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "TargetIDs",
							Path:   "Targets[<~*req.6>]",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ActionFilterIDs",
							Path:    "Actions[<~*req.8>].FilterIDs",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.8:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "ActionTTL",
							Path:    "Actions[<~*req.8>].TTL",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.8:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "ActionType",
							Path:    "Actions[<~*req.8>].Type",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.8:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "ActionOpts",
							Path:    "Actions[<~*req.8>].Opts",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.8:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "ActionWeights",
							Path:    "Actions[<~*req.8>].Weights",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.8:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.13", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "ActionBlockers",
							Path:    "Actions[<~*req.8>].Blockers",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.8:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.14", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "ActionDiktatsID",
							Path:      "Actions[<~*req.8>].Diktats.ID",
							Type:      utils.MetaVariable,
							Filters:   []string{"*notempty:~*req.8:"},
							Value:     utils.NewRSRParsersMustCompile("~*req.15", utils.InfieldSep),
							NewBranch: true,
							Layout:    time.RFC3339},
						{Tag: "ActionDiktatsFilterIDs",
							Path:    "Actions[<~*req.8>].Diktats.FilterIDs",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.8:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.16", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "ActionDiktatsOpts",
							Path:    "Actions[<~*req.8>].Diktats.Opts",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.8:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.17", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "ActionDiktatsWeights",
							Path:    "Actions[<~*req.8>].Diktats.Weights",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.8:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.18", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "ActionDiktatsBlockers",
							Path:    "Actions[<~*req.8>].Diktats.Blockers",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.8:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.19", utils.InfieldSep),
							Layout:  time.RFC3339},
					},
				},
				{
					Type:     utils.MetaAccounts,
					Filename: utils.AccountsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Blockers",
							Path:   "Blockers",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Opts",
							Path:   "Opts",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "BalanceFilterIDs",
							Path:    "Balances[<~*req.6>].FilterIDs",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceWeights",
							Path:    "Balances[<~*req.6>].Weights",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceBlockers",
							Path:    "Balances[<~*req.6>].Blockers",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceType",
							Path:    "Balances[<~*req.6>].Type",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceUnits",
							Path:    "Balances[<~*req.6>].Units",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceUnitFactors",
							Path:    "Balances[<~*req.6>].UnitFactors",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceOpts",
							Path:    "Balances[<~*req.6>].Opts",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.13", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceCostIncrements",
							Path:    "Balances[<~*req.6>].CostIncrements",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.14", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceAttributeIDs",
							Path:    "Balances[<~*req.6>].AttributeIDs",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.15", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceRateProfileIDs",
							Path:    "Balances[<~*req.6>].RateProfileIDs",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.16", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "ThresholdIDs",
							Path:   "ThresholdIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.17", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
			},
		},
	}
	for _, profile := range eCfg {
		for _, fields := range profile.Data {
			for _, v := range fields.Fields {
				v.ComputePath()
			}
		}
	}
	if !reflect.DeepEqual(eCfg, cgrCfg.loaderCfg) {
		t.Errorf("Expecting: %+v,\n received: %+v ",
			utils.ToJSON(eCfg), utils.ToJSON(cgrCfg.loaderCfg))
	}
}

func TestCgrLoaderCfgDefault(t *testing.T) {
	eLdrCfg := &LoaderCgrCfg{
		TpID:            "",
		DataPath:        "./",
		DisableReverse:  false,
		FieldSeparator:  utils.CSVSep,
		CachesConns:     []string{utils.MetaLocalHost},
		ActionSConns:    []string{utils.MetaLocalHost},
		GapiCredentials: json.RawMessage(`".gapi/credentials.json"`),
		GapiToken:       json.RawMessage(`".gapi/token.json"`),
	}
	if !reflect.DeepEqual(cgrCfg.LoaderCgrCfg(), eLdrCfg) {
		t.Errorf("received: %+v, expecting: %+v", utils.ToJSON(cgrCfg.LoaderCgrCfg()), utils.ToJSON(eLdrCfg))
	}
}

func TestLoadConfigDBCfgErr(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	if err := cfg.configDBCfg.Load(context.Background(), &mockDb{}, cfg); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestGetReloadChan(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	rcv := cfg.GetReloadChan()
	expected := make(chan string)
	if len(rcv) != len(expected) {
		t.Error("Channels should have the same length")
	}
}

func TestCgrMigratorCfgDefault(t *testing.T) {
	eMgrCfg := &MigratorCgrCfg{
		FromItems: map[string]*MigratorFromItem{
			utils.MetaAccounts:          {DBConn: utils.MetaDefault},
			utils.CacheChargerProfiles:  {DBConn: utils.MetaDefault},
			utils.MetaFilters:           {DBConn: utils.MetaDefault},
			utils.MetaLoadIDs:           {DBConn: utils.MetaDefault},
			utils.MetaStatQueueProfiles: {DBConn: utils.MetaDefault},
			utils.CacheVersions:         {DBConn: utils.MetaDefault},
		},
		OutDBOpts: &DBOpts{
			RedisMaxConns:           10,
			RedisConnectAttempts:    20,
			RedisClusterSync:        5 * time.Second,
			MongoQueryTimeout:       10 * time.Second,
			MongoConnScheme:         "mongodb",
			RedisPoolPipelineWindow: 150 * time.Microsecond,
		},
	}
	if !reflect.DeepEqual(cgrCfg.MigratorCgrCfg(), eMgrCfg) {
		t.Errorf("expected: %+v, received: %+v", utils.ToJSON(eMgrCfg), utils.ToJSON(cgrCfg.MigratorCgrCfg()))
	}
}

func TestCfgTlsCfg(t *testing.T) {
	jsnCfg := `
	{
	"tls":{
		"server_certificate" : "path/To/Server/Cert",
		"server_key":"path/To/Server/Key",
		"client_certificate" : "path/To/Client/Cert",
		"client_key":"path/To/Client/Key",
		"ca_certificate":"path/To/CA/Cert",
		"server_name":"TestServerName",
		"server_policy":3,
		},
	}`
	eCgrCfg := NewDefaultCGRConfig()
	eCgrCfg.tlsCfg = &TLSCfg{
		ServerCerificate: "path/To/Server/Cert",
		ServerKey:        "path/To/Server/Key",
		CaCertificate:    "path/To/CA/Cert",
		ClientCerificate: "path/To/Client/Cert",
		ClientKey:        "path/To/Client/Key",
		ServerName:       "TestServerName",
		ServerPolicy:     3,
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrCfg.TLSCfg(), cgrCfg.TLSCfg()) {
		t.Errorf("Expected: %s, received: %s",
			utils.ToJSON(eCgrCfg.tlsCfg), utils.ToJSON(cgrCfg.tlsCfg))
	}
}

func TestCgrCfgJSONDefaultAnalyzerSCfg(t *testing.T) {
	aSCfg := &AnalyzerSCfg{
		Enabled:         false,
		CleanupInterval: time.Hour,
		DBPath:          "/var/spool/cgrates/analyzers",
		IndexType:       utils.MetaScorch,
		EEsConns:        []string{},
		TTL:             24 * time.Hour,
		Opts: &AnalyzerSOpts{
			ExporterIDs: []*DynamicStringSliceOpt{},
		},
	}
	if !reflect.DeepEqual(cgrCfg.analyzerSCfg, aSCfg) {
		t.Errorf("received: %+v, expecting: %+v", utils.ToJSON(cgrCfg.analyzerSCfg), utils.ToJSON(aSCfg))
	}
}

func TestNewCGRConfigFromPathNotFound(t *testing.T) {
	fpath := path.Join("/usr", "share", "cgrates", "conf", "samples", "notValid")
	_, err := NewCGRConfigFromPath(context.Background(), fpath)
	if err == nil || err.Error() != utils.ErrPathNotReachable(fpath).Error() {
		t.Fatalf("Expected %s ,received %s", utils.ErrPathNotReachable(fpath), err)
	}
	fpath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo", "cgrates.json")
	cfg, err := NewCGRConfigFromPath(context.Background(), fpath)
	if err == nil {
		t.Fatalf("Expected error,received %v", cfg)
	}
	fpath = "https://not_a_reacheble_website"
	_, err = NewCGRConfigFromPath(context.Background(), fpath)
	if err == nil || err.Error() != utils.ErrPathNotReachable(fpath).Error() {
		t.Fatalf("Expected %s ,received %s", utils.ErrPathNotReachable(fpath), err)
	}
	cfg, err = NewCGRConfigFromPath(context.Background(), "https://github.com/")
	if err == nil {
		t.Fatalf("Expected error,received %v", cfg)
	}
}

func TestCgrCfgJSONDefaultApierCfg(t *testing.T) {
	aCfg := &AdminSCfg{
		Enabled:         false,
		CachesConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
		ActionSConns:    []string{},
		AttributeSConns: []string{},
		EEsConns:        []string{},
	}
	if !reflect.DeepEqual(cgrCfg.admS, aCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.admS, aCfg)
	}
}

func TestCgrCfgJSONDefaultRateCfg(t *testing.T) {
	eCfg := &RateSCfg{
		Enabled:                    false,
		IndexedSelects:             true,
		StringIndexedFields:        nil,
		PrefixIndexedFields:        &[]string{},
		SuffixIndexedFields:        &[]string{},
		ExistsIndexedFields:        &[]string{},
		NotExistsIndexedFields:     &[]string{},
		NestedFields:               false,
		RateIndexedSelects:         true,
		RateStringIndexedFields:    nil,
		RatePrefixIndexedFields:    &[]string{},
		RateSuffixIndexedFields:    &[]string{},
		RateExistsIndexedFields:    &[]string{},
		RateNotExistsIndexedFields: &[]string{},
		RateNestedFields:           false,
		Verbosity:                  1000,
		Opts: &RatesOpts{
			ProfileIDs:           []*DynamicStringSliceOpt{},
			StartTime:            []*DynamicStringOpt{{value: RatesStartTimeDftOpt}},
			Usage:                []*DynamicDecimalOpt{{value: RatesUsageDftOpt}},
			IntervalStart:        []*DynamicDecimalOpt{{value: RatesIntervalStartDftOpt}},
			ProfileIgnoreFilters: []*DynamicBoolOpt{{value: RatesProfileIgnoreFiltersDftOpt}},
		},
	}
	if !reflect.DeepEqual(cgrCfg.rateSCfg, eCfg) {
		t.Errorf("received: %+v, expecting: %+v", utils.ToJSON(cgrCfg.rateSCfg), utils.ToJSON(eCfg))
	}
}

func TestCgrCfgV1GetConfigAllConfig(t *testing.T) {
	var rcv map[string]any
	cgrCfg := NewDefaultCGRConfig()
	expected := cgrCfg.AsMapInterface()
	if err := cgrCfg.V1GetConfig(context.Background(), &SectionWithAPIOpts{}, &rcv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
	if err := cgrCfg.V1GetConfig(context.Background(), &SectionWithAPIOpts{}, &rcv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestCgrCfgV1GetConfigSectionLoader(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		LoaderSJSON: []map[string]any{
			{
				utils.IDCfg:           "*default",
				utils.EnabledCfg:      false,
				utils.TenantCfg:       utils.EmptyString,
				utils.RunDelayCfg:     "0",
				utils.LockFilePathCfg: ".cgr.lck",
				utils.CachesConnsCfg:  []string{utils.MetaInternal},
				utils.FieldSepCfg:     ",",
				utils.TpInDirCfg:      "/var/spool/cgrates/loader/in",
				utils.TpOutDirCfg:     "/var/spool/cgrates/loader/out",
				utils.DataCfg:         []map[string]any{},
				utils.ActionCfg:       utils.MetaStore,
				utils.OptsCfg: map[string]any{
					utils.MetaCache:       "",
					utils.MetaWithIndex:   true,
					utils.MetaForceLock:   false,
					utils.MetaStopOnError: false,
				},
				utils.CacheCfg: map[string]any{
					utils.MetaFilters:        map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
					utils.MetaAttributes:     map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
					utils.MetaResources:      map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
					utils.MetaIPs:            map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
					utils.MetaStats:          map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
					utils.MetaThresholds:     map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
					utils.MetaRoutes:         map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
					utils.MetaChargers:       map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
					utils.MetaRateProfiles:   map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
					utils.MetaActionProfiles: map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
					utils.MetaAccounts:       map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
					utils.MetaTrends:         map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
					utils.MetaRankings:       map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
				},
			},
		},
	}
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{LoaderSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if mp, can := reply[LoaderSJSON].([]map[string]any); !can {
		t.Errorf("Unexpected type: %t", reply[LoaderSJSON])
	} else {
		mp[0][utils.DataCfg] = []map[string]any{}
		if !reflect.DeepEqual(expected[LoaderSJSON], mp) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected[LoaderSJSON]), utils.ToJSON(mp))
		}
	}
}

func TestCgrCfgV1GetConfigSectionHTTPAgent(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		HTTPAgentJSON: []map[string]any{},
	}
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{HTTPAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestCgrCfgV1GetConfigSectionCoreS(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		CoreSJSON: map[string]any{
			utils.CapsCfg:              0,
			utils.CapsStrategyCfg:      utils.MetaBusy,
			utils.CapsStatsIntervalCfg: "0",
			utils.EEsConnsCfg:          []string{},
			utils.ShutdownTimeoutCfg:   "1s",
		},
	}
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{CoreSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestCgrCfgV1GetConfigListen(t *testing.T) {
	jsnCfg := `
{
"listen": {
	"rpc_json": ":2012",
	"rpc_gob": ":2013",
	"http": ":2080",
	}
}`
	expected := map[string]any{
		"listen": map[string]any{
			"http":         ":2080",
			"http_tls":     "127.0.0.1:2280",
			"rpc_gob":      ":2013",
			"rpc_gob_tls":  "127.0.0.1:2023",
			"rpc_json":     ":2012",
			"rpc_json_tls": "127.0.0.1:2022",
		},
	}
	var rcv map[string]any
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if err := cgrCfg.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{ListenJSON}}, &rcv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestV1GetConfigGeneral(t *testing.T) {
	var reply map[string]any
	cfgJSONStr := `{
      "general": {
            "node_id": "ENGINE1",
            "locking_timeout": "0",
            "failed_posts_ttl": "0s",
			"max_reconnect_interval": "5m",
            "connect_timeout": "0s",
            "reply_timeout": "0s",
        }
}`
	expected := map[string]any{
		utils.NodeIDCfg:               "ENGINE1",
		utils.RoundingDecimalsCfg:     5,
		utils.DBDataEncodingCfg:       "*msgpack",
		utils.TpExportPathCfg:         "/var/spool/cgrates/tpe",
		utils.DefaultReqTypeCfg:       "*rated",
		utils.DefaultCategoryCfg:      "call",
		utils.DefaultTenantCfg:        "cgrates.org",
		utils.DefaultTimezoneCfg:      "Local",
		utils.DefaultCachingCfg:       "*reload",
		utils.CachingDlayCfg:          "0",
		utils.ConnectAttemptsCfg:      5,
		utils.ReconnectsCfg:           -1,
		utils.MaxReconnectIntervalCfg: "5m0s",
		utils.ConnectTimeoutCfg:       "0",
		utils.ReplyTimeoutCfg:         "0",
		utils.LockingTimeoutCfg:       "0",
		utils.DigestSeparatorCfg:      ",",
		utils.DigestEqualCfg:          ":",
		utils.MaxParallelConnsCfg:     100,
		utils.DecimalMaxScaleCfg:      0,
		utils.DecimalMinScaleCfg:      0,
		utils.DecimalPrecisionCfg:     0,
		utils.DecimalRoundingModeCfg:  "*toNearestEven",
		utils.OptsCfg: map[string]any{
			utils.MetaExporterIDs: []*DynamicStringSliceOpt{},
		},
	}
	expected = map[string]any{
		GeneralJSON: expected,
	}
	cfgCgr, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr)
	if err != nil {
		t.Error(err)
	}
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{GeneralJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigDataDB(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		utils.DataDbConnsCfg: map[string]map[string]any{
			utils.MetaDefault: {
				utils.DataDbTypeCfg:           utils.MetaInternal,
				utils.DataDbHostCfg:           utils.EmptyString,
				utils.DataDbPortCfg:           0,
				utils.DataDbNameCfg:           utils.EmptyString,
				utils.DataDbUserCfg:           utils.EmptyString,
				utils.DataDbPassCfg:           utils.EmptyString,
				utils.StringIndexedFieldsCfg:  []string{},
				utils.PrefixIndexedFieldsCfg:  []string{},
				utils.ReplicationFilteredCfg:  false,
				utils.RemoteConnIDCfg:         utils.EmptyString,
				utils.ReplicationCache:        utils.EmptyString,
				utils.ReplicationFailedDirCfg: utils.EmptyString,
				utils.ReplicationIntervalCfg:  "0s",
				utils.RemoteConnsCfg:          []string{},
				utils.ReplicationConnsCfg:     []string{},
				utils.OptsCfg: map[string]any{
					"internalDBBackupPath":      "/var/lib/cgrates/internal_db/backup/db",
					"internalDBDumpInterval":    "1m0s",
					"internalDBDumpPath":        "/var/lib/cgrates/internal_db/db",
					"internalDBFileSizeLimit":   int64(1073741824),
					"internalDBRewriteInterval": "1h0m0s",
					"internalDBStartTimeout":    "5m0s",
					"mongoConnScheme":           "mongodb",
					"mongoQueryTimeout":         "10s",
					"mysqlDSNParams":            map[string]string{},
					"mysqlLocation":             "Local",
					"pgSSLMode":                 "disable",
					"redisCACertificate":        "",
					"redisClientCertificate":    "",
					"redisClientKey":            "",
					"redisCluster":              false,
					"redisClusterOndownDelay":   "0s",
					"redisClusterSync":          "5s",
					"redisConnectAttempts":      20,
					"redisConnectTimeout":       "0s",
					"redisMaxConns":             10,
					"redisPoolPipelineLimit":    0,
					"redisPoolPipelineWindow":   "150µs",
					"redisReadTimeout":          "0s", "redisSentinel": "",
					"redisTLS": false, "redisWriteTimeout": "0s",
					"sqlConnMaxLifetime": "0s", "sqlLogLevel": 3,
					"sqlMaxIdleConns": 10, "sqlMaxOpenConns": 100},
			},
		},
		utils.ItemsCfg: map[string]any{},
	}
	expected = map[string]any{
		DBJSON: expected,
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{DBJSON}}, &reply); err != nil {
		t.Error(err)
	} else if mp, can := reply[DBJSON].(map[string]any); !can {
		t.Errorf("Unexpected type: %t", reply[DBJSON])
	} else {
		mp[utils.ItemsCfg] = map[string]any{}
		if diff := cmp.Diff(expected, reply); diff != "" {
			t.Errorf("Config mismatch (-expected +got):\n%s", diff)
		}
	}
}

func TestV1GetConfigTLS(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		TlsJSON: map[string]any{
			utils.ServerCerificateCfg: "",
			utils.ServerKeyCfg:        "",
			utils.ServerPolicyCfg:     4,
			utils.ServerNameCfg:       "",
			utils.ClientCerificateCfg: "",
			utils.ClientKeyCfg:        "",
			utils.CaCertificateCfg:    "",
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{TlsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigCache(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		CacheJSON: map[string]any{
			utils.PartitionsCfg:       map[string]any{},
			utils.ReplicationConnsCfg: []string{},
			utils.RemoteConnsCfg:      []string{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{CacheJSON}}, &reply); err != nil {
		t.Error(err)
	} else if mp, can := reply[CacheJSON].(map[string]any); !can {
		t.Errorf("Unexpected type: %t", reply[CacheJSON])
	} else {
		mp[utils.PartitionsCfg] = map[string]any{}
		if !reflect.DeepEqual(reply, expected) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
		}
	}
}

func TestV1GetConfigHTTP(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		HTTPJSON: map[string]any{
			utils.HTTPJsonRPCURLCfg:        "/jsonrpc",
			utils.RegistrarSURLCfg:         "/registrar",
			utils.HTTPWSURLCfg:             "/ws",
			utils.HTTPFreeswitchCDRsURLCfg: "/freeswitch_json",
			utils.HTTPCDRsURLCfg:           "/cdr_http",
			utils.PprofPathCfg:             "/debug/pprof/",
			utils.HTTPUseBasicAuthCfg:      false,
			utils.HTTPAuthUsersCfg:         map[string]string{},
			utils.HTTPClientOptsCfg: map[string]any{
				utils.HTTPClientSkipTLSVerificationCfg:   false,
				utils.HTTPClientTLSHandshakeTimeoutCfg:   "10s",
				utils.HTTPClientDisableKeepAlivesCfg:     false,
				utils.HTTPClientDisableCompressionCfg:    false,
				utils.HTTPClientMaxIdleConnsCfg:          100,
				utils.HTTPClientMaxIdleConnsPerHostCfg:   2,
				utils.HTTPClientMaxConnsPerHostCfg:       0,
				utils.HTTPClientIdleConnTimeoutCfg:       "1m30s",
				utils.HTTPClientResponseHeaderTimeoutCfg: "0s",
				utils.HTTPClientExpectContinueTimeoutCfg: "0s",
				utils.HTTPClientForceAttemptHTTP2Cfg:     true,
				utils.HTTPClientDialTimeoutCfg:           "30s",
				utils.HTTPClientDialFallbackDelayCfg:     "300ms",
				utils.HTTPClientDialKeepAliveCfg:         "30s",
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{HTTPJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigFilterS(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		FilterSJSON: map[string]any{
			utils.StatSConnsCfg:     []string{},
			utils.ResourceSConnsCfg: []string{},
			utils.AccountSConnsCfg:  []string{},
			utils.TrendSConnsCfg:    []string{},
			utils.RankingSConnsCfg:  []string{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{FilterSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigCdrs(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		CDRsJSON: map[string]any{
			utils.EnabledCfg:          false,
			utils.ExtraFieldsCfg:      []string{},
			utils.SessionCostRetires:  5,
			utils.ChargerSConnsCfg:    []string{},
			utils.AttributeSConnsCfg:  []string{},
			utils.ThresholdSConnsCfg:  []string{},
			utils.StatSConnsCfg:       []string{},
			utils.OnlineCDRExportsCfg: []string(nil),
			utils.ActionSConnsCfg:     []string{},
			utils.EEsConnsCfg:         []string{},
			utils.RateSConnsCfg:       []string{},
			utils.AccountSConnsCfg:    []string{},
			utils.OptsCfg: map[string]any{
				utils.MetaAccounts:   []*DynamicBoolOpt{{}},
				utils.MetaAttributes: []*DynamicBoolOpt{{}},
				utils.MetaChargers:   []*DynamicBoolOpt{{}},
				utils.MetaEEs:        []*DynamicBoolOpt{{}},
				utils.MetaRates:      []*DynamicBoolOpt{{}},
				utils.MetaStats:      []*DynamicBoolOpt{{}},
				utils.MetaThresholds: []*DynamicBoolOpt{{}},
				utils.MetaRefund:     []*DynamicBoolOpt{{}},
				utils.MetaRerate:     []*DynamicBoolOpt{{}},
				utils.MetaStore:      []*DynamicBoolOpt{{value: true}},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{CDRsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSessionS(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		SessionSJSON: map[string]any{
			utils.EnabledCfg:             false,
			utils.ListenBijsonCfg:        "127.0.0.1:2014",
			utils.ListenBigobCfg:         "",
			utils.ChargerSConnsCfg:       []string{},
			utils.CDRsConnsCfg:           []string{},
			utils.ResourceSConnsCfg:      []string{},
			utils.IPsConnsCfg:            []string{},
			utils.ThresholdSConnsCfg:     []string{},
			utils.StatSConnsCfg:          []string{},
			utils.RouteSConnsCfg:         []string{},
			utils.AttributeSConnsCfg:     []string{},
			utils.ActionSConnsCfg:        []string{},
			utils.RateSConnsCfg:          []string{},
			utils.AccountSConnsCfg:       []string{},
			utils.ReplicationConnsCfg:    []string{},
			utils.StoreSCostsCfg:         false,
			utils.SessionIndexesCfg:      []string{},
			utils.ClientProtocolCfg:      1.0,
			utils.ChannelSyncIntervalCfg: "0",
			utils.TerminateAttemptsCfg:   5,
			utils.MinDurLowBalanceCfg:    "0",
			utils.AlterableFieldsCfg:     []string{},
			utils.STIRCfg: map[string]any{
				utils.AllowedAtestCfg:       []string{"*any"},
				utils.PayloadMaxdurationCfg: "-1",
				utils.DefaultAttestCfg:      "A",
				utils.PublicKeyPathCfg:      "",
				utils.PrivateKeyPathCfg:     "",
			},
			utils.DefaultUsageCfg: map[string]string{
				utils.MetaAny:   "3h0m0s",
				utils.MetaVoice: "3h0m0s",
				utils.MetaData:  "1048576",
				utils.MetaSMS:   "1",
			},
			utils.OptsCfg: map[string]any{
				utils.MetaAccounts:                  []*DynamicBoolOpt{{}},
				utils.MetaAttributes:                []*DynamicBoolOpt{{}},
				utils.MetaCDRs:                      []*DynamicBoolOpt{{}},
				utils.MetaChargers:                  []*DynamicBoolOpt{{}},
				utils.MetaResources:                 []*DynamicBoolOpt{{}},
				utils.MetaIPs:                       []*DynamicBoolOpt{{}},
				utils.MetaRoutes:                    []*DynamicBoolOpt{{}},
				utils.MetaStats:                     []*DynamicBoolOpt{{}},
				utils.MetaThresholds:                []*DynamicBoolOpt{{}},
				utils.MetaInitiate:                  []*DynamicBoolOpt{{}},
				utils.MetaUpdate:                    []*DynamicBoolOpt{{}},
				utils.MetaTerminate:                 []*DynamicBoolOpt{{}},
				utils.MetaMessage:                   []*DynamicBoolOpt{{}},
				utils.MetaAttributesDerivedReplyCfg: []*DynamicBoolOpt{{}},
				utils.MetaBlockerErrorCfg:           []*DynamicBoolOpt{{}},
				utils.MetaCDRsDerivedReplyCfg:       []*DynamicBoolOpt{{}},
				utils.MetaResourcesAuthorizeCfg:     []*DynamicBoolOpt{{}},
				utils.MetaResourcesAllocateCfg:      []*DynamicBoolOpt{{}},
				utils.MetaResourcesReleaseCfg:       []*DynamicBoolOpt{{}},
				utils.MetaResourcesDerivedReplyCfg:  []*DynamicBoolOpt{{}},
				utils.MetaIPsAuthorizeCfg:           []*DynamicBoolOpt{{}},
				utils.MetaIPsAllocateCfg:            []*DynamicBoolOpt{{}},
				utils.MetaIPsReleaseCfg:             []*DynamicBoolOpt{{}},
				utils.MetaRoutesDerivedReplyCfg:     []*DynamicBoolOpt{{}},
				utils.MetaStatsDerivedReplyCfg:      []*DynamicBoolOpt{{}},
				utils.MetaThresholdsDerivedReplyCfg: []*DynamicBoolOpt{{}},
				utils.MetaMaxUsageCfg:               []*DynamicBoolOpt{{}},
				utils.MetaTTLCfg:                    []*DynamicDurationOpt{{value: SessionsTTLDftOpt}},
				utils.MetaChargeableCfg:             []*DynamicBoolOpt{{value: SessionsChargeableDftOpt}},
				utils.MetaDebitIntervalCfg:          []*DynamicDurationOpt{{value: SessionsDebitIntervalDftOpt}},
				utils.MetaTTLLastUsageCfg:           []*DynamicDurationPointerOpt{},
				utils.MetaTTLLastUsedCfg:            []*DynamicDurationPointerOpt{},
				utils.MetaTTLMaxDelayCfg:            []*DynamicDurationOpt{{value: SessionsTTLMaxDelayDftOpt}},
				utils.MetaTTLUsageCfg:               []*DynamicDurationPointerOpt{},
				utils.MetaForceUsageCfg:             []*DynamicBoolOpt{},
				utils.MetaOriginID:                  []*DynamicStringOpt{},
				utils.MetaAccountsForceUsage:        []*DynamicBoolOpt{},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{SessionSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigFsAgent(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		FreeSWITCHAgentJSON: map[string]any{
			utils.EnabledCfg:                false,
			utils.SessionSConnsCfg:          []string{rpcclient.BiRPCInternal},
			utils.SubscribeParkCfg:          true,
			utils.CreateCdrCfg:              false,
			utils.ExtraFieldsCfg:            []string{},
			utils.LowBalanceAnnFileCfg:      "",
			utils.EmptyBalanceContextCfg:    "",
			utils.EmptyBalanceAnnFileCfg:    "",
			utils.MaxWaitConnectionCfg:      "2s",
			utils.ActiveSessionDelimiterCfg: ",",
			utils.EventSocketConnsCfg: []map[string]any{
				{
					utils.AddressCfg:              "127.0.0.1:8021",
					utils.Password:                "ClueCon",
					utils.ReconnectsCfg:           5,
					utils.MaxReconnectIntervalCfg: "0s",
					utils.ReplyTimeoutCfg:         "1m0s",
					utils.AliasCfg:                "127.0.0.1:8021",
				},
			},
			utils.RequestProcessorsCfg: []map[string]any{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{FreeSWITCHAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigKamailioAgent(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		KamailioAgentJSON: map[string]any{
			utils.EnabledCfg:       false,
			utils.SessionSConnsCfg: []string{rpcclient.BiRPCInternal},
			utils.CreateCdrCfg:     false,
			utils.TimezoneCfg:      "",
			utils.EvapiConnsCfg: []map[string]any{
				{
					utils.AddressCfg:              "127.0.0.1:8448",
					utils.ReconnectsCfg:           5,
					utils.MaxReconnectIntervalCfg: "0s",
					utils.AliasCfg:                "",
				},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{KamailioAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigAsteriskAgent(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		AsteriskAgentJSON: map[string]any{
			utils.EnabledCfg:       false,
			utils.SessionSConnsCfg: []string{rpcclient.BiRPCInternal},
			utils.CreateCdrCfg:     false,
			utils.AsteriskConnsCfg: []map[string]any{
				{
					utils.AliasCfg:                "",
					utils.AddressCfg:              "127.0.0.1:8088",
					utils.UserCf:                  "cgrates",
					utils.Password:                "CGRateS.org",
					utils.ConnectAttemptsCfg:      3,
					utils.ReconnectsCfg:           5,
					utils.MaxReconnectIntervalCfg: "0s",
				},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{AsteriskAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigDiameterAgent(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		DiameterAgentJSON: map[string]any{
			utils.ASRTemplateCfg:             "",
			utils.DictionariesPathCfg:        "/usr/share/cgrates/diameter/dict/",
			utils.EnabledCfg:                 false,
			utils.ForcedDisconnectCfg:        "*none",
			utils.ListenCfg:                  "127.0.0.1:3868",
			utils.ListenNetCfg:               "tcp",
			utils.OriginHostCfg:              "CGR-DA",
			utils.OriginRealmCfg:             "cgrates.org",
			utils.ProductNameCfg:             "CGRateS",
			utils.RARTemplateCfg:             "",
			utils.SessionSConnsCfg:           []string{rpcclient.BiRPCInternal},
			utils.StatSConnsCfg:              []string{},
			utils.ThresholdSConnsCfg:         []string{},
			utils.SyncedConnReqsCfg:          false,
			utils.VendorIDCfg:                0,
			utils.ConnHealthCheckIntervalCfg: "0s",
			utils.RequestProcessorsCfg:       []map[string]any{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{DiameterAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigRadiusAgent(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		RadiusAgentJSON: map[string]any{
			utils.EnabledCfg: false,
			utils.ListenersCfg: []map[string]any{
				{
					utils.NetworkCfg:  utils.UDP,
					utils.AuthAddrCfg: "127.0.0.1:1812",
					utils.AcctAddrCfg: "127.0.0.1:1813",
				},
			},
			utils.ClientSecretsCfg: map[string]string{
				utils.MetaDefault: "CGRateS.org",
			},
			utils.ClientDictionariesCfg: map[string][]string{
				utils.MetaDefault: {"/usr/share/cgrates/radius/dict/"},
			},
			utils.ClientDaAddressesCfg: map[string]any{},
			utils.SessionSConnsCfg:     []string{"*internal"},
			utils.StatSConnsCfg:        []string{},
			utils.ThresholdSConnsCfg:   []string{},
			utils.DMRTemplateCfg:       utils.MetaDMR,
			utils.CoATemplateCfg:       utils.MetaCoA,
			utils.RequestsCacheKeyCfg:  "",
			utils.RequestProcessorsCfg: []map[string]any{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{RadiusAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigDNSAgent(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		DNSAgentJSON: map[string]any{
			utils.EnabledCfg: false,
			utils.ListenersCfg: []map[string]any{
				{
					utils.AddressCfg: "127.0.0.1:53",
					utils.NetworkCfg: "udp",
				},
			},
			utils.SessionSConnsCfg:     []string{utils.MetaInternal},
			utils.StatSConnsCfg:        []string{},
			utils.ThresholdSConnsCfg:   []string{},
			utils.TimezoneCfg:          "",
			utils.RequestProcessorsCfg: []map[string]any{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{DNSAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigAttribute(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		AttributeSJSON: map[string]any{
			utils.EnabledCfg:                false,
			utils.StatSConnsCfg:             []string{},
			utils.ResourceSConnsCfg:         []string{},
			utils.AccountSConnsCfg:          []string{},
			utils.IndexedSelectsCfg:         true,
			utils.PrefixIndexedFieldsCfg:    []string{},
			utils.SuffixIndexedFieldsCfg:    []string{},
			utils.ExistsIndexedFieldsCfg:    []string{},
			utils.NotExistsIndexedFieldsCfg: []string{},
			utils.NestedFieldsCfg:           false,
			utils.OptsCfg: map[string]any{
				utils.MetaProfileIDs:           []*DynamicStringSliceOpt{},
				utils.MetaProcessRunsCfg:       []*DynamicIntOpt{{value: AttributesProcessRunsDftOpt}},
				utils.MetaProfileRunsCfg:       []*DynamicIntOpt{{value: AttributesProfileRunsDftOpt}},
				utils.MetaProfileIgnoreFilters: []*DynamicBoolOpt{{value: AttributesProfileIgnoreFiltersDftOpt}},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{AttributeSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigChargers(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		ChargerSJSON: map[string]any{
			utils.EnabledCfg:                false,
			utils.AttributeSConnsCfg:        []string{},
			utils.IndexedSelectsCfg:         true,
			utils.PrefixIndexedFieldsCfg:    []string{},
			utils.NestedFieldsCfg:           false,
			utils.SuffixIndexedFieldsCfg:    []string{},
			utils.ExistsIndexedFieldsCfg:    []string{},
			utils.NotExistsIndexedFieldsCfg: []string{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{ChargerSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigResourceS(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		ResourceSJSON: map[string]any{
			utils.EnabledCfg:                false,
			utils.StoreIntervalCfg:          utils.EmptyString,
			utils.ThresholdSConnsCfg:        []string{},
			utils.IndexedSelectsCfg:         true,
			utils.PrefixIndexedFieldsCfg:    []string{},
			utils.SuffixIndexedFieldsCfg:    []string{},
			utils.ExistsIndexedFieldsCfg:    []string{},
			utils.NotExistsIndexedFieldsCfg: []string{},
			utils.NestedFieldsCfg:           false,
			utils.OptsCfg: map[string]any{
				utils.MetaUsageIDCfg:  []*DynamicStringOpt{{value: ResourcesUsageIDDftOpt}},
				utils.MetaUsageTTLCfg: []*DynamicDurationOpt{{value: ResourcesUsageTTLDftOpt}},
				utils.MetaUnitsCfg:    []*DynamicFloat64Opt{{value: ResourcesUnitsDftOpt}},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{ResourceSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigStats(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		StatSJSON: map[string]any{
			utils.EnabledCfg:                false,
			utils.StoreIntervalCfg:          utils.EmptyString,
			utils.StoreUncompressedLimitCfg: 0,
			utils.ThresholdSConnsCfg:        []string{},
			utils.IndexedSelectsCfg:         true,
			utils.PrefixIndexedFieldsCfg:    []string{},
			utils.SuffixIndexedFieldsCfg:    []string{},
			utils.ExistsIndexedFieldsCfg:    []string{},
			utils.NotExistsIndexedFieldsCfg: []string{},
			utils.NestedFieldsCfg:           false,
			utils.OptsCfg: map[string]any{
				utils.MetaProfileIDs:           []*DynamicStringSliceOpt{},
				utils.MetaProfileIgnoreFilters: []*DynamicBoolOpt{{value: StatsProfileIgnoreFilters}},
				utils.OptsRoundingDecimals:     []*DynamicIntOpt{},
			},
			utils.EEsConnsCfg:       []string{},
			utils.EEsExporterIDsCfg: []string(nil),
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{StatSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigThresholds(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		ThresholdSJSON: map[string]any{
			utils.EnabledCfg:                false,
			utils.StoreIntervalCfg:          utils.EmptyString,
			utils.IndexedSelectsCfg:         true,
			utils.PrefixIndexedFieldsCfg:    []string{},
			utils.SuffixIndexedFieldsCfg:    []string{},
			utils.ExistsIndexedFieldsCfg:    []string{},
			utils.NotExistsIndexedFieldsCfg: []string{},
			utils.NestedFieldsCfg:           false,
			utils.ActionSConnsCfg:           []string{},
			utils.OptsCfg: map[string]any{
				utils.MetaProfileIDs:           []*DynamicStringSliceOpt{},
				utils.MetaProfileIgnoreFilters: []*DynamicBoolOpt{{value: ThresholdsProfileIgnoreFiltersDftOpt}},
			},
			utils.EEsConnsCfg:       []string{},
			utils.EEsExporterIDsCfg: []string{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{ThresholdSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigAcounts(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		AccountSJSON: map[string]any{
			utils.EnabledCfg:                false,
			utils.IndexedSelectsCfg:         true,
			utils.AttributeSConnsCfg:        []string{},
			utils.RateSConnsCfg:             []string{},
			utils.ThresholdSConnsCfg:        []string{},
			utils.PrefixIndexedFieldsCfg:    []string{},
			utils.SuffixIndexedFieldsCfg:    []string{},
			utils.ExistsIndexedFieldsCfg:    []string{},
			utils.NotExistsIndexedFieldsCfg: []string{},
			utils.NestedFieldsCfg:           false,
			utils.MaxIterations:             1000,
			utils.MaxUsage:                  "259200000000000", // 72h in ns
			utils.OptsCfg: map[string]any{
				utils.MetaProfileIDs:           []*DynamicStringSliceOpt{},
				utils.MetaUsage:                []*DynamicDecimalOpt{{value: AccountsUsageDftOpt}},
				utils.MetaProfileIgnoreFilters: []*DynamicBoolOpt{{value: AccountsProfileIgnoreFiltersDftOpt}},
			},
		},
	}
	cfg := NewDefaultCGRConfig()
	if err := cfg.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{AccountSJSON}}, &reply); err != nil {
		t.Error(expected)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigRoutes(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		RouteSJSON: map[string]any{
			utils.EnabledCfg:                false,
			utils.IndexedSelectsCfg:         true,
			utils.PrefixIndexedFieldsCfg:    []string{},
			utils.SuffixIndexedFieldsCfg:    []string{},
			utils.ExistsIndexedFieldsCfg:    []string{},
			utils.NotExistsIndexedFieldsCfg: []string{},
			utils.NestedFieldsCfg:           false,
			utils.AttributeSConnsCfg:        []string{},
			utils.ResourceSConnsCfg:         []string{},
			utils.StatSConnsCfg:             []string{},
			utils.RateSConnsCfg:             []string{},
			utils.AccountSConnsCfg:          []string{},
			utils.DefaultRatioCfg:           1,
			utils.OptsCfg: map[string]any{
				utils.OptsContext:         []*DynamicStringOpt{{value: RoutesContextDftOpt}},
				utils.MetaProfileCountCfg: []*DynamicIntPointerOpt{{value: RoutesProfileCountDftOpt}},
				utils.MetaIgnoreErrorsCfg: []*DynamicBoolOpt{{value: RoutesIgnoreErrorsDftOpt}},
				utils.MetaMaxCostCfg:      []*DynamicInterfaceOpt{{Value: RoutesMaxCostDftOpt}},
				utils.MetaLimitCfg:        []*DynamicIntPointerOpt{},
				utils.MetaOffsetCfg:       []*DynamicIntPointerOpt{},
				utils.MetaMaxItemsCfg:     []*DynamicIntPointerOpt{},
				utils.MetaUsage:           []*DynamicDecimalOpt{{value: RoutesUsageDftOpt}},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{RouteSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSuretax(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		SureTaxJSON: map[string]any{
			utils.URLCfg:                  utils.EmptyString,
			utils.ClientNumberCfg:         utils.EmptyString,
			utils.ValidationKeyCfg:        utils.EmptyString,
			utils.BusinessUnitCfg:         utils.EmptyString,
			utils.TimezoneCfg:             "UTC",
			utils.IncludeLocalCostCfg:     false,
			utils.ReturnFileCodeCfg:       "0",
			utils.ResponseGroupCfg:        "03",
			utils.ResponseTypeCfg:         "D4",
			utils.RegulatoryCodeCfg:       "03",
			utils.ClientTrackingCfg:       "~*opts.*originID",
			utils.CustomerNumberCfg:       "~*req.Subject",
			utils.OrigNumberCfg:           "~*req.Subject",
			utils.TermNumberCfg:           "~*req.Destination",
			utils.BillToNumberCfg:         utils.EmptyString,
			utils.ZipcodeCfg:              utils.EmptyString,
			utils.Plus4Cfg:                utils.EmptyString,
			utils.P2PZipcodeCfg:           utils.EmptyString,
			utils.P2PPlus4Cfg:             utils.EmptyString,
			utils.UnitsCfg:                "1",
			utils.UnitTypeCfg:             "00",
			utils.TaxIncludedCfg:          "0",
			utils.TaxSitusRuleCfg:         "04",
			utils.TransTypeCodeCfg:        "010101",
			utils.SalesTypeCodeCfg:        "R",
			utils.TaxExemptionCodeListCfg: utils.EmptyString,
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	cfgCgr.SureTaxCfg().Timezone = time.UTC
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{SureTaxJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionLoader(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		LoaderJSON: map[string]any{
			utils.TpIDCfg:            "",
			utils.DataPathCfg:        "./",
			utils.DisableReverseCfg:  false,
			utils.FieldSepCfg:        ",",
			utils.CachesConnsCfg:     []string{"*localhost"},
			utils.ActionSConnsCfg:    []string{"*localhost"},
			utils.GapiCredentialsCfg: json.RawMessage(`".gapi/credentials.json"`),
			utils.GapiTokenCfg:       json.RawMessage(`".gapi/token.json"`),
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{LoaderJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionMigrator(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		MigratorJSON: map[string]any{
			utils.FromItemsCfg: map[string]any{
				utils.MetaAccounts:          map[string]any{utils.DBConnCfg: utils.MetaDefault},
				utils.CacheChargerProfiles:  map[string]any{utils.DBConnCfg: utils.MetaDefault},
				utils.MetaFilters:           map[string]any{utils.DBConnCfg: utils.MetaDefault},
				utils.MetaLoadIDs:           map[string]any{utils.DBConnCfg: utils.MetaDefault},
				utils.MetaStatQueueProfiles: map[string]any{utils.DBConnCfg: utils.MetaDefault},
				utils.CacheVersions:         map[string]any{utils.DBConnCfg: utils.MetaDefault},
			},
			utils.UsersFiltersCfg: []string(nil),
			utils.OutDBOptsCfg: map[string]any{
				utils.RedisMaxConnsCfg:           10,
				utils.RedisConnectAttemptsCfg:    20,
				utils.RedisSentinelNameCfg:       "",
				utils.RedisClusterCfg:            false,
				utils.RedisClusterSyncCfg:        "5s",
				utils.RedisClusterOnDownDelayCfg: "0s",
				utils.RedisConnectTimeoutCfg:     "0s",
				utils.RedisReadTimeoutCfg:        "0s",
				utils.RedisWriteTimeoutCfg:       "0s",
				utils.RedisPoolPipelineWindowCfg: "150µs",
				utils.RedisPoolPipelineLimitCfg:  0,
				utils.RedisTLSCfg:                false,
				utils.RedisClientCertificateCfg:  "",
				utils.RedisClientKeyCfg:          "",
				utils.RedisCACertificateCfg:      "",
				utils.MongoQueryTimeoutCfg:       "10s",
				utils.MongoConnSchemeCfg:         "mongodb",
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{MigratorJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionApierS(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		AdminSJSON: map[string]any{
			utils.EnabledCfg:         false,
			utils.CachesConnsCfg:     []string{utils.MetaInternal},
			utils.ActionSConnsCfg:    []string{},
			utils.AttributeSConnsCfg: []string{},
			utils.EEsConnsCfg:        []string{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{AdminSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionEES(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		EEsJSON: map[string]any{
			utils.EnabledCfg:         false,
			utils.AttributeSConnsCfg: []string{},
			utils.CacheCfg: map[string]any{
				utils.MetaFileCSV: map[string]any{
					utils.LimitCfg:     -1,
					utils.PrecacheCfg:  false,
					utils.RemoteCfg:    false,
					utils.ReplicateCfg: false,
					utils.TTLCfg:       "5s",
					utils.StaticTTLCfg: false,
				},
			},
			utils.ExportersCfg: []map[string]any{
				{
					utils.IDCfg:                   utils.MetaDefault,
					utils.TypeCfg:                 utils.MetaNone,
					utils.ExportPathCfg:           "/var/spool/cgrates/ees",
					utils.OptsCfg:                 map[string]any{},
					utils.TimezoneCfg:             utils.EmptyString,
					utils.FiltersCfg:              []string{},
					utils.FlagsCfg:                []string{},
					utils.AttributeIDsCfg:         []string{},
					utils.AttributeContextCfg:     utils.EmptyString,
					utils.SynchronousCfg:          false,
					utils.BlockerCfg:              false,
					utils.AttemptsCfg:             1,
					utils.FieldsCfg:               []map[string]any{},
					utils.ConcurrentRequestsCfg:   0,
					utils.MetricsResetScheduleCfg: "",
					utils.FailedPostsDirCfg:       "/var/spool/cgrates/failed_posts",
					utils.EFsConnsCfg:             []string{utils.MetaInternal},
				},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{EEsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionERS(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		ERsJSON: map[string]any{
			utils.EnabledCfg:         false,
			utils.SessionSConnsCfg:   []string{utils.MetaInternal},
			utils.EEsConnsCfg:        []string{},
			utils.StatSConnsCfg:      []string{},
			utils.ThresholdSConnsCfg: []string{},
			utils.ReadersCfg: []map[string]any{
				{
					utils.FiltersCfg:              []string{},
					utils.FlagsCfg:                []string{},
					utils.IDCfg:                   "*default",
					utils.ProcessedPathCfg:        "/var/spool/cgrates/ers/out",
					utils.RunDelayCfg:             "0",
					utils.StartDelayCfg:           "0",
					utils.SourcePathCfg:           "/var/spool/cgrates/ers/in",
					utils.TenantCfg:               utils.EmptyString,
					utils.TimezoneCfg:             utils.EmptyString,
					utils.EEsIDsCfg:               []string{},
					utils.EEsSuccessIDsCfg:        []string{},
					utils.EEsFailedIDsCfg:         []string{},
					utils.CacheDumpFieldsCfg:      []map[string]any{},
					utils.PartialCommitFieldsCfg:  []map[string]any{},
					utils.ConcurrentRequestsCfg:   1024,
					utils.TypeCfg:                 utils.MetaNone,
					utils.ReconnectsCfg:           -1,
					utils.MaxReconnectIntervalCfg: "5m0s",
					utils.FieldsCfg:               []string{},
					utils.OptsCfg: map[string]any{
						utils.CSVFieldSepOpt:        ",",
						utils.HeaderDefineCharOpt:   ":",
						utils.CSVRowLengthOpt:       0,
						utils.PartialOrderFieldOpt:  "~*req.AnswerTime",
						utils.NatsSubject:           "cgrates_cdrs",
						utils.PartialCacheActionOpt: utils.MetaNone,
					},
				},
			},
			utils.PartialCacheTTLCfg: "1s",
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{ERsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if mp, can := reply[ERsJSON].(map[string]any); !can {
		t.Errorf("Unexpected type: %t", reply[ERsJSON])
	} else {
		mp[utils.ReadersCfg].([]map[string]any)[0][utils.FieldsCfg] = []string{}
		if !reflect.DeepEqual(mp, expected[ERsJSON]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected[ERsJSON]), utils.ToJSON(mp))
		}
	}
}

func TestV1GetConfigSectionRPConns(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		RPCConnsJSON: map[string]any{
			utils.MetaBiJSONLocalHost: map[string]any{
				utils.PoolSize:    0,
				utils.StrategyCfg: utils.MetaFirst,
				utils.Conns: []map[string]any{
					{
						utils.AddressCfg:   "127.0.0.1:2014",
						utils.TransportCfg: rpcclient.BiRPCJSON,
					},
				},
			},
			utils.MetaLocalHost: map[string]any{
				utils.PoolSize:    0,
				utils.StrategyCfg: utils.MetaFirst,
				utils.Conns: []map[string]any{
					{
						utils.AddressCfg:   "127.0.0.1:2012",
						utils.TransportCfg: "*json",
					},
				},
			},
			utils.MetaInternal: map[string]any{
				utils.StrategyCfg: utils.MetaFirst,
				utils.PoolSize:    0,
				utils.Conns: []map[string]any{
					{
						utils.AddressCfg:   utils.MetaInternal,
						utils.TransportCfg: utils.EmptyString,
					},
				},
			},
			rpcclient.BiRPCInternal: map[string]any{
				utils.StrategyCfg: utils.MetaFirst,
				utils.PoolSize:    0,
				utils.Conns: []map[string]any{
					{
						utils.AddressCfg:   rpcclient.BiRPCInternal,
						utils.TransportCfg: utils.EmptyString,
					},
				},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{RPCConnsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionSIPAgent(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		SIPAgentJSON: map[string]any{
			utils.EnabledCfg:             false,
			utils.ListenCfg:              "127.0.0.1:5060",
			utils.ListenNetCfg:           "udp",
			utils.SessionSConnsCfg:       []string{utils.MetaInternal},
			utils.StatSConnsCfg:          []string{},
			utils.ThresholdSConnsCfg:     []string{},
			utils.TimezoneCfg:            utils.EmptyString,
			utils.RetransmissionTimerCfg: "1s",
			utils.RequestProcessorsCfg:   []map[string]any{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{SIPAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionTemplates(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		TemplatesJSON: map[string][]map[string]any{
			utils.MetaErr: {
				{utils.TagCfg: "SessionId", utils.PathCfg: "*rep.Session-Id", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*req.Session-Id", utils.MandatoryCfg: true},
				{utils.TagCfg: "OriginHost", utils.PathCfg: "*rep.Origin-Host", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*vars.OriginHost", utils.MandatoryCfg: true},
				{utils.TagCfg: "OriginRealm", utils.PathCfg: "*rep.Origin-Realm", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*vars.OriginRealm", utils.MandatoryCfg: true},
			},
			utils.MetaASR: {
				{utils.TagCfg: "SessionId", utils.PathCfg: "*diamreq.Session-Id", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*req.Session-Id", utils.MandatoryCfg: true},
				{utils.TagCfg: "OriginHost", utils.PathCfg: "*diamreq.Origin-Host", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*req.Destination-Host", utils.MandatoryCfg: true},
				{utils.TagCfg: "OriginRealm", utils.PathCfg: "*diamreq.Origin-Realm", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*req.Destination-Realm", utils.MandatoryCfg: true},
				{utils.TagCfg: "DestinationRealm", utils.PathCfg: "*diamreq.Destination-Realm", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*req.Origin-Realm", utils.MandatoryCfg: true},
				{utils.TagCfg: "DestinationHost", utils.PathCfg: "*diamreq.Destination-Host", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*req.Origin-Host", utils.MandatoryCfg: true},
				{utils.TagCfg: "AuthApplicationId", utils.PathCfg: "*diamreq.Auth-Application-Id", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*vars.*appid", utils.MandatoryCfg: true},
			},
			utils.MetaCCA:    {},
			utils.MetaRAR:    {},
			"*errSip":        {},
			utils.MetaCdrLog: {},
			"*fsa":           {},
			utils.MetaDMR: {
				{utils.TagCfg: "User-Name", utils.PathCfg: "*radDAReq.User-Name", utils.TypeCfg: "*variable", utils.ValueCfg: "~*oreq.User-Name"},
				{utils.TagCfg: "NAS-IP-Address", utils.PathCfg: "*radDAReq.NAS-IP-Address", utils.TypeCfg: "*variable", utils.ValueCfg: "~*oreq.NAS-IP-Address"},
				{utils.TagCfg: "Acct-Session-Id", utils.PathCfg: "*radDAReq.Acct-Session-Id", utils.TypeCfg: "*variable", utils.ValueCfg: "~*oreq.Acct-Session-Id"},
				{utils.TagCfg: "Reply-Message", utils.PathCfg: "*radDAReq.Reply-Message", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.DisconnectCause"},
			},
			utils.MetaCoA: {
				{utils.TagCfg: "User-Name", utils.PathCfg: "*radDAReq.User-Name", utils.TypeCfg: "*variable", utils.ValueCfg: "~*oreq.User-Name"},
				{utils.TagCfg: "NAS-IP-Address", utils.PathCfg: "*radDAReq.NAS-IP-Address", utils.TypeCfg: "*variable", utils.ValueCfg: "~*oreq.NAS-IP-Address"},
				{utils.TagCfg: "Acct-Session-Id", utils.PathCfg: "*radDAReq.Acct-Session-Id", utils.TypeCfg: "*variable", utils.ValueCfg: "~*oreq.Acct-Session-Id"},
				{utils.TagCfg: "Filter-Id", utils.PathCfg: "*radDAReq.Filter-Id", utils.TypeCfg: "*variable", utils.ValueCfg: "~*req.CustomFilter"},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{TemplatesJSON}}, &reply); err != nil {
		t.Error(err)
	} else if mp, can := reply[TemplatesJSON].(map[string][]map[string]any); !can {
		t.Errorf("Unexpected type: %t", reply[TemplatesJSON])
	} else {
		mp[utils.MetaCCA] = []map[string]any{}
		mp[utils.MetaRAR] = []map[string]any{}
		mp["*errSip"] = []map[string]any{}
		mp["*fsa"] = []map[string]any{}
		mp[utils.MetaCdrLog] = []map[string]any{}
		if !reflect.DeepEqual(reply, expected) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
		}
	}
}

func TestV1GetConfigSectionConfigs(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		ConfigSJSON: map[string]any{
			utils.EnabledCfg: true,
			utils.URLCfg:     "/configs/",
			utils.RootDirCfg: "/var/spool/cgrates/configs",
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	for key := range utils.StatelessDataDBPartitions {
		cfgCgr.cacheCfg.Partitions[key].Limit = 0
	}
	cfgCgr.ConfigSCfg().Enabled = true
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{ConfigSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}

	var result string
	cfgCgr2 := NewDefaultCGRConfig()
	for key := range utils.StatelessDataDBPartitions {
		cfgCgr2.cacheCfg.Partitions[key].Limit = 0
	}
	if err := cfgCgr2.V1SetConfig(context.Background(), &SetConfigArgs{Config: reply, DryRun: true}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Unexpected result")
	} else if cfgCgr := NewDefaultCGRConfig(); !reflect.DeepEqual(cfgCgr.ConfigSCfg(), cfgCgr2.ConfigSCfg()) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(cfgCgr.ConfigSCfg()), utils.ToJSON(cfgCgr2.ConfigSCfg()))
	}

	cfgCgr2 = NewDefaultCGRConfig()
	for key := range utils.StatelessDataDBPartitions {
		cfgCgr2.cacheCfg.Partitions[key].Limit = 0
	}
	if err := cfgCgr2.V1SetConfig(context.Background(), &SetConfigArgs{Config: reply}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Unexpected result")
	} else if !reflect.DeepEqual(cfgCgr.ConfigSCfg(), cfgCgr2.ConfigSCfg()) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(cfgCgr.ConfigSCfg()), utils.ToJSON(cfgCgr2.ConfigSCfg()))
	}
}

func TestV1GetConfigSectionAPIBans(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		APIBanJSON: map[string]any{
			utils.EnabledCfg: false,
			utils.KeysCfg:    []string{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{APIBanJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionAnalyzer(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		AnalyzerSJSON: map[string]any{
			utils.EnabledCfg:         false,
			utils.CleanupIntervalCfg: "1h0m0s",
			utils.DBPathCfg:          "/var/spool/cgrates/analyzers",
			utils.IndexTypeCfg:       utils.MetaScorch,
			utils.EEsConnsCfg:        []string{},
			utils.TTLCfg:             "24h0m0s",
			utils.OptsCfg: map[string]any{
				utils.MetaExporterIDs: []*DynamicStringSliceOpt{},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{AnalyzerSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionRateS(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		RateSJSON: map[string]any{
			utils.EnabledCfg:                    false,
			utils.IndexedSelectsCfg:             true,
			utils.PrefixIndexedFieldsCfg:        []string{},
			utils.SuffixIndexedFieldsCfg:        []string{},
			utils.ExistsIndexedFieldsCfg:        []string{},
			utils.NotExistsIndexedFieldsCfg:     []string{},
			utils.NestedFieldsCfg:               false,
			utils.RateIndexedSelectsCfg:         true,
			utils.RatePrefixIndexedFieldsCfg:    []string{},
			utils.RateSuffixIndexedFieldsCfg:    []string{},
			utils.RateExistsIndexedFieldsCfg:    []string{},
			utils.RateNotExistsIndexedFieldsCfg: []string{},
			utils.RateNestedFieldsCfg:           false,
			utils.Verbosity:                     1000,
			utils.OptsCfg: map[string]any{
				utils.MetaProfileIDs:           []*DynamicStringSliceOpt{},
				utils.MetaStartTime:            []*DynamicStringOpt{{value: RatesStartTimeDftOpt}},
				utils.MetaUsage:                []*DynamicDecimalOpt{{value: RatesUsageDftOpt}},
				utils.MetaIntervalStartCfg:     []*DynamicDecimalOpt{{value: RatesIntervalStartDftOpt}},
				utils.MetaProfileIgnoreFilters: []*DynamicBoolOpt{{value: RatesProfileIgnoreFiltersDftOpt}},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{RateSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionInvalidSection(t *testing.T) {
	var reply map[string]any
	expected := "Invalid section "
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{"invalidSection"}}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestV1ReloadConfigEmptyConfig(t *testing.T) {
	var reply string
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1SetConfig(context.Background(), &SetConfigArgs{}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected output: %+v", reply)
	}
}

func TestV1ReloadConfigUnmarshalError(t *testing.T) {
	var reply string
	expected := "json: unsupported type: chan int"
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1SetConfig(context.Background(), &SetConfigArgs{
		Config: map[string]any{
			StatSJSON: make(chan int),
		},
	},
		&reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestV1ReloadConfigJSONWithLocks(t *testing.T) {
	var reply string
	section := map[string]any{
		"inexistentSection": map[string]any{},
	}
	expected := "Invalid section <inexistentSection> "
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1SetConfig(context.Background(), &SetConfigArgs{Config: section}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+q, received %+q", expected, err)
	}
}

func TestV1GetConfigAsJSONGeneral(t *testing.T) {
	var reply string
	strJSON := `{
		"general": {
			"node_id": "ENGINE1",
		}
	}`
	expected := `{"general":{"caching_delay":"0","connect_attempts":5,"connect_timeout":"1s","dbdata_encoding":"*msgpack","decimal_max_scale":0,"decimal_min_scale":0,"decimal_precision":0,"decimal_rounding_mode":"*toNearestEven","default_caching":"*reload","default_category":"call","default_request_type":"*rated","default_tenant":"cgrates.org","default_timezone":"Local","digest_equal":":","digest_separator":",","locking_timeout":"0","max_parallel_conns":100,"max_reconnect_interval":"0","node_id":"ENGINE1","opts":{"*exporterIDs":[]},"reconnects":-1,"reply_timeout":"2s","rounding_decimals":5,"tpexport_dir":"/var/spool/cgrates/tpe"}}`
	if cfgCgr, err := NewCGRConfigFromJSONStringWithDefaults(strJSON); err != nil {
		t.Error(err)
	} else if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{GeneralJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONDataDB(t *testing.T) {
	var reply string
	expected := `{"db":{"db_conns":{"*default":{"db_host":"","db_name":"","db_password":"","db_port":0,"db_type":"*internal","db_user":"","opts":{"internalDBBackupPath":"/var/lib/cgrates/internal_db/backup/db","internalDBDumpInterval":"1m0s","internalDBDumpPath":"/var/lib/cgrates/internal_db/db","internalDBFileSizeLimit":1073741824,"internalDBRewriteInterval":"1h0m0s","internalDBStartTimeout":"5m0s","mongoConnScheme":"mongodb","mongoQueryTimeout":"10s","mysqlDSNParams":{},"mysqlLocation":"Local","pgSSLMode":"disable","redisCACertificate":"","redisClientCertificate":"","redisClientKey":"","redisCluster":false,"redisClusterOndownDelay":"0s","redisClusterSync":"5s","redisConnectAttempts":20,"redisConnectTimeout":"0s","redisMaxConns":10,"redisPoolPipelineLimit":0,"redisPoolPipelineWindow":"150µs","redisReadTimeout":"0s","redisSentinel":"","redisTLS":false,"redisWriteTimeout":"0s","sqlConnMaxLifetime":"0s","sqlLogLevel":3,"sqlMaxIdleConns":10,"sqlMaxOpenConns":100},"prefix_indexed_fields":[],"remote_conn_id":"","remote_conns":[],"replication_cache":"","replication_conns":[],"replication_failed_dir":"","replication_filtered":false,"replication_interval":"0s","string_indexed_fields":[]}},"items":{"*account_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*accounts":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*action_profile_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*action_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*attribute_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*attribute_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*cdrs":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*charger_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*charger_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*filters":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*ip_allocations":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*ip_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*ip_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*load_ids":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*ranking_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*rankings":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*rate_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*rate_profile_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*rate_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*resource_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*resource_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*resources":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*reverse_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*route_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*route_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*stat_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*statqueue_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*statqueues":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*threshold_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*threshold_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*thresholds":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*trend_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*trends":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*versions":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false}}}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{DBJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONTls(t *testing.T) {
	var reply string
	expected := `{"tls":{"ca_certificate":"","client_certificate":"","client_key":"","server_certificate":"","server_key":"","server_name":"","server_policy":4}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{TlsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONTCache(t *testing.T) {
	var reply string
	expected := `{"caches":{"partitions":{"*account_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*accounts":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*action_profile_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*action_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*apiban":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"2m0s"},"*attribute_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*attribute_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*caps_events":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*cdr_ids":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"10m0s"},"*charger_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*charger_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*closed_sessions":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"10s"},"*diameter_messages":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*event_charges":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"10s"},"*event_ips":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*event_resources":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*filters":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*ip_allocations":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*ip_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*ip_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*load_ids":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*radius_packets":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*ranking_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rankings":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rate_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rate_profile_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rate_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*replication_hosts":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*resource_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*resource_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*resources":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*reverse_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*route_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*route_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rpc_connections":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rpc_responses":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"2s"},"*sentrypeer":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":true,"ttl":"24h0m0s"},"*stat_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*statqueue_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*statqueues":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*stir":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*threshold_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*threshold_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*thresholds":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*trend_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*trends":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*uch":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"}},"remote_conns":[],"replication_conns":[]}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{CacheJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONTListen(t *testing.T) {
	var reply string
	expected := `{"listen":{"http":"127.0.0.1:2080","http_tls":"127.0.0.1:2280","rpc_gob":"127.0.0.1:2013","rpc_gob_tls":"127.0.0.1:2023","rpc_json":"127.0.0.1:2012","rpc_json_tls":"127.0.0.1:2022"}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{ListenJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONAccounts(t *testing.T) {
	var reply string
	expected := `{"accounts":{"attributes_conns":[],"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"max_iterations":1000,"max_usage":"259200000000000","nested_fields":false,"notexists_indexed_fields":[],"opts":{"*profileIDs":[],"*profileIgnoreFilters":[{"FilterIDs":null,"Tenant":""}],"*usage":[{"FilterIDs":null,"Tenant":""}]},"prefix_indexed_fields":[],"rates_conns":[],"suffix_indexed_fields":[],"thresholds_conns":[]}}`
	cfg := NewDefaultCGRConfig()
	if err := cfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{AccountSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONHTTP(t *testing.T) {
	var reply string
	expected := `{"http":{"auth_users":{},"client_opts":{"dialFallbackDelay":"300ms","dialKeepAlive":"30s","dialTimeout":"30s","disableCompression":false,"disableKeepAlives":false,"expectContinueTimeout":"0s","forceAttemptHttp2":true,"idleConnTimeout":"1m30s","maxConnsPerHost":0,"maxIdleConns":100,"maxIdleConnsPerHost":2,"responseHeaderTimeout":"0s","skipTLSVerification":false,"tlsHandshakeTimeout":"10s"},"freeswitch_cdrs_url":"/freeswitch_json","http_cdrs":"/cdr_http","json_rpc_url":"/jsonrpc","pprof_path":"/debug/pprof/","registrars_url":"/registrar","use_basic_auth":false,"ws_url":"/ws"}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{HTTPJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONFilterS(t *testing.T) {
	var reply string
	expected := `{"filters":{"accounts_conns":[],"rankings_conns":[],"resources_conns":[],"stats_conns":[],"trends_conns":[]}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{FilterSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONCdrs(t *testing.T) {
	var reply string
	expected := `{"cdrs":{"accounts_conns":[],"actions_conns":[],"attributes_conns":[],"chargers_conns":[],"ees_conns":[],"enabled":false,"extra_fields":[],"online_cdr_exports":null,"opts":{"*accounts":[{"FilterIDs":null,"Tenant":""}],"*attributes":[{"FilterIDs":null,"Tenant":""}],"*chargers":[{"FilterIDs":null,"Tenant":""}],"*ees":[{"FilterIDs":null,"Tenant":""}],"*rates":[{"FilterIDs":null,"Tenant":""}],"*refund":[{"FilterIDs":null,"Tenant":""}],"*rerate":[{"FilterIDs":null,"Tenant":""}],"*stats":[{"FilterIDs":null,"Tenant":""}],"*store":[{"FilterIDs":null,"Tenant":""}],"*thresholds":[{"FilterIDs":null,"Tenant":""}]},"rates_conns":[],"session_cost_retries":5,"stats_conns":[],"thresholds_conns":[]}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{CDRsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONSessionS(t *testing.T) {
	var reply string
	expected := `{"sessions":{"accounts_conns":[],"actions_conns":[],"alterable_fields":[],"attributes_conns":[],"cdrs_conns":[],"channel_sync_interval":"0","chargers_conns":[],"client_protocol":1,"default_usage":{"*any":"3h0m0s","*data":"1048576","*sms":"1","*voice":"3h0m0s"},"enabled":false,"ips_conns":[],"listen_bigob":"","listen_bijson":"127.0.0.1:2014","min_dur_low_balance":"0","opts":{"*accounts":[{"FilterIDs":null,"Tenant":""}],"*accountsForceUsage":[],"*attributes":[{"FilterIDs":null,"Tenant":""}],"*attributesDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*blockerError":[{"FilterIDs":null,"Tenant":""}],"*cdrs":[{"FilterIDs":null,"Tenant":""}],"*cdrsDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*chargeable":[{"FilterIDs":null,"Tenant":""}],"*chargers":[{"FilterIDs":null,"Tenant":""}],"*debitInterval":[{"FilterIDs":null,"Tenant":""}],"*forceUsage":[],"*initiate":[{"FilterIDs":null,"Tenant":""}],"*ips":[{"FilterIDs":null,"Tenant":""}],"*ipsAllocate":[{"FilterIDs":null,"Tenant":""}],"*ipsAuthorize":[{"FilterIDs":null,"Tenant":""}],"*ipsRelease":[{"FilterIDs":null,"Tenant":""}],"*maxUsage":[{"FilterIDs":null,"Tenant":""}],"*message":[{"FilterIDs":null,"Tenant":""}],"*originID":[],"*resources":[{"FilterIDs":null,"Tenant":""}],"*resourcesAllocate":[{"FilterIDs":null,"Tenant":""}],"*resourcesAuthorize":[{"FilterIDs":null,"Tenant":""}],"*resourcesDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*resourcesRelease":[{"FilterIDs":null,"Tenant":""}],"*routes":[{"FilterIDs":null,"Tenant":""}],"*routesDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*stats":[{"FilterIDs":null,"Tenant":""}],"*statsDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*terminate":[{"FilterIDs":null,"Tenant":""}],"*thresholds":[{"FilterIDs":null,"Tenant":""}],"*thresholdsDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*ttl":[{"FilterIDs":null,"Tenant":""}],"*ttlLastUsage":[],"*ttlLastUsed":[],"*ttlMaxDelay":[{"FilterIDs":null,"Tenant":""}],"*ttlUsage":[],"*update":[{"FilterIDs":null,"Tenant":""}]},"rates_conns":[],"replication_conns":[],"resources_conns":[],"routes_conns":[],"session_indexes":[],"stats_conns":[],"stir":{"allowed_attest":["*any"],"default_attest":"A","payload_maxduration":"-1","privatekey_path":"","publickey_path":""},"store_session_costs":false,"terminate_attempts":5,"thresholds_conns":[]}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{SessionSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONFreeSwitchAgent(t *testing.T) {
	var reply string
	expected := `{"freeswitch_agent":{"active_session_delimiter":",","create_cdr":false,"empty_balance_ann_file":"","empty_balance_context":"","enabled":false,"event_socket_conns":[{"address":"127.0.0.1:8021","alias":"127.0.0.1:8021","max_reconnect_interval":"0s","password":"ClueCon","reconnects":5,"reply_timeout":"1m0s"}],"extra_fields":[],"low_balance_ann_file":"","max_wait_connection":"2s","request_processors":[],"sessions_conns":["*birpc_internal"],"subscribe_park":true}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{FreeSWITCHAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONFKamailioAgent(t *testing.T) {
	var reply string
	expected := `{"kamailio_agent":{"create_cdr":false,"enabled":false,"evapi_conns":[{"address":"127.0.0.1:8448","alias":"","max_reconnect_interval":"0s","reconnects":5}],"sessions_conns":["*birpc_internal"],"timezone":""}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{KamailioAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONAsteriskAgent(t *testing.T) {
	var reply string
	expected := `{"asterisk_agent":{"asterisk_conns":[{"address":"127.0.0.1:8088","alias":"","connect_attempts":3,"max_reconnect_interval":"0s","password":"CGRateS.org","reconnects":5,"user":"cgrates"}],"create_cdr":false,"enabled":false,"sessions_conns":["*birpc_internal"]}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{AsteriskAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONADiameterAgent(t *testing.T) {
	var reply string
	expected := `{"diameter_agent":{"asr_template":"","conn_health_check_interval":"0s","dictionaries_path":"/usr/share/cgrates/diameter/dict/","enabled":false,"forced_disconnect":"*none","listen":"127.0.0.1:3868","listen_net":"tcp","origin_host":"CGR-DA","origin_realm":"cgrates.org","product_name":"CGRateS","rar_template":"","request_processors":[],"sessions_conns":["*birpc_internal"],"stats_conns":[],"synced_conn_requests":false,"thresholds_conns":[],"vendor_id":0}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{DiameterAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONARadiusAgent(t *testing.T) {
	var reply string
	expected := `{"radius_agent":{"client_da_addresses":{},"client_dictionaries":{"*default":["/usr/share/cgrates/radius/dict/"]},"client_secrets":{"*default":"CGRateS.org"},"coa_template":"*coa","dmr_template":"*dmr","enabled":false,"listeners":[{"acct_address":"127.0.0.1:1813","auth_address":"127.0.0.1:1812","network":"udp"}],"request_processors":[],"requests_cache_key":"","sessions_conns":["*internal"],"stats_conns":[],"thresholds_conns":[]}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{RadiusAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONDNSAgent(t *testing.T) {
	var reply string
	expected := `{"dns_agent":{"enabled":false,"listeners":[{"address":"127.0.0.1:53","network":"udp"}],"request_processors":[],"sessions_conns":["*internal"],"stats_conns":[],"thresholds_conns":[],"timezone":""}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{DNSAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONAttributes(t *testing.T) {
	var reply string
	expected := `{"attributes":{"accounts_conns":[],"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*processRuns":[{"FilterIDs":null,"Tenant":""}],"*profileIDs":[],"*profileIgnoreFilters":[{"FilterIDs":null,"Tenant":""}],"*profileRuns":[{"FilterIDs":null,"Tenant":""}]},"prefix_indexed_fields":[],"resources_conns":[],"stats_conns":[],"suffix_indexed_fields":[]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{AttributeSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONChargerS(t *testing.T) {
	var reply string
	expected := `{"chargers":{"attributes_conns":[],"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"prefix_indexed_fields":[],"suffix_indexed_fields":[]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{ChargerSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONResourceS(t *testing.T) {
	var reply string
	expected := `{"resources":{"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*units":[{"FilterIDs":null,"Tenant":""}],"*usageID":[{"FilterIDs":null,"Tenant":""}],"*usageTTL":[{"FilterIDs":null,"Tenant":""}]},"prefix_indexed_fields":[],"store_interval":"","suffix_indexed_fields":[],"thresholds_conns":[]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{ResourceSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONStatS(t *testing.T) {
	var reply string
	expected := `{"stats":{"ees_conns":[],"ees_exporter_ids":null,"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*profileIDs":[],"*profileIgnoreFilters":[{"FilterIDs":null,"Tenant":""}],"*roundingDecimals":[]},"prefix_indexed_fields":[],"store_interval":"","store_uncompressed_limit":0,"suffix_indexed_fields":[],"thresholds_conns":[]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{StatSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONThresholdS(t *testing.T) {
	var reply string
	expected := `{"thresholds":{"actions_conns":[],"ees_conns":[],"ees_exporter_ids":[],"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*profileIDs":[],"*profileIgnoreFilters":[{"FilterIDs":null,"Tenant":""}]},"prefix_indexed_fields":[],"store_interval":"","suffix_indexed_fields":[]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{ThresholdSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONRouteS(t *testing.T) {
	var reply string
	expected := `{"routes":{"accounts_conns":[],"attributes_conns":[],"default_ratio":1,"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*context":[{"FilterIDs":null,"Tenant":""}],"*ignoreErrors":[{"FilterIDs":null,"Tenant":""}],"*limit":[],"*maxCost":[{"Tenant":"","Value":""}],"*maxItems":[],"*offset":[],"*profileCount":[{"FilterIDs":null,"Tenant":""}],"*usage":[{"FilterIDs":null,"Tenant":""}]},"prefix_indexed_fields":[],"rates_conns":[],"resources_conns":[],"stats_conns":[],"suffix_indexed_fields":[]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{RouteSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONSureTax(t *testing.T) {
	var reply string
	expected := `{"suretax":{"bill_to_number":"","business_unit":"","client_number":"","client_tracking":"~*opts.*originID","customer_number":"~*req.Subject","include_local_cost":false,"orig_number":"~*req.Subject","p2pplus4":"","p2pzipcode":"","plus4":"","regulatory_code":"03","response_group":"03","response_type":"D4","return_file_code":"0","sales_type_code":"R","tax_exemption_code_list":"","tax_included":"0","tax_situs_rule":"04","term_number":"~*req.Destination","timezone":"UTC","trans_type_code":"010101","unit_type":"00","units":"1","url":"","validation_key":"","zipcode":""}}`
	cgrCfg := NewDefaultCGRConfig()

	cgrCfg.SureTaxCfg().Timezone = time.UTC
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{SureTaxJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONLoaders(t *testing.T) {
	var reply string
	expected := `{"loaders":[{"action":"*store","cache":{"*accounts":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*action_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*attributes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*chargers":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*filters":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*ips":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*rankings":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*rate_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*resources":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*routes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*stats":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*thresholds":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*trends":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"}},"caches_conns":["*internal"],"data":[{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"new_branch":true,"path":"Rules.Type","tag":"Type","type":"*variable","value":"~*req.2"},{"path":"Rules.Element","tag":"Element","type":"*variable","value":"~*req.3"},{"path":"Rules.Values","tag":"Values","type":"*variable","value":"~*req.4"}],"file_name":"Filters.csv","flags":null,"type":"*filters"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"TenantID","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ProfileID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"new_branch":true,"path":"Attributes.FilterIDs","tag":"AttributeFilterIDs","type":"*variable","value":"~*req.5"},{"path":"Attributes.Blockers","tag":"AttributeBlockers","type":"*variable","value":"~*req.6"},{"path":"Attributes.Path","tag":"Path","type":"*variable","value":"~*req.7"},{"path":"Attributes.Type","tag":"Type","type":"*variable","value":"~*req.8"},{"path":"Attributes.Value","tag":"Value","type":"*variable","value":"~*req.9"}],"file_name":"Attributes.csv","flags":null,"type":"*attributes"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"UsageTTL","tag":"TTL","type":"*variable","value":"~*req.4"},{"path":"Limit","tag":"Limit","type":"*variable","value":"~*req.5"},{"path":"AllocationMessage","tag":"AllocationMessage","type":"*variable","value":"~*req.6"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.7"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.8"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.9"}],"file_name":"Resources.csv","flags":null,"type":"*resources"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"TTL","tag":"TTL","type":"*variable","value":"~*req.4"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.5"},{"new_branch":true,"path":"Pools.ID","tag":"PoolID","type":"*variable","value":"~*req.6"},{"path":"Pools.FilterIDs","tag":"PoolFilterIDs","type":"*variable","value":"~*req.7"},{"path":"Pools.Type","tag":"PoolType","type":"*variable","value":"~*req.8"},{"path":"Pools.Range","tag":"PoolRange","type":"*variable","value":"~*req.9"},{"path":"Pools.Strategy","tag":"PoolStrategy","type":"*variable","value":"~*req.10"},{"path":"Pools.Message","tag":"PoolMessage","type":"*variable","value":"~*req.11"},{"path":"Pools.Weights","tag":"PoolWeights","type":"*variable","value":"~*req.12"},{"path":"Pools.Blockers","tag":"PoolBlockers","type":"*variable","value":"~*req.13"}],"file_name":"IPs.csv","flags":null,"type":"*ips"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"QueueLength","tag":"QueueLength","type":"*variable","value":"~*req.5"},{"path":"TTL","tag":"TTL","type":"*variable","value":"~*req.6"},{"path":"MinItems","tag":"MinItems","type":"*variable","value":"~*req.7"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.8"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.9"},{"new_branch":true,"path":"Metrics.MetricID","tag":"MetricIDs","type":"*variable","value":"~*req.10"},{"path":"Metrics.FilterIDs","tag":"MetricFilterIDs","type":"*variable","value":"~*req.11"},{"path":"Metrics.Blockers","tag":"MetricBlockers","type":"*variable","value":"~*req.12"}],"file_name":"Stats.csv","flags":null,"type":"*stats"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"MaxHits","tag":"MaxHits","type":"*variable","value":"~*req.4"},{"path":"MinHits","tag":"MinHits","type":"*variable","value":"~*req.5"},{"path":"MinSleep","tag":"MinSleep","type":"*variable","value":"~*req.6"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.7"},{"path":"ActionProfileIDs","tag":"ActionProfileIDs","type":"*variable","value":"~*req.8"},{"path":"Async","tag":"Async","type":"*variable","value":"~*req.9"},{"path":"EeIDs","tag":"EeIDs","type":"*variable","value":"~*req.10"}],"file_name":"Thresholds.csv","flags":null,"type":"*thresholds"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"Schedule","tag":"Schedule","type":"*variable","value":"~*req.2"},{"path":"StatID","tag":"StatID","type":"*variable","value":"~*req.3"},{"path":"Metrics","tag":"Metrics","type":"*variable","value":"~*req.4"},{"path":"TTL","tag":"TTL","type":"*variable","value":"~*req.5"},{"path":"QueueLength","tag":"QueueLength","type":"*variable","value":"~*req.6"},{"path":"MinItems","tag":"MinItems","type":"*variable","value":"~*req.7"},{"path":"CorrelationType","tag":"CorrelationType","type":"*variable","value":"~*req.8"},{"path":"Tolerance","tag":"Tolerance","type":"*variable","value":"~*req.9"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.10"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.11"}],"file_name":"Trends.csv","flags":null,"type":"*trends"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"Schedule","tag":"Schedule","type":"*variable","value":"~*req.2"},{"path":"StatIDs","tag":"StatIDs","type":"*variable","value":"~*req.3"},{"path":"MetricIDs","tag":"MetricIDs","type":"*variable","value":"~*req.4"},{"path":"Sorting","tag":"Sorting","type":"*variable","value":"~*req.5"},{"path":"SortingParameters","tag":"SortingParameters","type":"*variable","value":"~*req.6"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.7"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.8"}],"file_name":"Rankings.csv","flags":null,"type":"*rankings"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Sorting","tag":"Sorting","type":"*variable","value":"~*req.5"},{"path":"SortingParameters","tag":"SortingParameters","type":"*variable","value":"~*req.6"},{"new_branch":true,"path":"Routes.ID","tag":"RouteID","type":"*variable","value":"~*req.7"},{"path":"Routes.FilterIDs","tag":"RouteFilterIDs","type":"*variable","value":"~*req.8"},{"path":"Routes.AccountIDs","tag":"RouteAccountIDs","type":"*variable","value":"~*req.9"},{"path":"Routes.RateProfileIDs","tag":"RouteRateProfileIDs","type":"*variable","value":"~*req.10"},{"path":"Routes.ResourceIDs","tag":"RouteResourceIDs","type":"*variable","value":"~*req.11"},{"path":"Routes.StatIDs","tag":"RouteStatIDs","type":"*variable","value":"~*req.12"},{"path":"Routes.Weights","tag":"RouteWeights","type":"*variable","value":"~*req.13"},{"path":"Routes.Blockers","tag":"RouteBlockers","type":"*variable","value":"~*req.14"},{"path":"Routes.RouteParameters","tag":"RouteParameters","type":"*variable","value":"~*req.15"}],"file_name":"Routes.csv","flags":null,"type":"*routes"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"RunID","tag":"RunID","type":"*variable","value":"~*req.5"},{"path":"AttributeIDs","tag":"AttributeIDs","type":"*variable","value":"~*req.6"}],"file_name":"Chargers.csv","flags":null,"type":"*chargers"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"MinCost","tag":"MinCost","type":"*variable","value":"~*req.4"},{"path":"MaxCost","tag":"MaxCost","type":"*variable","value":"~*req.5"},{"path":"MaxCostStrategy","tag":"MaxCostStrategy","type":"*variable","value":"~*req.6"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].FilterIDs","tag":"RateFilterIDs","type":"*variable","value":"~*req.8"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].ActivationTimes","tag":"RateActivationTimes","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].Weights","tag":"RateWeights","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].Blocker","tag":"RateBlocker","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.7:"],"new_branch":true,"path":"Rates[\u003c~*req.7\u003e].IntervalRates.IntervalStart","tag":"RateIntervalStart","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.FixedFee","tag":"RateFixedFee","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.RecurrentFee","tag":"RateRecurrentFee","type":"*variable","value":"~*req.14"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.Unit","tag":"RateUnit","type":"*variable","value":"~*req.15"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.Increment","tag":"RateIncrement","type":"*variable","value":"~*req.16"}],"file_name":"Rates.csv","flags":null,"type":"*rate_profiles"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Schedule","tag":"Schedule","type":"*variable","value":"~*req.5"},{"path":"Targets[\u003c~*req.6\u003e]","tag":"TargetIDs","type":"*variable","value":"~*req.7"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].FilterIDs","tag":"ActionFilterIDs","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].TTL","tag":"ActionTTL","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Type","tag":"ActionType","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Opts","tag":"ActionOpts","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Weights","tag":"ActionWeights","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Blockers","tag":"ActionBlockers","type":"*variable","value":"~*req.14"},{"filters":["*notempty:~*req.8:"],"new_branch":true,"path":"Actions[\u003c~*req.8\u003e].Diktats.ID","tag":"ActionDiktatsID","type":"*variable","value":"~*req.15"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.FilterIDs","tag":"ActionDiktatsFilterIDs","type":"*variable","value":"~*req.16"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.Opts","tag":"ActionDiktatsOpts","type":"*variable","value":"~*req.17"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.Weights","tag":"ActionDiktatsWeights","type":"*variable","value":"~*req.18"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.Blockers","tag":"ActionDiktatsBlockers","type":"*variable","value":"~*req.19"}],"file_name":"Actions.csv","flags":null,"type":"*action_profiles"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Opts","tag":"Opts","type":"*variable","value":"~*req.5"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].FilterIDs","tag":"BalanceFilterIDs","type":"*variable","value":"~*req.7"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Weights","tag":"BalanceWeights","type":"*variable","value":"~*req.8"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Blockers","tag":"BalanceBlockers","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Type","tag":"BalanceType","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Units","tag":"BalanceUnits","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].UnitFactors","tag":"BalanceUnitFactors","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Opts","tag":"BalanceOpts","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].CostIncrements","tag":"BalanceCostIncrements","type":"*variable","value":"~*req.14"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].AttributeIDs","tag":"BalanceAttributeIDs","type":"*variable","value":"~*req.15"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].RateProfileIDs","tag":"BalanceRateProfileIDs","type":"*variable","value":"~*req.16"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.17"}],"file_name":"Accounts.csv","flags":null,"type":"*accounts"}],"enabled":false,"field_separator":",","id":"*default","lockfile_path":".cgr.lck","opts":{"*cache":"","*forceLock":false,"*stopOnError":false,"*withIndex":true},"run_delay":"0","tenant":"","tp_in_dir":"/var/spool/cgrates/loader/in","tp_out_dir":"/var/spool/cgrates/loader/out"}]}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{LoaderSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONCgrLoader(t *testing.T) {
	var reply string
	expected := `{"loader":{"actions_conns":["*localhost"],"caches_conns":["*localhost"],"data_path":"./","disable_reverse":false,"field_separator":",","gapi_credentials":".gapi/credentials.json","gapi_token":".gapi/token.json","tpid":""}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{LoaderJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONCgrMigrator(t *testing.T) {
	var reply string
	expected := `{"migrator":{"fromItems":{"*accounts":{"dbConn":"*default"},"*charger_profiles":{"dbConn":"*default"},"*filters":{"dbConn":"*default"},"*load_ids":{"dbConn":"*default"},"*statqueue_profiles":{"dbConn":"*default"},"*versions":{"dbConn":"*default"}},"out_db_opts":{"mongoConnScheme":"mongodb","mongoQueryTimeout":"10s","redisCACertificate":"","redisClientCertificate":"","redisClientKey":"","redisCluster":false,"redisClusterOndownDelay":"0s","redisClusterSync":"5s","redisConnectAttempts":20,"redisConnectTimeout":"0s","redisMaxConns":10,"redisPoolPipelineLimit":0,"redisPoolPipelineWindow":"150µs","redisReadTimeout":"0s","redisSentinel":"","redisTLS":false,"redisWriteTimeout":"0s"},"users_filters":null}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{MigratorJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONApierS(t *testing.T) {
	var reply string
	expected := `{"admins":{"actions_conns":[],"attributes_conns":[],"caches_conns":["*internal"],"ees_conns":[],"enabled":false}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{AdminSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONCfgEES(t *testing.T) {
	var reply string
	expected := `{"ees":{"attributes_conns":[],"cache":{"*fileCSV":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"}},"enabled":false,"exporters":[{"attempts":1,"attribute_context":"","attribute_ids":[],"blocker":false,"concurrent_requests":0,"efs_conns":["*internal"],"export_path":"/var/spool/cgrates/ees","failed_posts_dir":"/var/spool/cgrates/failed_posts","fields":[],"filters":[],"flags":[],"id":"*default","metrics_reset_schedule":"","opts":{},"synchronous":false,"timezone":"","type":"*none"}]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{EEsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONCfgERS(t *testing.T) {
	var reply string
	expected := `{"ers":{"ees_conns":[],"enabled":false,"partial_cache_ttl":"1s","readers":[{"cache_dump_fields":[],"concurrent_requests":1024,"ees_failed_ids":[],"ees_ids":[],"ees_success_ids":[],"fields":[{"mandatory":true,"path":"*cgreq.ToR","tag":"ToR","type":"*variable","value":"~*req.2"},{"mandatory":true,"path":"*cgreq.OriginID","tag":"OriginID","type":"*variable","value":"~*req.3"},{"mandatory":true,"path":"*cgreq.RequestType","tag":"RequestType","type":"*variable","value":"~*req.4"},{"mandatory":true,"path":"*cgreq.Tenant","tag":"Tenant","type":"*variable","value":"~*req.6"},{"mandatory":true,"path":"*cgreq.Category","tag":"Category","type":"*variable","value":"~*req.7"},{"mandatory":true,"path":"*cgreq.Account","tag":"Account","type":"*variable","value":"~*req.8"},{"mandatory":true,"path":"*cgreq.Subject","tag":"Subject","type":"*variable","value":"~*req.9"},{"mandatory":true,"path":"*cgreq.Destination","tag":"Destination","type":"*variable","value":"~*req.10"},{"mandatory":true,"path":"*cgreq.SetupTime","tag":"SetupTime","type":"*variable","value":"~*req.11"},{"mandatory":true,"path":"*cgreq.AnswerTime","tag":"AnswerTime","type":"*variable","value":"~*req.12"},{"mandatory":true,"path":"*cgreq.Usage","tag":"Usage","type":"*variable","value":"~*req.13"}],"filters":[],"flags":[],"id":"*default","max_reconnect_interval":"5m0s","opts":{"csvFieldSeparator":",","csvHeaderDefineChar":":","csvRowLength":0,"natsSubject":"cgrates_cdrs","partialCacheAction":"*none","partialOrderField":"~*req.AnswerTime"},"partial_commit_fields":[],"processed_path":"/var/spool/cgrates/ers/out","reconnects":-1,"run_delay":"0","source_path":"/var/spool/cgrates/ers/in","start_delay":"0","tenant":"","timezone":"","type":"*none"}],"sessions_conns":["*internal"],"stats_conns":[],"thresholds_conns":[]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{ERsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONSIPAgent(t *testing.T) {
	var reply string
	expected := `{"sip_agent":{"enabled":false,"listen":"127.0.0.1:5060","listen_net":"udp","request_processors":[],"retransmission_timer":"1s","sessions_conns":["*internal"],"stats_conns":[],"thresholds_conns":[],"timezone":""}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{SIPAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONConfigS(t *testing.T) {
	var reply string
	expected := `{"configs":{"enabled":false,"root_dir":"/var/spool/cgrates/configs","url":"/configs/"}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{ConfigSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONApiBan(t *testing.T) {
	var reply string
	expected := `{"apiban":{"enabled":false,"keys":[]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{APIBanJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONRPCConns(t *testing.T) {
	var reply string
	expected := `{"rpc_conns":{"*bijson_localhost":{"conns":[{"address":"127.0.0.1:2014","transport":"*birpc_json"}],"poolSize":0,"strategy":"*first"},"*birpc_internal":{"conns":[{"address":"*birpc_internal","transport":""}],"poolSize":0,"strategy":"*first"},"*internal":{"conns":[{"address":"*internal","transport":""}],"poolSize":0,"strategy":"*first"},"*localhost":{"conns":[{"address":"127.0.0.1:2012","transport":"*json"}],"poolSize":0,"strategy":"*first"}}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{RPCConnsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONTemplates(t *testing.T) {
	var reply string
	expected := `{"templates":{"*asr":[{"mandatory":true,"path":"*diamreq.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*diamreq.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*req.Destination-Host"},{"mandatory":true,"path":"*diamreq.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*req.Destination-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Realm","tag":"DestinationRealm","type":"*variable","value":"~*req.Origin-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Host","tag":"DestinationHost","type":"*variable","value":"~*req.Origin-Host"},{"mandatory":true,"path":"*diamreq.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"}],"*cca":[{"mandatory":true,"path":"*rep.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"path":"*rep.Result-Code","tag":"ResultCode","type":"*constant","value":"2001"},{"mandatory":true,"path":"*rep.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*vars.OriginHost"},{"mandatory":true,"path":"*rep.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*vars.OriginRealm"},{"mandatory":true,"path":"*rep.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"},{"mandatory":true,"path":"*rep.CC-Request-Type","tag":"CCRequestType","type":"*variable","value":"~*req.CC-Request-Type"},{"mandatory":true,"path":"*rep.CC-Request-Number","tag":"CCRequestNumber","type":"*variable","value":"~*req.CC-Request-Number"}],"*cdrLog":[{"mandatory":true,"path":"*cdr.ToR","tag":"ToR","type":"*variable","value":"~*req.BalanceType"},{"mandatory":true,"path":"*cdr.OriginHost","tag":"OriginHost","type":"*constant","value":"127.0.0.1"},{"mandatory":true,"path":"*cdr.RequestType","tag":"RequestType","type":"*constant","value":"*none"},{"mandatory":true,"path":"*cdr.Tenant","tag":"Tenant","type":"*variable","value":"~*req.Tenant"},{"mandatory":true,"path":"*cdr.Account","tag":"Account","type":"*variable","value":"~*req.Account"},{"mandatory":true,"path":"*cdr.Subject","tag":"Subject","type":"*variable","value":"~*req.Account"},{"mandatory":true,"path":"*cdr.Cost","tag":"Cost","type":"*variable","value":"~*req.Cost"},{"mandatory":true,"path":"*cdr.Source","tag":"Source","type":"*constant","value":"*cdrLog"},{"mandatory":true,"path":"*cdr.Usage","tag":"Usage","type":"*constant","value":"1"},{"mandatory":true,"path":"*cdr.RunID","tag":"RunID","type":"*variable","value":"~*req.ActionType"},{"mandatory":true,"path":"*cdr.SetupTime","tag":"SetupTime","type":"*constant","value":"*now"},{"mandatory":true,"path":"*cdr.AnswerTime","tag":"AnswerTime","type":"*constant","value":"*now"},{"mandatory":true,"path":"*cdr.PreRated","tag":"PreRated","type":"*constant","value":"true"}],"*coa":[{"path":"*radDAReq.User-Name","tag":"User-Name","type":"*variable","value":"~*oreq.User-Name"},{"path":"*radDAReq.NAS-IP-Address","tag":"NAS-IP-Address","type":"*variable","value":"~*oreq.NAS-IP-Address"},{"path":"*radDAReq.Acct-Session-Id","tag":"Acct-Session-Id","type":"*variable","value":"~*oreq.Acct-Session-Id"},{"path":"*radDAReq.Filter-Id","tag":"Filter-Id","type":"*variable","value":"~*req.CustomFilter"}],"*dmr":[{"path":"*radDAReq.User-Name","tag":"User-Name","type":"*variable","value":"~*oreq.User-Name"},{"path":"*radDAReq.NAS-IP-Address","tag":"NAS-IP-Address","type":"*variable","value":"~*oreq.NAS-IP-Address"},{"path":"*radDAReq.Acct-Session-Id","tag":"Acct-Session-Id","type":"*variable","value":"~*oreq.Acct-Session-Id"},{"path":"*radDAReq.Reply-Message","tag":"Reply-Message","type":"*variable","value":"~*req.DisconnectCause"}],"*err":[{"mandatory":true,"path":"*rep.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*rep.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*vars.OriginHost"},{"mandatory":true,"path":"*rep.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*vars.OriginRealm"}],"*errSip":[{"mandatory":true,"path":"*rep.Request","tag":"Request","type":"*constant","value":"SIP/2.0 500 Internal Server Error"}],"*fsa":[{"path":"*cgreq.ToR","tag":"ToR","type":"*constant","value":"*voice"},{"path":"*cgreq.PDD","tag":"PDD","type":"*composed","value":"~*req.variable_progress_mediamsec;ms"},{"path":"*cgreq.ACD","tag":"ACD","type":"*composed","value":"~*req.variable_cdr_acd;s"},{"path":"*cgreq.OriginID","tag":"OriginID","type":"*variable","value":"~*req.Unique-ID"},{"path":"*opts.*originID","tag":"*originID","type":"*variable","value":"~*req.Unique-ID"},{"path":"*cgreq.OriginHost","tag":"OriginHost","type":"*variable","value":"~*req.variable_cgr_originhost"},{"path":"*cgreq.Account","tag":"Account","type":"*variable","value":"~*req.Caller-Username"},{"path":"*cgreq.Source","tag":"Source","type":"*composed","value":"FS_;~*req.Event-Name"},{"filters":["*string:*req.variable_process_cdr:false"],"path":"*cgreq.RequestType","tag":"RequestType","type":"*constant","value":"*none"},{"filters":["*string:*req.Caller-Dialplan:inline"],"path":"*cgreq.RequestType","tag":"RequestType","type":"*constant","value":"*none"},{"filters":["*exists:*cgreq.RequestType:"],"path":"*cgreq.RequestType","tag":"RequestType","type":"*constant","value":"*prepaid"},{"path":"*cgreq.Tenant","tag":"Tenant","type":"*constant","value":"cgrates.org"},{"path":"*cgreq.Category","tag":"Category","type":"*constant","value":"call"},{"path":"*cgreq.Subject","tag":"Subject","type":"*variable","value":"~*req.Caller-Username"},{"path":"*cgreq.Destination","tag":"Destination","type":"*variable","value":"~*req.Caller-Destination-Number"},{"path":"*cgreq.SetupTime","tag":"SetupTime","type":"*variable","value":"~*req.Caller-Channel-Created-Time"},{"path":"*cgreq.AnswerTime","tag":"AnswerTime","type":"*variable","value":"~*req.Caller-Channel-Answered-Time"},{"path":"*cgreq.Usage","tag":"Usage","type":"*composed","value":"~*req.variable_billsec;s"},{"path":"*cgreq.Route","tag":"Route","type":"*variable","value":"~*req.variable_cgr_route"},{"path":"*cgreq.Cost","tag":"Cost","type":"*constant","value":"-1.0"},{"filters":["*notempty:*req.Hangup-Cause:"],"path":"*cgreq.DisconnectCause","tag":"DisconnectCause","type":"*variable","value":"~*req.Hangup-Cause"}],"*rar":[{"mandatory":true,"path":"*diamreq.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*diamreq.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*req.Destination-Host"},{"mandatory":true,"path":"*diamreq.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*req.Destination-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Realm","tag":"DestinationRealm","type":"*variable","value":"~*req.Origin-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Host","tag":"DestinationHost","type":"*variable","value":"~*req.Origin-Host"},{"mandatory":true,"path":"*diamreq.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"},{"path":"*diamreq.Re-Auth-Request-Type","tag":"ReAuthRequestType","type":"*constant","value":"0"}]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{TemplatesJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONHTTPAgent(t *testing.T) {
	var reply string
	expected := `{"http_agent":[]}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{HTTPAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONAnalyzer(t *testing.T) {
	var reply string
	expected := `{"analyzers":{"cleanup_interval":"1h0m0s","db_path":"/var/spool/cgrates/analyzers","ees_conns":[],"enabled":false,"index_type":"*scorch","opts":{"*exporterIDs":[]},"ttl":"24h0m0s"}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{AnalyzerSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONRateS(t *testing.T) {
	var reply string
	expected := `{"rates":{"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*intervalStart":[{"FilterIDs":null,"Tenant":""}],"*profileIDs":[],"*profileIgnoreFilters":[{"FilterIDs":null,"Tenant":""}],"*startTime":[{"FilterIDs":null,"Tenant":""}],"*usage":[{"FilterIDs":null,"Tenant":""}]},"prefix_indexed_fields":[],"rate_exists_indexed_fields":[],"rate_indexed_selects":true,"rate_nested_fields":false,"rate_notexists_indexed_fields":[],"rate_prefix_indexed_fields":[],"rate_suffix_indexed_fields":[],"suffix_indexed_fields":[],"verbosity":1000}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{RateSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONCoreS(t *testing.T) {
	var reply string
	expected := `{"cores":{"caps":10,"caps_stats_interval":"0","caps_strategy":"*busy","ees_conns":[],"shutdown_timeout":"1s"}}`
	cgrCfg := NewDefaultCGRConfig()

	cgrCfg.coreSCfg.Caps = 10
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{CoreSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v,\n received %+v", expected, reply)
	}

	var result string
	cfgCgr2 := NewDefaultCGRConfig()
	cfgCgr2.rldCh = make(chan string, 100)
	if err := cfgCgr2.V1SetConfigFromJSON(context.Background(), &SetConfigFromJSONArgs{Config: reply, DryRun: true}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Unexpected result")
	} else if cgrCfg := NewDefaultCGRConfig(); !reflect.DeepEqual(cgrCfg.CoreSCfg(), cfgCgr2.CoreSCfg()) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(cgrCfg.CoreSCfg()), utils.ToJSON(cfgCgr2.CoreSCfg()))
	}
	cfgCgr2.rldCh = make(chan string, 100)
	if err := cfgCgr2.V1SetConfigFromJSON(context.Background(), &SetConfigFromJSONArgs{Config: reply}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Unexpected result")
	} else if !reflect.DeepEqual(cgrCfg.CoreSCfg(), cfgCgr2.CoreSCfg()) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(cgrCfg.CoreSCfg()), utils.ToJSON(cfgCgr2.CoreSCfg()))
	}
}

func TestV1GetConfigAsJSONCheckConfigSanity(t *testing.T) {
	var result string
	args := `{
		"chargers": {
	        "enabled": true,
            "attributes_conns": ["*internal"]
    }
}`
	expected := `<AttributeS> not enabled but requested by <ChargerS> component`
	cfgCgr2 := NewDefaultCGRConfig()
	cfgCgr2.rldCh = make(chan string, 100)

	if err := cfgCgr2.V1SetConfigFromJSON(context.Background(), &SetConfigFromJSONArgs{Config: args}, &result); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestV1GetConfigAsJSONInvalidSection(t *testing.T) {
	var reply string
	expected := `Invalid section `
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{"InvalidSection"}}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestV1GetConfigAsJSONAllConfig(t *testing.T) {
	cfgJSON := `{
      "general": {
	      "node_id": "ENGINE1",
	  }
}`
	var reply string
	expected := `{"accounts":{"attributes_conns":[],"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"max_iterations":1000,"max_usage":"259200000000000","nested_fields":false,"notexists_indexed_fields":[],"opts":{"*profileIDs":[],"*profileIgnoreFilters":[{"FilterIDs":null,"Tenant":""}],"*usage":[{"FilterIDs":null,"Tenant":""}]},"prefix_indexed_fields":[],"rates_conns":[],"suffix_indexed_fields":[],"thresholds_conns":[]},"actions":{"accounts_conns":[],"admins_conns":[],"cdrs_conns":[],"dynaprepaid_actionprofile":[],"ees_conns":[],"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*posterAttempts":[{"FilterIDs":null,"Tenant":""}],"*profileIDs":[],"*profileIgnoreFilters":[{"FilterIDs":null,"Tenant":""}]},"prefix_indexed_fields":[],"stats_conns":[],"suffix_indexed_fields":[],"tenants":[],"thresholds_conns":[]},"admins":{"actions_conns":[],"attributes_conns":[],"caches_conns":["*internal"],"ees_conns":[],"enabled":false},"analyzers":{"cleanup_interval":"1h0m0s","db_path":"/var/spool/cgrates/analyzers","ees_conns":[],"enabled":false,"index_type":"*scorch","opts":{"*exporterIDs":[]},"ttl":"24h0m0s"},"apiban":{"enabled":false,"keys":[]},"asterisk_agent":{"asterisk_conns":[{"address":"127.0.0.1:8088","alias":"","connect_attempts":3,"max_reconnect_interval":"0s","password":"CGRateS.org","reconnects":5,"user":"cgrates"}],"create_cdr":false,"enabled":false,"sessions_conns":["*birpc_internal"]},"attributes":{"accounts_conns":[],"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*processRuns":[{"FilterIDs":null,"Tenant":""}],"*profileIDs":[],"*profileIgnoreFilters":[{"FilterIDs":null,"Tenant":""}],"*profileRuns":[{"FilterIDs":null,"Tenant":""}]},"prefix_indexed_fields":[],"resources_conns":[],"stats_conns":[],"suffix_indexed_fields":[]},"caches":{"partitions":{"*account_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*accounts":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*action_profile_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*action_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*apiban":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"2m0s"},"*attribute_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*attribute_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*caps_events":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*cdr_ids":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"10m0s"},"*charger_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*charger_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*closed_sessions":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"10s"},"*diameter_messages":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*event_charges":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"10s"},"*event_ips":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*event_resources":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*filters":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*ip_allocations":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*ip_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*ip_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*load_ids":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*radius_packets":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*ranking_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rankings":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rate_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rate_profile_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rate_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*replication_hosts":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*resource_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*resource_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*resources":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*reverse_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*route_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*route_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rpc_connections":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rpc_responses":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"2s"},"*sentrypeer":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":true,"ttl":"24h0m0s"},"*stat_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*statqueue_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*statqueues":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*stir":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*threshold_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*threshold_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*thresholds":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*trend_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*trends":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*uch":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"}},"remote_conns":[],"replication_conns":[]},"cdrs":{"accounts_conns":[],"actions_conns":[],"attributes_conns":[],"chargers_conns":[],"ees_conns":[],"enabled":false,"extra_fields":[],"online_cdr_exports":null,"opts":{"*accounts":[{"FilterIDs":null,"Tenant":""}],"*attributes":[{"FilterIDs":null,"Tenant":""}],"*chargers":[{"FilterIDs":null,"Tenant":""}],"*ees":[{"FilterIDs":null,"Tenant":""}],"*rates":[{"FilterIDs":null,"Tenant":""}],"*refund":[{"FilterIDs":null,"Tenant":""}],"*rerate":[{"FilterIDs":null,"Tenant":""}],"*stats":[{"FilterIDs":null,"Tenant":""}],"*store":[{"FilterIDs":null,"Tenant":""}],"*thresholds":[{"FilterIDs":null,"Tenant":""}]},"rates_conns":[],"session_cost_retries":5,"stats_conns":[],"thresholds_conns":[]},"chargers":{"attributes_conns":[],"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"prefix_indexed_fields":[],"suffix_indexed_fields":[]},"config_db":{"db_host":"","db_name":"","db_password":"","db_port":0,"db_type":"*internal","db_user":"","opts":{"internalDBBackupPath":"/var/lib/cgrates/internal_db/backup/configdb","internalDBDumpInterval":"0s","internalDBDumpPath":"/var/lib/cgrates/internal_db/configdb","internalDBFileSizeLimit":1073741824,"internalDBRewriteInterval":"0s","internalDBStartTimeout":"5m0s","mongoConnScheme":"mongodb","mongoQueryTimeout":"10s","redisCACertificate":"","redisClientCertificate":"","redisClientKey":"","redisCluster":false,"redisClusterOndownDelay":"0s","redisClusterSync":"5s","redisConnectAttempts":20,"redisConnectTimeout":"0s","redisMaxConns":10,"redisReadTimeout":"0s","redisSentinel":"","redisTLS":false,"redisWriteTimeout":"0s"}},"configs":{"enabled":false,"root_dir":"/var/spool/cgrates/configs","url":"/configs/"},"cores":{"caps":0,"caps_stats_interval":"0","caps_strategy":"*busy","ees_conns":[],"shutdown_timeout":"1s"},"db":{"db_conns":{"*default":{"db_host":"","db_name":"","db_password":"","db_port":0,"db_type":"*internal","db_user":"","opts":{"internalDBBackupPath":"/var/lib/cgrates/internal_db/backup/db","internalDBDumpInterval":"1m0s","internalDBDumpPath":"/var/lib/cgrates/internal_db/db","internalDBFileSizeLimit":1073741824,"internalDBRewriteInterval":"1h0m0s","internalDBStartTimeout":"5m0s","mongoConnScheme":"mongodb","mongoQueryTimeout":"10s","mysqlDSNParams":{},"mysqlLocation":"Local","pgSSLMode":"disable","redisCACertificate":"","redisClientCertificate":"","redisClientKey":"","redisCluster":false,"redisClusterOndownDelay":"0s","redisClusterSync":"5s","redisConnectAttempts":20,"redisConnectTimeout":"0s","redisMaxConns":10,"redisPoolPipelineLimit":0,"redisPoolPipelineWindow":"150µs","redisReadTimeout":"0s","redisSentinel":"","redisTLS":false,"redisWriteTimeout":"0s","sqlConnMaxLifetime":"0s","sqlLogLevel":3,"sqlMaxIdleConns":10,"sqlMaxOpenConns":100},"prefix_indexed_fields":[],"remote_conn_id":"","remote_conns":[],"replication_cache":"","replication_conns":[],"replication_failed_dir":"","replication_filtered":false,"replication_interval":"0s","string_indexed_fields":[]}},"items":{"*account_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*accounts":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*action_profile_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*action_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*attribute_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*attribute_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*cdrs":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*charger_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*charger_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*filters":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*ip_allocations":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*ip_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*ip_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*load_ids":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*ranking_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*rankings":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*rate_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*rate_profile_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*rate_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*resource_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*resource_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*resources":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*reverse_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*route_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*route_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*stat_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*statqueue_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*statqueues":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*threshold_filter_indexes":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*threshold_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*thresholds":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*trend_profiles":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*trends":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false},"*versions":{"dbConn":"*default","limit":-1,"remote":false,"replicate":false,"static_ttl":false}}},"diameter_agent":{"asr_template":"","conn_health_check_interval":"0s","dictionaries_path":"/usr/share/cgrates/diameter/dict/","enabled":false,"forced_disconnect":"*none","listen":"127.0.0.1:3868","listen_net":"tcp","origin_host":"CGR-DA","origin_realm":"cgrates.org","product_name":"CGRateS","rar_template":"","request_processors":[],"sessions_conns":["*birpc_internal"],"stats_conns":[],"synced_conn_requests":false,"thresholds_conns":[],"vendor_id":0},"dns_agent":{"enabled":false,"listeners":[{"address":"127.0.0.1:53","network":"udp"}],"request_processors":[],"sessions_conns":["*internal"],"stats_conns":[],"thresholds_conns":[],"timezone":""},"ees":{"attributes_conns":[],"cache":{"*fileCSV":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"}},"enabled":false,"exporters":[{"attempts":1,"attribute_context":"","attribute_ids":[],"blocker":false,"concurrent_requests":0,"efs_conns":["*internal"],"export_path":"/var/spool/cgrates/ees","failed_posts_dir":"/var/spool/cgrates/failed_posts","fields":[],"filters":[],"flags":[],"id":"*default","metrics_reset_schedule":"","opts":{},"synchronous":false,"timezone":"","type":"*none"}]},"efs":{"enabled":false,"failed_posts_dir":"/var/spool/cgrates/failed_posts","failed_posts_ttl":"5s","poster_attempts":3},"ers":{"ees_conns":[],"enabled":false,"partial_cache_ttl":"1s","readers":[{"cache_dump_fields":[],"concurrent_requests":1024,"ees_failed_ids":[],"ees_ids":[],"ees_success_ids":[],"fields":[{"mandatory":true,"path":"*cgreq.ToR","tag":"ToR","type":"*variable","value":"~*req.2"},{"mandatory":true,"path":"*cgreq.OriginID","tag":"OriginID","type":"*variable","value":"~*req.3"},{"mandatory":true,"path":"*cgreq.RequestType","tag":"RequestType","type":"*variable","value":"~*req.4"},{"mandatory":true,"path":"*cgreq.Tenant","tag":"Tenant","type":"*variable","value":"~*req.6"},{"mandatory":true,"path":"*cgreq.Category","tag":"Category","type":"*variable","value":"~*req.7"},{"mandatory":true,"path":"*cgreq.Account","tag":"Account","type":"*variable","value":"~*req.8"},{"mandatory":true,"path":"*cgreq.Subject","tag":"Subject","type":"*variable","value":"~*req.9"},{"mandatory":true,"path":"*cgreq.Destination","tag":"Destination","type":"*variable","value":"~*req.10"},{"mandatory":true,"path":"*cgreq.SetupTime","tag":"SetupTime","type":"*variable","value":"~*req.11"},{"mandatory":true,"path":"*cgreq.AnswerTime","tag":"AnswerTime","type":"*variable","value":"~*req.12"},{"mandatory":true,"path":"*cgreq.Usage","tag":"Usage","type":"*variable","value":"~*req.13"}],"filters":[],"flags":[],"id":"*default","max_reconnect_interval":"5m0s","opts":{"csvFieldSeparator":",","csvHeaderDefineChar":":","csvRowLength":0,"natsSubject":"cgrates_cdrs","partialCacheAction":"*none","partialOrderField":"~*req.AnswerTime"},"partial_commit_fields":[],"processed_path":"/var/spool/cgrates/ers/out","reconnects":-1,"run_delay":"0","source_path":"/var/spool/cgrates/ers/in","start_delay":"0","tenant":"","timezone":"","type":"*none"}],"sessions_conns":["*internal"],"stats_conns":[],"thresholds_conns":[]},"filters":{"accounts_conns":[],"rankings_conns":[],"resources_conns":[],"stats_conns":[],"trends_conns":[]},"freeswitch_agent":{"active_session_delimiter":",","create_cdr":false,"empty_balance_ann_file":"","empty_balance_context":"","enabled":false,"event_socket_conns":[{"address":"127.0.0.1:8021","alias":"127.0.0.1:8021","max_reconnect_interval":"0s","password":"ClueCon","reconnects":5,"reply_timeout":"1m0s"}],"extra_fields":[],"low_balance_ann_file":"","max_wait_connection":"2s","request_processors":[],"sessions_conns":["*birpc_internal"],"subscribe_park":true},"general":{"caching_delay":"0","connect_attempts":5,"connect_timeout":"1s","dbdata_encoding":"*msgpack","decimal_max_scale":0,"decimal_min_scale":0,"decimal_precision":0,"decimal_rounding_mode":"*toNearestEven","default_caching":"*reload","default_category":"call","default_request_type":"*rated","default_tenant":"cgrates.org","default_timezone":"Local","digest_equal":":","digest_separator":",","locking_timeout":"0","max_parallel_conns":100,"max_reconnect_interval":"0","node_id":"ENGINE1","opts":{"*exporterIDs":[]},"reconnects":-1,"reply_timeout":"2s","rounding_decimals":5,"tpexport_dir":"/var/spool/cgrates/tpe"},"http":{"auth_users":{},"client_opts":{"dialFallbackDelay":"300ms","dialKeepAlive":"30s","dialTimeout":"30s","disableCompression":false,"disableKeepAlives":false,"expectContinueTimeout":"0s","forceAttemptHttp2":true,"idleConnTimeout":"1m30s","maxConnsPerHost":0,"maxIdleConns":100,"maxIdleConnsPerHost":2,"responseHeaderTimeout":"0s","skipTLSVerification":false,"tlsHandshakeTimeout":"10s"},"freeswitch_cdrs_url":"/freeswitch_json","http_cdrs":"/cdr_http","json_rpc_url":"/jsonrpc","pprof_path":"/debug/pprof/","registrars_url":"/registrar","use_basic_auth":false,"ws_url":"/ws"},"http_agent":[],"ips":{"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":null,"opts":{"*allocationID":[{"FilterIDs":null,"Tenant":""}],"*ttl":[{"FilterIDs":null,"Tenant":""}]},"prefix_indexed_fields":[],"store_interval":"0s","string_indexed_fields":null,"suffix_indexed_fields":[]},"janus_agent":{"enabled":false,"janus_conns":[{"address":"127.0.0.1:8088","admin_address":"localhost:7188","admin_password":"","type":"*ws"}],"request_processors":[],"sessions_conns":["*internal"],"url":"/janus"},"kamailio_agent":{"create_cdr":false,"enabled":false,"evapi_conns":[{"address":"127.0.0.1:8448","alias":"","max_reconnect_interval":"0s","reconnects":5}],"sessions_conns":["*birpc_internal"],"timezone":""},"listen":{"http":"127.0.0.1:2080","http_tls":"127.0.0.1:2280","rpc_gob":"127.0.0.1:2013","rpc_gob_tls":"127.0.0.1:2023","rpc_json":"127.0.0.1:2012","rpc_json_tls":"127.0.0.1:2022"},"loader":{"actions_conns":["*localhost"],"caches_conns":["*localhost"],"data_path":"./","disable_reverse":false,"field_separator":",","gapi_credentials":".gapi/credentials.json","gapi_token":".gapi/token.json","tpid":""},"loaders":[{"action":"*store","cache":{"*accounts":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*action_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*attributes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*chargers":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*filters":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*ips":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*rankings":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*rate_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*resources":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*routes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*stats":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*thresholds":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*trends":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"}},"caches_conns":["*internal"],"data":[{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"new_branch":true,"path":"Rules.Type","tag":"Type","type":"*variable","value":"~*req.2"},{"path":"Rules.Element","tag":"Element","type":"*variable","value":"~*req.3"},{"path":"Rules.Values","tag":"Values","type":"*variable","value":"~*req.4"}],"file_name":"Filters.csv","flags":null,"type":"*filters"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"TenantID","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ProfileID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"new_branch":true,"path":"Attributes.FilterIDs","tag":"AttributeFilterIDs","type":"*variable","value":"~*req.5"},{"path":"Attributes.Blockers","tag":"AttributeBlockers","type":"*variable","value":"~*req.6"},{"path":"Attributes.Path","tag":"Path","type":"*variable","value":"~*req.7"},{"path":"Attributes.Type","tag":"Type","type":"*variable","value":"~*req.8"},{"path":"Attributes.Value","tag":"Value","type":"*variable","value":"~*req.9"}],"file_name":"Attributes.csv","flags":null,"type":"*attributes"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"UsageTTL","tag":"TTL","type":"*variable","value":"~*req.4"},{"path":"Limit","tag":"Limit","type":"*variable","value":"~*req.5"},{"path":"AllocationMessage","tag":"AllocationMessage","type":"*variable","value":"~*req.6"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.7"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.8"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.9"}],"file_name":"Resources.csv","flags":null,"type":"*resources"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"TTL","tag":"TTL","type":"*variable","value":"~*req.4"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.5"},{"new_branch":true,"path":"Pools.ID","tag":"PoolID","type":"*variable","value":"~*req.6"},{"path":"Pools.FilterIDs","tag":"PoolFilterIDs","type":"*variable","value":"~*req.7"},{"path":"Pools.Type","tag":"PoolType","type":"*variable","value":"~*req.8"},{"path":"Pools.Range","tag":"PoolRange","type":"*variable","value":"~*req.9"},{"path":"Pools.Strategy","tag":"PoolStrategy","type":"*variable","value":"~*req.10"},{"path":"Pools.Message","tag":"PoolMessage","type":"*variable","value":"~*req.11"},{"path":"Pools.Weights","tag":"PoolWeights","type":"*variable","value":"~*req.12"},{"path":"Pools.Blockers","tag":"PoolBlockers","type":"*variable","value":"~*req.13"}],"file_name":"IPs.csv","flags":null,"type":"*ips"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"QueueLength","tag":"QueueLength","type":"*variable","value":"~*req.5"},{"path":"TTL","tag":"TTL","type":"*variable","value":"~*req.6"},{"path":"MinItems","tag":"MinItems","type":"*variable","value":"~*req.7"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.8"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.9"},{"new_branch":true,"path":"Metrics.MetricID","tag":"MetricIDs","type":"*variable","value":"~*req.10"},{"path":"Metrics.FilterIDs","tag":"MetricFilterIDs","type":"*variable","value":"~*req.11"},{"path":"Metrics.Blockers","tag":"MetricBlockers","type":"*variable","value":"~*req.12"}],"file_name":"Stats.csv","flags":null,"type":"*stats"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"MaxHits","tag":"MaxHits","type":"*variable","value":"~*req.4"},{"path":"MinHits","tag":"MinHits","type":"*variable","value":"~*req.5"},{"path":"MinSleep","tag":"MinSleep","type":"*variable","value":"~*req.6"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.7"},{"path":"ActionProfileIDs","tag":"ActionProfileIDs","type":"*variable","value":"~*req.8"},{"path":"Async","tag":"Async","type":"*variable","value":"~*req.9"},{"path":"EeIDs","tag":"EeIDs","type":"*variable","value":"~*req.10"}],"file_name":"Thresholds.csv","flags":null,"type":"*thresholds"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"Schedule","tag":"Schedule","type":"*variable","value":"~*req.2"},{"path":"StatID","tag":"StatID","type":"*variable","value":"~*req.3"},{"path":"Metrics","tag":"Metrics","type":"*variable","value":"~*req.4"},{"path":"TTL","tag":"TTL","type":"*variable","value":"~*req.5"},{"path":"QueueLength","tag":"QueueLength","type":"*variable","value":"~*req.6"},{"path":"MinItems","tag":"MinItems","type":"*variable","value":"~*req.7"},{"path":"CorrelationType","tag":"CorrelationType","type":"*variable","value":"~*req.8"},{"path":"Tolerance","tag":"Tolerance","type":"*variable","value":"~*req.9"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.10"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.11"}],"file_name":"Trends.csv","flags":null,"type":"*trends"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"Schedule","tag":"Schedule","type":"*variable","value":"~*req.2"},{"path":"StatIDs","tag":"StatIDs","type":"*variable","value":"~*req.3"},{"path":"MetricIDs","tag":"MetricIDs","type":"*variable","value":"~*req.4"},{"path":"Sorting","tag":"Sorting","type":"*variable","value":"~*req.5"},{"path":"SortingParameters","tag":"SortingParameters","type":"*variable","value":"~*req.6"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.7"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.8"}],"file_name":"Rankings.csv","flags":null,"type":"*rankings"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Sorting","tag":"Sorting","type":"*variable","value":"~*req.5"},{"path":"SortingParameters","tag":"SortingParameters","type":"*variable","value":"~*req.6"},{"new_branch":true,"path":"Routes.ID","tag":"RouteID","type":"*variable","value":"~*req.7"},{"path":"Routes.FilterIDs","tag":"RouteFilterIDs","type":"*variable","value":"~*req.8"},{"path":"Routes.AccountIDs","tag":"RouteAccountIDs","type":"*variable","value":"~*req.9"},{"path":"Routes.RateProfileIDs","tag":"RouteRateProfileIDs","type":"*variable","value":"~*req.10"},{"path":"Routes.ResourceIDs","tag":"RouteResourceIDs","type":"*variable","value":"~*req.11"},{"path":"Routes.StatIDs","tag":"RouteStatIDs","type":"*variable","value":"~*req.12"},{"path":"Routes.Weights","tag":"RouteWeights","type":"*variable","value":"~*req.13"},{"path":"Routes.Blockers","tag":"RouteBlockers","type":"*variable","value":"~*req.14"},{"path":"Routes.RouteParameters","tag":"RouteParameters","type":"*variable","value":"~*req.15"}],"file_name":"Routes.csv","flags":null,"type":"*routes"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"RunID","tag":"RunID","type":"*variable","value":"~*req.5"},{"path":"AttributeIDs","tag":"AttributeIDs","type":"*variable","value":"~*req.6"}],"file_name":"Chargers.csv","flags":null,"type":"*chargers"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"MinCost","tag":"MinCost","type":"*variable","value":"~*req.4"},{"path":"MaxCost","tag":"MaxCost","type":"*variable","value":"~*req.5"},{"path":"MaxCostStrategy","tag":"MaxCostStrategy","type":"*variable","value":"~*req.6"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].FilterIDs","tag":"RateFilterIDs","type":"*variable","value":"~*req.8"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].ActivationTimes","tag":"RateActivationTimes","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].Weights","tag":"RateWeights","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].Blocker","tag":"RateBlocker","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.7:"],"new_branch":true,"path":"Rates[\u003c~*req.7\u003e].IntervalRates.IntervalStart","tag":"RateIntervalStart","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.FixedFee","tag":"RateFixedFee","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.RecurrentFee","tag":"RateRecurrentFee","type":"*variable","value":"~*req.14"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.Unit","tag":"RateUnit","type":"*variable","value":"~*req.15"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.Increment","tag":"RateIncrement","type":"*variable","value":"~*req.16"}],"file_name":"Rates.csv","flags":null,"type":"*rate_profiles"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Schedule","tag":"Schedule","type":"*variable","value":"~*req.5"},{"path":"Targets[\u003c~*req.6\u003e]","tag":"TargetIDs","type":"*variable","value":"~*req.7"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].FilterIDs","tag":"ActionFilterIDs","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].TTL","tag":"ActionTTL","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Type","tag":"ActionType","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Opts","tag":"ActionOpts","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Weights","tag":"ActionWeights","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Blockers","tag":"ActionBlockers","type":"*variable","value":"~*req.14"},{"filters":["*notempty:~*req.8:"],"new_branch":true,"path":"Actions[\u003c~*req.8\u003e].Diktats.ID","tag":"ActionDiktatsID","type":"*variable","value":"~*req.15"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.FilterIDs","tag":"ActionDiktatsFilterIDs","type":"*variable","value":"~*req.16"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.Opts","tag":"ActionDiktatsOpts","type":"*variable","value":"~*req.17"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.Weights","tag":"ActionDiktatsWeights","type":"*variable","value":"~*req.18"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.Blockers","tag":"ActionDiktatsBlockers","type":"*variable","value":"~*req.19"}],"file_name":"Actions.csv","flags":null,"type":"*action_profiles"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Opts","tag":"Opts","type":"*variable","value":"~*req.5"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].FilterIDs","tag":"BalanceFilterIDs","type":"*variable","value":"~*req.7"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Weights","tag":"BalanceWeights","type":"*variable","value":"~*req.8"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Blockers","tag":"BalanceBlockers","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Type","tag":"BalanceType","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Units","tag":"BalanceUnits","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].UnitFactors","tag":"BalanceUnitFactors","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Opts","tag":"BalanceOpts","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].CostIncrements","tag":"BalanceCostIncrements","type":"*variable","value":"~*req.14"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].AttributeIDs","tag":"BalanceAttributeIDs","type":"*variable","value":"~*req.15"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].RateProfileIDs","tag":"BalanceRateProfileIDs","type":"*variable","value":"~*req.16"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.17"}],"file_name":"Accounts.csv","flags":null,"type":"*accounts"}],"enabled":false,"field_separator":",","id":"*default","lockfile_path":".cgr.lck","opts":{"*cache":"","*forceLock":false,"*stopOnError":false,"*withIndex":true},"run_delay":"0","tenant":"","tp_in_dir":"/var/spool/cgrates/loader/in","tp_out_dir":"/var/spool/cgrates/loader/out"}],"logger":{"efs_conns":["*internal"],"level":6,"opts":{"failed_posts_dir":"/var/spool/cgrates/failed_posts","kafka_attempts":1,"kafka_conn":"","kafka_topic":""},"type":"*syslog"},"migrator":{"fromItems":{"*accounts":{"dbConn":"*default"},"*charger_profiles":{"dbConn":"*default"},"*filters":{"dbConn":"*default"},"*load_ids":{"dbConn":"*default"},"*statqueue_profiles":{"dbConn":"*default"},"*versions":{"dbConn":"*default"}},"out_db_opts":{"mongoConnScheme":"mongodb","mongoQueryTimeout":"10s","redisCACertificate":"","redisClientCertificate":"","redisClientKey":"","redisCluster":false,"redisClusterOndownDelay":"0s","redisClusterSync":"5s","redisConnectAttempts":20,"redisConnectTimeout":"0s","redisMaxConns":10,"redisPoolPipelineLimit":0,"redisPoolPipelineWindow":"150µs","redisReadTimeout":"0s","redisSentinel":"","redisTLS":false,"redisWriteTimeout":"0s"},"users_filters":null},"prometheus_agent":{"admins_conns":[],"cache_ids":[],"caches_conns":[],"collect_go_metrics":false,"collect_process_metrics":false,"cores_conns":[],"enabled":false,"path":"/prometheus","stat_queue_ids":[],"stats_conns":[]},"radius_agent":{"client_da_addresses":{},"client_dictionaries":{"*default":["/usr/share/cgrates/radius/dict/"]},"client_secrets":{"*default":"CGRateS.org"},"coa_template":"*coa","dmr_template":"*dmr","enabled":false,"listeners":[{"acct_address":"127.0.0.1:1813","auth_address":"127.0.0.1:1812","network":"udp"}],"request_processors":[],"requests_cache_key":"","sessions_conns":["*internal"],"stats_conns":[],"thresholds_conns":[]},"rankings":{"ees_conns":[],"ees_exporter_ids":null,"enabled":false,"scheduled_ids":{},"stats_conns":[],"store_interval":"","thresholds_conns":[]},"rates":{"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*intervalStart":[{"FilterIDs":null,"Tenant":""}],"*profileIDs":[],"*profileIgnoreFilters":[{"FilterIDs":null,"Tenant":""}],"*startTime":[{"FilterIDs":null,"Tenant":""}],"*usage":[{"FilterIDs":null,"Tenant":""}]},"prefix_indexed_fields":[],"rate_exists_indexed_fields":[],"rate_indexed_selects":true,"rate_nested_fields":false,"rate_notexists_indexed_fields":[],"rate_prefix_indexed_fields":[],"rate_suffix_indexed_fields":[],"suffix_indexed_fields":[],"verbosity":1000},"registrarc":{"rpc":{"hosts":[],"refresh_interval":"5m0s","registrars_conns":[]}},"resources":{"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*units":[{"FilterIDs":null,"Tenant":""}],"*usageID":[{"FilterIDs":null,"Tenant":""}],"*usageTTL":[{"FilterIDs":null,"Tenant":""}]},"prefix_indexed_fields":[],"store_interval":"","suffix_indexed_fields":[],"thresholds_conns":[]},"routes":{"accounts_conns":[],"attributes_conns":[],"default_ratio":1,"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*context":[{"FilterIDs":null,"Tenant":""}],"*ignoreErrors":[{"FilterIDs":null,"Tenant":""}],"*limit":[],"*maxCost":[{"Tenant":"","Value":""}],"*maxItems":[],"*offset":[],"*profileCount":[{"FilterIDs":null,"Tenant":""}],"*usage":[{"FilterIDs":null,"Tenant":""}]},"prefix_indexed_fields":[],"rates_conns":[],"resources_conns":[],"stats_conns":[],"suffix_indexed_fields":[]},"rpc_conns":{"*bijson_localhost":{"conns":[{"address":"127.0.0.1:2014","transport":"*birpc_json"}],"poolSize":0,"strategy":"*first"},"*birpc_internal":{"conns":[{"address":"*birpc_internal","transport":""}],"poolSize":0,"strategy":"*first"},"*internal":{"conns":[{"address":"*internal","transport":""}],"poolSize":0,"strategy":"*first"},"*localhost":{"conns":[{"address":"127.0.0.1:2012","transport":"*json"}],"poolSize":0,"strategy":"*first"}},"sentrypeer":{"audience":"https://sentrypeer.com/api","client_id":"","client_secret":"","grant_type":"client_credentials","ips_url":"https://sentrypeer.com/api/ip-addresses","numbers_url":"https://sentrypeer.com/api/phone-numbers","token_url":"https://authz.sentrypeer.com/oauth/token"},"sessions":{"accounts_conns":[],"actions_conns":[],"alterable_fields":[],"attributes_conns":[],"cdrs_conns":[],"channel_sync_interval":"0","chargers_conns":[],"client_protocol":1,"default_usage":{"*any":"3h0m0s","*data":"1048576","*sms":"1","*voice":"3h0m0s"},"enabled":false,"ips_conns":[],"listen_bigob":"","listen_bijson":"127.0.0.1:2014","min_dur_low_balance":"0","opts":{"*accounts":[{"FilterIDs":null,"Tenant":""}],"*accountsForceUsage":[],"*attributes":[{"FilterIDs":null,"Tenant":""}],"*attributesDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*blockerError":[{"FilterIDs":null,"Tenant":""}],"*cdrs":[{"FilterIDs":null,"Tenant":""}],"*cdrsDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*chargeable":[{"FilterIDs":null,"Tenant":""}],"*chargers":[{"FilterIDs":null,"Tenant":""}],"*debitInterval":[{"FilterIDs":null,"Tenant":""}],"*forceUsage":[],"*initiate":[{"FilterIDs":null,"Tenant":""}],"*ips":[{"FilterIDs":null,"Tenant":""}],"*ipsAllocate":[{"FilterIDs":null,"Tenant":""}],"*ipsAuthorize":[{"FilterIDs":null,"Tenant":""}],"*ipsRelease":[{"FilterIDs":null,"Tenant":""}],"*maxUsage":[{"FilterIDs":null,"Tenant":""}],"*message":[{"FilterIDs":null,"Tenant":""}],"*originID":[],"*resources":[{"FilterIDs":null,"Tenant":""}],"*resourcesAllocate":[{"FilterIDs":null,"Tenant":""}],"*resourcesAuthorize":[{"FilterIDs":null,"Tenant":""}],"*resourcesDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*resourcesRelease":[{"FilterIDs":null,"Tenant":""}],"*routes":[{"FilterIDs":null,"Tenant":""}],"*routesDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*stats":[{"FilterIDs":null,"Tenant":""}],"*statsDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*terminate":[{"FilterIDs":null,"Tenant":""}],"*thresholds":[{"FilterIDs":null,"Tenant":""}],"*thresholdsDerivedReply":[{"FilterIDs":null,"Tenant":""}],"*ttl":[{"FilterIDs":null,"Tenant":""}],"*ttlLastUsage":[],"*ttlLastUsed":[],"*ttlMaxDelay":[{"FilterIDs":null,"Tenant":""}],"*ttlUsage":[],"*update":[{"FilterIDs":null,"Tenant":""}]},"rates_conns":[],"replication_conns":[],"resources_conns":[],"routes_conns":[],"session_indexes":[],"stats_conns":[],"stir":{"allowed_attest":["*any"],"default_attest":"A","payload_maxduration":"-1","privatekey_path":"","publickey_path":""},"store_session_costs":false,"terminate_attempts":5,"thresholds_conns":[]},"sip_agent":{"enabled":false,"listen":"127.0.0.1:5060","listen_net":"udp","request_processors":[],"retransmission_timer":"1s","sessions_conns":["*internal"],"stats_conns":[],"thresholds_conns":[],"timezone":""},"stats":{"ees_conns":[],"ees_exporter_ids":null,"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*profileIDs":[],"*profileIgnoreFilters":[{"FilterIDs":null,"Tenant":""}],"*roundingDecimals":[]},"prefix_indexed_fields":[],"store_interval":"","store_uncompressed_limit":0,"suffix_indexed_fields":[],"thresholds_conns":[]},"suretax":{"bill_to_number":"","business_unit":"","client_number":"","client_tracking":"~*opts.*originID","customer_number":"~*req.Subject","include_local_cost":false,"orig_number":"~*req.Subject","p2pplus4":"","p2pzipcode":"","plus4":"","regulatory_code":"03","response_group":"03","response_type":"D4","return_file_code":"0","sales_type_code":"R","tax_exemption_code_list":"","tax_included":"0","tax_situs_rule":"04","term_number":"~*req.Destination","timezone":"UTC","trans_type_code":"010101","unit_type":"00","units":"1","url":"","validation_key":"","zipcode":""},"templates":{"*asr":[{"mandatory":true,"path":"*diamreq.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*diamreq.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*req.Destination-Host"},{"mandatory":true,"path":"*diamreq.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*req.Destination-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Realm","tag":"DestinationRealm","type":"*variable","value":"~*req.Origin-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Host","tag":"DestinationHost","type":"*variable","value":"~*req.Origin-Host"},{"mandatory":true,"path":"*diamreq.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"}],"*cca":[{"mandatory":true,"path":"*rep.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"path":"*rep.Result-Code","tag":"ResultCode","type":"*constant","value":"2001"},{"mandatory":true,"path":"*rep.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*vars.OriginHost"},{"mandatory":true,"path":"*rep.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*vars.OriginRealm"},{"mandatory":true,"path":"*rep.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"},{"mandatory":true,"path":"*rep.CC-Request-Type","tag":"CCRequestType","type":"*variable","value":"~*req.CC-Request-Type"},{"mandatory":true,"path":"*rep.CC-Request-Number","tag":"CCRequestNumber","type":"*variable","value":"~*req.CC-Request-Number"}],"*cdrLog":[{"mandatory":true,"path":"*cdr.ToR","tag":"ToR","type":"*variable","value":"~*req.BalanceType"},{"mandatory":true,"path":"*cdr.OriginHost","tag":"OriginHost","type":"*constant","value":"127.0.0.1"},{"mandatory":true,"path":"*cdr.RequestType","tag":"RequestType","type":"*constant","value":"*none"},{"mandatory":true,"path":"*cdr.Tenant","tag":"Tenant","type":"*variable","value":"~*req.Tenant"},{"mandatory":true,"path":"*cdr.Account","tag":"Account","type":"*variable","value":"~*req.Account"},{"mandatory":true,"path":"*cdr.Subject","tag":"Subject","type":"*variable","value":"~*req.Account"},{"mandatory":true,"path":"*cdr.Cost","tag":"Cost","type":"*variable","value":"~*req.Cost"},{"mandatory":true,"path":"*cdr.Source","tag":"Source","type":"*constant","value":"*cdrLog"},{"mandatory":true,"path":"*cdr.Usage","tag":"Usage","type":"*constant","value":"1"},{"mandatory":true,"path":"*cdr.RunID","tag":"RunID","type":"*variable","value":"~*req.ActionType"},{"mandatory":true,"path":"*cdr.SetupTime","tag":"SetupTime","type":"*constant","value":"*now"},{"mandatory":true,"path":"*cdr.AnswerTime","tag":"AnswerTime","type":"*constant","value":"*now"},{"mandatory":true,"path":"*cdr.PreRated","tag":"PreRated","type":"*constant","value":"true"}],"*coa":[{"path":"*radDAReq.User-Name","tag":"User-Name","type":"*variable","value":"~*oreq.User-Name"},{"path":"*radDAReq.NAS-IP-Address","tag":"NAS-IP-Address","type":"*variable","value":"~*oreq.NAS-IP-Address"},{"path":"*radDAReq.Acct-Session-Id","tag":"Acct-Session-Id","type":"*variable","value":"~*oreq.Acct-Session-Id"},{"path":"*radDAReq.Filter-Id","tag":"Filter-Id","type":"*variable","value":"~*req.CustomFilter"}],"*dmr":[{"path":"*radDAReq.User-Name","tag":"User-Name","type":"*variable","value":"~*oreq.User-Name"},{"path":"*radDAReq.NAS-IP-Address","tag":"NAS-IP-Address","type":"*variable","value":"~*oreq.NAS-IP-Address"},{"path":"*radDAReq.Acct-Session-Id","tag":"Acct-Session-Id","type":"*variable","value":"~*oreq.Acct-Session-Id"},{"path":"*radDAReq.Reply-Message","tag":"Reply-Message","type":"*variable","value":"~*req.DisconnectCause"}],"*err":[{"mandatory":true,"path":"*rep.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*rep.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*vars.OriginHost"},{"mandatory":true,"path":"*rep.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*vars.OriginRealm"}],"*errSip":[{"mandatory":true,"path":"*rep.Request","tag":"Request","type":"*constant","value":"SIP/2.0 500 Internal Server Error"}],"*fsa":[{"path":"*cgreq.ToR","tag":"ToR","type":"*constant","value":"*voice"},{"path":"*cgreq.PDD","tag":"PDD","type":"*composed","value":"~*req.variable_progress_mediamsec;ms"},{"path":"*cgreq.ACD","tag":"ACD","type":"*composed","value":"~*req.variable_cdr_acd;s"},{"path":"*cgreq.OriginID","tag":"OriginID","type":"*variable","value":"~*req.Unique-ID"},{"path":"*opts.*originID","tag":"*originID","type":"*variable","value":"~*req.Unique-ID"},{"path":"*cgreq.OriginHost","tag":"OriginHost","type":"*variable","value":"~*req.variable_cgr_originhost"},{"path":"*cgreq.Account","tag":"Account","type":"*variable","value":"~*req.Caller-Username"},{"path":"*cgreq.Source","tag":"Source","type":"*composed","value":"FS_;~*req.Event-Name"},{"filters":["*string:*req.variable_process_cdr:false"],"path":"*cgreq.RequestType","tag":"RequestType","type":"*constant","value":"*none"},{"filters":["*string:*req.Caller-Dialplan:inline"],"path":"*cgreq.RequestType","tag":"RequestType","type":"*constant","value":"*none"},{"filters":["*exists:*cgreq.RequestType:"],"path":"*cgreq.RequestType","tag":"RequestType","type":"*constant","value":"*prepaid"},{"path":"*cgreq.Tenant","tag":"Tenant","type":"*constant","value":"cgrates.org"},{"path":"*cgreq.Category","tag":"Category","type":"*constant","value":"call"},{"path":"*cgreq.Subject","tag":"Subject","type":"*variable","value":"~*req.Caller-Username"},{"path":"*cgreq.Destination","tag":"Destination","type":"*variable","value":"~*req.Caller-Destination-Number"},{"path":"*cgreq.SetupTime","tag":"SetupTime","type":"*variable","value":"~*req.Caller-Channel-Created-Time"},{"path":"*cgreq.AnswerTime","tag":"AnswerTime","type":"*variable","value":"~*req.Caller-Channel-Answered-Time"},{"path":"*cgreq.Usage","tag":"Usage","type":"*composed","value":"~*req.variable_billsec;s"},{"path":"*cgreq.Route","tag":"Route","type":"*variable","value":"~*req.variable_cgr_route"},{"path":"*cgreq.Cost","tag":"Cost","type":"*constant","value":"-1.0"},{"filters":["*notempty:*req.Hangup-Cause:"],"path":"*cgreq.DisconnectCause","tag":"DisconnectCause","type":"*variable","value":"~*req.Hangup-Cause"}],"*rar":[{"mandatory":true,"path":"*diamreq.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*diamreq.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*req.Destination-Host"},{"mandatory":true,"path":"*diamreq.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*req.Destination-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Realm","tag":"DestinationRealm","type":"*variable","value":"~*req.Origin-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Host","tag":"DestinationHost","type":"*variable","value":"~*req.Origin-Host"},{"mandatory":true,"path":"*diamreq.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"},{"path":"*diamreq.Re-Auth-Request-Type","tag":"ReAuthRequestType","type":"*constant","value":"0"}]},"thresholds":{"actions_conns":[],"ees_conns":[],"ees_exporter_ids":[],"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*profileIDs":[],"*profileIgnoreFilters":[{"FilterIDs":null,"Tenant":""}]},"prefix_indexed_fields":[],"store_interval":"","suffix_indexed_fields":[]},"tls":{"ca_certificate":"","client_certificate":"","client_key":"","server_certificate":"","server_key":"","server_name":"","server_policy":4},"tpes":{"enabled":false},"trends":{"ees_conns":[],"ees_exporter_ids":null,"enabled":false,"scheduled_ids":{},"stats_conns":[],"store_interval":"","store_uncompressed_limit":0,"thresholds_conns":[]}}`

	cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSON)
	if err != nil {
		t.Fatal(err)
	}
	cgrCfg.SureTaxCfg().Timezone = time.UTC
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{}, &reply); err != nil {
		t.Fatal(err)
	} else if expected != reply {
		t.Fatalf("Expected %+v \n, received %+v", expected, reply)
	}
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{}, &reply); err != nil {
		t.Fatal(err)
	} else if expected != reply {
		t.Fatalf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1ReloadConfigFromJSONEmptyConfig(t *testing.T) {
	var reply string
	cgrCfg := NewDefaultCGRConfig()
	cgrCfg.rldCh = make(chan string, 100)
	if err := cgrCfg.V1SetConfigFromJSON(context.Background(), &SetConfigFromJSONArgs{Config: utils.EmptyString}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply")
	}
}

func TestV1ReloadConfigFromJSONInvalidSection(t *testing.T) {
	var reply string
	expected := "invalid character 'I' looking for beginning of value around line 1 and position 1\n line: \"InvalidSection\""
	cgrCfg := NewDefaultCGRConfig()
	cgrCfg.rldCh = make(chan string, 100)
	if err := cgrCfg.V1SetConfigFromJSON(context.Background(), &SetConfigFromJSONArgs{Config: "InvalidSection"}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestCgrCdfEventReader(t *testing.T) {
	eCfg := &ERsCfg{
		Enabled:         false,
		SessionSConns:   []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		EEsConns:        []string{},
		StatSConns:      []string{},
		ThresholdSConns: []string{},
		Readers: []*EventReaderCfg{
			{
				ID:                   utils.MetaDefault,
				Type:                 utils.MetaNone,
				RunDelay:             0,
				ConcurrentReqs:       1024,
				SourcePath:           "/var/spool/cgrates/ers/in",
				ProcessedPath:        "/var/spool/cgrates/ers/out",
				Tenant:               nil,
				Timezone:             utils.EmptyString,
				EEsIDs:               []string{},
				EEsSuccessIDs:        []string{},
				EEsFailedIDs:         []string{},
				Filters:              []string{},
				Reconnects:           -1,
				MaxReconnectInterval: 5 * time.Minute,
				Flags:                utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
						Value: utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: utils.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: utils.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: utils.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: utils.NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     []*FCTemplate{},
				PartialCommitFields: []*FCTemplate{},
				Opts: &EventReaderOpts{
					CSVFieldSeparator:   utils.StringPointer(","),
					CSVHeaderDefineChar: utils.StringPointer(":"),
					CSVRowLength:        utils.IntPointer(0),
					PartialOrderField:   utils.StringPointer("~*req.AnswerTime"),
					PartialCacheAction:  utils.StringPointer(utils.MetaNone),
					NATSSubject:         utils.StringPointer("cgrates_cdrs"),
				},
			},
		},
		PartialCacheTTL: time.Second,
	}
	for _, profile := range eCfg.Readers {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
	}
	if !reflect.DeepEqual(cgrCfg.ersCfg, eCfg) {
		t.Errorf("received: %+v,\n expecting: %+v", utils.ToJSON(cgrCfg.ersCfg), utils.ToJSON(eCfg))
	}
}

func TestCgrCdfEventExporter(t *testing.T) {
	eCfg := &EEsCfg{
		Enabled:         false,
		AttributeSConns: []string{},
		Cache: map[string]*CacheParamCfg{
			utils.MetaFileCSV: {
				Limit: -1,
				TTL:   5 * time.Second,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.MetaNone,
				ExportPath:     "/var/spool/cgrates/ees",
				Attempts:       1,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				AttributeSIDs:  []string{},
				Flags:          utils.FlagsWithParams{},
				Fields:         []*FCTemplate{},
				contentFields:  []*FCTemplate{},
				headerFields:   []*FCTemplate{},
				trailerFields:  []*FCTemplate{},
				Opts:           &EventExporterOpts{},
				FailedPostsDir: "/var/spool/cgrates/failed_posts",
				EFsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEFs)},
			},
		},
	}
	if !reflect.DeepEqual(cgrCfg.eesCfg, eCfg) {
		t.Errorf("received: %+v,\n expecting: %+v", utils.ToJSON(cgrCfg.eesCfg), utils.ToJSON(eCfg))
	}
}

func TestRpcConnsDefaults(t *testing.T) {
	eCfg := make(RPCConns)
	// hardoded the *internal and *localhost connections
	eCfg[utils.MetaBiJSONLocalHost] = &RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*RemoteHost{{
			Address:   "127.0.0.1:2014",
			Transport: rpcclient.BiRPCJSON,
		}},
	}
	eCfg[utils.MetaInternal] = &RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*RemoteHost{
			{
				Address: utils.MetaInternal,
			},
		},
	}
	eCfg[rpcclient.BiRPCInternal] = &RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*RemoteHost{
			{
				Address: rpcclient.BiRPCInternal,
			},
		},
	}
	eCfg[utils.MetaLocalHost] = &RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*RemoteHost{
			{
				Address:   "127.0.0.1:2012",
				Transport: utils.MetaJSON,
			},
		},
	}
	if !reflect.DeepEqual(cgrCfg.rpcConns, eCfg) {
		t.Errorf("received: %+v,\n expecting: %+v", utils.ToJSON(cgrCfg.rpcConns), utils.ToJSON(eCfg))
	}
}

func TestCgrCfgJSONDefaultsConfigS(t *testing.T) {
	eCfg := &ConfigSCfg{
		Enabled: false,
		URL:     "/configs/",
		RootDir: "/var/spool/cgrates/configs",
	}
	if !reflect.DeepEqual(cgrCfg.configSCfg, eCfg) {
		t.Errorf("received: %+v, expecting: %+v", utils.ToJSON(cgrCfg.configSCfg), utils.ToJSON(eCfg))
	}
}

func TestLoadConfigFromReaderError(t *testing.T) {
	expectedErrFile := "open randomfile.go: no such file or directory"
	file, err := os.Open("randomfile.go")
	expectedErr := "invalid argument"
	cgrCfg := NewDefaultCGRConfig()
	if err == nil || err.Error() != expectedErrFile {
		t.Errorf("Expected %+v, receivewd %+v", expectedErrFile, err)
	} else if err := loadConfigFromReader(context.Background(), file, cgrCfg.sections, true, cgrCfg); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadConfigFromReaderLoadFunctionsError(t *testing.T) {
	cfgJSONStr := `{
     "db": {				
	 	"db_conns": {
			"*default": {	
				"db_type": 123		
			},
		},				
     }
}`
	expected := `json: cannot unmarshal number into Go struct field DbConnJson.Db_conns.Db_type of type string`
	cgrCfg := NewDefaultCGRConfig()
	if err := loadConfigFromReader(context.Background(), strings.NewReader(cfgJSONStr),
		cgrCfg.sections, true, cgrCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadCfgFromJSONWithLocksInvalidSeciton(t *testing.T) {
	expected := "Invalid section: <invalidSection> "
	cfg := NewDefaultCGRConfig()
	if err := cfg.loadCfgWithLocks(context.Background(), "/random/path", "invalidSection"); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestCGRConfigClone(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	rcv := cfg.Clone()
	cfg.rldCh = nil
	rcv.rldCh = nil
	cfg.lks = nil
	rcv.lks = nil
	if !reflect.DeepEqual(cfg.AsMapInterface(),
		rcv.AsMapInterface()) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg), utils.ToJSON(rcv))
	}
	if !reflect.DeepEqual(cfg.loaderCfg, rcv.loaderCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.loaderCfg), utils.ToJSON(rcv.loaderCfg))
	}
	if !reflect.DeepEqual(cfg.httpAgentCfg, rcv.httpAgentCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.httpAgentCfg), utils.ToJSON(rcv.httpAgentCfg))
	}
	if !reflect.DeepEqual(cfg.rpcConns, rcv.rpcConns) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.rpcConns), utils.ToJSON(rcv.rpcConns))
	}
	if !reflect.DeepEqual(cfg.templates, rcv.templates) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.templates), utils.ToJSON(rcv.templates))
	}
	if !reflect.DeepEqual(cfg.generalCfg, rcv.generalCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.generalCfg), utils.ToJSON(rcv.generalCfg))
	}
	if !reflect.DeepEqual(cfg.dbCfg, rcv.dbCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.dbCfg), utils.ToJSON(rcv.dbCfg))
	}
	if !reflect.DeepEqual(cfg.tlsCfg, rcv.tlsCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.tlsCfg), utils.ToJSON(rcv.tlsCfg))
	}
	if !reflect.DeepEqual(cfg.cacheCfg, rcv.cacheCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.cacheCfg), utils.ToJSON(rcv.cacheCfg))
	}
	if !reflect.DeepEqual(cfg.listenCfg, rcv.listenCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.listenCfg), utils.ToJSON(rcv.listenCfg))
	}
	if !reflect.DeepEqual(cfg.httpCfg.AsMapInterface(),
		rcv.httpCfg.AsMapInterface()) {
		t.Errorf("Expected: %+v\nReceived: %+v",
			utils.ToJSON(cfg.httpCfg.AsMapInterface()),
			utils.ToJSON(rcv.httpCfg.AsMapInterface()))
	}
	if !reflect.DeepEqual(cfg.filterSCfg, rcv.filterSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.filterSCfg), utils.ToJSON(rcv.filterSCfg))
	}
	if !reflect.DeepEqual(cfg.cdrsCfg, rcv.cdrsCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.cdrsCfg), utils.ToJSON(rcv.cdrsCfg))
	}
	if !reflect.DeepEqual(cfg.sessionSCfg, rcv.sessionSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.sessionSCfg), utils.ToJSON(rcv.sessionSCfg))
	}
	if !reflect.DeepEqual(cfg.fsAgentCfg, rcv.fsAgentCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.fsAgentCfg), utils.ToJSON(rcv.fsAgentCfg))
	}
	if !reflect.DeepEqual(cfg.kamAgentCfg, rcv.kamAgentCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.kamAgentCfg), utils.ToJSON(rcv.kamAgentCfg))
	}
	if !reflect.DeepEqual(cfg.asteriskAgentCfg, rcv.asteriskAgentCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.asteriskAgentCfg), utils.ToJSON(rcv.asteriskAgentCfg))
	}
	if !reflect.DeepEqual(cfg.diameterAgentCfg, rcv.diameterAgentCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.diameterAgentCfg), utils.ToJSON(rcv.diameterAgentCfg))
	}
	if !reflect.DeepEqual(cfg.radiusAgentCfg, rcv.radiusAgentCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.radiusAgentCfg), utils.ToJSON(rcv.radiusAgentCfg))
	}
	if !reflect.DeepEqual(cfg.dnsAgentCfg, rcv.dnsAgentCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.dnsAgentCfg), utils.ToJSON(rcv.dnsAgentCfg))
	}
	if !reflect.DeepEqual(cfg.attributeSCfg, rcv.attributeSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.attributeSCfg), utils.ToJSON(rcv.attributeSCfg))
	}
	if !reflect.DeepEqual(cfg.chargerSCfg, rcv.chargerSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.chargerSCfg), utils.ToJSON(rcv.chargerSCfg))
	}
	if !reflect.DeepEqual(cfg.resourceSCfg, rcv.resourceSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.resourceSCfg), utils.ToJSON(rcv.resourceSCfg))
	}
	if !reflect.DeepEqual(cfg.statsCfg, rcv.statsCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.statsCfg), utils.ToJSON(rcv.statsCfg))
	}
	if !reflect.DeepEqual(cfg.thresholdSCfg, rcv.thresholdSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.thresholdSCfg), utils.ToJSON(rcv.thresholdSCfg))
	}
	if !reflect.DeepEqual(cfg.routeSCfg, rcv.routeSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.routeSCfg), utils.ToJSON(rcv.routeSCfg))
	}
	if !reflect.DeepEqual(cfg.sureTaxCfg, rcv.sureTaxCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.sureTaxCfg), utils.ToJSON(rcv.sureTaxCfg))
	}
	if !reflect.DeepEqual(cfg.registrarCCfg, rcv.registrarCCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.registrarCCfg), utils.ToJSON(rcv.registrarCCfg))
	}
	if !reflect.DeepEqual(cfg.loaderCgrCfg, rcv.loaderCgrCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.loaderCgrCfg), utils.ToJSON(rcv.loaderCgrCfg))
	}
	if !reflect.DeepEqual(cfg.migratorCgrCfg, rcv.migratorCgrCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.migratorCgrCfg), utils.ToJSON(rcv.migratorCgrCfg))
	}
	if !reflect.DeepEqual(cfg.analyzerSCfg, rcv.analyzerSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.analyzerSCfg), utils.ToJSON(rcv.analyzerSCfg))
	}
	if !reflect.DeepEqual(cfg.admS, rcv.admS) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.admS), utils.ToJSON(rcv.admS))
	}
	if !reflect.DeepEqual(cfg.ersCfg, rcv.ersCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.ersCfg), utils.ToJSON(rcv.ersCfg))
	}
	if !reflect.DeepEqual(cfg.eesCfg, rcv.eesCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.eesCfg), utils.ToJSON(rcv.eesCfg))
	}
	if !reflect.DeepEqual(cfg.rateSCfg, rcv.rateSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.rateSCfg), utils.ToJSON(rcv.rateSCfg))
	}
	if !reflect.DeepEqual(cfg.sipAgentCfg, rcv.sipAgentCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.sipAgentCfg), utils.ToJSON(rcv.sipAgentCfg))
	}
	if !reflect.DeepEqual(cfg.configSCfg, rcv.configSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.configSCfg), utils.ToJSON(rcv.configSCfg))
	}
	if !reflect.DeepEqual(cfg.apiBanCfg, rcv.apiBanCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.apiBanCfg), utils.ToJSON(rcv.apiBanCfg))
	}
	if !reflect.DeepEqual(cfg.coreSCfg, rcv.coreSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.coreSCfg), utils.ToJSON(rcv.coreSCfg))
	}

}

func TestActionSConfig(t *testing.T) {
	expected := &ActionSCfg{
		Enabled:                  false,
		EEsConns:                 []string{},
		CDRsConns:                []string{},
		ThresholdSConns:          []string{},
		StatSConns:               []string{},
		AccountSConns:            []string{},
		AdminSConns:              []string{},
		IndexedSelects:           true,
		Tenants:                  &[]string{},
		StringIndexedFields:      nil,
		PrefixIndexedFields:      &[]string{},
		SuffixIndexedFields:      &[]string{},
		ExistsIndexedFields:      &[]string{},
		NotExistsIndexedFields:   &[]string{},
		NestedFields:             false,
		DynaprepaidActionProfile: []string{},
		Opts: &ActionsOpts{
			ProfileIDs:           []*DynamicStringSliceOpt{},
			ProfileIgnoreFilters: []*DynamicBoolOpt{{}},
			PosterAttempts:       []*DynamicIntOpt{{value: ActionsPosterAttempsDftOpt}},
		},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.ActionSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestV1GetConfigSectionActionSJSON(t *testing.T) {
	var reply map[string]any
	expected := map[string]any{
		ActionSJSON: map[string]any{
			utils.EnabledCfg:                false,
			utils.EEsConnsCfg:               []string{},
			utils.CDRsConnsCfg:              []string{},
			utils.ThresholdSConnsCfg:        []string{},
			utils.StatSConnsCfg:             []string{},
			utils.AccountSConnsCfg:          []string{},
			utils.AdminSConnsCfg:            []string{},
			utils.Tenants:                   []string{},
			utils.IndexedSelectsCfg:         true,
			utils.PrefixIndexedFieldsCfg:    []string{},
			utils.SuffixIndexedFieldsCfg:    []string{},
			utils.ExistsIndexedFieldsCfg:    []string{},
			utils.NotExistsIndexedFieldsCfg: []string{},
			utils.NestedFieldsCfg:           false,
			utils.DynaprepaidActionplansCfg: []string{},
			utils.OptsCfg: map[string]any{
				utils.MetaProfileIDs:           []*DynamicStringSliceOpt{},
				utils.MetaProfileIgnoreFilters: []*DynamicBoolOpt{{}},
				utils.MetaPosterAttempts:       []*DynamicIntOpt{{value: ActionsPosterAttempsDftOpt}},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{ActionSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestLoadActionSCfgError(t *testing.T) {
	cfgJSONStr := `{
      "actions": { 
            "string_indexed_fields": "*req.index",
	  }
    }`
	expected := "json: cannot unmarshal string into Go struct field ActionSJsonCfg.String_indexed_fields of type []string"
	cgrConfig := NewDefaultCGRConfig()

	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.actionSCfg.Load(context.Background(), cgrCfgJSON, nil); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadAccountSCfgError(t *testing.T) {
	cfgJSONStr := `{
"accounts": {								
	"enabled": "not_bool",
   }
}`
	expected := "json: cannot unmarshal string into Go struct field AccountSJsonCfg.Enabled of type bool"
	cfg := NewDefaultCGRConfig()

	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cfg.accountSCfg.Load(context.Background(), cgrCfgJSON, nil); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestCGRConfigGetDP(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.LockSections(HTTPAgentJSON, LoaderJSON, ChargerSJSON)
	cfg.UnlockSections(HTTPAgentJSON, LoaderJSON, ChargerSJSON)
	exp := utils.MapStorage(cfg.AsMapInterface())
	dp := cfg.GetDataProvider()
	if !reflect.DeepEqual(dp, exp) {
		t.Errorf("Expected %+v, received %+v", exp, dp)
	}
}

func TestStoreCfgInDb(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	db := &CgrJsonCfg{}
	cfg.db = db
	cfg.attributeSCfg.Enabled = true
	cfg.attributeSCfg.AccountSConns = []string{"*internal"}
	args := &SectionWithAPIOpts{
		Sections: []string{AttributeSJSON},
	}
	var reply string
	if err := cfg.V1StoreCfgInDB(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	rcv := new(AttributeSJsonCfg)
	if err := db.GetSection(context.Background(), AttributeSJSON, rcv); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv.Enabled, utils.BoolPointer(true)) {
		t.Errorf("Expected %v \n but received \n %v", utils.BoolPointer(true), rcv.Enabled)
	} else if !reflect.DeepEqual(rcv.Accounts_conns, &[]string{"*internal"}) {
		t.Errorf("Expected %v \n but received \n %v", &[]string{"*internal"}, rcv.Accounts_conns)
	}
}

func TestSetCfgInDb(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for key := range utils.StatelessDataDBPartitions {
		cfg.cacheCfg.Partitions[key].Limit = 0
	}
	db := &CgrJsonCfg{}
	cfg.db = db
	cfg.attributeSCfg = &AttributeSCfg{
		Enabled:                true,
		ResourceSConns:         []string{"*internal"},
		StatSConns:             []string{"*internal"},
		AccountSConns:          []string{"*internal"},
		IndexedSelects:         true,
		StringIndexedFields:    &[]string{"field1"},
		SuffixIndexedFields:    &[]string{"field1"},
		PrefixIndexedFields:    &[]string{"field1"},
		ExistsIndexedFields:    &[]string{"field1"},
		NotExistsIndexedFields: &[]string{"field1"},
		Opts: &AttributesOpts{
			ProcessRuns: []*DynamicIntOpt{
				{
					value: int(2),
				},
			},
		},
		NestedFields: true,
	}
	cfg.sections = newSections(cfg)
	cfg.rldCh = make(chan string, 10)
	args := &SetConfigArgs{
		Config: map[string]any{
			"attributes": &AttributeSJsonCfg{
				Enabled:                  utils.BoolPointer(false),
				Resources_conns:          &[]string{"*localhost"},
				Stats_conns:              &[]string{"*localhost"},
				Accounts_conns:           &[]string{"*localhost"},
				Indexed_selects:          utils.BoolPointer(false),
				String_indexed_fields:    &[]string{"field2"},
				Suffix_indexed_fields:    &[]string{"field2"},
				Prefix_indexed_fields:    &[]string{"field2"},
				Exists_indexed_fields:    &[]string{"field2"},
				Notexists_indexed_fields: &[]string{"field2"},
				Opts: &AttributesOptsJson{
					ProcessRuns: []*DynamicInterfaceOpt{
						{
							Value: int(3),
						},
					},
				},
				Nested_fields: utils.BoolPointer(false),
			},
		},
	}
	expected := &AttributeSJsonCfg{
		Enabled:                  utils.BoolPointer(false),
		Resources_conns:          &[]string{"*localhost"},
		Stats_conns:              &[]string{"*localhost"},
		Accounts_conns:           &[]string{"*localhost"},
		Indexed_selects:          utils.BoolPointer(false),
		String_indexed_fields:    &[]string{"field2"},
		Suffix_indexed_fields:    &[]string{"field2"},
		Prefix_indexed_fields:    &[]string{"field2"},
		Exists_indexed_fields:    &[]string{"field2"},
		Notexists_indexed_fields: &[]string{"field2"},
		Opts: &AttributesOptsJson{
			ProcessRuns: []*DynamicInterfaceOpt{
				{
					Value: float64(3),
				},
				{
					Value: float64(2),
				},
			},
		},
		Nested_fields: utils.BoolPointer(false),
	}
	var reply string
	if err := cfg.V1SetConfig(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}

	rcv := new(AttributeSJsonCfg)
	if err := db.GetSection(context.Background(), AttributeSJSON, rcv); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestSetNilCfgInDb(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for key := range utils.StatelessDataDBPartitions {
		cfg.cacheCfg.Partitions[key].Limit = 0
	}
	db := &CgrJsonCfg{}
	cfg.db = db
	cfg.attributeSCfg = &AttributeSCfg{
		Enabled:                true,
		ResourceSConns:         []string{"*internal"},
		StatSConns:             []string{"*internal"},
		AccountSConns:          []string{"*internal"},
		IndexedSelects:         true,
		StringIndexedFields:    &[]string{"field1"},
		SuffixIndexedFields:    &[]string{"field1"},
		PrefixIndexedFields:    &[]string{"field1"},
		ExistsIndexedFields:    &[]string{"field1"},
		NotExistsIndexedFields: &[]string{"field1"},
		Opts: &AttributesOpts{
			ProcessRuns: []*DynamicIntOpt{
				{
					FilterIDs: []string{},
					value:     2,
				},
			},
		},
		NestedFields: true,
	}
	cfg.sections = newSections(cfg)
	cfg.rldCh = make(chan string, 100)
	attributes := &AttributeSJsonCfg{}
	args := &SetConfigArgs{
		Config: map[string]any{
			"attributes": attributes,
		},
	}
	expected := &AttributeSJsonCfg{
		Opts: &AttributesOptsJson{},
	}
	var reply string
	if err := cfg.V1SetConfig(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}

	rcv := new(AttributeSJsonCfg)
	if err := db.GetSection(context.Background(), AttributeSJSON, rcv); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestAsteriskAgentLoadFromJSONCfgNil(t *testing.T) {
	acCfg := &AsteriskAgentCfg{}
	if err := acCfg.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	}
}

func TestAsteriskAgentCloneSection(t *testing.T) {
	astCfg := AsteriskAgentCfg{
		Enabled:       true,
		SessionSConns: []string{"*internal"},
		CreateCDR:     false,
		AsteriskConns: []*AsteriskConnCfg{
			{
				Alias:           "asterisk",
				Address:         ":8080",
				User:            "ast_user",
				Password:        "ast_pass",
				ConnectAttempts: 2,
				Reconnects:      3,
			},
		},
	}

	exp := &AsteriskAgentCfg{
		Enabled:       true,
		SessionSConns: []string{"*internal"},
		CreateCDR:     false,
		AsteriskConns: []*AsteriskConnCfg{
			{
				Alias:           "asterisk",
				Address:         ":8080",
				User:            "ast_user",
				Password:        "ast_pass",
				ConnectAttempts: 2,
				Reconnects:      3,
			},
		},
	}

	rcv := astCfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}
}

func TestConfigAddSection(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	coreSCfg := &CoreSCfg{
		Caps:         2,
		CapsStrategy: utils.MetaReload,
	}
	cfg.AddSection(coreSCfg)

	_, has := cfg.sections.Get("cores")
	if !has {
		t.Error("could not retrieve 'cores' section")
	}
}

func TestConfigDBCfg(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	rcv := cfg.ConfigDBCfg()
	exp := &ConfigDBCfg{
		Type: "*internal",
		Port: "0",
		Opts: &DBOpts{
			InternalDBDumpPath:        "/var/lib/cgrates/internal_db/configdb",
			InternalDBBackupPath:      "/var/lib/cgrates/internal_db/backup/configdb",
			InternalDBStartTimeout:    5 * time.Minute,
			InternalDBDumpInterval:    0,
			InternalDBRewriteInterval: 0,
			InternalDBFileSizeLimit:   1073741824,
			RedisMaxConns:             10,
			RedisConnectAttempts:      20,
			RedisCluster:              false,
			RedisClusterSync:          5 * time.Second,
			RedisClusterOndownDelay:   0,
			RedisPoolPipelineWindow:   150 * time.Microsecond,
			RedisTLS:                  false,
			MongoQueryTimeout:         10 * time.Second,
			MongoConnScheme:           "mongodb",
		},
	}
	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}
}

func TestConfigDBLoadFromJson(t *testing.T) {
	dbCfg := &ConfigDBCfg{}
	jsonCfg := &ConfigDbJsonCfg{
		Db_port: utils.IntPointer(-1),
	}
	if err := dbCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	}
	if err := dbCfg.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	}
}

func TestConfigDBCloneSection(t *testing.T) {
	dbCfg := ConfigDBCfg{
		Type:     "*internal",
		Host:     "localhost",
		Port:     "2013",
		Name:     "dbname",
		User:     "cgrates",
		Password: "superSecretPassword",
		Opts:     &DBOpts{},
	}

	exp := &ConfigDBCfg{
		Type:     "*internal",
		Host:     "localhost",
		Port:     "2013",
		Name:     "dbname",
		User:     "cgrates",
		Password: "superSecretPassword",
		Opts:     &DBOpts{},
	}

	rcv := dbCfg.CloneSection()
	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected %v \n but recevived \n %v", exp, rcv)
	}
}
func TestFreeSwitchAgentCloneSection(t *testing.T) {
	fsAgCfg := FsAgentCfg{
		Enabled:             true,
		SessionSConns:       []string{"*json"},
		SubscribePark:       true,
		CreateCDR:           false,
		ExtraFields:         nil,
		LowBalanceAnnFile:   "lwb_file",
		EmptyBalanceContext: "eb_ctx",
		MaxWaitConnection:   1 * time.Second,
	}

	exp := &FsAgentCfg{
		Enabled:             true,
		SessionSConns:       []string{"*json"},
		SubscribePark:       true,
		CreateCDR:           false,
		ExtraFields:         nil,
		LowBalanceAnnFile:   "lwb_file",
		EmptyBalanceContext: "eb_ctx",
		MaxWaitConnection:   1 * time.Second,
	}

	rcv := fsAgCfg.CloneSection()
	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}
}

func TestFreewitchLoadFromJsonNil(t *testing.T) {
	fsAgCfg := FsAgentCfg{}
	if err := fsAgCfg.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	}
}

func TestLoadConfigFromHTTPErrorNewReqCtx(t *testing.T) {
	cfgCgr := NewDefaultCGRConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Nanosecond)
	cancel()
	url := "inexistentURL"
	expected := `path:"inexistentURL" is not reachable`
	if err := loadConfigFromHTTP(ctx, url, cfgCgr.sections, nil); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, \nreceived %+v", expected, err)
	}

}

func TestGetDftRemoteHostCfg(t *testing.T) {
	cgrCfg := NewDefaultCGRConfig()

	exp := cgrCfg.rpcConns[utils.MetaLocalHost].Conns[0]

	if rcv := getDftRemHstCfg(); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected <%+v>,\nReceived <%+v>", exp, rcv)
	}

}

func TestConfignewCGRConfigFromPathWithoutEnv(t *testing.T) {
	rcv, err := newCGRConfigFromPathWithoutEnv(nil, "")
	if err != nil {
		if err.Error() != `path:"" is not reachable` {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv == nil {
		t.Error(rcv)
	}
}

func TestCfgloadCfgWithLocks(t *testing.T) {
	cfg := &CGRConfig{
		sections: Sections{
			&CoreSCfg{
				Caps:              2,
				CapsStrategy:      "*busy",
				CapsStatsInterval: 2 * time.Second,
				ShutdownTimeout:   3 * time.Second,
			},
		},
	}

	err := cfg.loadCfgWithLocks(nil, "test", "")

	if err != nil {
		if err.Error() != `path:"test" is not reachable` {
			t.Error(err)
		}
	} else {
		t.Error(err)
	}
}
