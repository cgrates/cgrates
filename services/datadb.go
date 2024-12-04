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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewDataDBService returns the DataDB Service
func NewDataDBService(cfg *config.CGRConfig, connMgr *engine.ConnManager, setVersions bool,
	srvDep map[string]*sync.WaitGroup,
	srvIndexer *servmanager.ServiceIndexer) *DataDBService {
	return &DataDBService{
		cfg:         cfg,
		dbchan:      make(chan *engine.DataManager, 1),
		connMgr:     connMgr,
		setVersions: setVersions,
		srvDep:      srvDep,
		srvIndexer:  srvIndexer,
		stateDeps:   NewStateDependencies([]string{utils.StateServiceUP}),
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

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start handles the service start.
func (db *DataDBService) Start(*context.Context, context.CancelFunc) (err error) {
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
	if err != nil { // Cannot configure getter database, show stopper
		utils.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
		return
	}
	db.dm = engine.NewDataManager(dbConn, db.cfg.CacheCfg(), db.connMgr)

	if db.setVersions {
		err = engine.OverwriteDBVersions(dbConn)
	} else {
		err = engine.CheckVersions(db.dm.DataDB())
	}
	if err != nil {
		return err
	}

	db.dbchan <- db.dm
	close(db.stateDeps.StateChan(utils.StateServiceUP))
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
	return db.dm != nil && db.dm.DataDB() != nil
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

// StateChan returns signaling channel of specific state
func (db *DataDBService) StateChan(stateID string) chan struct{} {
	return db.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (db *DataDBService) IntRPCConn() birpc.ClientConnector {
	return db.intRPCconn
}
