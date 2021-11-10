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

package ees

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewRpcEE(t *testing.T) {
	eeSCfg := config.NewDefaultCGRConfig().EEsCfg().GetDefaultExporter()
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig())

	rcv, err := NewRpcEE(eeSCfg, dc, connMgr)
	if err != nil {
		t.Error(err)
	}

	exp := &RPCee{
		cfg:     eeSCfg,
		dc:      dc,
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
	eeSCfg := config.NewDefaultCGRConfig().EEsCfg().GetDefaultExporter()
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig())
	rpcEe, err := NewRpcEE(eeSCfg, dc, connMgr)
	if err != nil {
		t.Error(err)
	}
	if err := rpcEe.Connect(); err != nil {
		t.Error(err)
	}
}

// func TestRPCExportEvent(t *testing.T) {
// 	eeSCfg := config.NewDefaultCGRConfig().EEsCfg().GetDefaultExporter()
// 	dc, err := newEEMetrics("Local")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig())
// 	rpcEe, err := NewRpcEE(eeSCfg, dc, connMgr)
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
// 		APIOpts: map[string]interface{}{},
// 	}

// 	if err := rpcEe.ExportEvent(context.Background(), args, ""); err != nil {
// 		t.Error(err)
// 	}
// }

func TestRPCClose(t *testing.T) {
	eeSCfg := config.NewDefaultCGRConfig().EEsCfg().GetDefaultExporter()
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig())
	rpcEe, err := NewRpcEE(eeSCfg, dc, connMgr)
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
	eeSCfg := config.NewDefaultCGRConfig().EEsCfg().GetDefaultExporter()
	dc := &utils.SafeMapStorage{
		MapStorage: utils.MapStorage{
			"time":         "now",
			"just_a_field": "just_a_value",
		},
	}
	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig())
	rpcEe, err := NewRpcEE(eeSCfg, dc, connMgr)
	if err != nil {
		t.Error(err)
	}

	if rcv := rpcEe.GetMetrics(); !reflect.DeepEqual(rcv, dc) {
		t.Errorf("Expected %v \n but received \n %v", dc, rcv)
	}
}

func TestRPCPrepareMap(t *testing.T) {
	eeSCfg := config.NewDefaultCGRConfig().EEsCfg().GetDefaultExporter()
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig())
	rpcEe, err := NewRpcEE(eeSCfg, dc, connMgr)
	if err != nil {
		t.Error(err)
	}

	exp := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID1",
		Event: map[string]interface{}{
			utils.Usage: 21,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: "*attributes",
		},
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID1",
		Event: map[string]interface{}{
			utils.Usage: 21,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: "*attributes",
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

func TestRPCPrepareOrderMap(t *testing.T) {
	eeSCfg := config.NewDefaultCGRConfig().EEsCfg().GetDefaultExporter()
	dc := &utils.SafeMapStorage{
		MapStorage: utils.MapStorage{
			"time":         "now",
			"just_a_field": "just_a_value",
		},
	}
	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig())
	rpcEe, err := NewRpcEE(eeSCfg, dc, connMgr)
	if err != nil {
		t.Error(err)
	}

	onm := utils.NewOrderedNavigableMap()
	fullPath := &utils.FullPath{
		PathSlice: []string{utils.MetaReq, utils.MetaTenant},
		Path:      utils.MetaTenant,
	}
	val := &utils.DataLeaf{
		Data: "value1",
	}
	onm.Append(fullPath, val)

	rcv, err := rpcEe.PrepareOrderMap(onm)
	if err != nil {
		t.Error(err)
	}

	exp := map[string]interface{}{
		utils.MetaReq + "." + utils.MetaTenant: val.Data,
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %+v \n but received \n %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestRPCParseOpts(t *testing.T) {
	eeSCfg := &config.EventExporterCfg{
		Opts: &config.EventExporterOpts{
			RPCCodec:        utils.StringPointer(utils.MetaJSON),
			ServiceMethod:   utils.StringPointer(utils.APIerSv1ExportCDRs),
			KeyPath:         utils.StringPointer("/key/path"),
			CertPath:        utils.StringPointer("/cert/path"),
			CAPath:          utils.StringPointer("/ca/path"),
			TLS:             utils.BoolPointer(true),
			ConnIDs:         &[]string{"*internal", "*birpc"},
			RPCConnTimeout:  utils.DurationPointer(2 * time.Second),
			RPCReplyTimeout: utils.DurationPointer(3 * time.Second),
		},
	}
	dc := &utils.SafeMapStorage{
		MapStorage: utils.MapStorage{
			"time":         "now",
			"just_a_field": "just_a_value",
		},
	}
	connMgr := engine.NewConnManager(config.NewDefaultCGRConfig())
	rpcEe, err := NewRpcEE(eeSCfg, dc, connMgr)
	if err != nil {
		t.Error(err)
	}

	exp := &RPCee{
		cfg:           eeSCfg,
		dc:            dc,
		connMgr:       connMgr,
		codec:         utils.MetaJSON,
		serviceMethod: utils.APIerSv1ExportCDRs,
		keyPath:       "/key/path",
		certPath:      "/cert/path",
		caPath:        "/ca/path",
		tls:           true,
		connIDs:       []string{"*internal", "*birpc"},
		connTimeout:   2 * time.Second,
		replyTimeout:  3 * time.Second,
	}

	if err := rpcEe.parseOpts(); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(rpcEe, exp) {
		t.Errorf("Expected %+v \n but received \n %+v", exp, rpcEe)
	}
}
