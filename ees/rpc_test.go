// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewRpcEE(t *testing.T) {
	eeSCfg := config.NewDefaultCGRConfig().EEsCfg().ExporterCfg(utils.MetaDefault)
	em, err := utils.NewExporterMetrics("", "Local")
	if err != nil {
		t.Fatal(err)
	}
	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig(), make(map[string]chan birpc.ClientConnector))

	rcv, err := NewRpcEE(eeSCfg, em, connMgr)
	if err != nil {
		t.Error(err)
	}

	exp := &RPCee{
		cfg:     eeSCfg,
		em:      em,
		connMgr: connMgr,
	}

	err = exp.parseOpts()
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %+v \n but received \n %+v", exp, rcv)
	}
}

func TestRPCCfg(t *testing.T) {
	cfg := &RPCee{
		cfg: &config.EventExporterCfg{
			ID:                 utils.MetaDefault,
			Type:               utils.MetaNone,
			Attempts:           1,
			Opts:               new(config.EventExporterOpts),
			ExportPath:         "/var/spool/cgrates/ees",
			FailedPostsDir:     "/var/spool/cgrates/failed_posts",
			AttributeSIDs:      []string{},
			Fields:             []*config.FCTemplate{},
			Filters:            []string{},
			Flags:              utils.FlagsWithParams{},
			Synchronous:        false,
			Timezone:           "",
			ConcurrentRequests: 0,
		},
		codec:         utils.MetaJSON,
		serviceMethod: utils.APIerSv1ComputeFilterIndexes,
	}
	exp := &config.EventExporterCfg{
		ID:                 utils.MetaDefault,
		Type:               utils.MetaNone,
		Attempts:           1,
		Opts:               new(config.EventExporterOpts),
		ExportPath:         "/var/spool/cgrates/ees",
		FailedPostsDir:     "/var/spool/cgrates/failed_posts",
		AttributeSIDs:      []string{},
		Fields:             []*config.FCTemplate{},
		Filters:            []string{},
		Flags:              utils.FlagsWithParams{},
		Synchronous:        false,
		Timezone:           "",
		ConcurrentRequests: 0,
	}

	rcv := cfg.Cfg()
	rcv.HeaderFields()

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %+v \n but received \n %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestRPCConnect(t *testing.T) {
	eeSCfg := config.NewDefaultCGRConfig().EEsCfg().ExporterCfg(utils.MetaDefault)
	em, err := utils.NewExporterMetrics("", "Local")
	if err != nil {
		t.Fatal(err)
	}
	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig(), make(map[string]chan birpc.ClientConnector))
	rpcEe, err := NewRpcEE(eeSCfg, em, connMgr)
	if err != nil {
		t.Error(err)
	}
	if err := rpcEe.Connect(); err != nil {
		t.Error(err)
	}
}

func TestRPCClose(t *testing.T) {
	eeSCfg := config.NewDefaultCGRConfig().EEsCfg().ExporterCfg(utils.MetaDefault)
	em, err := utils.NewExporterMetrics("", "Local")
	if err != nil {
		t.Fatal(err)
	}
	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig(), make(map[string]chan birpc.ClientConnector))
	rpcEe, err := NewRpcEE(eeSCfg, em, connMgr)
	if err != nil {
		t.Error(err)
	}

	if err := rpcEe.Close(); err != nil {
		t.Error(err)
	} else if rpcEe.connMgr != nil {
		t.Errorf("Expected connMgr to be nil")
	}
}

func TestRPCGetMetrics(t *testing.T) {
	eeSCfg := config.NewDefaultCGRConfig().EEsCfg().ExporterCfg(utils.MetaDefault)
	em := &utils.ExporterMetrics{
		MapStorage: utils.MapStorage{
			"time":         "now",
			"just_a_field": "just_a_value",
		},
	}
	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig(), make(map[string]chan birpc.ClientConnector))
	rpcEe, err := NewRpcEE(eeSCfg, em, connMgr)
	if err != nil {
		t.Error(err)
	}

	if rcv := rpcEe.GetMetrics(); !reflect.DeepEqual(rcv, em) {
		t.Errorf("Expected %v \n but received \n %v", em, rcv)
	}
}

func TestRPCPrepareMap(t *testing.T) {
	eeSCfg := config.NewDefaultCGRConfig().EEsCfg().ExporterCfg(utils.MetaDefault)
	em, err := utils.NewExporterMetrics("", "Local")
	if err != nil {
		t.Fatal(err)
	}
	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig(), make(map[string]chan birpc.ClientConnector))
	rpcEe, err := NewRpcEE(eeSCfg, em, connMgr)
	if err != nil {
		t.Error(err)
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "CGRID1",
		Event: map[string]any{
			utils.Usage: 21,
		},
		APIOpts: map[string]any{
			utils.MetaSubsys: "*attributes",
		},
	}

	exp := cgrEv

	rcv, err := rpcEe.PrepareMap(cgrEv)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %+v \n but received \n %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestRPCParseOpts(t *testing.T) {
	rpcEE := &RPCee{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				RPC: &config.RPCOpts{
					RPCCodec:        utils.StringPointer("RPCCodec"),
					ServiceMethod:   utils.StringPointer("ServiceMethod"),
					KeyPath:         utils.StringPointer("KeyPath"),
					CertPath:        utils.StringPointer("CertPath"),
					CAPath:          utils.StringPointer("CAPath"),
					TLS:             utils.BoolPointer(true),
					ConnIDs:         utils.SliceStringPointer([]string{"ConnID"}),
					RPCConnTimeout:  utils.DurationPointer(time.Second),
					RPCReplyTimeout: utils.DurationPointer(time.Minute),
				},
			},
		},
	}

	exp := &RPCee{
		cfg: &config.EventExporterCfg{
			Opts: &config.EventExporterOpts{
				RPC: &config.RPCOpts{
					RPCCodec:        utils.StringPointer("RPCCodec"),
					ServiceMethod:   utils.StringPointer("ServiceMethod"),
					KeyPath:         utils.StringPointer("KeyPath"),
					CertPath:        utils.StringPointer("CertPath"),
					CAPath:          utils.StringPointer("CAPath"),
					TLS:             utils.BoolPointer(true),
					ConnIDs:         utils.SliceStringPointer([]string{"ConnID"}),
					RPCConnTimeout:  utils.DurationPointer(time.Second),
					RPCReplyTimeout: utils.DurationPointer(time.Minute),
				},
			},
		},
		codec:         "RPCCodec",
		serviceMethod: "ServiceMethod",
		keyPath:       "KeyPath",
		certPath:      "CertPath",
		caPath:        "CAPath",
		connIDs:       []string{"ConnID"},
		connTimeout:   time.Second,
		replyTimeout:  time.Minute,
		tls:           true,
	}
	if err := rpcEE.parseOpts(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rpcEE, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rpcEE)
	}
}
