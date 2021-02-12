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

	"github.com/cgrates/rpcclient"

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
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg)
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	srvMngr.AddServices(NewAttributeService(cfg, db,
		chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), anz, srvDep),
		NewLoaderService(cfg, db, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz, srvDep), db)
	if err := srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	cfg.AttributeSCfg().Enabled = true
	if err := cfg.V1ReloadConfig(&config.ReloadArgs{
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
		DataDbType: utils.Mongo,
		DataDbHost: "127.0.0.1",
		DataDbPort: "27017",
		DataDbName: "10",
		DataDbUser: "cgrates",
		Opts: map[string]interface{}{
			utils.QueryTimeoutCfg:            "10s",
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
			utils.MetaIndexes: {
				Replicate: false,
				Remote:    false},
			utils.MetaRateProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaActionProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaAccountProfiles: {
				Replicate: false,
				Remote:    false},
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
func TestDataDBReload2(t *testing.T) {
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
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg)
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	srvMngr.AddServices(NewAttributeService(cfg, db,
		chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), anz, srvDep),
		NewLoaderService(cfg, db, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz, srvDep), db)
	if err := srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	cfg.AttributeSCfg().Enabled = true
	if err := cfg.V1ReloadConfig(&config.ReloadArgs{
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

	err := db.Reload()
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	if err := cfg.V1ReloadConfig(&config.ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmysql"),
		Section: config.DATADB_JSN,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}

	err = db.Reload()
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

func TestDataDBReload3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	cfg.DataDbCfg().DataDbType = ""
	shdChan := utils.NewSyncedChan()
	shdWg := new(sync.WaitGroup)
	chS := engine.NewCacheS(cfg, nil, nil)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	close(chS.GetPrecacheChannel(utils.CacheAttributeProfiles))
	close(chS.GetPrecacheChannel(utils.CacheAttributeFilterIndexes))
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg)
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	srvMngr.AddServices(NewAttributeService(cfg, db,
		chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), anz, srvDep),
		NewLoaderService(cfg, db, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz, srvDep), db)
	if err := srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	cfg.AttributeSCfg().Enabled = true
	err := db.Start()
	if err == nil {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", "unsupported db_type <>", err)
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
func TestDataDBReload4(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	cfg.DataDbCfg().DataDbType = ""
	shdChan := utils.NewSyncedChan()
	shdWg := new(sync.WaitGroup)
	chS := engine.NewCacheS(cfg, nil, nil)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	close(chS.GetPrecacheChannel(utils.CacheAttributeProfiles))
	close(chS.GetPrecacheChannel(utils.CacheAttributeFilterIndexes))
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg)
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	srvMngr.AddServices(NewAttributeService(cfg, db,
		chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), anz, srvDep),
		NewLoaderService(cfg, db, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz, srvDep), db)
	if err := srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	cfg.SessionSCfg().Enabled = true
	err := db.Start()
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	cfg.SessionSCfg().Enabled = false
	cfg.GetReloadChan(config.DATADB_JSN) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	shdChan.CloseOnce()
	time.Sleep(10 * time.Millisecond)
}

func TestDataDBReload5(t *testing.T) {
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
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg)
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	srvMngr.AddServices(NewAttributeService(cfg, db,
		chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), anz, srvDep),
		NewLoaderService(cfg, db, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz, srvDep), db)
	if err := srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	cfg.AttributeSCfg().Enabled = true
	if err := cfg.V1ReloadConfig(&config.ReloadArgs{
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

	err := db.Reload()
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	if err := cfg.V1ReloadConfig(&config.ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmysql"),
		Section: config.DATADB_JSN,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	err = db.Reload()
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}

	cfg.DataDbCfg().DataDbType = "bad_type"
	err = db.Reload()
	if err == nil {
		t.Errorf("\nExpecting <unsupported db_type <bad_type>>,\n Received <%+v>", err)
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
func TestDataDBReload6(t *testing.T) {
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
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg)
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	srvMngr.AddServices(NewAttributeService(cfg, db,
		chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), anz, srvDep),
		NewLoaderService(cfg, db, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz, srvDep), db)
	if err := srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	cfg.AttributeSCfg().Enabled = true
	if err := cfg.V1ReloadConfig(&config.ReloadArgs{
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

	err := db.Reload()
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	if err := cfg.V1ReloadConfig(&config.ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
		Section: config.DATADB_JSN,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond)
	err = db.Reload()
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	db.cfg.DataDbCfg().DataDbType = utils.Mongo
	db.cfg.DataDbCfg().Opts = map[string]interface{}{
		utils.QueryTimeoutCfg: false,
	}
	err = db.Reload()
	if err == nil {
		t.Errorf("\nExpecting <cannot convert field: false to time.Duration>,\n Received <%+v>", err)
	}

	shdChan.CloseOnce()
	time.Sleep(10 * time.Millisecond)
}

/*
func TestDataDBReload7(t *testing.T) {
	dataDir := flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	var versionsConfigDIR string
	cfg, err := config.NewCGRConfigFromPath(path.Join(*dataDir, "conf", "samples", versionsConfigDIR))
	if err != nil {
		t.Fatal(err)
	}
	dbConn, err := engine.NewDataDBConn(cfg.DataDbCfg().DataDbType,
		cfg.DataDbCfg().DataDbHost, cfg.DataDbCfg().DataDbPort,
		cfg.DataDbCfg().DataDbName, cfg.DataDbCfg().DataDbUser,
		cfg.DataDbCfg().DataDbPass, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	dm3 := engine.NewDataManager(dbConn, cfg.CacheCfg(), nil)
	var test string
	var currentVersion engine.Versions
	var testVersion engine.Versions
	dataDbVersions := engine.CurrentDataDBVersions()
	storDbVersions := engine.CurrentStorDBVersions()

	allVersions := make(engine.Versions)
	for k, v := range dataDbVersions {
		allVersions[k] = v
	}
	for k, v := range storDbVersions {
		allVersions[k] = v
	}

	storType := dm3.DataDB().GetStorageType()
	switch storType {
	case utils.INTERNAL:
		currentVersion = allVersions
		testVersion = allVersions
		testVersion[utils.Accounts] = 1
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -exec=*accounts>"
	case utils.Mongo, utils.Redis:
		currentVersion = dataDbVersions
		testVersion = dataDbVersions
		testVersion[utils.Accounts] = 1

		test = "Migration needed: please backup cgr data and run : <cgr-migrator -exec=*accounts>"
	}

	//dataDB
	if _, rcvErr := dm3.DataDB().GetVersions(""); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := engine.CheckVersions(dm3.DataDB()); err != nil {
		t.Error(err)
	}
	if rcv, err := dm3.DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if len(currentVersion) != len(rcv) {
		t.Errorf("Expecting: %v, received: %v", currentVersion, rcv)
	}
	if err = dm3.DataDB().RemoveVersions(currentVersion); err != nil {
		t.Error(err)
	}
	if _, rcvErr := dm3.DataDB().GetVersions(""); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := dm3.DataDB().SetVersions(testVersion, false); err != nil {
		t.Error(err)
	}
	if err := engine.CheckVersions(dm3.DataDB()); err.Error() != test {
		t.Error(err)
	}
	if err = dm3.DataDB().RemoveVersions(testVersion); err != nil {
		t.Error(err)
	}
	var storageDb engine.Storage
	storType = storageDb.GetStorageType()
	switch storType {
	case utils.INTERNAL:
		currentVersion = allVersions
		testVersion = allVersions
		testVersion[utils.Accounts] = 1
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -exec=*accounts>"
	case utils.Mongo, utils.Postgres, utils.MySQL:
		currentVersion = storDbVersions
		testVersion = allVersions
		testVersion[utils.CostDetails] = 1
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -exec=*cost_details>"
	}
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
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg)
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	srvMngr.AddServices(NewAttributeService(cfg, db,
		chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), anz, srvDep),
		NewLoaderService(cfg, db, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz, srvDep), db)
	if err := srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	cfg.AttributeSCfg().Enabled = true
	if err := cfg.V1ReloadConfig(&config.ReloadArgs{
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

	err = db.Reload()
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	if err := cfg.V1ReloadConfig(&config.ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
		Section: config.DATADB_JSN,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}

	time.Sleep(10 * time.Millisecond)
	err = db.Reload()
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}

	cfg.DataDbCfg().DataDbType = utils.MySQL

	shdChan.CloseOnce()
	time.Sleep(10 * time.Millisecond)
}
*/
