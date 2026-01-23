/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package services

import (
	"fmt"
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewDataDBService returns the DataDB Service
func NewDataDBService(cfg *config.CGRConfig, connMgr *engine.ConnManager, setVersions bool,
	srvDep map[string]*sync.WaitGroup) *DataDBService {
	return &DataDBService{
		cfg:         cfg,
		dbchan:      make(chan *engine.DataManager, 1),
		connMgr:     connMgr,
		setVersions: setVersions,
		srvDep:      srvDep,
	}
}

// DataDBService implements Service interface
type DataDBService struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	oldDBCfg *config.DataDbCfg
	connMgr  *engine.ConnManager

	dm          *engine.DataManager
	dbchan      chan *engine.DataManager
	setVersions bool

	srvDep map[string]*sync.WaitGroup
}

// Start handles the service start.
func (db *DataDBService) Start() (err error) {
	if db.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	db.Lock()
	defer db.Unlock()
	db.oldDBCfg = db.cfg.DataDbCfg().Clone()
	dbConn, err := engine.NewDataDBConn(db.cfg.DataDbCfg().Type,
		db.cfg.DataDbCfg().Host, db.cfg.DataDbCfg().Port,
		db.cfg.DataDbCfg().Name, db.cfg.DataDbCfg().User,
		db.cfg.DataDbCfg().Password, db.cfg.GeneralCfg().DBDataEncoding,
		db.cfg.DataDbCfg().Opts, db.cfg.DataDbCfg().Items)
	if db.mandatoryDB() && err != nil { // cannot configure mandatory database
		utils.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
		return
	} else if db.cfg.SessionSCfg().Enabled && err != nil {
		utils.Logger.Warning(fmt.Sprintf("Could not configure dataDb: %s. Some SessionS APIs will not work", err))
		err = nil // reset the error only if SessionS is enabled
		return
	}
	db.dm = engine.NewDataManager(dbConn, db.cfg.CacheCfg(), db.connMgr)
	engine.SetDataStorage(db.dm)

	if db.setVersions {
		err = engine.OverwriteDBVersions(dbConn)
	} else {
		err = engine.CheckVersions(db.dm.DataDB())
	}
	if err != nil {
		return err
	}

	db.dbchan <- db.dm
	return
}

// Reload handles the change of config
func (db *DataDBService) Reload() (err error) {
	db.Lock()
	defer db.Unlock()
	if db.needsConnectionReload() {
		if err = db.dm.Reconnect(db.cfg.GeneralCfg().DBDataEncoding,
			db.cfg.DataDbCfg(), db.cfg.DataDbCfg().Items); err != nil {
			return
		}
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
func (db *DataDBService) Shutdown() (err error) {
	db.srvDep[utils.DataDB].Wait()
	db.Lock()
	db.dm.Close()
	db.dm = nil
	db.Unlock()
	return
}

// IsRunning returns if the service is running
func (db *DataDBService) IsRunning() bool {
	db.RLock()
	defer db.RUnlock()
	return db.dm != nil && db.dm.DataDB() != nil
}

// ServiceName returns the service name
func (db *DataDBService) ServiceName() string {
	return utils.DataDB
}

// ShouldRun returns if the service should be running
func (db *DataDBService) ShouldRun() bool {
	return db.mandatoryDB() || db.cfg.SessionSCfg().Enabled
}

// mandatoryDB returns if the current configuration needs the DB
func (db *DataDBService) mandatoryDB() bool {
	return db.cfg.RalsCfg().Enabled || db.cfg.SchedulerCfg().Enabled || db.cfg.ChargerSCfg().Enabled ||
		db.cfg.AttributeSCfg().Enabled || db.cfg.ResourceSCfg().Enabled || db.cfg.StatSCfg().Enabled ||
		db.cfg.ThresholdSCfg().Enabled || db.cfg.RouteSCfg().Enabled || db.cfg.DispatcherSCfg().Enabled ||
		db.cfg.ApierCfg().Enabled || db.cfg.AnalyzerSCfg().Enabled || db.cfg.ERsCfg().Enabled
}

// GetDM returns the DataManager
func (db *DataDBService) GetDM() *engine.DataManager {
	db.RLock()
	defer db.RUnlock()
	return db.dm
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
			db.oldDBCfg.Opts.RedisReadTimeout != db.cfg.DataDbCfg().Opts.RedisReadTimeout ||
			db.oldDBCfg.Opts.RedisWriteTimeout != db.cfg.DataDbCfg().Opts.RedisWriteTimeout ||
			db.oldDBCfg.Opts.RedisConnectTimeout != db.cfg.DataDbCfg().Opts.RedisConnectTimeout ||
			db.oldDBCfg.Opts.RedisPoolPipelineWindow != db.cfg.DataDbCfg().Opts.RedisPoolPipelineWindow ||
			db.oldDBCfg.Opts.RedisPoolPipelineLimit != db.cfg.DataDbCfg().Opts.RedisPoolPipelineLimit)
}

// GetDMChan returns the DataManager chanel
func (db *DataDBService) GetDMChan() chan *engine.DataManager {
	db.RLock()
	defer db.RUnlock()
	return db.dbchan
}
