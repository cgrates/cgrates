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

package engine

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

func newLockerTestDataManager(t *testing.T, cfg *config.CGRConfig,
	locker *guardian.Locker) *DataManager {
	t.Helper()
	data, err := NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbConns := NewDBConnManager(map[string]DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	return NewDataManager(dbConns, cfg, nil, locker)
}

func TestDataManagerCacheLockCoordination(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := NewLocker(cfg)
	dm := newLockerTestDataManager(t, cfg, locker)
	cache := NewCacheS(cfg, dm, nil, nil, locker)

	unlock := dm.Lock("shared")
	started := make(chan struct{})
	acquired := make(chan func(), 1)
	go func() {
		close(started)
		acquired <- cache.LockRPCResponse("shared")
	}()
	<-started
	select {
	case cacheUnlock := <-acquired:
		cacheUnlock()
		t.Fatal("CacheS did not wait for DataManager")
	case <-time.After(20 * time.Millisecond):
	}
	unlock()
	select {
	case cacheUnlock := <-acquired:
		cacheUnlock()
	case <-time.After(time.Second):
		t.Fatal("CacheS remained blocked after DataManager unlock")
	}
}

func testReplicatorLockCoordination(t *testing.T, dm *DataManager, replicator *replicator) {
	t.Helper()
	unlock := dm.Lock("replicator")
	started := make(chan struct{})
	acquired := make(chan func(), 1)
	go func() {
		close(started)
		acquired <- replicator.locker.Lock("replicator")
	}()
	<-started
	select {
	case replicatorUnlock := <-acquired:
		replicatorUnlock()
		t.Fatal("replicator did not wait for DataManager")
	case <-time.After(20 * time.Millisecond):
	}
	unlock()
	select {
	case replicatorUnlock := <-acquired:
		replicatorUnlock()
	case <-time.After(time.Second):
		t.Fatal("replicator remained blocked after DataManager unlock")
	}
}

func TestDataManagerReplicatorsKeepLocker(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DbCfg().DBConns[utils.MetaDefault].Opts.InternalDBDumpInterval = 0
	cfg.DbCfg().DBConns[utils.MetaDefault].Opts.InternalDBRewriteInterval = 0
	locker := NewLocker(cfg)
	dm := newLockerTestDataManager(t, cfg, locker)
	defer dm.Close()

	for _, replicator := range dm.dbConns.replicators {
		testReplicatorLockCoordination(t, dm, replicator)
	}
	if err := dm.ReconnectAll(cfg); err != nil {
		t.Fatal(err)
	}
	for _, replicator := range dm.dbConns.replicators {
		testReplicatorLockCoordination(t, dm, replicator)
	}
}
