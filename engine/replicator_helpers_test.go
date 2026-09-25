// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

type mockConnector struct {
	err error
}

func (c *mockConnector) Call(_ *context.Context, _ string, _, _ any) error {
	return c.err
}

func setupReplicator(tb testing.TB, failedDir string,
	connector birpc.ClientConnector) *ConnManager {
	tb.Helper()
	cfg := config.NewDefaultCGRConfig()
	connID := "replicator-test"
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	dbCfg := cfg.DbCfg().DBConns[utils.MetaDefault]
	dbCfg.RplConns = []string{connID}
	dbCfg.RplInterval = 0
	dbCfg.RplFailedDir = failedDir
	dbCfg.RplFiltered = false

	locker := NewLocker(cfg)
	cache := NewCacheS(cfg, nil, nil, nil, locker)
	cm := NewConnManager(cfg)
	cm.SetCache(cache)
	// Keep ConnManager dispatch without adding RPC transport overhead.
	cache.SetWithoutReplicate(utils.CacheRPCConnections, connID,
		connector, nil, true, utils.NonTransactional)
	return cm
}

func testReplicator(dm *DataManager) *replicator {
	return dm.dbConns.GetReplicator(utils.MetaDefault)
}

func newTestReplicator(tb testing.TB, interval time.Duration, failedDir string,
	connector birpc.ClientConnector) *replicator {
	tb.Helper()
	cm := setupReplicator(tb, failedDir, connector)
	r := newReplicator(cm.cfg.DbCfg().DBConns[utils.MetaDefault], cm, NewLocker(cm.cfg))
	// Set the interval after construction so callers control flushing without a ticker.
	r.interval = interval
	return r
}
