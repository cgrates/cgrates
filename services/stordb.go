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
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewStorDBService returns the StorDB Service
func NewStorDBService(cfg *config.CGRConfig, setVersions bool,
	srvDep map[string]*sync.WaitGroup,
	srvIndexer *servmanager.ServiceIndexer) *StorDBService {
	return &StorDBService{
		cfg:         cfg,
		setVersions: setVersions,
		srvDep:      srvDep,
		srvIndexer:  srvIndexer,
	}
}

// StorDBService implements Service interface
type StorDBService struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	oldDBCfg *config.StorDbCfg

	db          engine.StorDB
	syncChans   []chan engine.StorDB
	setVersions bool

	srvDep map[string]*sync.WaitGroup

	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the service start
func (db *StorDBService) Start(*context.Context, context.CancelFunc) (err error) {
	if db.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	db.Lock()
	defer db.Unlock()
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

	db.sync()
	return
}

// Reload handles the change of config
func (db *StorDBService) Reload(*context.Context, context.CancelFunc) (err error) {
	db.Lock()
	defer db.Unlock()
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
		db.sync() // sync only if needed
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
func (db *StorDBService) Shutdown() (_ error) {
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
	return db.cfg.StorDbCfg().Type == utils.MetaPostgres &&
		(db.oldDBCfg.Opts.PgSSLMode != db.cfg.StorDbCfg().Opts.PgSSLMode ||
			db.oldDBCfg.Opts.PgSSLCert != db.cfg.StorDbCfg().Opts.PgSSLCert ||
			db.oldDBCfg.Opts.PgSSLKey != db.cfg.StorDbCfg().Opts.PgSSLKey ||
			db.oldDBCfg.Opts.PgSSLPassword != db.cfg.StorDbCfg().Opts.PgSSLPassword ||
			db.oldDBCfg.Opts.PgSSLCertMode != db.cfg.StorDbCfg().Opts.PgSSLCertMode ||
			db.oldDBCfg.Opts.PgSSLRootCert != db.cfg.StorDbCfg().Opts.PgSSLRootCert)
}

// StateChan returns signaling channel of specific state
func (db *StorDBService) StateChan(stateID string) chan struct{} {
	return db.stateDeps.StateChan(stateID)
}
