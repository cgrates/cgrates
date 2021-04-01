// +build integration

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
package services

import (
	"fmt"
	"io"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func init() {
	log.SetOutput(io.Discard)
}

type testMockClients struct {
	calls func(args interface{}, reply interface{}) error
}

func (sT *testMockClients) Call(method string, arg interface{}, rply interface{}) error {
	return sT.calls(arg, rply)
}

func TestSessionSReload1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ChargerSCfg().Enabled = true
	cfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	cfg.RPCConns()["cache1"] = &config.RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*config.RemoteHost{
			{
				Address:     "127.0.0.1:9999",
				Transport:   utils.MetaGOB,
				Synchronous: true,
			},
		},
	}
	cfg.CacheCfg().ReplicationConns = []string{"cache1"}
	cfg.CacheCfg().Partitions[utils.CacheClosedSessions].Limit = 0
	cfg.CacheCfg().Partitions[utils.CacheClosedSessions].Replicate = true
	temporaryCache := engine.Cache
	defer func() {
		engine.Cache = temporaryCache
	}()
	engine.Cache = engine.NewCacheS(cfg, nil, nil)
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}

	clientConect := make(chan rpcclient.ClientConnector, 1)
	clientConect <- &testMockClients{
		calls: func(args interface{}, reply interface{}) error {
			rply, cancast := reply.(*[]*engine.ChrgSProcessEventReply)
			if !cancast {
				return fmt.Errorf("can't cast")
			}
			*rply = []*engine.ChrgSProcessEventReply{
				{
					ChargerSProfile:    "raw",
					AttributeSProfiles: []string{utils.MetaNone},
					AlteredFields:      []string{"~*req.RunID"},
					CGREvent:           args.(*utils.CGREvent),
				},
			}
			return nil
		},
	}
	conMng := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): clientConect,
	})
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	srv := NewSessionService(cfg, new(DataDBService), server, make(chan rpcclient.ClientConnector, 1), shdChan, conMng, nil, anz, srvDep)
	err := srv.Start()
	if err != nil {
		t.Fatal(err)
	}
	if !srv.IsRunning() {
		t.Fatal("Expected service to be running")
	}
	args := &sessions.V1InitSessionArgs{
		InitSession:       true,
		ProcessThresholds: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventInitiateSession",
			Event: map[string]interface{}{
				utils.Tenant:           "cgrates.org",
				utils.ToR:              utils.MetaVoice,
				utils.OriginID:         "testSSv1ItProcessEvent",
				utils.RequestType:      utils.MetaPostpaid,
				utils.AccountField:     "1001",
				utils.CGRDebitInterval: 10,
				utils.Destination:      "1002",
				utils.SetupTime:        time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:       time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:            0,
			},
		},
	}

	rply := new(sessions.V1InitSessionReply)
	srv.(*SessionService).sm.BiRPCv1InitiateSession(nil, args, rply)
	err = srv.Shutdown()
	if err == nil || err != utils.ErrPartiallyExecuted {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrPartiallyExecuted, err)
	}
}

func TestSessionSReload2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cfg.ChargerSCfg().Enabled = true
	cfg.RalsCfg().Enabled = true
	cfg.CdrsCfg().Enabled = true
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	close(chS.GetPrecacheChannel(utils.CacheChargerProfiles))
	close(chS.GetPrecacheChannel(utils.CacheChargerFilterIndexes))
	close(chS.GetPrecacheChannel(utils.CacheDestinations))
	close(chS.GetPrecacheChannel(utils.CacheReverseDestinations))
	close(chS.GetPrecacheChannel(utils.CacheActions))
	close(chS.GetPrecacheChannel(utils.CacheTimings))

	internalChan := make(chan rpcclient.ClientConnector, 1)
	internalChan <- nil
	cacheSChan := make(chan rpcclient.ClientConnector, 1)
	cacheSChan <- chS

	server := cores.NewServer(nil)

	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	cfg.StorDbCfg().Type = utils.Internal
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	srv := NewSessionService(cfg, db, server, make(chan rpcclient.ClientConnector, 1), shdChan, nil, nil, anz, srvDep)
	engine.NewConnManager(cfg, nil)

	srv.(*SessionService).sm = &sessions.SessionS{}
	if !srv.IsRunning() {
		t.Fatalf("\nExpecting service to be running")
	}
	err2 := srv.Start()
	if err2 != utils.ErrServiceAlreadyRunning {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err2)
	}
	cfg.SessionSCfg().Enabled = false
	err := srv.Reload()
	if err != nil {
		t.Fatalf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	time.Sleep(10 * time.Millisecond)
	srv.(*SessionService).sm = nil
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	shdChan.CloseOnce()
	time.Sleep(10 * time.Millisecond)

}

func TestSessionSReload3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cfg.ChargerSCfg().Enabled = true
	cfg.RalsCfg().Enabled = true
	cfg.CdrsCfg().Enabled = true
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	close(chS.GetPrecacheChannel(utils.CacheChargerProfiles))
	close(chS.GetPrecacheChannel(utils.CacheChargerFilterIndexes))
	close(chS.GetPrecacheChannel(utils.CacheDestinations))
	close(chS.GetPrecacheChannel(utils.CacheReverseDestinations))
	close(chS.GetPrecacheChannel(utils.CacheActions))
	close(chS.GetPrecacheChannel(utils.CacheTimings))

	internalChan := make(chan rpcclient.ClientConnector, 1)
	internalChan <- nil
	cacheSChan := make(chan rpcclient.ClientConnector, 1)
	cacheSChan <- chS

	server := cores.NewServer(nil)

	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	cfg.StorDbCfg().Type = utils.Internal
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	srv := NewSessionService(cfg, db, server, make(chan rpcclient.ClientConnector, 1), shdChan, nil, nil, anz, srvDep)
	engine.NewConnManager(cfg, nil)

	srv.(*SessionService).sm = &sessions.SessionS{}
	if !srv.IsRunning() {
		t.Fatalf("\nExpecting service to be running")
	}
	err2 := srv.(*SessionService).start()
	if err2 != nil {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", nil, err2)
	}

}
