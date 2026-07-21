// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"reflect"
	"testing"

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
	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig())
	connMgr.SetCache(engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil))

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
		serviceMethod: utils.AdminSv1ComputeFilterIndexIDs,
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
	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig())
	connMgr.SetCache(engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil))
	rpcEe, err := NewRpcEE(eeSCfg, em, connMgr)
	if err != nil {
		t.Error(err)
	}
	if err := rpcEe.Connect(); err != nil {
		t.Error(err)
	}
}

// func TestRPCExportEvent(t *testing.T) {
// 	eeSCfg := config.NewDefaultCGRConfig().EEsCfg().ExporterCfg(utils.MetaDefault)
// 	em := utils.NewExporterMetrics("",time.Local)
// 	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig())
// 	rpcEe, err := NewRpcEE(eeSCfg, em, connMgr)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	// rpcEe.connMgr.

// 	// internalCacheSChann := make(chan birpc.ClientConnector, 1)
// 	// rpcEe.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaJSON, utils.MetaCaches), "", internalCacheSChann)
// 	rpcEe.connIDs = []string{utils.ConcatenatedKey(utils.MetaJSON, utils.MetaCaches)utils.MetaInternal}

// 	rpcEe.serviceMethod = utils.APIerSv1ExportToFolder
// 	args := &utils.TenantWithAPIOpts{
// 		Tenant:  "cgrates.org",
// 		APIOpts: map[string]any{},
// 	}

// 	if err := rpcEe.ExportEvent(context.Background(), args, ""); err != nil {
// 		t.Error(err)
// 	}
// }

func TestRPCClose(t *testing.T) {
	eeSCfg := config.NewDefaultCGRConfig().EEsCfg().ExporterCfg(utils.MetaDefault)
	em, err := utils.NewExporterMetrics("", "Local")
	if err != nil {
		t.Fatal(err)
	}
	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig())
	connMgr.SetCache(engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil))
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
	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig())
	connMgr.SetCache(engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil))
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
	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig())
	connMgr.SetCache(engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil))
	rpcEe, err := NewRpcEE(eeSCfg, em, connMgr)
	if err != nil {
		t.Error(err)
	}

	exp := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID1",
		Event: map[string]any{
			utils.Usage: 21,
		},
		APIOpts: map[string]any{
			utils.MetaSubsys: "*attributes",
		},
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID1",
		Event: map[string]any{
			utils.Usage: 21,
		},
		APIOpts: map[string]any{
			utils.MetaSubsys: "*attributes",
		},
	}

	rcv, err := rpcEe.PrepareMap(cgrEv)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %+v \n but received \n %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
