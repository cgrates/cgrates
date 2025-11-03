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
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewDataDBService returns the DataDB Service
func NewDataDBService(cfg *config.CGRConfig, setVersions bool) *DataDBService {
	return &DataDBService{
		cfg:         cfg,
		setVersions: setVersions,
		stateDeps:   NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// DataDBService implements Service interface
type DataDBService struct {
	mu          sync.RWMutex
	cfg         *config.CGRConfig
	oldDBCfg    *config.DbCfg
	dm          *engine.DataManager
	setVersions bool
	stateDeps   *StateDependencies // channel subscriptions for state changes
}

// Start handles the service start.
func (db *DataDBService) Start(_ *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	cms, err := WaitForServiceState(utils.StateServiceUP, utils.ConnManager, registry, db.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	db.oldDBCfg = db.cfg.DbCfg().Clone()
	dbConnMap := new(engine.DBConnManager)
	for dbConnKey, dbconn := range db.cfg.DbCfg().DBConns {
		dbConn, err := engine.NewDataDBConn(dbconn.Type,
			dbconn.Host, dbconn.Port, dbconn.Name, dbconn.User,
			dbconn.Password, db.cfg.GeneralCfg().DBDataEncoding, dbconn.StringIndexedFields,
			dbconn.PrefixIndexedFields, dbconn.Opts, db.cfg.DbCfg().Items)
		if err != nil { // Cannot configure getter database, show stopper
			utils.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
			return err
		}
		dbConnMap.AddDataDBDriver(dbConnKey, dbConn)
		if dbconn.Type != utils.MetaInternal {
			utils.Logger.Info(fmt.Sprintf("<DB> connection established with <%s:%s> with DB name <%s>, Type <%s>", dbconn.Host, dbconn.Port, dbconn.Name, dbconn.Type))
		} else {
			utils.Logger.Info("<DB> Internal DB established")
		}
	}
	db.dm = engine.NewDataManager(dbConnMap, db.cfg, cms.(*ConnManagerService).ConnManager())
	if db.setVersions {
		dataDB, _, err := dbConnMap.GetConn(utils.CacheVersions)
		if err != nil {
			return err
		}
		if err = engine.OverwriteDBVersions(dataDB); err != nil {
			return err
		}
	} else {
		for _, dataDB := range db.dm.DataDB() {
			if err = engine.CheckVersions(dataDB); err != nil {
				return err
			}
		}
	}
	return
}

// Reload handles the change of config
func (db *DataDBService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.needsConnectionReload() {
		if err = db.dm.ReconnectAll(db.cfg); err != nil {
			return
		}
		db.oldDBCfg = db.cfg.DbCfg().Clone()
		return
	}
	for dbKey, dbConn := range db.cfg.DbCfg().DBConns {
		switch dbConn.Type {
		case utils.MetaMongo:
			mgo, canCast := db.dm.DataDB()[dbKey].(*engine.MongoStorage)
			if !canCast {
				return fmt.Errorf("can't conver DataDB of type %s to MongoStorage",
					dbConn.Type)
			}
			mgo.SetTTL(dbConn.Opts.MongoQueryTimeout)
		case utils.MetaPostgres, utils.MetaMySQL:
			msql, canCast := db.dm.DataDB()[dbKey].(*engine.SQLStorage)
			if !canCast {
				return fmt.Errorf("can't convert DB of type %s to SQLStorage",
					dbConn.Type)
			}
			msql.DB.SetMaxOpenConns(dbConn.Opts.SQLMaxOpenConns)
			msql.DB.SetMaxIdleConns(dbConn.Opts.SQLMaxIdleConns)
			msql.DB.SetConnMaxLifetime(dbConn.Opts.SQLConnMaxLifetime)
		case utils.MetaInternal:
			idb, canCast := db.dm.DataDB()[dbKey].(*engine.InternalDB)
			if !canCast {
				return fmt.Errorf("can't convert DB of type %s to InternalDB",
					dbConn.Type)
			}
			idb.SetStringIndexedFields(dbConn.StringIndexedFields)
			idb.SetPrefixIndexedFields(dbConn.PrefixIndexedFields)
		}

	}
	return
}

// Shutdown stops the service
func (db *DataDBService) Shutdown(registry *servmanager.ServiceRegistry) error {
	deps := []string{
		utils.ResourceS,
		utils.IPs,
		utils.TrendS,
		utils.RankingS,
		utils.StatS,
		utils.ThresholdS,
	}
	for _, svcID := range deps {
		if servmanager.State(registry.Lookup(svcID)) != utils.StateServiceUP {
			continue
		}
		_, err := WaitForServiceState(utils.StateServiceDOWN, svcID, registry, db.cfg.GeneralCfg().ConnectTimeout)
		if err != nil {
			return err
		}
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	for dataDBKey := range db.dm.DataDB() {
		db.dm.DataDB()[dataDBKey].Close()
	}
	return nil
}

// ServiceName returns the service name
func (db *DataDBService) ServiceName() string {
	return utils.DB
}

// ShouldRun returns if the service should be running
func (db *DataDBService) ShouldRun() bool { // db should allways run
	return true // ||db.mandatoryDB() || db.cfg.SessionSCfg().Enabled
}

// needsConnectionReload returns if the DB connection needs to reloaded
func (db *DataDBService) needsConnectionReload() bool {
	if len(db.oldDBCfg.DBConns) != len(db.cfg.DbCfg().DBConns) {
		return true
	}
	for dbConnKey, dbConn := range db.oldDBCfg.DBConns {
		if _, has := db.cfg.DbCfg().DBConns[dbConnKey]; !has {
			return true
		}
		if dbConn.Type != db.cfg.DbCfg().DBConns[dbConnKey].Type ||
			dbConn.Host != db.cfg.DbCfg().DBConns[dbConnKey].Host ||
			dbConn.Name != db.cfg.DbCfg().DBConns[dbConnKey].Name ||
			dbConn.Port != db.cfg.DbCfg().DBConns[dbConnKey].Port ||
			dbConn.User != db.cfg.DbCfg().DBConns[dbConnKey].User ||
			dbConn.Password != db.cfg.DbCfg().DBConns[dbConnKey].Password ||
			!utils.EqualUnorderedStringSlices(dbConn.StringIndexedFields,
				db.cfg.DbCfg().DBConns[dbConnKey].StringIndexedFields) ||
			!utils.EqualUnorderedStringSlices(dbConn.PrefixIndexedFields,
				db.cfg.DbCfg().DBConns[dbConnKey].PrefixIndexedFields) {
			return true
		}
		if db.cfg.DbCfg().DBConns[dbConnKey].Type == utils.MetaInternal { // in case of internal recreate the db using the new config
			for key, itm := range db.oldDBCfg.Items {
				if db.cfg.DbCfg().Items[key].Limit != itm.Limit &&
					db.cfg.DbCfg().Items[key].StaticTTL != itm.StaticTTL &&
					db.cfg.DbCfg().Items[key].TTL != itm.TTL &&
					db.cfg.DbCfg().Items[key].DBConn != itm.DBConn {
					return true
				}
			}
		}
		if db.oldDBCfg.DBConns[dbConnKey].Type == utils.MetaRedis &&
			(dbConn.Opts.RedisMaxConns != db.cfg.DbCfg().DBConns[dbConnKey].Opts.RedisMaxConns ||
				dbConn.Opts.RedisConnectAttempts != db.cfg.DbCfg().DBConns[dbConnKey].Opts.RedisConnectAttempts ||
				dbConn.Opts.RedisSentinel != db.cfg.DbCfg().DBConns[dbConnKey].Opts.RedisSentinel ||
				dbConn.Opts.RedisCluster != db.cfg.DbCfg().DBConns[dbConnKey].Opts.RedisCluster ||
				dbConn.Opts.RedisClusterSync != db.cfg.DbCfg().DBConns[dbConnKey].Opts.RedisClusterSync ||
				dbConn.Opts.RedisClusterOndownDelay != db.cfg.DbCfg().DBConns[dbConnKey].Opts.RedisClusterOndownDelay ||
				dbConn.Opts.RedisConnectTimeout != db.cfg.DbCfg().DBConns[dbConnKey].Opts.RedisConnectTimeout ||
				dbConn.Opts.RedisReadTimeout != db.cfg.DbCfg().DBConns[dbConnKey].Opts.RedisReadTimeout ||
				dbConn.Opts.RedisWriteTimeout != db.cfg.DbCfg().DBConns[dbConnKey].Opts.RedisWriteTimeout ||
				dbConn.Opts.RedisPoolPipelineWindow != db.cfg.DbCfg().DBConns[dbConnKey].Opts.RedisPoolPipelineWindow ||
				dbConn.Opts.RedisPoolPipelineLimit != db.cfg.DbCfg().DBConns[dbConnKey].Opts.RedisPoolPipelineLimit) {
			return true
		}
		if db.cfg.DbCfg().DBConns[dbConnKey].Type == utils.MetaPostgres &&
			(dbConn.Opts.PgSSLMode != db.cfg.DbCfg().DBConns[dbConnKey].Opts.PgSSLMode ||
				dbConn.Opts.PgSSLCert != db.cfg.DbCfg().DBConns[dbConnKey].Opts.PgSSLCert ||
				dbConn.Opts.PgSSLKey != db.cfg.DbCfg().DBConns[dbConnKey].Opts.PgSSLKey ||
				dbConn.Opts.PgSSLPassword != db.cfg.DbCfg().DBConns[dbConnKey].Opts.PgSSLPassword ||
				dbConn.Opts.PgSSLCertMode != db.cfg.DbCfg().DBConns[dbConnKey].Opts.PgSSLCertMode ||
				dbConn.Opts.PgSSLRootCert != db.cfg.DbCfg().DBConns[dbConnKey].Opts.PgSSLRootCert) {
			return true
		}
	}
	return false
}

// DataManager returns the DataManager object.
func (db *DataDBService) DataManager() *engine.DataManager {
	return db.dm
}

// StateChan returns signaling channel of specific state
func (db *DataDBService) StateChan(stateID string) chan struct{} {
	return db.stateDeps.StateChan(stateID)
}
