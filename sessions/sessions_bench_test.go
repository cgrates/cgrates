// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/chargers"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// benchMockClient is a minimal mock that returns canned responses for chargers.
type benchMockClient struct{}

func (benchMockClient) Call(ctx *context.Context, method string, args any, reply any) error {
	switch method {
	case utils.ChargerSv1ProcessEvent:
		*reply.(*[]*chargers.ChrgSProcessEventReply) = []*chargers.ChrgSProcessEventReply{
			{
				ChargerSProfile: "DEFAULT",
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "bench",
					Event: map[string]any{
						utils.AccountField: "1001",
						utils.Destination:  "1002",
					},
					APIOpts: map[string]any{
						utils.MetaRunID: utils.MetaDefault,
					},
				},
			},
		}
	}
	return nil
}

// setupBenchSessionS creates a fully wired SessionS with mock internal connections
// for benchmarking.
func setupBenchSessionS(b *testing.B, enableChargers bool) *SessionS {
	b.Helper()
	cfg := config.NewDefaultCGRConfig()

	// Disable RPC caching to avoid guardian lock overhead in benchmark
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0

	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		b.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	dm := engine.NewDataManager(dbCM, cfg, nil)
	dm.SetCache(cacheS)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	// Create ConnManager and register mock internal connections
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	mockCh := make(chan birpc.ClientConnector, 1)
	mockCh <- benchMockClient{}

	if enableChargers {
		chrgConnID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)
		connMgr.AddInternalConn(chrgConnID, utils.MetaChargers, mockCh)
		cfg.SessionSCfg().Conns[utils.MetaChargers] = []*config.DynamicConns{
			{
				Tenant:    "",
				FilterIDs: nil,
				ConnIDs:   []string{chrgConnID},
			},
		}
	}

	// Enable chargers in opts so ProcessEvent triggers charger processing
	cfg.SessionSCfg().Opts.Chargers = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", enableChargers, nil),
	}

	return NewSessionS(cfg, dm, cacheS, fltrs, connMgr)
}

// BenchmarkProcessEventChargersOnly benchmarks a full BiRPCv1ProcessEvent
// with only chargers enabled.
func BenchmarkProcessEventChargersOnly(b *testing.B) {
	sS := setupBenchSessionS(b, true)
	ctx := context.TODO()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.GenUUID(),
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.GenUUID(),
			},
		}
		var rply V1ProcessEventReply
		if err := sS.BiRPCv1ProcessEvent(ctx, ev, &rply); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkProcessEventNoSubsystems benchmarks ProcessEvent when no
// subsystem flags are enabled.
func BenchmarkProcessEventNoSubsystems(b *testing.B) {
	sS := setupBenchSessionS(b, false)
	// Disable chargers so nothing triggers
	sS.cfg.SessionSCfg().Opts.Chargers = nil
	ctx := context.TODO()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.GenUUID(),
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.GenUUID(),
			},
		}
		var rply V1ProcessEventReply
		if err := sS.BiRPCv1ProcessEvent(ctx, ev, &rply); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkProcessEventChargersParallel benchmarks concurrent ProcessEvent
// calls with chargers.
func BenchmarkProcessEventChargersParallel(b *testing.B) {
	sS := setupBenchSessionS(b, true)
	ctx := context.TODO()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.GenUUID(),
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
				},
				APIOpts: map[string]any{
					utils.MetaOriginID: utils.GenUUID(),
				},
			}
			var rply V1ProcessEventReply
			if err := sS.BiRPCv1ProcessEvent(ctx, ev, &rply); err != nil {
				b.Fatal(err)
			}
		}
	})
}
