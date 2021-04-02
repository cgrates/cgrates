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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewStorDBService returns the StorDB Service
func NewStorDBService(cfg *config.CGRConfig,
	srvDep map[string]*sync.WaitGroup) *StorDBService {
	return &StorDBService{
		cfg:    cfg,
		srvDep: srvDep,
	}
}

// StorDBService implements Service interface
type StorDBService struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	oldDBCfg *config.StorDbCfg

	db        engine.StorDB
	syncChans []chan engine.StorDB

	srvDep map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (db *StorDBService) Start() (err error) {
	if db.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	db.Lock()
	defer db.Unlock()
	db.oldDBCfg = db.cfg.StorDbCfg().Clone()
	d, err := engine.NewStorDBConn(db.cfg.StorDbCfg().Type, db.cfg.StorDbCfg().Host,
		db.cfg.StorDbCfg().Port, db.cfg.StorDbCfg().Name, db.cfg.StorDbCfg().User,
		db.cfg.StorDbCfg().Password, db.cfg.GeneralCfg().DBDataEncoding,
		db.cfg.StorDbCfg().StringIndexedFields, db.cfg.StorDbCfg().PrefixIndexedFields,
		db.cfg.StorDbCfg().Opts)
	if err != nil { // Cannot configure getter database, show stopper
		utils.Logger.Crit(fmt.Sprintf("Could not configure storDB: %s exiting!", err))
		return
	}
	db.db = d
	engine.SetCdrStorage(db.db)
	if err = engine.CheckVersions(db.db); err != nil {
		fmt.Println(err)
		return
	}
	db.sync()
	return
}

// Reload handles the change of config
func (db *StorDBService) Reload() (err error) {
	db.Lock()
	defer db.Unlock()
	if db.needsConnectionReload() {
		var d engine.StorDB
		if d, err = engine.NewStorDBConn(db.cfg.StorDbCfg().Type, db.cfg.StorDbCfg().Host,
			db.cfg.StorDbCfg().Port, db.cfg.StorDbCfg().Name, db.cfg.StorDbCfg().User,
			db.cfg.StorDbCfg().Password, db.cfg.GeneralCfg().DBDataEncoding,
			db.cfg.StorDbCfg().StringIndexedFields, db.cfg.StorDbCfg().PrefixIndexedFields,
			db.cfg.StorDbCfg().Opts); err != nil {
			return
		}
		db.db.Close()
		db.db = d
		db.oldDBCfg = db.cfg.StorDbCfg().Clone()
		db.sync() // sync only if needed
		return
	}
	if db.cfg.StorDbCfg().Type == utils.Mongo {
		mgo, canCast := db.db.(*engine.MongoStorage)
		if !canCast {
			return fmt.Errorf("can't conver StorDB of type %s to MongoStorage",
				db.cfg.StorDbCfg().Type)
		}
		var ttl time.Duration
		if ttl, err = utils.IfaceAsDuration(db.cfg.StorDbCfg().Opts[utils.QueryTimeoutCfg]); err != nil {
			return
		}
		mgo.SetTTL(ttl)
	} else if db.cfg.StorDbCfg().Type == utils.Postgres ||
		db.cfg.StorDbCfg().Type == utils.MySQL {
		msql, canCast := db.db.(*engine.SQLStorage)
		if !canCast {
			return fmt.Errorf("can't conver StorDB of type %s to SQLStorage",
				db.cfg.StorDbCfg().Type)
		}
		var maxConn, maxIdleConn, connMaxLifetime int64
		if maxConn, err = utils.IfaceAsTInt64(db.cfg.StorDbCfg().Opts[utils.MaxOpenConnsCfg]); err != nil {
			return
		}
		if maxIdleConn, err = utils.IfaceAsTInt64(db.cfg.StorDbCfg().Opts[utils.MaxIdleConnsCfg]); err != nil {
			return
		}
		if connMaxLifetime, err = utils.IfaceAsTInt64(db.cfg.StorDbCfg().Opts[utils.ConnMaxLifetimeCfg]); err != nil {
			return
		}
		msql.DB.SetMaxOpenConns(int(maxConn))
		msql.DB.SetMaxIdleConns(int(maxIdleConn))
		msql.DB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)
	} else if db.cfg.StorDbCfg().Type == utils.Internal {
		idb, canCast := db.db.(*engine.InternalDB)
		if !canCast {
			return fmt.Errorf("can't conver StorDB of type %s to InternalDB",
				db.cfg.StorDbCfg().Type)
		}
		idb.SetStringIndexedFields(db.cfg.StorDbCfg().StringIndexedFields)
		idb.SetPrefixIndexedFields(db.cfg.StorDbCfg().PrefixIndexedFields)
	}
	return
}

// Shutdown stops the service
func (db *StorDBService) Shutdown() (err error) {
	db.Lock()
	db.db.Close()
	db.db = nil
	for _, c := range db.syncChans {
		close(c)
	}
	db.syncChans = nil
	db.Unlock()
	return
}

// IsRunning returns if the service is running
func (db *StorDBService) IsRunning() bool {
	db.RLock()
	defer db.RUnlock()
	return db.isRunning()
}

// isRunning returns if the service is running (not thread safe)
func (db *StorDBService) isRunning() bool {
	return db != nil && db.db != nil
}

// ServiceName returns the service name
func (db *StorDBService) ServiceName() string {
	return utils.StorDB
}

// ShouldRun returns if the service should be running
func (db *StorDBService) ShouldRun() bool {
	return db.cfg.CdrsCfg().Enabled || db.cfg.ApierCfg().Enabled
}

// RegisterSyncChan used by dependent subsystems to register a chanel to reload only the storDB(thread safe)
func (db *StorDBService) RegisterSyncChan(c chan engine.StorDB) {
	db.Lock()
	db.syncChans = append(db.syncChans, c)
	if db.isRunning() {
		c <- db.db
	}
	db.Unlock()
}

// sync sends the storDB over syncChansv (not thrad safe)
func (db *StorDBService) sync() {
	if db.isRunning() {
		for _, c := range db.syncChans {
			c <- db.db
		}
	}
}

// needsConnectionReload returns if the DB connection needs to reloaded
func (db *StorDBService) needsConnectionReload() bool {
	if db.oldDBCfg.Type != db.cfg.StorDbCfg().Type ||
		db.oldDBCfg.Host != db.cfg.StorDbCfg().Host ||
		db.oldDBCfg.Name != db.cfg.StorDbCfg().Name ||
		db.oldDBCfg.Port != db.cfg.StorDbCfg().Port ||
		db.oldDBCfg.User != db.cfg.StorDbCfg().User ||
		db.oldDBCfg.Password != db.cfg.StorDbCfg().Password {
		return true
	}
	return db.cfg.StorDbCfg().Type == utils.Postgres &&
		utils.IfaceAsString(db.oldDBCfg.Opts[utils.SSLModeCfg]) != utils.IfaceAsString(db.cfg.StorDbCfg().Opts[utils.SSLModeCfg])
}
