// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"context"
	"testing"

	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func TestNewRPCServiceAddsPing(t *testing.T) {
	srv, err := NewRPCService(new(rpcTestReceiver), "TestSv1")
	if err != nil {
		t.Fatal(err)
	}
	if _, has := srv.Methods[utils.Ping]; !has {
		t.Fatal("missing Ping")
	}
	var reply string
	if err := srv.Call(context.Background(), "TestSv1.Ping", &utils.CGREvent{}, &reply); err != nil {
		t.Fatal(err)
	}
	if reply != utils.Pong {
		t.Errorf("Ping reply = %q, want %q", reply, utils.Pong)
	}
}

func TestConfigServiceDuplicateRegistration(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cl := NewCommonListenerService(cfg)
	cl.cls = commonlisteners.NewCommonListenerS(nil)
	cm := NewConnManagerService(cfg)
	cm.connMgr = engine.NewConnManager(cfg)
	cm.anz = NewAnalyzerService(cfg)
	registry := servmanager.NewRegistry()
	registry.Register(cl, cm)
	for _, name := range []string{utils.CommonListenerS, utils.ConnManager} {
		if err := registry.SetState(name, utils.StateServiceUP); err != nil {
			t.Fatal(err)
		}
	}
	svc := NewConfigService(cfg)
	shutdown := utils.NewSyncedChan()
	if err := svc.Start(shutdown, registry); err != nil {
		t.Fatal(err)
	}
	if err := svc.Start(shutdown, registry); err == nil {
		t.Fatal("duplicate ConfigSv1 registration succeeded")
	}
}

func TestAddInternalRoute(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cms := NewConnManagerService(cfg)
	cms.connMgr = engine.NewConnManager(cfg)
	cms.connMgr.SetCache(engine.NewCacheS(cfg, nil, nil, nil, engine.NewLocker(cfg)))
	cms.anz = NewAnalyzerService(cfg)
	srv, err := NewRPCService(new(rpcTestReceiver), "TestSv1")
	if err != nil {
		t.Fatal(err)
	}
	connID := "*internal:*tests"
	cms.AddInternalRoute("TestS", connID, srv)
	var reply string
	if err := cms.ConnManager().Call(t.Context(), []string{connID}, "TestSv1.Status", &utils.CGREvent{}, &reply); err != nil || reply != utils.OK {
		t.Fatalf("Status = %q, %v", reply, err)
	}
}

type rpcTestReceiver struct{}

func (*rpcTestReceiver) Status(_ context.Context, _ *utils.CGREvent, reply *string) error {
	*reply = utils.OK
	return nil
}
