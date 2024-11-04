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
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func TestDataDBReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	shdWg := new(sync.WaitGroup)
	chS := engine.NewCacheS(cfg, nil, nil, nil)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	close(chS.GetPrecacheChannel(utils.CacheAttributeProfiles))
	close(chS.GetPrecacheChannel(utils.CacheAttributeFilterIndexes))
	chSCh := make(chan *engine.CacheS, 1)
	chSCh <- chS
	css := &CacheService{cacheCh: chSCh}
	server := commonlisteners.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srvMngr := servmanager.NewServiceManager(shdWg, nil, cfg)
	cM := engine.NewConnManager(cfg)
	db := NewDataDBService(cfg, cM, false, srvDep)
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
		Type: utils.MetaMongo,
		Host: "127.0.0.1",
		Port: "27017",
		Name: "10",
		User: "cgrates",
		Opts: &config.DataDBOpts{
			MongoQueryTimeout: 10 * time.Second,
			RedisClusterSync:  5 * time.Second,
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpts{
			utils.MetaAccounts:           {Limit: -1},
			utils.MetaActions:            {Limit: -1},
			utils.MetaResourceProfile:    {Limit: -1},
			utils.MetaStatQueues:         {Limit: -1},
			utils.MetaResources:          {Limit: -1},
			utils.MetaStatQueueProfiles:  {Limit: -1},
			utils.MetaThresholds:         {Limit: -1},
			utils.MetaThresholdProfiles:  {Limit: -1},
			utils.MetaFilters:            {Limit: -1},
			utils.MetaRouteProfiles:      {Limit: -1},
			utils.MetaAttributeProfiles:  {Limit: -1},
			utils.MetaDispatcherHosts:    {Limit: -1},
			utils.MetaChargerProfiles:    {Limit: -1},
			utils.MetaDispatcherProfiles: {Limit: -1},
			utils.MetaLoadIDs:            {Limit: -1},
			utils.MetaRateProfiles:       {Limit: -1},
			utils.MetaActionProfiles:     {Limit: -1},

			utils.CacheResourceFilterIndexes:       {Limit: -1},
			utils.CacheStatFilterIndexes:           {Limit: -1},
			utils.CacheThresholdFilterIndexes:      {Limit: -1},
			utils.CacheRouteFilterIndexes:          {Limit: -1},
			utils.CacheAttributeFilterIndexes:      {Limit: -1},
			utils.CacheChargerFilterIndexes:        {Limit: -1},
			utils.CacheDispatcherFilterIndexes:     {Limit: -1},
			utils.CacheRateProfilesFilterIndexes:   {Limit: -1},
			utils.CacheActionProfilesFilterIndexes: {Limit: -1},
			utils.CacheAccountsFilterIndexes:       {Limit: -1},
			utils.CacheVersions:                    {Limit: -1},
			utils.CacheReverseFilterIndexes:        {Limit: -1},
			utils.CacheRateFilterIndexes:           {Limit: -1},
		},
	}
	if !reflect.DeepEqual(oldcfg, db.oldDBCfg) {
		t.Errorf("Expected %s \n received:%s", utils.ToJSON(oldcfg), utils.ToJSON(db.oldDBCfg))
	}

	err := db.Reload(ctx, cancel)
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	// cfg.AttributeSCfg().Enabled = false
	// cfg.GetReloadChan() <- config.SectionToService[config.DataDBJSON]
	// runtime.Gosched()
	// time.Sleep(10 * time.Millisecond)
	// if db.IsRunning() {
	// 	t.Errorf("Expected service to be down")
	// }
	cancel()
	time.Sleep(10 * time.Millisecond)
}

func TestDataDBReloadBadType(t *testing.T) {
	cfg, err := config.NewCGRConfigFromPath(context.Background(), path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"))
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
		utils.Stats:      4,
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
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg)
	db := NewDataDBService(cfg, cM, false, srvDep)
	db.oldDBCfg = &config.DataDbCfg{
		Type: utils.MetaMongo,
		Host: "127.0.0.1",
		Port: "27017",
		Name: "10",
		User: "cgrates",
		Opts: &config.DataDBOpts{
			MongoQueryTimeout: 10 * time.Second,
			RedisClusterSync:  5 * time.Second,
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpts{
			utils.MetaAccounts:           {},
			utils.MetaActions:            {},
			utils.MetaCronExp:            {},
			utils.MetaResourceProfile:    {},
			utils.MetaStatQueues:         {},
			utils.MetaResources:          {},
			utils.MetaStatQueueProfiles:  {},
			utils.MetaThresholds:         {},
			utils.MetaThresholdProfiles:  {},
			utils.MetaFilters:            {},
			utils.MetaRouteProfiles:      {},
			utils.MetaAttributeProfiles:  {},
			utils.MetaDispatcherHosts:    {},
			utils.MetaChargerProfiles:    {},
			utils.MetaDispatcherProfiles: {},
			utils.MetaLoadIDs:            {},
			utils.MetaRateProfiles:       {},
			utils.MetaActionProfiles:     {},
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
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg)
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
			MongoQueryTimeout: 10 * time.Second,
			RedisClusterSync:  5 * time.Second,
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpts{
			utils.MetaAccounts:           {},
			utils.MetaActions:            {},
			utils.MetaCronExp:            {},
			utils.MetaResourceProfile:    {},
			utils.MetaStatQueues:         {},
			utils.MetaResources:          {},
			utils.MetaStatQueueProfiles:  {},
			utils.MetaThresholds:         {},
			utils.MetaThresholdProfiles:  {},
			utils.MetaFilters:            {},
			utils.MetaRouteProfiles:      {},
			utils.MetaAttributeProfiles:  {},
			utils.MetaDispatcherHosts:    {},
			utils.MetaChargerProfiles:    {},
			utils.MetaDispatcherProfiles: {},
			utils.MetaLoadIDs:            {},
			utils.MetaRateProfiles:       {},
			utils.MetaActionProfiles:     {},
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
	cfg, err := config.NewCGRConfigFromPath(context.Background(), path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"))
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
		utils.Stats:      4,
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
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg)
	db := NewDataDBService(cfg, cM, false, srvDep)
	ctx, cancel := context.WithCancel(context.TODO())
	err = db.Start(ctx, cancel)
	if err == nil || err.Error() != "Migration needed: please backup cgr data and run : <cgr-migrator -exec=*attributes>" {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", "Migration needed: please backup cgr data and run : <cgr-migrator -exec=*attributes>", err)
	}
	cancel()
	time.Sleep(10 * time.Millisecond)
}

func TestDataDBReloadCastError(t *testing.T) {
	cfg, err := config.NewCGRConfigFromPath(context.Background(), path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"))
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
		utils.Stats:      4,
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
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg)
	db := NewDataDBService(cfg, cM, false, srvDep)
	db.oldDBCfg = &config.DataDbCfg{
		Type: utils.MetaMongo,
		Host: "127.0.0.1",
		Port: "27017",
		Name: "10",
		User: "cgrates",
		Opts: &config.DataDBOpts{
			MongoQueryTimeout: 10 * time.Second,
			RedisClusterSync:  5 * time.Second,
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpts{
			utils.MetaAccounts:           {},
			utils.MetaActions:            {},
			utils.MetaCronExp:            {},
			utils.MetaResourceProfile:    {},
			utils.MetaStatQueues:         {},
			utils.MetaResources:          {},
			utils.MetaStatQueueProfiles:  {},
			utils.MetaThresholds:         {},
			utils.MetaThresholdProfiles:  {},
			utils.MetaFilters:            {},
			utils.MetaRouteProfiles:      {},
			utils.MetaAttributeProfiles:  {},
			utils.MetaDispatcherHosts:    {},
			utils.MetaChargerProfiles:    {},
			utils.MetaDispatcherProfiles: {},
			utils.MetaLoadIDs:            {},
			utils.MetaRateProfiles:       {},
			utils.MetaActionProfiles:     {},
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

func TestDataDBStartAttributeSCfgErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg)
	db := NewDataDBService(cfg, cM, false, srvDep)
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
	cM := engine.NewConnManager(cfg)
	db := NewDataDBService(cfg, cM, false, srvDep)
	cfg.GeneralCfg().DBDataEncoding = utils.JSON
	db.oldDBCfg = &config.DataDbCfg{
		Type: utils.MetaMongo,
		Host: "127.0.0.1",
		Port: "27017",
		Name: "10",
		User: "cgrates",
		Opts: &config.DataDBOpts{
			MongoQueryTimeout: 10 * time.Second,
			RedisClusterSync:  5 * time.Second,
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpts{
			utils.MetaAccounts:           {},
			utils.MetaActions:            {},
			utils.MetaCronExp:            {},
			utils.MetaResourceProfile:    {},
			utils.MetaStatQueues:         {},
			utils.MetaResources:          {},
			utils.MetaStatQueueProfiles:  {},
			utils.MetaThresholds:         {},
			utils.MetaThresholdProfiles:  {},
			utils.MetaFilters:            {},
			utils.MetaRouteProfiles:      {},
			utils.MetaAttributeProfiles:  {},
			utils.MetaDispatcherHosts:    {},
			utils.MetaChargerProfiles:    {},
			utils.MetaDispatcherProfiles: {},
			utils.MetaLoadIDs:            {},
			utils.MetaRateProfiles:       {},
			utils.MetaActionProfiles:     {},
		},
	}
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	db.dm = engine.NewDataManager(data, nil, nil)
	ctx, cancel := context.WithCancel(context.TODO())
	err := db.Reload(ctx, cancel)
	if err != nil {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}
