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
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewDataDBService returns the DataDB Service
func NewDataDBService(cfg *config.CGRConfig, setVersions bool,
	srvDep map[string]*sync.WaitGroup) *DataDBService {
	return &DataDBService{
		cfg:         cfg,
		setVersions: setVersions,
		srvDep:      srvDep,
		stateDeps:   NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// DataDBService implements Service interface
type DataDBService struct {
	mu          sync.Mutex
	cfg         *config.CGRConfig
	oldDBCfg    *config.DataDbCfg
	dm          *engine.DataManager
	setVersions bool
	srvDep      map[string]*sync.WaitGroup
	stateDeps   *StateDependencies // channel subscriptions for state changes
}

// Start handles the service start.
func (db *DataDBService) Start(_ chan struct{}, registry *servmanager.ServiceRegistry) (err error) {
	cms, err := WaitForServiceState(utils.StateServiceUP, utils.ConnManager, registry, db.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	db.oldDBCfg = db.cfg.DataDbCfg().Clone()
	dbConn, err := engine.NewDataDBConn(db.cfg.DataDbCfg().Type,
		db.cfg.DataDbCfg().Host, db.cfg.DataDbCfg().Port,
		db.cfg.DataDbCfg().Name, db.cfg.DataDbCfg().User,
		db.cfg.DataDbCfg().Password, db.cfg.GeneralCfg().DBDataEncoding,
		db.cfg.DataDbCfg().Opts, db.cfg.DataDbCfg().Items)
	if err != nil { // Cannot configure getter database, show stopper
		utils.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
		return
	}
	db.dm = engine.NewDataManager(dbConn, db.cfg.CacheCfg(), cms.(*ConnManagerService).ConnManager())

	if db.setVersions {
		err = engine.OverwriteDBVersions(dbConn)
	} else {
		err = engine.CheckVersions(db.dm.DataDB())
	}
	if err != nil {
		return err
	}

	close(db.stateDeps.StateChan(utils.StateServiceUP))
	return
}

// Reload handles the change of config
func (db *DataDBService) Reload(_ chan struct{}, _ *servmanager.ServiceRegistry) (err error) {
	if db.needsConnectionReload() {
		var d engine.DataDBDriver
		d, err = engine.NewDataDBConn(db.cfg.DataDbCfg().Type,
			db.cfg.DataDbCfg().Host, db.cfg.DataDbCfg().Port,
			db.cfg.DataDbCfg().Name, db.cfg.DataDbCfg().User,
			db.cfg.DataDbCfg().Password, db.cfg.GeneralCfg().DBDataEncoding,
			db.cfg.DataDbCfg().Opts, db.cfg.DataDbCfg().Items)
		if err != nil {
			return
		}
		db.dm.Reconnect(d)
		db.oldDBCfg = db.cfg.DataDbCfg().Clone()
		return
	}
	if db.cfg.DataDbCfg().Type == utils.MetaMongo {
		mgo, canCast := db.dm.DataDB().(*engine.MongoStorage)
		if !canCast {
			return fmt.Errorf("can't conver DataDB of type %s to MongoStorage",
				db.cfg.DataDbCfg().Type)
		}
		mgo.SetTTL(db.cfg.DataDbCfg().Opts.MongoQueryTimeout)
	}
	return
}

// Shutdown stops the service
func (db *DataDBService) Shutdown(_ *servmanager.ServiceRegistry) (_ error) {
	db.srvDep[utils.DataDB].Wait()
	db.dm.DataDB().Close()
	db.dm = nil
	close(db.StateChan(utils.StateServiceDOWN))
	return
}

// ServiceName returns the service name
func (db *DataDBService) ServiceName() string {
	return utils.DataDB
}

// ShouldRun returns if the service should be running
func (db *DataDBService) ShouldRun() bool { // db should allways run
	return true // ||db.mandatoryDB() || db.cfg.SessionSCfg().Enabled
}

// needsConnectionReload returns if the DB connection needs to reloaded
func (db *DataDBService) needsConnectionReload() bool {
	if db.oldDBCfg.Type != db.cfg.DataDbCfg().Type ||
		db.oldDBCfg.Host != db.cfg.DataDbCfg().Host ||
		db.oldDBCfg.Name != db.cfg.DataDbCfg().Name ||
		db.oldDBCfg.Port != db.cfg.DataDbCfg().Port ||
		db.oldDBCfg.User != db.cfg.DataDbCfg().User ||
		db.oldDBCfg.Password != db.cfg.DataDbCfg().Password {
		return true
	}
	if db.cfg.DataDbCfg().Type == utils.MetaInternal { // in case of internal recreate the db using the new config
		for key, itm := range db.oldDBCfg.Items {
			if db.cfg.DataDbCfg().Items[key].Limit != itm.Limit &&
				db.cfg.DataDbCfg().Items[key].StaticTTL != itm.StaticTTL &&
				db.cfg.DataDbCfg().Items[key].TTL != itm.TTL {
				return true
			}
		}
	}
	return db.oldDBCfg.Type == utils.MetaRedis &&
		(db.oldDBCfg.Opts.RedisMaxConns != db.cfg.DataDbCfg().Opts.RedisMaxConns ||
			db.oldDBCfg.Opts.RedisConnectAttempts != db.cfg.DataDbCfg().Opts.RedisConnectAttempts ||
			db.oldDBCfg.Opts.RedisSentinel != db.cfg.DataDbCfg().Opts.RedisSentinel ||
			db.oldDBCfg.Opts.RedisCluster != db.cfg.DataDbCfg().Opts.RedisCluster ||
			db.oldDBCfg.Opts.RedisClusterSync != db.cfg.DataDbCfg().Opts.RedisClusterSync ||
			db.oldDBCfg.Opts.RedisClusterOndownDelay != db.cfg.DataDbCfg().Opts.RedisClusterOndownDelay ||
			db.oldDBCfg.Opts.RedisConnectTimeout != db.cfg.DataDbCfg().Opts.RedisConnectTimeout ||
			db.oldDBCfg.Opts.RedisReadTimeout != db.cfg.DataDbCfg().Opts.RedisReadTimeout ||
			db.oldDBCfg.Opts.RedisWriteTimeout != db.cfg.DataDbCfg().Opts.RedisWriteTimeout ||
			db.oldDBCfg.Opts.RedisPoolPipelineWindow != db.cfg.DataDbCfg().Opts.RedisPoolPipelineWindow ||
			db.oldDBCfg.Opts.RedisPoolPipelineLimit != db.cfg.DataDbCfg().Opts.RedisPoolPipelineLimit)
}

// DataManager returns the DataManager object.
func (db *DataDBService) DataManager() *engine.DataManager {
	return db.dm
}

// StateChan returns signaling channel of specific state
func (db *DataDBService) StateChan(stateID string) chan struct{} {
	return db.stateDeps.StateChan(stateID)
}

// Lock implements the sync.Locker interface
func (db *DataDBService) Lock() {
	db.mu.Lock()
}

// Unlock implements the sync.Locker interface
func (db *DataDBService) Unlock() {
	db.mu.Unlock()
}
