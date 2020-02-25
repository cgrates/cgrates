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
	"github.com/cgrates/rpcclient"
)

// NewStorDBService returns the StorDB Service
func NewStorDBService(cfg *config.CGRConfig) *StorDBService {
	return &StorDBService{
		cfg: cfg,
		// db:     engine.NewInternalDB([]string{}, []string{}), // to be removed
	}
}

// StorDBService implements Service interface
type StorDBService struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	oldDBCfg *config.StorDbCfg

	db        engine.StorDB
	syncChans []chan engine.StorDB

	reconnected bool
}

// Start should handle the sercive start
func (db *StorDBService) Start() (err error) {
	if db.IsRunning() {
		return fmt.Errorf("service aleady running")
	}
	db.Lock()
	defer db.Unlock()
	db.oldDBCfg = db.cfg.StorDbCfg().Clone()
	d, err := engine.NewStorDBConn(db.cfg.StorDbCfg().Type, db.cfg.StorDbCfg().Host,
		db.cfg.StorDbCfg().Port, db.cfg.StorDbCfg().Name, db.cfg.StorDbCfg().User,
		db.cfg.StorDbCfg().Password, db.cfg.GeneralCfg().DBDataEncoding,
		db.cfg.StorDbCfg().SSLMode, db.cfg.StorDbCfg().MaxOpenConns,
		db.cfg.StorDbCfg().MaxIdleConns, db.cfg.StorDbCfg().ConnMaxLifetime,
		db.cfg.StorDbCfg().StringIndexedFields, db.cfg.StorDbCfg().PrefixIndexedFields,
		db.cfg.StorDbCfg().Items)
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

// GetIntenternalChan returns the internal connection chanel
func (db *StorDBService) GetIntenternalChan() (conn chan rpcclient.ClientConnector) {
	return nil
}

// Reload handles the change of config
func (db *StorDBService) Reload() (err error) {
	db.Lock()
	defer db.Unlock()
	if db.reconnected = db.needsConnectionReload(); db.reconnected {
		var d engine.StorDB
		if d, err = engine.NewStorDBConn(db.cfg.StorDbCfg().Type, db.cfg.StorDbCfg().Host,
			db.cfg.StorDbCfg().Port, db.cfg.StorDbCfg().Name, db.cfg.StorDbCfg().User,
			db.cfg.StorDbCfg().Password, db.cfg.GeneralCfg().DBDataEncoding,
			db.cfg.StorDbCfg().SSLMode, db.cfg.StorDbCfg().MaxOpenConns,
			db.cfg.StorDbCfg().MaxIdleConns, db.cfg.StorDbCfg().ConnMaxLifetime,
			db.cfg.StorDbCfg().StringIndexedFields, db.cfg.StorDbCfg().PrefixIndexedFields,
			db.cfg.StorDbCfg().Items); err != nil {
			return
		}
		db.db.Close()
		db.db = d
		db.oldDBCfg = db.cfg.StorDbCfg().Clone()
		db.sync() // sync only if needed
		return
	}
	if db.cfg.StorDbCfg().Type == utils.MONGO {
		mgo, canCast := db.db.(*engine.MongoStorage)
		if !canCast {
			return fmt.Errorf("can't conver StorDB of type %s to MongoStorage",
				db.cfg.StorDbCfg().Type)
		}
		mgo.SetTTL(db.cfg.StorDbCfg().QueryTimeout)
	} else if db.cfg.StorDbCfg().Type == utils.POSTGRES ||
		db.cfg.StorDbCfg().Type == utils.MYSQL {
		msql, canCast := db.db.(*engine.SQLStorage)
		if !canCast {
			return fmt.Errorf("can't conver StorDB of type %s to SQLStorage",
				db.cfg.StorDbCfg().Type)
		}
		msql.Db.SetMaxOpenConns(db.cfg.StorDbCfg().MaxOpenConns)
		msql.Db.SetMaxIdleConns(db.cfg.StorDbCfg().MaxIdleConns)
		msql.Db.SetConnMaxLifetime(time.Duration(db.cfg.StorDbCfg().ConnMaxLifetime) * time.Second)
	} else if db.cfg.StorDbCfg().Type == utils.INTERNAL {
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
	db.reconnected = false
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
	return db.cfg.RalsCfg().Enabled || db.cfg.CdrsCfg().Enabled
}

// RegisterSyncChan used by dependent subsystems to register a chanel to reload only the storDB(thread safe)
func (db *StorDBService) RegisterSyncChan(c chan engine.StorDB) {
	db.Lock()
	db.syncChans = append(db.syncChans, c)
	if db.isRunning() {
		for _, c := range db.syncChans {
			c <- db.db
		}
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
	if db.cfg.StorDbCfg().Type == utils.POSTGRES &&
		db.oldDBCfg.SSLMode != db.cfg.StorDbCfg().SSLMode {
		return true
	}
	return false
}
