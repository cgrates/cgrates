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
	dm := engine.NewDataManager(dbCM, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	// Create ConnManager and register mock internal connections
	connMgr := engine.NewConnManager(cfg)
	mockCh := make(chan birpc.ClientConnector, 1)
	mockCh <- benchMockClient{}

	if enableChargers {
		chrgConnID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)
		connMgr.AddInternalConn(chrgConnID, utils.MetaChargers, mockCh)
		cfg.SessionSCfg().Conns[utils.MetaChargers] = []*config.DynamicStringSliceOpt{
			{
				Tenant:    "",
				FilterIDs: nil,
				Values:    []string{chrgConnID},
			},
		}
	}

	// Enable chargers in opts so ProcessEvent triggers charger processing
	cfg.SessionSCfg().Opts.Chargers = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", enableChargers, nil),
	}

	return NewSessionS(cfg, dm, fltrs, connMgr)
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
			APIOpts: map[string]any{},
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
			APIOpts: map[string]any{},
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
				APIOpts: map[string]any{},
			}
			var rply V1ProcessEventReply
			if err := sS.BiRPCv1ProcessEvent(ctx, ev, &rply); err != nil {
				b.Fatal(err)
			}
		}
	})
}
