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

	shdChan := utils.NewSyncedChan()
	shdWg := new(sync.WaitGroup)
	chS := engine.NewCacheS(cfg, nil, nil)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	close(chS.GetPrecacheChannel(utils.CacheAttributeProfiles))
	close(chS.GetPrecacheChannel(utils.CacheAttributeFilterIndexes))
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg, nil)
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, false, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	srvMngr.AddServices(NewAttributeService(cfg, db,
		chS, filterSChan, server, make(chan birpc.ClientConnector, 1), anz, srvDep),
		NewLoaderService(cfg, db, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep), db)
	if err := srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	cfg.AttributeSCfg().Enabled = true
	if err := cfg.V1ReloadConfig(context.Background(),
		&config.ReloadArgs{
			Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
			Section: config.DATADB_JSN,
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !db.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	getDm := db.GetDM()
	if !reflect.DeepEqual(getDm, db.dm) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", db.dm, getDm)
	}
	oldcfg := &config.DataDbCfg{
		Type: utils.MetaMongo,
		Host: "127.0.0.1",
		Port: "27017",
		Name: "10",
		User: "cgrates",
		Opts: &config.DataDBOpts{
			InternalDBDumpPath:      "/var/lib/cgrates/internal_db/datadb",
			InternalDBBackupPath:    "/var/lib/cgrates/internal_db/backup/datadb",
			InternalDBStartTimeout:  5 * time.Minute,
			InternalDBWriteLimit:    100,
			MongoConnScheme:         "mongodb",
			RedisMaxConns:           10,
			RedisConnectAttempts:    20,
			RedisSentinel:           "",
			RedisCluster:            false,
			RedisClusterSync:        5 * time.Second,
			RedisClusterOndownDelay: 0,
			RedisPoolPipelineWindow: 150 * time.Microsecond,
			RedisConnectTimeout:     0,
			RedisReadTimeout:        0,
			RedisWriteTimeout:       0,
			MongoQueryTimeout:       10 * time.Second,
			RedisTLS:                false,
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpt{
			utils.MetaAccounts:            {Limit: -1},
			utils.MetaReverseDestinations: {Limit: -1},
			utils.MetaDestinations:        {Limit: -1},
			utils.MetaRatingPlans:         {Limit: -1},
			utils.MetaRatingProfiles:      {Limit: -1},
			utils.MetaActions:             {Limit: -1},
			utils.MetaActionPlans:         {Limit: -1},
			utils.MetaAccountActionPlans:  {Limit: -1},
			utils.MetaActionTriggers:      {Limit: -1},
			utils.MetaSharedGroups:        {Limit: -1},
			utils.MetaTimings:             {Limit: -1},
			utils.MetaResourceProfile:     {Limit: -1},
			utils.MetaStatQueues:          {Limit: -1},
			utils.MetaResources:           {Limit: -1},
			utils.MetaStatQueueProfiles:   {Limit: -1},
			utils.MetaRankings:            {Limit: -1},
			utils.MetaRankingProfiles:     {Limit: -1},
			utils.MetaTrends:              {Limit: -1},
			utils.MetaTrendProfiles:       {Limit: -1},
			utils.MetaThresholds:          {Limit: -1},
			utils.MetaThresholdProfiles:   {Limit: -1},
			utils.MetaFilters:             {Limit: -1},
			utils.MetaRouteProfiles:       {Limit: -1},
			utils.MetaAttributeProfiles:   {Limit: -1},
			utils.MetaDispatcherHosts:     {Limit: -1},
			utils.MetaChargerProfiles:     {Limit: -1},
			utils.MetaDispatcherProfiles:  {Limit: -1},
			utils.MetaLoadIDs:             {Limit: -1},
			utils.MetaSessionsBackup:      {Limit: -1},
			utils.CacheVersions:           {Limit: -1},

			utils.CacheResourceFilterIndexes:   {Limit: -1},
			utils.CacheStatFilterIndexes:       {Limit: -1},
			utils.CacheThresholdFilterIndexes:  {Limit: -1},
			utils.CacheRouteFilterIndexes:      {Limit: -1},
			utils.CacheAttributeFilterIndexes:  {Limit: -1},
			utils.CacheChargerFilterIndexes:    {Limit: -1},
			utils.CacheDispatcherFilterIndexes: {Limit: -1},
			utils.CacheReverseFilterIndexes:    {Limit: -1},
		},
	}
	if !reflect.DeepEqual(oldcfg, db.oldDBCfg) {
		t.Errorf("Expected %s \n received:%s", utils.ToJSON(oldcfg), utils.ToJSON(db.oldDBCfg))
	}

	err := db.Reload()
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	cfg.AttributeSCfg().Enabled = false
	cfg.GetReloadChan(config.DATADB_JSN) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	shdChan.CloseOnce()
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
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		dbConn.Flush("")
		dbConn.Close()
	}()

	err = dbConn.SetVersions(engine.Versions{
		utils.StatS:          4,
		utils.Accounts:       3,
		utils.Actions:        2,
		utils.ActionTriggers: 2,
		utils.ActionPlans:    3,
		utils.SharedGroups:   2,
		utils.Thresholds:     4,
		utils.Routes:         2,
		// old version for Attributes
		utils.Attributes:          5,
		utils.Timing:              1,
		utils.RQF:                 5,
		utils.Resource:            1,
		utils.Subscribers:         1,
		utils.Destinations:        1,
		utils.ReverseDestinations: 1,
		utils.RatingPlan:          1,
		utils.RatingProfile:       1,
		utils.Chargers:            2,
		utils.Dispatchers:         2,
		utils.LoadIDsVrs:          1,
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	shdChan := utils.NewSyncedChan()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, false, srvDep)
	db.oldDBCfg = &config.DataDbCfg{
		Type: utils.MetaMongo,
		Host: "127.0.0.1",
		Port: "27017",
		Name: "10",
		User: "cgrates",
		Opts: &config.DataDBOpts{
			RedisMaxConns:           10,
			RedisConnectAttempts:    20,
			RedisSentinel:           "",
			RedisCluster:            false,
			RedisClusterSync:        5 * time.Second,
			RedisClusterOndownDelay: 0,
			RedisPoolPipelineWindow: 150 * time.Microsecond,
			RedisConnectTimeout:     0,
			RedisReadTimeout:        0,
			RedisWriteTimeout:       0,
			MongoQueryTimeout:       10 * time.Second,
			RedisTLS:                false,
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpt{
			utils.MetaAccounts: {
				Replicate: false,
				Remote:    false},
			utils.MetaReverseDestinations: {
				Replicate: false,
				Remote:    false},
			utils.MetaDestinations: {
				Replicate: false,
				Remote:    false},
			utils.MetaRatingPlans: {
				Replicate: false,
				Remote:    false},
			utils.MetaRatingProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaActions: {
				Replicate: false,
				Remote:    false},
			utils.MetaActionPlans: {
				Replicate: false,
				Remote:    false},
			utils.MetaAccountActionPlans: {
				Replicate: false,
				Remote:    false},
			utils.MetaActionTriggers: {
				Replicate: false,
				Remote:    false},
			utils.MetaSharedGroups: {
				Replicate: false,
				Remote:    false},
			utils.MetaTimings: {
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
		},
	}
	cfg.DataDbCfg().Type = "dbtype"
	db.dm = nil
	err = db.Reload()
	if err == nil || err.Error() != "unsupported db_type <dbtype>" {
		t.Fatal(err)
	}

	shdChan.CloseOnce()
	time.Sleep(10 * time.Millisecond)
}

func TestDataDBReloadErrorMarsheler(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DBDataEncoding = ""
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	shdChan := utils.NewSyncedChan()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, false, srvDep)

	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	db.oldDBCfg = &config.DataDbCfg{
		Type: utils.MetaMongo,
		Host: "127.0.0.1",
		Port: "27017",
		Name: "10",
		User: "cgrates",
		Opts: &config.DataDBOpts{
			RedisMaxConns:           10,
			RedisConnectAttempts:    20,
			RedisSentinel:           "",
			RedisCluster:            false,
			RedisClusterSync:        5 * time.Second,
			RedisClusterOndownDelay: 0,
			RedisPoolPipelineWindow: 150 * time.Microsecond,
			RedisConnectTimeout:     0,
			RedisReadTimeout:        0,
			RedisWriteTimeout:       0,
			MongoQueryTimeout:       10 * time.Second,
			RedisTLS:                false,
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpt{
			utils.MetaAccounts: {
				Replicate: false,
				Remote:    false},
			utils.MetaReverseDestinations: {
				Replicate: false,
				Remote:    false},
			utils.MetaDestinations: {
				Replicate: false,
				Remote:    false},
			utils.MetaRatingPlans: {
				Replicate: false,
				Remote:    false},
			utils.MetaRatingProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaActions: {
				Replicate: false,
				Remote:    false},
			utils.MetaActionPlans: {
				Replicate: false,
				Remote:    false},
			utils.MetaAccountActionPlans: {
				Replicate: false,
				Remote:    false},
			utils.MetaActionTriggers: {
				Replicate: false,
				Remote:    false},
			utils.MetaSharedGroups: {
				Replicate: false,
				Remote:    false},
			utils.MetaTimings: {
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
		},
	}

	err := db.Reload()
	if err == nil || err.Error() != "Unsupported marshaler: " {
		t.Fatal(err)
	}
	shdChan.CloseOnce()
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
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		dbConn.Flush("")
		dbConn.Close()
	}()
	err = dbConn.SetVersions(engine.Versions{
		utils.StatS:          4,
		utils.Accounts:       3,
		utils.Actions:        2,
		utils.ActionTriggers: 2,
		utils.ActionPlans:    3,
		utils.SharedGroups:   2,
		utils.Thresholds:     4,
		utils.Routes:         2,
		// old version for Attributes
		utils.Attributes:          5,
		utils.Timing:              1,
		utils.RQF:                 5,
		utils.Resource:            1,
		utils.Subscribers:         1,
		utils.Destinations:        1,
		utils.ReverseDestinations: 1,
		utils.RatingPlan:          1,
		utils.RatingProfile:       1,
		utils.Chargers:            2,
		utils.Dispatchers:         2,
		utils.LoadIDsVrs:          1,
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)

	shdChan := utils.NewSyncedChan()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, false, srvDep)
	err = db.Start()
	if err == nil || err.Error() != "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*attributes>" {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", "Migration needed: please backup cgr data and run: <cgr-migrator -exec=*attributes>", err)
	}
	shdChan.CloseOnce()
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
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		dbConn.Flush("")
		dbConn.Close()
	}()

	err = dbConn.SetVersions(engine.Versions{
		utils.StatS:          4,
		utils.Accounts:       3,
		utils.Actions:        2,
		utils.ActionTriggers: 2,
		utils.ActionPlans:    3,
		utils.SharedGroups:   2,
		utils.Thresholds:     4,
		utils.Routes:         2,
		// old version for Attributes
		utils.Attributes:          5,
		utils.Timing:              1,
		utils.RQF:                 5,
		utils.Resource:            1,
		utils.Subscribers:         1,
		utils.Destinations:        1,
		utils.ReverseDestinations: 1,
		utils.RatingPlan:          1,
		utils.RatingProfile:       1,
		utils.Chargers:            2,
		utils.Dispatchers:         2,
		utils.LoadIDsVrs:          1,
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	shdChan := utils.NewSyncedChan()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, false, srvDep)
	db.oldDBCfg = &config.DataDbCfg{
		Type: utils.MetaMongo,
		Host: "127.0.0.1",
		Port: "27017",
		Name: "10",
		User: "cgrates",
		Opts: &config.DataDBOpts{
			RedisMaxConns:           10,
			RedisConnectAttempts:    20,
			RedisSentinel:           "",
			RedisCluster:            false,
			RedisClusterSync:        5 * time.Second,
			RedisClusterOndownDelay: 0,
			RedisPoolPipelineWindow: 150 * time.Microsecond,
			RedisConnectTimeout:     0,
			RedisReadTimeout:        0,
			RedisWriteTimeout:       0,
			MongoQueryTimeout:       10 * time.Second,
			RedisTLS:                false,
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpt{
			utils.MetaAccounts: {
				Replicate: false,
				Remote:    false},
			utils.MetaReverseDestinations: {
				Replicate: false,
				Remote:    false},
			utils.MetaDestinations: {
				Replicate: false,
				Remote:    false},
			utils.MetaRatingPlans: {
				Replicate: false,
				Remote:    false},
			utils.MetaRatingProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaActions: {
				Replicate: false,
				Remote:    false},
			utils.MetaActionPlans: {
				Replicate: false,
				Remote:    false},
			utils.MetaAccountActionPlans: {
				Replicate: false,
				Remote:    false},
			utils.MetaActionTriggers: {
				Replicate: false,
				Remote:    false},
			utils.MetaSharedGroups: {
				Replicate: false,
				Remote:    false},
			utils.MetaTimings: {
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
		},
	}

	db.dm = nil
	err = db.Reload()
	if err == nil || err.Error() != "can't conver DataDB of type *mongo to MongoStorage" {
		t.Fatal(err)
	}

	shdChan.CloseOnce()
	time.Sleep(10 * time.Millisecond)
}

func TestDataDBStartSessionSCfgErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, false, srvDep)
	cfg.DataDbCfg().Type = "badtype"
	cfg.SessionSCfg().Enabled = true
	cfg.SessionSCfg().ListenBijson = ""
	err := db.Start()
	if err != nil {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}

func TestDataDBStartRalsSCfgErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, false, srvDep)
	cfg.DataDbCfg().Type = "badtype"
	db.cfg.RalsCfg().Enabled = true
	cfg.SessionSCfg().ListenBijson = ""
	err := db.Start()
	if err == nil || err.Error() != "unsupported db_type <badtype>" {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", "unsupported db_type <badtype>", err)
	}
}

func TestDataDBReloadError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, false, srvDep)
	cfg.GeneralCfg().DBDataEncoding = utils.JSON
	db.oldDBCfg = &config.DataDbCfg{
		Type: utils.MetaMongo,
		Host: "127.0.0.1",
		Port: "27017",
		Name: "10",
		User: "cgrates",
		Opts: &config.DataDBOpts{
			RedisMaxConns:           10,
			RedisConnectAttempts:    20,
			RedisSentinel:           "",
			RedisCluster:            false,
			RedisClusterSync:        5 * time.Second,
			RedisClusterOndownDelay: 0,
			RedisPoolPipelineWindow: 150 * time.Microsecond,
			RedisConnectTimeout:     0,
			RedisReadTimeout:        0,
			RedisWriteTimeout:       0,
			MongoQueryTimeout:       10 * time.Second,
			RedisTLS:                false,
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpt{
			utils.MetaAccounts: {
				Replicate: false,
				Remote:    false},
			utils.MetaReverseDestinations: {
				Replicate: false,
				Remote:    false},
			utils.MetaDestinations: {
				Replicate: false,
				Remote:    false},
			utils.MetaRatingPlans: {
				Replicate: false,
				Remote:    false},
			utils.MetaRatingProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaActions: {
				Replicate: false,
				Remote:    false},
			utils.MetaActionPlans: {
				Replicate: false,
				Remote:    false},
			utils.MetaAccountActionPlans: {
				Replicate: false,
				Remote:    false},
			utils.MetaActionTriggers: {
				Replicate: false,
				Remote:    false},
			utils.MetaSharedGroups: {
				Replicate: false,
				Remote:    false},
			utils.MetaTimings: {
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
		},
	}
	data, derr := engine.NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	if derr != nil {
		t.Error(derr)
	}
	db.dm = engine.NewDataManager(data, nil, nil)
	err := db.Reload()
	if err != nil {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}
