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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewDataDBService returns the DataDB Service
func NewDataDBService(cfg *config.CGRConfig, connMgr *engine.ConnManager,
	srvDep map[string]*sync.WaitGroup) *DataDBService {
	return &DataDBService{
		cfg:     cfg,
		dbchan:  make(chan *engine.DataManager, 1),
		connMgr: connMgr,
		srvDep:  srvDep,
	}
}

// DataDBService implements Service interface
type DataDBService struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	oldDBCfg *config.DataDbCfg
	connMgr  *engine.ConnManager

	dm     *engine.DataManager
	dbchan chan *engine.DataManager
	srvDep map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (db *DataDBService) Start(*context.Context, context.CancelFunc) (err error) {
	if db.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	db.Lock()
	defer db.Unlock()
	db.oldDBCfg = db.cfg.DataDbCfg().Clone()
	var d engine.DataDBDriver
	d, err = engine.NewDataDBConn(db.cfg.DataDbCfg().Type,
		db.cfg.DataDbCfg().Host, db.cfg.DataDbCfg().Port,
		db.cfg.DataDbCfg().Name, db.cfg.DataDbCfg().User,
		db.cfg.DataDbCfg().Password, db.cfg.GeneralCfg().DBDataEncoding,
		db.cfg.DataDbCfg().Opts)
	if err != nil { // Cannot configure getter database, show stopper
		utils.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
		return
	}
	db.dm = engine.NewDataManager(d, db.cfg.CacheCfg(), db.connMgr)
	if err = engine.CheckVersions(db.dm.DataDB()); err != nil {
		fmt.Println(err)
		return
	}
	db.dbchan <- db.dm
	return
}

// Reload handles the change of config
func (db *DataDBService) Reload(*context.Context, context.CancelFunc) (err error) {
	db.Lock()
	defer db.Unlock()
	if db.needsConnectionReload() {
		var d engine.DataDBDriver
		d, err = engine.NewDataDBConn(db.cfg.DataDbCfg().Type,
			db.cfg.DataDbCfg().Host, db.cfg.DataDbCfg().Port,
			db.cfg.DataDbCfg().Name, db.cfg.DataDbCfg().User,
			db.cfg.DataDbCfg().Password, db.cfg.GeneralCfg().DBDataEncoding,
			db.cfg.DataDbCfg().Opts)
		if err != nil {
			return
		}
		db.dm.Reconnect(d)
		db.oldDBCfg = db.cfg.DataDbCfg().Clone()
		return
	}
	if db.cfg.DataDbCfg().Type == utils.Mongo {
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
func (db *DataDBService) Shutdown() (_ error) {
	db.srvDep[utils.DataDB].Wait()
	db.Lock()
	db.dm.DataDB().Close()
	db.dm = nil
	<-db.dbchan
	db.Unlock()
	return
}

// IsRunning returns if the service is running
func (db *DataDBService) IsRunning() bool {
	db.RLock()
	defer db.RUnlock()
	return db != nil && db.dm != nil && db.dm.DataDB() != nil
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
	return db.oldDBCfg.Type == utils.Redis &&
		(db.oldDBCfg.Opts.RedisSentinel != db.cfg.DataDbCfg().Opts.RedisSentinel ||
			db.oldDBCfg.Opts.RedisCluster != db.cfg.DataDbCfg().Opts.RedisCluster ||
			db.oldDBCfg.Opts.RedisClusterSync != db.cfg.DataDbCfg().Opts.RedisClusterSync ||
			db.oldDBCfg.Opts.RedisClusterOndownDelay != db.cfg.DataDbCfg().Opts.RedisClusterOndownDelay)
}

// GetDM returns the DataManager
func (db *DataDBService) WaitForDM(ctx *context.Context) (datadb *engine.DataManager, err error) {
	db.RLock()
	dbCh := db.dbchan
	db.RUnlock()
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case datadb = <-dbCh:
		dbCh <- datadb
	}
	return
}
