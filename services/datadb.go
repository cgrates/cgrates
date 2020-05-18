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
	"github.com/cgrates/cgrates/utils"
)

// NewDataDBService returns the DataDB Service
func NewDataDBService(cfg *config.CGRConfig, connMgr *engine.ConnManager) *DataDBService {
	return &DataDBService{
		cfg:     cfg,
		dbchan:  make(chan *engine.DataManager, 1),
		connMgr: connMgr,
	}
}

// DataDBService implements Service interface
type DataDBService struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	oldDBCfg *config.DataDbCfg
	connMgr  *engine.ConnManager

	db     *engine.DataManager
	dbchan chan *engine.DataManager
}

// Start should handle the sercive start
func (db *DataDBService) Start() (err error) {
	if db.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	db.Lock()
	defer db.Unlock()
	db.oldDBCfg = db.cfg.DataDbCfg().Clone()
	d, err := engine.NewDataDBConn(db.cfg.DataDbCfg().DataDbType,
		db.cfg.DataDbCfg().DataDbHost, db.cfg.DataDbCfg().DataDbPort,
		db.cfg.DataDbCfg().DataDbName, db.cfg.DataDbCfg().DataDbUser,
		db.cfg.DataDbCfg().DataDbPass, db.cfg.GeneralCfg().DBDataEncoding,
		db.cfg.DataDbCfg().DataDbSentinelName, db.cfg.DataDbCfg().Items)
	if db.mandatoryDB() && err != nil { // Cannot configure getter database, show stopper
		utils.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
		return
	} else if db.cfg.SessionSCfg().Enabled && err != nil {
		utils.Logger.Warning(fmt.Sprintf("Could not configure dataDb: %s. Some SessionS APIs will not work", err))
		err = nil // reset the error in case of only SessionS active
		return
	}

	db.db = engine.NewDataManager(d, db.cfg.CacheCfg(), db.connMgr)
	engine.SetDataStorage(db.db)
	if err = engine.CheckVersions(db.db.DataDB()); err != nil {
		fmt.Println(err)
		return
	}
	db.dbchan <- db.db
	return
}

// Reload handles the change of config
func (db *DataDBService) Reload() (err error) {
	db.Lock()
	defer db.Unlock()
	if db.needsConnectionReload() {
		if err = db.db.Reconnect(db.cfg.GeneralCfg().DBDataEncoding, db.cfg.DataDbCfg()); err != nil {
			return
		}
		db.oldDBCfg = db.cfg.DataDbCfg().Clone()
		return
	}
	if db.cfg.DataDbCfg().DataDbType == utils.MONGO {
		mgo, canCast := db.db.DataDB().(*engine.MongoStorage)
		if !canCast {
			return fmt.Errorf("can't conver DataDB of type %s to MongoStorage",
				db.cfg.DataDbCfg().DataDbType)
		}
		mgo.SetTTL(db.cfg.DataDbCfg().QueryTimeout)
	}
	return
}

// Shutdown stops the service
func (db *DataDBService) Shutdown() (err error) {
	db.Lock()
	db.db.DataDB().Close()
	db.db = nil
	db.Unlock()
	return
}

// IsRunning returns if the service is running
func (db *DataDBService) IsRunning() bool {
	db.RLock()
	defer db.RUnlock()
	return db != nil && db.db != nil && db.db.DataDB() != nil
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
		db.cfg.LoaderCfg().Enabled() || db.cfg.ApierCfg().Enabled
}

// GetDM returns the DataManager
func (db *DataDBService) GetDM() *engine.DataManager {
	db.RLock()
	defer db.RUnlock()
	return db.db
}

// needsConnectionReload returns if the DB connection needs to reloaded
func (db *DataDBService) needsConnectionReload() bool {
	if db.oldDBCfg.DataDbType != db.cfg.DataDbCfg().DataDbType ||
		db.oldDBCfg.DataDbHost != db.cfg.DataDbCfg().DataDbHost ||
		db.oldDBCfg.DataDbName != db.cfg.DataDbCfg().DataDbName ||
		db.oldDBCfg.DataDbPort != db.cfg.DataDbCfg().DataDbPort ||
		db.oldDBCfg.DataDbUser != db.cfg.DataDbCfg().DataDbUser ||
		db.oldDBCfg.DataDbPass != db.cfg.DataDbCfg().DataDbPass {
		return true
	}
	if db.oldDBCfg.DataDbType == utils.REDIS {
		return db.oldDBCfg.DataDbSentinelName != db.cfg.DataDbCfg().DataDbSentinelName
	}
	return false
}

// GetDMChan returns the DataManager chanel
func (db *DataDBService) GetDMChan() chan *engine.DataManager {
	db.RLock()
	defer db.RUnlock()
	return db.dbchan
}
