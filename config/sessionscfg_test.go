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
package config

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestFsAgentCfgloadFromJsonCfg1(t *testing.T) {
	fsAgentJsnCfg := &FreeswitchAgentJsonCfg{
		Enabled:       utils.BoolPointer(true),
		CreateCDR:     utils.BoolPointer(true),
		SubscribePark: utils.BoolPointer(true),
		EventSocketConns: &[]*FsConnJsonCfg{
			{
				Address:    utils.StringPointer("1.2.3.4:8021"),
				Password:   utils.StringPointer("ClueCon"),
				Reconnects: utils.IntPointer(5),
			},
			{
				Address:    utils.StringPointer("2.3.4.5:8021"),
				Password:   utils.StringPointer("ClueCon"),
				Reconnects: utils.IntPointer(5),
			},
		},
	}
	eFsAgentConfig := &FsAgentCfg{
		Enabled:       true,
		CreateCDR:     true,
		SubscribePark: true,
		EventSocketConns: []*FsConnCfg{
			{Address: "1.2.3.4:8021", Password: "ClueCon", Reconnects: 5, ReplyTimeout: time.Minute, Alias: "1.2.3.4:8021"},
			{Address: "2.3.4.5:8021", Password: "ClueCon", Reconnects: 5, ReplyTimeout: time.Minute, Alias: "2.3.4.5:8021"},
		},
	}
	fsAgentCfg := new(FsAgentCfg)
	if err := fsAgentCfg.loadFromJSONCfg(fsAgentJsnCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFsAgentConfig, fsAgentCfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(eFsAgentConfig), utils.ToJSON(fsAgentCfg))
	}
}

func TestSessionSCfgloadFromJsonCfgCase1(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Listen_bijson:         utils.StringPointer("127.0.0.1:2018"),
		Chargers_conns:        &[]string{utils.MetaInternal, "*conn1"},
		Resources_conns:       &[]string{utils.MetaInternal, "*conn1"},
		Thresholds_conns:      &[]string{utils.MetaInternal, "*conn1"},
		Stats_conns:           &[]string{utils.MetaInternal, "*conn1"},
		Routes_conns:          &[]string{utils.MetaInternal, "*conn1"},
		Attributes_conns:      &[]string{utils.MetaInternal, "*conn1"},
		Cdrs_conns:            &[]string{utils.MetaInternal, "*conn1"},
		Actions_conns:         &[]string{utils.MetaInternal, "*conn1"},
		Rates_conns:           &[]string{utils.MetaInternal, "*conn1"},
		Accounts_conns:        &[]string{utils.MetaInternal, "*conn1"},
		Replication_conns:     &[]string{"*conn1"},
		Store_session_costs:   utils.BoolPointer(true),
		Session_indexes:       &[]string{},
		Client_protocol:       utils.Float64Pointer(2.5),
		Channel_sync_interval: utils.StringPointer("10"),
		Terminate_attempts:    utils.IntPointer(6),
		Alterable_fields:      &[]string{},
		Min_dur_low_balance:   utils.StringPointer("1"),
		Stir: &STIRJsonCfg{
			Allowed_attest:      &[]string{utils.MetaAny},
			Payload_maxduration: utils.StringPointer("-1"),
			Default_attest:      utils.StringPointer("A"),
			Publickey_path:      utils.StringPointer("randomPath"),
			Privatekey_path:     utils.StringPointer("randomPath"),
		},
		Opts: &SessionsOptsJson{
			DebitInterval: []*DynamicInterfaceOpt{
				{
					Value: 2 * time.Second,
				},
			},
		},
	}
	expected := &SessionSCfg{
		Enabled:             true,
		ListenBijson:        "127.0.0.1:2018",
		ChargerSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers), "*conn1"},
		ResourceSConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), "*conn1"},
		ThresholdSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
		StatSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"},
		RouteSConns:         []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes), "*conn1"},
		AttributeSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"},
		CDRsConns:           []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs), "*conn1"},
		ActionSConns:        []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions), "*conn1"},
		RateSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates), "*conn1"},
		AccountSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), "*conn1"},
		ReplicationConns:    []string{"*conn1"},
		StoreSCosts:         true,
		SessionIndexes:      utils.StringSet{},
		ClientProtocol:      2.5,
		ChannelSyncInterval: 10,
		TerminateAttempts:   6,
		AlterableFields:     utils.StringSet{},
		MinDurLowBalance:    1,
		STIRCfg: &STIRcfg{
			AllowedAttest:      utils.StringSet{utils.MetaAny: {}},
			PayloadMaxduration: -1,
			DefaultAttest:      "A",
			PrivateKeyPath:     "randomPath",
			PublicKeyPath:      "randomPath",
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
			RoutesDerivedReply:     []*DynamicBoolOpt{{}},
			StatsDerivedReply:      []*DynamicBoolOpt{{}},
			ThresholdsDerivedReply: []*DynamicBoolOpt{{}},
			MaxUsage:               []*DynamicBoolOpt{{}},
			TTL:                    []*DynamicDurationOpt{{value: SessionsTTLDftOpt}},
			Chargeable:             []*DynamicBoolOpt{{value: SessionsChargeableDftOpt}},
			ForceUsage:             []*DynamicBoolOpt{},
			DebitInterval: []*DynamicDurationOpt{
				{
					value: 2 * time.Second,
				},
				{
					value: SessionsDebitIntervalDftOpt,
				},
			},
			TTLLastUsage:       []*DynamicDurationPointerOpt{},
			TTLLastUsed:        []*DynamicDurationPointerOpt{},
			TTLMaxDelay:        []*DynamicDurationOpt{{value: SessionsTTLMaxDelayDftOpt}},
			TTLUsage:           []*DynamicDurationPointerOpt{},
			SessionsOriginID:   []*DynamicStringOpt{},
			AccountsForceUsage: []*DynamicBoolOpt{},
		},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.sessionSCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.sessionSCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.sessionSCfg))
	}
	cfgJSON = nil
	if err := jsonCfg.sessionSCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	}

	if err := expected.Opts.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	}

	if err := expected.STIRCfg.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	}
}

func TestSessionSCfgloadFromJsonCfgCase13(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Opts: &SessionsOptsJson{
			TTL: []*DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "1c",
				},
			},
		},
	}
	errExpect := `time: unknown unit "c" in duration "1c"`
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.sessionSCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v \n but received \n %v", errExpect, err.Error())
	}
	cfgJSON.Opts.TTL = nil

	/////
	cfgJSON.Opts.DebitInterval = []*DynamicInterfaceOpt{
		{
			Tenant: "cgrates.org",
			Value:  "1c",
		},
	}
	if err := jsonCfg.sessionSCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v \n but received \n %v", errExpect, err.Error())
	}
	cfgJSON.Opts.DebitInterval = nil

	/////
	cfgJSON.Opts.TTLLastUsage = []*DynamicInterfaceOpt{
		{
			Tenant: "cgrates.org",
			Value:  "1c",
		},
	}
	if err := jsonCfg.sessionSCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v \n but received \n %v", errExpect, err.Error())
	}
	cfgJSON.Opts.TTLLastUsage = nil

	/////
	cfgJSON.Opts.TTLLastUsed = []*DynamicInterfaceOpt{
		{
			Tenant: "cgrates.org",
			Value:  "1c",
		},
	}
	if err := jsonCfg.sessionSCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v \n but received \n %v", errExpect, err.Error())
	}
	cfgJSON.Opts.TTLLastUsed = nil

	/////
	cfgJSON.Opts.TTLUsage = []*DynamicInterfaceOpt{
		{
			Tenant: "cgrates.org",
			Value:  "1c",
		},
	}
	if err := jsonCfg.sessionSCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v \n but received \n %v", errExpect, err.Error())
	}
	cfgJSON.Opts.TTLUsage = nil

	/////
	cfgJSON.Opts.TTLMaxDelay = []*DynamicInterfaceOpt{
		{
			Tenant: "cgrates.org",
			Value:  "1c",
		},
	}
	if err := jsonCfg.sessionSCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v \n but received \n %v", errExpect, err.Error())
	}
	cfgJSON.Opts.TTLMaxDelay = nil
}

func TestSessionSCfgloadFromJsonCfgCase2(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Replication_conns: &[]string{utils.MetaInternal},
	}
	expected := "Replication connection ID needs to be different than *internal "
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.sessionSCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSessionSCfgloadFromJsonCfgCase7(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Channel_sync_interval: utils.StringPointer("1ss"),
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.sessionSCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSessionSCfgloadFromJsonCfgCase8(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Min_dur_low_balance: utils.StringPointer("1ss"),
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.sessionSCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSessionSCfgloadFromJsonCfgCase10(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Opts: &SessionsOptsJson{
			TTLLastUsage: []*DynamicInterfaceOpt{
				{
					Value: "1",
				},
			},
			TTLLastUsed: []*DynamicInterfaceOpt{
				{
					Value: "10",
				},
			},
			TTLMaxDelay: []*DynamicInterfaceOpt{
				{
					Value: "100",
				},
			},
			TTLUsage: []*DynamicInterfaceOpt{
				{
					Value: "1",
				},
			},
		},
	}
	expected := &SessionSCfg{
		Enabled:             false,
		ListenBijson:        "127.0.0.1:2014",
		ChargerSConns:       []string{},
		ResourceSConns:      []string{},
		ThresholdSConns:     []string{},
		StatSConns:          []string{},
		RouteSConns:         []string{},
		AttributeSConns:     []string{},
		CDRsConns:           []string{},
		ReplicationConns:    []string{},
		ActionSConns:        []string{},
		RateSConns:          []string{},
		AccountSConns:       []string{},
		StoreSCosts:         false,
		SessionIndexes:      utils.StringSet{},
		ClientProtocol:      1.0,
		ChannelSyncInterval: 0,
		TerminateAttempts:   5,
		AlterableFields:     utils.StringSet{},
		MinDurLowBalance:    0,
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
			RoutesDerivedReply:     []*DynamicBoolOpt{{}},
			StatsDerivedReply:      []*DynamicBoolOpt{{}},
			ThresholdsDerivedReply: []*DynamicBoolOpt{{}},
			MaxUsage:               []*DynamicBoolOpt{{}},
			TTL:                    []*DynamicDurationOpt{{value: SessionsTTLDftOpt}},
			Chargeable:             []*DynamicBoolOpt{{value: SessionsChargeableDftOpt}},
			DebitInterval:          []*DynamicDurationOpt{{value: SessionsDebitIntervalDftOpt}},
			ForceUsage:             []*DynamicBoolOpt{},
			SessionsOriginID:       []*DynamicStringOpt{},
			AccountsForceUsage:     []*DynamicBoolOpt{},
			TTLLastUsage: []*DynamicDurationPointerOpt{
				{
					value: utils.DurationPointer(1),
				},
			},
			TTLLastUsed: []*DynamicDurationPointerOpt{
				{
					value: utils.DurationPointer(10),
				},
			},
			TTLMaxDelay: []*DynamicDurationOpt{
				{
					value: 100,
				},
				{
					value: SessionsTTLMaxDelayDftOpt,
				},
			},
			TTLUsage: []*DynamicDurationPointerOpt{
				{
					value: utils.DurationPointer(1),
				},
			},
		},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.sessionSCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(jsonCfg.sessionSCfg, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.sessionSCfg))
	}
}

func TestSessionSCfgloadFromJsonCfgCase11(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Stir: &STIRJsonCfg{
			Payload_maxduration: utils.StringPointer("1ss"),
		},
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.sessionSCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSessionSCfgloadFromJsonCfgCase12(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Default_usage: map[string]string{
			utils.MetaAny:   "1ss",
			utils.MetaVoice: "1ss",
			utils.MetaData:  "1ss",
			utils.MetaSMS:   "1ss",
		},
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.sessionSCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestGetDefaultUsage(t *testing.T) {
	session := &SessionSCfg{
		DefaultUsage: map[string]time.Duration{
			"test":        time.Hour,
			utils.MetaAny: time.Second,
		},
	}
	expected := time.Hour
	if rcv := session.GetDefaultUsage("test"); !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
	expected = time.Second
	if rcv := session.GetDefaultUsage(utils.EmptyString); !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
}

func TestSessionSCfgAsMapInterfaceCase1(t *testing.T) {
	cfgJSONStr := `{
	"sessions": {
          "channel_sync_interval": "1s",
          "session_ttl_max_delay": "3h0m0s",
          "session_ttl_last_used": "0s",
          "session_ttl_usage": "1s",
          "session_ttl_last_usage": "10s",
           "sessions": {
			"stir": {
				"payload_maxduration": "-1",
			},
		},
    },
}`
	eMap := map[string]any{
		utils.EnabledCfg:             false,
		utils.ListenBijsonCfg:        "127.0.0.1:2014",
		utils.ListenBigobCfg:         "",
		utils.ChargerSConnsCfg:       []string{},
		utils.CDRsConnsCfg:           []string{},
		utils.ResourceSConnsCfg:      []string{},
		utils.ThresholdSConnsCfg:     []string{},
		utils.StatSConnsCfg:          []string{},
		utils.RouteSConnsCfg:         []string{},
		utils.AttributeSConnsCfg:     []string{},
		utils.ReplicationConnsCfg:    []string{},
		utils.ActionSConnsCfg:        []string{},
		utils.RateSConnsCfg:          []string{},
		utils.AccountSConnsCfg:       []string{},
		utils.StoreSCostsCfg:         false,
		utils.SessionIndexesCfg:      []string{},
		utils.ClientProtocolCfg:      1.0,
		utils.ChannelSyncIntervalCfg: "1s",
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
			utils.MetaSessionsOriginID:          []*DynamicStringOpt{},
			utils.MetaAccountsForceUsage:        []*DynamicBoolOpt{},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.sessionSCfg.AsMapInterface(""); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestSessionSCfgAsMapInterfaceCase2(t *testing.T) {
	cfgJSONStr := `{
		"sessions": {
			"enabled": true,
			"listen_bijson": "127.0.0.1:2018",
			"chargers_conns": ["*internal:*chargers", "*conn1"],
			"cdrs_conns": ["*internal:*cdrs", "*conn1"],
			"resources_conns": ["*internal:*resources", "*conn1"],
			"thresholds_conns": ["*internal:*thresholds", "*conn1"],
			"stats_conns": ["*internal:*stats", "*conn1"],
			"routes_conns": ["*internal:*routes", "*conn1"],
			"attributes_conns": ["*internal:*attributes", "*conn1"],
			"actions_conns": ["*internal:*actions", "*conn1"],
			"rates_conns": ["*internal:*rates", "*conn1"],
			"accounts_conns": ["*internal:*accounts", "*conn1"],
			"replication_conns": ["*localhost"],
			"store_session_costs": true,
            "min_dur_low_balance": "1s",
			"client_protocol": 2.0,
			"terminate_attempts": 10,
			"stir": {
				"allowed_attest": ["any1","any2"],
				"payload_maxduration": "1s",
				"default_attest": "B",
				"publickey_path": "",
				"privatekey_path": "",
			},
			"opts": {
				"*ttl": [
					{
						"Value": "1s",
					},
				],
				"*debitInterval": [
					{
						"Value": "8s",
					},
				],
			},
		},
	}`
	eMap := map[string]any{
		utils.EnabledCfg:             true,
		utils.ListenBijsonCfg:        "127.0.0.1:2018",
		utils.ListenBigobCfg:         "",
		utils.ChargerSConnsCfg:       []string{utils.MetaInternal, "*conn1"},
		utils.CDRsConnsCfg:           []string{utils.MetaInternal, "*conn1"},
		utils.ResourceSConnsCfg:      []string{utils.MetaInternal, "*conn1"},
		utils.ThresholdSConnsCfg:     []string{utils.MetaInternal, "*conn1"},
		utils.StatSConnsCfg:          []string{utils.MetaInternal, "*conn1"},
		utils.RouteSConnsCfg:         []string{utils.MetaInternal, "*conn1"},
		utils.AttributeSConnsCfg:     []string{utils.MetaInternal, "*conn1"},
		utils.ActionSConnsCfg:        []string{utils.MetaInternal, "*conn1"},
		utils.RateSConnsCfg:          []string{utils.MetaInternal, "*conn1"},
		utils.AccountSConnsCfg:       []string{utils.MetaInternal, "*conn1"},
		utils.ReplicationConnsCfg:    []string{utils.MetaLocalHost},
		utils.StoreSCostsCfg:         true,
		utils.MinDurLowBalanceCfg:    "1s",
		utils.SessionIndexesCfg:      []string{},
		utils.ClientProtocolCfg:      2.0,
		utils.ChannelSyncIntervalCfg: "0",
		utils.TerminateAttemptsCfg:   10,
		utils.AlterableFieldsCfg:     []string{},
		utils.STIRCfg: map[string]any{
			utils.AllowedAtestCfg:       []string{"any1", "any2"},
			utils.PayloadMaxdurationCfg: "1s",
			utils.DefaultAttestCfg:      "B",
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
			utils.MetaRoutesDerivedReplyCfg:     []*DynamicBoolOpt{{}},
			utils.MetaStatsDerivedReplyCfg:      []*DynamicBoolOpt{{}},
			utils.MetaThresholdsDerivedReplyCfg: []*DynamicBoolOpt{{}},
			utils.MetaMaxUsageCfg:               []*DynamicBoolOpt{{}},
			utils.MetaTTLCfg: []*DynamicDurationOpt{
				{
					value: time.Second,
				},
				{
					value: SessionsTTLDftOpt,
				},
			},
			utils.MetaChargeableCfg: []*DynamicBoolOpt{{value: SessionsChargeableDftOpt}},
			utils.MetaDebitIntervalCfg: []*DynamicDurationOpt{
				{
					value: 8 * time.Second,
				},
				{
					value: SessionsDebitIntervalDftOpt,
				},
			},
			utils.MetaTTLLastUsageCfg:    []*DynamicDurationPointerOpt{},
			utils.MetaTTLLastUsedCfg:     []*DynamicDurationPointerOpt{},
			utils.MetaTTLMaxDelayCfg:     []*DynamicDurationOpt{{value: SessionsTTLMaxDelayDftOpt}},
			utils.MetaTTLUsageCfg:        []*DynamicDurationPointerOpt{},
			utils.MetaForceUsageCfg:      []*DynamicBoolOpt{},
			utils.MetaSessionsOriginID:   []*DynamicStringOpt{},
			utils.MetaAccountsForceUsage: []*DynamicBoolOpt{},
		},
	}
	cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr)
	if err != nil {
		t.Error(err)
	}
	rcv := cgrCfg.sessionSCfg.AsMapInterface("").(map[string]any)
	sort.Strings(rcv[utils.STIRCfg].(map[string]any)[utils.AllowedAtestCfg].([]string))
	if !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestSessionSCfgAsMapInterfaceCase3(t *testing.T) {
	cfgJSONStr := `{
	"sessions": {
			"stir": {
				"payload_maxduration": "0",
			},
		},
    },
}`
	eMap := map[string]any{
		utils.STIRCfg: map[string]any{
			utils.AllowedAtestCfg:       []string{"*any"},
			utils.PayloadMaxdurationCfg: "0",
			utils.DefaultAttestCfg:      "A",
			utils.PublicKeyPathCfg:      "",
			utils.PrivateKeyPathCfg:     "",
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.sessionSCfg.AsMapInterface("").(map[string]any); !reflect.DeepEqual(eMap[utils.STIRCfg], rcv[utils.STIRCfg]) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap[utils.STIRCfg]), utils.ToJSON(rcv[utils.STIRCfg]))
	}
}

func TestFsAgentCfgloadFromJsonCfgCase1(t *testing.T) {
	fsAgentJsnCfg := &FreeswitchAgentJsonCfg{
		Enabled:                utils.BoolPointer(true),
		SessionSConns:          &[]string{utils.MetaInternal},
		CreateCDR:              utils.BoolPointer(true),
		SubscribePark:          utils.BoolPointer(true),
		LowBalanceAnnFile:      utils.StringPointer("randomFile"),
		EmptyBalanceAnnFile:    utils.StringPointer("randomEmptyFile"),
		EmptyBalanceContext:    utils.StringPointer("randomEmptyContext"),
		ActiveSessionDelimiter: utils.StringPointer("/"),
		MaxWaitConnection:      utils.StringPointer("2"),
		ExtraFields:            &[]string{},
		EventSocketConns: &[]*FsConnJsonCfg{
			{
				Address:      utils.StringPointer("1.2.3.4:8021"),
				Password:     utils.StringPointer("ClueCon"),
				Reconnects:   utils.IntPointer(5),
				Alias:        utils.StringPointer("127.0.0.1:8021"),
				ReplyTimeout: utils.StringPointer("2m"),
			},
		},
	}
	expected := &FsAgentCfg{
		Enabled:                true,
		SessionSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		SubscribePark:          true,
		CreateCDR:              true,
		LowBalanceAnnFile:      "randomFile",
		EmptyBalanceAnnFile:    "randomEmptyFile",
		EmptyBalanceContext:    "randomEmptyContext",
		ActiveSessionDelimiter: "/",
		MaxWaitConnection:      2,
		ExtraFields:            RSRParsers{},
		EventSocketConns: []*FsConnCfg{
			{
				Address:      "1.2.3.4:8021",
				Password:     "ClueCon",
				Reconnects:   5,
				Alias:        "127.0.0.1:8021",
				ReplyTimeout: 2 * time.Minute,
			},
		},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.fsAgentCfg.loadFromJSONCfg(fsAgentJsnCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.fsAgentCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.fsAgentCfg))
	}
}

func TestFsAgentCfgloadFromJsonCfgCase2(t *testing.T) {
	fsAgentJsnCfg := &FreeswitchAgentJsonCfg{
		MaxWaitConnection: utils.StringPointer("1ss"),
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.fsAgentCfg.loadFromJSONCfg(fsAgentJsnCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFsAgentCfgloadFromJsonCfgCase3(t *testing.T) {
	fsAgentJsnCfg := &FreeswitchAgentJsonCfg{
		ExtraFields: &[]string{"a{*"},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.fsAgentCfg.loadFromJSONCfg(fsAgentJsnCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFsAgentCfgAsMapInterfaceCase1(t *testing.T) {
	cfgJSONStr := `{
	"freeswitch_agent": {},
}`
	eMap := map[string]any{
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
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.fsAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, recevied %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestFsAgentCfgAsMapInterfaceCase2(t *testing.T) {
	cfgJSONStr := `{
"freeswitch_agent": {
	"enabled": true,
	"sessions_conns": ["*birpc_internal", "*conn1","*conn2"],
	"subscribe_park": false,
	"create_cdr": true,
	"max_wait_connection": "7s",
	"active_session_delimiter": "//",
	"event_socket_conns": [
	{"address": "127.0.0.1:8000", "password": "ClueCon123", "reconnects": 8, "max_reconnect_interval": "5m", "reply_timeout": "2m", "alias": "127.0.0.1:8000"}
],}
}`
	eMap := map[string]any{
		utils.EnabledCfg:                true,
		utils.SessionSConnsCfg:          []string{rpcclient.BiRPCInternal, "*conn1", "*conn2"},
		utils.SubscribeParkCfg:          false,
		utils.CreateCdrCfg:              true,
		utils.ExtraFieldsCfg:            []string{},
		utils.LowBalanceAnnFileCfg:      "",
		utils.EmptyBalanceContextCfg:    "",
		utils.EmptyBalanceAnnFileCfg:    "",
		utils.MaxWaitConnectionCfg:      "7s",
		utils.ActiveSessionDelimiterCfg: "//",
		utils.EventSocketConnsCfg: []map[string]any{
			{
				utils.AddressCfg:              "127.0.0.1:8000",
				utils.Password:                "ClueCon123",
				utils.ReconnectsCfg:           8,
				utils.MaxReconnectIntervalCfg: "5m0s",
				utils.ReplyTimeoutCfg:         "2m0s",
				utils.AliasCfg:                "127.0.0.1:8000",
			},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.fsAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, recevied %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestFsAgentCfgAsMapInterfaceCase3(t *testing.T) {
	cfgJSONStr := `{
	"freeswitch_agent": {
          "extra_fields": ["randomFields"],		
          "max_wait_connection": "0",
		  "sessions_conns": ["*internal"]
    }
}`
	eMap := map[string]any{
		utils.EnabledCfg:                false,
		utils.SessionSConnsCfg:          []string{utils.MetaInternal},
		utils.SubscribeParkCfg:          true,
		utils.CreateCdrCfg:              false,
		utils.ExtraFieldsCfg:            []string{"randomFields"},
		utils.LowBalanceAnnFileCfg:      "",
		utils.EmptyBalanceContextCfg:    "",
		utils.EmptyBalanceAnnFileCfg:    "",
		utils.MaxWaitConnectionCfg:      "",
		utils.ActiveSessionDelimiterCfg: ",",
		utils.EventSocketConnsCfg: []map[string]any{
			{
				utils.AddressCfg:              "127.0.0.1:8021",
				utils.Password:                "ClueCon",
				utils.ReconnectsCfg:           5,
				utils.MaxReconnectIntervalCfg: "0s",
				utils.AliasCfg:                "127.0.0.1:8021",
				utils.ReplyTimeoutCfg:         "1m0s",
			},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.fsAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, recevied %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

}

func TestFsConnCfgloadFromJsonCfg(t *testing.T) {
	var fscocfg, expected FsConnCfg
	if err := fscocfg.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fscocfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, fscocfg)
	}
	if err := fscocfg.loadFromJSONCfg(new(FsConnJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fscocfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, fscocfg)
	}
	json := &FsConnJsonCfg{
		Address:    utils.StringPointer("127.0.0.1:8448"),
		Password:   utils.StringPointer("pass123"),
		Reconnects: utils.IntPointer(5),
		Alias:      utils.StringPointer("127.0.0.1:8448"),
	}
	expected = FsConnCfg{
		Address:    "127.0.0.1:8448",
		Password:   "pass123",
		Reconnects: 5,
		Alias:      "127.0.0.1:8448",
	}
	if err := fscocfg.loadFromJSONCfg(json); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, fscocfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(fscocfg))
	}
}

func TestRemoteHostloadFromJsonCfg(t *testing.T) {
	var hpoolcfg, expected RemoteHost
	hpoolcfg.loadFromJSONCfg(nil)
	if !reflect.DeepEqual(hpoolcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, hpoolcfg)
	}
	hpoolcfg.loadFromJSONCfg(new(RemoteHostJson))
	if !reflect.DeepEqual(hpoolcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, hpoolcfg)
	}
	json := &RemoteHostJson{
		Address: utils.StringPointer("127.0.0.1:8448"),
	}
	expected = RemoteHost{
		Address: "127.0.0.1:8448",
	}
	hpoolcfg.loadFromJSONCfg(json)
	if !reflect.DeepEqual(expected, hpoolcfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(hpoolcfg))
	}
}

func TestAsteriskAgentCfgloadFromJsonCfg(t *testing.T) {
	cfgJSON := &AsteriskAgentJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{utils.MetaInternal},
		Create_cdr:     utils.BoolPointer(true),
		Asterisk_conns: &[]*AstConnJsonCfg{
			{
				Alias:                  utils.StringPointer("127.0.0.1:8448"),
				Address:                utils.StringPointer("127.0.0.1:8088"),
				User:                   utils.StringPointer(utils.CGRateSLwr),
				Password:               utils.StringPointer("CGRateS.org"),
				Max_reconnect_interval: utils.StringPointer("5m"),
				Connect_attempts:       utils.IntPointer(3),
				Reconnects:             utils.IntPointer(5),
			},
		},
	}
	expected := &AsteriskAgentCfg{
		Enabled:       true,
		SessionSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		CreateCDR:     true,
		AsteriskConns: []*AsteriskConnCfg{{
			Alias:                "127.0.0.1:8448",
			Address:              "127.0.0.1:8088",
			User:                 "cgrates",
			Password:             "CGRateS.org",
			ConnectAttempts:      3,
			Reconnects:           5,
			MaxReconnectInterval: 5 * time.Minute,
		}},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.asteriskAgentCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.asteriskAgentCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.asteriskAgentCfg))
	}
}

func TestAsteriskAgentCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"asterisk_agent": {
		"sessions_conns": ["*internal"],
	},
}`
	eMap := map[string]any{
		utils.EnabledCfg:       false,
		utils.SessionSConnsCfg: []string{utils.MetaInternal},
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
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.asteriskAgentCfg.AsMapInterface(""); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestAsteriskAgentCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
	"asterisk_agent": {
		"enabled": true,
		"sessions_conns": ["*birpc_internal", "*conn1","*conn2"],
		"create_cdr": true,
		"asterisk_conns":[
			{"address": "127.0.0.1:8089","connect_attempts": 5,"reconnects": 8, "max_reconnect_interval": "5m"}
		],
	},
}`
	eMap := map[string]any{
		utils.EnabledCfg:       true,
		utils.SessionSConnsCfg: []string{rpcclient.BiRPCInternal, "*conn1", "*conn2"},
		utils.CreateCdrCfg:     true,
		utils.AsteriskConnsCfg: []map[string]any{
			{
				utils.AliasCfg:                "",
				utils.AddressCfg:              "127.0.0.1:8089",
				utils.UserCf:                  "cgrates",
				utils.Password:                "CGRateS.org",
				utils.ConnectAttemptsCfg:      5,
				utils.ReconnectsCfg:           8,
				utils.MaxReconnectIntervalCfg: "5m0s",
			},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.asteriskAgentCfg.AsMapInterface(""); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestAsteriskConnCfgloadFromJsonCfg(t *testing.T) {
	var asconcfg, expected AsteriskConnCfg
	if err := asconcfg.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(asconcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, asconcfg)
	}
	if err := asconcfg.loadFromJSONCfg(new(AstConnJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(asconcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, asconcfg)
	}
	json := &AstConnJsonCfg{
		Address:          utils.StringPointer("127.0.0.1:8088"),
		User:             utils.StringPointer("cgrates"),
		Password:         utils.StringPointer("CGRateS.org"),
		Connect_attempts: utils.IntPointer(3),
		Reconnects:       utils.IntPointer(5),
	}
	expected = AsteriskConnCfg{
		Address:         "127.0.0.1:8088",
		User:            "cgrates",
		Password:        "CGRateS.org",
		ConnectAttempts: 3,
		Reconnects:      5,
	}
	if err := asconcfg.loadFromJSONCfg(json); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, asconcfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(asconcfg))
	}
}

func TestAsteriskAgentCfgClone(t *testing.T) {
	ban := &AsteriskAgentCfg{
		Enabled:       true,
		SessionSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS), "*conn1"},
		CreateCDR:     true,
		AsteriskConns: []*AsteriskConnCfg{{
			Alias:           "127.0.0.1:8448",
			Address:         "127.0.0.1:8088",
			User:            "cgrates",
			Password:        "CGRateS.org",
			ConnectAttempts: 3,
			Reconnects:      5,
		}},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.SessionSConns[1] = ""; ban.SessionSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.AsteriskConns[0].User = ""; ban.AsteriskConns[0].User != "cgrates" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffAstConnJsonCfg(t *testing.T) {
	v1 := &AsteriskConnCfg{
		Alias:           "AsteriskAlias",
		Address:         "localhost:8080",
		User:            "cgrates.org",
		Password:        "CGRateS.org",
		ConnectAttempts: 2,
		Reconnects:      2,
	}

	v2 := &AsteriskConnCfg{
		Alias:           "",
		Address:         "localhost:8037",
		User:            "itsyscom.com",
		Password:        "ITSysCOM.com",
		ConnectAttempts: 3,
		Reconnects:      3,
	}

	expected := &AstConnJsonCfg{
		Alias:            utils.StringPointer(""),
		Address:          utils.StringPointer("localhost:8037"),
		User:             utils.StringPointer("itsyscom.com"),
		Password:         utils.StringPointer("ITSysCOM.com"),
		Connect_attempts: utils.IntPointer(3),
		Reconnects:       utils.IntPointer(3),
	}

	rcv := diffAstConnJsonCfg(v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2_2 := v1
	expected2 := &AstConnJsonCfg{
		Alias:            nil,
		Address:          nil,
		User:             nil,
		Password:         nil,
		Connect_attempts: nil,
		Reconnects:       nil,
	}

	rcv = diffAstConnJsonCfg(v1, v2_2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}

func TestEqualsAstConnJsonCfg(t *testing.T) {

	//When not equal
	v1 := []*AsteriskConnCfg{
		{
			Alias:           "AsteriskAlias",
			Address:         "localhost:8080",
			User:            "cgrates.org",
			Password:        "CGRateS.org",
			ConnectAttempts: 2,
			Reconnects:      2,
		},
	}

	v2 := []*AsteriskConnCfg{
		{
			Alias:           "",
			Address:         "localhost:8037",
			User:            "itsyscom.com",
			Password:        "ITSysCOM.com",
			ConnectAttempts: 3,
			Reconnects:      3,
		},
	}

	rcv := equalsAstConnJsonCfg(v1, v2)
	if rcv {
		t.Error("Cfgs should not match")
	}

	//When equal
	v2 = v1
	rcv = equalsAstConnJsonCfg(v1, v2)
	if !rcv {
		t.Error("Cfgs should match")
	}

	v2 = []*AsteriskConnCfg{
		{
			Alias:           "",
			Address:         "localhost:8037",
			User:            "itsyscom.com",
			Password:        "ITSysCOM.com",
			ConnectAttempts: 3,
			Reconnects:      3,
		},
		{
			Alias:           "AsteriskAlias",
			Address:         "localhost:8080",
			User:            "cgrates.org",
			Password:        "CGRateS.org",
			ConnectAttempts: 2,
			Reconnects:      2,
		},
	}

	rcv = equalsAstConnJsonCfg(v1, v2)
	if rcv {
		t.Error("Length of cfgs should not match")
	}
}

func TestDiffAsteriskAgentJsonCfg(t *testing.T) {
	var d *AsteriskAgentJsonCfg

	v1 := &AsteriskAgentCfg{
		Enabled:       false,
		SessionSConns: []string{"*localhost"},
		CreateCDR:     false,
		AsteriskConns: []*AsteriskConnCfg{
			{
				Alias:           "",
				Address:         "localhost:8037",
				User:            "itsyscom.com",
				Password:        "ITSysCOM.com",
				ConnectAttempts: 3,
				Reconnects:      3,
			},
		},
	}

	v2 := &AsteriskAgentCfg{
		Enabled:       true,
		SessionSConns: []string{"*birpc"},
		CreateCDR:     true,
		AsteriskConns: []*AsteriskConnCfg{
			{
				Alias:           "AsteriskAlias",
				Address:         "localhost:8080",
				User:            "cgrates.org",
				Password:        "CGRateS.org1",
				ConnectAttempts: 2,
				Reconnects:      2,
			},
		},
	}

	expected := &AsteriskAgentJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*birpc"},
		Create_cdr:     utils.BoolPointer(true),
		Asterisk_conns: &[]*AstConnJsonCfg{
			{
				Alias:            utils.StringPointer("AsteriskAlias"),
				Address:          utils.StringPointer("localhost:8080"),
				User:             utils.StringPointer("cgrates.org"),
				Password:         utils.StringPointer("CGRateS.org1"),
				Connect_attempts: utils.IntPointer(2),
				Reconnects:       utils.IntPointer(2),
			},
		},
	}

	rcv := diffAsteriskAgentJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestFsAgentCfgClone(t *testing.T) {
	ban := &FsAgentCfg{
		Enabled:             true,
		CreateCDR:           true,
		SubscribePark:       true,
		EmptyBalanceAnnFile: "file",
		EmptyBalanceContext: "context",
		ExtraFields:         NewRSRParsersMustCompile("tenant", utils.InfieldSep),
		LowBalanceAnnFile:   "file2",
		MaxWaitConnection:   time.Second,
		SessionSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS), "*conn1"},
		EventSocketConns: []*FsConnCfg{
			{Address: "1.2.3.4:8021", Password: "ClueCon", Reconnects: 5, Alias: "1.2.3.4:8021"},
			{Address: "2.3.4.5:8021", Password: "ClueCon", Reconnects: 5, Alias: "2.3.4.5:8021"},
		},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.SessionSConns[1] = ""; ban.SessionSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.EventSocketConns[0].Password = ""; ban.EventSocketConns[0].Password != "ClueCon" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffFsConnJsonCfg(t *testing.T) {
	v1 := &FsConnCfg{
		Address:    "localhost:8080",
		Password:   "FsPassword",
		Reconnects: 3,
		Alias:      "FS",
	}

	v2 := &FsConnCfg{
		Address:    "localhost:8037",
		Password:   "AnotherFsPassword",
		Reconnects: 1,
		Alias:      "FS_AGENT",
	}

	expected := &FsConnJsonCfg{
		Address:    utils.StringPointer("localhost:8037"),
		Password:   utils.StringPointer("AnotherFsPassword"),
		Reconnects: utils.IntPointer(1),
		Alias:      utils.StringPointer("FS_AGENT"),
	}

	rcv := diffFsConnJsonCfg(v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &FsConnJsonCfg{}

	rcv = diffFsConnJsonCfg(v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestEqualsFsConnsJsonCfg(t *testing.T) {
	v1 := []*FsConnCfg{
		{
			Address:    "localhost:8080",
			Password:   "FsPassword",
			Reconnects: 3,
			Alias:      "FS",
		},
	}

	v2 := []*FsConnCfg{
		{
			Address:    "localhost:8037",
			Password:   "AnotherFsPassword",
			Reconnects: 1,
			Alias:      "FS_AGENT",
		},
	}

	if equalsFsConnsJsonCfg(v1, v2) {
		t.Error("Conns should not match")
	}

	v2 = []*FsConnCfg{
		{
			Address:    "localhost:8080",
			Password:   "FsPassword",
			Reconnects: 3,
			Alias:      "FS",
		},
	}

	if !equalsFsConnsJsonCfg(v1, v2) {
		t.Error("Conns should match")
	}
}

func TestDiffFreeswitchAgentJsonCfg(t *testing.T) {
	var d *FreeswitchAgentJsonCfg

	v1 := &FsAgentCfg{
		Enabled:       false,
		SessionSConns: []string{},
		SubscribePark: false,
		CreateCDR:     false,
		ExtraFields: RSRParsers{
			{
				Rules: "ExtraField",
			},
		},
		LowBalanceAnnFile:   "LBAF",
		EmptyBalanceContext: "EBC",
		EmptyBalanceAnnFile: "EBAF",
		MaxWaitConnection:   5 * time.Second,
		EventSocketConns:    []*FsConnCfg{},
	}

	v2 := &FsAgentCfg{
		Enabled:       true,
		SessionSConns: []string{"*localhost"},
		SubscribePark: true,
		CreateCDR:     true,
		ExtraFields: RSRParsers{
			{
				Rules: "ExtraField2",
			},
		},
		LowBalanceAnnFile:   "LBAF2",
		EmptyBalanceContext: "EBC2",
		EmptyBalanceAnnFile: "EBAF2",
		MaxWaitConnection:   3 * time.Second,
		EventSocketConns: []*FsConnCfg{
			{
				Address:      "localhost:8080",
				Password:     "FsPassword",
				Reconnects:   3,
				Alias:        "FS",
				ReplyTimeout: 30 * time.Second,
			},
		},
	}

	expected := &FreeswitchAgentJsonCfg{
		Enabled:             utils.BoolPointer(true),
		SessionSConns:       &[]string{"*localhost"},
		SubscribePark:       utils.BoolPointer(true),
		CreateCDR:           utils.BoolPointer(true),
		ExtraFields:         &[]string{"ExtraField2"},
		LowBalanceAnnFile:   utils.StringPointer("LBAF2"),
		EmptyBalanceContext: utils.StringPointer("EBC2"),
		EmptyBalanceAnnFile: utils.StringPointer("EBAF2"),
		MaxWaitConnection:   utils.StringPointer("3s"),
		EventSocketConns: &[]*FsConnJsonCfg{
			{
				Address:      utils.StringPointer("localhost:8080"),
				Password:     utils.StringPointer("FsPassword"),
				Reconnects:   utils.IntPointer(3),
				Alias:        utils.StringPointer("FS"),
				ReplyTimeout: utils.StringPointer("30s"),
			},
		},
	}

	rcv := diffFreeswitchAgentJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &FreeswitchAgentJsonCfg{}

	rcv = diffFreeswitchAgentJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestSessionSCfgClone(t *testing.T) {
	ban := &SessionSCfg{
		Enabled:             true,
		ListenBijson:        "127.0.0.1:2018",
		ChargerSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers), "*conn1"},
		ResourceSConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), "*conn1"},
		ThresholdSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
		StatSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"},
		RouteSConns:         []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes), "*conn1"},
		AttributeSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"},
		CDRsConns:           []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs), "*conn1"},
		ReplicationConns:    []string{"*conn1"},
		StoreSCosts:         true,
		SessionIndexes:      utils.StringSet{},
		ClientProtocol:      2.5,
		ChannelSyncInterval: 10,
		TerminateAttempts:   6,
		AlterableFields:     utils.StringSet{},
		MinDurLowBalance:    1,
		ActionSConns:        []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions), "*conn1"},
		STIRCfg: &STIRcfg{
			AllowedAttest:      utils.StringSet{utils.MetaAny: {}},
			PayloadMaxduration: -1,
			DefaultAttest:      "A",
			PrivateKeyPath:     "randomPath",
			PublicKeyPath:      "randomPath",
		},
		DefaultUsage: map[string]time.Duration{
			utils.MetaAny:   3 * time.Hour,
			utils.MetaVoice: 3 * time.Hour,
			utils.MetaData:  1048576,
			utils.MetaSMS:   1,
		},
		Opts: &SessionsOpts{
			DebitInterval: []*DynamicDurationOpt{
				{
					value: 2,
				},
			},
			TTL: []*DynamicDurationOpt{
				{
					value: 0,
				},
			},
			TTLMaxDelay: []*DynamicDurationOpt{
				{
					value: time.Second,
				},
			},
			TTLLastUsed: []*DynamicDurationPointerOpt{
				{
					value: utils.DurationPointer(time.Second),
				},
			},
			TTLLastUsage: []*DynamicDurationPointerOpt{
				{
					value: utils.DurationPointer(time.Second),
				},
			},
			TTLUsage: []*DynamicDurationPointerOpt{
				{
					value: utils.DurationPointer(time.Second),
				},
			},
		},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.ChargerSConns[1] = ""; ban.ChargerSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}

	if rcv.ResourceSConns[1] = ""; ban.ResourceSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.ThresholdSConns[1] = ""; ban.ThresholdSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.StatSConns[1] = ""; ban.StatSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.RouteSConns[1] = ""; ban.RouteSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.AttributeSConns[1] = ""; ban.AttributeSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.CDRsConns[1] = ""; ban.CDRsConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.ReplicationConns[0] = ""; ban.ReplicationConns[0] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}

	if rcv.STIRCfg.DefaultAttest = ""; ban.STIRCfg.DefaultAttest != "A" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffSTIRJsonCfg(t *testing.T) {
	var d *STIRJsonCfg

	v1 := &STIRcfg{
		AllowedAttest: utils.StringSet{
			"A_TEST1": {},
		},
		PayloadMaxduration: 2 * time.Second,
		DefaultAttest:      "default_attest",
		PublicKeyPath:      "/public/key/path",
		PrivateKeyPath:     "/private/key/path",
	}

	v2 := &STIRcfg{
		AllowedAttest:      nil,
		PayloadMaxduration: 4 * time.Second,
		DefaultAttest:      "default_attest2",
		PublicKeyPath:      "/public/key/path/2",
		PrivateKeyPath:     "/private/key/path/2",
	}

	expected := &STIRJsonCfg{
		Allowed_attest:      nil,
		Payload_maxduration: utils.StringPointer("4s"),
		Default_attest:      utils.StringPointer("default_attest2"),
		Publickey_path:      utils.StringPointer("/public/key/path/2"),
		Privatekey_path:     utils.StringPointer("/private/key/path/2"),
	}

	rcv := diffSTIRJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &STIRJsonCfg{}
	rcv = diffSTIRJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestDiffSessionSJsonCfg(t *testing.T) {
	var d *SessionSJsonCfg

	v1 := &SessionSCfg{
		Enabled:             false,
		ListenBijson:        "*bijson_rpc",
		ListenBigob:         "*bigob_rpc",
		ChargerSConns:       []string{"*localhost"},
		ResourceSConns:      []string{"*localhost"},
		ThresholdSConns:     []string{"*localhost"},
		StatSConns:          []string{"*localhost"},
		RouteSConns:         []string{"*localhost"},
		CDRsConns:           []string{"*localhost"},
		ReplicationConns:    []string{"*localhost"},
		AttributeSConns:     []string{"*localhost"},
		RateSConns:          []string{"*localhost"},
		AccountSConns:       []string{"*localhost"},
		StoreSCosts:         false,
		SessionIndexes:      nil,
		ClientProtocol:      12.2,
		ChannelSyncInterval: 1 * time.Second,
		TerminateAttempts:   3,
		AlterableFields:     nil,
		MinDurLowBalance:    1 * time.Second,
		ActionSConns:        []string{"*localhost"},
		DefaultUsage: map[string]time.Duration{
			"DFLT_1": 1 * time.Second,
		},
		STIRCfg: &STIRcfg{
			AllowedAttest: utils.StringSet{
				"A_TEST1": {},
			},
			PayloadMaxduration: 2 * time.Second,
			DefaultAttest:      "default_attest",
			PublicKeyPath:      "/public/key/path",
			PrivateKeyPath:     "/private/key/path",
		},
		Opts: &SessionsOpts{
			DebitInterval: []*DynamicDurationOpt{
				{
					value: time.Second,
				},
			},
			TTL: []*DynamicDurationOpt{
				{
					value: time.Second,
				},
			},
			TTLMaxDelay: []*DynamicDurationOpt{
				{
					value: time.Second,
				},
			},
			TTLLastUsage: []*DynamicDurationPointerOpt{
				{
					value: utils.DurationPointer(time.Second),
				},
			},
			TTLLastUsed: []*DynamicDurationPointerOpt{
				{
					value: utils.DurationPointer(time.Second),
				},
			},
			TTLUsage: []*DynamicDurationPointerOpt{
				{
					value: utils.DurationPointer(time.Second),
				},
			},
		},
	}

	v2 := &SessionSCfg{
		Enabled:          true,
		ListenBijson:     "*bijson",
		ListenBigob:      "*bigob",
		ChargerSConns:    []string{"*birpc"},
		ResourceSConns:   []string{"*birpc"},
		ThresholdSConns:  []string{"*birpc"},
		StatSConns:       []string{"*birpc"},
		RouteSConns:      []string{"*birpc"},
		CDRsConns:        []string{"*birpc"},
		ReplicationConns: []string{"*birpc"},
		AttributeSConns:  []string{"*birpc"},
		RateSConns:       []string{"*birpc"},
		AccountSConns:    []string{"*birpc"},
		StoreSCosts:      true,
		SessionIndexes: utils.StringSet{
			"index1": struct{}{},
		},
		ClientProtocol:      13.2,
		ChannelSyncInterval: 2 * time.Second,
		TerminateAttempts:   5,
		AlterableFields: utils.StringSet{
			"index1": struct{}{},
		},
		MinDurLowBalance: 2 * time.Second,
		ActionSConns:     []string{"*birpc"},
		DefaultUsage: map[string]time.Duration{
			"DFLT_1": 2 * time.Second,
		},
		STIRCfg: &STIRcfg{
			AllowedAttest:      nil,
			PayloadMaxduration: 4 * time.Second,
			DefaultAttest:      "default_attest2",
			PublicKeyPath:      "/public/key/path/2",
			PrivateKeyPath:     "/private/key/path/2",
		},
		Opts: &SessionsOpts{
			DebitInterval: []*DynamicDurationOpt{
				{
					value: 2 * time.Second,
				},
			},
			TTL: []*DynamicDurationOpt{
				{
					value: 2 * time.Second,
				},
			},
			TTLMaxDelay: []*DynamicDurationOpt{
				{
					value: 2 * time.Second,
				},
			},
			TTLLastUsage: []*DynamicDurationPointerOpt{
				{
					value: utils.DurationPointer(2 * time.Second),
				},
			},
			TTLLastUsed: []*DynamicDurationPointerOpt{
				{
					value: utils.DurationPointer(2 * time.Second),
				},
			},
			TTLUsage: []*DynamicDurationPointerOpt{
				{
					value: utils.DurationPointer(2 * time.Second),
				},
			},
		},
	}

	expected := &SessionSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Listen_bijson:         utils.StringPointer("*bijson"),
		Listen_bigob:          utils.StringPointer("*bigob"),
		Chargers_conns:        &[]string{"*birpc"},
		Resources_conns:       &[]string{"*birpc"},
		Thresholds_conns:      &[]string{"*birpc"},
		Stats_conns:           &[]string{"*birpc"},
		Routes_conns:          &[]string{"*birpc"},
		Cdrs_conns:            &[]string{"*birpc"},
		Replication_conns:     &[]string{"*birpc"},
		Attributes_conns:      &[]string{"*birpc"},
		Rates_conns:           &[]string{"*birpc"},
		Accounts_conns:        &[]string{"*birpc"},
		Store_session_costs:   utils.BoolPointer(true),
		Session_indexes:       &[]string{"index1"},
		Client_protocol:       utils.Float64Pointer(13.2),
		Channel_sync_interval: utils.StringPointer("2s"),
		Terminate_attempts:    utils.IntPointer(5),
		Alterable_fields:      &[]string{"index1"},
		Min_dur_low_balance:   utils.StringPointer("2s"),
		Actions_conns:         &[]string{"*birpc"},
		Default_usage: map[string]string{
			"DFLT_1": "2s",
		},
		Stir: &STIRJsonCfg{
			Allowed_attest:      nil,
			Payload_maxduration: utils.StringPointer("4s"),
			Default_attest:      utils.StringPointer("default_attest2"),
			Publickey_path:      utils.StringPointer("/public/key/path/2"),
			Privatekey_path:     utils.StringPointer("/private/key/path/2"),
		},
		Opts: &SessionsOptsJson{
			DebitInterval: []*DynamicInterfaceOpt{
				{
					Value: 2 * time.Second,
				},
			},
			TTL: []*DynamicInterfaceOpt{
				{
					Value: 2 * time.Second,
				},
			},
			TTLMaxDelay: []*DynamicInterfaceOpt{
				{
					Value: 2 * time.Second,
				},
			},
			TTLLastUsage: []*DynamicInterfaceOpt{
				{
					Value: utils.DurationPointer(2 * time.Second),
				},
			},
			TTLLastUsed: []*DynamicInterfaceOpt{
				{
					Value: utils.DurationPointer(2 * time.Second),
				},
			},
			TTLUsage: []*DynamicInterfaceOpt{
				{
					Value: utils.DurationPointer(2 * time.Second),
				},
			},
		},
	}

	rcv := diffSessionSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2.Opts.TTLMaxDelay = nil
	v2.Opts.TTLLastUsed = nil
	v2.Opts.TTLLastUsage = nil
	v2.Opts.TTLUsage = nil

	expected.Opts.TTLMaxDelay = []*DynamicInterfaceOpt{}
	expected.Opts.TTLLastUsed = []*DynamicInterfaceOpt{}
	expected.Opts.TTLLastUsage = []*DynamicInterfaceOpt{}
	expected.Opts.TTLUsage = []*DynamicInterfaceOpt{}

	rcv = diffSessionSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestSessionSCloneSection(t *testing.T) {
	sessCfg := &SessionSCfg{
		Enabled:             false,
		ListenBijson:        "*bijson_rpc",
		ListenBigob:         "*bigob_rpc",
		ChargerSConns:       []string{"*localhost"},
		ResourceSConns:      []string{"*localhost"},
		ThresholdSConns:     []string{"*localhost"},
		StatSConns:          []string{"*localhost"},
		RouteSConns:         []string{"*localhost"},
		CDRsConns:           []string{"*localhost"},
		ReplicationConns:    []string{"*localhost"},
		AttributeSConns:     []string{"*localhost"},
		StoreSCosts:         false,
		SessionIndexes:      nil,
		ClientProtocol:      12.2,
		ChannelSyncInterval: 1 * time.Second,
		TerminateAttempts:   3,
		AlterableFields:     nil,
		MinDurLowBalance:    1 * time.Second,
		ActionSConns:        []string{"*localhost"},
		DefaultUsage: map[string]time.Duration{
			"DFLT_1": 1 * time.Second,
		},
		STIRCfg: &STIRcfg{
			AllowedAttest: utils.StringSet{
				"A_TEST1": {},
			},
			PayloadMaxduration: 2 * time.Second,
			DefaultAttest:      "default_attest",
			PublicKeyPath:      "/public/key/path",
			PrivateKeyPath:     "/private/key/path",
		},
		Opts: &SessionsOpts{
			DebitInterval: []*DynamicDurationOpt{
				{
					value: time.Second,
				},
			},
			TTL: []*DynamicDurationOpt{
				{
					value: time.Second,
				},
			},
			TTLMaxDelay: []*DynamicDurationOpt{
				{
					value: time.Second,
				},
			},
			TTLLastUsage: []*DynamicDurationPointerOpt{
				{
					value: utils.DurationPointer(time.Second),
				},
			},
			TTLLastUsed: []*DynamicDurationPointerOpt{
				{
					value: utils.DurationPointer(time.Second),
				},
			},
		},
	}

	exp := &SessionSCfg{
		Enabled:             false,
		ListenBijson:        "*bijson_rpc",
		ListenBigob:         "",
		ChargerSConns:       []string{"*localhost"},
		ResourceSConns:      []string{"*localhost"},
		ThresholdSConns:     []string{"*localhost"},
		StatSConns:          []string{"*localhost"},
		RouteSConns:         []string{"*localhost"},
		CDRsConns:           []string{"*localhost"},
		ReplicationConns:    []string{"*localhost"},
		AttributeSConns:     []string{"*localhost"},
		StoreSCosts:         false,
		SessionIndexes:      nil,
		ClientProtocol:      12.2,
		ChannelSyncInterval: 1 * time.Second,
		TerminateAttempts:   3,
		AlterableFields:     nil,
		MinDurLowBalance:    1 * time.Second,
		ActionSConns:        []string{"*localhost"},
		DefaultUsage: map[string]time.Duration{
			"DFLT_1": 1 * time.Second,
		},
		STIRCfg: &STIRcfg{
			AllowedAttest: utils.StringSet{
				"A_TEST1": {},
			},
			PayloadMaxduration: 2 * time.Second,
			DefaultAttest:      "default_attest",
			PublicKeyPath:      "/public/key/path",
			PrivateKeyPath:     "/private/key/path",
		},
		Opts: &SessionsOpts{
			DebitInterval: []*DynamicDurationOpt{
				{
					value: time.Second,
				},
			},
			TTL: []*DynamicDurationOpt{
				{
					value: time.Second,
				},
			},
			TTLMaxDelay: []*DynamicDurationOpt{
				{
					value: time.Second,
				},
			},
			TTLLastUsage: []*DynamicDurationPointerOpt{
				{
					value: utils.DurationPointer(time.Second),
				},
			},
			TTLLastUsed: []*DynamicDurationPointerOpt{
				{
					value: utils.DurationPointer(time.Second),
				},
			},
		},
	}

	rcv := sessCfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestDiffSessionsOptsJsonCfg(t *testing.T) {
	var d *SessionsOptsJson

	v1 := &SessionsOpts{
		Accounts: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		Attributes: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		CDRs: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		Chargers: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		Resources: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		Routes: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		Stats: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		Thresholds: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		Initiate: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		Update: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		Terminate: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		Message: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		AttributesDerivedReply: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		BlockerError: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		CDRsDerivedReply: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		ResourcesAuthorize: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		ResourcesAllocate: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		ResourcesRelease: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		ResourcesDerivedReply: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		RoutesDerivedReply: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		StatsDerivedReply: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		ThresholdsDerivedReply: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		MaxUsage: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		TTL: []*DynamicDurationOpt{
			{
				Tenant: "cgrates.org",
				value:  3 * time.Second,
			},
		},
		Chargeable: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.org",
				value:  false,
			},
		},
		TTLLastUsage: []*DynamicDurationPointerOpt{
			{
				Tenant: "cgrates.org",
				value:  utils.DurationPointer(5 * time.Second),
			},
		},
		TTLLastUsed: []*DynamicDurationPointerOpt{
			{
				Tenant: "cgrates.org",
				value:  utils.DurationPointer(5 * time.Second),
			},
		},
		DebitInterval: []*DynamicDurationOpt{
			{
				Tenant: "cgrates.org",
				value:  3 * time.Second,
			},
		},
		TTLMaxDelay: []*DynamicDurationOpt{
			{
				Tenant: "cgrates.org",
				value:  3 * time.Second,
			},
		},
		TTLUsage: []*DynamicDurationPointerOpt{
			{
				Tenant: "cgrates.org",
				value:  utils.DurationPointer(5 * time.Second),
			},
		},
	}

	v2 := &SessionsOpts{
		Accounts: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		Attributes: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		CDRs: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		Chargers: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		Resources: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		Routes: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		Stats: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		Thresholds: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		Initiate: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		Update: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		Terminate: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		Message: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		AttributesDerivedReply: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		BlockerError: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		CDRsDerivedReply: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		ResourcesAuthorize: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		ResourcesAllocate: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		ResourcesRelease: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		ResourcesDerivedReply: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		RoutesDerivedReply: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		StatsDerivedReply: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		ThresholdsDerivedReply: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		MaxUsage: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		TTL: []*DynamicDurationOpt{
			{
				Tenant: "cgrates.net",
				value:  4 * time.Second,
			},
		},
		Chargeable: []*DynamicBoolOpt{
			{
				Tenant: "cgrates.net",
				value:  true,
			},
		},
		TTLLastUsage: []*DynamicDurationPointerOpt{
			{
				Tenant: "cgrates.net",
				value:  utils.DurationPointer(6 * time.Second),
			},
		},
		TTLLastUsed: []*DynamicDurationPointerOpt{
			{
				Tenant: "cgrates.net",
				value:  utils.DurationPointer(6 * time.Second),
			},
		},
		DebitInterval: []*DynamicDurationOpt{
			{
				Tenant: "cgrates.net",
				value:  4 * time.Second,
			},
		},
		TTLMaxDelay: []*DynamicDurationOpt{
			{
				Tenant: "cgrates.net",
				value:  4 * time.Second,
			},
		},
		TTLUsage: []*DynamicDurationPointerOpt{
			{
				Tenant: "cgrates.net",
				value:  utils.DurationPointer(4 * time.Second),
			},
		},
	}

	expected := &SessionsOptsJson{
		Accounts: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		Attributes: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		CDRs: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		Chargers: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		Resources: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		Routes: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		Stats: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		Thresholds: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		Initiate: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		Update: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		Terminate: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		Message: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		AttributesDerivedReply: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		BlockerError: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		CDRsDerivedReply: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		ResourcesAuthorize: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		ResourcesAllocate: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		ResourcesRelease: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		ResourcesDerivedReply: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		RoutesDerivedReply: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		StatsDerivedReply: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		ThresholdsDerivedReply: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		MaxUsage: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},

		TTL: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  4 * time.Second,
			},
		},
		Chargeable: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  true,
			},
		},
		TTLLastUsage: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  utils.DurationPointer(6 * time.Second),
			},
		},
		TTLLastUsed: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  utils.DurationPointer(6 * time.Second),
			},
		},
		DebitInterval: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  4 * time.Second,
			},
		},
		TTLMaxDelay: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  4 * time.Second,
			},
		},
		TTLUsage: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.net",
				Value:  utils.DurationPointer(4 * time.Second),
			},
		},
	}

	rcv := diffSessionsOptsJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
