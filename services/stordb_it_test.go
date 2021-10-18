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
	"sync"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestStorDBReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	shdWg := new(sync.WaitGroup)
	chS := engine.NewCacheS(cfg, nil, nil)
	cfg.ChargerSCfg().Enabled = true
	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg, nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	cfg.StorDbCfg().Password = "CGRateS.org"
	stordb := NewStorDBService(cfg, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	chrS := NewChargerService(cfg, db, chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz, srvDep)
	schS := NewSchedulerService(cfg, db, chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz, srvDep)
	ralS := NewRalService(cfg, chS, server,
		make(chan rpcclient.ClientConnector, 1),
		make(chan rpcclient.ClientConnector, 1),
		shdChan, nil, anz, srvDep, filterSChan)
	cdrsRPC := make(chan rpcclient.ClientConnector, 1)
	cdrS := NewCDRServer(cfg, db, stordb, filterSChan, server,
		cdrsRPC, nil, anz, srvDep)
	srvMngr.AddServices(cdrS, ralS, schS, chrS,
		NewLoaderService(cfg, db, filterSChan, server,
			make(chan rpcclient.ClientConnector, 1), nil, anz, srvDep), db, stordb)
	if err := engine.InitStorDb(cfg); err != nil {
		t.Fatal(err)
	}

	if err := srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	if cdrS.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	if stordb.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	cfg.RalsCfg().Enabled = true
	var reply string
	if err := cfg.V1ReloadConfig(&config.ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
		Section: config.CDRS_JSN,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}

	select {
	case d := <-cdrsRPC:
		cdrsRPC <- d
	case <-time.After(time.Second):
		t.Fatal("It took to long to reload the cache")
	}
	if !cdrS.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if !stordb.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	time.Sleep(10 * time.Millisecond)
	if err := stordb.Reload(); err != nil {
		t.Fatalf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	time.Sleep(10 * time.Millisecond)
	cfg.StorDbCfg().Password = ""
	if err := cfg.V1ReloadConfig(&config.ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
		Section: config.STORDB_JSN,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}

	if err := cdrS.Reload(); err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}

	if err := stordb.Reload(); err != nil {
		t.Fatalf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	cfg.StorDbCfg().Type = utils.Internal
	if err := stordb.Reload(); err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	err := stordb.Start()
	if err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	if err := stordb.Reload(); err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	if err := db.Start(); err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	if err := cdrS.Start(); err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}

	if err := cdrS.Reload(); err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}

	cfg.CdrsCfg().Enabled = false
	cfg.GetReloadChan(config.CDRS_JSN) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if cdrS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	shdChan.CloseOnce()
	time.Sleep(10 * time.Millisecond)
}

func TestStorDBReloadVersion1(t *testing.T) {
	cfg, err := config.NewCGRConfigFromPath(path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"))
	if err != nil {
		t.Fatal(err)
	}

	storageDb, err := engine.NewStorDBConn(cfg.StorDbCfg().Type,
		cfg.StorDbCfg().Host, cfg.StorDbCfg().Port,
		cfg.StorDbCfg().Name, cfg.StorDbCfg().User,
		cfg.StorDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.StorDbCfg().StringIndexedFields, cfg.StorDbCfg().PrefixIndexedFields,
		cfg.StorDbCfg().Opts, cfg.StorDbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		storageDb.Flush("")
		storageDb.Close()
	}()

	err = storageDb.SetVersions(engine.Versions{
		utils.CostDetails:   2,
		utils.SessionSCosts: 3,
		//old version for CDRs
		utils.CDRs:               1,
		utils.TpRatingPlans:      1,
		utils.TpFilters:          1,
		utils.TpDestinationRates: 1,
		utils.TpActionTriggers:   1,
		utils.TpAccountActionsV:  1,
		utils.TpActionPlans:      1,
		utils.TpActions:          1,
		utils.TpThresholds:       1,
		utils.TpRoutes:           1,
		utils.TpStats:            1,
		utils.TpSharedGroups:     1,
		utils.TpRatingProfiles:   1,
		utils.TpResources:        1,
		utils.TpRates:            1,
		utils.TpTiming:           1,
		utils.TpResource:         1,
		utils.TpDestinations:     1,
		utils.TpRatingPlan:       1,
		utils.TpRatingProfile:    1,
		utils.TpChargers:         1,
		utils.TpDispatchers:      1,
	}, true)

	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	cfg.ChargerSCfg().Enabled = true
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	stordb := NewStorDBService(cfg, srvDep)
	stordb.oldDBCfg = cfg.StorDbCfg().Clone()
	if err := engine.InitStorDb(cfg); err != nil {
		t.Fatal(err)
	}

	err = stordb.Start()
	if err != nil {
		t.Fatal(err)
	}
	stordb.db = nil
	err = stordb.Reload()
	if err == nil || err.Error() != "can't conver StorDB of type mongo to MongoStorage" {
		t.Fatal(err)
	}

	cfg.CdrsCfg().Enabled = false
	err = stordb.Reload()
	if err == nil || err.Error() != "can't conver StorDB of type mongo to MongoStorage" {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	shdChan.CloseOnce()
	time.Sleep(10 * time.Millisecond)
}

func TestStorDBReloadVersion2(t *testing.T) {
	cfg, err := config.NewCGRConfigFromPath(path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmysql"))
	if err != nil {
		t.Fatal(err)
	}

	storageDb, err := engine.NewStorDBConn(cfg.StorDbCfg().Type,
		cfg.StorDbCfg().Host, cfg.StorDbCfg().Port,
		cfg.StorDbCfg().Name, cfg.StorDbCfg().User,
		cfg.StorDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.StorDbCfg().StringIndexedFields, cfg.StorDbCfg().PrefixIndexedFields,
		cfg.StorDbCfg().Opts, cfg.StorDbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		storageDb.Flush("")
		storageDb.Close()
	}()

	err = storageDb.SetVersions(engine.Versions{
		utils.CostDetails:   2,
		utils.SessionSCosts: 3,
		//old version for CDRs
		utils.CDRs:               1,
		utils.TpRatingPlans:      1,
		utils.TpFilters:          1,
		utils.TpDestinationRates: 1,
		utils.TpActionTriggers:   1,
		utils.TpAccountActionsV:  1,
		utils.TpActionPlans:      1,
		utils.TpActions:          1,
		utils.TpThresholds:       1,
		utils.TpRoutes:           1,
		utils.TpStats:            1,
		utils.TpSharedGroups:     1,
		utils.TpRatingProfiles:   1,
		utils.TpResources:        1,
		utils.TpRates:            1,
		utils.TpTiming:           1,
		utils.TpResource:         1,
		utils.TpDestinations:     1,
		utils.TpRatingPlan:       1,
		utils.TpRatingProfile:    1,
		utils.TpChargers:         1,
		utils.TpDispatchers:      1,
	}, true)

	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	cfg.ChargerSCfg().Enabled = true
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cfg.StorDbCfg().Password = "CGRateS.org"
	stordb := NewStorDBService(cfg, srvDep)
	stordb.oldDBCfg = cfg.StorDbCfg().Clone()
	if err := engine.InitStorDb(cfg); err != nil {
		t.Fatal(err)
	}
	err = stordb.Start()
	if err != nil {
		t.Fatal(err)
	}
	stordb.db = nil
	err = stordb.Reload()
	if err == nil || err.Error() != "can't conver StorDB of type mysql to SQLStorage" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "can't convert StorDB of type mysql to SQLStorage", err)
	}
	cfg.CdrsCfg().Enabled = false
	err = stordb.Reload()
	if err == nil || err.Error() != "can't conver StorDB of type mysql to SQLStorage" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "can't convert StorDB of type mysql to SQLStorage", err)

		time.Sleep(10 * time.Millisecond)
		shdChan.CloseOnce()
		time.Sleep(10 * time.Millisecond)
	}
}
func TestStorDBReloadVersion3(t *testing.T) {
	cfg, err := config.NewCGRConfigFromPath(path.Join("/usr", "share", "cgrates", "conf", "samples", "tutinternal"))
	if err != nil {
		t.Fatal(err)
	}
	storageDb, err := engine.NewStorDBConn(cfg.StorDbCfg().Type,
		cfg.StorDbCfg().Host, cfg.StorDbCfg().Port,
		cfg.StorDbCfg().Name, cfg.StorDbCfg().User,
		cfg.StorDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.StorDbCfg().StringIndexedFields, cfg.StorDbCfg().PrefixIndexedFields,
		cfg.StorDbCfg().Opts, cfg.StorDbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		storageDb.Flush("")
		storageDb.Close()
	}()
	err = storageDb.SetVersions(engine.Versions{
		utils.CostDetails:   2,
		utils.SessionSCosts: 3,
		//old version for CDRs
		utils.CDRs:               1,
		utils.TpRatingPlans:      1,
		utils.TpFilters:          1,
		utils.TpDestinationRates: 1,
		utils.TpActionTriggers:   1,
		utils.TpAccountActionsV:  1,
		utils.TpActionPlans:      1,
		utils.TpActions:          1,
		utils.TpThresholds:       1,
		utils.TpRoutes:           1,
		utils.TpStats:            1,
		utils.TpSharedGroups:     1,
		utils.TpRatingProfiles:   1,
		utils.TpResources:        1,
		utils.TpRates:            1,
		utils.TpTiming:           1,
		utils.TpResource:         1,
		utils.TpDestinations:     1,
		utils.TpRatingPlan:       1,
		utils.TpRatingProfile:    1,
		utils.TpChargers:         1,
		utils.TpDispatchers:      1,
	}, true)

	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	// shdChan := utils.NewSyncedChan()
	cfg.ChargerSCfg().Enabled = true
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cfg.StorDbCfg().Password = "CGRateS.org"
	stordb := NewStorDBService(cfg, srvDep)
	stordb.oldDBCfg = cfg.StorDbCfg().Clone()
	stordb.db = nil
	err = stordb.Reload()
	if err == nil || err.Error() != "can't conver StorDB of type internal to InternalDB" {
		t.Fatal(err)
	}
	/* the internal now uses its own cache
	err = stordb.Start()
	if err == nil {
		t.Fatal(err)
	}
	cfg.CdrsCfg().Enabled = false
	err = stordb.Reload()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	shdChan.CloseOnce()
	time.Sleep(10 * time.Millisecond)
	*/
}

func TestStorDBReloadNewStorDBConnError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	cfg.ChargerSCfg().Enabled = true
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cfg.StorDbCfg().Password = "CGRateS.org"
	stordb := NewStorDBService(cfg, srvDep)
	stordb.oldDBCfg = &config.StorDbCfg{
		Type:     utils.Internal,
		Host:     "test_host",
		Port:     "test_port",
		Name:     "test_name",
		User:     "test_user",
		Password: "test_pass",
	}
	cfg.StorDbCfg().Type = "badType"
	err := stordb.Reload()
	if err == nil || err.Error() != "unknown db 'badType' valid options are [mysql, mongo, postgres, internal]" {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", "unknown db 'badType' valid options are [mysql, mongo, postgres, internal]", err)
	}
	shdChan.CloseOnce()
}

func TestStorDBReloadCanCastError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	cfg.ChargerSCfg().Enabled = true
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cfg.StorDbCfg().Password = "CGRateS.org"
	cfg.StorDbCfg().Type = utils.Mongo
	stordb := NewStorDBService(cfg, srvDep)
	stordb.cfg.StorDbCfg().Opts = map[string]interface{}{
		utils.MongoQueryTimeoutCfg: false,
	}
	stordb.db = &engine.MongoStorage{}
	stordb.oldDBCfg = cfg.StorDbCfg().Clone()
	err := stordb.Reload()
	if err == nil || err.Error() != "cannot convert field: false to time.Duration" {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", "cannot convert field: false to time.Duration", err)
	}
	shdChan.CloseOnce()
}

func TestStorDBReloadIfaceAsTIntMaxOpenConnsCfgError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	cfg.ChargerSCfg().Enabled = true
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cfg.StorDbCfg().Password = "CGRateS.org"
	cfg.StorDbCfg().Type = utils.MySQL
	stordb := NewStorDBService(cfg, srvDep)
	stordb.cfg.StorDbCfg().Opts = map[string]interface{}{
		utils.SQLMaxOpenConnsCfg: false,
	}
	stordb.db = &engine.SQLStorage{}
	stordb.oldDBCfg = cfg.StorDbCfg().Clone()
	err := stordb.Reload()
	expected := "cannot convert field<bool>: false to int"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, err)
	}
	shdChan.CloseOnce()
}

func TestStorDBReloadIfaceAsTIntConnMaxLifetimeCfgError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	cfg.ChargerSCfg().Enabled = true
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cfg.StorDbCfg().Password = "CGRateS.org"
	cfg.StorDbCfg().Type = utils.MySQL
	stordb := NewStorDBService(cfg, srvDep)
	stordb.cfg.StorDbCfg().Opts = map[string]interface{}{
		utils.SQLMaxOpenConnsCfg:    1,
		utils.SQLMaxIdleConnsCfg:    1,
		utils.SQLConnMaxLifetimeCfg: false,
	}
	stordb.db = &engine.SQLStorage{}
	stordb.oldDBCfg = cfg.StorDbCfg().Clone()
	err := stordb.Reload()
	expected := "cannot convert field<bool>: false to int"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, err)
	}
	shdChan.CloseOnce()
}

func TestStorDBReloadIfaceAsTIntMaxIdleConnsCfgError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	cfg.ChargerSCfg().Enabled = true
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cfg.StorDbCfg().Password = "CGRateS.org"
	cfg.StorDbCfg().Type = utils.MySQL
	stordb := NewStorDBService(cfg, srvDep)
	stordb.cfg.StorDbCfg().Opts = map[string]interface{}{
		utils.SQLMaxOpenConnsCfg:    1,
		utils.SQLMaxIdleConnsCfg:    false,
		utils.SQLConnMaxLifetimeCfg: 1,
	}
	stordb.db = &engine.SQLStorage{}
	stordb.oldDBCfg = cfg.StorDbCfg().Clone()
	err := stordb.Reload()
	expected := "cannot convert field<bool>: false to int"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, err)
	}
	shdChan.CloseOnce()
}

func TestStorDBReloadStartDBError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	cfg.ChargerSCfg().Enabled = true
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cfg.StorDbCfg().Password = "CGRateS.org"
	stordb := NewStorDBService(cfg, srvDep)
	cfg.StorDbCfg().Type = "badType"
	err := stordb.Start()
	expected := "unknown db 'badType' valid options are [mysql, mongo, postgres, internal]"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, err)
	}
	shdChan.CloseOnce()
}
