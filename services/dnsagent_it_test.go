//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func TestDNSAgentReload(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.SessionSCfg().Enabled = true
	utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	engineShutdown := make(chan bool, 1)
	chS := engine.NewCacheS(cfg, nil)

	cacheSChan := make(chan birpc.ClientConnector, 1)
	cacheSChan <- chS

	server := utils.NewServer()
	srvMngr := servmanager.NewServiceManager(cfg, engineShutdown)
	db := NewDataDBService(cfg, nil)
	sS := NewSessionService(cfg, db, server, make(chan birpc.ClientConnector, 1),
		engineShutdown, nil)
	srv := NewDNSAgent(cfg, filterSChan, engineShutdown, nil)
	engine.NewConnManager(cfg, nil)
	srvMngr.AddServices(srv, sS,
		NewLoaderService(cfg, db, filterSChan, server, engineShutdown, make(chan birpc.ClientConnector, 1), nil), db)
	if err = srvMngr.StartServices(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	if err := cfg.V1ReloadConfigFromPath(&config.ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "dnsagent_reload"),
		Section: config.DNSAgentJson,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !srv.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	cfg.DNSAgentCfg().Enabled = false
	cfg.GetReloadChan(config.DNSAgentJson) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	engineShutdown <- true
	time.Sleep(10 * time.Millisecond) // wait to stop session bijsonrpc server
}
