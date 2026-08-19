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
	cfg.DataDbCfg().RplConns = []string{connID}
	cfg.DataDbCfg().RplInterval = 0
	cfg.DataDbCfg().RplFailedDir = failedDir
	cfg.DataDbCfg().RplFiltered = false

	oldCfg := config.CgrConfig()
	oldCache := Cache
	oldConnMgr := connMgr
	config.SetCgrConfig(cfg)
	Cache = NewCacheS(cfg, nil, nil)
	cm := NewConnManager(cfg, nil)
	// Keep ConnManager dispatch without adding RPC transport overhead.
	Cache.SetWithoutReplicate(utils.CacheRPCConnections, connID,
		connector, nil, true, utils.NonTransactional)
	tb.Cleanup(func() {
		SetConnManager(oldConnMgr)
		Cache = oldCache
		config.SetCgrConfig(oldCfg)
	})
	return cm
}

func newTestReplicator(tb testing.TB, interval time.Duration, failedDir string,
	connector birpc.ClientConnector) *replicator {
	tb.Helper()
	r := newReplicator(setupReplicator(tb, failedDir, connector))
	// Set the interval after construction so callers control flushing without a ticker.
	r.interval = interval
	return r
}
