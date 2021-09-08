//go:build integration
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
	"path"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func TestDataDBReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)

	shdWg := new(sync.WaitGroup)
	chS := engine.NewCacheS(cfg, nil, nil)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	close(chS.GetPrecacheChannel(utils.CacheAttributeProfiles))
	close(chS.GetPrecacheChannel(utils.CacheAttributeFilterIndexes))
	chSCh := make(chan *engine.CacheS, 1)
	chSCh <- chS
	css := &CacheService{cacheCh: chSCh}
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srvMngr := servmanager.NewServiceManager(cfg, shdWg, nil)
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, make(chan birpc.ClientConnector, 1), srvDep)
	srvMngr.AddServices(NewAttributeService(cfg, db,
		css, filterSChan, server, make(chan birpc.ClientConnector, 1), anz, &DispatcherService{srvsReload: make(map[string]chan struct{})}, srvDep),
		NewLoaderService(cfg, db, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep), db)
	ctx, cancel := context.WithCancel(context.TODO())
	srvMngr.StartServices(ctx, cancel)
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	cfg.AttributeSCfg().Enabled = true
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo")
	if err := cfg.V1ReloadConfig(context.Background(), &config.ReloadArgs{
		Section: config.DataDBJSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !db.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	oldcfg := &config.DataDbCfg{
		Type: utils.Mongo,
		Host: "127.0.0.1",
		Port: "27017",
		Name: "10",
		User: "cgrates",
		Opts: map[string]interface{}{
			utils.MongoQueryTimeoutCfg:       "10s",
			utils.RedisClusterOnDownDelayCfg: "0",
			utils.RedisClusterSyncCfg:        "5s",
			utils.RedisClusterCfg:            false,
			utils.RedisSentinelNameCfg:       "",
			utils.RedisTLS:                   false,
			utils.RedisClientCertificate:     "",
			utils.RedisClientKey:             "",
			utils.RedisCACertificate:         "",
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpt{
			utils.MetaAccounts: {
				Replicate: false,
				Remote:    false},
			utils.MetaActions: {
				Replicate: false,
				Remote:    false},
			utils.MetaResourceProfile: {
				Replicate: false,
				Remote:    false},
			utils.MetaStatQueues: {
				Replicate: false,
				Remote:    false},
			utils.MetaResources: {
				Replicate: false,
				Remote:    false},
			utils.MetaStatQueueProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaThresholds: {
				Replicate: false,
				Remote:    false},
			utils.MetaThresholdProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaFilters: {
				Replicate: false,
				Remote:    false},
			utils.MetaRouteProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaAttributeProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaDispatcherHosts: {
				Replicate: false,
				Remote:    false},
			utils.MetaChargerProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaDispatcherProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaLoadIDs: {
				Replicate: false,
				Remote:    false},
			utils.MetaIndexes: {
				Replicate: false,
				Remote:    false},
			utils.MetaRateProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaActionProfiles: {
				Replicate: false,
				Remote:    false},
		},
	}
	if !reflect.DeepEqual(oldcfg, db.oldDBCfg) {
		t.Errorf("Expected %s \n received:%s", utils.ToJSON(oldcfg), utils.ToJSON(db.oldDBCfg))
	}

	err := db.Reload(ctx, cancel)
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	cfg.AttributeSCfg().Enabled = false
	cfg.GetReloadChan(config.DataDBJSON) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	cancel()
	time.Sleep(10 * time.Millisecond)
}

func TestDataDBReloadBadType(t *testing.T) {
	cfg, err := config.NewCGRConfigFromPath(path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"))
	if err != nil {
		t.Fatal(err)
	}
	dbConn, err := engine.NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		dbConn.Flush("")
		dbConn.Close()
	}()

	err = dbConn.SetVersions(engine.Versions{
		utils.StatS:      4,
		utils.Accounts:   3,
		utils.Actions:    2,
		utils.Thresholds: 4,
		utils.Routes:     2,
		// old version for Attributes
		utils.Attributes:     5,
		utils.RQF:            5,
		utils.Resource:       1,
		utils.Subscribers:    1,
		utils.Chargers:       2,
		utils.Dispatchers:    2,
		utils.LoadIDsVrs:     1,
		utils.RateProfiles:   1,
		utils.ActionProfiles: 1,
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, srvDep)
	db.oldDBCfg = &config.DataDbCfg{
		Type: utils.Mongo,
		Host: "127.0.0.1",
		Port: "27017",
		Name: "10",
		User: "cgrates",
		Opts: map[string]interface{}{
			utils.MongoQueryTimeoutCfg:       "10s",
			utils.RedisClusterOnDownDelayCfg: "0",
			utils.RedisClusterSyncCfg:        "5s",
			utils.RedisClusterCfg:            false,
			utils.RedisSentinelNameCfg:       "",
			utils.RedisTLS:                   false,
			utils.RedisClientCertificate:     "",
			utils.RedisClientKey:             "",
			utils.RedisCACertificate:         "",
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpt{
			utils.MetaAccounts: {
				Replicate: false,
				Remote:    false},
			utils.MetaActions: {
				Replicate: false,
				Remote:    false},
			utils.MetaCronExp: {
				Replicate: false,
				Remote:    false},
			utils.MetaResourceProfile: {
				Replicate: false,
				Remote:    false},
			utils.MetaStatQueues: {
				Replicate: false,
				Remote:    false},
			utils.MetaResources: {
				Replicate: false,
				Remote:    false},
			utils.MetaStatQueueProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaThresholds: {
				Replicate: false,
				Remote:    false},
			utils.MetaThresholdProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaFilters: {
				Replicate: false,
				Remote:    false},
			utils.MetaRouteProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaAttributeProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaDispatcherHosts: {
				Replicate: false,
				Remote:    false},
			utils.MetaChargerProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaDispatcherProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaLoadIDs: {
				Replicate: false,
				Remote:    false},
			utils.MetaIndexes: {
				Replicate: false,
				Remote:    false},
			utils.MetaRateProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaActionProfiles: {
				Replicate: false,
				Remote:    false},
		},
	}
	cfg.DataDbCfg().Type = "dbtype"
	db.dm = nil
	ctx, cancel := context.WithCancel(context.TODO())
	err = db.Reload(ctx, cancel)
	if err == nil || err.Error() != "unsupported db_type <dbtype>" {
		t.Fatal(err)
	}

	cancel()
	time.Sleep(10 * time.Millisecond)
}

func TestDataDBReloadErrorMarsheler(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DBDataEncoding = ""
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, srvDep)

	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	db.oldDBCfg = &config.DataDbCfg{
		Type: utils.Mongo,
		Host: "127.0.0.1",
		Port: "27017",
		Name: "10",
		User: "cgrates",
		Opts: map[string]interface{}{
			utils.MongoQueryTimeoutCfg:       "10s",
			utils.RedisClusterOnDownDelayCfg: "0",
			utils.RedisClusterSyncCfg:        "5s",
			utils.RedisClusterCfg:            false,
			utils.RedisSentinelNameCfg:       "",
			utils.RedisTLS:                   false,
			utils.RedisClientCertificate:     "",
			utils.RedisClientKey:             "",
			utils.RedisCACertificate:         "",
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpt{
			utils.MetaAccounts: {
				Replicate: false,
				Remote:    false},
			utils.MetaActions: {
				Replicate: false,
				Remote:    false},
			utils.MetaCronExp: {
				Replicate: false,
				Remote:    false},
			utils.MetaResourceProfile: {
				Replicate: false,
				Remote:    false},
			utils.MetaStatQueues: {
				Replicate: false,
				Remote:    false},
			utils.MetaResources: {
				Replicate: false,
				Remote:    false},
			utils.MetaStatQueueProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaThresholds: {
				Replicate: false,
				Remote:    false},
			utils.MetaThresholdProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaFilters: {
				Replicate: false,
				Remote:    false},
			utils.MetaRouteProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaAttributeProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaDispatcherHosts: {
				Replicate: false,
				Remote:    false},
			utils.MetaChargerProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaDispatcherProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaLoadIDs: {
				Replicate: false,
				Remote:    false},
			utils.MetaIndexes: {
				Replicate: false,
				Remote:    false},
			utils.MetaRateProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaActionProfiles: {
				Replicate: false,
				Remote:    false},
		},
	}

	ctx, cancel := context.WithCancel(context.TODO())
	err := db.Reload(ctx, cancel)
	if err == nil || err.Error() != "Unsupported marshaler: " {
		t.Fatal(err)
	}
	cancel()
	time.Sleep(10 * time.Millisecond)
}

func TestDataDBStartVersion(t *testing.T) {
	cfg, err := config.NewCGRConfigFromPath(path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"))
	if err != nil {
		t.Fatal(err)
	}
	dbConn, err := engine.NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		dbConn.Flush("")
		dbConn.Close()
	}()
	err = dbConn.SetVersions(engine.Versions{
		utils.StatS:      4,
		utils.Accounts:   3,
		utils.Actions:    2,
		utils.Thresholds: 4,
		utils.Routes:     2,
		// old version for Attributes
		utils.Attributes:     5,
		utils.RQF:            5,
		utils.Resource:       1,
		utils.Subscribers:    1,
		utils.Chargers:       2,
		utils.Dispatchers:    2,
		utils.LoadIDsVrs:     1,
		utils.RateProfiles:   1,
		utils.ActionProfiles: 1,
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, srvDep)
	ctx, cancel := context.WithCancel(context.TODO())
	err = db.Start(ctx, cancel)
	if err == nil || err.Error() != "Migration needed: please backup cgr data and run : <cgr-migrator -exec=*attributes>" {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", "Migration needed: please backup cgr data and run : <cgr-migrator -exec=*attributes>", err)
	}
	cancel()
	time.Sleep(10 * time.Millisecond)
}

func TestDataDBReloadCastError(t *testing.T) {
	cfg, err := config.NewCGRConfigFromPath(path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"))
	if err != nil {
		t.Fatal(err)
	}
	dbConn, err := engine.NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		dbConn.Flush("")
		dbConn.Close()
	}()

	err = dbConn.SetVersions(engine.Versions{
		utils.StatS:      4,
		utils.Accounts:   3,
		utils.Actions:    2,
		utils.Thresholds: 4,
		utils.Routes:     2,
		// old version for Attributes
		utils.Attributes:     5,
		utils.RQF:            5,
		utils.Resource:       1,
		utils.Subscribers:    1,
		utils.Chargers:       2,
		utils.Dispatchers:    2,
		utils.LoadIDsVrs:     1,
		utils.RateProfiles:   1,
		utils.ActionProfiles: 1,
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, srvDep)
	db.oldDBCfg = &config.DataDbCfg{
		Type: utils.Mongo,
		Host: "127.0.0.1",
		Port: "27017",
		Name: "10",
		User: "cgrates",
		Opts: map[string]interface{}{
			utils.MongoQueryTimeoutCfg:       "10s",
			utils.RedisClusterOnDownDelayCfg: "0",
			utils.RedisClusterSyncCfg:        "5s",
			utils.RedisClusterCfg:            false,
			utils.RedisSentinelNameCfg:       "",
			utils.RedisTLS:                   false,
			utils.RedisClientCertificate:     "",
			utils.RedisClientKey:             "",
			utils.RedisCACertificate:         "",
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpt{
			utils.MetaAccounts: {
				Replicate: false,
				Remote:    false},
			utils.MetaActions: {
				Replicate: false,
				Remote:    false},
			utils.MetaCronExp: {
				Replicate: false,
				Remote:    false},
			utils.MetaResourceProfile: {
				Replicate: false,
				Remote:    false},
			utils.MetaStatQueues: {
				Replicate: false,
				Remote:    false},
			utils.MetaResources: {
				Replicate: false,
				Remote:    false},
			utils.MetaStatQueueProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaThresholds: {
				Replicate: false,
				Remote:    false},
			utils.MetaThresholdProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaFilters: {
				Replicate: false,
				Remote:    false},
			utils.MetaRouteProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaAttributeProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaDispatcherHosts: {
				Replicate: false,
				Remote:    false},
			utils.MetaChargerProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaDispatcherProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaLoadIDs: {
				Replicate: false,
				Remote:    false},
			utils.MetaIndexes: {
				Replicate: false,
				Remote:    false},
			utils.MetaRateProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaActionProfiles: {
				Replicate: false,
				Remote:    false},
		},
	}

	db.dm = nil
	ctx, cancel := context.WithCancel(context.TODO())
	err = db.Reload(ctx, cancel)
	if err == nil || err.Error() != "can't conver DataDB of type mongo to MongoStorage" {
		t.Fatal(err)
	}

	cancel()
	time.Sleep(10 * time.Millisecond)
}

func TestDataDBReloadIfaceAsDurationError(t *testing.T) {
	cfg, err := config.NewCGRConfigFromPath(path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"))
	if err != nil {
		t.Fatal(err)
	}
	dbConn, err := engine.NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		dbConn.Flush("")
		dbConn.Close()
	}()

	err = dbConn.SetVersions(engine.Versions{
		utils.StatS:      4,
		utils.Accounts:   3,
		utils.Actions:    2,
		utils.Thresholds: 4,
		utils.Routes:     2,
		// old version for Attributes
		utils.Attributes:     5,
		utils.RQF:            5,
		utils.Resource:       1,
		utils.Subscribers:    1,
		utils.Chargers:       2,
		utils.Dispatchers:    2,
		utils.LoadIDsVrs:     1,
		utils.RateProfiles:   1,
		utils.ActionProfiles: 1,
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, srvDep)
	db.oldDBCfg = &config.DataDbCfg{
		Type: utils.Mongo,
		Host: "127.0.0.1",
		Port: "27017",
		Name: "10",
		User: "cgrates",
		Opts: map[string]interface{}{
			utils.MongoQueryTimeoutCfg:       "10s",
			utils.RedisClusterOnDownDelayCfg: "0",
			utils.RedisClusterSyncCfg:        "5s",
			utils.RedisClusterCfg:            false,
			utils.RedisSentinelNameCfg:       "",
			utils.RedisTLS:                   false,
			utils.RedisClientCertificate:     "",
			utils.RedisClientKey:             "",
			utils.RedisCACertificate:         "",
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpt{
			utils.MetaAccounts: {
				Replicate: false,
				Remote:    false},
			utils.MetaActions: {
				Replicate: false,
				Remote:    false},
			utils.MetaCronExp: {
				Replicate: false,
				Remote:    false},
			utils.MetaResourceProfile: {
				Replicate: false,
				Remote:    false},
			utils.MetaStatQueues: {
				Replicate: false,
				Remote:    false},
			utils.MetaResources: {
				Replicate: false,
				Remote:    false},
			utils.MetaStatQueueProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaThresholds: {
				Replicate: false,
				Remote:    false},
			utils.MetaThresholdProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaFilters: {
				Replicate: false,
				Remote:    false},
			utils.MetaRouteProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaAttributeProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaDispatcherHosts: {
				Replicate: false,
				Remote:    false},
			utils.MetaChargerProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaDispatcherProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaLoadIDs: {
				Replicate: false,
				Remote:    false},
			utils.MetaIndexes: {
				Replicate: false,
				Remote:    false},
			utils.MetaRateProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaActionProfiles: {
				Replicate: false,
				Remote:    false},
		},
	}
	cfg.DataDbCfg().Opts[utils.MongoQueryTimeoutCfg] = true
	db.dm = nil
	ctx, cancel := context.WithCancel(context.TODO())
	err = db.Reload(ctx, cancel)
	if err == nil || err.Error() != "cannot convert field: true to time.Duration" {
		t.Fatal(err)
	}

	cancel()
	time.Sleep(10 * time.Millisecond)
}

func TestDataDBStartSessionSCfgErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, srvDep)
	cfg.DataDbCfg().Type = "badtype"
	cfg.SessionSCfg().Enabled = true
	cfg.SessionSCfg().ListenBijson = ""
	ctx, cancel := context.WithCancel(context.TODO())
	err := db.Start(ctx, cancel)
	if err != nil {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}

func TestDataDBStartAttributeSCfgErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, srvDep)
	cfg.DataDbCfg().Type = "badtype"
	cfg.AttributeSCfg().Enabled = true
	cfg.SessionSCfg().ListenBijson = ""
	ctx, cancel := context.WithCancel(context.TODO())
	err := db.Start(ctx, cancel)
	if err == nil || err.Error() != "unsupported db_type <badtype>" {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", "unsupported db_type <badtype>", err)
	}
}

func TestDataDBReloadError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, srvDep)
	cfg.GeneralCfg().DBDataEncoding = utils.JSON
	db.oldDBCfg = &config.DataDbCfg{
		Type: utils.Mongo,
		Host: "127.0.0.1",
		Port: "27017",
		Name: "10",
		User: "cgrates",
		Opts: map[string]interface{}{
			utils.MongoQueryTimeoutCfg:       "10s",
			utils.RedisClusterOnDownDelayCfg: "0",
			utils.RedisClusterSyncCfg:        "5s",
			utils.RedisClusterCfg:            false,
			utils.RedisSentinelNameCfg:       "",
			utils.RedisTLS:                   false,
			utils.RedisClientCertificate:     "",
			utils.RedisClientKey:             "",
			utils.RedisCACertificate:         "",
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpt{
			utils.MetaAccounts: {
				Replicate: false,
				Remote:    false},
			utils.MetaActions: {
				Replicate: false,
				Remote:    false},
			utils.MetaCronExp: {
				Replicate: false,
				Remote:    false},
			utils.MetaResourceProfile: {
				Replicate: false,
				Remote:    false},
			utils.MetaStatQueues: {
				Replicate: false,
				Remote:    false},
			utils.MetaResources: {
				Replicate: false,
				Remote:    false},
			utils.MetaStatQueueProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaThresholds: {
				Replicate: false,
				Remote:    false},
			utils.MetaThresholdProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaFilters: {
				Replicate: false,
				Remote:    false},
			utils.MetaRouteProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaAttributeProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaDispatcherHosts: {
				Replicate: false,
				Remote:    false},
			utils.MetaChargerProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaDispatcherProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaLoadIDs: {
				Replicate: false,
				Remote:    false},
			utils.MetaIndexes: {
				Replicate: false,
				Remote:    false},
			utils.MetaRateProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaActionProfiles: {
				Replicate: false,
				Remote:    false},
		},
	}
	data := engine.NewInternalDB(nil, nil, true)
	db.dm = engine.NewDataManager(data, nil, nil)
	ctx, cancel := context.WithCancel(context.TODO())
	err := db.Reload(ctx, cancel)
	if err != nil {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}
