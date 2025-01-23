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

// NewStorDBService returns the StorDB Service
func NewStorDBService(cfg *config.CGRConfig, setVersions bool) *StorDBService {
	return &StorDBService{
		cfg:         cfg,
		setVersions: setVersions,
		stateDeps:   NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// StorDBService implements Service interface
type StorDBService struct {
	mu          sync.RWMutex
	cfg         *config.CGRConfig
	oldDBCfg    *config.StorDbCfg
	db          engine.StorDB
	setVersions bool
	stateDeps   *StateDependencies // channel subscriptions for state changes
}

// Start should handle the service start
func (db *StorDBService) Start(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.oldDBCfg = db.cfg.StorDbCfg().Clone()
	dbConn, err := engine.NewStorDBConn(db.cfg.StorDbCfg().Type, db.cfg.StorDbCfg().Host,
		db.cfg.StorDbCfg().Port, db.cfg.StorDbCfg().Name, db.cfg.StorDbCfg().User,
		db.cfg.StorDbCfg().Password, db.cfg.GeneralCfg().DBDataEncoding,
		db.cfg.StorDbCfg().StringIndexedFields, db.cfg.StorDbCfg().PrefixIndexedFields,
		db.cfg.StorDbCfg().Opts, db.cfg.StorDbCfg().Items)
	if err != nil { // Cannot configure getter database, show stopper
		utils.Logger.Crit(fmt.Sprintf("Could not configure storDB: %s exiting!", err))
		return
	}
	db.db = dbConn

	if db.setVersions {
		err = engine.OverwriteDBVersions(dbConn)
	} else {
		err = engine.CheckVersions(db.db)
	}
	if err != nil {
		return err
	}
	return
}

// Reload handles the change of config
func (db *StorDBService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.needsConnectionReload() {
		var d engine.StorDB
		if d, err = engine.NewStorDBConn(db.cfg.StorDbCfg().Type, db.cfg.StorDbCfg().Host,
			db.cfg.StorDbCfg().Port, db.cfg.StorDbCfg().Name, db.cfg.StorDbCfg().User,
			db.cfg.StorDbCfg().Password, db.cfg.GeneralCfg().DBDataEncoding,
			db.cfg.StorDbCfg().StringIndexedFields, db.cfg.StorDbCfg().PrefixIndexedFields,
			db.cfg.StorDbCfg().Opts, db.cfg.StorDbCfg().Items); err != nil {
			return
		}
		db.db.Close()
		db.db = d
		db.oldDBCfg = db.cfg.StorDbCfg().Clone()
		return
	}
	if db.cfg.StorDbCfg().Type == utils.MetaMongo {
		mgo, canCast := db.db.(*engine.MongoStorage)
		if !canCast {
			return fmt.Errorf("can't conver StorDB of type %s to MongoStorage",
				db.cfg.StorDbCfg().Type)
		}
		mgo.SetTTL(db.cfg.StorDbCfg().Opts.MongoQueryTimeout)
	} else if db.cfg.StorDbCfg().Type == utils.MetaPostgres ||
		db.cfg.StorDbCfg().Type == utils.MetaMySQL {
		msql, canCast := db.db.(*engine.SQLStorage)
		if !canCast {
			return fmt.Errorf("can't conver StorDB of type %s to SQLStorage",
				db.cfg.StorDbCfg().Type)
		}
		msql.DB.SetMaxOpenConns(db.cfg.StorDbCfg().Opts.SQLMaxOpenConns)
		msql.DB.SetMaxIdleConns(db.cfg.StorDbCfg().Opts.SQLMaxIdleConns)
		msql.DB.SetConnMaxLifetime(db.cfg.StorDbCfg().Opts.SQLConnMaxLifetime)
	} else if db.cfg.StorDbCfg().Type == utils.MetaInternal {
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
func (db *StorDBService) Shutdown(_ *servmanager.ServiceRegistry) (_ error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.db.Close()
	db.db = nil
	return
}

// isRunning returns if the service is running (not thread safe)
func (db *StorDBService) isRunning() bool {
	return db.db != nil
}

// ServiceName returns the service name
func (db *StorDBService) ServiceName() string {
	return utils.StorDB
}

// ShouldRun returns if the service should be running
func (db *StorDBService) ShouldRun() bool {
	return db.cfg.CdrsCfg().Enabled || db.cfg.AdminSCfg().Enabled
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
	return db.cfg.StorDbCfg().Type == utils.MetaPostgres &&
		(db.oldDBCfg.Opts.PgSSLMode != db.cfg.StorDbCfg().Opts.PgSSLMode ||
			db.oldDBCfg.Opts.PgSSLCert != db.cfg.StorDbCfg().Opts.PgSSLCert ||
			db.oldDBCfg.Opts.PgSSLKey != db.cfg.StorDbCfg().Opts.PgSSLKey ||
			db.oldDBCfg.Opts.PgSSLPassword != db.cfg.StorDbCfg().Opts.PgSSLPassword ||
			db.oldDBCfg.Opts.PgSSLCertMode != db.cfg.StorDbCfg().Opts.PgSSLCertMode ||
			db.oldDBCfg.Opts.PgSSLRootCert != db.cfg.StorDbCfg().Opts.PgSSLRootCert)
}

// DB returns the db connection object.
func (db *StorDBService) DB() engine.StorDB {
	return db.db
}

// StateChan returns signaling channel of specific state
func (db *StorDBService) StateChan(stateID string) chan struct{} {
	return db.stateDeps.StateChan(stateID)
}
