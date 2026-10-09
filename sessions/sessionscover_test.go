// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"bytes"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

type mkCall struct{}

func (sT *mkCall) Call(ctx *context.Context, method string, arg any, rply any) error {
	if arg.(*utils.DPRArgs).OriginHost != "cgrates" {
		return utils.ErrNoActiveSession
	}
	return nil
}

func TestBiRPCv1DisconnectPeer(t *testing.T) {
	client := new(mkCall)
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, nil, nil)

	sessions.biJIDs = map[string]*biJClient{
		"client1": {
			conn: client,
		},
	}
	args := &utils.DPRArgs{
		OriginHost: "cgrates",
	}
	var reply string
	if err := sessions.BiRPCv1DisconnectPeer(nil, args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply result")
	}

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	args.OriginHost = "changed_host"
	expLogg := "[WARNING] <SessionS> failed sending DPR for connection with id: <client1>, err: <NO_ACTIVE_SESSION>"
	if err := sessions.BiRPCv1DisconnectPeer(nil, args, &reply); err == nil || err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %+v, received %+v", utils.ErrPartiallyExecuted, err)
	} else if rcv := buf.String(); !strings.Contains(rcv, expLogg) {
		t.Errorf("Expected %+v, received %+v", expLogg, rcv)
	}
}

type mkCallForces struct{}

func (sT *mkCallForces) Call(ctx *context.Context, method string, arg any, rply any) error {
	return utils.ErrNoActiveSession
}

func TestBiRPCv1ForceDisconnect(t *testing.T) {
	ctx := context.WithClient(context.Background(), new(mkCall))
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, nil, nil)

	var reply string
	if err := sessions.BiRPCv1ForceDisconnect(ctx, nil, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
	args := &utils.SessionFilter{
		Tenant:  "cgrates.org",
		Filters: []string{"*string:~*opts.Disconnected:*asap"},
		Limit:   utils.IntPointer(7),
	}
	sessions.dm = nil

	if err := sessions.BiRPCv1ForceDisconnect(ctx, args, &reply); err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %+v, received %+v", utils.ErrNoDatabaseConn, err)
	}
	sessions.dm = dm
	args.Filters = nil
	sessions.aSessions = map[string]*Session{
		"sess1": {

			ID: "1001",
			sRuns: map[string]*SRun{
				"Id": {
					ID: "Id",
				},
			},
		},
		"CGRATES_ID": {

			ClientConnID: "client1",
		},
	}
	time.Sleep(50 * time.Millisecond)

	if err := sessions.BiRPCv1ForceDisconnect(ctx, args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	testMk := &mkCallForces{}
	clntConn := make(chan birpc.ClientConnector, 1)
	clntConn <- testMk
	sessions.cfg.SessionSCfg().Conns[utils.MetaResources] = []*config.DynamicConns{
		{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}},
	}
	connMngr := engine.NewConnManager(cfg)
	connMngr.SetCache(cacheS)
	connMngr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), utils.ResourceSv1, clntConn)
	sessions.connMgr = connMngr
	sessions.biJIDs = map[string]*biJClient{
		"client1": {
			conn: testMk,
		},
	}
	if err := sessions.BiRPCv1ForceDisconnect(ctx, args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
}

func TestSyncSessionsSync(t *testing.T) {
	log.SetOutput(io.Discard)
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheClosedSessions] = &config.CacheParamCfg{
		Replicate: true,
	}
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr, locker)
	dm.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, nil, connMgr)
	sessions.aSessions = map[string]*Session{}
	sessions.cfg.GeneralCfg().ReplyTimeout = 1
	var reply string
	if err := sessions.BiRPCv1SyncSessions(nil, nil, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected to be OK")
	}
}
