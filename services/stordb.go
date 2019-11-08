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
		cfg:    cfg,
		dbchan: make(chan engine.StorDB, 1),
		// db:     engine.NewInternalDB([]string{}, []string{}), // to be removed
	}
}

// StorDBService implements Service interface
type StorDBService struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	oldDBCfg *config.StorDbCfg

	db     engine.StorDB
	dbchan chan engine.StorDB

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
		db.cfg.StorDbCfg().Password, db.cfg.StorDbCfg().SSLMode, db.cfg.StorDbCfg().MaxOpenConns,
		db.cfg.StorDbCfg().MaxIdleConns, db.cfg.StorDbCfg().ConnMaxLifetime,
		db.cfg.StorDbCfg().StringIndexedFields, db.cfg.StorDbCfg().PrefixIndexedFields)
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
	db.dbchan <- db.db
	return
}

// GetIntenternalChan returns the internal connection chanel
func (db *StorDBService) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
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
			db.cfg.StorDbCfg().Password, db.cfg.StorDbCfg().SSLMode, db.cfg.StorDbCfg().MaxOpenConns,
			db.cfg.StorDbCfg().MaxIdleConns, db.cfg.StorDbCfg().ConnMaxLifetime,
			db.cfg.StorDbCfg().StringIndexedFields, db.cfg.StorDbCfg().PrefixIndexedFields); err != nil {
			return
		}
		db.db.Close()
		db.db = d
		db.oldDBCfg = db.cfg.StorDbCfg().Clone()
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
	db.Unlock()
	return
}

// IsRunning returns if the service is running
func (db *StorDBService) IsRunning() bool {
	db.RLock()
	defer db.RUnlock()
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

// GetDM returns the StorDB
func (db *StorDBService) GetDM() engine.StorDB {
	db.RLock()
	defer db.RUnlock()
	return db.db
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

// GetStorDBchan returns the StorDB chanel
func (db *StorDBService) GetStorDBchan() chan engine.StorDB {
	db.RLock()
	defer db.RUnlock()
	return db.dbchan
}

// WasReconnected returns if after reload the DB was recreated
func (db *StorDBService) WasReconnected() bool {
	db.RLock()
	defer db.RUnlock()
	return db.reconnected
}
